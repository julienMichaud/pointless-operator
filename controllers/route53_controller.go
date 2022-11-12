/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	cachev1alpha1 "github.com/julienMichaud/pointless-operator/api/v1alpha1"
)

const route53Finalizer = "route53.jmichaud.net/finalizer"

// Route53Reconciler reconciles a Route53 object
type Route53Reconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=cache.jmichaud.net,resources=route53s,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.jmichaud.net,resources=route53s/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.jmichaud.net,resources=route53s/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Route53 object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *Route53Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the Memcached instance
	// The purpose is check if the Custom Resource for the Kind Memcached
	// is applied on the cluster if not we return nil to stop the reconciliation
	route53 := &cachev1alpha1.Route53{}
	err := r.Get(ctx, req.NamespacedName, route53)

	if err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then, it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			log.Info("route53 resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get memcached")
		return ctrl.Result{}, err
	}

	// Let's just set the status as Unknown when no status are available
	if route53.Status.Conditions == nil || len(route53.Status.Conditions) == 0 {
		meta.SetStatusCondition(&route53.Status.Conditions, metav1.Condition{Type: "UNKNOWN", Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "Starting reconciliation"})
		if err = r.Status().Update(ctx, route53); err != nil {
			log.Error(err, "Failed to update route53 status")
			return ctrl.Result{}, err
		}

		// Let's re-fetch the route53 Custom Resource after update the status
		// so that we have the latest state of the resource on the cluster and we will avoid
		// raise the issue "the object has been modified, please apply
		// your changes to the latest version and try again" which would re-trigger the reconciliation
		// if we try to update it again in the following operations
		if err := r.Get(ctx, req.NamespacedName, route53); err != nil {
			log.Error(err, "Failed to re-fetch route53")
			return ctrl.Result{}, err
		}
	}

	// Let's add a finalizer. Then, we can define some operations which should
	// occurs before the custom resource to be deleted.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers
	// if !controllerutil.ContainsFinalizer(memcached, memcachedFinalizer) {
	// 	log.Info("Adding Finalizer for Memcached")
	// 	if ok := controllerutil.AddFinalizer(memcached, memcachedFinalizer); !ok {
	// 		log.Error(err, "Failed to add finalizer into the custom resource")
	// 		return ctrl.Result{Requeue: true}, nil
	// 	}

	// 	if err = r.Update(ctx, memcached); err != nil {
	// 		log.Error(err, "Failed to update custom resource to add finalizer")
	// 		return ctrl.Result{}, err
	// 	}
	// }

	// Check if the Memcached instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isRoute53MarkedToBeDeleted := route53.GetDeletionTimestamp() != nil
	if isRoute53MarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(route53, route53Finalizer) {
			log.Info("Performing Finalizer Operations for Route53 before delete CR")

			// Let's add here an status "Downgrade" to define that this resource begin its process to be terminated.
			meta.SetStatusCondition(&route53.Status.Conditions, metav1.Condition{Type: "DEGRADED",
				Status: metav1.ConditionUnknown, Reason: "Finalizing",
				Message: fmt.Sprintf("Performing finalizer operations for the custom resource: %s ", route53.Name)})

			if err := r.Status().Update(ctx, route53); err != nil {
				log.Error(err, "Failed to update Route53 status")
				return ctrl.Result{}, err
			}

			// Perform all operations required before remove the finalizer and allow
			// the Kubernetes API to remove the custom resource.
			// r.doFinalizerOperationsForMemcached(memcached)
			err := r.handleDeleteFinalizer(req, route53)
			if err != nil {
				meta.SetStatusCondition(&route53.Status.Conditions, metav1.Condition{Type: "DEGRADED",
					Status: metav1.ConditionUnknown, Reason: "Finalizing",
					Message: fmt.Sprintf("Error while deleting CR aws resources: %s ", route53.Name)})

				if err := r.Status().Update(ctx, route53); err != nil {
					log.Error(err, "Failed to update Route53 status")
					return ctrl.Result{}, err
				}
			}

			// TODO(user): If you add operations to the doFinalizerOperationsForMemcached method
			// then you need to ensure that all worked fine before deleting and updating the Downgrade status
			// otherwise, you should requeue here.

			// Re-fetch the memcached Custom Resource before update the status
			// so that we have the latest state of the resource on the cluster and we will avoid
			// raise the issue "the object has been modified, please apply
			// your changes to the latest version and try again" which would re-trigger the reconciliation
			if err := r.Get(ctx, req.NamespacedName, route53); err != nil {
				log.Error(err, "Failed to re-fetch route53")
				return ctrl.Result{}, err
			}

			meta.SetStatusCondition(&route53.Status.Conditions, metav1.Condition{Type: "DEGRADED",
				Status: metav1.ConditionTrue, Reason: "Finalizing",
				Message: fmt.Sprintf("Finalizer operations for custom resource %s name were successfully accomplished", route53.Name)})

			if err := r.Status().Update(ctx, route53); err != nil {
				log.Error(err, "Failed to update Route53 status")
				return ctrl.Result{}, err
			}

			log.Info("Removing Finalizer for route53 after successfully perform the operations")
			if ok := controllerutil.RemoveFinalizer(route53, route53Finalizer); !ok {
				log.Error(err, "Failed to remove finalizer for route53")
				return ctrl.Result{Requeue: true}, nil
			}

			if err := r.Update(ctx, route53); err != nil {
				log.Error(err, "Failed to remove finalizer for route53")
				return ctrl.Result{}, err
			}

			err = r.handleDelete(req, route53)
			if err != nil {
				meta.SetStatusCondition(&route53.Status.Conditions, metav1.Condition{Type: "DEGRADED",
					Status: metav1.ConditionUnknown, Reason: "Finalizing",
					Message: fmt.Sprintf("Error while deleting CR: %s ", route53.Name)})

				if err := r.Status().Update(ctx, route53); err != nil {
					log.Error(err, "Failed to update Route53 status")
					return ctrl.Result{}, err
				}
			}
		}
		return ctrl.Result{}, nil
	}

	err = r.handleCreate(req, route53)
	if err != nil {
		log.Error(err, "Failed to create new record")

		// The following implementation will update the status
		meta.SetStatusCondition(&route53.Status.Conditions, metav1.Condition{Type: "Available",
			Status: metav1.ConditionFalse, Reason: "Reconciling",
			Message: fmt.Sprintf("Failed to create Record for the custom resource (%s): (%s)", route53.Name, err)})

		if err := r.Status().Update(ctx, route53); err != nil {
			log.Error(err, "Failed to update Route53 status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	// The following implementation will update the status
	meta.SetStatusCondition(&route53.Status.Conditions, metav1.Condition{Type: "Available",
		Status: metav1.ConditionTrue, Reason: "Reconciling",
		Message: fmt.Sprintf("Deployment for custom resource (%s) created successfully", route53.Name)})

	if err := r.Status().Update(ctx, route53); err != nil {
		log.Error(err, "Failed to update Memcached status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Route53Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.Route53{}).
		Complete(r)
}

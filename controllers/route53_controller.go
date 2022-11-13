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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	cachev1alpha1 "github.com/julienMichaud/pointless-operator/api/v1alpha1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const route53Finalizer = "route53.jmichaud.net/finalizer"

// Route53Reconciler reconciles a Route53 object
type Route53Reconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Log      logr.Logger
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

	const route53Finalizer = "route53.jmichaud.net/finalizer"

	const (
		// typeAvailableMemcached represents the status of the Deployment reconciliation
		typeAvailableRoute53 = "Available"
		// typeDegradedMemcached represents the status used when the custom resource is deleted and the finalizer operations are must to occur.
		typeDegradedRoute53 = "Degraded"
	)

	route53 := &cachev1alpha1.Route53{}

	err := r.Get(context.TODO(), req.NamespacedName, route53)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("route53 resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get memcached")
		return reconcile.Result{}, err
	}

	// Let's add a finalizer. Then, we can define some operations which should
	// occurs before the custom resource to be deleted.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers
	if !controllerutil.ContainsFinalizer(route53, route53Finalizer) {
		log.Info("Adding Finalizer for Route53")
		if ok := controllerutil.AddFinalizer(route53, route53Finalizer); !ok {
			log.Error(err, "Failed to add finalizer into the custom resource")
			return ctrl.Result{Requeue: true}, nil
		}

		if err = r.Update(ctx, route53); err != nil {
			log.Error(err, "Failed to update custom resource to add finalizer")
			return ctrl.Result{}, err
		}
	}

	isRoute53MarkedToBeDeleted := route53.GetDeletionTimestamp() != nil

	if isRoute53MarkedToBeDeleted {
		log.Info("marked to be deleted")
		if controllerutil.ContainsFinalizer(route53, route53Finalizer) {
			log.Info("Performing Finalizer Operations for Route53 before delete CR")

			// Let's add here an status "Downgrade" to define that this resource begin its process to be terminated.
			meta.SetStatusCondition(&route53.Status.Conditions, metav1.Condition{Type: typeDegradedRoute53,
				Status: metav1.ConditionUnknown, Reason: "Finalizing",
				Message: fmt.Sprintf("Performing finalizer operations for the custom resource: %s ", route53.Name)})

			if err := r.Status().Update(ctx, route53); err != nil {
				log.Error(err, "Failed to update Route53 status")
				return ctrl.Result{}, err
			}

			// Perform all operations required before remove the finalizer and allow
			// the Kubernetes API to remove the custom resource.
			// TODO(user): If you add operations to the doFinalizerOperationsForMemcached method
			// then you need to ensure that all worked fine before deleting and updating the Downgrade status
			// otherwise, you should requeue here.
			if err := r.doFinalizerOperationsForRoute53(route53); err != nil {
				log.Infof("error while executing doFinalizerOperationsForRoute53 %s ", err)
				return ctrl.Result{}, err
			}

			if err := r.Status().Update(ctx, route53); err != nil {
				log.Error(err, "Failed to update Route53 status")
				return ctrl.Result{}, err
			}

			// Re-fetch the memcached Custom Resource before update the status
			// so that we have the latest state of the resource on the cluster and we will avoid
			// raise the issue "the object has been modified, please apply
			// your changes to the latest version and try again" which would re-trigger the reconciliation
			if err := r.Get(ctx, req.NamespacedName, route53); err != nil {
				log.Error(err, "Failed to re-fetch route53")
				return ctrl.Result{}, err
			}

			meta.SetStatusCondition(&route53.Status.Conditions, metav1.Condition{Type: typeDegradedRoute53,
				Status: metav1.ConditionTrue, Reason: "Finalizing",
				Message: fmt.Sprintf("Finalizer operations for custom resource %s name were successfully accomplished", route53.Name)})

			if err := r.Status().Update(ctx, route53); err != nil {
				log.Error(err, "Failed to update Route53 status")
				return ctrl.Result{}, err
			}

			log.Info("Removing Finalizer for Route53 after successfully perform the operations")
			if ok := controllerutil.RemoveFinalizer(route53, route53Finalizer); !ok {
				log.Error(err, "Failed to remove finalizer for Route53")
				return ctrl.Result{Requeue: true}, nil
			}

			if err := r.Update(ctx, route53); err != nil {
				log.Error(err, "Failed to remove finalizer for Route53")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if err = r.handleCreate(ctx, req, route53); err != nil {
		log.Infof("error while creating record %s ", err)
		return ctrl.Result{}, err
	}

	if err := r.Status().Update(ctx, route53); err != nil {
		log.Error(err, "Failed to update Route53 status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *Route53Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.Route53{}).
		Complete(r)
}

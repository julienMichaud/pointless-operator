package controllers

import (
	cachev1alpha1 "github.com/julienMichaud/pointless-operator/api/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *Route53Reconciler) handleDeleteFinalizer(request reconcile.Request, instance *cachev1alpha1.Route53) error {
	// log := r.Log.WithValues("route53", instance.ObjectMeta.Namespace)

	// log.Info("Creating route53 record: " + instance.GetObjectMeta().GetName())

	// url, err := util.GetMonitorURL(r.Client, instance)
	// if err != nil {
	// 	return err
	// }

	// // Extract provider specific configuration
	// providerConfig := monitorService.ExtractConfig(instance.Spec)

	// // Create monitor Model
	// monitor := models.Monitor{Name: monitorName, URL: url, Config: providerConfig}

	// Add monitor for provider
	// monitorService.Add(monitor)
	log.Log.Info("for the CR %s, delete things on AWS now....", instance.Name)

	return nil
}

func (r *Route53Reconciler) handleDelete(request reconcile.Request, instance *cachev1alpha1.Route53) error {
	// log := r.Log.WithValues("route53", instance.ObjectMeta.Namespace)

	// log.Info("Creating route53 record: " + instance.GetObjectMeta().GetName())

	// url, err := util.GetMonitorURL(r.Client, instance)
	// if err != nil {
	// 	return err
	// }

	// // Extract provider specific configuration
	// providerConfig := monitorService.ExtractConfig(instance.Spec)

	// // Create monitor Model
	// monitor := models.Monitor{Name: monitorName, URL: url, Config: providerConfig}

	// Add monitor for provider
	// monitorService.Add(monitor)
	log.Log.Info("for the CR %s, delete Cr now....", instance.Name)

	return nil
}

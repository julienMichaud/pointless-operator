package controllers

import (
	"context"

	cachev1alpha1 "github.com/julienMichaud/pointless-operator/api/v1alpha1"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *Route53Reconciler) handleDeleteFinalizer(ctx context.Context, request reconcile.Request, instance *cachev1alpha1.Route53) error {

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
	log.Printf("for the CR %s, delete things on AWS now....", instance.Name)

	return nil
}

func (r *Route53Reconciler) doFinalizerOperationsForRoute53(instance *cachev1alpha1.Route53) error {

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

	log.Printf("for the CR %s with domain %s and record %s, delete record now on AWS...", instance.Name, instance.Spec.Domain, instance.Spec.RecordType)
	return nil

}

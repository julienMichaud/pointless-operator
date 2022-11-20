package controllers

import (
	"context"
	"fmt"

	cachev1alpha1 "github.com/julienMichaud/pointless-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"

	checkAWS "github.com/julienMichaud/pointless-operator/pkg/aws"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *Route53Reconciler) handleCreate(ctx context.Context, request reconcile.Request, instance *cachev1alpha1.Route53) error {

	log.Printf("for the CR %s, will check if record with domain name %s and type %s already exist", instance.Name, instance.Spec.Domain, instance.Spec.RecordType)

	exist, err := checkAWS.DoesRecordExist(*r.AWS, instance.Spec.Domain, instance.Spec.RecordType)
	if err != nil {
		log.Error(err, "Failed to check if record was set")
		return err
	}

	if exist {
		log.Printf("record already exist, not doing anything")

		meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{Type: "Available",
			Status: metav1.ConditionTrue, Reason: "Reconciling",
			Message: fmt.Sprintf("DNS record for custom resource (%s) with domain %s and type  %s already exist", instance.Name, instance.Spec.Domain, instance.Spec.RecordType)})

		if err := r.Status().Update(ctx, instance); err != nil {
			log.Error(err, "Failed to update Route53 status")
			return err

		}
		return nil
	}

	log.Printf("for the CR %s, will add record with domain name %s and type %s", instance.Name, instance.Spec.Domain, instance.Spec.RecordType)

	err = checkAWS.CreateRecord(*r.AWS, instance.Spec.Domain, instance.Spec.RecordType, instance.Spec.Value)
	if err != nil {
		log.Error(err, "Failed to create record %s", instance.Spec.Domain)
		return err
	}

	// The following implementation will update the status
	meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{Type: "Available",
		Status: metav1.ConditionTrue, Reason: "Reconciling",
		Message: fmt.Sprintf("DNS record for custom resource (%s) with domain %s and type  %s created successfully", instance.Name, instance.Spec.Domain, instance.Spec.RecordType)})

	if err := r.Status().Update(ctx, instance); err != nil {
		log.Error(err, "Failed to update Route53 status")
		return err

	}

	return nil
}

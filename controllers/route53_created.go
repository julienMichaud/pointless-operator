package controllers

import (
	"context"
	"fmt"

	cachev1alpha1 "github.com/julienMichaud/pointless-operator/api/v1alpha1"
	"go.opentelemetry.io/otel"
	"k8s.io/apimachinery/pkg/api/meta"

	checkAWS "github.com/julienMichaud/pointless-operator/pkg/aws"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *Route53Reconciler) handleCreate(ctx context.Context, contextLogging log.Entry, request reconcile.Request, instance *cachev1alpha1.Route53) error {

	spanCtx, span := otel.Tracer(name).Start(ctx, "handleCreate")
	defer span.End()

	contextLogging.Printf("for the CR %s, will check if record with domain name %s and type %s already exist", instance.Name, instance.Spec.Domain, instance.Spec.RecordType)

	recordChanger := checkAWS.Route53RecordChanger{Client: r.AWS}
	recordRetriever := checkAWS.Route53RecordRetriever{Client: r.AWS}

	exist, recordName, recordType, recordValue, recordTTL, err := checkAWS.RetrieveRecordOnR53(spanCtx, recordRetriever, instance.Spec.Domain)
	if err != nil {
		contextLogging.Error(err, "Failed to check if record exist")
		return err
	}

	if exist {
		contextLogging.Printf("record already exist, checking if every values is correct on AWS side")

		if recordName != fmt.Sprintf(instance.Spec.Domain+".") || recordType != instance.Spec.RecordType || recordValue != instance.Spec.Value || recordTTL != instance.Spec.TTL {

			contextLogging.Printf("got: %s,%s,%s,%v want: %s.,%s,%s,%v", recordName, recordType, recordValue, recordTTL, instance.Spec.Domain, instance.Spec.RecordType, instance.Spec.Value, instance.Spec.TTL)

			err = checkAWS.CreateRecord(spanCtx, recordChanger, instance.Spec.Domain, instance.Spec.RecordType, instance.Spec.Value, instance.Spec.TTL)
			if err != nil {
				contextLogging.Error(err, "Failed to update record %s", instance.Spec.Domain)

				meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{Type: "Available",
					Status: metav1.ConditionTrue, Reason: "Reconciling",
					Message: fmt.Sprintf("Fail to update dns record on route53 (%s) with domain %s and type  %s with value %s ", instance.Name, instance.Spec.Domain, instance.Spec.RecordType, instance.Spec.Value)})

				if err := r.Status().Update(ctx, instance); err != nil {
					contextLogging.Error(err, "Failed to update Route53 status")
					return err

				}
				return err
			}

		}

		meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{Type: "Available",
			Status: metav1.ConditionTrue, Reason: "Reconciling",
			Message: fmt.Sprintf("DNS record for custom resource (%s) with domain %s and type  %s already exist", instance.Name, instance.Spec.Domain, instance.Spec.RecordType)})

		if err := r.Status().Update(ctx, instance); err != nil {
			contextLogging.Error(err, "Failed to update Route53 status")
			return err

		}
		return nil
	} else {

		contextLogging.Printf("for the CR %s, will add record with domain name %s and type %s", instance.Name, instance.Spec.Domain, instance.Spec.RecordType)

		err = checkAWS.CreateRecord(ctx, recordChanger, instance.Spec.Domain, instance.Spec.RecordType, instance.Spec.Value, instance.Spec.TTL)
		if err != nil {
			contextLogging.Error(err, "Failed to create record %s", instance.Spec.Domain)
			return err
		}

		// The following implementation will update the status
		meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{Type: "Available",
			Status: metav1.ConditionTrue, Reason: "Reconciling",
			Message: fmt.Sprintf("DNS record for custom resource (%s) with domain %s and type  %s created successfully", instance.Name, instance.Spec.Domain, instance.Spec.RecordType)})

		if err := r.Status().Update(ctx, instance); err != nil {
			contextLogging.Error(err, "Failed to update Route53 status")
			return err

		}

		return nil
	}
}

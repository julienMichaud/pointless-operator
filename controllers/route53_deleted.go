package controllers

import (
	"context"

	cachev1alpha1 "github.com/julienMichaud/pointless-operator/api/v1alpha1"
	checkAWS "github.com/julienMichaud/pointless-operator/pkg/aws"
	"go.opentelemetry.io/otel"

	log "github.com/sirupsen/logrus"
)

func (r *Route53Reconciler) doFinalizerOperationsForRoute53(ctx context.Context, instance *cachev1alpha1.Route53, contextLogging log.Entry) error {

	spanCtx, span := otel.Tracer(name).Start(ctx, "doFinalizerOperationsForRoute53")
	defer span.End()

	contextLogging.Printf("for the CR %s, will check if record with domain name %s and type %s exist", instance.Name, instance.Spec.Domain, instance.Spec.RecordType)
	recordRetriever := checkAWS.Route53RecordRetriever{Client: r.AWS}

	exist, _, _, _, _, err := checkAWS.RetrieveRecordOnR53(spanCtx, recordRetriever, instance.Spec.Domain)
	if err != nil {
		contextLogging.Error(err, "Failed to check if record was set")
		return err
	}

	if exist {

		contextLogging.Printf("for the CR %s with domain %s and record %s, delete record now on AWS...", instance.Name, instance.Spec.Domain, instance.Spec.RecordType)
		err = checkAWS.DeleteRecord(spanCtx, *r.AWS, instance.Spec.Domain, instance.Spec.RecordType, instance.Spec.Value, instance.Spec.TTL)
		if err != nil {
			contextLogging.Error(err, "Failed to delete record")
			return err
		}
	}

	return nil

}

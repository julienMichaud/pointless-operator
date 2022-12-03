package controllers

import (
	cachev1alpha1 "github.com/julienMichaud/pointless-operator/api/v1alpha1"
	checkAWS "github.com/julienMichaud/pointless-operator/pkg/aws"

	log "github.com/sirupsen/logrus"
)

func (r *Route53Reconciler) doFinalizerOperationsForRoute53(instance *cachev1alpha1.Route53) error {

	log.Printf("for the CR %s, will check if record with domain name %s and type %s exist", instance.Name, instance.Spec.Domain, instance.Spec.RecordType)

	exist, _, _, _, _, err := checkAWS.RetrieveRecordOnR53(*r.AWS, instance.Spec.Domain)
	if err != nil {
		log.Error(err, "Failed to check if record was set")
		return err
	}

	if exist {

		log.Printf("for the CR %s with domain %s and record %s, delete record now on AWS...", instance.Name, instance.Spec.Domain, instance.Spec.RecordType)
		err = checkAWS.DeleteRecord(*r.AWS, instance.Spec.Domain, instance.Spec.RecordType, instance.Spec.Value, instance.Spec.TTL)
		if err != nil {
			log.Error(err, "Failed to delete record")
			return err
		}
	}

	return nil

}

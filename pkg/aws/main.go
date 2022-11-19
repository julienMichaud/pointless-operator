package pointlessAws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	log "github.com/sirupsen/logrus"
)

func DoesRecordExist(route53client route53.Client, record, recordType string) (bool, error) {
	id := "Z041619718A5JIL5IXWDC"

	records, err := route53client.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{HostedZoneId: &id})
	if err != nil {
		return false, err
	}

	for _, v := range records.ResourceRecordSets { // should modify to work with paginate
		if *v.Name == fmt.Sprintf("%s.", record) {
			log.Printf("found record %s on route53", record)
			return true, nil
		}

	}
	log.Printf("Record %s dont exist on route53", record)
	return false, nil

}

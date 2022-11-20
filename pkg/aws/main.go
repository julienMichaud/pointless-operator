package pointlessAws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

func DoesRecordExist(route53client route53.Client, record, recordType string) (bool, error) {
	id := "Z041619718A5JIL5IXWDC"

	records, err := route53client.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{HostedZoneId: &id})
	if err != nil {
		return false, err
	}

	for _, v := range records.ResourceRecordSets { // should modify to work with paginate
		if *v.Name == fmt.Sprintf("%s.", record) {
			return true, nil
		}

	}
	return false, nil

}

func CreateRecord(route53client route53.Client, record, recordType, value string) error {
	id := "Z041619718A5JIL5IXWDC"

	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionCreate,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: aws.String(record + "."),
						ResourceRecords: []types.ResourceRecord{
							{
								Value: &value,
							},
						},
						Type: types.RRTypeA,
						TTL:  aws.Int64(60),
					},
				},
			},
		},
		HostedZoneId: &id,
	}

	_, err := route53client.ChangeResourceRecordSets(context.TODO(), input)

	if err != nil {
		return err
	}
	return nil

}

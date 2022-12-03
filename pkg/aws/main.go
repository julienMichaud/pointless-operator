package pointlessAws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

func RetrieveRecordOnR53(route53client route53.Client, record string) (found bool, recordName string, recordTypereturned string, recordValue string, ttl int64, error error) {
	id := "Z041619718A5JIL5IXWDC"

	records, err := route53client.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{HostedZoneId: &id})
	if err != nil {
		return false, "", "", "", 0, err
	}

	for _, v := range records.ResourceRecordSets { // should modify to work with paginate
		if *v.Name == fmt.Sprintf("%s.", record) {
			return true, *v.Name, string(v.Type), *v.ResourceRecords[0].Value, *v.TTL, nil
		}

	}
	return false, "", "", "", 0, nil

}

func CreateRecord(route53client route53.Client, record, recordType, value string, ttl int64) error {
	id := "Z041619718A5JIL5IXWDC"

	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionUpsert,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: aws.String(record + "."),
						ResourceRecords: []types.ResourceRecord{
							{
								Value: &value,
							},
						},
						Type: types.RRTypeA,
						TTL:  &ttl,
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

func DeleteRecord(route53client route53.Client, record, recordType, value string, ttl int64) error {

	id := "Z041619718A5JIL5IXWDC"

	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionDelete,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: aws.String(record + "."),
						ResourceRecords: []types.ResourceRecord{
							{
								Value: &value,
							},
						},
						Type: types.RRTypeA,
						TTL:  &ttl,
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

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

type RecordChanger interface {
	ChangeRecordSet(input *route53.ChangeResourceRecordSetsInput) error
}

type Route53RecordChanger struct {
	Client *route53.Client
}

func (r Route53RecordChanger) ChangeRecordSet(input *route53.ChangeResourceRecordSetsInput) error {

	_, err := r.Client.ChangeResourceRecordSets(context.TODO(), input)

	return err
}

func CreateRecord(recordChanger RecordChanger, record, recordType, value string, ttl int64) error {
	id := "Z041619718A5JIL5IXWDC"

	var recordTypeAWS types.RRType

	if recordType == "A" {
		recordTypeAWS = types.RRTypeA
	}

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
						Type: recordTypeAWS,
						TTL:  &ttl,
					},
				},
			},
		},
		HostedZoneId: &id,
	}

	err := recordChanger.ChangeRecordSet(input)

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

package pointlessAws

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

func validIP4(ipAddress string) bool {
	ipAddress = strings.Trim(ipAddress, " ")

	re, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	return re.MatchString(ipAddress)
}

type Route53RecordChangerMock struct {
}

func (r Route53RecordChangerMock) ChangeRecordSet(input *route53.ChangeResourceRecordSetsInput) error {
	validateIP := validIP4(*input.ChangeBatch.Changes[0].ResourceRecordSet.ResourceRecords[0].Value)
	if validateIP != true {
		return fmt.Errorf("not a valid IP")
	}

	if *input.ChangeBatch.Changes[0].ResourceRecordSet.TTL > int64(2147483647) {
		return fmt.Errorf("TTL must have value less than or equal to 2147483647")
	}

	return nil
}

func TestCreateRecord(t *testing.T) {
	recordChanger := Route53RecordChangerMock{}

	type test struct {
		domain     string
		recordType string
		record     string
		TTL        int64
		wantErr    bool
	}

	tests := []test{
		{domain: "tes1", recordType: "A", record: "8.8.8.8", TTL: int64(60), wantErr: false},        // ok record
		{domain: "tes2", recordType: "A", record: "8.8.8.8test2", TTL: int64(60), wantErr: true},    // invalid ip
		{domain: "tes1", recordType: "A", record: "8.8.8.8", TTL: int64(2147483648), wantErr: true}, // invalid TTL
	}

	for _, tc := range tests {
		err := CreateRecord(recordChanger, tc.domain, tc.recordType, tc.record, tc.TTL)
		if err != nil {
			if tc.wantErr == false {
				t.Errorf("didnt want err, got '%s'", err)
			}
		}
	}
}

type Route53RecordRetrieverMock struct {
}

func (r Route53RecordRetrieverMock) ListRecords(input *route53.ListResourceRecordSetsInput) (*route53.ListResourceRecordSetsOutput, error) {
	ttl := int64(60)
	output := route53.ListResourceRecordSetsOutput{ResourceRecordSets: []types.ResourceRecordSet{
		{
			Name: aws.String("test1."),
			Type: types.RRTypeA,
			ResourceRecords: []types.ResourceRecord{
				{Value: aws.String("8.8.8.8")},
			},
			TTL: &ttl,
		},
	},
	}

	return &output, nil
}

func TestRetrieveRecordOnR53(t *testing.T) {
	recordRetriever := Route53RecordRetrieverMock{}

	type test struct {
		recordGiven     string
		foundWant       bool
		recordWant      string
		recordTypeWant  string
		recordValueWant string
		ttlWant         int64
		wantErr         bool
	}

	tests := []test{
		{recordGiven: "test1", foundWant: true, recordWant: "test1", recordTypeWant: "A", recordValueWant: "8.8.8.8", ttlWant: 60, wantErr: false}, // ok record
		{recordGiven: "test2", foundWant: false, recordWant: "", recordTypeWant: "", recordValueWant: "", ttlWant: 0, wantErr: false},              // record dont exist

	}

	for _, tc := range tests {
		found, recordName, recordTypeReturned, recordValueReturned, ttlReturned, err := RetrieveRecordOnR53(recordRetriever, tc.recordGiven)
		if err != nil {
			if tc.wantErr == false {
				t.Errorf("didnt want err, got '%s'", err)
			}
		}

		if found != tc.foundWant {
			t.Errorf("didnt want err, got %t, wanted %t", found, tc.foundWant)
		}

		if strings.TrimRight(recordName, ".") != tc.recordWant {
			t.Errorf("didnt want err, got %s, wanted %s", recordName, tc.recordWant)
		}

		if recordTypeReturned != tc.recordTypeWant {
			t.Errorf("didnt want err, got %s, wanted %s", recordTypeReturned, tc.recordTypeWant)
		}

		if recordValueReturned != tc.recordValueWant {
			t.Errorf("didnt want err, got %s, wanted %s", recordValueReturned, tc.recordValueWant)
		}

		if ttlReturned != tc.ttlWant {
			t.Errorf("didnt want err, got %d, wanted %d", ttlReturned, tc.ttlWant)
		}
	}
}

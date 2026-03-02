package sources

import (
	"bytes"
	"testing"
)

func TestParseAWSRanges(t *testing.T) {
	data := `{
		"prefixes": [
			{"ip_prefix": "1.2.3.0/24", "region": "us-east-1", "service": "EC2"},
			{"ip_prefix": "5.6.7.0/24", "region": "eu-west-1", "service": "ROUTE53"}
		]
	}`
	r := bytes.NewReader([]byte(data))
	prefixes, err := ParseAWSRanges(r)
	if err != nil {
		t.Fatalf("ParseAWSRanges failed: %v", err)
	}

	if len(prefixes) != 2 {
		t.Errorf("Expected 2 prefixes, got %d", len(prefixes))
	}

	if prefixes[0].Region != "us-east-1" || prefixes[0].Service != "EC2" {
		t.Errorf("Unexpected prefix data: %+v", prefixes[0])
	}
}

func TestParseGoogleRanges(t *testing.T) {
	data := `{
		"prefixes": [
			{"ipv4Prefix": "8.8.8.0/24", "location": "us-east1"},
			{"ipv4Prefix": "35.192.0.0/12", "location": "us-central1"}
		]
	}`
	r := bytes.NewReader([]byte(data))
	prefixes, err := ParseGoogleRanges(r)
	if err != nil {
		t.Fatalf("ParseGoogleRanges failed: %v", err)
	}

	if len(prefixes) != 2 {
		t.Errorf("Expected 2 prefixes, got %d", len(prefixes))
	}

	if prefixes[0].Region != "us-east1" {
		t.Errorf("Unexpected region: %s", prefixes[0].Region)
	}
}

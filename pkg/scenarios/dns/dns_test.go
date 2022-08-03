package dns_test

import (
	"context"
	"testing"

	"github.com/hidracloud/hidra/v2/pkg/scenarios/dns"
)

func TestScenario(t *testing.T) {
	s := &dns.Scenario{}
	s.Init()

	params := map[string]string{
		"domain": "google.com",
	}

	ctx := context.TODO()

	_, err := s.RunStep(ctx, "whoisFrom", params, 0, false)
	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["for"] = "7d"

	_, err = s.RunStep(ctx, "shouldBeValidFor", params, 0, false)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["domain"] = "brutal.systems"

	_, err = s.RunStep(ctx, "dnsSecShouldBeValid", params, 0, false)

	if err != nil {
		t.Error(err)
	}
}

package whois_test

import (
	"context"
	"testing"

	"github.com/hidracloud/hidra/pkg/scenarios/whois"
)

func TestScenario(t *testing.T) {
	s := &whois.Scenario{}
	s.Init()

	params := map[string]string{
		"domain": "google.com",
	}

	ctx := context.TODO()

	_, err := s.RunStep(ctx, "whoisFrom", params, 0)
	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["for"] = "7d"

	_, err = s.RunStep(ctx, "shouldBeValidFor", params, 0)

	if err != nil {
		t.Error(err)
	}
}

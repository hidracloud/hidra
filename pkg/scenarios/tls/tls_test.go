package tls_test

import (
	"context"
	"testing"

	"github.com/hidracloud/hidra/v2/pkg/scenarios/tls"
)

func TestScenario(t *testing.T) {
	// Initialize scenario
	s := &tls.Scenario{}
	s.Init()

	ctx := context.TODO()

	params := make(map[string]string)
	params["to"] = "google.com:443"
	_, err := s.RunStep(ctx, "connectTo", params, 0, false)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["dns"] = "google.com"

	_, err = s.RunStep(ctx, "dnsShouldBePresent", params, 0, false)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["for"] = "7d"

	_, err = s.RunStep(ctx, "shouldBeValidFor", params, 0, false)

	if err != nil {
		t.Error(err)
	}
}

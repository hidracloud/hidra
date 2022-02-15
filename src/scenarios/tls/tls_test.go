package tls_test

import (
	"testing"

	"github.com/hidracloud/hidra/src/scenarios/tls"
)

func TestScenario(t *testing.T) {
	// Initialize scenario
	s := &tls.Scenario{}
	s.Init()

	params := make(map[string]string)
	params["to"] = "google.com:443"
	_, err := s.RunStep("connectTo", params)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["dns"] = "google.com"

	_, err = s.RunStep("dnsShouldBePresent", params)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["for"] = "7d"

	_, err = s.RunStep("shouldBeValidFor", params)

	if err != nil {
		t.Error(err)
	}
}

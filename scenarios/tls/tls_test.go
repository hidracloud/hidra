package tls_test

import (
	"testing"

	"github.com/hidracloud/hidra/scenarios/tls"
)

func TestTLSScenario(t *testing.T) {
	// Initialize scenario
	s := &tls.TLSScneario{}
	s.Init()

	params := make(map[string]string)
	params["to"] = "google.com:443"
	err := s.RunStep("connectTo", params)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["dns"] = "google.com"

	err = s.RunStep("dnsShouldBePresent", params)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["for"] = "7d"

	err = s.RunStep("shouldBeValidFor", params)

	if err != nil {
		t.Error(err)
	}
}

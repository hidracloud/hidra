package whois_test

import (
	"testing"

	"github.com/hidracloud/hidra/src/scenarios/whois"
)

func TestScenario(t *testing.T) {
	s := &whois.Scenario{}
	s.Init()

	params := map[string]string{
		"domain": "latostadora.com",
	}

	_, err := s.RunStep("whoisFrom", params, 0)
	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["for"] = "7d"

	_, err = s.RunStep("shouldBeValidFor", params, 0)

	if err != nil {
		t.Error(err)
	}
}

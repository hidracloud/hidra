package udp_test

import (
	"testing"

	"github.com/hidracloud/hidra/src/scenarios/udp"
)

func TestScenario(t *testing.T) {
	// Initialize scenario
	s := &udp.Scenario{}
	s.Init()

	params := make(map[string]string)
	params["to"] = "8.8.8.8:53"

	_, err := s.RunStep("connectTo", params, 0)

	if err != nil {
		t.Error(err)
	}

}

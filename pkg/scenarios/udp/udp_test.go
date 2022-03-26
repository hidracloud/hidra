package udp_test

import (
	"context"
	"testing"

	"github.com/hidracloud/hidra/pkg/scenarios/udp"
)

func TestScenario(t *testing.T) {
	// Initialize scenario
	s := &udp.Scenario{}
	s.Init()

	ctx := context.TODO()

	params := make(map[string]string)
	params["to"] = "8.8.8.8:53"

	_, err := s.RunStep(ctx, "connectTo", params, 0)

	if err != nil {
		t.Error(err)
	}

}

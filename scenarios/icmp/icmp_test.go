package icmp_test

import (
	"testing"

	"github.com/hidracloud/hidra/scenarios/icmp"
)

func TestPing(t *testing.T) {
	// Initialize scenario
	h := &icmp.IcmpScenario{}
	h.Init()

	params := make(map[string]string)
	params["hostname"] = "8.8.8.8"

	_, err := h.RunStep("ping", params)

	if err != nil {
		t.Errorf("TestPing failed: %s", err)
	}
}

func TestTraceroute(t *testing.T) {
	// Initialize scenario
	h := &icmp.IcmpScenario{}
	h.Init()

	params := make(map[string]string)
	params["hostname"] = "8.8.8.8"

	_, err := h.RunStep("traceroute", params)

	if err != nil {
		t.Errorf("TestPing failed: %s", err)
	}
}

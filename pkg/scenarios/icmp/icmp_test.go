package icmp_test

import (
	"context"
	"testing"
	"time"

	"github.com/hidracloud/hidra/pkg/scenarios/icmp"
)

func TestPing(t *testing.T) {
	// Initialize scenario
	h := &icmp.Scenario{}
	h.Init()

	ctx := context.TODO()

	params := make(map[string]string)
	params["hostname"] = "8.8.8.8"

	_, err := h.RunStep(ctx, "ping", params, time.Second*60)

	if err != nil {
		t.Errorf("TestPing failed: %s", err)
	}
}

func TestTraceroute(t *testing.T) {
	// Initialize scenario
	h := &icmp.Scenario{}
	h.Init()

	ctx := context.TODO()

	params := make(map[string]string)
	params["hostname"] = "8.8.8.8"

	_, err := h.RunStep(ctx, "traceroute", params, time.Second*60)

	if err != nil {
		t.Errorf("TestPing failed: %s", err)
	}
}

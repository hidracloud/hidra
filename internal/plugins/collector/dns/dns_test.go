package dns_test

import (
	"context"
	"testing"

	"github.com/hidracloud/hidra/v3/internal/plugins"
	"github.com/hidracloud/hidra/v3/internal/plugins/collector/dns"
)

// TestRequestByMethod
func TestHTTPRequestParameters(t *testing.T) {
	h := dns.DNS{}
	h.Init()

	ctx := context.TODO()
	previous := make(map[string]any, 0)

	_, err := h.RunStep(ctx, previous, &plugins.Step{
		Name: "whoisFrom",
		Args: map[string]string{
			"domain": "google.com",
		},
	})

	if err != nil {
		t.Error(err)
	}

	_, err = h.RunStep(ctx, previous, &plugins.Step{
		Name: "shouldBeValidFor",
		Args: map[string]string{
			"for": "7d",
		},
	})

	if err != nil {
		t.Error(err)
	}

	_, err = h.RunStep(ctx, previous, &plugins.Step{
		Name: "askRegisterToNS",
		Args: map[string]string{
			"ns":   "10.0.0.0",
			"type": "a",
			"host": "testss",
		},
	})

	if err == nil {
		t.Error("Should return error")
	}
}

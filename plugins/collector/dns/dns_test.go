package dns_test

import (
	"context"
	"testing"

	"github.com/hidracloud/hidra/v3/plugins"
	"github.com/hidracloud/hidra/v3/plugins/collector/dns"
)

// TestRequestByMethod
func TestHTTPRequestParameters(t *testing.T) {
	h := dns.DNS{}
	h.Init()

	ctx := context.TODO()

	ctx, _, err := h.RunStep(ctx, &plugins.Step{
		Name: "whoisFrom",
		Args: map[string]string{
			"domain": "google.com",
		},
	})

	if err != nil {
		t.Error(err)
	}

	_, _, err = h.RunStep(ctx, &plugins.Step{
		Name: "shouldBeValidFor",
		Args: map[string]string{
			"for": "7d",
		},
	})

	if err != nil {
		t.Error(err)
	}
}

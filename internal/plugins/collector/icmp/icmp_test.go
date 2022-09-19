package icmp_test

import (
	"context"
	"testing"

	"github.com/hidracloud/hidra/v3/internal/plugins"
	"github.com/hidracloud/hidra/v3/internal/plugins/collector/icmp"
)

// TestRequestByMethod
func TestHTTPRequestParameters(t *testing.T) {
	h := icmp.ICMP{}
	h.Init()

	ctx := context.TODO()

	_, _, err := h.RunStep(ctx, &plugins.Step{
		Name: "ping",
		Args: map[string]string{
			"hostname": "8.8.8.8",
		},
	})

	if err != nil {
		t.Error(err)
	}
}

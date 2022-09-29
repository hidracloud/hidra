package tls_test

import (
	"context"
	"testing"

	"github.com/hidracloud/hidra/v3/internal/plugins"
	"github.com/hidracloud/hidra/v3/internal/plugins/collector/tls"
)

func TestScenario(t *testing.T) {
	// Initialize scenario
	h := &tls.TLS{}
	h.Init()

	ctx := context.TODO()

	previous := make(map[string]any, 0)

	_, err := h.RunStep(ctx, previous, &plugins.Step{
		Name: "connectTo",
		Args: map[string]string{
			"to": "google.com:443",
		},
	})

	if err != nil {
		t.Error(err)
	}

	_, err = h.RunStep(ctx, previous, &plugins.Step{
		Name: "dnsShouldBePresent",
		Args: map[string]string{
			"dns": "google.com",
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
}

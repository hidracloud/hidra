package tls_test

import (
	"context"
	"testing"

	"github.com/hidracloud/hidra/v3/plugins"
	"github.com/hidracloud/hidra/v3/plugins/services/tls"
)

func TestScenario(t *testing.T) {
	// Initialize scenario
	h := &tls.TLS{}
	h.Init()

	ctx := context.TODO()

	ctx, _, err := h.RunStep(ctx, &plugins.Step{
		Name: "connectTo",
		Args: map[string]string{
			"to": "google.com:443",
		},
	})

	if err != nil {
		t.Error(err)
	}

	ctx, _, err = h.RunStep(ctx, &plugins.Step{
		Name: "dnsShouldBePresent",
		Args: map[string]string{
			"dns": "google.com",
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

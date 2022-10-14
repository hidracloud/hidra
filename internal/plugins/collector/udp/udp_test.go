package udp_test

import (
	"context"
	"testing"

	"github.com/hidracloud/hidra/v3/internal/plugins"
	"github.com/hidracloud/hidra/v3/internal/plugins/collector/udp"
)

func TestScenario(t *testing.T) {
	h := &udp.UDP{}
	h.Init()

	ctx := context.TODO()
	previous := make(map[string]any, 0)

	_, err := h.RunStep(ctx, previous, &plugins.Step{
		Name: "connectTo",
		Args: map[string]string{
			"to": "8.8.8.8:53",
		},
	})

	if err != nil {
		t.Error(err)
	}
}

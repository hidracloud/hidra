package dummy

import (
	"context"

	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/internal/plugins"
)

// Dummy represents a dummy plugin.
type Dummy struct {
	plugins.BasePlugin
}

// doNothing does nothing.
func (p *Dummy) doNothing(ctx context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	return nil, nil
}

// Init initializes the plugin.
func (p *Dummy) Init() {
	p.Primitives()

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "doNothing",
		Description: "Yes, it does nothing.",
		Fn:          p.doNothing,
		Params:      []plugins.StepParam{},
	})

}

// Init initializes the plugin.
func init() {
	h := &Dummy{}
	h.Init()
	plugins.AddPlugin("dummy", "Dummy plugin is used to test features", h)
}

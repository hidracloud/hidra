package str

import (
	"context"
	"fmt"

	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/internal/misc"
	"github.com/hidracloud/hidra/v3/internal/plugins"
	"github.com/hidracloud/hidra/v3/internal/utils"
)

// HTTP represents a HTTP plugin.
type Strings struct {
	plugins.BasePlugin
}

// outputShouldContain returns true if the output contains the expected value.
func (p *Strings) outputShouldContain(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	search := args["search"]

	if _, ok := stepsgen[misc.ContextOutput].([]byte); !ok {
		return nil, fmt.Errorf("output is not a string")
	}

	output := stepsgen[misc.ContextOutput].([]byte)

	if utils.BytesContainsString(output, search) {
		return []*metrics.Metric{}, nil
	}

	return []*metrics.Metric{}, fmt.Errorf("output does not contain %s", search)
}

// Init initializes the plugin.
func (p *Strings) Init() {
	p.Primitives()

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "outputShouldContain",
		Description: "Checks if the output contains a string",
		Params: []plugins.StepParam{
			{
				Name:        "search",
				Description: "The string to search",
				Optional:    false,
			},
		},
		Fn: p.outputShouldContain,
	})

}

// Init initializes the plugin.
func init() {
	h := &Strings{}
	h.Init()
	plugins.AddPlugin("string", "String plugin is used to check strings", h)
}

package str

import (
	"context"
	"fmt"
	"strings"

	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/plugins"
)

// HTTP represents a HTTP plugin.
type Strings struct {
	plugins.BasePlugin
}

// outputShouldContain returns true if the output contains the expected value.
func (p *Strings) outputShouldContain(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	search := args["search"]

	if _, ok := ctx.Value(plugins.ContextOutput).(string); !ok {
		return ctx, nil, fmt.Errorf("output is not a string")
	}

	output := ctx.Value(plugins.ContextOutput).(string)

	if strings.Contains(output, search) {
		return ctx, []*metrics.Metric{}, nil
	}

	return ctx, []*metrics.Metric{}, fmt.Errorf("output does not contain %s", search)
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
	plugins.AddPlugin("string", h)
}

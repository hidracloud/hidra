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
func (p *Strings) outputShouldContain(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	search := args["search"]

	if _, ok := ctx.Value(misc.ContextOutput).([]byte); !ok {
		return ctx, nil, fmt.Errorf("output is not a string")
	}

	output := ctx.Value(misc.ContextOutput).([]byte)

	if utils.BytesContainsString(output, search) {
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

package str

import (
	"context"
	"fmt"
	"strconv"

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

	var err error
	times := 1

	if args["times"] != "" {
		times, err = strconv.Atoi(args["times"])

		if err != nil {
			return nil, err
		}
	}

	if times == 1 {
		if utils.BytesContainsString(output, search) {
			return []*metrics.Metric{}, nil
		}

		return []*metrics.Metric{}, fmt.Errorf("output does not contain %s", search)
	}

	appear := utils.BytesContainsStringTimes(output, search)

	if times <= appear {
		return []*metrics.Metric{}, nil
	}

	return []*metrics.Metric{}, fmt.Errorf("output does not contain %s %d times, appear %d", search, times, appear)
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
			{
				Name:        "times",
				Description: "The number of times the string should appear",
				Optional:    true,
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

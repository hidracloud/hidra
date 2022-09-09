package plugins

import (
	"context"
	"fmt"

	"github.com/hidracloud/hidra/v3/internal/metrics"
)

// PluginInterface represents a plugin.
type PluginInterface interface {
	// Init initializes the plugin.
	Init()
	// Primitives returns the plugin primitives.
	Primitives()
	// RunStep runs a step.
	RunStep(context.Context, *Step) (context.Context, []*metrics.Metric, error)
	// RegisterStep registers a step.
	RegisterStep(*StepDefinition)
}

// Step represents a step.
type Step struct {
	// Name is the name of the step.
	Name string
	// Args is the arguments of the step.
	Args map[string]string
	// Timeout is the timeout of the step.
	Timeout int
}

type stepFn func(context.Context, map[string]string) (context.Context, []*metrics.Metric, error)

// StepParam returns the value of a step parameter
type StepParam struct {
	Name        string
	Description string
	Optional    bool
}

// StepDefinition represents a step definition.
type StepDefinition struct {
	Name             string
	Description      string
	Params           []StepParam
	Fn               stepFn `json:"-"`
	ContextGenerator []ContextGenerator
}

// ContextGenerator represents a context generator.
type ContextGenerator struct {
	// Name is the name of the context generator.
	Name string
	// Description is the description of the context generator.
	Description string
}

// BasePlugin represents a base plugin.
type BasePlugin struct {
	StepDefinitions map[string]*StepDefinition
}

// Primitives initializes the plugin.
func (p *BasePlugin) Primitives() {
	p.StepDefinitions = map[string]*StepDefinition{}
}

// RunStep runs a step.
func (p *BasePlugin) RunStep(ctx context.Context, step *Step) (context.Context, []*metrics.Metric, error) {
	// get step definition
	stepDefinition, ok := p.StepDefinitions[step.Name]

	if !ok && step.Name == "onFailure" {
		return ctx, nil, ctx.Value(LastError).(error)
	}

	if !ok {
		return ctx, nil, fmt.Errorf("step %s not found", step.Name)
	}

	// validate step arguments
	for _, param := range stepDefinition.Params {
		if _, ok := step.Args[param.Name]; !ok && !param.Optional {
			return ctx, nil, fmt.Errorf("missing argument %s", param.Name)
		}
	}

	// run step
	ctx, metrics, err := stepDefinition.Fn(ctx, step.Args)

	if err != nil {
		ctx = context.WithValue(ctx, LastError, err)
		step.Name = "onFailure"

		ctx, metrics, _ = p.RunStep(ctx, step)

		return ctx, metrics, err
	}

	return ctx, metrics, nil
}

// RegisterStep registers a step.
func (p *BasePlugin) RegisterStep(step *StepDefinition) {
	p.StepDefinitions[step.Name] = step
}

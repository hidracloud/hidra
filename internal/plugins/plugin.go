package plugins

import (
	"context"
	"fmt"

	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/internal/misc"
)

// PluginInterface represents a plugin.
type PluginInterface interface {
	// Init initializes the plugin.
	Init()
	// Primitives returns the plugin primitives.
	Primitives()
	// RunStep runs a step.
	RunStep(context.Context, map[string]any, *Step) ([]*metrics.Metric, error)
	// RegisterStep registers a step.
	RegisterStep(*StepDefinition)
	// StepExists returns true if the step exists.
	StepExists(string) bool
	// GetSteps returns all steps.
	GetSteps() map[string]*StepDefinition
}

// Step represents a step.
type Step struct {
	// Name is the name of the step.
	Name string
	// Args is the arguments of the step.
	Args map[string]string
	// Timeout is the timeout of the step.
	Timeout int
	// Negate is true if the step should be negated.
	Negate bool
}

type stepFn func(context.Context, map[string]string, map[string]any) ([]*metrics.Metric, error)

// StepParam returns the value of a step parameter
type StepParam struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Optional    bool   `json:"optional"`
}

// StepDefinition represents a step definition.
type StepDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Params      []StepParam `json:"params"`
	Fn          stepFn      `json:"-"`
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
func (p *BasePlugin) RunStep(ctx context.Context, stepsgen map[string]any, step *Step) ([]*metrics.Metric, error) {
	// get step definition
	stepDefinition, ok := p.StepDefinitions[step.Name]

	if !ok && step.Name == "onFailure" {
		return nil, stepsgen[misc.ContextLastError].(error)
	}

	if !ok {
		return nil, fmt.Errorf("step %s not found", step.Name)
	}

	// validate step arguments
	for _, param := range stepDefinition.Params {
		if _, ok := step.Args[param.Name]; !ok && !param.Optional {
			return nil, fmt.Errorf("missing argument %s", param.Name)
		}
	}

	// run step
	metrics, err := stepDefinition.Fn(ctx, step.Args, stepsgen)

	if err != nil && step.Negate {
		return metrics, nil
	} else if err == nil && step.Negate {
		return metrics, fmt.Errorf("step %s should have failed", step.Name)
	}

	if err != nil {
		stepsgen[misc.ContextLastError] = err
		step.Name = "onFailure"

		metrics, _ = p.RunStep(ctx, stepsgen, step)

		return metrics, err
	}

	return metrics, nil
}

// StepExists returns true if the step exists.
func (p *BasePlugin) StepExists(name string) bool {
	_, ok := p.StepDefinitions[name]

	return ok
}

// GetSteps returns all steps.
func (p *BasePlugin) GetSteps() map[string]*StepDefinition {
	return p.StepDefinitions
}

// RegisterStep registers a step.
func (p *BasePlugin) RegisterStep(step *StepDefinition) {
	p.StepDefinitions[step.Name] = step
}

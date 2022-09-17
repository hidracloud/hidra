package browser_test

import (
	"context"
	"testing"

	"github.com/hidracloud/hidra/v3/plugins"
	"github.com/hidracloud/hidra/v3/plugins/services/browser"
)

func TestScenario(t *testing.T) {
	h := &browser.Browser{}
	h.Init()

	ctx := context.TODO()

	ctx, _, err := h.RunStep(ctx, &plugins.Step{
		Name: "navigateTo",
		Args: map[string]string{
			"url": "https://hidra.cloud",
		},
	})

	if err != nil {
		t.Error(err)
	}

	ctx, _, err = h.RunStep(ctx, &plugins.Step{
		Name: "urlShouldBe",
		Args: map[string]string{
			"url": "https://hidra.cloud/",
		},
	})

	if err != nil {
		t.Error(err)
	}

	ctx, _, err = h.RunStep(ctx, &plugins.Step{
		Name: "textShouldBe",
		Args: map[string]string{
			"selector": "#td-cover-block-0 > div > div > div > div > h1",
			"text":     "Welcome to Hidra! Your new solution for monitoring",
		},
	})

	if err != nil {
		t.Error(err)
	}

	ctx, _, err = h.RunStep(ctx, &plugins.Step{
		Name: "waitVisible",
		Args: map[string]string{
			"selector": "#td-cover-block-0 > div > div > div > div > div > div > a.btn.btn-lg.btn-primary.mr-3.mb-4",
		},
	})

	if err != nil {
		t.Error(err)
	}

	ctx, _, err = h.RunStep(ctx, &plugins.Step{
		Name: "click",
		Args: map[string]string{
			"selector": "#td-cover-block-0 > div > div > div > div > div > div > a.btn.btn-lg.btn-primary.mr-3.mb-4",
		},
	})

	if err != nil {
		t.Error(err)
	}

	_, _, err = h.RunStep(ctx, &plugins.Step{
		Name: "urlShouldBe",
		Args: map[string]string{
			"url": "https://hidra.cloud/docs/",
		},
	})

	if err != nil {
		t.Error(err)
	}

	_, _, err = h.RunStep(ctx, &plugins.Step{
		Name: "onClose",
	})

	if err != nil {
		t.Error(err)
	}

}

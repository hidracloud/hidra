package browser

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/internal/misc"
	"github.com/hidracloud/hidra/v3/internal/plugins"

	"github.com/chromedp/cdproto/performance"
	"github.com/chromedp/chromedp"
)

var (
	errPluginNotInitialized = errors.New("plugin not initialized")
)

// Browser represents a Browser plugin.
type Browser struct {
	plugins.BasePlugin
}

// navigateTo implements the browser.navigateTo primitive.
func (p *Browser) navigateTo(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	if _, ok := ctx.Value(misc.ContextBrowserChromedpCtx).(context.Context); !ok {
		timeout := 120 * time.Second

		if _, ok := ctx.Value(misc.ContextTimeout).(time.Duration); ok {
			timeout = ctx.Value(misc.ContextTimeout).(time.Duration)
		}

		// initialize chromedp
		initialCtx := context.Background()

		if os.Getenv("BROWSER_NO_HEADLESS") != "" {
			initialCtx, _ = chromedp.NewExecAllocator(initialCtx, append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", false))...)
		}

		chromedpCtx, cancel := chromedp.NewContext(initialCtx)
		chromedpCtx, cancelTimeout := context.WithTimeout(chromedpCtx, timeout)

		ctx = context.WithValue(ctx, misc.ContextBrowserChromedpCtx, chromedpCtx)
		ctx = context.WithValue(ctx, misc.ContextBrowserChromedpCancel, cancel)
		ctx = context.WithValue(ctx, misc.ContextBrowserChromedpCancelTimeout, cancelTimeout)
	}

	chromedpCtx := ctx.Value(misc.ContextBrowserChromedpCtx).(context.Context)

	err := chromedp.Run(chromedpCtx, performance.Enable(), chromedp.Navigate(args["url"]))

	return ctx, nil, err
}

func selector2By(selector string) func(*chromedp.Selector) {
	opt := chromedp.BySearch

	if selector != "" {
		switch selector {
		case "bySearch":
			opt = chromedp.BySearch
		case "byID":
			opt = chromedp.ByID
		case "byQuery":
			opt = chromedp.ByQuery
		}
	}

	return opt
}

// urlShouldBe implements the browser.urlShouldBe primitive.
func (p *Browser) urlShouldBe(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	if _, ok := ctx.Value(misc.ContextBrowserChromedpCtx).(context.Context); !ok {
		return ctx, nil, errPluginNotInitialized
	}

	chromedpCtx := ctx.Value(misc.ContextBrowserChromedpCtx).(context.Context)

	var url string

	err := chromedp.Run(chromedpCtx, chromedp.Location(&url))

	if err != nil {
		return ctx, nil, err
	}

	if url != args["url"] {
		return ctx, nil, fmt.Errorf("url is not %s is %s", args["url"], url)
	}

	return ctx, nil, nil
}

// textShouldBe implements the browser.textShouldBe primitive.
func (p *Browser) textShouldBe(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	if _, ok := ctx.Value(misc.ContextBrowserChromedpCtx).(context.Context); !ok {
		return ctx, nil, errPluginNotInitialized
	}

	chromedpCtx := ctx.Value(misc.ContextBrowserChromedpCtx).(context.Context)

	var text string

	err := chromedp.Run(chromedpCtx, chromedp.Text(args["selector"], &text, selector2By(args["selectorBy"])))

	if err != nil {
		return ctx, nil, err
	}

	if text != args["text"] {
		return ctx, nil, fmt.Errorf("text is not %s is %s", args["text"], text)
	}

	return ctx, nil, nil
}

// sendKeys implements the browser.sendKeys primitive.
func (p *Browser) sendKeys(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	if _, ok := ctx.Value(misc.ContextBrowserChromedpCtx).(context.Context); !ok {
		return ctx, nil, errPluginNotInitialized
	}

	chromedpCtx := ctx.Value(misc.ContextBrowserChromedpCtx).(context.Context)

	err := chromedp.Run(
		chromedpCtx,
		chromedp.SendKeys(args["selector"], args["keys"], selector2By(args["selectorBy"])),
	)

	return ctx, nil, err
}

// waitVisible implements the browser.waitVisible primitive.
func (p *Browser) waitVisible(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	if _, ok := ctx.Value(misc.ContextBrowserChromedpCtx).(context.Context); !ok {
		return ctx, nil, errPluginNotInitialized
	}

	chromedpCtx := ctx.Value(misc.ContextBrowserChromedpCtx).(context.Context)

	err := chromedp.Run(chromedpCtx, chromedp.WaitVisible(args["selector"], selector2By(args["selectorBy"])))

	return ctx, nil, err
}

// onClose implements the browser.onClose primitive.
func (p *Browser) onClose(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	var err error
	if _, ok := ctx.Value(misc.ContextBrowserChromedpCtx).(context.Context); ok {
		chromedpCtx := ctx.Value(misc.ContextBrowserChromedpCtx).(context.Context)

		err = chromedp.Stop().Do(chromedpCtx)

		ctx.Value(misc.ContextBrowserChromedpCancel).(context.CancelFunc)()
		ctx.Value(misc.ContextBrowserChromedpCancelTimeout).(context.CancelFunc)()
	}

	return ctx, nil, err
}

// click implements the browser.click primitive.
func (p *Browser) click(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	if _, ok := ctx.Value(misc.ContextBrowserChromedpCtx).(context.Context); !ok {
		return ctx, nil, errPluginNotInitialized
	}

	chromedpCtx := ctx.Value(misc.ContextBrowserChromedpCtx).(context.Context)

	err := chromedp.Run(chromedpCtx, chromedp.Click(args["selector"], selector2By(args["selectorBy"])))

	return ctx, nil, err
}

// wait implements the browser.wait primitive.
func (p *Browser) wait(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	duration, err := time.ParseDuration(args["duration"])

	if err != nil {
		return ctx, nil, err
	}

	time.Sleep(duration)

	return ctx, nil, nil
}

// Init initializes the plugin.
func (p *Browser) Init() {
	p.Primitives()

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "navigateTo",
		Description: "Navigates to a URL",
		Params: []plugins.StepParam{
			{
				Name:        "url",
				Description: "URL to navigate to",
				Optional:    false,
			},
		},
		Fn: p.navigateTo,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "urlShouldBe",
		Description: "Checks if the current URL is the expected one",
		Params: []plugins.StepParam{
			{
				Name:        "url",
				Description: "Expected URL",
				Optional:    false,
			},
		},
		Fn: p.urlShouldBe,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "textShouldBe",
		Description: "Checks if the text of an element is the expected one",
		Params: []plugins.StepParam{
			{
				Name:        "selector",
				Description: "Selector of the element",
				Optional:    false,
			},
			{
				Name:        "selectorBy",
				Description: "Selector type",
				Optional:    true,
			},
			{
				Name:        "text",
				Description: "Expected text",
				Optional:    false,
			},
		},
		Fn: p.textShouldBe,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "sendKeys",
		Description: "Sends keys to an element",
		Params: []plugins.StepParam{
			{
				Name:        "selector",
				Description: "Selector of the element",
				Optional:    false,
			},
			{
				Name:        "selectorBy",
				Description: "Selector type",
				Optional:    true,
			},
			{
				Name:        "keys",
				Description: "Keys to send",
				Optional:    false,
			},
		},
		Fn: p.sendKeys,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "waitVisible",
		Description: "Waits for an element to be visible",
		Params: []plugins.StepParam{
			{
				Name:        "selector",
				Description: "Selector of the element",
				Optional:    false,
			},
			{
				Name:        "selectorBy",
				Description: "Selector type",
				Optional:    true,
			},
		},
		Fn: p.waitVisible,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "click",
		Description: "Clicks on an element",
		Params: []plugins.StepParam{
			{
				Name:        "selector",
				Description: "Selector of the element",
				Optional:    false,
			},
			{
				Name:        "selectorBy",
				Description: "Selector type",
				Optional:    true,
			},
		},
		Fn: p.click,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "wait",
		Description: "Waits for a duration",
		Params: []plugins.StepParam{
			{
				Name:        "duration",
				Description: "Duration to wait",
				Optional:    false,
			},
		},
		Fn: p.wait,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "onClose",
		Description: "Close the connection",
		Params:      []plugins.StepParam{},
		Fn:          p.onClose,
	})
}

// Init initializes the plugin.
func init() {
	h := &Browser{}
	h.Init()
	plugins.AddPlugin("browser", h)
}

package browser

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/internal/misc"
	"github.com/hidracloud/hidra/v3/internal/plugins"
	"github.com/hidracloud/hidra/v3/internal/utils"
	log "github.com/sirupsen/logrus"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/performance"
	"github.com/chromedp/chromedp"
)

var (
	errPluginNotInitialized = errors.New("plugin not initialized")
	// selectorDesc describes the selector primitive.
	selectorDesc = "Selector of the element"
	// selectorTypeDesc describes the selectorType primitive.
	selectorTypeDesc = "Selector type"
)

// Browser represents a Browser plugin.
type Browser struct {
	plugins.BasePlugin
}

// RequestInfo represents a request info.
type RequestInfo struct {
	// Type is the type of the request.
	Type string
	// URL is the URL of the request.
	URL string
	// Timestamp is the timestamp of the request.
	Timestamp time.Time
	// ResponseReceivedTimestamp is the timestamp of the response received.
	ResponseReceivedTimestamp time.Time
	// ResponseFinishedTimestamp is the timestamp of the response finished.
	ResponseFinishedTimestamp time.Time
}

// navigateTo implements the browser.navigateTo primitive.
func (p *Browser) navigateTo(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	if _, ok := stepsgen[misc.ContextBrowserChromedpCtx].(context.Context); !ok {
		// initialize chromedp
		initialCtx := context.Background()

		if os.Getenv("BROWSER_NO_HEADLESS") != "" {
			initialCtx, _ = chromedp.NewExecAllocator(initialCtx, append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", false))...)
		}

		chromedpCtx, cancel := chromedp.NewContext(initialCtx)

		stepsgen[misc.ContextBrowserChromedpCtx] = chromedpCtx
		stepsgen[misc.ContextBrowserChromedpCancel] = cancel
	}

	chromedpCtx := stepsgen[misc.ContextBrowserChromedpCtx].(context.Context)

	timeout := 30 * time.Second

	if _, ok := stepsgen[misc.ContextTimeout].(time.Duration); ok {
		timeout = stepsgen[misc.ContextTimeout].(time.Duration)
	}

	ackCtx, _ := context.WithTimeout(chromedpCtx, timeout) //nolint:all
	customMetrics := make([]*metrics.Metric, 0)

	requestsInfo := make(map[string]*RequestInfo)

	chromedp.ListenTarget(
		ackCtx,
		func(ev interface{}) {
			if ev, ok := ev.(*network.EventRequestWillBeSent); ok {
				requestsInfo[string(ev.RequestID)] = &RequestInfo{
					Type:      string(ev.Type),
					Timestamp: ev.Timestamp.Time(),
				}
			}

			if ev, ok := ev.(*network.EventResponseReceived); ok {
				if requestInfo, ok := requestsInfo[string(ev.RequestID)]; ok {
					requestInfo.ResponseReceivedTimestamp = ev.Timestamp.Time()
					requestInfo.URL = ev.Response.URL
				}
			}

			if ev, ok := ev.(*network.EventLoadingFinished); ok {
				if requestInfo, ok := requestsInfo[string(ev.RequestID)]; ok {
					requestInfo.ResponseFinishedTimestamp = ev.Timestamp.Time()
				}
			}
		},
	)

	err := chromedp.Run(ackCtx, performance.Enable(), chromedp.ActionFunc(func(cxt context.Context) error {
		perfMetrics, err := performance.GetMetrics().Do(cxt)

		if err != nil {
			return err
		}

		for _, metric := range perfMetrics {
			customMetrics = append(customMetrics, &metrics.Metric{
				Name:        utils.CamelCaseToSnakeCase(metric.Name),
				Description: fmt.Sprintf("Performance metric %s", metric.Name),
				Labels: map[string]string{
					"url": args["url"],
				},
				Value: metric.Value,
				Purge: false,
			})
		}

		return nil
	}), chromedp.Navigate(args["url"]))

	if err != nil {
		return nil, err
	}

	for _, requestInfo := range requestsInfo {
		customMetrics = append(customMetrics, &metrics.Metric{
			Name:        "request_part_time",
			Description: fmt.Sprintf("Request time for %s", requestInfo.Type),
			Labels: map[string]string{
				"part_url": requestInfo.URL,
				"type":     strings.ToLower(requestInfo.Type),
			},
			Value: float64(requestInfo.ResponseFinishedTimestamp.Sub(requestInfo.Timestamp).Milliseconds()),
			Purge: true,
		})
	}

	return customMetrics, err
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
func (p *Browser) urlShouldBe(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	if _, ok := stepsgen[misc.ContextBrowserChromedpCtx].(context.Context); !ok {
		return nil, errPluginNotInitialized
	}

	chromedpCtx := stepsgen[misc.ContextBrowserChromedpCtx].(context.Context)

	var url string

	timeout := 30 * time.Second

	if _, ok := stepsgen[misc.ContextTimeout].(time.Duration); ok {
		timeout = stepsgen[misc.ContextTimeout].(time.Duration)
	}

	ackCtx, _ := context.WithTimeout(chromedpCtx, timeout) //nolint:all

	err := chromedp.Run(ackCtx, chromedp.Location(&url))

	if err != nil {
		return nil, err
	}

	if url != args["url"] {
		return nil, fmt.Errorf("url is not %s is %s", args["url"], url)
	}

	return nil, nil
}

// setViewPort implements the browser.setViewPort primitive.
func (p *Browser) setViewPort(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	if _, ok := stepsgen[misc.ContextBrowserChromedpCtx].(context.Context); !ok {
		return nil, errPluginNotInitialized
	}

	chromedpCtx := stepsgen[misc.ContextBrowserChromedpCtx].(context.Context)

	timeout := 30 * time.Second

	if _, ok := stepsgen[misc.ContextTimeout].(time.Duration); ok {
		timeout = stepsgen[misc.ContextTimeout].(time.Duration)
	}

	ackCtx, _ := context.WithTimeout(chromedpCtx, timeout) //nolint:all

	err := chromedp.Run(
		ackCtx,
		chromedp.EmulateViewport(
			int64(utils.StringToInt(args["width"])),
			int64(utils.StringToInt(args["height"])),
		),
	)

	return nil, err
}

// textShouldBe implements the browser.textShouldBe primitive.
func (p *Browser) textShouldBe(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	if _, ok := stepsgen[misc.ContextBrowserChromedpCtx].(context.Context); !ok {
		return nil, errPluginNotInitialized
	}

	chromedpCtx := stepsgen[misc.ContextBrowserChromedpCtx].(context.Context)

	var text string

	timeout := 30 * time.Second

	if _, ok := stepsgen[misc.ContextTimeout].(time.Duration); ok {
		timeout = stepsgen[misc.ContextTimeout].(time.Duration)
	}

	ackCtx, _ := context.WithTimeout(chromedpCtx, timeout) //nolint:all

	err := chromedp.Run(ackCtx, chromedp.Text(args["selector"], &text, selector2By(args["selectorBy"])))

	if err != nil {
		return nil, err
	}

	if text != args["text"] {
		return nil, fmt.Errorf("text is not %s is %s", args["text"], text)
	}

	return nil, nil
}

// sendKeys implements the browser.sendKeys primitive.
func (p *Browser) sendKeys(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	if _, ok := stepsgen[misc.ContextBrowserChromedpCtx].(context.Context); !ok {
		return nil, errPluginNotInitialized
	}

	chromedpCtx := stepsgen[misc.ContextBrowserChromedpCtx].(context.Context)

	timeout := 30 * time.Second

	if _, ok := stepsgen[misc.ContextTimeout].(time.Duration); ok {
		timeout = stepsgen[misc.ContextTimeout].(time.Duration)
	}

	ackCtx, _ := context.WithTimeout(chromedpCtx, timeout) //nolint:all

	err := chromedp.Run(
		ackCtx,
		chromedp.SendKeys(args["selector"], args["keys"], selector2By(args["selectorBy"])),
	)

	return nil, err
}

// waitVisible implements the browser.waitVisible primitive.
func (p *Browser) waitVisible(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	if _, ok := stepsgen[misc.ContextBrowserChromedpCtx].(context.Context); !ok {
		return nil, errPluginNotInitialized
	}

	chromedpCtx := stepsgen[misc.ContextBrowserChromedpCtx].(context.Context)

	timeout := 30 * time.Second

	if _, ok := stepsgen[misc.ContextTimeout].(time.Duration); ok {
		timeout = stepsgen[misc.ContextTimeout].(time.Duration)
	}

	ackCtx, _ := context.WithTimeout(chromedpCtx, timeout) //nolint:all

	err := chromedp.Run(ackCtx, chromedp.WaitVisible(args["selector"], selector2By(args["selectorBy"])))

	return nil, err
}

// onClose implements the browser.onClose primitive.
func (p *Browser) onClose(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	var err error

	if _, ok := stepsgen[misc.ContextBrowserChromedpCtx].(context.Context); ok {
		chromedpCtx := stepsgen[misc.ContextBrowserChromedpCtx].(context.Context)
		cancel := stepsgen[misc.ContextBrowserChromedpCancel].(context.CancelFunc)

		err = chromedp.Stop().Do(chromedpCtx)

		cancel()
	}

	return nil, err
}

// onFailure implements the browser.onFailure primitive.
func (p *Browser) onFailure(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	var err error

	if _, ok := stepsgen[misc.ContextAttachment].(map[string][]byte); ok {
		log.Debug("Generating screenshot")
		if _, ok := stepsgen[misc.ContextBrowserChromedpCtx].(context.Context); !ok {
			return nil, errPluginNotInitialized
		}
		chromedpCtx := stepsgen[misc.ContextBrowserChromedpCtx].(context.Context)

		timeout := 30 * time.Second

		if _, ok := stepsgen[misc.ContextTimeout].(time.Duration); ok {
			timeout = stepsgen[misc.ContextTimeout].(time.Duration)
		}

		ackCtx, _ := context.WithTimeout(chromedpCtx, timeout) //nolint:all

		// Take a screenshot
		var buf []byte

		err = chromedp.Run(ackCtx, chromedp.CaptureScreenshot(&buf))

		if err != nil {
			log.Debugf("Error taking screenshot: %s", err)
			return nil, err
		}

		log.Debug("Screenshot taken")
		stepsgen[misc.ContextAttachment].(map[string][]byte)["screenshot.png"] = buf
	}

	return nil, err
}

// click implements the browser.click primitive.
func (p *Browser) click(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	if _, ok := stepsgen[misc.ContextBrowserChromedpCtx].(context.Context); !ok {
		return nil, errPluginNotInitialized
	}

	chromedpCtx := stepsgen[misc.ContextBrowserChromedpCtx].(context.Context)

	timeout := 30 * time.Second

	if _, ok := stepsgen[misc.ContextTimeout].(time.Duration); ok {
		timeout = stepsgen[misc.ContextTimeout].(time.Duration)
	}

	ackCtx, _ := context.WithTimeout(chromedpCtx, timeout) //nolint:all

	err := chromedp.Run(ackCtx, chromedp.Click(args["selector"], selector2By(args["selectorBy"])))

	return nil, err
}

// wait implements the browser.wait primitive.
func (p *Browser) wait(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	duration, err := time.ParseDuration(args["duration"])

	if err != nil {
		return nil, err
	}

	time.Sleep(duration)

	return nil, nil
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
				Description: selectorDesc,
				Optional:    false,
			},
			{
				Name:        "selectorBy",
				Description: selectorTypeDesc,
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
				Description: selectorDesc,
				Optional:    false,
			},
			{
				Name:        "selectorBy",
				Description: selectorTypeDesc,
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
				Description: selectorDesc,
				Optional:    false,
			},
			{
				Name:        "selectorBy",
				Description: selectorTypeDesc,
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
				Description: selectorDesc,
				Optional:    false,
			},
			{
				Name:        "selectorBy",
				Description: selectorTypeDesc,
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
		Name:        "setViewPort",
		Description: "Sets the viewport size",
		Params: []plugins.StepParam{
			{
				Name:        "width",
				Description: "Width of the viewport",
				Optional:    false,
			},
			{
				Name:        "height",
				Description: "Height of the viewport",
				Optional:    false,
			},
		},
		Fn: p.setViewPort,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "onClose",
		Description: "Close the connection",
		Params:      []plugins.StepParam{},
		Fn:          p.onClose,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "onFailure",
		Description: "Close the connection on failure",
		Params:      []plugins.StepParam{},
		Fn:          p.onFailure,
	})
}

// Init initializes the plugin.
func init() {
	h := &Browser{}
	h.Init()
	plugins.AddPlugin("browser", "Browser plugin is used to interact with a browser", h)
}

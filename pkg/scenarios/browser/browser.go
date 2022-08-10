// Package browser run tests in real browser
package browser

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/hidracloud/hidra/v2/pkg/models"
	"github.com/hidracloud/hidra/v2/pkg/scenarios"
)

// Scenario Represent an browser scenario
type Scenario struct {
	models.Scenario
	ctx    context.Context
	cancel context.CancelFunc
}

// RCA generate RCAs for scenario
func (b *Scenario) RCA(result *models.ScenarioResult) error {
	log.Println("Chrome RCA")
	return nil
}

func (b *Scenario) urlShouldBe(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	var url string
	err := chromedp.Run(b.ctx,
		chromedp.Location(&url),
	)

	if err != nil {
		return nil, err
	}

	if url != c["url"] {
		return nil, fmt.Errorf("url is not %s is %s", c["url"], url)
	}
	return nil, nil
}

func (b *Scenario) textShouldBe(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	var text string
	err := chromedp.Run(b.ctx,
		chromedp.Text(c["selector"], &text, selector2By(c["selector"])),
	)

	if err != nil {
		return nil, err
	}

	if text != c["text"] {
		return nil, fmt.Errorf("text is not %s is %s", c["text"], text)
	}
	return nil, nil
}

func (b *Scenario) sendKeys(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	err := chromedp.Run(b.ctx,
		chromedp.SendKeys(c["selector"], c["keys"], selector2By(c["selector"])),
	)

	if err != nil {
		return nil, err
	}

	return nil, nil
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

func (b *Scenario) waitVisible(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	err := chromedp.Run(b.ctx,
		chromedp.WaitVisible(c["selector"], selector2By(c["selector"])),
	)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *Scenario) click(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	err := chromedp.Run(b.ctx,
		chromedp.Click(c["selector"], selector2By(c["selector"])),
	)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *Scenario) navigateTo(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	err := chromedp.Run(b.ctx,
		chromedp.Navigate(c["url"]),
	)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *Scenario) takeScreenshot(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	var screenshot []byte
	var err error

	if c["selector"] == "" {
		err = chromedp.Run(b.ctx,
			chromedp.FullScreenshot(&screenshot, 90),
		)
	} else {
		err = chromedp.Run(b.ctx,
			chromedp.Screenshot(c["selector"], &screenshot),
		)
	}

	if err != nil {
		return nil, err
	}

	// save screenshot bytes to file
	f, err := os.OpenFile("screenshot.png", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()

	if err != nil {
		return nil, err
	}

	f.Write(screenshot)

	return nil, nil
}

func (b *Scenario) wait(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	sleep, _ := time.ParseDuration(c["sleep"])

	err := chromedp.Run(b.ctx,
		chromedp.Sleep(sleep),
	)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Description return scenario description
func (b *Scenario) Description() string {
	return "Run scenario in real browser"
}

// Close closes the scenario
func (b *Scenario) Close() {
	err := chromedp.Stop().Do(b.ctx)
	if err != nil {
		log.Println(err)
	}
	b.ctx.Done()

	// keep browser open
	if os.Getenv("DEBUG") == "" {
		b.cancel()
	}
}

// Init initialize scenario
func (b *Scenario) Init() {
	b.StartPrimitives()

	// Initialize chrome context
	initialCtx := context.Background()

	// If debug env is set, use debug mode
	if os.Getenv("DEBUG") != "" {
		initialCtx, _ = chromedp.NewExecAllocator(initialCtx, append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", false))...)
	}

	b.ctx, b.cancel = chromedp.NewContext(
		initialCtx,
	)

	b.RegisterStep("navigateTo", models.StepDefinition{
		Description: "It navigates to a url",
		Params: []models.StepParam{
			{Name: "url", Optional: false, Description: "The url to navigate to"},
		},
		Fn: b.navigateTo,
	})

	b.RegisterStep("urlShouldBe", models.StepDefinition{
		Description: "It checks if the url is the expected one",
		Params: []models.StepParam{
			{Name: "url", Optional: false, Description: "The url to check"},
		},
		Fn: b.urlShouldBe,
	})

	b.RegisterStep("textShouldBe", models.StepDefinition{
		Description: "It checks if the text is the expected one",
		Params: []models.StepParam{
			{Name: "text", Optional: false, Description: "The text to check"},
			{Name: "selector", Optional: false, Description: "The selector to check"},
			{Name: "searchMethod", Optional: true, Description: "The search method to use. "},
		},
		Fn: b.textShouldBe,
	})

	b.RegisterStep("sendKeys", models.StepDefinition{
		Description: "It sends keys to a selector",
		Params: []models.StepParam{
			{Name: "keys", Optional: false, Description: "The keys to send to the selector. "},
			{Name: "selector", Optional: false, Description: "The selector to send the keys to. "},
			{Name: "searchMethod", Optional: true, Description: "The search method to use. "},
		},
		Fn: b.sendKeys,
	})

	b.RegisterStep("waitVisible", models.StepDefinition{
		Description: "It waits until the selector is visible",
		Params: []models.StepParam{
			{Name: "selector", Optional: false, Description: "The selector to wait for. "},
			{Name: "searchMethod", Optional: true, Description: "The search method to use. "},
		},
		Fn: b.waitVisible,
	})

	b.RegisterStep("click", models.StepDefinition{
		Description: "It clicks on a selector",
		Params: []models.StepParam{
			{Name: "selector", Optional: false, Description: "The selector to click. "},
			{Name: "searchMethod", Optional: true, Description: "The search method to use. "},
		},
		Fn: b.click,
	})

	b.RegisterStep("takeScreenshot", models.StepDefinition{
		Description: "It takes a screenshot",
		Params: []models.StepParam{
			{Name: "selector", Optional: true, Description: "The selector to take the screenshot of. "},
		},
		Fn: b.takeScreenshot,
	})

	b.RegisterStep("wait", models.StepDefinition{
		Description: "It waits for a given amount of time",
		Params: []models.StepParam{
			{Name: "sleep", Optional: false, Description: "The amount of time to wait. "},
		},
		Fn: b.wait,
	})

}

func init() {
	scenarios.Add("browser", func() models.IScenario {
		return &Scenario{}
	})
}

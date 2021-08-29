// Monitoring webapp
package browser

import (
	"context"
	"fmt"

	"github.com/hidracloud/hidra/models"
	"github.com/hidracloud/hidra/scenarios"

	"github.com/chromedp/chromedp"
)

// Represent an browser scenario
type BrowserScenario struct {
	models.Scenario
	ctx context.Context
}

func (b *BrowserScenario) urlShouldBe(c map[string]string) ([]models.Metric, error) {
	if _, ok := c["url"]; !ok {
		return nil, fmt.Errorf("url parameter missing")
	}

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

func (b *BrowserScenario) textShouldBe(c map[string]string) ([]models.Metric, error) {
	if _, ok := c["text"]; !ok {
		return nil, fmt.Errorf("text parameter missing")
	}

	if _, ok := c["selector"]; !ok {
		return nil, fmt.Errorf("selector parameter missing")
	}

	var text string
	err := chromedp.Run(b.ctx,
		chromedp.Text(c["selector"], &text),
	)

	if err != nil {
		return nil, err
	}

	if text != c["text"] {
		return nil, fmt.Errorf("text is not %s is %s", c["text"], text)
	}
	return nil, nil
}

func (b *BrowserScenario) sendKeys(c map[string]string) ([]models.Metric, error) {
	if _, ok := c["keys"]; !ok {
		return nil, fmt.Errorf("keys parameter missing")
	}

	if _, ok := c["selector"]; !ok {
		return nil, fmt.Errorf("selector parameter missing")
	}

	err := chromedp.Run(b.ctx,
		chromedp.SendKeys(c["selector"], c["keys"]),
	)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *BrowserScenario) waitVisible(c map[string]string) ([]models.Metric, error) {
	if _, ok := c["selector"]; !ok {
		return nil, fmt.Errorf("selector parameter missing")
	}

	err := chromedp.Run(b.ctx,
		chromedp.WaitVisible(c["selector"]),
	)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *BrowserScenario) click(c map[string]string) ([]models.Metric, error) {
	if _, ok := c["selector"]; !ok {
		return nil, fmt.Errorf("selector parameter missing")
	}

	err := chromedp.Run(b.ctx,
		chromedp.Click(c["selector"], chromedp.NodeVisible),
	)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *BrowserScenario) navigateTo(c map[string]string) ([]models.Metric, error) {
	if _, ok := c["url"]; !ok {
		return nil, fmt.Errorf("url parameter missing")
	}

	err := chromedp.Run(b.ctx,
		chromedp.Navigate(c["url"]),
	)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *BrowserScenario) Description() string {
	return "It executes actions on a real browser, as if it were being executed by Google Chrome."
}

func (b *BrowserScenario) Init() {
	b.StartPrimitives()

	// Initialize chrome context
	b.ctx, _ = chromedp.NewContext(
		context.Background(),
		// chromedp.WithDebugf(log.Printf),
	)

	b.RegisterStep("navigateTo", b.navigateTo)
	b.RegisterStep("waitVisible", b.waitVisible)
	b.RegisterStep("click", b.click)
	b.RegisterStep("urlShouldBe", b.urlShouldBe)
	b.RegisterStep("sendKeys", b.sendKeys)
	b.RegisterStep("textShouldBe", b.textShouldBe)
}

func init() {
	scenarios.Add("browser", func() models.IScenario {
		return &BrowserScenario{}
	})
}

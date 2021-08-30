// HTTP scenario do test using HTTP protocol
package http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"

	"github.com/hidracloud/hidra/models"
	"github.com/hidracloud/hidra/scenarios"
)

// Represent an http scenario
type HttpScenario struct {
	models.Scenario

	Url      string
	Method   string
	Response *http.Response
	Body     string
	Redirect string
	Headers  map[string]string
	Client   *http.Client
}

// Set user agent
func (h *HttpScenario) setUserAgent(c map[string]string) ([]models.Metric, error) {
	h.Headers["User-Agent"] = c["user-agent"]
	return nil, nil
}

// Add new HTTP header
func (h *HttpScenario) addHttpHeader(c map[string]string) ([]models.Metric, error) {
	h.Headers[c["key"]] = c["value"]
	return nil, nil
}

// Send a request depends of the method
func (h *HttpScenario) requestByMethod(c map[string]string) ([]models.Metric, error) {
	var err error

	body := ""

	if _, ok := c["body"]; ok {
		body = c["body"]
	}

	jar, err := cookiejar.New(nil)

	if err != nil {
		return nil, err
	}

	h.Client.Jar = jar

	req, err := http.NewRequest(h.Method, h.Url, strings.NewReader(body))

	if err != nil {
		return nil, err
	}

	for k, v := range h.Headers {
		req.Header.Set(k, v)
	}

	resp, err := h.Client.Do(req)

	if err != nil {
		return nil, err
	}

	h.Response = resp

	b, err := ioutil.ReadAll(h.Response.Body)

	if err != nil {
		return nil, err
	}

	h.Body = strings.ToLower(string(b))
	h.Response.Body.Close()

	return nil, err
}

// Make http request to given URL
func (h *HttpScenario) request(c map[string]string) ([]models.Metric, error) {
	var err error
	var ok bool

	h.Url = c["url"]

	h.Method = "GET"

	if _, ok = c["method"]; ok {
		h.Method = strings.ToUpper(c["method"])
	}

	_, err = h.requestByMethod(c)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Check if status code match
func (h *HttpScenario) statusCodeShouldBe(c map[string]string) ([]models.Metric, error) {
	if h.Response == nil {
		return nil, fmt.Errorf("request should be initialized first")
	}

	if strconv.Itoa(h.Response.StatusCode) != c["statusCode"] {
		return nil, fmt.Errorf("statusCode expected %s, but %d", c["statusCode"], h.Response.StatusCode)
	}

	return nil, nil
}

func (h *HttpScenario) bodyShouldContain(c map[string]string) ([]models.Metric, error) {
	if h.Response == nil {
		return nil, fmt.Errorf("request should be initialized first")
	}

	if !strings.Contains(h.Body, strings.ToLower(c["search"])) {
		return nil, fmt.Errorf("expected %s in body, but not found", c["search"])
	}

	return nil, nil
}

func (h *HttpScenario) shouldRedirectTo(c map[string]string) ([]models.Metric, error) {

	if h.Response == nil {
		return nil, fmt.Errorf("request should be initialized first")
	}

	// Check if header Location is present
	if h.Response.Header.Get("Location") == "" {
		return nil, fmt.Errorf("expected Location header, but not found")
	}

	if h.Response.Header.Get("Location") != c["url"] {
		return nil, fmt.Errorf("expected redirect to %s, but got %s", c["url"], h.Response.Header.Get("Location"))
	}

	return nil, nil
}

// Clear parameters
func (h *HttpScenario) clear(c map[string]string) ([]models.Metric, error) {
	h.Url = ""
	h.Response = nil
	h.Method = ""
	h.Headers = make(map[string]string)

	return nil, nil
}

func (h *HttpScenario) Description() string {
	return "Run a HTTP scenario"
}

func (h *HttpScenario) Init() {
	h.StartPrimitives()

	h.Headers = make(map[string]string)
	h.Headers["User-Agent"] = "hidra/monitoring"

	h.Client = &http.Client{}

	h.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	h.RegisterStep("request", models.StepDefinition{
		Description: "Send a HTTP request",
		Params: []models.StepParam{
			{Name: "url", Optional: false, Description: "URL to send the request"},
			{Name: "method", Optional: true, Description: "HTTP method to use"},
			{Name: "body", Optional: true, Description: "Body to send"},
		},
		Fn: h.request,
	})

	h.RegisterStep("statusCodeShouldBe", models.StepDefinition{
		Description: "Check if status code is as expected",
		Params: []models.StepParam{
			{Name: "statusCode", Optional: false, Description: "Status code to check"},
		},
		Fn: h.statusCodeShouldBe,
	})

	h.RegisterStep("bodyShouldContain", models.StepDefinition{
		Description: "Check if body contains a string",
		Params: []models.StepParam{
			{Name: "search", Optional: false, Description: "String to search"},
		},
		Fn: h.bodyShouldContain,
	})

	h.RegisterStep("shouldRedirectTo", models.StepDefinition{
		Description: "Check if redirect to a given URL",
		Params: []models.StepParam{
			{Name: "url", Optional: false, Description: "URL to check"},
		},
		Fn: h.shouldRedirectTo,
	})

	h.RegisterStep("clear", models.StepDefinition{
		Description: "Clear parameters",
		Params:      []models.StepParam{},
		Fn:          h.clear,
	})

	h.RegisterStep("addHttpHeader", models.StepDefinition{
		Description: "Add a HTTP header",
		Params: []models.StepParam{
			{Name: "key", Optional: false, Description: "Header name"},
			{Name: "value", Optional: false, Description: "Header value"},
		},
		Fn: h.addHttpHeader,
	})

	h.RegisterStep("setUserAgent", models.StepDefinition{
		Description: "Set User-Agent header",
		Params: []models.StepParam{
			{Name: "user-agent", Optional: false, Description: "User-Agent value"},
		},
		Fn: h.setUserAgent,
	})

}

func init() {
	scenarios.Add("http", func() models.IScenario {
		return &HttpScenario{}
	})
}

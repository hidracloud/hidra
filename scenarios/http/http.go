package http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/JoseCarlosGarcia95/hidra/models"
	"github.com/JoseCarlosGarcia95/hidra/scenarios"
)

type HttpScenario struct {
	models.Scenario

	Url      string
	Method   string
	Response *http.Response
	Body     string
	Headers  map[string]string
	Client   *http.Client
}

// Set user agent
func (h *HttpScenario) setUserAgent(c map[string]string) error {
	var ok bool
	if h.Headers["User-Agent"], ok = c["user-agent"]; !ok {
		return fmt.Errorf("user-agent parameter missing")
	}
	return nil
}

// Add new HTTP header
func (h *HttpScenario) addHttpHeader(c map[string]string) error {
	if _, ok := c["key"]; !ok {
		return fmt.Errorf("key parameter missing")
	}
	if _, ok := c["value"]; !ok {
		return fmt.Errorf("value parameter missing")
	}

	h.Headers[c["key"]] = c["value"]
	return nil
}

// Send a request depends of the method
func (h *HttpScenario) requestByMethod(c map[string]string) error {
	var err error

	body := ""

	if _, ok := c["body"]; ok {
		body = c["body"]
	}

	req, err := http.NewRequest(h.Method, h.Url, strings.NewReader(body))

	if err != nil {
		return err
	}

	for k, v := range h.Headers {
		req.Header.Set(k, v)
	}

	resp, err := h.Client.Do(req)

	if err != nil {
		return err
	}

	h.Response = resp

	b, err := ioutil.ReadAll(h.Response.Body)

	if err != nil {
		return err
	}

	h.Body = strings.ToLower(string(b))
	h.Response.Body.Close()

	return err
}

// Make http request to given URL
func (h *HttpScenario) request(c map[string]string) error {
	var err error
	var ok bool

	if h.Url, ok = c["url"]; !ok {
		return fmt.Errorf("url parameter missing")
	}

	h.Method = "GET"

	if _, ok = c["method"]; ok {
		h.Method = strings.ToUpper(c["method"])
	}

	err = h.requestByMethod(c)

	if err != nil {
		return err
	}

	return nil
}

// Check if status code match
func (h *HttpScenario) statusCodeShouldBe(c map[string]string) error {
	if h.Response == nil {
		return fmt.Errorf("request should be initialized first")
	}

	if _, ok := c["statusCode"]; !ok {
		return fmt.Errorf("statusCode parameter missing")
	}

	if strconv.Itoa(h.Response.StatusCode) != c["statusCode"] {
		return fmt.Errorf("statusCode expected %s, but %d", c["statusCode"], h.Response.StatusCode)
	}

	return nil
}

func (h *HttpScenario) bodyShouldContain(c map[string]string) error {
	if _, ok := c["search"]; !ok {
		return fmt.Errorf("search parameter missing")
	}

	if !strings.Contains(h.Body, strings.ToLower(c["search"])) {
		return fmt.Errorf("expected %s in body, but not found", c["search"])
	}

	return nil
}

// Clear parameters
func (h *HttpScenario) clear(c map[string]string) error {
	h.Url = ""
	h.Response = nil
	h.Method = ""
	h.Headers = make(map[string]string)

	return nil
}

func (h *HttpScenario) Init() {
	h.StartPrimitives()

	h.Headers = make(map[string]string)
	h.Headers["User-Agent"] = "hidra/monitoring"

	h.Client = &http.Client{}

	h.RegisterStep("request", h.request)
	h.RegisterStep("statusCodeShouldBe", h.statusCodeShouldBe)
	h.RegisterStep("setUserAgent", h.setUserAgent)
	h.RegisterStep("addHttpHeader", h.addHttpHeader)
	h.RegisterStep("bodyShouldContain", h.bodyShouldContain)

	h.RegisterStep("clear", h.clear)
}

func init() {
	scenarios.Add("http", func() models.IScenario {
		return &HttpScenario{}
	})
}

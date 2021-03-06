// Package http is a http scenario
package http

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptrace"
	"strconv"
	"strings"

	"github.com/hidracloud/hidra/v2/pkg/models"
	"github.com/hidracloud/hidra/v2/pkg/scenarios"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	errRequestShouldBeInitialized = "request should be initialized"
)

// Cache of bodies
var bodyCache = make(map[string]string)

// Scenario Represent an http scenario
type Scenario struct {
	models.Scenario

	URL             string
	Method          string
	Response        *http.Response
	Body            string
	Redirect        string
	ForceIP         string
	Headers         map[string]string
	Client          *http.Client
	SharedJar       *cookiejar.Jar
	SkipInsecureTLS bool `default:"false"`
}

// dialContext is a dialer that uses context to cancel a dial.
func (h *Scenario) dialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	d := &net.Dialer{}

	if h.ForceIP != "" {
		port := 80

		if strings.Contains(h.URL, "https") {
			port = 443
		}

		addr = h.ForceIP + ":" + strconv.Itoa(port)
	}

	return d.DialContext(ctx, network, addr)
}

// Set user agent
func (h *Scenario) setUserAgent(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	h.Headers["User-Agent"] = c["user-agent"]
	return nil, nil
}

// Add new HTTP header
func (h *Scenario) addHTTPHeader(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	h.Headers[c["key"]] = c["value"]
	return nil, nil
}

// Allow insecure TLS connections
func (h *Scenario) allowInsecureTLS(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	h.SkipInsecureTLS = true

	httpTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: h.SkipInsecureTLS},
		DialContext:     h.dialContext,
	}

	h.Client = &http.Client{Transport: otelhttp.NewTransport(httpTransport, otelhttp.WithClientTrace(func(ctx context.Context) *httptrace.ClientTrace {
		return otelhttptrace.NewClientTrace(ctx, otelhttptrace.WithoutSubSpans())
	}))}
	return nil, nil
}

// Resolve force IP resolution
func (h *Scenario) forceIP(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	h.ForceIP = c["ip"]
	return nil, nil
}

// Send a request depends of the method
func (h *Scenario) requestByMethod(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	var err error

	body := ""

	if _, ok := c["body"]; ok {
		body = c["body"]
	}

	h.Client.Jar = h.SharedJar

	// convert body to bytes
	bodyBytes := []byte(body)

	req, err := http.NewRequestWithContext(ctx, h.Method, h.URL, bytes.NewBuffer(bodyBytes))

	if err != nil {
		return nil, err
	}

	for k, v := range h.Headers {
		req.Header.Set(k, v)
	}

	resp, err := h.Client.Do(req)

	// clean headers after request
	h.Headers = make(map[string]string)

	if err != nil {
		return nil, err
	}

	h.Response = resp

	b, err := ioutil.ReadAll(h.Response.Body)

	if err != nil {
		return nil, err
	}

	h.Body = strings.ToLower(string(b))
	defer h.Response.Body.Close()

	return nil, err
}

// Make http request to given URL
func (h *Scenario) request(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	var err error
	var ok bool

	h.URL = c["url"]

	h.Method = "GET"

	if _, ok = c["method"]; ok {
		h.Method = strings.ToUpper(c["method"])
	}

	_, err = h.requestByMethod(ctx, c)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Check if status code match
func (h *Scenario) statusCodeShouldBe(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	if h.Response == nil {
		return nil, fmt.Errorf(errRequestShouldBeInitialized)
	}

	if strconv.Itoa(h.Response.StatusCode) != c["statusCode"] {
		return nil, fmt.Errorf("statusCode expected %s, but %d", c["statusCode"], h.Response.StatusCode)
	}

	return nil, nil
}

func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (h *Scenario) bodyShouldChange(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	if h.Response == nil {
		return nil, fmt.Errorf(errRequestShouldBeInitialized)
	}

	urlMD5 := getMD5Hash(h.URL)
	bodyMD5 := getMD5Hash(h.Body)

	// Get all the data from the cache
	if bodyCache[urlMD5] == "" {
		bodyCache[urlMD5] = bodyMD5
		return nil, nil
	}

	// Check if the body has changed
	if bodyCache[urlMD5] == bodyMD5 {
		return nil, fmt.Errorf("body should change")
	}

	bodyCache[urlMD5] = bodyMD5

	return nil, nil
}

func (h *Scenario) bodyShouldContain(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	if h.Response == nil {
		return nil, fmt.Errorf(errRequestShouldBeInitialized)
	}

	if !strings.Contains(h.Body, strings.ToLower(c["search"])) {
		return nil, fmt.Errorf("expected %s in body, but not found", c["search"])
	}

	return nil, nil
}

func (h *Scenario) shouldRedirectTo(ctx context.Context, c map[string]string) ([]models.Metric, error) {

	if h.Response == nil {
		return nil, fmt.Errorf(errRequestShouldBeInitialized)
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
func (h *Scenario) clear(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	h.URL = ""
	h.Response = nil
	h.Method = ""
	h.Headers = make(map[string]string)

	return nil, nil
}

// Description return scenario description
func (h *Scenario) Description() string {
	return "Run a HTTP scenario"
}

// RCA generate RCAs for scenario
func (h *Scenario) RCA(result *models.ScenarioResult) error {
	log.Println("HTTP RCA")
	return nil
}

// Close closes the scenario
func (h *Scenario) Close() {
	h.Client.CloseIdleConnections()
}

// Init initialize scenario
func (h *Scenario) Init() {
	h.StartPrimitives()

	h.Headers = make(map[string]string)
	h.Headers["User-Agent"] = "hidra/monitoring"

	httpTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: h.SkipInsecureTLS},
		DialContext:     h.dialContext,
	}

	h.Client = &http.Client{Transport: otelhttp.NewTransport(httpTransport, otelhttp.WithClientTrace(func(ctx context.Context) *httptrace.ClientTrace {
		return otelhttptrace.NewClientTrace(ctx, otelhttptrace.WithoutSubSpans())
	}))}

	h.SharedJar, _ = cookiejar.New(nil)
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

	h.RegisterStep("bodyShouldChange", models.StepDefinition{
		Description: "Check if body has changed",
		Params:      []models.StepParam{},
		Fn:          h.bodyShouldChange,
	})

	h.RegisterStep("addHTTPHeader", models.StepDefinition{
		Description: "Add a HTTP header",
		Params: []models.StepParam{
			{Name: "key", Optional: false, Description: "Header name"},
			{Name: "value", Optional: false, Description: "Header value"},
		},
		Fn: h.addHTTPHeader,
	})

	h.RegisterStep("allowInsecureTLS", models.StepDefinition{
		Description: "Allow insecure TLS",
		Params:      []models.StepParam{},
		Fn:          h.allowInsecureTLS,
	})

	h.RegisterStep("forceIP", models.StepDefinition{
		Description: "Resolve IP",
		Params: []models.StepParam{
			{Name: "ip", Optional: false, Description: "IP to resolve"},
		},
		Fn: h.forceIP,
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
		return &Scenario{}
	})
}

package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hidracloud/hidra/v3/config"
	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/internal/misc"
	"github.com/hidracloud/hidra/v3/internal/plugins"
	"github.com/hidracloud/hidra/v3/internal/runner"
	"github.com/hidracloud/hidra/v3/internal/utils"
	log "github.com/sirupsen/logrus"
)

var (
	httpClient = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Get context from request
			ctx := req.Context()

			// Check if context has followRedirects
			if _, ok := ctx.Value(misc.ContextHTTPFollowRedirects).(bool); !ok {
				return http.ErrUseLastResponse
			}

			return nil
		},
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     10 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				d := &net.Dialer{}

				if _, ip := ctx.Value(misc.ContextHTTPForceIP).(string); ip {
					addr = fmt.Sprintf("%s:%s", ctx.Value(misc.ContextHTTPForceIP), strings.Split(addr, ":")[1])
				}

				return d.DialContext(ctx, network, addr)
			},
		},
	}

	errContextNotFound = errors.New("context doesn't have the expected")
)

// HTTP represents a HTTP plugin.
type HTTP struct {
	plugins.BasePlugin
}

// CacheAgeShouldBeLowerThan represents a HTTP cache age should be lower than.
func (p *HTTP) cacheAgeShouldBeLowerThan(ctx context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	var err error

	// get context for current step
	if _, ok := stepsgen[misc.ContextHTTPResponse].(*http.Response); !ok {
		return nil, errContextNotFound
	}

	resp := stepsgen[misc.ContextHTTPResponse].(*http.Response)

	ageStr := resp.Header.Get("Age")

	if ageStr == "" {
		return nil, fmt.Errorf("cache age not found")
	}

	age, err := strconv.ParseInt(ageStr, 10, 64)

	if err != nil {
		return nil, err
	}

	maxAge, err := strconv.ParseInt(args["maxAge"], 10, 64)

	customMetrics := []*metrics.Metric{
		{
			Name:        "http_response_cache_age",
			Description: "The HTTP response cache age",
			Value:       float64(age),
			Labels: map[string]string{
				"method": stepsgen[misc.ContextHTTPMethod].(string),
				"url":    stepsgen[misc.ContextHTTPURL].(string),
			},
		},
	}

	if err != nil {
		return customMetrics, err
	}

	if age > maxAge {
		return customMetrics, fmt.Errorf("cache age is %d, expected to be lower than %d", age, maxAge)
	}

	return customMetrics, err
}

// RequestByMethod makes a HTTP request by method.
func (p *HTTP) requestByMethod(ctx context.Context, c map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	var err error

	body := ""

	if _, ok := c["body"]; ok {
		body = c["body"]
	}

	dnsStartTime := time.Time{}
	dnsStopTime := time.Time{}

	tcpStartTime := time.Time{}
	tcpStopTime := time.Time{}

	tlsStartTime := time.Time{}
	tlsStopTime := time.Time{}

	var certificates []*x509.Certificate

	clientTrace := &httptrace.ClientTrace{
		DNSStart: func(dnsInfo httptrace.DNSStartInfo) {
			dnsStartTime = time.Now()
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			dnsAddr := ""

			if len(dnsInfo.Addrs) > 0 {
				dnsAddr = dnsInfo.Addrs[0].String()
			}

			dnsStopTime = time.Now()
			stepsgen[misc.ContextConnectionIP] = dnsAddr
		},
		ConnectStart: func(network, addr string) {
			tcpStartTime = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			tcpStopTime = time.Now()
		},
		TLSHandshakeStart: func() {
			tlsStartTime = time.Now()
		},
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			tlsStopTime = time.Now()
			certificates = cs.PeerCertificates
		},
	}

	timeout := 30 * time.Second

	if _, ok := stepsgen[misc.ContextTimeout].(time.Duration); ok {
		timeout = stepsgen[misc.ContextTimeout].(time.Duration)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if _, ok := stepsgen[misc.ContextHTTPForceIP].(string); ok {
		// nolint:staticcheck
		ctx = context.WithValue(ctx, misc.ContextHTTPForceIP, stepsgen[misc.ContextHTTPForceIP])
	}

	if _, ok := stepsgen[misc.ContextHTTPFollowRedirects].(bool); ok {
		// nolint:staticcheck
		ctx = context.WithValue(ctx, misc.ContextHTTPFollowRedirects, stepsgen[misc.ContextHTTPFollowRedirects])
	}

	ctx = httptrace.WithClientTrace(ctx, clientTrace)
	req, err := http.NewRequestWithContext(ctx, stepsgen[misc.ContextHTTPMethod].(string), stepsgen[misc.ContextHTTPURL].(string), bytes.NewBuffer([]byte(body)))

	if err != nil {
		return nil, err
	}

	userAgentSet := false
	if ctxHeaders, ok := stepsgen[misc.ContextHTTPHeaders].(map[string]string); ok {
		for k, v := range ctxHeaders {
			if strings.ToLower(k) == "user-agent" {
				userAgentSet = true
			}
			req.Header.Set(k, v)
		}
	}

	if !userAgentSet {
		req.Header.Set("User-Agent", fmt.Sprintf("hidra/monitoring %s", misc.Version))
	}

	startTime := time.Now()
	resp, err := httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	_, err = io.Copy(io.Discard, resp.Body)

	if err != nil {
		return nil, err
	}

	stepsgen[misc.ContextHTTPResponse] = resp
	stepsgen[misc.ContextOutput] = b

	dnsTime := dnsStopTime.Sub(dnsStartTime).Seconds()

	tcpTime := tcpStopTime.Sub(tcpStartTime).Seconds()

	tlsTime := tlsStopTime.Sub(tlsStartTime).Seconds()

	customMetrics := []*metrics.Metric{
		{
			Name:        "http_response_status_code",
			Description: "The HTTP response status code",
			Value:       float64(resp.StatusCode),
			Labels: map[string]string{
				"method": stepsgen[misc.ContextHTTPMethod].(string),
				"url":    stepsgen[misc.ContextHTTPURL].(string),
			},
		},
		{
			Name:        "http_response_content_length",
			Description: "The HTTP response content length",
			Value:       float64(len(b)),
			Labels: map[string]string{
				"method": stepsgen[misc.ContextHTTPMethod].(string),
				"url":    stepsgen[misc.ContextHTTPURL].(string),
			},
		},
		{
			Name:        "http_response_time",
			Description: "The HTTP response time",
			Value:       time.Since(startTime).Seconds(),
			Labels: map[string]string{
				"method": stepsgen[misc.ContextHTTPMethod].(string),
				"url":    stepsgen[misc.ContextHTTPURL].(string),
			},
		},
		{
			Name:        "http_response_dns_time",
			Description: "The HTTP response DNS time",
			Value:       dnsTime,
			Labels: map[string]string{
				"method": stepsgen[misc.ContextHTTPMethod].(string),
				"url":    stepsgen[misc.ContextHTTPURL].(string),
			},
		},
		{
			Name:        "http_response_tcp_connect_time",
			Description: "The HTTP response TCP connect time",
			Value:       tcpTime,
			Labels: map[string]string{
				"method": stepsgen[misc.ContextHTTPMethod].(string),
				"url":    stepsgen[misc.ContextHTTPURL].(string),
			},
		},
		{
			Name:        "http_response_tls_handshake_time",
			Description: "The HTTP response TLS handshake time",
			Value:       tlsTime,
			Labels: map[string]string{
				"method": stepsgen[misc.ContextHTTPMethod].(string),
				"url":    stepsgen[misc.ContextHTTPURL].(string),
			},
		},
	}

	certificatesShouldBeValidated := len(certificates) > 0

	if val, ok := stepsgen[misc.ContextHTTPTlsInsecureSkipVerify].(bool); ok && val {
		certificatesShouldBeValidated = false
	}

	if certificatesShouldBeValidated {
		// extract hostname from URL
		u, err := url.Parse(stepsgen[misc.ContextHTTPURL].(string))

		if err != nil {
			return nil, err
		}

		for _, certificate := range certificates {
			customMetrics = append(customMetrics, &metrics.Metric{
				Name: "tls_certificate_not_after",
				Labels: map[string]string{
					"serial_number": certificate.SerialNumber.String(),
					"subject":       certificate.Subject.String(),
					"host":          u.Host,
				},
				Value: float64(certificate.NotAfter.Unix()),
			})

			customMetrics = append(customMetrics, &metrics.Metric{
				Name: "tls_certificate_not_before",
				Labels: map[string]string{
					"serial_number": certificate.SerialNumber.String(),
					"subject":       certificate.Subject.String(),
					"host":          u.Host,
				},
				Value: float64(certificate.NotBefore.Unix()),
			})

			customMetrics = append(customMetrics, &metrics.Metric{
				Name: "tls_certificate_version",
				Labels: map[string]string{
					"serial_number": certificate.SerialNumber.String(),
					"subject":       certificate.Subject.String(),
					"host":          u.Host,
				},
				Value: float64(certificate.Version),
			})
		}

		dnsPlugin := plugins.GetPlugin("dns")

		if dnsPlugin != nil {
			dnsMetrics, err := dnsPlugin.RunStep(ctx, stepsgen, &plugins.Step{
				Name:    "whoisFrom",
				Args:    map[string]string{"domain": u.Host},
				Timeout: int(timeout.Seconds()),
				Negate:  false,
			})

			if err != nil {
				log.Debugf("failed to generate whois metrics: %s", err)
			} else {
				customMetrics = append(customMetrics, dnsMetrics...)
			}
		}

		var sample *config.SampleConfig

		if val, ok := stepsgen[misc.ContextSample].(*config.SampleConfig); ok {
			sample = val
		}

		runner.RegisterBackgroundTask(func() ([]*metrics.Metric, *config.SampleConfig, error) {
			icmpPlugin := plugins.GetPlugin("icmp")

			if icmpPlugin != nil {
				tracerouteMetrics, err := icmpPlugin.RunStep(ctx, stepsgen, &plugins.Step{
					Name:    "traceroute",
					Args:    map[string]string{"hostname": u.Host},
					Timeout: int(timeout.Seconds()),
					Negate:  false,
				})

				return tracerouteMetrics, sample, err
			}

			return nil, sample, fmt.Errorf("icmp plugin not found")
		})
	}

	return customMetrics, err
}

// request represents a HTTP request.
func (p *HTTP) request(ctx context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	if _, ok := args["method"]; !ok {
		args["method"] = "GET"
	}

	// set context for current step
	stepsgen[misc.ContextHTTPMethod] = args["method"]
	stepsgen[misc.ContextHTTPURL] = args["url"]
	stepsgen[misc.ContextHTTPBody] = args["body"]

	return p.requestByMethod(ctx, args, stepsgen)
}

// statusCodeShouldBe represents a HTTP status code should be.
func (p *HTTP) statusCodeShouldBe(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	var err error

	// get context for current step

	if _, ok := stepsgen[misc.ContextHTTPResponse].(*http.Response); !ok {
		return nil, errContextNotFound
	}

	resp := stepsgen[misc.ContextHTTPResponse].(*http.Response)

	expectedStatusCode, err := strconv.ParseInt(args["statusCode"], 10, 64)

	if err != nil {
		return nil, err
	}

	if int64(resp.StatusCode) != expectedStatusCode {
		return nil, fmt.Errorf("expected status code %d but got %d", expectedStatusCode, resp.StatusCode)
	}

	return nil, err
}

// bodyShouldContain represents a HTTP body should contain.
func (p *HTTP) bodyShouldContain(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	var err error

	// get context for current step

	if _, ok := stepsgen[misc.ContextOutput].([]byte); !ok {
		return nil, errContextNotFound
	}

	output := utils.BytesToLowerCase(stepsgen[misc.ContextOutput].([]byte))

	times := 1

	if val, ok := args["times"]; ok {
		times, err = strconv.Atoi(val)

		if err != nil {
			return nil, err
		}
	}

	if times == 1 {
		if !bytes.Contains(output, []byte(strings.ToLower(args["search"]))) {
			return nil, fmt.Errorf("expected body to contain %s", args["search"])
		}

		return nil, err
	}

	appear := utils.BytesContainsStringTimes(output, strings.ToLower(args["search"]))

	if times > appear {
		return nil, fmt.Errorf("expected body to contain %s %d times, but only %d", args["search"], times, appear)
	}

	return nil, err
}

// shouldRedirectTo represents a HTTP should redirect to.
func (p *HTTP) shouldRedirectTo(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	var err error

	// get context for current step
	if _, ok := stepsgen[misc.ContextHTTPResponse].(*http.Response); !ok {
		return nil, errContextNotFound
	}

	resp := stepsgen[misc.ContextHTTPResponse].(*http.Response)

	if resp.Header.Get("Location") != args["url"] {
		return nil, fmt.Errorf("expected redirect to %s but got %s", args["url"], resp.Header.Get("Location"))
	}

	return nil, err
}

// addHTTPHeader represents a HTTP add header.
func (p *HTTP) addHTTPHeader(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	var err error

	// get context for current step
	if _, ok := stepsgen[misc.ContextHTTPHeaders].(map[string]string); !ok {
		stepsgen[misc.ContextHTTPHeaders] = make(map[string]string)
	}

	headers := stepsgen[misc.ContextHTTPHeaders].(map[string]string)

	headers[args["key"]] = args["value"]

	return nil, err
}

// setUserAgent represents a HTTP set user agent.
func (p *HTTP) setUserAgent(ctx context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	return p.addHTTPHeader(ctx, map[string]string{
		"key":   "User-Agent",
		"value": args["user-agent"],
	}, stepsgen)
}

// onFailure implements the plugins.Plugin interface.
func (p *HTTP) onFailure(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {

	if _, ok := stepsgen[misc.ContextAttachment].(map[string][]byte); ok {
		// get output from context
		if _, ok := stepsgen[misc.ContextOutput].([]byte); !ok {
			return nil, nil
		}

		if output, ok := stepsgen[misc.ContextOutput].([]byte); ok {
			stepsgen[misc.ContextAttachment].(map[string][]byte)["response.html"] = output
		}
	}
	// Generate an screenshot of current response
	return nil, nil
}

// onClose implements the plugins.Plugin interface.
func (p *HTTP) onClose(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	httpClient.CloseIdleConnections()
	return nil, nil
}

// Init initializes the plugin.
func (p *HTTP) Init() {
	p.Primitives()

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "request",
		Description: "Makes a HTTP request",
		Params: []plugins.StepParam{
			{Name: "method", Description: "The HTTP method", Optional: true},
			{Name: "url", Description: "The URL", Optional: false},
			{Name: "body", Description: "The body", Optional: true},
		},
		Fn: p.request,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "statusCodeShouldBe",
		Description: "Checks if the status code is equal to the expected value",
		Params: []plugins.StepParam{
			{Name: "statusCode", Description: "The expected status code", Optional: false},
		},
		Fn: p.statusCodeShouldBe,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "bodyShouldContain",
		Description: "[DEPRECATED] Please use outputShouldContain from string plugin. Checks if the body contains the expected value",
		Params: []plugins.StepParam{
			{Name: "search", Description: "The expected value", Optional: false},
			{Name: "times", Description: "The number of times the value should appear in the body", Optional: true},
		},
		Fn: p.bodyShouldContain,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "shouldRedirectTo",
		Description: "Checks if the response redirects to the expected URL",
		Params: []plugins.StepParam{
			{Name: "url", Description: "The expected URL", Optional: false},
		},
		Fn: p.shouldRedirectTo,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "addHTTPHeader",
		Description: "Adds a HTTP header to the request. If the header already exists, it will be overwritten",
		Params: []plugins.StepParam{
			{Name: "key", Description: "The header name", Optional: false},
			{Name: "value", Description: "The header value", Optional: false},
		},
		Fn: p.addHTTPHeader,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "setUserAgent",
		Description: "Sets the User-Agent header",
		Params: []plugins.StepParam{
			{Name: "user-agent", Description: "The User-Agent value", Optional: false},
		},
		Fn: p.setUserAgent,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "allowInsecureTLS",
		Description: "Allows insecure TLS connections. This is useful for testing purposes, but should not be used in production",
		Fn: func(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
			stepsgen[misc.ContextHTTPTlsInsecureSkipVerify] = true
			return nil, nil
		},
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "forceIP",
		Description: "Forces the IP address to use for the request",
		Params: []plugins.StepParam{
			{Name: "ip", Description: "The IP address", Optional: false},
		},
		Fn: func(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
			stepsgen[misc.ContextHTTPForceIP] = args["ip"]
			return nil, nil
		},
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "followRedirects",
		Description: "Follows the redirect",
		Fn: func(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
			stepsgen[misc.ContextHTTPFollowRedirects] = true
			return nil, nil
		},
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "cacheAgeShouldBeLowerThan",
		Description: "Checks if the cache age is lower than the expected value",
		Params: []plugins.StepParam{
			{Name: "maxAge", Description: "The max age", Optional: false},
		},
		Fn: p.cacheAgeShouldBeLowerThan,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "onFailure",
		Description: "Executes the steps if the previous step failed",
		Fn:          p.onFailure,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "onClose",
		Description: "Executes the steps when the test is finished",
		Fn:          p.onClose,
	})
}

// Init initializes the plugin.
func init() {
	h := &HTTP{}
	h.Init()
	plugins.AddPlugin("http", "HTTP plugin is used to make HTTP requests", h)
}

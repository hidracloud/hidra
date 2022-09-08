package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"strconv"
	"strings"
	"time"

	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/plugins"
)

// HTTP represents a HTTP plugin.
type HTTP struct {
	plugins.BasePlugin
}

// RequestByMethod makes a HTTP request by method.
func (p *HTTP) requestByMethod(ctx context.Context, c map[string]string) (context.Context, []*metrics.Metric, error) {
	var err error

	body := ""

	if _, ok := c["body"]; ok {
		body = c["body"]
	}

	var clientJar http.CookieJar
	if sharedJar, ok := ctx.Value(plugins.ContextSharedJar).(http.CookieJar); ok {
		clientJar = sharedJar
	} else {
		ctx = context.WithValue(ctx, plugins.ContextSharedJar, clientJar)
	}

	tlsSkipInsecure := false

	if _, ok := ctx.Value(plugins.ContextHTTPTlsInsecureSkipVerify).(bool); ok {
		tlsSkipInsecure = ctx.Value(plugins.ContextHTTPTlsInsecureSkipVerify).(bool)
	}

	var httpClient *http.Client

	if _, ok := ctx.Value(plugins.ContextHTTPClient).(*http.Client); ok {
		httpClient = ctx.Value(plugins.ContextHTTPClient).(*http.Client)
	} else {
		timeout := 30 * time.Second

		if _, ok := ctx.Value(plugins.ContextTimeout).(time.Duration); ok {
			timeout = ctx.Value(plugins.ContextTimeout).(time.Duration)
		}
		
		httpClient = &http.Client{
			Jar: clientJar,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Timeout: timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: tlsSkipInsecure,
				},
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					d := &net.Dialer{}

					if _, ok := ctx.Value(plugins.ContextHTTPForceIP).(string); ok {
						addr = fmt.Sprintf("%s:%s", ctx.Value(plugins.ContextHTTPForceIP), strings.Split(addr, ":")[1])
					}

					return d.DialContext(ctx, network, addr)
				},
			},
		}

		ctx = context.WithValue(ctx, plugins.ContextHTTPClient, httpClient)
	}

	clientTrace := &httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			ctx = context.WithValue(ctx, plugins.ContextHTTPConnInfo, connInfo)
		},
		DNSStart: func(dnsInfo httptrace.DNSStartInfo) {
			ctx = context.WithValue(ctx, plugins.ContextHTTPDNSStartInfo, dnsInfo)
			ctx = context.WithValue(ctx, plugins.ContextHTTPDNSStartTime, time.Now())
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			ctx = context.WithValue(ctx, plugins.ContextHTTPDNSDoneInfo, dnsInfo)
			ctx = context.WithValue(ctx, plugins.ContextHTTPDNSStopTime, time.Now())
		},
		ConnectStart: func(network, addr string) {
			ctx = context.WithValue(ctx, plugins.ContextHTTPNetwork, network)
			ctx = context.WithValue(ctx, plugins.ContextHTTPAddr, addr)
			ctx = context.WithValue(ctx, plugins.ContextHTTPTcpConnectStartTime, time.Now())
		},
		ConnectDone: func(network, addr string, err error) {
			ctx = context.WithValue(ctx, plugins.ContextHTTPNetwork, network)
			ctx = context.WithValue(ctx, plugins.ContextHTTPAddr, addr)
			ctx = context.WithValue(ctx, plugins.ContextHTTPTcpConnectStopTime, time.Now())
		},
		TLSHandshakeStart: func() {
			ctx = context.WithValue(ctx, plugins.ContextHTTPTlsHandshakeStartTime, time.Now())
		},
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			ctx = context.WithValue(ctx, plugins.ContextHTTPTlsHandshakeStopTime, time.Now())
		},
	}

	ctx = httptrace.WithClientTrace(ctx, clientTrace)
	req, err := http.NewRequestWithContext(ctx, ctx.Value(plugins.ContextHTTPMethod).(string), ctx.Value(plugins.ContextHTTPURL).(string), bytes.NewBuffer([]byte(body)))

	if err != nil {
		return ctx, nil, err
	}

	if ctxHeaders, ok := ctx.Value(plugins.ContextHTTPHeaders).(map[string]string); ok {
		for k, v := range ctxHeaders {
			req.Header.Set(k, v)
		}
	}

	startTime := time.Now()
	resp, err := httpClient.Do(req)

	if err != nil {
		return ctx, nil, err
	}

	b, err := io.ReadAll(resp.Body)

	if err != nil {
		return ctx, nil, err
	}

	defer resp.Body.Close()

	ctx = context.WithValue(ctx, plugins.ContextHTTPResponse, resp)
	ctx = context.WithValue(ctx, plugins.ContextOutput, string(b))

	dnsTime := 0.0

	if dnsStartTime, ok := ctx.Value(plugins.ContextHTTPDNSStartTime).(time.Time); ok {
		if dnsStopTime, ok := ctx.Value(plugins.ContextHTTPDNSStopTime).(time.Time); ok {
			dnsTime = dnsStopTime.Sub(dnsStartTime).Seconds()
		}
	}

	tcpTime := 0.0

	if tcpStartTime, ok := ctx.Value(plugins.ContextHTTPTcpConnectStartTime).(time.Time); ok {
		if tcpStopTime, ok := ctx.Value(plugins.ContextHTTPTcpConnectStopTime).(time.Time); ok {
			tcpTime = tcpStopTime.Sub(tcpStartTime).Seconds()
		}
	}

	tlsTime := 0.0

	if tlsStartTime, ok := ctx.Value(plugins.ContextHTTPTlsHandshakeStartTime).(time.Time); ok {
		if tlsStopTime, ok := ctx.Value(plugins.ContextHTTPTlsHandshakeStopTime).(time.Time); ok {
			tlsTime = tlsStopTime.Sub(tlsStartTime).Seconds()
		}
	}

	customMetrics := []*metrics.Metric{
		{
			Name:        "http_response_status_code",
			Description: "The HTTP response status code",
			Value:       float64(resp.StatusCode),
			Labels: map[string]string{
				"method": ctx.Value(plugins.ContextHTTPMethod).(string),
				"url":    ctx.Value(plugins.ContextHTTPURL).(string),
			},
		},
		{
			Name:        "http_response_content_length",
			Description: "The HTTP response content length",
			Value:       float64(len(b)),
			Labels: map[string]string{
				"method": ctx.Value(plugins.ContextHTTPMethod).(string),
				"url":    ctx.Value(plugins.ContextHTTPURL).(string),
			},
		},
		{
			Name:        "http_response_time",
			Description: "The HTTP response time",
			Value:       time.Since(startTime).Seconds(),
			Labels: map[string]string{
				"method": ctx.Value(plugins.ContextHTTPMethod).(string),
				"url":    ctx.Value(plugins.ContextHTTPURL).(string),
			},
		},
		{
			Name:        "http_response_dns_time",
			Description: "The HTTP response DNS time",
			Value:       dnsTime,
			Labels: map[string]string{
				"method": ctx.Value(plugins.ContextHTTPMethod).(string),
				"url":    ctx.Value(plugins.ContextHTTPURL).(string),
			},
		},
		{
			Name:        "http_response_tcp_connect_time",
			Description: "The HTTP response TCP connect time",
			Value:       tcpTime,
			Labels: map[string]string{
				"method": ctx.Value(plugins.ContextHTTPMethod).(string),
				"url":    ctx.Value(plugins.ContextHTTPURL).(string),
			},
		},
		{
			Name:        "http_response_tls_handshake_time",
			Description: "The HTTP response TLS handshake time",
			Value:       tlsTime,
			Labels: map[string]string{
				"method": ctx.Value(plugins.ContextHTTPMethod).(string),
				"url":    ctx.Value(plugins.ContextHTTPURL).(string),
			},
		},
	}
	return ctx, customMetrics, err
}

// request represents a HTTP request.
func (p *HTTP) request(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	if _, ok := args["method"]; !ok {
		args["method"] = "GET"
	}
	// set context for current step
	ctx = context.WithValue(ctx, plugins.ContextHTTPMethod, args["method"])
	ctx = context.WithValue(ctx, plugins.ContextHTTPURL, args["url"])
	ctx = context.WithValue(ctx, plugins.ContextHTTPBody, args["body"])

	return p.requestByMethod(ctx, args)
}

// statusCodeShouldBe represents a HTTP status code should be.
func (p *HTTP) statusCodeShouldBe(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	var err error

	// get context for current step

	if _, ok := ctx.Value(plugins.ContextHTTPResponse).(*http.Response); !ok {
		return ctx, nil, fmt.Errorf("context doesn't have the expected value %s", plugins.ContextHTTPResponse.Name)
	}

	resp := ctx.Value(plugins.ContextHTTPResponse).(*http.Response)

	expectedStatusCode, err := strconv.ParseInt(args["statusCode"], 10, 64)

	if err != nil {
		return ctx, nil, err
	}

	if int64(resp.StatusCode) != expectedStatusCode {
		return ctx, nil, fmt.Errorf("expected status code %d but got %d", expectedStatusCode, resp.StatusCode)
	}

	return ctx, nil, err
}

// bodyShouldContain represents a HTTP body should contain.
func (p *HTTP) bodyShouldContain(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	var err error

	// get context for current step

	if _, ok := ctx.Value(plugins.ContextOutput).(string); !ok {
		return ctx, nil, fmt.Errorf("context doesn't have the expected value %s", plugins.ContextHTTPResponse.Name)
	}

	output := ctx.Value(plugins.ContextOutput).(string)

	if !strings.Contains(output, args["search"]) {
		return ctx, nil, fmt.Errorf("expected body to contain %s", args["search"])
	}

	return ctx, nil, err
}

// shouldRedirectTo represents a HTTP should redirect to.
func (p *HTTP) shouldRedirectTo(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	var err error

	// get context for current step
	if _, ok := ctx.Value(plugins.ContextHTTPResponse).(*http.Response); !ok {
		return ctx, nil, fmt.Errorf("context doesn't have the expected value %s", plugins.ContextHTTPResponse.Name)
	}

	resp := ctx.Value(plugins.ContextHTTPResponse).(*http.Response)

	if resp.Header.Get("Location") != args["url"] {
		return ctx, nil, fmt.Errorf("expected redirect to %s but got %s", args["url"], resp.Header.Get("Location"))
	}

	return ctx, nil, err
}

// addHTTPHeader represents a HTTP add header.
func (p *HTTP) addHTTPHeader(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	var err error

	// get context for current step
	if _, ok := ctx.Value(plugins.ContextHTTPHeaders).(map[string]string); !ok {
		ctx = context.WithValue(ctx, plugins.ContextHTTPHeaders, map[string]string{})
	}

	headers := ctx.Value(plugins.ContextHTTPHeaders).(map[string]string)

	headers[args["key"]] = args["value"]

	return ctx, nil, err
}

// setUserAgent represents a HTTP set user agent.
func (p *HTTP) setUserAgent(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	return p.addHTTPHeader(ctx, map[string]string{
		"key":   "User-Agent",
		"value": args["userAgent"],
	})
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
		ContextGenerator: []plugins.ContextGenerator{
			{
				Name:        plugins.ContextHTTPResponse.Name,
				Description: "The HTTP response",
			},
			{
				Name:        plugins.ContextOutput.Name,
				Description: "The HTTP response body",
			},
		},
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
		Fn: func(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
			ctx = context.WithValue(ctx, plugins.ContextHTTPTlsInsecureSkipVerify, true)
			return ctx, nil, nil
		},
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "forceIP",
		Description: "Forces the IP address to use for the request",
		Params: []plugins.StepParam{
			{Name: "ip", Description: "The IP address", Optional: false},
		},
		Fn: func(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
			ctx = context.WithValue(ctx, plugins.ContextHTTPForceIP, args["ip"])
			return ctx, nil, nil
		},
	})

}

// Init initializes the plugin.
func init() {
	h := &HTTP{}
	h.Init()
	plugins.AddPlugin("http", h)
}

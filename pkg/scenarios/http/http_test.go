package http_test

import (
	"context"
	"testing"

	"github.com/hidracloud/hidra/pkg/scenarios/http"
)

const (
	errUrlParameterMissing = "url parameter is missing"
)

func TestHTTPRequestParameters(t *testing.T) {
	// Initialize scenario
	h := &http.Scenario{}
	h.Init()

	ctx := context.TODO()

	// Create an invalid request, should be a fail.
	params := make(map[string]string)
	_, err := h.RunStep(ctx, "request", params, 0)

	if err == nil {
		t.Error(errUrlParameterMissing)
	}

	params = make(map[string]string)
	_, err = h.RunStep(ctx, "setUserAgent", params, 0)

	if err == nil {
		t.Error(errUrlParameterMissing)
	}

	params = make(map[string]string)
	_, err = h.RunStep(ctx, "addHTTPHeader", params, 0)

	if err == nil {
		t.Error(errUrlParameterMissing)
	}

	params = make(map[string]string)
	params["key"] = "test"
	_, err = h.RunStep(ctx, "addHTTPHeader", params, 0)

	if err == nil {
		t.Error(errUrlParameterMissing)
	}
}

func TestHTTPRequestTestGoogle(t *testing.T) {
	// Initialize scenario
	h := &http.Scenario{}
	h.Init()

	// Create an invalid request, should be a fail.
	params := make(map[string]string)
	params["user-agent"] = "hidra-test"

	ctx := context.TODO()

	_, err := h.RunStep(ctx, "setUserAgent", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["key"] = "accept"
	params["value"] = "text/html"

	_, err = h.RunStep(ctx, "addHTTPHeader", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["url"] = "https://example.org/"
	_, err = h.RunStep(ctx, "request", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["statusCode"] = "200"
	_, err = h.RunStep(ctx, "statusCodeShouldBe", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["search"] = "awhdvbiyf3ri"
	_, err = h.RunStep(ctx, "bodyShouldContain", params, 0)

	if err == nil {
		t.Error("not expected in body")
	}

	params = make(map[string]string)
	params["search"] = "example"
	_, err = h.RunStep(ctx, "bodyShouldContain", params, 0)

	if err != nil {
		t.Error(err)
	}

	_, err = h.RunStep(ctx, "clear", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["url"] = "http://google.com/"
	_, err = h.RunStep(ctx, "request", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["url"] = "http://www.google.com/"
	_, err = h.RunStep(ctx, "shouldRedirectTo", params, 0)

	if err != nil {
		t.Error(err)
	}
}

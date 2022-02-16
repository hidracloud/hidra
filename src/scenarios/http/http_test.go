package http_test

import (
	"testing"

	"github.com/hidracloud/hidra/src/scenarios/http"
)

func TestHTTPRequestParameters(t *testing.T) {
	// Initialize scenario
	h := &http.Scenario{}
	h.Init()

	// Create an invalid request, should be a fail.
	params := make(map[string]string)
	_, err := h.RunStep("request", params, 0)

	if err == nil {
		t.Error("url parameter missing but needed")
	}

	params = make(map[string]string)
	_, err = h.RunStep("setUserAgent", params, 0)

	if err == nil {
		t.Error("url parameter missing but needed")
	}

	params = make(map[string]string)
	_, err = h.RunStep("addHTTPHeader", params, 0)

	if err == nil {
		t.Error("url parameter missing but needed")
	}

	params = make(map[string]string)
	params["key"] = "test"
	_, err = h.RunStep("addHTTPHeader", params, 0)

	if err == nil {
		t.Error("url parameter missing but needed")
	}
}

func TestHTTPRequestTestGoogle(t *testing.T) {
	// Initialize scenario
	h := &http.Scenario{}
	h.Init()

	// Create an invalid request, should be a fail.
	params := make(map[string]string)
	params["user-agent"] = "hidra-test"

	_, err := h.RunStep("setUserAgent", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["key"] = "accept"
	params["value"] = "text/html"

	_, err = h.RunStep("addHTTPHeader", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["url"] = "https://example.org/"
	_, err = h.RunStep("request", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["statusCode"] = "200"
	_, err = h.RunStep("statusCodeShouldBe", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["search"] = "awhdvbiyf3ri"
	_, err = h.RunStep("bodyShouldContain", params, 0)

	if err == nil {
		t.Error("not expected in body")
	}

	params = make(map[string]string)
	params["search"] = "example"
	_, err = h.RunStep("bodyShouldContain", params, 0)

	if err != nil {
		t.Error(err)
	}

	h.RunStep("clear", params, 0)

	params = make(map[string]string)
	params["url"] = "http://google.com/"
	_, err = h.RunStep("request", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["url"] = "http://www.google.com/"
	_, err = h.RunStep("shouldRedirectTo", params, 0)

	if err != nil {
		t.Error(err)
	}
}

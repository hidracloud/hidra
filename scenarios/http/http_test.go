package http_test

import (
	"testing"

	"github.com/JoseCarlosGarcia95/hidra/scenarios/http"
)

func TestHttpRequestParameters(t *testing.T) {
	// Initialize scenario
	h := &http.HttpScenario{}
	h.Init()

	// Create an invalid request, should be a fail.
	params := make(map[string]string)
	err := h.RunStep("request", params)

	if err == nil {
		t.Error("url parameter missing but needed")
	}

	params = make(map[string]string)
	err = h.RunStep("setUserAgent", params)

	if err == nil {
		t.Error("url parameter missing but needed")
	}

	params = make(map[string]string)
	err = h.RunStep("addHttpHeader", params)

	if err == nil {
		t.Error("url parameter missing but needed")
	}

	params = make(map[string]string)
	params["key"] = "test"
	err = h.RunStep("addHttpHeader", params)

	if err == nil {
		t.Error("url parameter missing but needed")
	}
}

func TestHttpRequestTestGoogle(t *testing.T) {
	// Initialize scenario
	h := &http.HttpScenario{}
	h.Init()

	// Create an invalid request, should be a fail.
	params := make(map[string]string)
	params["user-agent"] = "hidra-test"

	err := h.RunStep("setUserAgent", params)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["key"] = "accept"
	params["value"] = "text/html"

	err = h.RunStep("addHttpHeader", params)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["url"] = "https://example.org/"
	err = h.RunStep("request", params)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["statusCode"] = "200"
	err = h.RunStep("statusCodeShouldBe", params)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["search"] = "awhdvbiyf3ri"
	err = h.RunStep("bodyShouldContain", params)

	if err == nil {
		t.Error("not expected in body")
	}

	params = make(map[string]string)
	params["search"] = "example"
	err = h.RunStep("bodyShouldContain", params)

	if err != nil {
		t.Error(err)
	}

	h.RunStep("clear", params)

}

package http_test

import (
	"context"
	"testing"

	"github.com/hidracloud/hidra/v3/plugins"
	"github.com/hidracloud/hidra/v3/plugins/services/http"
)

// TestRequestByMethod
func TestHTTPRequestParameters(t *testing.T) {
	h := http.HTTP{}
	h.Init()

	ctx := context.TODO()

	ctx, _, err := h.RunStep(ctx, &plugins.Step{
		Name: "request",
		Args: map[string]string{
			"method": "GET",
			"url":    "https://www.google.com",
		},
	})

	if err != nil {
		t.Error(err)
	}

	ctx, _, err = h.RunStep(ctx, &plugins.Step{
		Name: "statusCodeShouldBe",
		Args: map[string]string{
			"statusCode": "200",
		},
	})

	if err != nil {
		t.Error(err)
	}

	_, _, err = h.RunStep(ctx, &plugins.Step{
		Name: "statusCodeShouldBe",
		Args: map[string]string{
			"statusCode": "201",
		},
	})

	if err == nil {
		t.Error("expected error")
	}

}

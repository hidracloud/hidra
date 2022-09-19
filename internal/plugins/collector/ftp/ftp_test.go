package ftp_test

import (
	"context"
	"testing"

	"github.com/hidracloud/hidra/v3/internal/plugins"
	"github.com/hidracloud/hidra/v3/internal/plugins/collector/ftp"
)

func TestScenario(t *testing.T) {
	h := &ftp.FTP{}
	h.Init()

	ctx := context.TODO()

	ctx, _, err := h.RunStep(ctx, &plugins.Step{
		Name: "connectTo",
		Args: map[string]string{
			"to": "ftp.dlptest.com:21",
		},
	})

	if err != nil {
		t.Error(err)
	}

	ctx, _, err = h.RunStep(ctx, &plugins.Step{
		Name: "login",
		Args: map[string]string{
			"user":     "dlpuser",
			"password": "rNrKYTX9g7z3RgJRmxWuGHbeu",
		},
	})

	if err != nil {
		t.Error(err)
	}

	ctx, _, err = h.RunStep(ctx, &plugins.Step{
		Name: "write",
		Args: map[string]string{
			"data": "test",
			"file": "test.txt",
		},
	})

	if err != nil {
		t.Error(err)
	}

	ctx, _, err = h.RunStep(ctx, &plugins.Step{
		Name: "read",
		Args: map[string]string{
			"file": "test.txt",
		},
	})

	if err != nil {
		t.Error(err)
	}

	_, _, err = h.RunStep(ctx, &plugins.Step{
		Name: "delete",
		Args: map[string]string{
			"file": "test.txt",
		},
	})

	if err != nil {
		t.Error(err)
	}

}

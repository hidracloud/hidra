package ftp_test

import (
	"context"
	"testing"

	"github.com/hidracloud/hidra/src/scenarios/ftp"
)

func TestScenario(t *testing.T) {
	s := &ftp.Scenario{}
	s.Init()

	ctx := context.TODO()
	params := make(map[string]string)
	params["to"] = "ftp.dlptest.com:21"

	_, err := s.RunStep(ctx, "connectTo", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["user"] = "dlpuser"
	params["password"] = "rNrKYTX9g7z3RgJRmxWuGHbeu"

	_, err = s.RunStep(ctx, "login", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["data"] = "test"
	params["test-file"] = "test.txt"

	_, err = s.RunStep(ctx, "write", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["test-file"] = "test.txt"
	params["data"] = "test"

	_, err = s.RunStep(ctx, "read", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["test-file"] = "test.txt"

	_, err = s.RunStep(ctx, "delete", params, 0)

	if err != nil {
		t.Error(err)
	}
}

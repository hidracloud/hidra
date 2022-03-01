package ftp_test

import (
	"testing"

	"github.com/hidracloud/hidra/src/scenarios/ftp"
)

func TestScenario(t *testing.T) {
	s := &ftp.Scenario{}
	s.Init()

	params := make(map[string]string)
	params["to"] = "ftp.dlptest.com:21"

	_, err := s.RunStep("connectTo", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["user"] = "dlpuser"
	params["password"] = "rNrKYTX9g7z3RgJRmxWuGHbeu"

	_, err = s.RunStep("login", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["data"] = "test"
	params["test-file"] = "test.txt"

	_, err = s.RunStep("write", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["test-file"] = "test.txt"
	params["data"] = "test"

	_, err = s.RunStep("read", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["test-file"] = "test.txt"

	_, err = s.RunStep("delete", params, 0)

	if err != nil {
		t.Error(err)
	}
}

package tcp_test

import (
	"testing"

	"github.com/hidracloud/hidra/src/scenarios/tcp"
)

func TestScenario(t *testing.T) {
	// Initialize scenario
	s := &tcp.Scenario{}
	s.Init()

	params := make(map[string]string)
	params["to"] = "google.com:80"
	_, err := s.RunStep("connectTo", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["data"] = "SEVBRCAvIEhUVFAvMS4xDQoNCgo="
	_, err = s.RunStep("write", params, 0)

	if err != nil {
		t.Error(err)
	}

	params = make(map[string]string)
	params["data"] = ""
	_, err = s.RunStep("read", params, 0)

	if err != nil {
		t.Error(err)
	}
}

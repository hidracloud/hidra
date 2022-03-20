package security_test

import (
	"testing"

	"github.com/hidracloud/hidra/src/scenarios/security"
)

func TestPortscanner(t *testing.T) {
	s := &security.Scenario{}
	s.Init()

	params := make(map[string]string)
	params["hostname"] = "8.8.8.8"

	_, err := s.RunStep("portScanner", params, 0)

	if err != nil {
		t.Errorf("PortScanner failed: %s", err)
	}
}

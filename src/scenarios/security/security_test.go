package security_test

import (
	"testing"
	"time"

	"github.com/hidracloud/hidra/src/scenarios/security"
)

func TestPortscanner(t *testing.T) {
	s := &security.Scenario{}
	s.Init()

	params := make(map[string]string)
	params["hostname"] = "scanme.nmap.org"

	_, err := s.RunStep("portScanner", params, time.Second*60)

	if err != nil {
		t.Errorf("PortScanner failed: %s", err)
	}
}

package security_test

import (
	"context"
	"testing"

	"github.com/hidracloud/hidra/pkg/scenarios/security"
)

func TestPortscanner(t *testing.T) {
	s := &security.Scenario{}
	s.Init()

	params := make(map[string]string)
	params["hostname"] = "scanme.nmap.org"

	ctx := context.TODO()

	_, err := s.RunStep(ctx, "portScanner", params, 0)

	if err != nil {
		t.Errorf("PortScanner failed: %s", err)
	}
}

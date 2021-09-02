// Package all runs all scenarios
package all

import (
	// Run browser initialization
	_ "github.com/hidracloud/hidra/scenarios/browser"
	// Run http initialization
	_ "github.com/hidracloud/hidra/scenarios/http"
	// Run icmp initialization
	_ "github.com/hidracloud/hidra/scenarios/icmp"
	// Run tls initialization
	_ "github.com/hidracloud/hidra/scenarios/tls"
)

// Package all runs all scenarios
package all

import (
	// Run browser initialization
	_ "github.com/hidracloud/hidra/src/scenarios/browser"
	// Run http initialization
	_ "github.com/hidracloud/hidra/src/scenarios/http"
	// Run icmp initialization
	_ "github.com/hidracloud/hidra/src/scenarios/icmp"
	// Run tls initialization
	_ "github.com/hidracloud/hidra/src/scenarios/tls"
	// Run whois initialization
	_ "github.com/hidracloud/hidra/src/scenarios/whois"
	// Run ftp initialization
	_ "github.com/hidracloud/hidra/src/scenarios/ftp"
)

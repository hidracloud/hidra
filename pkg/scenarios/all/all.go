// Package all runs all scenarios
package all

import (
	// Run browser initialization
	_ "github.com/hidracloud/hidra/pkg/scenarios/browser"
	// Run http initialization
	_ "github.com/hidracloud/hidra/pkg/scenarios/http"
	// Run icmp initialization
	_ "github.com/hidracloud/hidra/pkg/scenarios/icmp"
	// Run tls initialization
	_ "github.com/hidracloud/hidra/pkg/scenarios/tls"
	// Run whois initialization
	_ "github.com/hidracloud/hidra/pkg/scenarios/whois"
	// Run ftp initialization
	_ "github.com/hidracloud/hidra/pkg/scenarios/ftp"
	// Run tcp initialization
	_ "github.com/hidracloud/hidra/pkg/scenarios/tcp"
	// Run udp initialization
	_ "github.com/hidracloud/hidra/pkg/scenarios/udp"
	// Run security initialization
	_ "github.com/hidracloud/hidra/pkg/scenarios/security"
)

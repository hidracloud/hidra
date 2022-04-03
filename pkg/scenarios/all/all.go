// Package all runs all scenarios
package all

import (
	// Run browser initialization
	_ "github.com/hidracloud/hidra/v2/pkg/scenarios/browser"
	// Run http initialization
	_ "github.com/hidracloud/hidra/v2/pkg/scenarios/http"
	// Run icmp initialization
	_ "github.com/hidracloud/hidra/v2/pkg/scenarios/icmp"
	// Run tls initialization
	_ "github.com/hidracloud/hidra/v2/pkg/scenarios/tls"
	// Run whois initialization
	_ "github.com/hidracloud/hidra/v2/pkg/scenarios/whois"
	// Run ftp initialization
	_ "github.com/hidracloud/hidra/v2/pkg/scenarios/ftp"
	// Run tcp initialization
	_ "github.com/hidracloud/hidra/v2/pkg/scenarios/tcp"
	// Run udp initialization
	_ "github.com/hidracloud/hidra/v2/pkg/scenarios/udp"
	// Run security initialization
	_ "github.com/hidracloud/hidra/v2/pkg/scenarios/security"
)

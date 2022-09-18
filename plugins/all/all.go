// Package all runs all scenarios
package all

import (
	// Run strings initializers
	_ "github.com/hidracloud/hidra/v3/plugins/misc/string"

	// Run http initialization
	_ "github.com/hidracloud/hidra/v3/plugins/collector/http"

	// Run http initialization
	_ "github.com/hidracloud/hidra/v3/plugins/collector/dns"

	// Run ftp initialization
	_ "github.com/hidracloud/hidra/v3/plugins/collector/ftp"

	// Run icmp initialization
	_ "github.com/hidracloud/hidra/v3/plugins/collector/icmp"

	// Run tcp initialization
	_ "github.com/hidracloud/hidra/v3/plugins/collector/tcp"

	// Run udp initialization
	_ "github.com/hidracloud/hidra/v3/plugins/collector/udp"

	// Run udp initialization
	_ "github.com/hidracloud/hidra/v3/plugins/collector/tls"

	// Run browser initialization
	_ "github.com/hidracloud/hidra/v3/plugins/collector/browser"
)

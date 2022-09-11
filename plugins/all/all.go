// Package all runs all scenarios
package all

import (
	// Run strings initializers
	_ "github.com/hidracloud/hidra/v3/plugins/misc/string"

	// Run http initialization
	_ "github.com/hidracloud/hidra/v3/plugins/services/http"

	// Run http initialization
	_ "github.com/hidracloud/hidra/v3/plugins/services/dns"

	// Run ftp initialization
	_ "github.com/hidracloud/hidra/v3/plugins/services/ftp"

	// Run ftp initialization
	_ "github.com/hidracloud/hidra/v3/plugins/services/icmp"
)

package plugins

// ContextKey represents a context key.
type ContextKey struct {
	// Name is the name of the context key.
	Name string
}

var (
	// ContextHTTPMethod is the context key for the HTTP method.
	ContextHTTPMethod = ContextKey{
		Name: "http.method",
	}
	// ContextHTTPURL is the context key for the HTTP URL.
	ContextHTTPURL = ContextKey{
		Name: "http.url",
	}
	// ContextHTTPBody is the context key for the HTTP body.
	ContextHTTPBody = ContextKey{
		Name: "http.body",
	}
	// ContextSharedJar is the context key for the shared jar.
	ContextSharedJar = ContextKey{
		Name: "http.sharedjar",
	}
	// ContextHTTPHeaders is the context key for the HTTP headers.
	ContextHTTPHeaders = ContextKey{
		Name: "http.headers",
	}
	// ContextHTTPResponse is the context key for the HTTP response.
	ContextHTTPResponse = ContextKey{
		Name: "http.response",
	}

	// ContextHTTPConnInfo is the context key for the HTTP connection info.
	ContextHTTPConnInfo = ContextKey{
		Name: "http.conninfo",
	}

	// ContextHTTPDNSStartInfo is the context key for the HTTP DNS start info.
	ContextHTTPDNSStartInfo = ContextKey{
		Name: "http.dnsstartinfo",
	}

	// ContextHTTPDNSStartTime is the context key for the HTTP DNS start time.
	ContextHTTPDNSStartTime = ContextKey{
		Name: "http.dnsstarttime",
	}

	// ContextHTTPDNSStopTime is the context key for the HTTP DNS stop time.
	ContextHTTPDNSStopTime = ContextKey{
		Name: "http.dnsstoptime",
	}

	// ContextHTTPTcpConnectStartTime is the context key for the HTTP TCP connect start time.
	ContextHTTPTcpConnectStartTime = ContextKey{
		Name: "http.tcpconnectstarttime",
	}

	// ContextHTTPTcpConnectStopTime is the context key for the HTTP TCP connect stop time.
	ContextHTTPTcpConnectStopTime = ContextKey{
		Name: "http.tcpconnectstoptime",
	}

	// ContextHTTPTlsHandshakeStartTime is the context key for the HTTP TLS handshake start time.
	ContextHTTPTlsHandshakeStartTime = ContextKey{
		Name: "http.tlshandshakestarttime",
	}

	// ContextHTTPTlsHandshakeStopTime is the context key for the HTTP TLS handshake stop time.
	ContextHTTPTlsHandshakeStopTime = ContextKey{
		Name: "http.tlshandshakestoptime",
	}

	// ContextHTTPDNSDoneInfo is the context key for the HTTP DNS done info.
	ContextHTTPDNSDoneInfo = ContextKey{
		Name: "http.dnsdoneinfo",
	}

	// ContextHTTPTlsInsecureSkipVerify is the context key for the HTTP TLS insecure skip verify.
	ContextHTTPTlsInsecureSkipVerify = ContextKey{
		Name: "http.tlsinsecureskipverify",
	}

	// ContextHTTPForceIP is the context key for the HTTP force IP.
	ContextHTTPForceIP = ContextKey{
		Name: "http.forceip",
	}

	// ContextHTTPNetwork is the context key for the HTTP network.
	ContextHTTPNetwork = ContextKey{
		Name: "http.network",
	}

	// ContextHTTPAddr is the context key for the HTTP address.
	ContextHTTPAddr = ContextKey{
		Name: "http.addr",
	}

	// ContextOutput is the context key for the output.
	ContextOutput = ContextKey{
		Name: "output",
	}

	// ContextTimeout is the context key for the timeout.
	ContextTimeout = ContextKey{
		Name: "timeout",
	}

	// ContextHTTPClient is the context key for the HTTP client.
	ContextHTTPClient = ContextKey{
		Name: "http.client",
	}
)

func (c *ContextKey) String() string {
	return c.Name
}

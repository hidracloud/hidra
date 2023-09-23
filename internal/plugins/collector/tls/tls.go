package tls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/internal/misc"
	"github.com/hidracloud/hidra/v3/internal/plugins"
	"github.com/hidracloud/hidra/v3/internal/utils"
)

// TLS represents a TLS plugin.
type TLS struct {
	plugins.BasePlugin
}

// connectTo connects to a TLS server.
func (p *TLS) connectTo(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	if _, ok := stepsgen[misc.ContextTLSConnection].(*tls.Conn); ok {
		err := stepsgen[misc.ContextTLSConnection].(*tls.Conn).Close()

		if err != nil {
			return nil, err
		}
	}

	// Check if host has protocol at the beginning, if yes, remove it
	if len(args["to"]) > 8 && args["to"][:8] == "https://" {
		args["to"] = args["to"][8:]
	}

	// check if has path at the end, if yes, remove it
	if strings.Contains(args["to"], "/") {
		parts := strings.Split(args["to"], "/")
		args["to"] = parts[0]
	}

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	startTime := time.Now()

	// if a port is not specified, use 443
	if !strings.Contains(args["to"], ":") {
		args["to"] = args["to"] + ":443"
	}

	conn, err := tls.Dial("tcp", args["to"], conf)

	if err != nil {
		return nil, err
	}

	certificates := conn.ConnectionState().PeerCertificates

	stepsgen[misc.ContextTLSConnection] = conn
	stepsgen[misc.ContextTLSCertificates] = certificates
	stepsgen[misc.ContextTLSHost] = args["to"]

	customMetrics := []*metrics.Metric{
		{
			Name: "tls_version",
			Labels: map[string]string{
				"host": args["to"],
			},
			Value:       float64(conn.ConnectionState().Version),
			Purge:       true,
			PurgeLabels: []string{"host", "subject"},
		},
		{
			Name: "tls_cipher_suite",
			Labels: map[string]string{
				"host": args["to"],
			},
			Value:       float64(conn.ConnectionState().CipherSuite),
			Purge:       true,
			PurgeLabels: []string{"host", "subject"},
		},
		{
			Name: "tls_handshake_duration_seconds",
			Labels: map[string]string{
				"host": args["to"],
			},
			Purge:       true,
			PurgeLabels: []string{"host", "subject"},
			Value:       time.Since(startTime).Seconds(),
		},
	}

	for _, certificate := range certificates {
		customMetrics = append(customMetrics, &metrics.Metric{
			Name: "tls_certificate_not_after",
			Labels: map[string]string{
				"serial_number": certificate.SerialNumber.String(),
				"subject":       certificate.Subject.String(),
				"host":          args["to"],
			},
			Purge:       true,
			PurgeLabels: []string{"host", "subject"},
			Value:       float64(certificate.NotAfter.Unix()),
		})

		customMetrics = append(customMetrics, &metrics.Metric{
			Name: "tls_certificate_not_before",
			Labels: map[string]string{
				"serial_number": certificate.SerialNumber.String(),
				"subject":       certificate.Subject.String(),
				"host":          args["to"],
			},
			Purge:       true,
			PurgeLabels: []string{"host", "subject"},
			Value:       float64(certificate.NotBefore.Unix()),
		})

		customMetrics = append(customMetrics, &metrics.Metric{
			Name: "tls_certificate_version",
			Labels: map[string]string{
				"serial_number": certificate.SerialNumber.String(),
				"subject":       certificate.Subject.String(),
				"host":          args["to"],
			},
			Value:       float64(certificate.Version),
			Purge:       true,
			PurgeLabels: []string{"host", "subject"},
		})
	}

	return customMetrics, nil
}

// onClose closes the connection.
func (p *TLS) onClose(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	if _, ok := stepsgen[misc.ContextTLSConnection].(*tls.Conn); ok {
		err := stepsgen[misc.ContextTLSConnection].(*tls.Conn).Close()

		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// dnsShouldBePresent checks if a DNS record should be present.
func (p *TLS) dnsShouldBePresent(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	if _, ok := stepsgen[misc.ContextTLSCertificates].([]*x509.Certificate); !ok {
		return nil, fmt.Errorf("no TLS connection found")
	}

	certificates := stepsgen[misc.ContextTLSCertificates].([]*x509.Certificate)

	for _, cert := range certificates {
		for _, dns := range cert.DNSNames {
			matched, err := filepath.Match(dns, args["dns"])

			if err != nil {
				return nil, err
			}

			if matched {
				return nil, nil
			}
		}
	}
	return nil, fmt.Errorf("DNS name %s should not be present", args["dns"])
}

// shouldBeValidFor checks if a certificate is valid for a given host.
func (p *TLS) shouldBeValidFor(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	if _, ok := stepsgen[misc.ContextTLSCertificates].([]*x509.Certificate); !ok {
		return nil, fmt.Errorf("no TLS connection found")
	}

	duration, err := utils.ParseDuration(args["for"])
	if err != nil {
		return nil, err
	}

	certificates := stepsgen[misc.ContextTLSCertificates].([]*x509.Certificate)
	limitDate := time.Now().Add(duration)

	for _, cert := range certificates {
		if limitDate.After(cert.NotAfter) {
			return nil, fmt.Errorf("certificate will be invalid after %s, and your limit date is %s", cert.NotAfter, limitDate)
		}
	}

	return nil, nil
}

// Init initializes the plugin.
func (p *TLS) Init() {
	p.Primitives()

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "connectTo",
		Description: "Connects to a TLS server",
		Params: []plugins.StepParam{
			{
				Name:        "to",
				Description: "Host to connect to",
				Optional:    false,
			},
		},
		Fn: p.connectTo,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "dnsShouldBePresent",
		Description: "Checks if a DNS record should be present",
		Params: []plugins.StepParam{
			{
				Name:        "dns",
				Description: "DNS name to check",
				Optional:    false,
			},
		},
		Fn: p.dnsShouldBePresent,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "shouldBeValidFor",
		Description: "Checks if a certificate is valid for a given host",
		Params: []plugins.StepParam{
			{
				Name:        "for",
				Description: "Duration for which the certificate should be valid",
				Optional:    false,
			},
		},
		Fn: p.shouldBeValidFor,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "onClose",
		Description: "Close the connection",
		Params:      []plugins.StepParam{},
		Fn:          p.onClose,
	})
}

// Init initializes the plugin.
func init() {
	h := &TLS{}
	h.Init()
	plugins.AddPlugin("tls", "TLS plugin is used to check TLS certificates", h)
}

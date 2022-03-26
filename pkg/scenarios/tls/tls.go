package tls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/hidracloud/hidra/pkg/models"
	"github.com/hidracloud/hidra/pkg/scenarios"
	"github.com/hidracloud/hidra/pkg/utils"
)

// Scenario Represent an ssl scenario
type Scenario struct {
	models.Scenario

	certificates []*x509.Certificate
}

func (s *Scenario) connectTo(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", c["to"], conf)

	if err != nil {
		return nil, err
	}

	defer conn.Close()
	s.certificates = conn.ConnectionState().PeerCertificates

	return nil, nil
}

// RCA generate RCAs for scenario
func (s *Scenario) RCA(result *models.ScenarioResult) error {
	log.Println("TLS RCA")
	return nil
}

func (s *Scenario) dnsShouldBePresent(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	if s.certificates == nil {
		return nil, fmt.Errorf("you should connect to an addr first")
	}

	for _, cert := range s.certificates {
		for _, dns := range cert.DNSNames {
			matched, err := filepath.Match(dns, c["dns"])

			if err != nil {
				return nil, err
			}

			if matched {
				return nil, nil
			}
		}
	}

	return nil, fmt.Errorf("dns missing")
}

func (s *Scenario) shouldBeValidFor(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	duration, err := utils.ParseDuration(c["for"])

	if err != nil {
		return nil, err
	}

	limitDate := time.Now().Add(duration)
	for _, cert := range s.certificates {
		if limitDate.After(cert.NotAfter) {
			return nil, fmt.Errorf("certificate will be invalid after %s, and your limit date is %s", cert.NotAfter, limitDate)
		}
	}
	return nil, nil
}

func (s *Scenario) dumpMetrics(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	customMetrics := make([]models.Metric, 0)

	for _, cert := range s.certificates {
		customMetrics = append(customMetrics, models.Metric{
			Name:  "certificate_not_after",
			Value: float64(cert.NotAfter.Unix()),
			Labels: map[string]string{
				"subject": cert.Subject.CommonName,
			},
		})

		customMetrics = append(customMetrics, models.Metric{
			Name:  "certificate_not_before",
			Value: float64(cert.NotBefore.Unix()),
			Labels: map[string]string{
				"subject": cert.Subject.CommonName,
			},
		})

		customMetrics = append(customMetrics, models.Metric{
			Name:  "certificate_serial_number",
			Value: float64(cert.SerialNumber.Int64()),
			Labels: map[string]string{
				"subject": cert.Subject.CommonName,
			},
		})
	}

	return customMetrics, nil
}

// Description return the description of the scenario
func (s *Scenario) Description() string {
	return "Run a TLS scenario"
}

// Close closes the scenario
func (s *Scenario) Close() {

}

// Init initialize the scenario
func (s *Scenario) Init() {
	s.StartPrimitives()

	s.RegisterStep("connectTo", models.StepDefinition{
		Description: "Connect to a host",
		Params: []models.StepParam{
			{
				Name:        "to",
				Description: "Host to connect to",
				Optional:    false,
			},
		},
		Fn: s.connectTo,
	})

	s.RegisterStep("dnsShouldBePresent", models.StepDefinition{
		Description: "Check if the dns is present in the certificate",
		Params: []models.StepParam{
			{
				Name:        "dns",
				Description: "DNS to check",
				Optional:    false,
			},
		},
		Fn: s.dnsShouldBePresent,
	})

	s.RegisterStep("dumpMetrics", models.StepDefinition{
		Description: "Dump metrics",
		Params:      []models.StepParam{},
		Fn:          s.dumpMetrics,
	})

	s.RegisterStep("shouldBeValidFor", models.StepDefinition{
		Description: "Check if the certificate is valid for a given duration",
		Params: []models.StepParam{
			{
				Name:        "for",
				Description: "Duration to check",
				Optional:    false,
			},
		},
		Fn: s.shouldBeValidFor,
	})
}

func init() {
	scenarios.Add("tls", func() models.IScenario {
		return &Scenario{}
	})
}

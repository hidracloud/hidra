package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"path/filepath"
	"time"

	"github.com/hidracloud/hidra/models"
	"github.com/hidracloud/hidra/scenarios"
	"github.com/hidracloud/hidra/utils"
)

// Represent an ssl scenario
type TLSScneario struct {
	models.Scenario

	certificates []*x509.Certificate
}

func (s *TLSScneario) connectTo(c map[string]string) error {
	if _, ok := c["to"]; !ok {
		return fmt.Errorf("to parameter missing")
	}

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", c["to"], conf)

	if err != nil {
		return err
	}

	defer conn.Close()
	s.certificates = conn.ConnectionState().PeerCertificates

	return nil
}

func (s *TLSScneario) dnsShouldBePresent(c map[string]string) error {
	if _, ok := c["dns"]; !ok {
		return fmt.Errorf("dns parameter missing")
	}

	if s.certificates == nil {
		return fmt.Errorf("you should connect to an addr first")
	}

	for _, cert := range s.certificates {
		for _, dns := range cert.DNSNames {
			matched, err := filepath.Match(dns, c["dns"])

			if err != nil {
				return err
			}

			if matched {
				return nil
			}
		}
	}

	return fmt.Errorf("dns missing")
}

func (s *TLSScneario) shouldBeValidFor(c map[string]string) error {
	if _, ok := c["for"]; !ok {
		return fmt.Errorf("for parameter missing")
	}

	duration, err := utils.ParseDuration(c["for"])

	if err != nil {
		return err
	}

	limitDate := time.Now().Add(duration)
	for _, cert := range s.certificates {
		if limitDate.After(cert.NotAfter) {
			return fmt.Errorf("certificate will be invalid after %s, and your limit date is %s", cert.NotAfter, limitDate)
		}
	}
	return nil
}

func (s *TLSScneario) Init() {
	s.StartPrimitives()

	s.RegisterStep("connectTo", s.connectTo)
	s.RegisterStep("dnsShouldBePresent", s.dnsShouldBePresent)
	s.RegisterStep("shouldBeValidFor", s.shouldBeValidFor)
}

func init() {
	scenarios.Add("ssl", func() models.IScenario {
		return &TLSScneario{}
	})
}

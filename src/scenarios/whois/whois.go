package whois

import (
	"fmt"
	"log"
	"time"

	"github.com/hidracloud/hidra/src/models"
	"github.com/hidracloud/hidra/src/scenarios"
	"github.com/hidracloud/hidra/src/utils"
	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
)

// Scenario Represent an ssl scenario
type Scenario struct {
	models.Scenario
	domain    string
	whoisInfo *whoisparser.WhoisInfo
}

func (s *Scenario) whoisFrom(c map[string]string) ([]models.Metric, error) {
	if c["domain"] == "" {
		return nil, fmt.Errorf("domain is required")
	}

	whoisResult, err := whois.Whois(c["domain"])

	if err != nil {
		return nil, err
	}

	result, err := whoisparser.Parse(whoisResult)

	if err != nil {
		return nil, err
	}

	s.whoisInfo = &result
	s.domain = c["domain"]

	return nil, nil
}

// RCA generate RCAs for scenario
func (h *Scenario) RCA(result *models.ScenarioResult) error {
	log.Println("WHOIS RCA")
	return nil
}

func (s *Scenario) dumpMetrics(c map[string]string) ([]models.Metric, error) {
	customMetrics := make([]models.Metric, 0)

	dateFormat := "2006-01-02T15:04:05.999Z"

	if c["dateFormat"] != "" {
		dateFormat = c["dateFormat"]
	}

	expirationDate, err := time.Parse(dateFormat, s.whoisInfo.Domain.ExpirationDate)

	if err != nil {
		return nil, err
	}

	customMetrics = append(customMetrics, models.Metric{
		Name:  "whois_expiration_date",
		Value: float64(expirationDate.Unix()),
		Labels: map[string]string{
			"domain": s.domain,
		},
	})

	return customMetrics, nil
}

func (s *Scenario) shouldBeValidFor(c map[string]string) ([]models.Metric, error) {
	duration, err := utils.ParseDuration(c["for"])

	if err != nil {
		return nil, err
	}

	limitDate := time.Now().Add(duration)

	dateFormat := "2006-01-02T15:04:05.999Z"

	if c["dateFormat"] != "" {
		dateFormat = c["dateFormat"]
	}

	expirationDate, err := time.Parse(dateFormat, s.whoisInfo.Domain.ExpirationDate)

	if err != nil {
		return nil, err
	}

	if limitDate.After(expirationDate) {
		return nil, fmt.Errorf("domain will be invalid after %s, and your limit date is %s", expirationDate, limitDate)
	}

	return nil, nil
}

// Description return the description of the scenario
func (s *Scenario) Description() string {
	return "Run a Whois scenario"
}

// Init initialize the scenario
func (s *Scenario) Init() {
	s.StartPrimitives()

	s.RegisterStep("whoisFrom", models.StepDefinition{
		Description: "Get whois from domain",
		Params: []models.StepParam{
			{
				Name:        "domain",
				Description: "Domain to get whois",
				Optional:    false,
			},
		},
		Fn: s.whoisFrom,
	})

	s.RegisterStep("shouldBeValidFor", models.StepDefinition{
		Description: "Check if domain is valid for a given duration",
		Params: []models.StepParam{
			{
				Name:        "for",
				Description: "Duration to check",
				Optional:    false,
			},
			{
				Name:        "dateFormat",
				Description: "Date format to parse, default is 2006-01-02T15:04:05.999Z",
				Optional:    true,
			},
		},
		Fn: s.shouldBeValidFor,
	})

	s.RegisterStep("dumpMetrics", models.StepDefinition{
		Description: "Dump metrics",
		Params:      []models.StepParam{},
		Fn:          s.dumpMetrics,
	})
}

func init() {
	scenarios.Add("whois", func() models.IScenario {
		return &Scenario{}
	})
}

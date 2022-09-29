package dns

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/internal/misc"
	"github.com/hidracloud/hidra/v3/internal/plugins"
	"github.com/hidracloud/hidra/v3/internal/utils"

	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"

	"github.com/StalkR/dnssec-analyzer/dnssec"
)

// DNS represents a DNS plugin.
type DNS struct {
	plugins.BasePlugin
}

// whoisFrom returns the whois information from a domain.
func (p *DNS) whoisFrom(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	whoisResult, err := whois.Whois(args["domain"])

	if err != nil {
		return nil, err
	}

	result, err := whoisparser.Parse(whoisResult)

	if err != nil {
		return nil, err
	}

	stepsgen[misc.ContextDNSInfo] = &result

	dateFormat := "2006-01-02T15:04:05.999Z"

	if args["dateFormat"] != "" {
		dateFormat = args["dateFormat"]
	}

	expirationDate, err := time.Parse(dateFormat, result.Domain.ExpirationDate)

	if err != nil {
		return nil, err
	}

	customMetrics := []*metrics.Metric{
		{
			Name:  "whois_expiration_date",
			Value: float64(expirationDate.Unix()),
			Labels: map[string]string{
				"domain": args["domain"],
			},
		},
	}

	return customMetrics, nil
}

// shouldBeValidFor checks if the domain is valid for a given number of duration.
func (p *DNS) shouldBeValidFor(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	duration, err := utils.ParseDuration(args["for"])

	if err != nil {
		return nil, err
	}

	dateFormat := "2006-01-02T15:04:05.999Z"

	if args["dateFormat"] != "" {
		dateFormat = args["dateFormat"]
	}

	result := stepsgen[misc.ContextDNSInfo].(*whoisparser.WhoisInfo)

	if result == nil {
		return nil, fmt.Errorf("whois info not found")
	}

	expirationDate, err := time.Parse(dateFormat, result.Domain.ExpirationDate)

	if err != nil {
		return nil, err
	}

	if expirationDate.Before(time.Now().Add(duration)) {
		return nil, fmt.Errorf("domain is not valid for %d days", duration)
	}

	return nil, nil
}

// dnsSecShouldBeValid checks if the domain has DNSSEC enabled.
func (p *DNS) dnsSecShouldBeValid(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	analysis, err := dnssec.Analyze(args["domain"])

	if err != nil {
		return nil, err
	}

	if strings.Contains(analysis.String(), "No DNSKEY records found") {
		return nil, nil
	}

	if analysis.Status() != dnssec.OK {
		return nil, fmt.Errorf("domain has invalid dnssec configuration")
	}

	return nil, nil
}

// Init initializes the plugin.
func (p *DNS) Init() {
	p.Primitives()

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "whoisFrom",
		Description: "Gets the whois information from a domain",
		Params: []plugins.StepParam{
			{
				Name:        "domain",
				Description: "The domain to get the whois information",
				Optional:    false,
			},
			{
				Name:        "dateFormat",
				Description: "Date format to parse, default is 2006-01-02T15:04:05.999Z",
				Optional:    true,
			},
		},
		Fn: p.whoisFrom,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "shouldBeValidFor",
		Description: "Checks if the domain is valid for a given number of duration",
		Params: []plugins.StepParam{
			{
				Name:        "for",
				Description: "The duration to check if the domain is valid",
				Optional:    false,
			},
			{
				Name:        "dateFormat",
				Description: "Date format to parse, default is 2006-01-02T15:04:05.999Z",
				Optional:    true,
			},
		},
		Fn: p.shouldBeValidFor,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "dnsSecShouldBeValid",
		Description: "Checks if the domain has DNSSEC enabled",
		Params: []plugins.StepParam{
			{
				Name:        "domain",
				Description: "The domain to check if DNSSEC is enabled",
				Optional:    false,
			},
		},
		Fn: p.dnsSecShouldBeValid,
	})

}

// Init initializes the plugin.
func init() {
	h := &DNS{}
	h.Init()
	plugins.AddPlugin("dns", h)
}

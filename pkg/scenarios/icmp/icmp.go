// Package icmp implements a plugin that runs a traceroute and ping command
package icmp

import (
	"context"
	"fmt"
	"log"
	"os/user"
	"strconv"
	"time"

	"github.com/aeden/traceroute"
	"github.com/go-ping/ping"
	"github.com/hidracloud/hidra/pkg/models"
	"github.com/hidracloud/hidra/pkg/scenarios"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Scenario Represent an ICMP scenario
type Scenario struct {
	models.Scenario
}

func (h *Scenario) traceroute(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	options := traceroute.TracerouteOptions{}
	options.SetRetries(0)
	options.SetMaxHops(traceroute.DEFAULT_MAX_HOPS + 1)
	options.SetFirstHop(traceroute.DEFAULT_FIRST_HOP)

	ctx, span := otel.Tracer("icmp").Start(ctx, "traceroute", trace.WithAttributes(
		attribute.String("hostname", c["hostname"]),
	))
	defer span.End()

	zz := make(chan traceroute.TracerouteHop)
	go func(ctx1 context.Context) {
		ctx, span := otel.Tracer("traceroute").Start(ctx1, "calculateTraceroute")

		defer span.End()
		for {
			var hopSpan trace.Span

			hop, ok := <-zz
			ctx, hopSpan = otel.Tracer("calculateTraceroute").Start(ctx, "hop", trace.WithAttributes(
				attribute.String("host", hop.Host),
				attribute.Int64("ttl", int64(hop.TTL)),
				attribute.String("ip", hop.AddressString()),
			))

			defer hopSpan.End()
			if !ok {
				return
			}
		}
	}(ctx)
	tcrresult, err := traceroute.Traceroute(c["hostname"], &options, zz)

	if err != nil {
		return nil, err
	}

	customMetrics := make([]models.Metric, 0)

	now := time.Now()

	for i := 0; i < len(tcrresult.Hops); i++ {
		customMetrics = append(customMetrics, models.Metric{
			Name:  fmt.Sprintf("hop_%d_elapsed", i),
			Value: float64(tcrresult.Hops[i].ElapsedTime.Milliseconds()),
			Labels: map[string]string{
				"host": tcrresult.Hops[i].Host,
				"ip":   fmt.Sprintf("%v.%v.%v.%v", tcrresult.Hops[i].Address[0], tcrresult.Hops[i].Address[1], tcrresult.Hops[i].Address[2], tcrresult.Hops[i].Address[3]),
			},
			Description: "time to completed hop",
			Expires:     time.Duration(now.Add(time.Minute * 5).Unix()),
		})

		status := 0

		if tcrresult.Hops[i].Success {
			status = 1
		}

		customMetrics = append(customMetrics, models.Metric{
			Name:  fmt.Sprintf("hop_%d_status", i),
			Value: float64(status),
			Labels: map[string]string{
				"host": tcrresult.Hops[i].Host,
				"ip":   fmt.Sprintf("%v.%v.%v.%v", tcrresult.Hops[i].Address[0], tcrresult.Hops[i].Address[1], tcrresult.Hops[i].Address[2], tcrresult.Hops[i].Address[3]),
			},
			Description: "hop status",
			Expires:     time.Duration(now.Add(time.Minute * 5).Unix()),
		})
	}

	customMetrics = append(customMetrics, models.Metric{
		Name:        "hops",
		Value:       float64(len(tcrresult.Hops)),
		Description: "number of hops",
		Expires:     time.Duration(now.Add(time.Minute * 5).Unix()),
	})

	return customMetrics, nil
}

// RCA generate RCAs for scenario
func (h *Scenario) RCA(result *models.ScenarioResult) error {
	log.Println("ICMP RCA")
	return nil
}

func (h *Scenario) ping(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	if _, ok := c["hostname"]; !ok {
		return nil, fmt.Errorf("hostname parameter missing")
	}

	pinger, err := ping.NewPinger(c["hostname"])
	if err != nil {
		return nil, err
	}

	pinger.Count = 3
	if _, ok := c["times"]; ok {
		tmp, err := strconv.Atoi(c["times"])

		if err != nil {
			return nil, err
		}

		pinger.Count = tmp
	}

	currentUser, err := user.Current()

	if err != nil {
		return nil, err
	}

	if currentUser.Uid == "0" {
		pinger.SetPrivileged(true)
	}

	err = pinger.Run() // Blocks until finished.
	if err != nil {
		return nil, err
	}

	stats := pinger.Statistics()

	customMetrics := make([]models.Metric, 0)

	customMetrics = append(customMetrics, models.Metric{
		Name:        "packet_loss",
		Value:       stats.PacketLoss,
		Description: "number of lost packets",
	})

	customMetrics = append(customMetrics, models.Metric{
		Name:        "min_rtt",
		Value:       float64(stats.MinRtt.Milliseconds()),
		Description: "min ping",
	})

	customMetrics = append(customMetrics, models.Metric{
		Name:        "max_rtt",
		Value:       float64(stats.MinRtt.Milliseconds()),
		Description: "max ping",
	})

	customMetrics = append(customMetrics, models.Metric{
		Name:        "avg_rtt",
		Value:       float64(stats.AvgRtt.Milliseconds()),
		Description: "avg ping",
	})

	customMetrics = append(customMetrics, models.Metric{
		Name:        "packet_duplicates",
		Value:       float64(stats.PacketsRecvDuplicates),
		Description: "duplicate packets",
	})

	customMetrics = append(customMetrics, models.Metric{
		Name:        "packet_receive",
		Value:       float64(stats.PacketsRecv),
		Description: "packets received",
	})

	customMetrics = append(customMetrics, models.Metric{
		Name:        "packet_send",
		Value:       float64(stats.PacketsSent),
		Description: "packets send",
	})

	return customMetrics, nil
}

// Description returns the scenario description
func (h *Scenario) Description() string {
	return "Run a ICMP scenario"
}

// Close closes the scenario
func (h *Scenario) Close() {
	// nothing to do
}

// Init initializes the scenario
func (h *Scenario) Init() {
	h.StartPrimitives()

	h.RegisterStep("ping", models.StepDefinition{
		Description: "Run a ICMP ping",
		Params: []models.StepParam{
			{Name: "hostname", Description: "Hostname to ping", Optional: false},
		},
		Fn: h.ping,
	})

	h.RegisterStep("traceroute", models.StepDefinition{
		Description: "Run a ICMP traceroute",
		Params: []models.StepParam{
			{Name: "hostname", Description: "Hostname to traceroute", Optional: false},
		},
		Fn: h.traceroute,
	})
}

func init() {
	scenarios.Add("icmp", func() models.IScenario {
		return &Scenario{}
	})
}

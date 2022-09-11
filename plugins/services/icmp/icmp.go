package icmp

import (
	"context"
	"os/user"
	"time"

	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/plugins"

	"github.com/go-ping/ping"
)

const (
	// PingerCount is the number of pings to send.
	PingerCount = 10
)

// ICMP represents a ICMP plugin.
type ICMP struct {
	plugins.BasePlugin
}

// ping sends a ping to a host.
func (p *ICMP) ping(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	pinger, err := ping.NewPinger(args["hostname"])
	if err != nil {
		return ctx, nil, err
	}

	timeout := 30 * time.Second

	if _, ok := ctx.Value(plugins.ContextTimeout).(time.Duration); ok {
		timeout = ctx.Value(plugins.ContextTimeout).(time.Duration)
	}

	pinger.Count = PingerCount
	pinger.Timeout = timeout

	currentUser, err := user.Current()

	if err != nil {
		return ctx, nil, err
	}

	if currentUser.Uid == "0" {
		pinger.SetPrivileged(true)
	}

	err = pinger.Run() // Blocks until finished.
	if err != nil {
		return ctx, nil, err
	}

	stats := pinger.Statistics()

	customMetrics := make([]*metrics.Metric, 0)

	customMetrics = append(customMetrics, &metrics.Metric{
		Name:        "packet_loss",
		Value:       stats.PacketLoss,
		Description: "number of lost packets",
		Labels: map[string]string{
			"hostname": args["hostname"],
		},
	})

	customMetrics = append(customMetrics, &metrics.Metric{
		Name:        "min_rtt",
		Value:       float64(stats.MinRtt.Milliseconds()),
		Description: "min ping",
		Labels: map[string]string{
			"hostname": args["hostname"],
		},
	})

	customMetrics = append(customMetrics, &metrics.Metric{
		Name:        "max_rtt",
		Value:       float64(stats.MinRtt.Milliseconds()),
		Description: "max ping",
		Labels: map[string]string{
			"hostname": args["hostname"],
		},
	})

	customMetrics = append(customMetrics, &metrics.Metric{
		Name:        "avg_rtt",
		Value:       float64(stats.AvgRtt.Milliseconds()),
		Description: "avg ping",
		Labels: map[string]string{
			"hostname": args["hostname"],
		},
	})

	customMetrics = append(customMetrics, &metrics.Metric{
		Name:        "packet_duplicates",
		Value:       float64(stats.PacketsRecvDuplicates),
		Description: "duplicate packets",
		Labels: map[string]string{
			"hostname": args["hostname"],
		},
	})

	customMetrics = append(customMetrics, &metrics.Metric{
		Name:        "packet_receive",
		Value:       float64(stats.PacketsRecv),
		Description: "packets received",
		Labels: map[string]string{
			"hostname": args["hostname"],
		},
	})

	customMetrics = append(customMetrics, &metrics.Metric{
		Name:        "packet_send",
		Value:       float64(stats.PacketsSent),
		Description: "packets send",
		Labels: map[string]string{
			"hostname": args["hostname"],
		},
	})

	return ctx, customMetrics, nil
}

// Init initializes the plugin.
func (p *ICMP) Init() {
	p.Primitives()

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "ping",
		Description: "Ping a host",
		Params: []plugins.StepParam{
			{
				Name:        "hostname",
				Description: "Hostname to ping",
				Optional:    false,
			},
		},
		Fn: p.ping,
	})
}

// Init initializes the plugin.
func init() {
	h := &ICMP{}
	h.Init()
	plugins.AddPlugin("icmp", h)
}

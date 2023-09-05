package icmp

import (
	"context"
	"fmt"
	"net"
	"os/user"
	"time"

	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/internal/misc"
	"github.com/hidracloud/hidra/v3/internal/plugins"

	"github.com/go-ping/ping"

	"github.com/pixelbender/go-traceroute/traceroute"

	log "github.com/sirupsen/logrus"
)

const (
	// PingerCount is the number of pings to send.
	PingerCount = 3
)

// ICMP represents a ICMP plugin.
type ICMP struct {
	plugins.BasePlugin
}

// ping sends a ping to a host.
func (p *ICMP) ping(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	pinger, err := ping.NewPinger(args["hostname"])
	if err != nil {
		return nil, err
	}

	timeout := 30 * time.Second

	if _, ok := stepsgen[misc.ContextTimeout].(time.Duration); ok {
		timeout = stepsgen[misc.ContextTimeout].(time.Duration)
	}

	pinger.Count = PingerCount
	pinger.Timeout = timeout

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

	customMetrics := make([]*metrics.Metric, 0)

	customMetrics = append(customMetrics, &metrics.Metric{
		Name:        "icmp_ping_packet_loss",
		Value:       stats.PacketLoss,
		Description: "number of lost packets",
		Labels: map[string]string{
			"hostname": args["hostname"],
		},
	})

	customMetrics = append(customMetrics, &metrics.Metric{
		Name:        "icmp_ping_min_rtt",
		Value:       float64(stats.MinRtt.Milliseconds()),
		Description: "min ping",
		Labels: map[string]string{
			"hostname": args["hostname"],
		},
	})

	customMetrics = append(customMetrics, &metrics.Metric{
		Name:        "icmp_ping_max_rtt",
		Value:       float64(stats.MinRtt.Milliseconds()),
		Description: "max ping",
		Labels: map[string]string{
			"hostname": args["hostname"],
		},
	})

	customMetrics = append(customMetrics, &metrics.Metric{
		Name:        "icmp_ping_rtt",
		Value:       float64(stats.AvgRtt.Milliseconds()),
		Description: "avg ping",
		Labels: map[string]string{
			"hostname": args["hostname"],
		},
	})

	customMetrics = append(customMetrics, &metrics.Metric{
		Name:        "icmp_ping_packet_duplicates",
		Value:       float64(stats.PacketsRecvDuplicates),
		Description: "duplicate packets",
		Labels: map[string]string{
			"hostname": args["hostname"],
		},
	})

	customMetrics = append(customMetrics, &metrics.Metric{
		Name:        "icmp_ping_packet_receive",
		Value:       float64(stats.PacketsRecv),
		Description: "packets received",
		Labels: map[string]string{
			"hostname": args["hostname"],
		},
	})

	customMetrics = append(customMetrics, &metrics.Metric{
		Name:        "icmp_ping_packet_send",
		Value:       float64(stats.PacketsSent),
		Description: "packets send",
		Labels: map[string]string{
			"hostname": args["hostname"],
		},
	})

	return customMetrics, nil
}

// traceroute sends a traceroute to a host.
func (p *ICMP) traceroute(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	ipAddresses, err := net.LookupIP(args["hostname"])
	if err != nil {
		return nil, err
	}

	timeout := 30 * time.Second

	if _, ok := stepsgen[misc.ContextTimeout].(time.Duration); ok {
		timeout = stepsgen[misc.ContextTimeout].(time.Duration)
	}

	traceroute.DefaultConfig.Timeout = timeout

	hops, err := traceroute.Trace(ipAddresses[0])
	if err != nil {
		return nil, fmt.Errorf("error while tracing route, maybe we don't have permissions: %s", err)
	}

	maxDistance := 0

	customMetrics := make([]*metrics.Metric, 0)

	for _, h := range hops {
		for _, n := range h.Nodes {
			log.Debugf("%d. %v %v", h.Distance, n.IP, n.RTT)

			if h.Distance > maxDistance {
				maxDistance = h.Distance
			}

			customMetrics = append(customMetrics, &metrics.Metric{
				Name:        "icmp_traceroute_rtt",
				Value:       float64(n.RTT[0].Milliseconds()),
				Description: "traceroute rtt",
				Labels: map[string]string{
					"hostname": args["hostname"],
					"hop":      n.IP.String(),
				},
				Purge:       true,
				PurgeLabels: []string{"hostname"},
			})

			customMetrics = append(customMetrics, &metrics.Metric{
				Name:        "icmp_traceroute_distance",
				Value:       float64(h.Distance),
				Description: "traceroute distance",
				Labels: map[string]string{
					"hostname": args["hostname"],
					"hop":      n.IP.String(),
				},
				Purge:       true,
				PurgeLabels: []string{"hostname"},
			})

			customMetrics = append(customMetrics, &metrics.Metric{
				Name:        "icmp_traceroute_ip",
				Value:       1,
				Description: "traceroute ip",
				Labels: map[string]string{
					"hostname": args["hostname"],
					"hop":      n.IP.String(),
				},
				Purge:       true,
				PurgeLabels: []string{"hostname"},
			})

			customMetrics = append(customMetrics, &metrics.Metric{
				Name:        "icmp_traceroute_last_time",
				Value:       float64(time.Now().Unix()),
				Description: "traceroute last time",
				Labels: map[string]string{
					"hostname": args["hostname"],
					"hop":      n.IP.String(),
				},
				Purge:       true,
				PurgeLabels: []string{"hostname"},
			})
		}
	}

	customMetrics = append(customMetrics, &metrics.Metric{
		Name:        "icmp_traceroute_max_distance",
		Value:       float64(maxDistance),
		Description: "max distance",
		Labels: map[string]string{
			"hostname": args["hostname"],
		},
	})

	return customMetrics, nil
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

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "traceroute",
		Description: "Traceroute a host",
		Params: []plugins.StepParam{
			{
				Name:        "hostname",
				Description: "Hostname to traceroute",
				Optional:    false,
			},
		},
		Fn: p.traceroute,
	})
}

// Init initializes the plugin.
func init() {
	h := &ICMP{}
	h.Init()
	plugins.AddPlugin("icmp", "ICMP plugin is used to ping and traceroute hosts", h)
}

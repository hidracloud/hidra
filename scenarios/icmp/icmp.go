// Monitoring ICMP
package icmp

import (
	"fmt"
	"strconv"

	"github.com/hidracloud/hidra/models"
	"github.com/hidracloud/hidra/scenarios"

	"github.com/aeden/traceroute"
	"github.com/go-ping/ping"
)

// Represent an ICMP scenario
type IcmpScenario struct {
	models.Scenario
}

func (h *IcmpScenario) traceroute(c map[string]string) ([]models.Metric, error) {
	if _, ok := c["hostname"]; !ok {
		return nil, fmt.Errorf("hostname parameter missing")
	}

	options := traceroute.TracerouteOptions{}
	options.SetRetries(0)
	options.SetMaxHops(traceroute.DEFAULT_MAX_HOPS + 1)
	options.SetFirstHop(traceroute.DEFAULT_FIRST_HOP)

	zz := make(chan traceroute.TracerouteHop)
	go func() {
		for {
			_, ok := <-zz
			if !ok {
				return
			}
		}
	}()
	tcrresult, err := traceroute.Traceroute(c["hostname"], &options, zz)

	if err != nil {
		return nil, err
	}

	custom_metrics := make([]models.Metric, 0)

	for i := 0; i < len(tcrresult.Hops); i++ {
		custom_metrics = append(custom_metrics, models.Metric{
			Name:  fmt.Sprintf("hop_%d_elapsed", i),
			Value: float64(tcrresult.Hops[i].ElapsedTime.Milliseconds()),
			Labels: map[string]string{
				"host": tcrresult.Hops[i].Host,
				"ip":   fmt.Sprintf("%v.%v.%v.%v", tcrresult.Hops[i].Address[0], tcrresult.Hops[i].Address[1], tcrresult.Hops[i].Address[2], tcrresult.Hops[i].Address[3]),
			},
		})

		status := 0

		if tcrresult.Hops[i].Success {
			status = 1
		}

		custom_metrics = append(custom_metrics, models.Metric{
			Name:  fmt.Sprintf("hop_%d_status", i),
			Value: float64(status),
			Labels: map[string]string{
				"host": tcrresult.Hops[i].Host,
				"ip":   fmt.Sprintf("%v.%v.%v.%v", tcrresult.Hops[i].Address[0], tcrresult.Hops[i].Address[1], tcrresult.Hops[i].Address[2], tcrresult.Hops[i].Address[3]),
			},
		})
	}

	custom_metrics = append(custom_metrics, models.Metric{
		Name:  "hops",
		Value: float64(len(tcrresult.Hops)),
	})

	return custom_metrics, nil
}

func (h *IcmpScenario) ping(c map[string]string) ([]models.Metric, error) {
	if _, ok := c["hostname"]; !ok {
		return nil, fmt.Errorf("hostname parameter missing")
	}

	pinger, err := ping.NewPinger(c["hostname"])
	if err != nil {
		return nil, err
	}

	pinger.Count = 3
	if _, ok := c["times"]; ok {
		tmp, err := strconv.ParseInt(c["times"], 10, 64)

		if err != nil {
			return nil, err
		}

		pinger.Count = int(tmp)
	}

	err = pinger.Run() // Blocks until finished.
	if err != nil {
		return nil, err
	}

	fmt.Println(err)

	stats := pinger.Statistics()

	custom_metrics := make([]models.Metric, 0)

	custom_metrics = append(custom_metrics, models.Metric{
		Name:  "packet_loss",
		Value: stats.PacketLoss,
	})

	custom_metrics = append(custom_metrics, models.Metric{
		Name:  "min_rtt",
		Value: float64(stats.MinRtt.Milliseconds()),
	})

	custom_metrics = append(custom_metrics, models.Metric{
		Name:  "max_rtt",
		Value: float64(stats.MinRtt.Milliseconds()),
	})

	custom_metrics = append(custom_metrics, models.Metric{
		Name:  "avg_rtt",
		Value: float64(stats.AvgRtt.Milliseconds()),
	})

	custom_metrics = append(custom_metrics, models.Metric{
		Name:  "packet_duplicates",
		Value: float64(stats.PacketsRecvDuplicates),
	})

	custom_metrics = append(custom_metrics, models.Metric{
		Name:  "packet_receive",
		Value: float64(stats.PacketsRecv),
	})

	custom_metrics = append(custom_metrics, models.Metric{
		Name:  "packet_send",
		Value: float64(stats.PacketsSent),
	})

	return custom_metrics, nil
}

func (h *IcmpScenario) Init() {
	h.StartPrimitives()

	h.RegisterStep("ping", h.ping)
	h.RegisterStep("traceroute", h.traceroute)

}

func init() {
	scenarios.Add("icmp", func() models.IScenario {
		return &IcmpScenario{}
	})
}

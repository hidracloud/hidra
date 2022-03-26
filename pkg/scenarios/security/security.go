package security

import (
	"context"
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/JoseCarlosGarcia95/go-port-scanner/portscanner"
	"github.com/go-ping/ping"
	"github.com/hidracloud/hidra/pkg/models"
	"github.com/hidracloud/hidra/pkg/scenarios"
)

// Scenario represent a security scenario
type Scenario struct {
	models.Scenario
}

// RCA represent a RCA
func (s *Scenario) RCA(result *models.ScenarioResult) error {
	log.Println("ICMP RCA")
	return nil
}

// Description returns the scenario description
func (s *Scenario) Description() string {
	return "Run a security scenario"
}

func readPortScannerConfig(c map[string]string) (hostname string, protocol string, startPort int, endPort int, workers int, fingerprint bool) {
	hostname = c["hostname"]
	protocol = "tcp"

	if _, ok := c["protocol"]; ok && len(c["protocol"]) > 0 {
		protocol = c["protocol"]
	}

	startPort = 1
	endPort = 65535

	if _, ok := c["port_start"]; ok && len(c["port_start"]) > 0 {
		startPort, _ = strconv.Atoi(c["port_start"])
	}

	if _, ok := c["port_end"]; ok && len(c["port_end"]) > 0 {
		endPort, _ = strconv.Atoi(c["port_end"])
	}

	workers = 4

	if _, ok := c["workers"]; ok && len(c["workers"]) > 0 {
		workers, _ = strconv.Atoi(c["workers"])
	}

	fingerprint = true

	if _, ok := c["fingerprint"]; ok && len(c["fingerprint"]) > 0 {
		fingerprint, _ = strconv.ParseBool(c["fingerprint"])
	}
	return
}

func (s *Scenario) portScanner(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	hostname, protocol, startPort, endPort, workers, fingerprint := readPortScannerConfig(c)

	pinger, err := ping.NewPinger(hostname)
	pinger.Timeout = time.Second
	if err != nil {
		return nil, err
	}

	pinger.Count = 1

	err = pinger.Run()
	if err != nil {
		return nil, err
	}

	stats := pinger.Statistics()

	if stats.PacketsRecv == 0 {
		return []models.Metric{
			{
				Name:        "host_status",
				Value:       0,
				Description: "Host status",
				Labels: map[string]string{
					"hostname": hostname,
				},
			},
		}, nil
	}
	openedPorts := portscanner.PortRange(hostname, protocol, uint32(startPort), uint32(endPort), uint32(workers))

	metrics := []models.Metric{
		{
			Name:        "host_status",
			Value:       1,
			Description: "Host status",
			Labels: map[string]string{
				"hostname": hostname,
			},
		},
	}

	for _, port := range openedPorts {
		serviceName := portscanner.Port2Service(hostname, protocol, port, fingerprint)

		metric := models.Metric{
			Name:        "opened_port",
			Value:       1,
			Description: "Opened port",
			Labels: map[string]string{
				"hostname":     hostname,
				"port":         strconv.Itoa(int(port)),
				"service_name": serviceName,
			},
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (s *Scenario) subnetPortScanner(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	_, ipv4Net, err := net.ParseCIDR(c["cidr"])
	delete(c, "cidr")

	if err != nil {
		return nil, err
	}

	mask := binary.BigEndian.Uint32(ipv4Net.Mask)
	start := binary.BigEndian.Uint32(ipv4Net.IP)

	finish := (start & mask) | (mask ^ 0xffffffff)

	metrics := make([]models.Metric, 0)
	for i := start; i <= finish; i++ {
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, i)
		c["hostname"] = ip.String()

		oneIPMetrics, err := s.portScanner(ctx, c)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, oneIPMetrics...)
	}
	return metrics, nil
}

// Close closes the scenario
func (s *Scenario) Close() {

}

// Init initializes the scenario
func (s *Scenario) Init() {
	s.StartPrimitives()

	s.RegisterStep("portScanner", models.StepDefinition{
		Description: "Run a port scanner",
		Params: []models.StepParam{
			{Name: "hostname", Description: "Hostname to make a port scanner", Optional: false},
			{Name: "protocol", Description: "Protocol to make a port scanner", Optional: true},
			{Name: "port_start", Description: "Port start to make a port scanner", Optional: true},
			{Name: "port_end", Description: "Port end to make a port scanner", Optional: true},
			{Name: "fingerprint", Description: "Fingerprint to make a port scanner", Optional: true},
			{Name: "workers", Description: "Workers to make a port scanner", Optional: true},
		},
		Fn: s.portScanner,
	})

	s.RegisterStep("subnetPortScanner", models.StepDefinition{
		Description: "Run a subnet port scanner",
		Params: []models.StepParam{
			{Name: "cidr", Description: "Subnet to make a port scanner", Optional: false},
			{Name: "protocol", Description: "Protocol to make a port scanner", Optional: true},
			{Name: "port_start", Description: "Port start to make a port scanner", Optional: true},
			{Name: "port_end", Description: "Port end to make a port scanner", Optional: true},
			{Name: "fingerprint", Description: "Fingerprint to make a port scanner", Optional: true},
			{Name: "workers", Description: "Workers to make a port scanner", Optional: true},
		},
		Fn: s.subnetPortScanner,
	})
}

func init() {
	scenarios.Add("security", func() models.IScenario {
		return &Scenario{}
	})
}

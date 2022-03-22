package security

import (
	"context"
	"log"
	"strconv"
	"sync"

	"github.com/JoseCarlosGarcia95/go-port-scanner/portscanner"
	"github.com/hidracloud/hidra/src/models"
	"github.com/hidracloud/hidra/src/scenarios"
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

func (s *Scenario) portScanner(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	wg := sync.WaitGroup{}

	hostname := c["hostname"]
	protocol := "tcp"

	if _, ok := c["protocol"]; ok && len(c["protocol"]) > 0 {
		protocol = c["protocol"]
	}

	startPort := 1
	endPort := 65535

	if _, ok := c["port_start"]; ok && len(c["port_start"]) > 0 {
		startPort, _ = strconv.Atoi(c["port_start"])
	}

	if _, ok := c["port_end"]; ok && len(c["port_end"]) > 0 {
		endPort, _ = strconv.Atoi(c["port_end"])
	}

	workers := 1000

	if _, ok := c["workers"]; ok && len(c["workers"]) > 0 {
		workers, _ = strconv.Atoi(c["workers"])
	}

	openedPorts := portscanner.PortRange(hostname, protocol, uint32(startPort), uint32(endPort), uint32(workers))

	fingerprint := true

	if _, ok := c["fingerprint"]; ok && len(c["fingerprint"]) > 0 {
		fingerprint, _ = strconv.ParseBool(c["fingerprint"])
	}

	metrics := make([]models.Metric, 0)
	for _, port := range openedPorts {
		serviceName := portscanner.Port2Service(hostname, protocol, port, fingerprint)

		metric := models.Metric{
			Name:        "oepened_port",
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
}

func init() {
	scenarios.Add("security", func() models.IScenario {
		return &Scenario{}
	})
}

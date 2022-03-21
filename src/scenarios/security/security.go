package security

import (
	"context"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

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

	mutex := sync.Mutex{}
	openedPorts := make(map[int]bool)

	for port := 1; port <= 65535; port++ {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			address := hostname + ":" + strconv.Itoa(port)

			conn, err := net.DialTimeout(protocol, address, 5*time.Second)

			if err != nil {
				return
			}

			mutex.Lock()
			openedPorts[port] = true
			mutex.Unlock()

			defer conn.Close()

		}(port)
	}

	wg.Wait()

	metrics := make([]models.Metric, 0)

	for port := range openedPorts {
		metric := models.Metric{
			Name:        "oepened_port",
			Value:       1,
			Description: "Opened port",
			Labels: map[string]string{
				"hostname": hostname,
				"port":     strconv.Itoa(port),
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
		},
		Fn: s.portScanner,
	})
}

func init() {
	scenarios.Add("security", func() models.IScenario {
		return &Scenario{}
	})
}

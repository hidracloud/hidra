package tcp

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"

	b64 "encoding/base64"

	"github.com/hidracloud/hidra/src/models"
	"github.com/hidracloud/hidra/src/scenarios"
)

// Scenario Represent an ssl scenario
type Scenario struct {
	models.Scenario
	conn *net.TCPConn
}

// RCA generate RCAs for scenario
func (h *Scenario) RCA(result *models.ScenarioResult) error {
	log.Println("TCP RCA")
	return nil
}

// Description return the description of the scenario
func (h *Scenario) Description() string {
	return "Run a TCP scenario"
}

func (h *Scenario) connectTo(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", c["to"])
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp4", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	h.conn = conn

	return nil, nil
}

func (h *Scenario) write(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	if h.conn == nil {
		return nil, fmt.Errorf("you should connect to an addr first")
	}

	data, err := b64.StdEncoding.DecodeString(c["data"])

	if err != nil {
		return nil, err
	}

	_, err = h.conn.Write(data)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Scenario) read(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	var err error

	if h.conn == nil {
		return nil, fmt.Errorf("you should connect to an addr first")
	}

	bytesToRead := 1024

	if c["bytesToRead"] != "" {
		bytesToRead, err = strconv.Atoi(c["bytesToRead"])
		if err != nil {
			return nil, err
		}
	}

	rcvData := make([]byte, bytesToRead)
	n, err := h.conn.Read(rcvData)
	if err != nil {
		return nil, err
	}

	rcvDataStr := string(rcvData[:n])

	if len(c["data"]) != 0 {
		dataExpected, err := b64.StdEncoding.DecodeString(c["data"])

		if err != nil {
			return nil, err
		}

		if string(dataExpected) != rcvDataStr {
			return nil, fmt.Errorf("data expected: %s, data received: %s", string(dataExpected), rcvDataStr)
		}
	}
	return nil, nil
}

// Close closes the scenario
func (h *Scenario) Close() {
	if h.conn != nil {
		h.conn.Close()
	}
}

// Init initialize the scenario
func (h *Scenario) Init() {
	h.StartPrimitives()

	h.RegisterStep("connectTo", models.StepDefinition{
		Description: "Connect to a host",
		Params: []models.StepParam{
			{
				Name:        "to",
				Description: "Host to connect to",
				Optional:    false,
			},
		},
		Fn: h.connectTo,
	})

	h.RegisterStep("write", models.StepDefinition{
		Description: "Write data to the connection",
		Params: []models.StepParam{
			{
				Name:        "data",
				Description: "Data to write in base64",
				Optional:    false,
			},
		},
		Fn: h.write,
	})

	h.RegisterStep("read", models.StepDefinition{
		Description: "Read data from the connection",
		Params: []models.StepParam{
			{
				Name:        "data",
				Description: "Data to read in base64",
				Optional:    false,
			},
			{
				Name:        "bytesToRead",
				Description: "Number of bytes to read",
				Optional:    true,
			},
		},
		Fn: h.read,
	})
}

func init() {
	scenarios.Add("tcp", func() models.IScenario {
		return &Scenario{}
	})
}

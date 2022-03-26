package tcp

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"

	b64 "encoding/base64"

	"github.com/hidracloud/hidra/pkg/models"
	"github.com/hidracloud/hidra/pkg/scenarios"
)

// Scenario Represent an ssl scenario
type Scenario struct {
	models.Scenario
	conn *net.TCPConn
}

// RCA generate RCAs for scenario
func (t *Scenario) RCA(result *models.ScenarioResult) error {
	log.Println("TCP RCA")
	return nil
}

// Description return the description of the scenario
func (t *Scenario) Description() string {
	return "Run a TCP scenario"
}

func (t *Scenario) connectTo(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", c["to"])
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp4", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	t.conn = conn

	return nil, nil
}

func (t *Scenario) write(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	if t.conn == nil {
		return nil, fmt.Errorf("you should connect to an addr first")
	}

	data, err := b64.StdEncoding.DecodeString(c["data"])

	if err != nil {
		return nil, err
	}

	_, err = t.conn.Write(data)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (t *Scenario) read(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	var err error

	if t.conn == nil {
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
	n, err := t.conn.Read(rcvData)
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
func (t *Scenario) Close() {
	if t.conn != nil {
		t.conn.Close()
	}
}

// Init initialize the scenario
func (t *Scenario) Init() {
	t.StartPrimitives()

	t.RegisterStep("connectTo", models.StepDefinition{
		Description: "Connect to a host",
		Params: []models.StepParam{
			{
				Name:        "to",
				Description: "Host to connect to",
				Optional:    false,
			},
		},
		Fn: t.connectTo,
	})

	t.RegisterStep("write", models.StepDefinition{
		Description: "Write data to the connection",
		Params: []models.StepParam{
			{
				Name:        "data",
				Description: "Data to write in base64",
				Optional:    false,
			},
		},
		Fn: t.write,
	})

	t.RegisterStep("read", models.StepDefinition{
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
		Fn: t.read,
	})
}

func init() {
	scenarios.Add("tcp", func() models.IScenario {
		return &Scenario{}
	})
}

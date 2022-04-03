package udp

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"

	b64 "encoding/base64"

	"github.com/hidracloud/hidra/v2/pkg/models"
	"github.com/hidracloud/hidra/v2/pkg/scenarios"
)

// Scenario Represent an ssl scenario
type Scenario struct {
	models.Scenario

	conn *net.UDPConn
}

// RCA generate RCAs for scenario
func (u *Scenario) RCA(result *models.ScenarioResult) error {
	log.Println("UDP RCA")
	return nil
}

// Description return the description of the scenario
func (u *Scenario) Description() string {
	return "Run a UDP scenario"
}

// Close closes the scenario
func (u *Scenario) Close() {
	if u.conn != nil {
		u.conn.Close()
	}
}

func (u *Scenario) connectTo(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", c["to"])
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, err
	}

	u.conn = conn

	return nil, nil
}

func (u *Scenario) write(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	if u.conn == nil {
		return nil, fmt.Errorf("you should connect to an addr first")
	}

	data, err := b64.StdEncoding.DecodeString(c["data"])

	if err != nil {
		return nil, err
	}

	_, err = u.conn.Write(data)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (u *Scenario) read(ctx context.Context, c map[string]string) ([]models.Metric, error) {
	var err error

	if u.conn == nil {
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
	n, err := u.conn.Read(rcvData)
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

// Init initialize the scenario
func (u *Scenario) Init() {
	u.StartPrimitives()

	u.RegisterStep("connectTo", models.StepDefinition{
		Description: "Connect to a host",
		Params: []models.StepParam{
			{
				Name:        "to",
				Description: "Host to connect to",
				Optional:    false,
			},
		},
		Fn: u.connectTo,
	})

	u.RegisterStep("write", models.StepDefinition{
		Description: "Write data to the connection",
		Params: []models.StepParam{
			{
				Name:        "data",
				Description: "Data to write in base64",
				Optional:    false,
			},
		},
		Fn: u.write,
	})

	u.RegisterStep("read", models.StepDefinition{
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
		Fn: u.read,
	})
}

func init() {
	scenarios.Add("udp", func() models.IScenario {
		return &Scenario{}
	})
}

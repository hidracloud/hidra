package tcp

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/internal/misc"
	"github.com/hidracloud/hidra/v3/internal/plugins"

	b64 "encoding/base64"
)

// TCP represents a TCP plugin.
type TCP struct {
	plugins.BasePlugin
}

// whoisFrom returns the whois information from a domain.
func (p *TCP) connectTo(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", args["to"])
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp4", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	stepsgen[misc.ContextTCPConnection] = conn

	return nil, nil
}

// write writes a file to the TCP server.
func (p *TCP) write(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	if _, ok := stepsgen[misc.ContextTCPConnection].(*net.TCPConn); !ok {
		return nil, fmt.Errorf("no tcp connection found")
	}

	conn := stepsgen[misc.ContextTCPConnection].(*net.TCPConn)

	data, err := b64.StdEncoding.DecodeString(args["data"])

	if err != nil {
		return nil, err
	}

	startTime := time.Now()

	byteLen, err := conn.Write(data)
	if err != nil {
		return nil, err
	}

	customMetrics := []*metrics.Metric{
		{
			Name:        "tcp_write_time",
			Description: "The time it took to write the data to the TCP server",
			Value:       time.Since(startTime).Seconds(),
		},
		{
			Name:        "tcp_write_size",
			Description: "The size of the data written to the TCP server",
			Value:       float64(byteLen),
		},
	}

	return customMetrics, nil
}

// read reads a file from the TCP server.
func (p *TCP) read(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	var err error

	if _, ok := stepsgen[misc.ContextTCPConnection].(*net.TCPConn); !ok {
		return nil, fmt.Errorf("no TCP connection found")
	}

	conn := stepsgen[misc.ContextTCPConnection].(*net.TCPConn)

	bytesToRead := 1024

	if args["bytesToRead"] != "" {
		bytesToRead, err = strconv.Atoi(args["bytesToRead"])
		if err != nil {
			return nil, err
		}
	}

	rcvData := make([]byte, bytesToRead)

	startTime := time.Now()

	n, err := conn.Read(rcvData)
	if err != nil {
		return nil, err
	}

	stepsgen[misc.ContextOutput] = rcvData[:n]

	customMetrics := []*metrics.Metric{
		{
			Name:        "tcp_read_time",
			Description: "The time it took to write the data to the TCP server",
			Value:       time.Since(startTime).Seconds(),
		},
		{
			Name:        "tcp_read_size",
			Description: "The size of the data written to the TCP server",
			Value:       float64(n),
		},
	}

	return customMetrics, nil
}

// onClose closes the connection.
func (p *TCP) onClose(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {

	if _, ok := stepsgen[misc.ContextTCPConnection].(*net.TCPConn); !ok {
		return nil, fmt.Errorf("no FTP connection found")
	}

	conn := stepsgen[misc.ContextTCPConnection].(*net.TCPConn)

	err := conn.Close()

	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Init initializes the plugin.
func (p *TCP) Init() {
	p.Primitives()

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "connectTo",
		Description: "Connect to a TCP server",
		Params: []plugins.StepParam{
			{
				Name:        "to",
				Description: "Host to connect to",
				Optional:    false,
			},
		},
		Fn: p.connectTo,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "write",
		Description: "Write a file to a TCP server",
		Params: []plugins.StepParam{
			{
				Name:        "data",
				Description: "Data to write",
				Optional:    false,
			},
		},
		Fn: p.write,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "read",
		Description: "Read a file from a FTP server",
		Params: []plugins.StepParam{
			{
				Name:        "bytesToRead",
				Description: "Number of bytes to read",
				Optional:    true,
			},
		},
		Fn: p.read,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "onClose",
		Description: "Close the connection",
		Params:      []plugins.StepParam{},
		Fn:          p.onClose,
	})
}

// Init initializes the plugin.
func init() {
	h := &TCP{}
	h.Init()
	plugins.AddPlugin("tcp", h)
}

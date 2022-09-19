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
func (p *TCP) connectTo(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", args["to"])
	if err != nil {
		return ctx, nil, err
	}

	conn, err := net.DialTCP("tcp4", nil, tcpAddr)
	if err != nil {
		return ctx, nil, err
	}

	ctx = context.WithValue(ctx, misc.ContextTCPConnection, conn)

	return ctx, nil, nil
}

// write writes a file to the TCP server.
func (p *TCP) write(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	if _, ok := ctx.Value(misc.ContextTCPConnection).(*net.TCPConn); !ok {
		return ctx, nil, fmt.Errorf("no tcp connection found")
	}

	conn := ctx.Value(misc.ContextTCPConnection).(*net.TCPConn)

	data, err := b64.StdEncoding.DecodeString(args["data"])

	if err != nil {
		return ctx, nil, err
	}

	startTime := time.Now()

	byteLen, err := conn.Write(data)
	if err != nil {
		return ctx, nil, err
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

	return ctx, customMetrics, nil
}

// read reads a file from the TCP server.
func (p *TCP) read(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	var err error

	if _, ok := ctx.Value(misc.ContextTCPConnection).(*net.TCPConn); !ok {
		return ctx, nil, fmt.Errorf("no TCP connection found")
	}

	conn := ctx.Value(misc.ContextTCPConnection).(*net.TCPConn)

	bytesToRead := 1024

	if args["bytesToRead"] != "" {
		bytesToRead, err = strconv.Atoi(args["bytesToRead"])
		if err != nil {
			return ctx, nil, err
		}
	}

	rcvData := make([]byte, bytesToRead)

	startTime := time.Now()

	n, err := conn.Read(rcvData)
	if err != nil {
		return ctx, nil, err
	}

	rcvDataStr := string(rcvData[:n])

	ctx = context.WithValue(ctx, misc.ContextOutput, rcvDataStr)

	customMetrics := []*metrics.Metric{
		{
			Name:        "tcp_read_time",
			Description: "The time it took to write the data to the TCP server",
			Value:       time.Since(startTime).Seconds(),
		},
		{
			Name:        "tcp_read_size",
			Description: "The size of the data written to the TCP server",
			Value:       float64(len(rcvDataStr)),
		},
	}

	return ctx, customMetrics, nil
}

// onClose closes the connection.
func (p *TCP) onClose(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {

	if _, ok := ctx.Value(misc.ContextTCPConnection).(*net.TCPConn); !ok {
		return ctx, nil, fmt.Errorf("no FTP connection found")
	}

	conn := ctx.Value(misc.ContextTCPConnection).(*net.TCPConn)

	err := conn.Close()

	if err != nil {
		return ctx, nil, err
	}

	return ctx, nil, nil
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
		ContextGenerator: []plugins.ContextGenerator{
			{
				Name:        misc.ContextOutput.Name,
				Description: "The TCP read contents",
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

package udp

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

// UDP represents a UDP plugin.
type UDP struct {
	plugins.BasePlugin
}

// whoisFrom returns the whois information from a domain.
func (p *UDP) connectTo(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", args["to"])
	if err != nil {
		return ctx, nil, err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return ctx, nil, err
	}

	ctx = context.WithValue(ctx, misc.ContextUDPConnection, conn)

	return ctx, nil, nil
}

// write writes a file to the UDP server.
func (p *UDP) write(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	if _, ok := ctx.Value(misc.ContextUDPConnection).(*net.UDPConn); !ok {
		return ctx, nil, fmt.Errorf("no udp connection found")
	}

	conn := ctx.Value(misc.ContextUDPConnection).(*net.UDPConn)

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
			Name:        "udp_write_time",
			Description: "The time it took to write the data to the UDP server",
			Value:       time.Since(startTime).Seconds(),
		},
		{
			Name:        "udp_write_size",
			Description: "The size of the data written to the UDP server",
			Value:       float64(byteLen),
		},
	}

	return ctx, customMetrics, nil
}

// read reads a file from the UDP server.
func (p *UDP) read(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	var err error

	if _, ok := ctx.Value(misc.ContextUDPConnection).(*net.UDPConn); !ok {
		return ctx, nil, fmt.Errorf("no UDP connection found")
	}

	conn := ctx.Value(misc.ContextUDPConnection).(*net.UDPConn)

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
			Name:        "udp_read_time",
			Description: "The time it took to write the data to the UDP server",
			Value:       time.Since(startTime).Seconds(),
		},
		{
			Name:        "udp_read_size",
			Description: "The size of the data written to the UDP server",
			Value:       float64(len(rcvDataStr)),
		},
	}

	return ctx, customMetrics, nil
}

// onClose closes the connection.
func (p *UDP) onClose(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {

	if _, ok := ctx.Value(misc.ContextUDPConnection).(*net.UDPConn); !ok {
		return ctx, nil, fmt.Errorf("no FTP connection found")
	}

	conn := ctx.Value(misc.ContextUDPConnection).(*net.UDPConn)

	err := conn.Close()

	if err != nil {
		return ctx, nil, err
	}

	return ctx, nil, nil
}

// Init initializes the plugin.
func (p *UDP) Init() {
	p.Primitives()

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "connectTo",
		Description: "Connect to a UDP server",
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
		Description: "Write a file to a UDP server",
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
				Description: "The UDP read contents",
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
	h := &UDP{}
	h.Init()
	plugins.AddPlugin("udp", h)
}

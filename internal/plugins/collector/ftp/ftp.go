package ftp

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/internal/misc"
	"github.com/hidracloud/hidra/v3/internal/plugins"
	ftpclient "github.com/jlaffaye/ftp"
)

var (
	errNoFTPConnection = fmt.Errorf("no FTP connection found")
)

// FTP represents a FTP plugin.
type FTP struct {
	plugins.BasePlugin
}

// whoisFrom returns the whois information from a domain.
func (p *FTP) connectTo(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	if _, ok := ctx.Value(misc.ContextFTPConnection).(*ftpclient.ServerConn); ok {
		err := ctx.Value(misc.ContextFTPConnection).(*ftpclient.ServerConn).Quit()

		if err != nil {
			return ctx, nil, err
		}
	}

	timeout := 30 * time.Second

	if _, ok := ctx.Value(misc.ContextTimeout).(time.Duration); ok {
		timeout = ctx.Value(misc.ContextTimeout).(time.Duration)
	}

	ftpConn, err := ftpclient.Dial(args["to"], ftpclient.DialWithTimeout(timeout))

	if err != nil {
		return ctx, nil, err
	}

	ctx = context.WithValue(ctx, misc.ContextFTPHost, args["to"])
	ctx = context.WithValue(ctx, misc.ContextFTPConnection, ftpConn)

	return ctx, nil, nil
}

// login logs in to the FTP server.
func (p *FTP) login(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	if _, ok := ctx.Value(misc.ContextFTPConnection).(*ftpclient.ServerConn); !ok {
		return ctx, nil, errNoFTPConnection
	}

	ftpConn := ctx.Value(misc.ContextFTPConnection).(*ftpclient.ServerConn)

	err := ftpConn.Login(args["user"], args["password"])

	if err != nil {
		return ctx, nil, err
	}

	return ctx, nil, nil
}

// write writes a file to the FTP server.
func (p *FTP) write(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	if _, ok := ctx.Value(misc.ContextFTPConnection).(*ftpclient.ServerConn); !ok {
		return ctx, nil, errNoFTPConnection
	}

	ftpConn := ctx.Value(misc.ContextFTPConnection).(*ftpclient.ServerConn)

	data := bytes.NewBufferString(args["data"])

	startTime := time.Now()

	err := ftpConn.Stor(args["file"], data)

	if err != nil {
		return ctx, nil, err
	}

	customMetrics := []*metrics.Metric{
		{
			Name:        "ftp_write_size",
			Description: "The size of the file written to the FTP server",
			Value:       float64(len(args["data"])),
			Labels: map[string]string{
				"host": ctx.Value(misc.ContextFTPHost).(string),
			},
		},
		{
			Name:        "ftp_write_time",
			Description: "The time it took to write the file to the FTP server",
			Value:       float64(time.Since(startTime).Milliseconds()),
			Labels: map[string]string{
				"host": ctx.Value(misc.ContextFTPHost).(string),
			},
		},
	}

	return ctx, customMetrics, nil
}

// read reads a file from the FTP server.
func (p *FTP) read(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	if _, ok := ctx.Value(misc.ContextFTPConnection).(*ftpclient.ServerConn); !ok {
		return ctx, nil, errNoFTPConnection
	}

	ftpConn := ctx.Value(misc.ContextFTPConnection).(*ftpclient.ServerConn)

	startTime := time.Now()

	data, err := ftpConn.Retr(args["file"])

	if err != nil {
		return ctx, nil, err
	}

	defer data.Close()
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(data)

	if err != nil {
		return ctx, nil, err
	}

	ctx = context.WithValue(ctx, misc.ContextOutput, buf.String())

	customMetrics := []*metrics.Metric{
		{
			Name:        "ftp_read_size",
			Description: "The size of the file read from the FTP server",
			Value:       float64(buf.Len()),
			Labels: map[string]string{
				"host": ctx.Value(misc.ContextFTPHost).(string),
			},
		},
		{
			Name:        "ftp_read_time",
			Description: "The time it took to read the file from the FTP server",
			Value:       float64(time.Since(startTime).Milliseconds()),
			Labels: map[string]string{
				"host": ctx.Value(misc.ContextFTPHost).(string),
			},
		},
	}

	return ctx, customMetrics, nil
}

// delete deletes a file from the FTP server.
func (p *FTP) delete(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	if _, ok := ctx.Value(misc.ContextFTPConnection).(*ftpclient.ServerConn); !ok {
		return ctx, nil, errNoFTPConnection
	}

	ftpConn := ctx.Value(misc.ContextFTPConnection).(*ftpclient.ServerConn)

	startTime := time.Now()

	err := ftpConn.Delete(args["file"])

	if err != nil {
		return ctx, nil, err
	}

	customMetrics := []*metrics.Metric{
		{
			Name:        "ftp_delete_time",
			Description: "The time it took to delete the file from the FTP server",
			Value:       float64(time.Since(startTime).Milliseconds()),
			Labels: map[string]string{
				"host": ctx.Value(misc.ContextFTPHost).(string),
			},
		},
	}

	return ctx, customMetrics, nil
}

// onClose closes the connection.
func (p *FTP) onClose(ctx context.Context, args map[string]string) (context.Context, []*metrics.Metric, error) {
	if _, ok := ctx.Value(misc.ContextFTPConnection).(*ftpclient.ServerConn); ok {
		err := ctx.Value(misc.ContextFTPConnection).(*ftpclient.ServerConn).Quit()

		if err != nil {
			return ctx, nil, err
		}
	}

	return ctx, nil, nil
}

// Init initializes the plugin.
func (p *FTP) Init() {
	p.Primitives()

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "connectTo",
		Description: "Connect to a FTP server",
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
		Name:        "login",
		Description: "Login to a FTP server",
		Params: []plugins.StepParam{
			{
				Name:        "user",
				Description: "User to login with",
				Optional:    false,
			},
			{
				Name:        "password",
				Description: "Password to login with",
				Optional:    false,
			},
		},
		Fn: p.login,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "write",
		Description: "Write a file to a FTP server",
		Params: []plugins.StepParam{
			{
				Name:        "file",
				Description: "File to write",
				Optional:    false,
			},
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
				Name:        "file",
				Description: "File to read",
				Optional:    false,
			},
		},
		ContextGenerator: []plugins.ContextGenerator{
			{
				Name:        misc.ContextOutput.Name,
				Description: "The FTP file contents",
			},
		},
		Fn: p.read,
	})

	p.RegisterStep(&plugins.StepDefinition{
		Name:        "delete",
		Description: "Delete a file from a FTP server",
		Params: []plugins.StepParam{
			{
				Name:        "file",
				Description: "File to delete",
				Optional:    false,
			},
		},
		Fn: p.delete,
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
	h := &FTP{}
	h.Init()
	plugins.AddPlugin("ftp", h)
}

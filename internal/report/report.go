package report

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hidracloud/hidra/v3/internal/config"
	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/internal/misc"
	"github.com/hidracloud/hidra/v3/internal/utils"
	"github.com/pixelbender/go-traceroute/traceroute"
)

// IsEnabled returns true if the report is enabled.
var IsEnabled = true

// Report is a report of a single test run.
type Report struct {
	// Name is the name of the test.
	Name string `json:"name,omitempty"`
	// Path is the path of the test.
	Path string `json:"path,omitempty"`
	// Duration is the duration of the test.
	Duration string `json:"duration,omitempty"`
	// Metrics is the metrics of the test.
	Metrics map[string]float64 `json:"metrics,omitempty"`
	// LastError is the last error of the test.
	LastError string `json:"last_error,omitempty"`
	// Attachments is the attachments of the report.
	Attachments map[string]string `json:"attachments,omitempty"`
	// Tags is the tags of the report.
	Tags map[string]string `json:"tags,omitempty"`
	// ConnectionInfo is the connection info of the report.
	ConnectionInfo ReportConnectionInfo `json:"connection_info,omitempty"`
	// HttpInfo is the HTTP info of the report.
	HttpInfo ReportHttpRespone `json:"http_info,omitempty"`
	// Output is the output of the report.
	Output string `json:"output,omitempty"`
	// Variables is the variables of the report.
	Variables map[string]string `json:"variables,omitempty"`
}

// ReportConnectionInfo is the connection info of the report.
type ReportConnectionInfo struct {
	IP         string   `json:"ip,omitempty"`
	Traceroute []string `json:"traceroute,omitempty"`
}

// ReportHttpRespone is the HTTP response of the report.
type ReportHttpRespone struct {
	// Headers is the headers of the report.
	Headers map[string]string `json:"headers,omitempty"`
	// ResponseCode
	ResponseCode int `json:"response_code,omitempty"`
}

// NewReport creates a new report.
func NewReport(sample *config.SampleConfig, allMetrics []*metrics.Metric, variables map[string]string, duration time.Duration, ctx context.Context, err error) *Report {
	if !IsEnabled {
		return nil
	}

	report := &Report{
		Name:      sample.Name,
		Path:      sample.Path,
		Duration:  duration.String(),
		Metrics:   metrics.MetricsToMap(allMetrics),
		LastError: err.Error(),
		Tags:      sample.Tags,
		Variables: variables,
	}

	report.GenerateConnectionInfo(ctx)
	report.GenerateReportHttpRespone(ctx)
	report.GenerateOutput(ctx)

	return report
}

// GenerateConnectionInfo returns the connection info of the report.
func (r *Report) GenerateConnectionInfo(ctx context.Context) {
	lastIP := ""

	tracerouteList := []string{}
	if lastIP, ok := ctx.Value(misc.ContextConnectionIP).(string); ok {
		// nolint: errcheck
		hops, _ := traceroute.Trace(net.ParseIP(lastIP))

		for _, hop := range hops {
			tracerouteList = append(tracerouteList, fmt.Sprintf("%d. %v %v", hop.Distance, hop.Nodes[0].IP, hop.Nodes[0].RTT))
		}
	}

	r.ConnectionInfo = ReportConnectionInfo{
		IP:         lastIP,
		Traceroute: tracerouteList,
	}
}

// GenerateReportHttpRespone returns the HTTP response of the report.
func (r *Report) GenerateReportHttpRespone(ctx context.Context) {
	if httpResp, ok := ctx.Value(misc.ContextHTTPResponse).(*http.Response); ok {
		headers := map[string]string{}

		for k, v := range httpResp.Header {
			headers[k] = strings.Join(v, ",")
		}

		r.HttpInfo = ReportHttpRespone{
			Headers:      headers,
			ResponseCode: httpResp.StatusCode,
		}
	}
}

// GenerateOutput set output into report
func (r *Report) GenerateOutput(ctx context.Context) {
	if output, ok := ctx.Value(misc.ContextOutput).(string); ok {
		r.Output = utils.HTMLStripTags(output)

		// convert r.Output to base64
		r.Output = utils.Base64Encode(r.Output)
	}
}

// Dump returns the string representation of the report.
func (r *Report) Dump() string {
	e, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		return ""
	}
	return string(e)
}

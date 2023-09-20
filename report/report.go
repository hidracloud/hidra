package report

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hidracloud/hidra/v3/config"
	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/internal/misc"

	log "github.com/sirupsen/logrus"
)

var (
	// IsEnabled returns true if the report is enabled.
	IsEnabled = false
	// ReportS3Conf is the S3 report configuration.
	ReportS3Conf *ReportS3Config
	// CallbackConf is the callback report configuration.
	CallbackConf *CallbackConfig
	// BasePath is the base path of the report.
	BasePath = "/tmp/hidra"
)

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
	Attachments map[string][]byte `json:"-"`
	// AttachmentList is the list of attachments.
	AttachmentList []string `json:"attachments,omitempty"`
	// Tags is the tags of the report.
	Tags map[string]string `json:"tags,omitempty"`
	// ConnectionInfo is the connection info of the report.
	ConnectionInfo ReportConnectionInfo `json:"connection_info,omitempty"`
	// HttpInfo is the HTTP info of the report.
	HttpInfo ReportHttpRespone `json:"http_info,omitempty"`
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

// ReportS3Config is the S3 report configuration.
type ReportS3Config struct {
	// Bucket is the bucket name.
	Bucket string `yaml:"bucket"`
	// Region is the region.
	Region string `yaml:"region"`
	// AccessKeyID is the access key ID.
	AccessKeyID string `yaml:"access_key_id"`
	// SecretAccessKey is the secret access key.
	SecretAccessKey string `yaml:"secret_access_key"`
	// Endpoint is the endpoint.
	Endpoint string `yaml:"endpoint"`
	// ForcePathStyle is the flag to force path style.
	ForcePathStyle bool `yaml:"force_path_style"`
	// UseSSL is the flag to use SSL.
	UseSSL bool `yaml:"use_ssl"`
}

// CallbackConfig is the callback report configuration.
type CallbackConfig struct {
	// URL is the URL of the callback.
	URL string `yaml:"url"`
}

// SetS3Configuration configures the S3 report.
func SetS3Configuration(reportS3Conf *ReportS3Config) {
	ReportS3Conf = reportS3Conf
}

// SetCallbackConfiguration configures the callback report.
func SetCallbackConfiguration(callbackConf *CallbackConfig) {
	CallbackConf = callbackConf
}

// SetBasePath set base path of report
func SetBasePath(basePath string) {
	BasePath = basePath
}

// NewReport creates a new report.
func NewReport(sample *config.SampleConfig, allMetrics []*metrics.Metric, variables map[string]string, duration time.Duration, stepsgen map[string]any, err error) *Report {
	if !IsEnabled {
		return nil
	}

	log.Debug("Generating new report")

	report := &Report{
		Name:      sample.Name,
		Path:      sample.Path,
		Duration:  duration.String(),
		Metrics:   metrics.MetricsToMap(allMetrics),
		LastError: err.Error(),
		Tags:      sample.Tags,
		Variables: variables,
	}

	report.GenerateConnectionInfo(stepsgen)
	report.GenerateReportHttpRespone(stepsgen)
	report.GenerateAttachments(stepsgen)

	return report
}

// GenerateConnectionInfo returns the connection info of the report.
func (r *Report) GenerateConnectionInfo(stepsgen map[string]any) {
	lastIP := ""

	if val, ok := stepsgen[misc.ContextConnectionIP].(string); ok {
		lastIP = val
	}

	r.ConnectionInfo = ReportConnectionInfo{
		IP: lastIP,
	}
}

// GenerateAttachments generates all attachnments of the report.
func (r *Report) GenerateAttachments(stepsgen map[string]any) {
	if attachments, ok := stepsgen[misc.ContextAttachment].(map[string][]byte); ok {
		r.Attachments = attachments
		r.AttachmentList = []string{}
		for k := range attachments {
			r.AttachmentList = append(r.AttachmentList, k)
		}
	}
}

// GenerateReportHttpRespone returns the HTTP response of the report.
func (r *Report) GenerateReportHttpRespone(stepsgen map[string]any) {
	if httpResp, ok := stepsgen[misc.ContextHTTPResponse].(*http.Response); ok {
		headers := map[string]string{}

		log.Debug("Generating HTTP response info")
		for k, v := range httpResp.Header {
			headers[k] = strings.Join(v, ",")
		}

		r.HttpInfo = ReportHttpRespone{
			Headers:      headers,
			ResponseCode: httpResp.StatusCode,
		}
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

// GenerateMoreIndexHTML generates the index.html for the report.
func (r *Report) GenerateMoreIndexHTML() string {
	ulHTML := ""
	for _, dest := range r.AttachmentList {
		ulHTML += fmt.Sprintf("<li><a href=\"%s\">%s</a></li>", dest, dest)
	}

	indexHTML := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<title>Attachments</title>
</head>
<body>
	<h1>Attachments</h1>
	<ul>
		%s
	</ul>
</body>
</html>`, ulHTML)

	return indexHTML
}

// Save saves the report to a file.
func (r *Report) Save() error {
	if !IsEnabled {
		return nil
	}

	if r == nil {
		return nil
	}

	if ReportS3Conf != nil {
		err := r.SaveS3()
		if err != nil {
			return err
		}
	}

	if CallbackConf != nil {
		err := r.SendCallback()
		if err != nil {
			return err
		}
	}

	return r.SaveFile()
}

// Save saves an slice of reports to a file.
func Save(reports []*Report) {
	if !IsEnabled {
		return
	}

	for _, oneReport := range reports {
		rErr := oneReport.Save()

		if rErr != nil {
			log.Errorf("Error saving report: %s", rErr)
		}
	}
}

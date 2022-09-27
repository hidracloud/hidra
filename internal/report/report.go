package report

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hidracloud/hidra/v3/internal/config"
	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/internal/misc"
	"github.com/hidracloud/hidra/v3/internal/utils"
	"github.com/pixelbender/go-traceroute/traceroute"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	log "github.com/sirupsen/logrus"
)

var (
	// IsEnabled returns true if the report is enabled.
	IsEnabled = false
	// ReportS3Conf
	ReportS3Conf *ReportS3Config
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

// SetS3Configuration configures the S3 report.
func SetS3Configuration(reportS3Conf *ReportS3Config) {
	ReportS3Conf = reportS3Conf
}

// SetBasePath set base path of report
func SetBasePath(basePath string) {
	BasePath = basePath
}

// NewReport creates a new report.
func NewReport(sample *config.SampleConfig, allMetrics []*metrics.Metric, variables map[string]string, duration time.Duration, ctx context.Context, err error) *Report {
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

	report.GenerateConnectionInfo(ctx)
	report.GenerateReportHttpRespone(ctx)
	report.GenerateOutput(ctx)
	report.GenerateAttachments(ctx)

	return report
}

// GenerateConnectionInfo returns the connection info of the report.
func (r *Report) GenerateConnectionInfo(ctx context.Context) {
	lastIP := ""

	log.Debug("Generating connection info with traceroute")
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

// GenerateAttachments generates all attachnments of the report.
func (r *Report) GenerateAttachments(ctx context.Context) {
	if attachments, ok := ctx.Value(misc.ContextAttachment).(map[string][]byte); ok {
		r.Attachments = attachments
		r.AttachmentList = []string{}
		for k := range attachments {
			r.AttachmentList = append(r.AttachmentList, k)
		}
	}
}

// GenerateReportHttpRespone returns the HTTP response of the report.
func (r *Report) GenerateReportHttpRespone(ctx context.Context) {
	if httpResp, ok := ctx.Value(misc.ContextHTTPResponse).(*http.Response); ok {
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

// SaveS3 saves the report to S3.
func (r *Report) SaveS3() error {
	if ReportS3Conf == nil {
		return nil
	}

	log.Debug("Saving report to S3")

	minioClient, err := minio.New(ReportS3Conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(ReportS3Conf.AccessKeyID, ReportS3Conf.SecretAccessKey, ""),
		Secure: ReportS3Conf.UseSSL,
		Region: ReportS3Conf.Region,
	})

	if err != nil {
		return err
	}

	rDump := r.Dump()

	// rDump to reader
	reader := strings.NewReader(rDump)

	if reader == nil {
		return errors.New("reader is nil")
	}

	_, err = minioClient.PutObject(context.Background(), ReportS3Conf.Bucket, r.Name+".json", reader, int64(len(rDump)), minio.PutObjectOptions{
		ContentType: "application/json",
	})

	// upload attachments to r.Name.more/ folder
	for dest, content := range r.Attachments {
		// create a reader from origin file
		reader := bytes.NewReader(content)

		contentType := "application/octet-stream"

		// get content type from file extension
		if ext := filepath.Ext(dest); ext != "" {
			if ct := mime.TypeByExtension(ext); ct != "" {
				contentType = ct
				// remove encoding from content type
				if i := strings.Index(contentType, ";"); i != -1 {
					contentType = contentType[:i]
				}
			}
		}

		_, err = minioClient.PutObject(context.Background(), ReportS3Conf.Bucket, r.Name+".more/"+dest, reader, int64(len(content)), minio.PutObjectOptions{
			ContentType: contentType,
		})

		if err != nil {
			log.Warnf("Failed to upload attachment to S3: %s", err)
			continue
		}
	}

	if len(r.Attachments) > 0 {
		indexHTML := r.GenerateMoreIndexHTML()

		_, err = minioClient.PutObject(context.Background(), ReportS3Conf.Bucket, r.Name+".more/index.html", strings.NewReader(indexHTML), int64(len(indexHTML)), minio.PutObjectOptions{
			ContentType: "text/html",
		})

		if err != nil {
			log.Warnf("Failed to upload index.html to S3: %s", err)
		}
	}
	return err
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

// SaveFile saves the report to a file.
func (r *Report) SaveFile() error {
	if r == nil {
		return nil
	}

	rDump := r.Dump()

	if err := os.MkdirAll(BasePath, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(BasePath, r.Name+".json")

	log.Debugf("Saving report to file %s", filePath)

	for dest, content := range r.Attachments {
		attachmentPath := filepath.Join(BasePath, r.Name+".more", dest)

		if err := os.MkdirAll(filepath.Dir(attachmentPath), 0755); err != nil {
			log.Errorf("Error creating attachment directory %s: %s", filepath.Dir(attachmentPath), err)
			continue
		}

		if err := os.WriteFile(attachmentPath, content, 0644); err != nil {
			log.Errorf("Error writing attachment %s: %s", attachmentPath, err)
			continue
		}
	}

	return os.WriteFile(filePath, []byte(rDump), 0644)
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

	return r.SaveFile()
}

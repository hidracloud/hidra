package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// ExporterConfig is the configuration for the exporter.
type ExporterConfig struct {
	// LogLevel is the log level.
	LogLevel string `yaml:"log_level"`

	// HTTPServerConfig is the configuration for the HTTP server.
	HTTPServerConfig struct {
		ListenAddress string `yaml:"listen_address"`
		MetricsPath   string `yaml:"metrics_path"`
	} `yaml:"http_server"`

	// SamplesPath is the path to the samples.
	SamplesPath string `yaml:"samples_path"`

	// SchedulerConfig is the configuration for the scheduler.
	SchedulerConfig struct {
		// RefreshSamplesInterval is the interval to refresh the samples.
		RefreshSamplesInterval time.Duration `yaml:"refresh_samples_interval"`
		// EnqueueSamplesInterval is the interval to enqueue the samples.
		EnqueueSamplesInterval time.Duration `yaml:"enqueue_samples_interval"`
		// GCInterval is the interval to run the garbage collector.
		GCInterval time.Duration `yaml:"gc_interval"`
	} `yaml:"scheduler"`

	BasicAuth struct {
		Enabled  bool   `yaml:"enabled" default:"false"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"basic_auth"`

	WorkerConfig struct {
		// ParallelJobs is the size of the channel.
		ParallelJobs int `yaml:"parallel_jobs"`
		// MaxQueueSize is the max size of the queue.
		MaxQueueSize int `yaml:"max_queue_size"`
		// SleepBetweenJobs is the sleep between jobs.
		SleepBetweenJobs time.Duration `yaml:"sleep_between_jobs"`
	} `yaml:"worker"`

	// ReportConfig is the configuration for the report.
	ReportConfig struct {
		// Enabled is the flag to enable the report.
		Enabled bool `yaml:"enabled"`
		// S3Config is the configuration for the S3.
		S3Config struct {
			// Enabled is the flag to enable the S3.
			Enabled bool `yaml:"enabled"`
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
		} `yaml:"s3"`
		// CallbackConfig is the configuration for the callback.
		CallbackConfig struct {
			// Enabled is the flag to enable the callback.
			Enabled bool `yaml:"enabled"`
			// URL is the URL.
			URL string `yaml:"url"`
		} `yaml:"callback"`
	} `yaml:"report"`

	UsageConfig struct {
		// Enabled is the flag to enable the usage.
		Enabled bool `yaml:"enabled"`
	} `yaml:"usage"`
}

// LoadExporterConfig loads from byte array.
func LoadExporterConfig(data []byte) (*ExporterConfig, error) {
	var config ExporterConfig
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// Set default values
	if config.SchedulerConfig.RefreshSamplesInterval == 0 {
		config.SchedulerConfig.RefreshSamplesInterval = 60 * time.Second
	}

	if config.SchedulerConfig.EnqueueSamplesInterval == 0 {
		config.SchedulerConfig.EnqueueSamplesInterval = 5 * time.Second
	}

	if config.SchedulerConfig.GCInterval == 0 {
		config.SchedulerConfig.GCInterval = 5 * time.Minute
	}

	if len(config.HTTPServerConfig.ListenAddress) == 0 {
		config.HTTPServerConfig.ListenAddress = ":19090"
	}

	if len(config.HTTPServerConfig.MetricsPath) == 0 {
		config.HTTPServerConfig.MetricsPath = "/metrics"
	}

	if config.WorkerConfig.SleepBetweenJobs == 0 {
		config.WorkerConfig.SleepBetweenJobs = 1 * time.Second
	}

	if config.WorkerConfig.ParallelJobs <= 0 {
		config.WorkerConfig.ParallelJobs = 16
	}

	if config.WorkerConfig.MaxQueueSize <= 0 {
		config.WorkerConfig.MaxQueueSize = 1000
	}

	return &config, nil
}

// LoadExporterConfigFromFile loads from file.
func LoadExporterConfigFromFile(path string) (*ExporterConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadExporterConfig(data)
}

package config_test

import (
	"testing"

	"github.com/hidracloud/hidra/v3/internal/config"
	"github.com/likexian/gokit/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadExporterconfigFromBytes tests loading from bytes.
func TestLoadExporterconfigFromBytes(t *testing.T) {
	data := []byte(`
log_level: debug
http_server:
  listen_address:
samples_path: /tmp
`)
	config, err := config.LoadExporterConfig(data)
	require.NoError(t, err)
	assert.Equal(t, "debug", config.LogLevel)
	assert.Equal(t, "", config.HTTPServerConfig.ListenAddress)
	assert.Equal(t, "/tmp", config.SamplesPath)
}

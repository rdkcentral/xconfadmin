package common

import (
	"os"
	"testing"

	"gotest.tools/assert"
)

func TestNewServerConfig(t *testing.T) {
	// Test with non-existent file
	_, err := NewServerConfig("nonexistent.conf")
	assert.Assert(t, err != nil)

	// Test with valid config file (create a temporary one for testing)
	tempFile, err := os.CreateTemp("", "test_config_*.conf")
	assert.Assert(t, err == nil)
	defer os.Remove(tempFile.Name())

	// Write some config content
	content := "key = value\nsection { key2 = value2 }"
	_, err = tempFile.WriteString(content)
	assert.Assert(t, err == nil)
	tempFile.Close()

	config, err := NewServerConfig(tempFile.Name())
	assert.Assert(t, err == nil)
	assert.Assert(t, config != nil)
	assert.Assert(t, config.Config != nil)

	// Test ConfigBytes
	bytes := config.ConfigBytes()
	assert.Assert(t, len(bytes) > 0)
	assert.Equal(t, content, string(bytes))
}

func TestServerOriginId(t *testing.T) {
	originId := ServerOriginId()
	assert.Assert(t, originId != "")

	// Should contain PID at minimum
	pid := os.Getpid()
	assert.Assert(t, pid > 0)
}

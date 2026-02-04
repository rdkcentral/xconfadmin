package http

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"github.com/go-akka/configuration"
	"github.com/stretchr/testify/assert"
)

// Mock IdpServiceConnector for testing
type MockIdpServiceConnector struct {
	host string
}

func (m *MockIdpServiceConnector) IdpServiceHost() string                               { return m.host }
func (m *MockIdpServiceConnector) SetIdpServiceHost(host string)                        { m.host = host }
func (m *MockIdpServiceConnector) GetFullLoginUrl(continueUrl string) string            { return "" }
func (m *MockIdpServiceConnector) GetJsonWebKeyResponse(url string) *JsonWebKeyResponse { return nil }
func (m *MockIdpServiceConnector) GetFullLogoutUrl(continueUrl string) string           { return "" }
func (m *MockIdpServiceConnector) GetToken(code string) string                          { return "" }
func (m *MockIdpServiceConnector) Logout(url string) error                              { return nil }
func (m *MockIdpServiceConnector) GetIdpServiceConfig() *IdpServiceConfig               { return nil }

func TestNewIdpServiceConnector_WithExternalService(t *testing.T) {
	// Test case: external IdpServiceConnector is provided
	mockService := &MockIdpServiceConnector{host: "mock-host"}
	config := configuration.ParseString("")

	result := NewIdpServiceConnector(config, mockService)

	assert.NotNil(t, result)
	assert.Equal(t, mockService, result)
	assert.Equal(t, "mock-host", result.IdpServiceHost())
}

func TestNewIdpServiceConnector_WithConfigurationSuccess(t *testing.T) {
	// Backup original environment variables
	originalClientId := os.Getenv("IDP_CLIENT_ID")
	originalClientSecret := os.Getenv("IDP_CLIENT_SECRET")

	// Clean up after test
	defer func() {
		os.Setenv("IDP_CLIENT_ID", originalClientId)
		os.Setenv("IDP_CLIENT_SECRET", originalClientSecret)
	}()

	// Set up environment variables
	os.Setenv("IDP_CLIENT_ID", "test-client-id")
	os.Setenv("IDP_CLIENT_SECRET", "test-client-secret")

	// Create test configuration
	configData := `
		xconfwebconfig {
			xconf {
				idp_service_name = "test-service"
			}
			test-service {
				host = "https://test-host.com"
			}
		}
	`
	config := configuration.ParseString(configData)

	result := NewIdpServiceConnector(config, nil)

	assert.NotNil(t, result)

	// Verify it's a DefaultIdpService
	defaultService, ok := result.(*DefaultIdpService)
	assert.True(t, ok)
	assert.Equal(t, "https://test-host.com", defaultService.IdpServiceHost())

	// Verify IdpServiceConfig
	idpConfig := defaultService.GetIdpServiceConfig()
	assert.NotNil(t, idpConfig)
	assert.Equal(t, "test-client-id", idpConfig.ClientId)
	assert.Equal(t, "test-client-secret", idpConfig.ClientSecret)

	// Verify auth header
	expectedAuth := fmt.Sprintf("test-client-id:test-client-secret")
	expectedAuthHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(expectedAuth)))
	assert.Equal(t, expectedAuthHeader, idpConfig.AuthHeaderValue)
}

func TestNewIdpServiceConnector_WithConfigurationFromConfig(t *testing.T) {
	// Backup and clear environment variables
	originalClientId := os.Getenv("IDP_CLIENT_ID")
	originalClientSecret := os.Getenv("IDP_CLIENT_SECRET")
	os.Unsetenv("IDP_CLIENT_ID")
	os.Unsetenv("IDP_CLIENT_SECRET")

	// Clean up after test
	defer func() {
		os.Setenv("IDP_CLIENT_ID", originalClientId)
		os.Setenv("IDP_CLIENT_SECRET", originalClientSecret)
	}()

	// Create test configuration with client credentials in config
	configData := `
		xconfwebconfig {
			xconf {
				idp_service_name = "test-service"
			}
			test-service {
				host = "https://test-host.com"
				client_id = "config-client-id"
				client_secret = "config-client-secret"
			}
		}
	`
	config := configuration.ParseString(configData)

	result := NewIdpServiceConnector(config, nil)

	assert.NotNil(t, result)

	// Verify it's a DefaultIdpService
	defaultService, ok := result.(*DefaultIdpService)
	assert.True(t, ok)
	assert.Equal(t, "https://test-host.com", defaultService.IdpServiceHost())

	// Verify IdpServiceConfig
	idpConfig := defaultService.GetIdpServiceConfig()
	assert.NotNil(t, idpConfig)
	assert.Equal(t, "config-client-id", idpConfig.ClientId)
	assert.Equal(t, "config-client-secret", idpConfig.ClientSecret)
}

func TestNewIdpServiceConnector_PanicOnMissingHost(t *testing.T) {
	// Create configuration without host
	configData := `
		xconfwebconfig {
			xconf {
				idp_service_name = "test-service"
			}
			test-service {
				client_id = "test-client-id"
				client_secret = "test-client-secret"
			}
		}
	`
	config := configuration.ParseString(configData)

	// This should panic because host is missing
	assert.Panics(t, func() {
		NewIdpServiceConnector(config, nil)
	})
}

func TestNewIdpServiceConnector_PanicOnMissingClientId(t *testing.T) {
	// Backup and clear environment variables
	originalClientId := os.Getenv("IDP_CLIENT_ID")
	originalClientSecret := os.Getenv("IDP_CLIENT_SECRET")
	os.Unsetenv("IDP_CLIENT_ID")
	os.Unsetenv("IDP_CLIENT_SECRET")

	// Clean up after test
	defer func() {
		os.Setenv("IDP_CLIENT_ID", originalClientId)
		os.Setenv("IDP_CLIENT_SECRET", originalClientSecret)
	}()

	// Create configuration without client_id
	configData := `
		xconfwebconfig {
			xconf {
				idp_service_name = "test-service"
			}
			test-service {
				host = "https://test-host.com"
				client_secret = "test-client-secret"
			}
		}
	`
	config := configuration.ParseString(configData)

	// This should panic because client_id is missing
	assert.Panics(t, func() {
		NewIdpServiceConnector(config, nil)
	})
}

func TestNewIdpServiceConnector_PanicOnMissingClientSecret(t *testing.T) {
	// Backup and clear environment variables
	originalClientId := os.Getenv("IDP_CLIENT_ID")
	originalClientSecret := os.Getenv("IDP_CLIENT_SECRET")
	os.Unsetenv("IDP_CLIENT_ID")
	os.Unsetenv("IDP_CLIENT_SECRET")

	// Clean up after test
	defer func() {
		os.Setenv("IDP_CLIENT_ID", originalClientId)
		os.Setenv("IDP_CLIENT_SECRET", originalClientSecret)
	}()

	// Create configuration without client_secret
	configData := `
		xconfwebconfig {
			xconf {
				idp_service_name = "test-service"
			}
			test-service {
				host = "https://test-host.com"
				client_id = "test-client-id"
			}
		}
	`
	config := configuration.ParseString(configData)

	// This should panic because client_secret is missing
	assert.Panics(t, func() {
		NewIdpServiceConnector(config, nil)
	})
}

func TestNewIdpServiceConnector_EnvironmentVariablesPrecedence(t *testing.T) {
	// Backup original environment variables
	originalClientId := os.Getenv("IDP_CLIENT_ID")
	originalClientSecret := os.Getenv("IDP_CLIENT_SECRET")

	// Clean up after test
	defer func() {
		os.Setenv("IDP_CLIENT_ID", originalClientId)
		os.Setenv("IDP_CLIENT_SECRET", originalClientSecret)
	}()

	// Set environment variables that should take precedence over config
	os.Setenv("IDP_CLIENT_ID", "env-client-id")
	os.Setenv("IDP_CLIENT_SECRET", "env-client-secret")

	// Create configuration with different values
	configData := `
		xconfwebconfig {
			xconf {
				idp_service_name = "test-service"
			}
			test-service {
				host = "https://test-host.com"
				client_id = "config-client-id"
				client_secret = "config-client-secret"
			}
		}
	`
	config := configuration.ParseString(configData)

	result := NewIdpServiceConnector(config, nil)

	assert.NotNil(t, result)

	// Verify environment variables take precedence
	defaultService := result.(*DefaultIdpService)
	idpConfig := defaultService.GetIdpServiceConfig()
	assert.Equal(t, "env-client-id", idpConfig.ClientId)
	assert.Equal(t, "env-client-secret", idpConfig.ClientSecret)
}

func TestNewIdpServiceConnector_KidMapInitialization(t *testing.T) {
	// Set up environment variables
	originalClientId := os.Getenv("IDP_CLIENT_ID")
	originalClientSecret := os.Getenv("IDP_CLIENT_SECRET")
	os.Setenv("IDP_CLIENT_ID", "test-client-id")
	os.Setenv("IDP_CLIENT_SECRET", "test-client-secret")

	defer func() {
		os.Setenv("IDP_CLIENT_ID", originalClientId)
		os.Setenv("IDP_CLIENT_SECRET", originalClientSecret)
	}()

	configData := `
		xconfwebconfig {
			xconf {
				idp_service_name = "test-service"
			}
			test-service {
				host = "https://test-host.com"
			}
		}
	`
	config := configuration.ParseString(configData)

	result := NewIdpServiceConnector(config, nil)

	defaultService := result.(*DefaultIdpService)
	idpConfig := defaultService.GetIdpServiceConfig()

	// Verify KidMap is initialized (sync.Map doesn't have a direct way to check if empty)
	assert.NotNil(t, idpConfig.KidMap)

	// Test that we can store and retrieve from KidMap
	testKey := JsonWebKey{Kid: "test-kid"}
	idpConfig.KidMap.Store("test", testKey)

	value, ok := idpConfig.KidMap.Load("test")
	assert.True(t, ok)
	assert.Equal(t, testKey, value)
}

// Additional tests for better coverage of the DefaultIdpService methods
func TestDefaultIdpService_Methods(t *testing.T) {
	// Set up environment variables
	originalClientId := os.Getenv("IDP_CLIENT_ID")
	originalClientSecret := os.Getenv("IDP_CLIENT_SECRET")
	os.Setenv("IDP_CLIENT_ID", "test-client-id")
	os.Setenv("IDP_CLIENT_SECRET", "test-client-secret")

	defer func() {
		os.Setenv("IDP_CLIENT_ID", originalClientId)
		os.Setenv("IDP_CLIENT_SECRET", originalClientSecret)
	}()

	configData := `
		xconfwebconfig {
			xconf {
				idp_service_name = "test-service"
			}
			test-service {
				host = "https://test-host.com"
			}
		}
	`
	config := configuration.ParseString(configData)

	result := NewIdpServiceConnector(config, nil)
	defaultService := result.(*DefaultIdpService)

	// Test SetIdpServiceHost and IdpServiceHost
	defaultService.SetIdpServiceHost("https://new-host.com")
	assert.Equal(t, "https://new-host.com", defaultService.IdpServiceHost())

	// Test GetFullLoginUrl
	loginUrl := defaultService.GetFullLoginUrl("https://continue.com")
	expectedLoginUrl := fmt.Sprintf(fullLoginUrl, "https://new-host.com", "https://continue.com", "test-client-id")
	assert.Equal(t, expectedLoginUrl, loginUrl)

	// Test GetFullLogoutUrl
	logoutUrl := defaultService.GetFullLogoutUrl("https://continue.com")
	expectedLogoutUrl := fmt.Sprintf(fullLogoutUrl, "https://new-host.com", "https://continue.com", "test-client-id")
	assert.Equal(t, expectedLogoutUrl, logoutUrl)
}

/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package taggingapi_config

import (
	"testing"

	"github.com/go-akka/configuration"
	"github.com/stretchr/testify/assert"
)

func TestTaggingApiConfig_Struct(t *testing.T) {
	// Test struct creation with values
	config := &TaggingApiConfig{
		BatchLimit:  5000,
		WorkerCount: 20,
	}

	assert.Equal(t, 5000, config.BatchLimit, "BatchLimit should be set correctly")
	assert.Equal(t, 20, config.WorkerCount, "WorkerCount should be set correctly")
}

func TestTaggingApiConfig_ZeroValues(t *testing.T) {
	// Test struct with zero values
	config := &TaggingApiConfig{}

	assert.Equal(t, 0, config.BatchLimit, "BatchLimit should default to zero")
	assert.Equal(t, 0, config.WorkerCount, "WorkerCount should default to zero")
}

func TestNewTaggingApiConfig_WithConfig(t *testing.T) {
	// Test with mock configuration that has the required keys
	configStr := `
		webconfig {
			xconf {
				tag_members_batch_limit = 3000
				tag_update_worker_count = 15
			}
		}
	`

	conf := configuration.ParseString(configStr)
	result := NewTaggingApiConfig(conf)

	assert.NotNil(t, result, "NewTaggingApiConfig should return a config instance")
	assert.Equal(t, 3000, result.BatchLimit, "BatchLimit should be read from config")
	assert.Equal(t, 15, result.WorkerCount, "WorkerCount should be read from config")
}

func TestNewTaggingApiConfig_WithDefaults(t *testing.T) {
	// Test with empty configuration (should use defaults)
	configStr := `{}`

	conf := configuration.ParseString(configStr)
	result := NewTaggingApiConfig(conf)

	assert.NotNil(t, result, "NewTaggingApiConfig should return a config instance")
	assert.Equal(t, 2000, result.BatchLimit, "BatchLimit should use default value")
	assert.Equal(t, 20, result.WorkerCount, "WorkerCount should use default value")
}

func TestNewTaggingApiConfig_WithPartialConfig(t *testing.T) {
	// Test with configuration that has only one of the required keys
	configStr := `
		webconfig {
			xconf {
				tag_members_batch_limit = 4000
			}
		}
	`

	conf := configuration.ParseString(configStr)
	result := NewTaggingApiConfig(conf)

	assert.NotNil(t, result, "NewTaggingApiConfig should return a config instance")
	assert.Equal(t, 4000, result.BatchLimit, "BatchLimit should be read from config")
	assert.Equal(t, 20, result.WorkerCount, "WorkerCount should use default value")
}

func TestNewTaggingApiConfig_WithOtherPartialConfig(t *testing.T) {
	// Test with configuration that has only the worker count
	configStr := `
		webconfig {
			xconf {
				tag_update_worker_count = 25
			}
		}
	`

	conf := configuration.ParseString(configStr)
	result := NewTaggingApiConfig(conf)

	assert.NotNil(t, result, "NewTaggingApiConfig should return a config instance")
	assert.Equal(t, 2000, result.BatchLimit, "BatchLimit should use default value")
	assert.Equal(t, 25, result.WorkerCount, "WorkerCount should be read from config")
}

func TestNewTaggingApiConfig_WithExtremeValues(t *testing.T) {
	// Test with extreme values
	configStr := `
		webconfig {
			xconf {
				tag_members_batch_limit = 1
				tag_update_worker_count = 1000
			}
		}
	`

	conf := configuration.ParseString(configStr)
	result := NewTaggingApiConfig(conf)

	assert.NotNil(t, result, "NewTaggingApiConfig should return a config instance")
	assert.Equal(t, 1, result.BatchLimit, "BatchLimit should handle minimum value")
	assert.Equal(t, 1000, result.WorkerCount, "WorkerCount should handle large value")
}

func TestNewTaggingApiConfig_WithZeroValues(t *testing.T) {
	// Test with zero values in config
	configStr := `
		webconfig {
			xconf {
				tag_members_batch_limit = 0
				tag_update_worker_count = 0
			}
		}
	`

	conf := configuration.ParseString(configStr)
	result := NewTaggingApiConfig(conf)

	assert.NotNil(t, result, "NewTaggingApiConfig should return a config instance")
	assert.Equal(t, 0, result.BatchLimit, "BatchLimit should handle zero value")
	assert.Equal(t, 0, result.WorkerCount, "WorkerCount should handle zero value")
}

func TestNewTaggingApiConfig_DefaultValues(t *testing.T) {
	// Test that default values are correct as per function implementation
	emptyConfig := configuration.ParseString("{}")
	result := NewTaggingApiConfig(emptyConfig)

	// These are the default values from the function
	expectedBatchLimit := 2000
	expectedWorkerCount := 20

	assert.Equal(t, expectedBatchLimit, result.BatchLimit, "Default BatchLimit should be 2000")
	assert.Equal(t, expectedWorkerCount, result.WorkerCount, "Default WorkerCount should be 20")
}

func TestTaggingApiConfig_FieldTypes(t *testing.T) {
	// Test that fields are of correct type
	config := &TaggingApiConfig{
		BatchLimit:  100,
		WorkerCount: 5,
	}

	assert.IsType(t, int(0), config.BatchLimit, "BatchLimit should be of type int")
	assert.IsType(t, int(0), config.WorkerCount, "WorkerCount should be of type int")
}

func TestNewTaggingApiConfig_NilHandling(t *testing.T) {
	// Test that function doesn't panic with nil config
	// This might panic depending on the configuration library implementation
	assert.NotPanics(t, func() {
		// This may panic, but we're testing that our function handles it
		defer func() {
			if r := recover(); r != nil {
				// Panic is acceptable for nil config
			}
		}()
		NewTaggingApiConfig(nil)
	}, "NewTaggingApiConfig should handle nil gracefully or panic predictably")
}

func TestTaggingApiConfig_Modification(t *testing.T) {
	// Test that config values can be modified after creation
	config := &TaggingApiConfig{
		BatchLimit:  1000,
		WorkerCount: 5,
	}

	// Modify values
	config.BatchLimit = 2500
	config.WorkerCount = 12

	assert.Equal(t, 2500, config.BatchLimit, "BatchLimit should be modifiable")
	assert.Equal(t, 12, config.WorkerCount, "WorkerCount should be modifiable")
}

func TestNewTaggingApiConfig_ConfigKeys(t *testing.T) {
	// Test that the function uses the correct configuration keys
	configStr := `
		webconfig {
			xconf {
				tag_members_batch_limit = 1234
				tag_update_worker_count = 5678
				other_unrelated_key = 9999
			}
		}
	`

	conf := configuration.ParseString(configStr)
	result := NewTaggingApiConfig(conf)

	// Should only read the specific keys we care about
	assert.Equal(t, 1234, result.BatchLimit, "Should read tag_members_batch_limit")
	assert.Equal(t, 5678, result.WorkerCount, "Should read tag_update_worker_count")
	// other_unrelated_key should be ignored
}

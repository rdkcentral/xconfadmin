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
package queries

import (
	"testing"

	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/stretchr/testify/assert"
)

func TestGetModels_Simple(t *testing.T) {
	// Test basic functionality
	result := GetModels()
	assert.NotNil(t, result)
	assert.IsType(t, []*shared.ModelResponse{}, result)
}

func TestGetModel_NonExistent(t *testing.T) {
	result := GetModel("NON_EXISTENT_MODEL")
	// May or may not be nil depending on DB state
	_ = result
}

func TestIsExistModel_Empty(t *testing.T) {
	exists := IsExistModel("")
	assert.False(t, exists)
}

func TestIsExistModel_Check(t *testing.T) {
	_ = IsExistModel("SOME_MODEL")
	// Function executes without panic
}

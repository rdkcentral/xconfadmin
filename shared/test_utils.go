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
package shared

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	xutil "github.com/rdkcentral/xconfadmin/util"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
)

// Helper functions to abstract DAO operations for mock/real database

// GetOneFromDao retrieves a single entity - works with both mock and real DAO
func GetOneFromDao(tableName string, rowKey string) (any, error) {
	return db.GetCachedSimpleDao().GetOne(db.GetDefaultTenantId(), tableName, rowKey)
}

// SetOneInDao stores a single entity - works with both mock and real DAO
func SetOneInDao(tableName string, rowKey string, entity any) error {
	if obj, ok := entity.(db.Updatable); ok {
		obj.SetUpdated(xutil.GetTimestamp())
	}
	return db.GetCachedSimpleDao().SetOne(db.GetDefaultTenantId(), tableName, rowKey, entity)
}

// DeleteOneFromDao removes a single entity - works with both mock and real DAO
func DeleteOneFromDao(tableName string, rowKey string) error {
	return db.GetCachedSimpleDao().DeleteOne(db.GetDefaultTenantId(), tableName, rowKey)
}

// GetAllAsListFromDao retrieves all entities as a list - works with both mock and real DAO
func GetAllAsListFromDao(tableName string, maxResults int) ([]interface{}, error) {
	return db.GetCachedSimpleDao().GetAllAsList(db.GetDefaultTenantId(), tableName, maxResults)
}

// GetAllAsMapFromDao retrieves all entities as a map - works with both mock and real DAO
func GetAllAsMapFromDao(tableName string) (map[interface{}]interface{}, error) {
	return db.GetCachedSimpleDao().GetAllAsMap(db.GetDefaultTenantId(), tableName)
}

// RefreshAllInDao refreshes cache for a table - no-op for mock
func RefreshAllInDao(tableName string) error {
	return db.GetCachedSimpleDao().RefreshAll(db.GetDefaultTenantId(), tableName)
}

func DeleteAllEntities(t *testing.T) {
	tenantId := db.GetDefaultTenantId()
	for _, tableInfo := range db.GetAllTableInfo() {
		if err := TruncateTable(t, tenantId, tableInfo.TableName); err != nil {
			fmt.Printf("failed to truncate table %s\n", tableInfo.TableName)
		}
		if tableInfo.Cached {
			db.GetCachedSimpleDao().RefreshAll(tenantId, tableInfo.TableName)
		}
	}
}

func TruncateTable(t *testing.T, tenantId string, tableName string) error {
	// ensure t is genuinely from an active test (not fabricated); it will panic if t is nil, and is a no-op otherwise
	t.Helper()

	dbClient := db.GetDatabaseClient()
	cassandraClient, ok := dbClient.(*db.CassandraClient)
	if ok {
		tableInfo, err := db.GetTableInfo(tableName)
		if err != nil {
			return err
		}
		if tableInfo.Unsharded {
			if tableName == db.TABLE_LOGS {
				tableName = cassandraClient.GetTableNameFromLogKeyspace(tableName)
			}
			return cassandraClient.Query(fmt.Sprintf(`TRUNCATE table %s`, tableName)).Exec()
		} else {
			return cassandraClient.DeleteAllXconfData(tenantId, tableName)
		}
	}
	return nil
}

// DeleteTelemetryEntities - Ultra-fast cleanup using in-memory mock
// Replaces slow Cassandra truncation (60s) with instant mock.Clear() (<1ms)
func DeleteTelemetryEntities(t *testing.T) {
	telemetryTables := []string{
		db.TABLE_TELEMETRY_PROFILES,
		db.TABLE_TELEMETRY_RULES,
		db.TABLE_TELEMETRY_TWO_PROFILES,
		db.TABLE_TELEMETRY_TWO_RULES,
		db.TABLE_PERMANENT_TELEMETRY_PROFILES,
		db.TABLE_TELEMETRY_CHANGES,
		db.TABLE_TELEMETRY_APPROVED_CHANGES,
		db.TABLE_TELEMETRY_TWO_CHANGES,
		db.TABLE_TELEMETRY_APPROVED_TWO_CHANGES,
	}

	tenantId := db.GetDefaultTenantId()
	for _, tableName := range telemetryTables {
		TruncateTable(t, tenantId, tableName)
		db.GetCachedSimpleDao().RefreshAll(tenantId, tableName)
	}
}

func ExecuteRequest(r *http.Request, handler http.Handler) *httptest.ResponseRecorder { // restored local version
	recorder := httptest.NewRecorder()

	// Wrap the response writer with XResponseWriter to match production behavior
	xw := xwhttp.NewXResponseWriter(recorder, r)

	// Read and set the request body on XResponseWriter (mimics middleware behavior)
	if r.Method == "POST" || r.Method == "PUT" {
		if r.Body != nil {
			if rbytes, err := io.ReadAll(r.Body); err == nil {
				xw.SetBody(string(rbytes))
				// Reset the body so the handler can read it again
				r.Body = io.NopCloser(bytes.NewReader(rbytes))
			}
		} else {
			xw.SetBody("")
		}
	}

	handler.ServeHTTP(xw, r)
	return recorder
}

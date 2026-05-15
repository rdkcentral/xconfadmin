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
package common

import (
	"fmt"

	"github.com/rdkcentral/xconfwebconfig/db"
)

// TruncateTable removes all data from a single table and refreshes its cache.
// This is the atomic operation used by all package-specific cleanup functions.
// If using a real Cassandra database, calls DeleteAllXconfData on the table.
// Always calls RefreshAll on the cache manager.
func TruncateTable(tableName string) error {
	dbClient := db.GetDatabaseClient()
	cassandraClient, ok := dbClient.(*db.CassandraClient)
	if ok {
		if err := cassandraClient.DeleteAllXconfData(tableName); err != nil {
			fmt.Printf("failed to truncate table %s: %v\n", tableName, err)
			return err
		}
	}
	return db.GetCachedSimpleDao().RefreshAll(tableName)
}

// TruncateAndRefresh removes all data from multiple tables and refreshes their caches.
// This is the pattern all scoped cleanup functions should use.
// Call this in your package-specific cleanup helpers (e.g., CleanupDCMFormulaTables).
func TruncateAndRefresh(tableNames []string) error {
	for _, tableName := range tableNames {
		if err := TruncateTable(tableName); err != nil {
			return err
		}
	}
	return nil
}

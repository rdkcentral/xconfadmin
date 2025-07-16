// Copyright 2025 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package queries

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	//"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	util "github.com/rdkcentral/xconfwebconfig/util"

	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	xcommon "github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	xutil "github.com/rdkcentral/xconfadmin/util"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"

	"github.com/gorilla/mux"
)

const (
	IMPORTED     = "IMPORTED"
	NOT_IMPORTED = "NOT_IMPORTED"
)

type Change struct {
	ChangedKey string           `json:"changedKey"`
	Operation  db.OperationType `json:"operationType"`
	CfName     string           `json:"cfName"`
	UserName   string           `json:"userName"`
}

// CacheStats statistics for the cache.LoadingCache
type CacheStats struct {
	DaoRefreshTime int64         `json:"daoRefreshTime"`
	CacheSize      int           `json:"cacheSize"`
	NonAbsentCount int           `json:"nonAbsentCount"`
	RequestCount   uint64        `json:"requestCount"`
	EvictionCount  uint64        `json:"evictionCount"`
	HitRate        float64       `json:"hitRate"`
	MissRate       float64       `json:"missRate"`
	TotalLoadTime  time.Duration `json:"totalLoadTime"`
}

// Statistics cache statistics
type Statistics struct {
	StatsMap map[string]CacheStats `json:"Statistics"`
}

func GetInfoTableNames(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.TOOL_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	tables := make(map[string]interface{})
	for _, tableInfo := range db.GetAllTableInfo() {
		entry := make(map[string]bool)
		entry["Split"] = tableInfo.Split
		entry["Compress"] = tableInfo.Compress
		entry["CacheData"] = tableInfo.CacheData
		tables[tableInfo.TableName] = entry
	}
	response, _ := util.JSONMarshal(tables)
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetInfoTable(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.TOOL_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	tableName := mux.Vars(r)[xcommon.TABLE_NAME]
	tableInfo, _ := db.GetTableInfo(tableName)
	if tableInfo == nil {
		xwhttp.WriteXconfResponse(w, http.StatusNotFound, []byte(fmt.Sprintf("Not found table definition: %s", tableName)))
		return
	}

	if tableInfo.IsCompressOnly() {
		xwhttp.WriteXconfResponse(w, http.StatusNotImplemented, []byte("Listing table not supported"))
		return
	}

	_, cacheData := r.URL.Query()["cache"]

	// Get Data from DB
	var data map[string]interface{}
	var err error

	if tableName == db.TABLE_XCONF_CHANGED_KEYS {
		if cacheData {
			xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Data is not cached for this table"))
			return
		}

		data, err = GetChangedKeysMapRaw()
	} else {
		if cacheData {
			res, err := db.GetCachedSimpleDao().GetAllAsMap(tableInfo.TableName)
			if err == nil {
				cacheData := make(map[string]interface{})
				for k, v := range res {
					cacheData[k.(string)] = v
				}

				response, _ := util.JSONMarshal(cacheData)
				xwhttp.WriteXconfResponse(w, http.StatusOK, response)
				return
			}
		} else {
			// Use the appropriate db based on compression policy
			if tableInfo.IsCompressAndSplit() {
				data, err = db.GetCompressingDataDao().GetAllAsMap(tableInfo.TableName)
			} else {
				data, err = db.GetSimpleDao().GetAllAsMap(tableInfo.TableName, 0)
			}
		}
	}

	if err != nil {
		xwhttp.WriteXconfResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}

	response, _ := util.JSONMarshal(data)
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetInfoTableRowKey(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.TOOL_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	tableName := mux.Vars(r)[xcommon.TABLE_NAME]
	tableInfo, _ := db.GetTableInfo(tableName)
	if tableInfo == nil {
		xwhttp.WriteXconfResponse(w, http.StatusNotFound, []byte(fmt.Sprintf("Not found table definition: %s", tableName)))
		return
	}

	if tableInfo.IsCompressOnly() {
		xwhttp.WriteXconfResponse(w, http.StatusNotImplemented, []byte("Listing table not supported"))
		return
	}

	rowKey := mux.Vars(r)[xcommon.ROW_KEY]

	// Get Data from DB
	var data interface{}
	var err error

	if _, found := r.URL.Query()["cache"]; found {
		if !tableInfo.CacheData {
			xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Data is not cached for this table"))
			return
		}

		data, err = db.GetCachedSimpleDao().GetOne(tableInfo.TableName, rowKey)
	} else {
		// Use the appropriate db based on compression policy
		if tableInfo.IsCompressAndSplit() {
			data, err = db.GetCompressingDataDao().GetOne(tableInfo.TableName, rowKey)
		} else {
			data, err = db.GetSimpleDao().GetOne(tableInfo.TableName, rowKey)
		}
	}

	if err != nil {
		xwhttp.WriteXconfResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}

	response, _ := util.JSONMarshal(data)
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

// Get all ChangedKeys records in the last 15 minutes as raw JSON data
func GetChangedKeysMapRaw() (map[string]interface{}, error) {
	changedKeysTimeWindowSize := db.GetCacheManager().GetChangedKeysTimeWindowSize()

	endTS := util.GetTimestamp(time.Now().UTC())
	endRowKey := endTS - (endTS % int64(changedKeysTimeWindowSize))

	startTS := xutil.UtcOffsetPriorMinTimestamp(15)
	currentRowKey := startTS - (startTS % int64(changedKeysTimeWindowSize))

	startUuid, err := util.UUIDFromTime(startTS, 0, 0)
	if err != nil {
		return nil, err
	}

	ranges := make(map[int64]*db.RangeInfo)
	ranges[currentRowKey] = &db.RangeInfo{StartValue: startUuid}
	currentRowKey += int64(changedKeysTimeWindowSize)
	for currentRowKey <= endRowKey {
		ranges[currentRowKey] = nil
		currentRowKey += int64(changedKeysTimeWindowSize)
	}

	data := make(map[string]interface{})

	for rowKey, rangeInfo := range ranges {
		list, err := db.GetListingDao().GetRange(db.TABLE_XCONF_CHANGED_KEYS, rowKey, rangeInfo)
		if err != nil {
			return nil, err
		}

		for i, jsonData := range list {
			data[fmt.Sprintf("%d-%d", rowKey, i)] = jsonData
		}
	}

	return data, nil
}

// This API can be used to update the raw JSON data in a table;
// should only be used for fixing table incompatibilities issues between Java and Go
func UpdateInfoTableRowKey(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanWrite(r, auth.TOOL_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	tableName := mux.Vars(r)[xcommon.TABLE_NAME]
	tableInfo, _ := db.GetTableInfo(tableName)
	if tableInfo == nil {
		xwhttp.WriteXconfResponse(w, http.StatusNotFound, []byte(fmt.Sprintf("Not found table definition: %s", tableName)))
		return
	}

	if tableInfo.IsCompressOnly() {
		xwhttp.WriteXconfResponse(w, http.StatusNotImplemented, []byte("Listing table not supported"))
		return
	}

	rowKey := mux.Vars(r)[xcommon.ROW_KEY]

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("Unable to extract Body"))
		return
	}

	replacer := strings.NewReplacer("\n", "", "\r", "", "\t", "")
	jsonData := []byte(replacer.Replace(xw.Body()))

	if !json.Valid(jsonData) {
		xwhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("JSON is not valid"))
		return
	}

	// Update data the DB as Json Data; first ensure the record exists
	// by using the GetOneRaw function and avoid unmarshalling the object
	var err error
	if tableInfo.IsCompressAndSplit() {
		_, err = db.GetCompressingDataDao().GetOne(tableName, rowKey)
		if err == nil {
			err = db.GetCompressingDataDao().SetOne(tableName, rowKey, jsonData)
		}
	} else {
		_, err = db.GetSimpleDao().GetOne(tableName, rowKey)
		if err == nil {
			err = db.GetSimpleDao().SetOne(tableName, rowKey, jsonData)
		}
	}
	if err != nil {
		xwhttp.WriteXconfResponse(w, http.StatusInternalServerError, []byte(fmt.Sprintf("Unable to update record: %s", err.Error())))
		return
	}

	if tableInfo.CacheData {
		// Write cache changed log
		cm := db.GetCacheManager()
		cm.WriteCacheLog(tableName, rowKey, db.UPDATE_OPERATION)

		// Refresh the cache entry
		if err = db.GetCachedSimpleDao().RefreshOne(tableInfo.TableName, rowKey); err != nil {
			xwhttp.WriteXconfResponse(w, http.StatusInternalServerError, []byte(fmt.Sprintf("Unable to refresh cache entry: %s", err.Error())))
		}
	}
}

func GetChangeLogForTheDay(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.TOOL_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	changeLogMap, _ := GetChangeLog()
	response, _ := util.JSONMarshal(changeLogMap)
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetChangeLog() (result map[int64][]Change, err error) {
	//period of time in nano-seconds to group changed data in result map
	intervalStep := time.Duration(60 * 60 * time.Microsecond * time.Millisecond)

	currentTime := time.Now().UTC()
	year, month, day := currentTime.Date()
	startOfInterval := time.Date(year, month, day, 0, 0, 0, 0, currentTime.Location())

	endOfInterval := startOfInterval
	result = make(map[int64][]Change)
	lastDataChunk := false
	for !lastDataChunk {
		startOfInterval = endOfInterval
		endOfInterval = startOfInterval.Add(intervalStep)
		if endOfInterval.After(currentTime) {
			endOfInterval = currentTime.Add(intervalStep)
			lastDataChunk = true
		}
		changedList, err := db.GetCacheManager().SyncChanges(startOfInterval, endOfInterval, false)
		cdList := []Change{}
		for _, obj := range changedList {
			cd := obj.(*db.ChangedData)
			chg := Change{
				ChangedKey: cd.ChangedKey,
				Operation:  cd.Operation,
				CfName:     cd.CfName,
				UserName:   cd.UserName,
			}
			cdList = append(cdList, chg)
		}
		result[startOfInterval.UnixNano()/int64(time.Millisecond)] = cdList
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func GetStats(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.TOOL_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	stats := db.GetCacheManager().GetStatistics()
	response, _ := util.JSONMarshal(stats.CacheMap)
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetInfoStatistics(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.TOOL_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	stats := *db.GetCacheManager().GetStatistics()
	response, _ := util.JSONMarshal(stats)
	xhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func GetAppSettings(w http.ResponseWriter, r *http.Request) {
	// For retrieving app settings, tools permission is required
	if !auth.HasReadPermissionForTool(r) {
		xhttp.WriteAdminErrorResponse(w, http.StatusUnauthorized, "")
		return
	}

	settings, err := xcommon.GetAppSettings()
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	response, _ := util.JSONMarshal(settings)
	xwhttp.WriteXconfResponse(w, http.StatusOK, response)
}

func UpdateAppSettings(w http.ResponseWriter, r *http.Request) {
	// For updating app settings, tools permission is required
	if !auth.HasWritePermissionForTool(r) {
		xhttp.WriteAdminErrorResponse(w, http.StatusUnauthorized, "")
	}

	// r.Body is already drained in the middleware
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.AdminError(w, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, "responsewriter cast error"))
		return
	}
	body := xw.Body()
	settings := make(map[string]interface{})
	err := json.Unmarshal([]byte(body), &settings)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	for k, v := range settings {
		if !xcommon.IsValidAppSetting(k) {
			xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid AppSetting: %s", k))
			return
		}
		if _, err := xcommon.SetAppSetting(k, v); err != nil {
			xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Unable to save AppSetting for %s: %s", k, err.Error()))
			return
		}
	}

	xwhttp.WriteXconfResponse(w, http.StatusNoContent, nil)
}

func GetInfoRefreshAllHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.TOOL_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	failedToRefreshTables := db.GetCacheManager().RefreshAll()
	if len(failedToRefreshTables) == 0 {
		stats := db.GetCacheManager().GetStatistics()
		response, _ := util.JSONMarshal(stats.CacheMap)
		xhttp.WriteXconfResponse(w, http.StatusOK, response)
	} else {
		xhttp.WriteXconfResponse(w, http.StatusInternalServerError, []byte(fmt.Sprintf("\"Couldn't refresh caches for tables: %s\"", strings.Join(failedToRefreshTables, ", "))))
	}
}

func GetInfoRefreshHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.TOOL_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	tableName := mux.Vars(r)[xcommon.TABLE_NAME]
	err := db.GetCacheManager().Refresh(tableName)
	if err == nil {
		if stats, err := db.GetCacheManager().GetCacheStats(tableName); err == nil {
			response, _ := util.JSONMarshal(stats)
			xhttp.WriteXconfResponse(w, http.StatusOK, response)
		} else {
			xhttp.WriteXconfResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		}
	} else {
		xhttp.WriteXconfResponse(w, http.StatusInternalServerError, []byte(err.Error()))
	}
}

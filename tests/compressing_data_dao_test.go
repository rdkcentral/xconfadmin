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
package tests

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/util"
	"gotest.tools/assert"
)

var humptyStrList = []string{
	"Humpty Dumpty sat on a wall",
	"Humpty Dumpty had a great fall",
	"All the king's horses and all the king's men",
	"Couldn't put Humpty together again",
}

func TestCompressingDataCRUD(t *testing.T) {
	nl := shared.NewGenericNamespacedList(fmt.Sprintf("NL-%s", uuid.New().String()), "STRING", humptyStrList)

	// test create
	jsonData, err := json.Marshal(nl)
	assert.NilError(t, err)

	err = ds.GetCompressingDataDao().SetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID, jsonData)
	assert.NilError(t, err)

	// test retrieve
	obj, err := ds.GetCompressingDataDao().GetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID)
	assert.NilError(t, err)
	assert.Assert(t, obj != nil)

	resNL := *obj.(*shared.GenericNamespacedList)
	assert.Equal(t, resNL.ID, nl.ID)
	assert.Equal(t, len(resNL.Data), len(humptyStrList))
	assert.Assert(t, util.StringElementsMatch(resNL.Data, humptyStrList))

	// test update
	macList := []string{
		util.GenerateRandomCpeMac(),
		util.GenerateRandomCpeMac(),
		util.GenerateRandomCpeMac(),
	}

	nl.TypeName = "MAC_LIST"
	nl.Data = macList

	jsonData, err = json.Marshal(nl)
	assert.NilError(t, err)

	err = ds.GetCompressingDataDao().SetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID, jsonData)
	assert.NilError(t, err)

	obj, err = ds.GetCompressingDataDao().GetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID)
	assert.NilError(t, err)
	assert.Assert(t, obj != nil)

	resNL = *obj.(*shared.GenericNamespacedList)
	assert.Equal(t, resNL.ID, nl.ID)
	assert.Equal(t, len(resNL.Data), len(macList))
	assert.Assert(t, util.StringElementsMatch(resNL.Data, macList))

	// test delete
	err = ds.GetCompressingDataDao().DeleteOne(ds.TABLE_GENERIC_NS_LIST, resNL.ID)
	assert.NilError(t, err)

	_, err = ds.GetCompressingDataDao().GetOne(ds.TABLE_GENERIC_NS_LIST, resNL.ID)
	assert.Assert(t, errors.Is(err, gocql.ErrNotFound))
}

func TestCompressingDataGetAllByKeys(t *testing.T) {
	// generate some data
	keys, err := generateTestNamespacedList(5)
	assert.NilError(t, err)
	assert.Assert(t, len(keys) == 5)

	rowKeys := keys[0:3]
	list, err := ds.GetCompressingDataDao().GetAllByKeys(ds.TABLE_GENERIC_NS_LIST, rowKeys)
	assert.NilError(t, err)
	assert.Equal(t, len(list), len(rowKeys))

	for _, obj := range list {
		nl := *obj.(*shared.GenericNamespacedList)
		assert.Assert(t, util.Contains(rowKeys, nl.ID))
	}
}

func TestCompressingDataGetAllAsList(t *testing.T) {
	truncateTable(ds.TABLE_GENERIC_NS_LIST)

	// generate some data
	keys, err := generateTestNamespacedList(5)
	assert.NilError(t, err)
	assert.Assert(t, len(keys) == 5)

	list, err := ds.GetCompressingDataDao().GetAllAsList(ds.TABLE_GENERIC_NS_LIST, false)
	assert.NilError(t, err)
	assert.Equal(t, len(list), len(keys))

	for _, obj := range list {
		nl := *obj.(*shared.GenericNamespacedList)
		assert.Assert(t, util.Contains(keys, nl.ID))
	}
}

func TestCompressingDataGetAllAsMap(t *testing.T) {
	truncateTable(ds.TABLE_GENERIC_NS_LIST)

	// generate some data
	keys, err := generateTestNamespacedList(5)
	assert.NilError(t, err)
	assert.Assert(t, len(keys) == 5)

	nlMap, err := ds.GetCompressingDataDao().GetAllAsMap(ds.TABLE_GENERIC_NS_LIST, false)
	assert.NilError(t, err)
	assert.Equal(t, len(nlMap), len(keys))

	for _, key := range keys {
		assert.Assert(t, nlMap[key] != nil)
	}
}

func TestCompressingDataGetKeys(t *testing.T) {
	truncateTable(ds.TABLE_GENERIC_NS_LIST)

	// generate some data
	keys, err := generateTestNamespacedList(5)
	assert.NilError(t, err)
	assert.Assert(t, len(keys) == 5)

	rowKeys := ds.GetCompressingDataDao().GetKeys(ds.TABLE_GENERIC_NS_LIST)
	assert.NilError(t, err)
	assert.Equal(t, len(rowKeys), len(keys))
	assert.Assert(t, util.StringElementsMatch(keys, rowKeys), fmt.Sprintf("%v : %v", keys, rowKeys))
}

func TestCompressingDataMultipleParts(t *testing.T) {
	strList := make([]string, 2000)
	for i := 0; i < 2000; i++ {
		strList = append(strList, uuid.New().String())
	}

	nl := shared.NewGenericNamespacedList(uuid.New().String(), "STRING", strList)

	// test create
	jsonData, err := json.Marshal(nl)
	assert.NilError(t, err)

	err = ds.GetCompressingDataDao().SetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID, jsonData)
	assert.NilError(t, err)

	// test retrieve
	obj, err := ds.GetCompressingDataDao().GetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID)
	assert.NilError(t, err)
	assert.Assert(t, obj != nil)

	resNL := *obj.(*shared.GenericNamespacedList)
	assert.Equal(t, resNL.ID, nl.ID)
	assert.Equal(t, len(resNL.Data), len(strList))
	assert.Assert(t, util.StringElementsMatch(resNL.Data, strList))
}

func generateTestNamespacedList(num int) ([]string, error) {
	var keys []string
	for i := 0; i < num; i++ {
		data := []string{
			util.GenerateRandomCpeMac(),
			util.GenerateRandomCpeMac(),
		}
		id := fmt.Sprintf("NL-%s", uuid.New().String())
		nl := shared.NewGenericNamespacedList(id, "MAC_LIST", data)
		jsonData, err := json.Marshal(nl)
		if err != nil {
			return nil, err
		}

		err = ds.GetCompressingDataDao().SetOne(ds.TABLE_GENERIC_NS_LIST, nl.ID, jsonData)
		if err != nil {
			return nil, err
		}

		keys = append(keys, id)
	}
	return keys, nil
}

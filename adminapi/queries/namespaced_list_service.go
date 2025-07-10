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
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	xwhttp "xconfwebconfig/http"
	ru "xconfwebconfig/rulesengine"
	xutil "xconfwebconfig/util"

	"xconfadmin/common"
	xrfc "xconfadmin/shared/rfc"
	"xconfadmin/util"
	ds "xconfwebconfig/db"
	re "xconfwebconfig/rulesengine"
	"xconfwebconfig/shared"
	corefw "xconfwebconfig/shared/firmware"
	firmware "xconfwebconfig/shared/firmware"
	"xconfwebconfig/shared/rfc"

	log "github.com/sirupsen/logrus"
)

var ruleTables = []string{
	ds.TABLE_DCM_RULE,
	ds.TABLE_FIRMWARE_RULE,
	ds.TABLE_FIRMWARE_RULE_TEMPLATE,
	ds.TABLE_TELEMETRY_RULES,
	ds.TABLE_TELEMETRY_TWO_RULES,
	ds.TABLE_FEATURE_CONTROL_RULE,
	ds.TABLE_SETTING_RULES,
}

const (
	STRING      = "STRING"
	MAC_LIST    = "MAC_LIST"
	IP_LIST     = "IP_LIST"
	RI_MAC_LIST = "RI_MAC_LIST"
)

func GetNamespacedListIdsByType(typeName string) []string {
	var list []*shared.GenericNamespacedList
	var err error
	if typeName == "" {
		list, err = shared.GetGenericNamedListListsDB()
	} else {
		list, err = shared.GetGenericNamedListListsByTypeDB(typeName)
	}
	if err != nil {
		log.Error(fmt.Sprintf("GetNamespacedLists: %v", err))
		return []string{}
	}

	result := make([]string, len(list))
	for i, nl := range list {
		result[i] = nl.ID
	}
	return result
}

func GetNamespacedListsByType(typeName string) []*shared.GenericNamespacedList {
	var list []*shared.GenericNamespacedList
	var err error
	if typeName == "" {
		list, err = shared.GetGenericNamedListListsDB()
	} else {
		list, err = shared.GetGenericNamedListListsByTypeDB(typeName)
	}
	if err != nil {
		log.Error(fmt.Sprintf("GetNamespacedLists: %v", err))
		return []*shared.GenericNamespacedList{}
	}

	return list
}

func GetNamespacedListById(id string) *shared.GenericNamespacedList {
	nl, err := shared.GetGenericNamedListOneDB(id)
	if err != nil {
		log.Error(fmt.Sprintf("GetNamespacedListById: %v", err))
		return nil
	}

	return nl
}

func GetNamespacedListByIdAndType(id string, typeName string) *shared.GenericNamespacedList {
	nl := GetNamespacedListById(id)
	if nl == nil || nl.TypeName != typeName {
		return nil
	}

	return nl
}

func GetNamespacedListsByIp(ip string) []*shared.GenericNamespacedList {
	result := []*shared.GenericNamespacedList{}
	list, err := shared.GetGenericNamedListListsByTypeDB(shared.IP_LIST)
	if err != nil {
		log.Error(fmt.Sprintf("GetNamespacedListsByIp: %v", err))
		return result
	}
	for _, nl := range list {
		ipGrp := shared.NewIpAddressGroupWithAddrStrings(nl.ID, nl.ID, nl.Data)
		if ipGrp.IsInRange(ip) {
			ipGrp.RawIpAddresses = nl.Data // For the response, we need list of ip address as string
			result = append(result, nl)
		}
	}
	return result
}

func GetMacListsByMacPart(macAddress string) []*shared.GenericNamespacedList {
	result := []*shared.GenericNamespacedList{}
	list, err := shared.GetGenericNamedListListsByTypeDB(shared.MAC_LIST)
	if err != nil {
		log.Error(fmt.Sprintf("GetMacListsByMac: %v", err))
		return result
	}
	for _, nl := range list {
		if isMacListHasMacPart(macAddress, nl.Data) {
			result = append(result, nl)
		}
	}
	return result
}

func GetNamespacedListsByContext(searchContext map[string]string) []*shared.GenericNamespacedList {
	lists, err := shared.GetGenericNamedListListsDB()
	if err != nil {
		log.Error(fmt.Sprintf("GetMacListsByMac: %v", err))
		return []*shared.GenericNamespacedList{}
	}

	filteredLists := make([]*shared.GenericNamespacedList, 0, len(lists))

	for _, list := range lists {
		if name, ok := util.FindEntryInContext(searchContext, common.NAME_UPPER, false); ok {
			if !strings.Contains(strings.ToLower(list.ID), strings.ToLower(name)) {
				continue
			}
		}
		if TypeName, ok := util.FindEntryInContext(searchContext, common.TYPE_UPPER, false); ok {
			if list.TypeName != TypeName {
				continue
			}
		}
		if data, ok := util.FindEntryInContext(searchContext, common.DATA_UPPER, false); ok {
			if list.IsIpList() {
				if !isIpAddressHasIpPart(data, list.Data) {
					continue
				}
			} else if list.IsMacList() {
				if !isMacListHasMacPart(data, list.Data) {
					continue
				}
			}
		}
		filteredLists = append(filteredLists, list)
	}
	return filteredLists
}

func GeneratePageNamespacedLists(list []*shared.GenericNamespacedList, page int, pageSize int) (result []*shared.GenericNamespacedList) {
	sort.Slice(list, func(i, j int) bool {
		return strings.Compare(strings.ToLower(list[i].ID), strings.ToLower(list[j].ID)) < 0
	})

	length := len(list)
	startIndex := page*pageSize - pageSize
	if page < 1 || startIndex > length || pageSize < 1 {
		return result
	}
	lastIndex := length
	if page*pageSize < length {
		lastIndex = page * pageSize
	}
	return list[startIndex:lastIndex]
}

func IsValidType(stype string) bool {
	return STRING == stype || MAC_LIST == stype || IP_LIST == stype || RI_MAC_LIST == stype
}
func ValidateListDataForAdmin(typeName string, listData []string) error {
	if !IsValidType(typeName) {
		return errors.New("Type is invalid")
	}

	if len(listData) == 0 {
		return errors.New("List must not be empty")
	}

	var invalidAddresses []string
	if typeName == IP_LIST {
		for _, ipAddress := range listData {
			if shared.NewIpAddress(ipAddress) == nil {
				invalidAddresses = append(invalidAddresses, ipAddress)
			}
		}
	} else if typeName == MAC_LIST {
		for _, mac := range listData {
			if !util.IsValidMacAddress(mac) {
				invalidAddresses = append(invalidAddresses, mac)
			}
		}
	}
	if len(invalidAddresses) > 0 {
		return fmt.Errorf("List contains invalid address(es): %v", invalidAddresses)
	}

	return nil
}

func AddNamespacedListData(listType string, listId string, stringListWrapper *shared.StringListWrapper) *xwhttp.ResponseEntity {
	if listId == "" {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Id is empty"), nil)
	}

	err := ValidateListDataForAdmin(listType, stringListWrapper.List)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	shared.LockGenericNamespacedList()
	defer shared.UnlockGenericNamespacedList()

	listToUpdate, err := shared.GetGenericNamedListOneByTypeNonCached(listId, listType)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("List with current ID doesn't exist"), nil)
	}

	itemsSet := xutil.Set{}
	itemsSet.Add(listToUpdate.Data...)

	if listType == shared.MAC_LIST {
		for _, mac := range stringListWrapper.List {
			if macAddress, err := util.ValidateAndNormalizeMacAddress(mac); err != nil {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
			} else {
				itemsSet.Add(macAddress)
			}
		}
	} else {
		itemsSet.Add(stringListWrapper.List...)
	}

	listToUpdate.Data = itemsSet.ToSlice()

	err = listToUpdate.ValidateDataIntersection()
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	err = shared.CreateGenericNamedListOneDB(listToUpdate)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	if listType == shared.IP_LIST {
		listToUpdate.CreateIpAddressGroupResponse()
		return xwhttp.NewResponseEntity(http.StatusOK, nil, listToUpdate.CreateIpAddressGroupResponse())
	}
	return xwhttp.NewResponseEntity(http.StatusOK, nil, listToUpdate)
}

func RemoveNamespacedListData(listType string, listId string, stringListWrapper *shared.StringListWrapper) *xwhttp.ResponseEntity {
	if listId == "" {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("Id is empty"), nil)
	}

	err := ValidateListDataForAdmin(listType, stringListWrapper.List)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	shared.LockGenericNamespacedList()
	defer shared.UnlockGenericNamespacedList()

	listToUpdate, err := shared.GetGenericNamedListOneByTypeNonCached(listId, listType)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, errors.New("List with current ID doesn't exist"), nil)
	}

	itemsSet := xutil.Set{}
	itemsSet.Add(listToUpdate.Data...)
	itemsNotInList := make([]string, 0)

	if listType == shared.MAC_LIST {
		for _, mac := range stringListWrapper.List {
			if macAddress, err := util.ValidateAndNormalizeMacAddress(mac); err != nil {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
			} else {
				if itemsSet.Contains(macAddress) {
					itemsSet.Remove(macAddress)
				} else {
					itemsNotInList = append(itemsNotInList, macAddress)
				}
			}
		}
	} else {
		for _, str := range stringListWrapper.List {
			if itemsSet.Contains(str) {
				itemsSet.Remove(str)
			} else {
				itemsNotInList = append(itemsNotInList, str)
			}
		}
	}

	if len(itemsNotInList) > 0 {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("List contains %ss, which are not present in current Namespaced list: %s", getItemName(listType), itemsNotInList), nil)
	}

	listToUpdate.Data = itemsSet.ToSlice()
	if len(listToUpdate.Data) == 0 {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, fmt.Errorf("Namespaced list should contain at least one %s address", getItemName(listType)), nil)
	}

	err = shared.CreateGenericNamedListOneDB(listToUpdate)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, listToUpdate)
}

func CreateNamespacedList(namespacedList *shared.GenericNamespacedList, updateIfExists bool) *xwhttp.ResponseEntity {
	err := namespacedList.ValidateForAdminService()
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	if namespacedList.TypeName == shared.MAC_LIST {
		for i, mac := range namespacedList.Data {
			if macAddress, err := util.ValidateAndNormalizeMacAddress(mac); err != nil {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
			} else {
				namespacedList.Data[i] = macAddress
			}
		}
	}

	// No need to check for existing record if update is allowed
	if !updateIfExists {
		existingList, _ := shared.GetGenericNamedListOneByTypeNonCached(namespacedList.ID, namespacedList.TypeName)
		if existingList != nil {
			return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("List with name %s already exists", namespacedList.ID), nil)
		}
	}

	err = shared.CreateGenericNamedListOneDB(namespacedList)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusCreated, nil, namespacedList)
}

func UpdateNamespacedList(namespacedList *shared.GenericNamespacedList, newId string) *xwhttp.ResponseEntity {
	err := namespacedList.ValidateForAdminService()
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
	}

	if namespacedList.TypeName == shared.MAC_LIST {
		for i, mac := range namespacedList.Data {
			if macAddress, err := util.ValidateAndNormalizeMacAddress(mac); err != nil {
				return xwhttp.NewResponseEntity(http.StatusBadRequest, err, nil)
			} else {
				namespacedList.Data[i] = macAddress
			}
		}
	}

	// When new ID is provided, performs rename operation otherwise update
	if !xutil.IsBlank(newId) && newId != namespacedList.ID {
		existingList, _ := shared.GetGenericNamedListOneByTypeNonCached(newId, namespacedList.TypeName)
		if existingList != nil {
			return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("\"%s %s already exists\"", namespacedList.TypeName, newId), nil)
		}

		if err = renameNamespacedListInUsedEntities(namespacedList.ID, newId); err != nil {
			return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
		}

		if err = shared.DeleteOneGenericNamedList(namespacedList.ID); err != nil {
			return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
		}
		namespacedList.ID = newId
	} else {
		existingList, err := shared.GetGenericNamedListOneByTypeNonCached(namespacedList.ID, namespacedList.TypeName)
		if err != nil {
			return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
		}
		if existingList == nil {
			return xwhttp.NewResponseEntity(http.StatusConflict, fmt.Errorf("\"List with id %s doesn't exist\"", namespacedList.ID), nil)
		}
	}

	err = shared.CreateGenericNamedListOneDB(namespacedList)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, namespacedList)
}

func DeleteNamespacedList(typeName string, id string) *xwhttp.ResponseEntity {
	ds.GetCacheManager().ForceSyncChanges()
	var namespacedList *shared.GenericNamespacedList
	if typeName == "" {
		namespacedList = GetNamespacedListById(id)
	} else {
		namespacedList = GetNamespacedListByIdAndType(id, typeName)
	}
	if namespacedList == nil {
		return xwhttp.NewResponseEntity(http.StatusNotFound, fmt.Errorf("List with id: %s does not exist", id), nil)
	}
	namespacedList.Updated = 0

	usage, err := validateUsageForNamespacedList(id)
	if err != nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
	}

	if usage != "" {
		return xwhttp.NewResponseEntity(http.StatusConflict, errors.New(usage), nil)
	}

	if err := shared.DeleteOneGenericNamedList(id); err == nil {
		return xwhttp.NewResponseEntity(http.StatusNoContent, nil, nil)
	}
	return xwhttp.NewResponseEntity(http.StatusInternalServerError, err, nil)
}

// Return usage info if NamespacedList is used by a rule, empty string otherwise
func validateUsageForNamespacedList(id string) (string, error) {
	for _, tableName := range ruleTables {
		ruleList, err := ds.GetCachedSimpleDao().GetAllAsList(tableName, 0)
		if err != nil {
			return "", err
		}

		for _, v := range ruleList {
			xrule, ok := v.(ru.XRule)
			if !ok {
				return "", fmt.Errorf("Failed to assert %s as XRule type", tableName)
			}

			ids := ru.GetFixedArgsFromRuleByOperation(xrule.GetRule(), re.StandardOperationInList)
			if xutil.Contains(ids, id) {
				return fmt.Sprintf("List is used by %s %s", xrule.GetRuleType(), xrule.GetName()), nil
			}

			if tableName == ds.TABLE_FIRMWARE_RULE {
				firmwareRule, ok := v.(*firmware.FirmwareRule)
				if !ok {
					return "", fmt.Errorf("Failed to parse Firmware Rule")
				}
				if id == firmwareRule.ApplicableAction.Whitelist && firmwareRule.Type == corefw.ENV_MODEL_RULE {
					return fmt.Sprintf("%v is used in a Percentage Filter %v", id, firmwareRule.Name), nil
				}
			}
		}
	}

	for _, feature := range rfc.GetFeatureList() {
		if feature != nil && feature.Whitelisted && feature.WhitelistProperty != nil && feature.WhitelistProperty.Value == id {
			return fmt.Sprintf("NamespacedList is used by %s feature", feature.FeatureName), nil
		}
	}

	return "", nil
}

func renameNamespacedListInUsedEntities(oldNamespacedListId string, newNamespacedListId string) error {
	for _, tableName := range ruleTables {
		ruleList, err := ds.GetCachedSimpleDao().GetAllAsList(tableName, 0)
		if err != nil {
			return err
		}

		for _, v := range ruleList {
			if xrule, ok := v.(ru.XRule); ok {
				rule := xrule.GetRule()
				if ru.ChangeFixedArgToNewValue(oldNamespacedListId, newNamespacedListId, *rule, re.StandardOperationInList) {
					if err := ds.GetCachedSimpleDao().SetOne(tableName, xrule.GetId(), v); err != nil {
						return err
					}
				}
			} else {
				return fmt.Errorf("Failed to assert %s as XRule type", tableName)
			}
		}

		for _, feature := range rfc.GetFeatureListForAS() {
			if feature != nil && feature.Whitelisted && feature.WhitelistProperty != nil && feature.WhitelistProperty.Value == oldNamespacedListId {
				feature.WhitelistProperty.Value = newNamespacedListId
				if _, err := xrfc.SetOneFeature(feature); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func getItemName(listType string) string {
	s := strings.Split(listType, "_LIST")
	return s[0]
}

func isMacListHasMacPart(macPart string, macs []string) bool {
	normalizedMacPart := xutil.AlphaNumericMacAddress(macPart)
	for _, v := range macs {
		mac := strings.ReplaceAll(v, ":", "")
		if strings.Contains(mac, normalizedMacPart) {
			return true
		}
	}
	return false
}

func isIpAddressHasIpPart(ipPart string, ipAddresses []string) bool {
	for _, ip := range ipAddresses {
		if strings.Contains(ip, ipPart) {
			return true
		}

		ipAddress := shared.NewIpAddress(ip)
		if ipAddress != nil && ipAddress.IsInRange(ipPart) {
			return true
		}
	}
	return false
}

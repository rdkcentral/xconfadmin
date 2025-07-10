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
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"

	core "xconfadmin/shared"
	xwcommon "xconfwebconfig/common"

	log "github.com/sirupsen/logrus"
)

func findPrioritizableById(itemId string, prioritizables []core.Prioritizable) bool {
	for _, item := range prioritizables {
		if item.GetID() == itemId {
			return true
		}
	}
	return false
}

func ChangePrioritizablePriorities(prioritizable core.Prioritizable, newPriority int, applicationType string) ([]core.Prioritizable, error) {
	if newPriority <= 0 {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusBadRequest, fmt.Sprintf("Invalid priority value %v", newPriority))
	}
	oldPriority := prioritizable.GetPriority()

	contextMap := map[string]string{core.APPLICATION_TYPE: applicationType}
	prioritizables := FeatureRulesToPrioritizables(FindFeatureRuleByContext(contextMap))
	reorganizedPrioritizables := UpdatePrioritizablesPriorities(prioritizables, oldPriority, newPriority)
	if !findPrioritizableById(prioritizable.GetID(), reorganizedPrioritizables) {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusConflict, fmt.Sprintf("Updated prioritizable '%s' is not present in reorganized prioritizables", prioritizable.GetID()))
	}
	if err := SaveFeatureRules(reorganizedPrioritizables); err != nil {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, fmt.Sprintf("Failed to save prioritizable after priority reorganization: %s", err.Error()))
	}
	log.Info("Priority of Prioritizable " + prioritizable.GetID() + " has been changed, oldPriority=" + strconv.Itoa(oldPriority) + ", newPriority=" + strconv.Itoa(newPriority))
	return reorganizedPrioritizables, nil
}

func reorganizePrioritizablePriorities(sortedItemsList []core.Prioritizable, oldPriority int, newPriority int) []core.Prioritizable {
	if newPriority < 1 || int(newPriority) > len(sortedItemsList) {
		newPriority = len(sortedItemsList)
	}
	item := sortedItemsList[oldPriority-1]
	item.SetPriority(newPriority)
	if oldPriority < newPriority {
		for i := oldPriority; i <= newPriority-1; i++ {
			buf := sortedItemsList[i]
			buf.SetPriority(i)
			sortedItemsList[i-1] = buf
		}
	}
	if oldPriority > newPriority {
		for i := oldPriority - 2; i >= newPriority-1; i-- {
			buf := sortedItemsList[i]
			buf.SetPriority(i + 2)
			sortedItemsList[i+1] = buf
		}
	}
	sortedItemsList[newPriority-1] = item
	return getAlteredPrioritizableSubList(sortedItemsList, oldPriority, newPriority)
}

func getAlteredPrioritizableSubList(itemsList []core.Prioritizable, oldPriority int, newPriority int) []core.Prioritizable {
	start := int(math.Min(float64(oldPriority), float64(newPriority)) - float64(1))
	end := int(math.Max(float64(oldPriority), float64(newPriority)))
	return itemsList[start:end]
}

func AddNewPrioritizableAndReorganizePriorities(newItem core.Prioritizable, itemsList []core.Prioritizable) []core.Prioritizable {
	if itemsList == nil {
		return itemsList
	}
	sort.Slice(itemsList, func(i, j int) bool {
		return itemsList[i].GetPriority() < itemsList[j].GetPriority()
	})
	itemsList = append(itemsList, newItem)
	return reorganizePrioritizablePriorities(itemsList, len(itemsList), newItem.GetPriority())
}

func UpdatePrioritizablePriorityAndReorganize(newItem core.Prioritizable, itemsList []core.Prioritizable, priority int) []core.Prioritizable {
	sort.Slice(itemsList, func(i, j int) bool {
		return itemsList[i].GetPriority() < itemsList[j].GetPriority()
	})
	if len(itemsList) > 0 {
		for i, item := range itemsList {
			if item.GetID() == newItem.GetID() {
				itemsList[i] = newItem
				break
			}
		}
	} else {
		itemsList = append(itemsList, newItem)
	}
	return reorganizePrioritizablePriorities(itemsList, priority, newItem.GetPriority())
}

func UpdatePrioritizablesPriorities(itemsList []core.Prioritizable, oldPriority int, newPriority int) []core.Prioritizable {
	sort.Slice(itemsList, func(i, j int) bool {
		return itemsList[i].GetPriority() < itemsList[j].GetPriority()
	})
	return reorganizePrioritizablePriorities(itemsList, oldPriority, newPriority)
}

func PackPriorities(allItems []core.Prioritizable, itemToDelete core.Prioritizable) []core.Prioritizable {
	altered := []core.Prioritizable{}
	// sort by ascending priority
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].GetPriority() < allItems[j].GetPriority()
	})
	priority := 1
	for _, item := range allItems {
		if item.GetID() == itemToDelete.GetID() {
			continue
		}
		oldpriority := item.GetPriority()
		item.SetPriority(priority)
		priority++
		if item.GetPriority() != oldpriority {
			altered = append(altered, item)
		}
	}
	return altered
}

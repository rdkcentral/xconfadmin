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
package change

import (
	"fmt"
	"net/http"
	"sort"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"

	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	xcommon "github.com/rdkcentral/xconfadmin/common"
	xchange "github.com/rdkcentral/xconfadmin/shared/change"
	xutil "github.com/rdkcentral/xconfadmin/util"

	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	xwchange "github.com/rdkcentral/xconfwebconfig/shared/change"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	xwutil "github.com/rdkcentral/xconfwebconfig/util"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func GetTelemetryTwoChangeEntityIds() []string {
	ids := []string{}
	changeList := xchange.GetAllTelemetryTwoChangeList()
	for _, change := range changeList {
		ids = append(ids, change.EntityID)
	}
	return ids
}

func GetTelemetryTwoChangesByEntityId(entityId string) []*xwchange.TelemetryTwoChange {
	result := []*xwchange.TelemetryTwoChange{}
	changes := xchange.GetAllTelemetryTwoChangeList()
	for _, change := range changes {
		if change.EntityID == entityId {
			result = append(result, change)
		}
	}
	return result
}

func GetTelemetryTwoChangesByIds(changeIds []string) []*xwchange.TelemetryTwoChange {
	result := []*xwchange.TelemetryTwoChange{}
	for _, changeId := range changeIds {
		change := xchange.GetOneTelemetryTwoChange(changeId)
		if change != nil {
			result = append(result, change)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Updated < result[j].Updated
	})

	return result
}

func GetTelemetryTwoChangesByContext(searchContext map[string]string) []*xwchange.TelemetryTwoChange {
	filteredChanges := []*xwchange.TelemetryTwoChange{}
	changes := xchange.GetAllTelemetryTwoChangeList()
	for _, change := range changes {
		if applicationType, ok := xutil.FindEntryInContext(searchContext, xwcommon.APPLICATION_TYPE, false); ok {
			if change.ApplicationType != applicationType {
				continue
			}
		}
		if author, ok := xutil.FindEntryInContext(searchContext, xcommon.AUTHOR, false); ok {
			if !xutil.ContainsIgnoreCase(change.Author, author) {
				continue
			}
		}
		if name, ok := xutil.FindEntryInContext(searchContext, xcommon.ENTITY, false); ok {
			var profileName string
			if change.NewEntity != nil {
				profileName = change.NewEntity.Name
			} else if change.OldEntity != nil {
				profileName = change.OldEntity.Name
			}
			if !xutil.ContainsIgnoreCase(profileName, name) {
				continue
			}
		}
		filteredChanges = append(filteredChanges, change)
	}
	return filteredChanges
}

func GetApprovedTelemetryTwoChangesByContext(searchContext map[string]string) []*xwchange.ApprovedTelemetryTwoChange {
	filteredChanges := []*xwchange.ApprovedTelemetryTwoChange{}
	changes := xchange.GetAllApprovedTelemetryTwoChangeList()
	for _, change := range changes {
		if applicationType, ok := xutil.FindEntryInContext(searchContext, xwcommon.APPLICATION_TYPE, false); ok {
			if change.ApplicationType != applicationType {
				continue
			}
		}
		if author, ok := xutil.FindEntryInContext(searchContext, xcommon.AUTHOR, false); ok {
			if !xutil.ContainsIgnoreCase(change.Author, author) {
				continue
			}
		}
		if name, ok := xutil.FindEntryInContext(searchContext, xcommon.ENTITY, false); ok {
			var profileName string
			if change.NewEntity != nil {
				profileName = change.NewEntity.Name
			} else if change.OldEntity != nil {
				profileName = change.OldEntity.Name
			}
			if !xutil.ContainsIgnoreCase(profileName, name) {
				continue
			}
		}
		filteredChanges = append(filteredChanges, change)
	}
	return filteredChanges
}

func ApproveTelemetryTwoChange(r *http.Request, changeId string) (*xwchange.ApprovedTelemetryTwoChange, error) {
	change := xchange.GetOneTelemetryTwoChange(changeId)
	if change == nil {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusNotFound, fmt.Sprintf("Entity with id  %s does not exist", changeId))
	}

	if change.Operation == xchange.Create {
		if _, err := CreateTelemetryTwoProfile(r, change.NewEntity); err != nil {
			return nil, err
		}

		approvedChange, err := SaveToApprovedApprovedTelemetryTwoChange(r, change)
		if err != nil {
			return nil, err
		}

		if err := DeleteTelemetryTwoChange(changeId); err != nil {
			return nil, err
		}

		return approvedChange, nil
	}

	approvedChange, err := updateDeleteEntityTelemetryTwoChange(r, change)
	if err != nil {
		return nil, err
	}

	return approvedChange, nil
}

func ApproveTelemetryTwoChanges(r *http.Request, changeIds []string) map[string]string {
	errorMessages := make(map[string]string)
	mergedUpdateChangesByEntityId := make(map[string]*logupload.TelemetryTwoProfile)
	entityToByCancelChange := []string{}
	changesToApprove := GetTelemetryTwoChangesByIds(changeIds)
	for _, change := range changesToApprove {
		var err error
		switch {
		case xchange.Create == change.Operation:
			_, err = CreateTelemetryTwoProfile(r, change.NewEntity)
		case xchange.Update == change.Operation:
			var mergeResult *logupload.TelemetryTwoProfile
			mergeResult, err = applyUpdateTelemetryTwoChange(mergedUpdateChangesByEntityId[change.EntityID], change)
			if err == nil {
				mergedUpdateChangesByEntityId[mergeResult.ID] = mergeResult
				_, err = UpdateTelemetryTwoProfile(r, mergeResult)
			}
		case xchange.Delete == change.Operation:
			err = DeleteTelemetryTwoProfile(r, change.OldEntity.ID)
		}
		if err == nil {
			entityToByCancelChange = append(entityToByCancelChange, change.EntityID)
			saveToApprovedAndCleanUpTelemetryTwoChange(r, change)
		} else {
			errorMessages[change.ID] = err.Error()
		}
	}

	ids := make([]string, 0, len(errorMessages))
	for k := range errorMessages {
		ids = append(ids, k)
	}
	if err := cancelApprovedTelemetryTwoChangesByEntityId(r, entityToByCancelChange, ids); err != nil {
		log.Errorf("Failed to cancel approved changes (%v): %v", entityToByCancelChange, err)
	}
	if len(errorMessages) > 0 {
		log.Errorf("Approving Error: %v", errorMessages)
	}

	return errorMessages
}

func SaveToApprovedApprovedTelemetryTwoChange(r *http.Request, change *xwchange.TelemetryTwoChange) (*xwchange.ApprovedTelemetryTwoChange, error) {
	approvedChange := xchange.NewApprovedTelemetryTwoChange(change)

	if err := beforeSavingApprovedTelemetryTwoChange(r, approvedChange); err != nil {
		return nil, err
	}

	if err := xchange.SetOneApprovedTelemetryTwoChange(approvedChange); err != nil {
		return nil, err
	}

	return approvedChange, nil
}

func DeleteTelemetryTwoChange(changeId string) error {
	if err := beforeDeleteTelemetryTwoChange(changeId); err != nil {
		return err
	}
	if err := xchange.DeleteOneTelemetryTwoChange(changeId); err != nil {
		return xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func DeleteApprovedTelemetryTwoChange(changeId string) error {
	if err := beforeDeleteApprovedTelemetryTwoChange(changeId); err != nil {
		return err
	}
	if err := xchange.DeleteOneApprovedTelemetryTwoChange(changeId); err != nil {
		return err
	}
	return nil
}

func RevertTelemetryTwoChange(r *http.Request, approvedId string) *xwhttp.ResponseEntity {
	approvedChange := xchange.GetOneApprovedTelemetryTwoChange(approvedId)
	if approvedChange == nil {
		return xwhttp.NewResponseEntity(http.StatusNotFound, fmt.Errorf("ApprovedTelemetryTwoChange with %s id does not exist", approvedId), nil)
	}

	if approvedChange.Operation == xchange.Delete {
		revertDeleteApprovedTelemetryTwoChange(r, approvedChange)
	} else {
		revertCreateOrUpdateApprovedTelemetryTwoChange(r, approvedChange)
	}

	userName := auth.GetUserNameOrUnknown(r)
	log.Infof("Change has been reverted by %s: %v", userName, approvedChange)
	return xwhttp.NewResponseEntity(http.StatusOK, nil, nil)
}

func RevertTelemetryTwoChanges(r *http.Request, approvedIds []string) map[string]string {
	errorMessages := make(map[string]string)
	changesToRevert := make([]xwchange.ApprovedTelemetryTwoChange, 0, len(approvedIds))
	for _, approvedId := range approvedIds {
		approvedChange := xchange.GetOneApprovedTelemetryTwoChange(approvedId)
		if approvedChange != nil {
			changesToRevert = append(changesToRevert, *approvedChange)
		}
	}
	sort.Slice(changesToRevert, func(i, j int) bool {
		return changesToRevert[i].Updated < changesToRevert[j].Updated
	})

	for _, approvedChange := range changesToRevert {
		res := RevertTelemetryTwoChange(r, approvedChange.ID)
		if res.Status != http.StatusOK {
			errorMessages[approvedChange.ID] = res.Error.Error()
		}
	}

	if len(errorMessages) > 0 {
		log.Errorf("Reverting Error: %v", errorMessages)
	}

	return errorMessages
}

func GeneratePageTelemetryTwoChanges(list []*xwchange.TelemetryTwoChange, page int, pageSize int) (result []*xwchange.TelemetryTwoChange) {
	sort.Slice(list, func(i, j int) bool {
		return list[j].Updated < list[i].Updated
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

func GeneratePageApprovedTelemetryTwoChanges(list []*xwchange.ApprovedTelemetryTwoChange, page int, pageSize int) (result []*xwchange.ApprovedTelemetryTwoChange) {
	sort.Slice(list, func(i, j int) bool {
		return list[j].Updated < list[i].Updated
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

func GroupTelemetryTwoChanges(changes []*xwchange.TelemetryTwoChange) map[string][]xwchange.TelemetryTwoChange {
	groupedChanges := make(map[string][]xwchange.TelemetryTwoChange)
	for _, change := range changes {
		if _, found := groupedChanges[change.EntityID]; !found {
			groupedChanges[change.EntityID] = make([]xwchange.TelemetryTwoChange, 0, 1)
		}
		groupedChanges[change.EntityID] = append(groupedChanges[change.EntityID], *change)
	}

	return groupedChanges
}

func GroupApprovedTelemetryTwoChanges(changes []*xwchange.ApprovedTelemetryTwoChange) map[string][]xwchange.ApprovedTelemetryTwoChange {
	groupedChanges := make(map[string][]xwchange.ApprovedTelemetryTwoChange)
	for _, change := range changes {
		if _, found := groupedChanges[change.EntityID]; !found {
			groupedChanges[change.EntityID] = make([]xwchange.ApprovedTelemetryTwoChange, 0, 1)
		}
		groupedChanges[change.EntityID] = append(groupedChanges[change.EntityID], *change)
	}

	return groupedChanges
}

func beforeSavingTelemetryTwoChange(r *http.Request, change *xwchange.TelemetryTwoChange) error {
	if change.ID == "" {
		change.ID = uuid.New().String()
	}
	if change.ApplicationType == "" {
		/* Todo: wait for permissionService.getWriteApplication() to be ready in Golang
		if (change != null && StringUtils.isBlank(change.getApplicationType())) {
			change.setApplicationType(permissionService.getWriteApplication());
		} */
		application, err := auth.CanWrite(r, auth.CHANGE_ENTITY)
		if err != nil {
			return err
		}
		change.ApplicationType = application
	}
	if err := change.Validate(); err != nil {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, err.Error())
	}
	return nil
}

func beforeSavingApprovedTelemetryTwoChange(r *http.Request, approvedChange *xwchange.ApprovedTelemetryTwoChange) error {
	if approvedChange.ApprovedUser == "" {
		approvedChange.ApprovedUser = auth.GetUserNameOrUnknown(r)
	}
	if approvedChange.ID == "" {
		approvedChange.ID = uuid.New().String()
	}
	if approvedChange.ApplicationType == "" {
		application, err := auth.CanWrite(r, auth.CHANGE_ENTITY)
		if err != nil {
			return err
		}
		approvedChange.ApplicationType = application
	}
	if err := approvedChange.Validate(); err != nil {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, err.Error())
	}
	return nil
}

func beforeDeleteTelemetryTwoChange(id string) error {
	if id == "" {
		xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Id is blank")
	}
	if change := xchange.GetOneTelemetryTwoChange(id); change == nil {
		xwcommon.NewRemoteErrorAS(http.StatusNotFound, fmt.Sprintf("TelemetryTwoChange with %s id does not exist", id))
	}
	return nil
}

func beforeDeleteApprovedTelemetryTwoChange(id string) error {
	if id == "" {
		xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Id is blank")
	}
	if change := xchange.GetOneApprovedTelemetryTwoChange(id); change == nil {
		xwcommon.NewRemoteErrorAS(http.StatusNotFound, fmt.Sprintf("ApprovedTelemetryTwoChange with %s id does not exist", id))
	}
	return nil
}

func revertDeleteApprovedTelemetryTwoChange(r *http.Request, approvedChange *xwchange.ApprovedTelemetryTwoChange) error {
	if approvedChange.OldEntity == nil {
		return xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, fmt.Sprintf("OldEntity is empty for ApprovedTelemetryTwoChange with id %s", approvedChange.ID))
	}
	if _, err := CreateTelemetryTwoProfile(r, approvedChange.OldEntity); err != nil {
		return err
	}

	if err := DeleteApprovedTelemetryTwoChange(approvedChange.ID); err != nil {
		return err
	}
	return nil
}

func revertCreateOrUpdateApprovedTelemetryTwoChange(r *http.Request, approvedChange *xwchange.ApprovedTelemetryTwoChange) error {
	entityToRevert := logupload.GetOneTelemetryTwoProfile(approvedChange.EntityID)
	if entityToRevert == nil {
		return xwcommon.NewRemoteErrorAS(http.StatusNotFound, fmt.Sprintf("TelemetryTwoProfile with id %s does not exist", approvedChange.EntityID))
	}

	if approvedChange.Operation == xchange.Create {
		if err := DeleteTelemetryTwoProfile(r, entityToRevert.ID); err != nil {
			return err
		}
	} else {
		if _, err := UpdateTelemetryTwoProfile(r, entityToRevert); err != nil {
			return err
		}
	}

	return DeleteApprovedTelemetryTwoChange(approvedChange.ID)
}

func buildToCreateTelemetryTwoChange(newEntity *logupload.TelemetryTwoProfile, applicationType string, userName string) *xwchange.TelemetryTwoChange {
	change := xchange.NewEmptyTelemetryTwoChange()
	change.ID = uuid.New().String()
	change.EntityID = newEntity.ID
	change.EntityType = xchange.TelemetryTwoProfile
	change.ApplicationType = applicationType
	change.NewEntity = newEntity
	change.Author = userName
	change.Operation = xchange.Create
	return change
}

func buildToUpdateTelemetryTwoChange(oldEntity *logupload.TelemetryTwoProfile, newEntity *logupload.TelemetryTwoProfile, applicationType string, userName string) *xwchange.TelemetryTwoChange {
	change := xchange.NewEmptyTelemetryTwoChange()
	change.ID = uuid.New().String()
	change.EntityID = oldEntity.ID
	change.EntityType = xchange.TelemetryTwoProfile
	change.ApplicationType = applicationType
	change.OldEntity = oldEntity
	change.NewEntity = newEntity
	change.Author = userName
	change.Operation = xchange.Update
	return change
}

func buildToDeleteTelemetryTwoChange(oldEntity *logupload.TelemetryTwoProfile, applicationType string, userName string) *xwchange.TelemetryTwoChange {
	change := xchange.NewEmptyTelemetryTwoChange()
	change.ID = uuid.New().String()
	change.EntityID = oldEntity.ID
	change.EntityType = xchange.TelemetryTwoProfile
	change.ApplicationType = applicationType
	change.OldEntity = oldEntity
	change.Operation = xchange.Delete
	change.Author = userName
	return change
}

func updateDeleteEntityTelemetryTwoChange(r *http.Request, change *xwchange.TelemetryTwoChange) (*xwchange.ApprovedTelemetryTwoChange, error) {
	currentEntity := change.OldEntity
	entityToChange := logupload.GetOneTelemetryTwoProfile(change.EntityID)
	// in Java, equalPendingEntities(currentEntity, entityToChange) always return true
	//if (entityToChange != null && equalPendingEntities(currentEntity, entityToChange)) {
	if entityToChange != nil {
		if change.Operation == xchange.Delete {
			if err := DeleteTelemetryTwoProfile(r, currentEntity.ID); err != nil {
				return nil, err
			}
		} else {
			if _, err := UpdateTelemetryTwoProfile(r, change.NewEntity); err != nil {
				return nil, err
			}
		}
		change.ApprovedUser = auth.GetUserNameOrUnknown(r)
		approvedChange, err := SaveToApprovedApprovedTelemetryTwoChange(r, change)
		if err != nil {
			return nil, err
		}
		if err := xchange.DeleteOneTelemetryTwoChange(change.ID); err != nil {
			return nil, err
		}
		return approvedChange, nil
	} else {
		msg := fmt.Sprintf("Change could not be approved, TelemetryTwoProfile have been already changed: TelemetryTwoChange %s - EntityID %s", change.ID, change.EntityID)
		return nil, xwcommon.NewRemoteErrorAS(http.StatusConflict, msg)
	}
}

func applyUpdateTelemetryTwoChange(mergeResult *logupload.TelemetryTwoProfile, change *xwchange.TelemetryTwoChange) (*logupload.TelemetryTwoProfile, error) {
	if mergeResult == nil {
		var err error
		if mergeResult, err = change.NewEntity.Clone(); err == nil {
			return mergeResult, nil
		} else {
			return nil, err
		}
	}
	oldProfile := change.OldEntity
	updatedProfile := change.NewEntity
	if oldProfile.Name != updatedProfile.Name {
		mergeResult.Name = updatedProfile.Name
	}
	if oldProfile.Jsonconfig != updatedProfile.Jsonconfig {
		mergeResult.Jsonconfig = updatedProfile.Jsonconfig
	}
	return mergeResult, nil
}

func saveToApprovedAndCleanUpTelemetryTwoChange(r *http.Request, change *xwchange.TelemetryTwoChange) error {
	approvedChange, err := SaveToApprovedApprovedTelemetryTwoChange(r, change)
	if err != nil {
		return err
	}

	if err := DeleteTelemetryTwoChange(change.ID); err != nil {
		return err
	}
	userName := auth.GetUserNameOrUnknown(r)
	log.Infof("Change approved by %s: %v", userName, approvedChange)
	return nil
}

func cancelApprovedTelemetryTwoChangesByEntityId(r *http.Request, entityIdsToByCancelChanges []string, changeIdsToBeExcluded []string) error {
	for _, entityId := range entityIdsToByCancelChanges {
		changes := GetTelemetryTwoChangesByEntityId(entityId)
		for _, change := range changes {
			if !xwutil.Contains(changeIdsToBeExcluded, change.ID) {
				if err := DeleteTelemetryTwoChange(change.ID); err != nil {
					return err
				}
				userName := auth.GetUserNameOrUnknown(r)
				log.Infof("Automatically canceled change by %s: %v", userName, change)
			}
		}
	}
	return nil
}

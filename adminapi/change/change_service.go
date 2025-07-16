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
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	xcommon "github.com/rdkcentral/xconfadmin/common"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"

	"github.com/rdkcentral/xconfwebconfig/shared"

	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	xshared "github.com/rdkcentral/xconfadmin/shared"
	xchange "github.com/rdkcentral/xconfadmin/shared/change"
	xutil "github.com/rdkcentral/xconfadmin/util"
	xwshared "github.com/rdkcentral/xconfwebconfig/shared"
	xwchange "github.com/rdkcentral/xconfwebconfig/shared/change"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func GetApprovedAll(r *http.Request) ([]*xwchange.ApprovedChange, error) {
	approvedChangesAll := xchange.GetApprovedChangeList()
	approvedChanges := []*xwchange.ApprovedChange{}
	application, err := auth.CanRead(r, auth.CHANGE_ENTITY)
	if err != nil {
		return nil, err
	}
	for _, approvedChange := range approvedChangesAll {
		if xshared.ApplicationTypeEquals(application, approvedChange.ApplicationType) || xshared.ApplicationTypeEquals(application, xwshared.ALL) {
			approvedChanges = append(approvedChanges, approvedChange)
		}
	}
	sort.Slice(approvedChanges, func(i, j int) bool {
		return approvedChanges[i].Updated < approvedChanges[j].Updated
	})
	return approvedChanges, nil
}

func Delete(changeId string) (*xwchange.Change, error) {
	err := beforeDelete(changeId)
	if err != nil {
		return nil, err
	}
	change := xchange.GetOneChange(changeId)
	xchange.DeleteOneChange(changeId)
	return change, nil
}

func beforeSavingChange(r *http.Request, change *xwchange.Change) error {
	if change != nil && change.ApplicationType == "" {
		application, err := auth.CanWrite(r, auth.CHANGE_ENTITY)
		if err != nil {
			return err
		}
		change.ApplicationType = application
	}
	if change.ID == "" {
		change.ID = uuid.New().String()
	}
	err := validateChange(*change)
	if err != nil {
		return err
	}

	return validateAllChanges(change)
}

func beforeSavingApprovedChange(r *http.Request, change *xwchange.Change) error {
	if change != nil && change.ApplicationType == "" {
		application, err := auth.CanWrite(r, auth.CHANGE_ENTITY)
		if err != nil {
			return err
		}
		change.ApplicationType = application
	}

	return validateApprovedChange(*change)
}

func validateChange(pendingChange xwchange.PendingChange) error {
	if pendingChange == nil {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Change is empty")
	}
	if pendingChange.GetID() == "" {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Id is blank")
	}
	if pendingChange.GetAuthor() == "" {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Author is empty")
	}
	if pendingChange.GetEntityID() == "" {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Entity id is empty")
	}
	if pendingChange.GetOperation() == "" {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Operation is empty")
	}
	if (xwchange.Create == pendingChange.GetOperation() || xwchange.Update == pendingChange.GetOperation()) && pendingChange.GetNewEntity().IsEmpty() {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "New entity is empty")
	}
	if (xwchange.Delete == pendingChange.GetOperation() || xwchange.Update == pendingChange.GetOperation()) && pendingChange.GetOldEntity().IsEmpty() {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Old entity is empty")
	}

	return nil
}

func validateApprovedChange(change xwchange.PendingChange) error {
	validateChange(change)
	if change.GetApprovedUser() == "" {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Approved user is empty")
	}
	return nil
}

func validateAllChanges(change *xwchange.Change) error {
	changesById := GetChangesByEntityId(change.EntityID)
	for _, existingChange := range changesById {
		if existingChange.EqualChangeData(change) {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, "The same change already exists")
		}
	}
	return nil
}

func beforeDelete(id string) error {
	if id == "" {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Id is blank")
	}
	change := xchange.GetOneChange(id)
	if change == nil {
		return xwcommon.NewRemoteErrorAS(http.StatusNotFound, " Change with "+id+" id does not exist")
	}
	return nil
}

func CreateApprovedChange(r *http.Request, change *xwchange.Change) (*xwchange.ApprovedChange, error) {
	err := beforeSavingApprovedChange(r, change)
	if err != nil {
		return nil, err
	}
	approvedChange := xwchange.ApprovedChange(*change)
	xchange.SetOneApprovedChange(&approvedChange)
	jsonBytes, _ := json.Marshal(change)
	log.Info("ApprovedChange saved: {}", string(jsonBytes))
	return &approvedChange, nil
}

// TODO remove it
func updateDeleteEntity(r *http.Request, change *xwchange.Change) (*xwchange.ApprovedChange, error) {
	entityToChange := logupload.GetOnePermanentTelemetryProfile(change.EntityID) //*PermanentTelemetryProfile
	if entityToChange != nil {                                                   // in Java, equalPendingEntities is hard-code to return true
		change.ApprovedUser = auth.GetUserNameOrUnknown(r)
		if xwchange.Delete == change.Operation {
			_, err := Delete(change.OldEntity.ID)
			if err != nil {
				return nil, err
			}
		} else {
			newEntity := change.NewEntity
			_, err := UpdatePermanentTelemetryProfile(newEntity)
			if err != nil {
				return nil, err
			}
		}
		approvedChange, err := CreateApprovedChange(r, change)
		if err != nil {
			return nil, err
		}
		_, err = Delete(change.ID)
		if err != nil {
			return nil, err
		}
		return approvedChange, nil
	} else {
		jsonBytes, _ := json.Marshal(entityToChange)
		return nil, xwcommon.NewRemoteErrorAS(http.StatusConflict, "Change could not be approved, "+change.OldEntity.Name+" have been already changed: "+string(jsonBytes))
	}
}

func Revert(r *http.Request, approvedId string) error {
	if approvedId == "" {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Id is blank")
	}
	approvedChange := xchange.GetOneApprovedChange(approvedId)
	if approvedChange == nil {
		return xwcommon.NewRemoteErrorAS(http.StatusNotFound, "ApprovedChange with "+approvedId+" id does not exist")
	}
	if xwchange.Delete == approvedChange.Operation {
		revertDelete(r, approvedId, approvedChange)
	} else {
		revertCreateOrUpdateChange(r, approvedId, approvedChange.EntityID, approvedChange)
	}
	userName := auth.GetUserNameOrUnknown(r)
	log.Info("Change has been reverted by {}: {}", userName, approvedChange)
	return nil
}

func revertDelete(r *http.Request, id string, approvedChange *xwchange.ApprovedChange) *xwchange.ApprovedChange {
	CreatePermanentTelemetryProfile(r, approvedChange.OldEntity)
	xchange.DeleteOneApprovedChange(id)
	return approvedChange
}

func revertCreateOrUpdateChange(r *http.Request, changeId string, entityId string, approvedChange *xwchange.ApprovedChange) *xwchange.ApprovedChange {
	entityToRevert := logupload.GetOnePermanentTelemetryProfile(entityId)
	// in Java, equalPendingEntities(PermanentTelemetryProfile oldEntity, PermanentTelemetryProfile newEntity) always returns true
	//if (equalPendingEntities(approvedChange.getNewEntity(), entityToRevert)) { is being ignored
	if xwchange.Create == approvedChange.Operation {
		DeletePermanentTelemetryProfile(r, entityToRevert.ID)
	} else {
		UpdatePermanentTelemetryProfile(approvedChange.OldEntity)
	}
	xchange.DeleteOneApprovedChange(changeId)
	return approvedChange
}

func CancelChange(r *http.Request, changeId string) error {
	canceledChange, err := Delete(changeId)
	if err != nil {
		return err
	}
	userName := auth.GetUserNameOrUnknown(r)
	log.Info("Change has been canceled by {}: {}", userName, canceledChange)
	return nil
}

func GroupChanges(changes []*xwchange.Change) map[string][]*xwchange.Change {
	groupedChanges := make(map[string][]*xwchange.Change)
	for _, change := range changes {
		groupChange(change, groupedChanges)
	}
	return groupedChanges
}

func groupChange(change *xwchange.Change, groupedChanges map[string][]*xwchange.Change) {
	if _, ok := groupedChanges[change.EntityID]; ok && groupedChanges[change.EntityID] != nil {
		groupedChanges[change.EntityID] = append(groupedChanges[change.EntityID], change)
	} else {
		var changeList []*xwchange.Change
		groupedChanges[change.EntityID] = append(changeList, change)
	}
}

func GroupApprovedChanges(changes []*xwchange.ApprovedChange) map[string][]*xwchange.ApprovedChange {
	groupedChanges := make(map[string][]*xwchange.ApprovedChange)
	for _, change := range changes {
		groupApprovedChange(change, groupedChanges)
	}
	return groupedChanges
}

func groupApprovedChange(change *xwchange.ApprovedChange, groupedChanges map[string][]*xwchange.ApprovedChange) {
	if _, ok := groupedChanges[change.EntityID]; ok && groupedChanges[change.EntityID] != nil {
		groupedChanges[change.EntityID] = append(groupedChanges[change.EntityID], change)
	} else {
		var changeList []*xwchange.ApprovedChange
		groupedChanges[change.EntityID] = append(changeList, change)
	}
}

func GetChangedEntityIds() *[]string {
	ids := []string{}
	changeList := xchange.GetChangeList()
	for _, change := range changeList {
		ids = append(ids, change.EntityID)
	}
	return &ids
}

func GetChangesByEntityIds(changeIds *[]string) ([]*xwchange.Change, error) {
	changes := []*xwchange.Change{}
	for _, id := range *changeIds {
		if id == "" {
			return nil, xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Id is blank")
		}
		change := xchange.GetOneChange(id)
		if change == nil {
			return nil, xwcommon.NewRemoteErrorAS(http.StatusNotFound, "Factory with "+id+" id does not exist")
		}
		changes = append(changes, change)
	}
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Updated < changes[j].Updated
	})
	return changes, nil
}

func GetChangesByEntityId(entityId string) []*xwchange.Change {
	result := []*xwchange.Change{}
	changes := xchange.GetChangeList()
	for _, change := range changes {
		if change.EntityID == entityId {
			result = append(result, change)
		}
	}
	return result
}

func Approve(r *http.Request, id string) (*xwchange.ApprovedChange, error) {
	change := xchange.GetOneChange(id)
	if change == nil {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusNotFound, "Change with "+id+" id does not exist")
	}

	var err error
	var approvedChange *xwchange.ApprovedChange
	switch {
	case xwchange.Create == change.Operation:
		_, err = CreatePermanentTelemetryProfile(r, change.NewEntity)
	case xwchange.Update == change.Operation:
		_, err = UpdatePermanentTelemetryProfile(change.NewEntity)
	case xwchange.Delete == change.Operation:
		_, err = DeletePermanentTelemetryProfile(r, change.OldEntity.ID)
	}
	if err != nil {
		return nil, err
	} else {
		approvedChange, err = SaveToApprovedAndCleanUpChange(r, change)
		if err != nil {
			return nil, err
		}
	}

	changesByProfileId := GetChangesByEntityId(change.EntityID)
	err = CancelApprovedChangesByEntityId(r, getChangeIds(changesByProfileId), []string{})
	if err != nil {
		return nil, err
	}

	return approvedChange, nil
}

func getChangeIds(changes []*xwchange.Change) []string {
	changeIds := []string{}
	for _, change := range changes {
		changeIds = append(changeIds, change.EntityID)
	}
	return changeIds
}

func ApproveChanges(r *http.Request, changeIds *[]string) (map[string]string, error) {
	changesToApprove, err := GetChangesByEntityIds(changeIds)
	if err != nil {
		return nil, err
	}
	errorMessages := make(map[string]string)
	mergedUpdateChangesByEntityId := make(map[string]*logupload.PermanentTelemetryProfile)
	entityToByCancelChange := []string{}
	for _, change := range changesToApprove {
		var err error
		switch {
		case xwchange.Create == change.Operation:
			_, err = CreatePermanentTelemetryProfile(r, change.NewEntity)
		case xwchange.Update == change.Operation:
			mergeResult := ApplyUpdateChange(mergedUpdateChangesByEntityId[change.EntityID], change)
			mergedUpdateChangesByEntityId[mergeResult.ID] = mergeResult
			_, err = UpdatePermanentTelemetryProfile(mergeResult)
		case xwchange.Delete == change.Operation:
			_, err = DeletePermanentTelemetryProfile(r, change.OldEntity.ID)
		}
		if err != nil {
			logAndCollectChangeException(change, err, errorMessages)
		} else {
			_, err = SaveToApprovedAndCleanUpChange(r, change)
			if err != nil {
				logAndCollectChangeException(change, err, errorMessages)
			} else {
				entityToByCancelChange = append(entityToByCancelChange, change.EntityID)
			}
		}
	}
	keys := make([]string, 0, len(errorMessages))
	for k := range errorMessages {
		keys = append(keys, k)
	}
	CancelApprovedChangesByEntityId(r, entityToByCancelChange, keys)
	return errorMessages, nil
}

func SaveToApprovedAndCleanUpChange(r *http.Request, change *xwchange.Change) (*xwchange.ApprovedChange, error) {
	userName := auth.GetUserNameOrUnknown(r)
	change.ApprovedUser = userName
	approvedChange, err := CreateApprovedChange(r, change)
	if err != nil {
		return approvedChange, err
	}
	Delete(change.ID)
	log.Info("Change approved by {}: {}", userName, approvedChange)
	return approvedChange, nil
}

func CancelApprovedChangesByEntityId(r *http.Request, entityIdsToByCancelChanges []string, changeIdsToBeExcluded []string) error {
	for _, entityId := range entityIdsToByCancelChanges {
		changes := GetChangesByEntityId(entityId)
		for _, changeByEntityId := range changes {
			if !xutil.StringSliceContains(changeIdsToBeExcluded, changeByEntityId.ID) {
				_, err := Delete(changeByEntityId.ID)
				if err != nil {
					return err
				}
				userName := auth.GetUserNameOrUnknown(r)
				log.Info("Automatically canceled change by {}: {}", userName, changeByEntityId)
			}
		}
	}
	return nil
}

func logAndCollectChangeException(change *xwchange.Change, err error, errorMessages map[string]string) {
	errMsg := fmt.Sprintf("ApprovingException:  %v", err)
	log.Error(errMsg)
	errorMessages[change.ID] = errMsg
}

func RevertChanges(r *http.Request, changeIds *[]string) (map[string]string, error) {
	changesToRevert := []*xwchange.ApprovedChange{}
	for _, changeId := range *changeIds {
		approvedChange := xchange.GetOneApprovedChange(changeId)
		if approvedChange == nil {
			return nil, xwcommon.NewRemoteErrorAS(http.StatusNotFound, "ApprovedChange with "+changeId+" id does not exist")
		}
		changesToRevert = append(changesToRevert, approvedChange)
	}
	sort.Slice(changesToRevert, func(i, j int) bool {
		return changesToRevert[i].Updated < changesToRevert[j].Updated
	})
	errorMessages := make(map[string]string)
	for _, approvedChange := range changesToRevert {
		err := Revert(r, approvedChange.ID)
		if err != nil {
			log.Error("RevertingException: ", err.Error())
			errorMessages[approvedChange.ID] = err.Error()
		}
	}
	return errorMessages, nil
}

func FindByContextForChanges(searchContext map[string]string) []*xwchange.Change {
	changes := xchange.GetChangeList()
	changesFound := []*xwchange.Change{}
	for _, change := range changes {
		if applicationType, ok := xutil.FindEntryInContext(searchContext, xwcommon.APPLICATION_TYPE, false); ok {
			if applicationType != "" && applicationType != shared.ALL {
				if change.ApplicationType != applicationType {
					continue
				}
			}
		}
		if author, ok := xutil.FindEntryInContext(searchContext, xcommon.AUTHOR, false); ok {
			if author != "" {
				if !strings.Contains(change.Author, author) {
					continue
				}
			}
		}
		if profileName, ok := xutil.FindEntryInContext(searchContext, xcommon.ENTITY, false); ok {
			if profileName != "" {
				entity := change.NewEntity
				if entity == nil {
					entity = change.OldEntity
					if entity == nil {
						continue
					}
				}
				if !strings.Contains(entity.Name, profileName) {
					continue
				}
			}
		}
		changesFound = append(changesFound, change)
	}
	return changesFound
}

func FindByContextForApprovedChanges(r *http.Request, searchContext map[string]string) []*xwchange.ApprovedChange {
	approvedChanges := xchange.GetApprovedChangeList()
	changesFound := []*xwchange.ApprovedChange{}
	for _, change := range approvedChanges {
		if applicationType, ok := xutil.FindEntryInContext(searchContext, xwcommon.APPLICATION_TYPE, false); ok {
			if applicationType != "" && applicationType != xwshared.ALL {
				if change.ApplicationType != applicationType {
					continue
				}
			}
		}
		if author, ok := xutil.FindEntryInContext(searchContext, xcommon.AUTHOR, false); ok {
			if author != "" {
				if !strings.Contains(change.Author, author) {
					continue
				}
			}
		}
		if profileName, ok := xutil.FindEntryInContext(searchContext, xcommon.PROFILE_NAME, false); ok {
			if profileName != "" {
				entity := change.NewEntity
				if entity == nil {
					entity = change.OldEntity
					if entity == nil {
						continue
					}
				}
				if !strings.Contains(entity.Name, profileName) {
					continue
				}
			}
		}
		changesFound = append(changesFound, change)
	}
	return changesFound
}

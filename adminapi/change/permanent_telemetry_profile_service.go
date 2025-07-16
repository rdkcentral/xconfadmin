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

	xcommon "xconfadmin/common"
	xshared "xconfadmin/shared"

	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

	"xconfadmin/adminapi/auth"
	xchange "xconfadmin/shared/change"
	xlogupload "xconfadmin/shared/logupload"
	xutil "xconfadmin/util"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	core_change "github.com/rdkcentral/xconfwebconfig/shared/change"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func GetTelemetryProfilesByContext(searchContext map[string]string) []*logupload.PermanentTelemetryProfile {
	filteredProfiles := []*logupload.PermanentTelemetryProfile{}
	profiles := logupload.GetPermanentTelemetryProfileList()
	for _, profile := range profiles {
		if applicationType, ok := xutil.FindEntryInContext(searchContext, xwcommon.APPLICATION_TYPE, false); ok {
			if profile.ApplicationType != applicationType {
				continue
			}
		}
		if name, ok := xutil.FindEntryInContext(searchContext, xcommon.NAME_UPPER, false); ok {
			if !xutil.ContainsIgnoreCase(profile.Name, name) {
				continue
			}
		}
		filteredProfiles = append(filteredProfiles, profile)
	}
	return filteredProfiles
}

func UpdatePermanentTelemetryProfile(updatedProfile *logupload.PermanentTelemetryProfile) (*logupload.PermanentTelemetryProfile, error) {
	normalizeOnSaveAfterApproving(updatedProfile)
	err := beforeUpdating(updatedProfile)
	if err != nil {
		return nil, err
	}
	if err := beforeSavingPermanentTelemetryProfile(updatedProfile); err != nil {
		return nil, err
	}
	err = xlogupload.SetOnePermanentTelemetryProfile(updatedProfile.ID, updatedProfile)
	if err != nil {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, err.Error())
	}
	return updatedProfile, nil
}

func CreatePermanentTelemetryProfile(r *http.Request, profile *logupload.PermanentTelemetryProfile) (*logupload.PermanentTelemetryProfile, error) {
	normalizeOnSaveAfterApproving(profile)
	err := beforeCreating(profile)
	if err != nil {
		return nil, err
	}
	return SavePermanentTelemetryProfile(r, profile)
}

func SavePermanentTelemetryProfile(r *http.Request, entity *logupload.PermanentTelemetryProfile) (*logupload.PermanentTelemetryProfile, error) {
	if err := auth.ValidateWrite(r, entity.ApplicationType, auth.TELEMETRY_ENTITY); err != nil {
		return nil, err
	}
	if err := beforeSavingPermanentTelemetryProfile(entity); err != nil {
		return nil, err
	}
	if err := xlogupload.SetOnePermanentTelemetryProfile(entity.ID, entity); err != nil {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, err.Error())
	}
	return entity, nil
}

func ValidateTelemetryProfilePendingChanges(entity *logupload.PermanentTelemetryProfile) error {
	pendingEntities := xchange.GetChangeList()
	for _, change := range pendingEntities {
		if change.ID != entity.ID {
			if &change.NewEntity != nil && entity.EqualChangeData(change.NewEntity) {
				return xwcommon.NewRemoteErrorAS(http.StatusConflict, "The same change already exists")
			} else if &change.OldEntity != nil && entity.EqualChangeData(change.OldEntity) {
				return xwcommon.NewRemoteErrorAS(http.StatusConflict, "The same change already exists")
			}
		}
	}
	return nil
}

func beforeSavingPermanentTelemetryProfile(entity *logupload.PermanentTelemetryProfile) error {
	if err := entity.Validate(); err != nil {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, err.Error())
	}
	if err := validateAll(entity); err != nil {
		return err
	}
	return nil
}

func normalizeOnSaveAfterApproving(profile *logupload.PermanentTelemetryProfile) {
	if profile != nil && len(profile.TelemetryProfile) < 1 {
		return
	}
	for i, telemetryElement := range profile.TelemetryProfile {
		if telemetryElement.ID == "" {
			profile.TelemetryProfile[i].ID = uuid.New().String()
		}
	}
}

func beforeCreating(entity *logupload.PermanentTelemetryProfile) error {
	id := entity.ID
	if id == "" {
		entity.ID = uuid.New().String()
	} else {
		existingEntity := logupload.GetOnePermanentTelemetryProfile(id)
		if existingEntity != nil {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, "Entity with id: "+id+" already exists")
		}
	}
	return nil
}

func beforeUpdating(updatedProfile *logupload.PermanentTelemetryProfile) error {
	id := updatedProfile.ID
	if id == "" {
		return xwcommon.NewRemoteErrorAS(http.StatusNotFound, "Entity id is empty")
	}
	existingEntity := logupload.GetOnePermanentTelemetryProfile(id)
	if existingEntity == nil {
		return xwcommon.NewRemoteErrorAS(http.StatusNotFound, "Entity with id: "+id+" does not exist")
	}
	return nil
}

func validateAll(entity *logupload.PermanentTelemetryProfile) error {
	existingEntities := logupload.GetPermanentTelemetryProfileList() //[]*PermanentTelemetryProfile
	for _, profile := range existingEntities {
		if profile.ID != entity.ID && profile.Name == entity.Name {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, "PermanentProfile with such name exists: "+entity.Name)
		}
	}
	return nil
}

func DeletePermanentTelemetryProfile(r *http.Request, id string) (*logupload.PermanentTelemetryProfile, error) {
	writeApplication, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		return nil, err
	}
	profile, err := beforeRemoving(id, writeApplication)
	if err != nil {
		return nil, err
	}
	xlogupload.DeletePermanentTelemetryProfile(id)
	return profile, nil
}

func beforeRemoving(id string, writeApplication string) (*logupload.PermanentTelemetryProfile, error) {
	entity := logupload.GetOnePermanentTelemetryProfile(id)
	if entity == nil || !xshared.ApplicationTypeEquals(writeApplication, entity.ApplicationType) {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusNotFound, "Entity with id: "+id+" does not exist")
	}
	if err := validateUsage(id); err != nil {
		return nil, err
	}
	return entity, nil
}

func buildToCreateChange(newEntity *logupload.PermanentTelemetryProfile, applicationType string, userName string) *core_change.Change {
	change := xchange.NewEmptyChange()
	change.ID = uuid.New().String()
	change.EntityID = newEntity.ID
	change.EntityType = core_change.TelemetryProfile
	change.ApplicationType = applicationType
	change.NewEntity = newEntity
	change.Author = userName
	change.Operation = core_change.Create
	return change
}

func buildToUpdateChange(oldEntity *logupload.PermanentTelemetryProfile, newEntity *logupload.PermanentTelemetryProfile, applicationType string, userName string) *core_change.Change {
	change := xchange.NewEmptyChange()
	change.ID = uuid.New().String()
	change.EntityID = oldEntity.ID
	change.EntityType = core_change.TelemetryProfile
	change.ApplicationType = applicationType
	change.OldEntity = oldEntity
	change.NewEntity = newEntity
	change.Author = userName
	change.Operation = core_change.Update
	return change
}

func buildToDeleteChange(oldEntity *logupload.PermanentTelemetryProfile, applicationType string, userName string) *core_change.Change {
	change := xchange.NewEmptyChange()
	change.ID = uuid.New().String()
	change.EntityID = oldEntity.ID
	change.EntityType = core_change.TelemetryProfile
	change.ApplicationType = applicationType
	change.OldEntity = oldEntity
	change.Operation = core_change.Delete
	change.Author = userName
	return change
}

func validateUsage(id string) error {
	all := logupload.GetTelemetryRuleListForAs() //[]*TelemetryRule
	for _, rule := range all {
		if rule.BoundTelemetryID == id {
			return xwcommon.NewRemoteErrorAS(http.StatusConflict, "Can't delete profile as it's used in telemetry rule: "+rule.Name)
		}
	}

	return nil
}

func WriteCreateChange(r *http.Request, profile *logupload.PermanentTelemetryProfile) (*core_change.Change, error) {
	if err := auth.ValidateWrite(r, profile.ApplicationType, auth.TELEMETRY_ENTITY); err != nil {
		return nil, err
	}
	if err := beforeCreating(profile); err != nil {
		return nil, err
	}
	if err := beforeSavingPermanentTelemetryProfile(profile); err != nil {
		return nil, err
	}

	if err := ValidateTelemetryProfilePendingChanges(profile); err != nil {
		return nil, err
	}

	application, err := auth.CanWrite(r, auth.CHANGE_ENTITY, profile.ApplicationType)
	if err != nil {
		return nil, err
	}
	change := buildToCreateChange(profile, application, auth.GetUserNameOrUnknown(r))

	err = beforeSavingChange(r, change)
	if err != nil {
		return nil, err
	}
	if err := xchange.CreateOneChange(change); err != nil {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, err.Error())
	}

	return change, nil
}

func WriteUpdateChangeOrSave(r *http.Request, newProfile *logupload.PermanentTelemetryProfile) (*core_change.Change, error) {
	if err := auth.ValidateWrite(r, newProfile.ApplicationType, auth.TELEMETRY_ENTITY); err != nil {
		return nil, err
	}
	if err := beforeUpdating(newProfile); err != nil {
		return nil, err
	}
	if err := beforeSavingPermanentTelemetryProfile(newProfile); err != nil {
		return nil, err
	}

	change := core_change.NewEmptyChange()
	oldProfile := logupload.GetOnePermanentTelemetryProfile(newProfile.ID)
	if newProfile.EqualChangeData(oldProfile) {
		normalizeOnSaveAfterApproving(newProfile)
		if err := xlogupload.SetOnePermanentTelemetryProfile(newProfile.ID, newProfile); err != nil {
			return nil, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, err.Error())
		}
	} else {
		application, err := auth.CanWrite(r, auth.CHANGE_ENTITY, newProfile.ApplicationType)
		if err != nil {
			return nil, err
		}
		change = buildToUpdateChange(oldProfile, newProfile, application, auth.GetUserNameOrUnknown(r))
		err = beforeSavingChange(r, change)
		if err != nil {
			return nil, err
		}
		if err := xchange.CreateOneChange(change); err != nil {
			return nil, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, err.Error())
		}
	}

	return change, nil
}

func WriteDeleteChange(r *http.Request, profileId string) (*core_change.Change, error) {
	writeApplication, err := auth.CanWrite(r, auth.TELEMETRY_ENTITY)
	if err != nil {
		return nil, err
	}
	profile, err := beforeRemoving(profileId, writeApplication)
	if err != nil {
		return nil, err
	}
	application, err := auth.CanWrite(r, auth.CHANGE_ENTITY)
	if err != nil {
		return nil, err
	}
	change := buildToDeleteChange(profile, application, auth.GetUserNameOrUnknown(r))
	err = beforeSavingChange(r, change)
	if err != nil {
		return nil, err
	}
	if err := xchange.CreateOneChange(change); err != nil {
		return nil, err
	}
	return change, nil
}

func CreateTelemetryIds() *xwhttp.ResponseEntity {
	var migratedProfileNames []string
	profiles := logupload.GetPermanentTelemetryProfileList()
	if profiles == nil {
		return xwhttp.NewResponseEntity(http.StatusInternalServerError, fmt.Errorf("failed to load PermanentTelemetryProfile"), nil)
	}
	for _, profile := range profiles {
		normalizeOnSaveAfterApproving(profile)
		if err := xlogupload.SetOnePermanentTelemetryProfile(profile.ID, profile); err != nil {
			log.Error(fmt.Sprintf("failed to set PermanentTelemetryProfile: %v", err))
		} else {
			migratedProfileNames = append(migratedProfileNames, profile.Name)
		}
	}

	return xwhttp.NewResponseEntity(http.StatusOK, nil, migratedProfileNames)
}

func ApplyUpdateChange(mergeResult *logupload.PermanentTelemetryProfile, change *core_change.Change) *logupload.PermanentTelemetryProfile {
	if mergeResult == nil {
		return change.NewEntity
	}
	oldProfile := change.OldEntity
	updatedProfile := change.NewEntity
	if oldProfile.Name != updatedProfile.Name {
		mergeResult.Name = updatedProfile.Name
	}
	if oldProfile.Schedule != updatedProfile.Schedule {
		mergeResult.Schedule = updatedProfile.Schedule
	}
	if oldProfile.UploadProtocol != updatedProfile.UploadProtocol {
		mergeResult.UploadProtocol = updatedProfile.UploadProtocol
	}
	if oldProfile.UploadRepository != updatedProfile.UploadRepository {
		mergeResult.UploadRepository = updatedProfile.UploadRepository
	}
	return ApplyTelemetryElementChanges(change, mergeResult)
}

func ApplyTelemetryElementChanges(change *core_change.Change, mergeResult *logupload.PermanentTelemetryProfile) *logupload.PermanentTelemetryProfile {
	oldTelemetryElements := change.OldEntity.TelemetryProfile
	updatedTelemetryElements := change.NewEntity.TelemetryProfile
	for _, updated := range updatedTelemetryElements {
		old := FindTelemetryElementById(updated.ID, &oldTelemetryElements)
		merged := FindTelemetryElementById(updated.ID, &mergeResult.TelemetryProfile)
		if isNewElement(&updated) || removedBefore(old, &updated, merged) {
			mergeResult.TelemetryProfile = append(mergeResult.TelemetryProfile, updated)
			continue
		}
		applyTelemetryElementChange(merged, old, &updated)
	}
	RemoveTelemetryElementsFromMergeResult(getRemovedTelemetryElementIds(&oldTelemetryElements, &updatedTelemetryElements), mergeResult)
	return mergeResult
}

func RemoveTelemetryElementsFromMergeResult(idsToRemove *[]string, mergeResult *logupload.PermanentTelemetryProfile) {
	for _, id := range *idsToRemove {
		telemetryElementToRemove := FindTelemetryElementById(id, &mergeResult.TelemetryProfile)
		//mergeResult.getTelemetryProfile().remove(telemetryElementToRemove):
		for i := 0; i < len(mergeResult.TelemetryProfile); i++ {
			if mergeResult.TelemetryProfile[i].ID == telemetryElementToRemove.ID {
				mergeResult.TelemetryProfile = append(mergeResult.TelemetryProfile[:i], mergeResult.TelemetryProfile[i+1:]...)
				i--
			}
		}
	}
}

func FindTelemetryElementById(id string, telemetryElements *[]logupload.TelemetryElement) *logupload.TelemetryElement {
	for _, telemetryElement := range *telemetryElements {
		if telemetryElement.ID == id {
			return &telemetryElement
		}
	}
	return nil
}

func isNewElement(telemetryElement *logupload.TelemetryElement) bool {
	return telemetryElement.ID == "" && telemetryElement != nil
}

func removedBefore(old *logupload.TelemetryElement, updated *logupload.TelemetryElement, merged *logupload.TelemetryElement) bool {
	return !old.Equals(updated) && merged == nil
}

func applyTelemetryElementChange(mergedElement *logupload.TelemetryElement, oldElement *logupload.TelemetryElement, newElement *logupload.TelemetryElement) {
	if oldElement != nil && mergedElement != nil {
		if oldElement.Header != newElement.Header {
			mergedElement.Header = newElement.Header
		}
		if oldElement.Content != newElement.Content {
			mergedElement.Content = newElement.Content
		}
		if oldElement.Type != newElement.Type {
			mergedElement.Type = newElement.Type
		}
		if oldElement.PollingFrequency != newElement.PollingFrequency {
			mergedElement.PollingFrequency = newElement.PollingFrequency
		}
	}
}

func getRemovedTelemetryElementIds(oldElements *[]logupload.TelemetryElement, newElements *[]logupload.TelemetryElement) *[]string {
	removedElements := []string{}
	for _, oldElement := range *oldElements {
		if FindTelemetryElementById(oldElement.ID, newElements) == nil {
			removedElements = append(removedElements, oldElement.ID)
		}
	}
	return &removedElements
}

func AddPermanentTelemetryProfileElement(entry *logupload.TelemetryElement, telemetryEntries []logupload.TelemetryElement) ([]logupload.TelemetryElement, error) {
	exists, _ := doesEntryExist(entry, telemetryEntries)
	if exists {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusConflict, fmt.Sprintf("Telemetry Profile entry already exists: %v", *entry))
	}

	telemetryEntries = append(telemetryEntries, *entry)
	return telemetryEntries, nil
}

// Returns bool value and index
func doesEntryExist(entry *logupload.TelemetryElement, telemetryEntries []logupload.TelemetryElement) (bool, int) {
	if entry == nil || telemetryEntries == nil || len(telemetryEntries) == 0 {
		return false, -1
	}
	for i, element := range telemetryEntries {
		if entry.EqualTelemetryData(&element) {
			return true, i
		}
	}
	return false, -1
}

func RemovePermanentTelemetryProfileElement(entryToRemove *logupload.TelemetryElement, telemetryEntries []logupload.TelemetryElement) ([]logupload.TelemetryElement, error) {
	exists, i := doesEntryExist(entryToRemove, telemetryEntries)
	if !exists {
		return nil, xwcommon.NewRemoteErrorAS(http.StatusNotFound, "Telemetry entry does not exist") //TODO add enty to the error message
	}

	telemetryEntries = append(telemetryEntries[:i], telemetryEntries[i+1:]...)
	return telemetryEntries, nil
}

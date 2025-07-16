package change

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/shared"
	xwchange "github.com/rdkcentral/xconfwebconfig/shared/change"
	"github.com/rdkcentral/xconfwebconfig/util"
	xwutil "github.com/rdkcentral/xconfwebconfig/util"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	TelemetryTwoProfile = "TELEMETRY_TWO_PROFILE"
)

const (
	Create xwchange.ChangeOperation = "CREATE"
	Update xwchange.ChangeOperation = "UPDATE"
	Delete xwchange.ChangeOperation = "DELETE"
)

func GetChangeList() []*xwchange.Change {
	all := []*xwchange.Change{}
	changeList, err := db.GetSimpleDao().GetAllAsList(db.TABLE_XCONF_CHANGE, 0)
	if err != nil {
		log.Warn("no Change found")
		return nil
	}
	for idx := range changeList {
		change := changeList[idx].(*xwchange.Change)
		all = append(all, change)
	}
	return all
}

func SetOneApprovedChange(approvedChange *xwchange.ApprovedChange) error {
	approvedChange.Updated = xwutil.GetTimestamp(time.Now().UTC())

	approvedChangeBytes, err := json.Marshal(approvedChange)
	if err != nil {
		return err
	}

	return db.GetSimpleDao().SetOne(db.TABLE_XCONF_APPROVED_CHANGE, approvedChange.ID, approvedChangeBytes)
}

func GetOneApprovedChange(id string) *xwchange.ApprovedChange {
	var change *xwchange.ApprovedChange
	changeInst, err := db.GetSimpleDao().GetOne(db.TABLE_XCONF_APPROVED_CHANGE, id)
	if err != nil {
		log.Warn(fmt.Sprintf("no Approved found for Id: %s", id))
		return nil
	}
	change = changeInst.(*xwchange.ApprovedChange)
	return change
}

func GetApprovedChangeList() []*xwchange.ApprovedChange {
	all := []*xwchange.ApprovedChange{}
	approvedList, err := db.GetSimpleDao().GetAllAsList(db.TABLE_XCONF_APPROVED_CHANGE, 0)
	if err != nil {
		log.Warn("no Change found")
		return nil
	}
	for idx := range approvedList {
		approved := approvedList[idx].(*xwchange.ApprovedChange)
		all = append(all, approved)
	}
	return all
}

func GetChangesByEntityId(entityId string) []*xwchange.Change {
	result := []*xwchange.Change{}
	all := GetChangeList()
	for _, change := range all {
		if change.EntityID == entityId {
			result = append(result, change)
		}
	}
	return result
}

func GetOneChange(id string) *xwchange.Change {
	var change *xwchange.Change
	changeInst, err := db.GetSimpleDao().GetOne(db.TABLE_XCONF_CHANGE, id)
	if err != nil {
		log.Warn(fmt.Sprintf("no Change found for Id: %s", id))
		return nil
	}
	change = changeInst.(*xwchange.Change)
	return change
}

func DeleteOneChange(id string) error {
	return db.GetSimpleDao().DeleteOne(db.TABLE_XCONF_CHANGE, id)
}

func DeleteOneApprovedChange(id string) error {
	return db.GetSimpleDao().DeleteOne(db.TABLE_XCONF_APPROVED_CHANGE, id)
}

func NewEmptyChange() *xwchange.Change {
	return &xwchange.Change{
		ApplicationType: shared.STB,
	}
}

func NewEmptyTelemetryTwoChange() *xwchange.TelemetryTwoChange {
	return &xwchange.TelemetryTwoChange{
		ApplicationType: shared.STB,
	}
}

func CreateOneChange(change *xwchange.Change) error {
	change.Updated = util.GetTimestamp()

	changeBytes, err := json.Marshal(change)
	if err != nil {
		return err
	}

	return db.GetSimpleDao().SetOne(db.TABLE_XCONF_CHANGE, change.ID, changeBytes)
}

func GetApprovedTelemetryTwoChangesByApplicationType(applicationType string) []*xwchange.ApprovedTelemetryTwoChange {
	all := []*xwchange.ApprovedTelemetryTwoChange{}
	list, err := db.GetSimpleDao().GetAllAsList(db.TABLE_XCONF_APPROVED_TELEMETRY_TWO_CHANGE, 0)
	if err != nil {
		log.Warn("no xwchange.ApprovedTelemetryTwoChange found")
		return nil
	}
	for _, inst := range list {
		change := inst.(*xwchange.ApprovedTelemetryTwoChange)
		if change.ApplicationType != applicationType {
			continue
		}
		all = append(all, change)
	}
	return all
}

func GetAllTelemetryTwoChangeList() []*xwchange.TelemetryTwoChange {
	all := []*xwchange.TelemetryTwoChange{}
	list, err := db.GetSimpleDao().GetAllAsList(db.TABLE_XCONF_TELEMETRY_TWO_CHANGE, 0)
	if err != nil {
		log.Warn("no TelemetryTwoChange found")
		return nil
	}
	for _, inst := range list {
		change := inst.(*xwchange.TelemetryTwoChange)
		all = append(all, change)
	}
	return all
}

func CreateOneTelemetryTwoChange(change *xwchange.TelemetryTwoChange) error {
	// create record in DB
	if util.IsBlank(change.ID) {
		change.ID = uuid.New().String()
	}
	change.Updated = util.GetTimestamp()

	changeBytes, err := json.Marshal(change)
	if err != nil {
		return err
	}

	return db.GetSimpleDao().SetOne(db.TABLE_XCONF_TELEMETRY_TWO_CHANGE, change.ID, changeBytes)
}

func GetAllApprovedTelemetryTwoChangeList() []*xwchange.ApprovedTelemetryTwoChange {
	all := []*xwchange.ApprovedTelemetryTwoChange{}
	list, err := db.GetSimpleDao().GetAllAsList(db.TABLE_XCONF_APPROVED_TELEMETRY_TWO_CHANGE, 0)
	if err != nil {
		log.Warn("no xwchange.ApprovedTelemetryTwoChange found")
		return nil
	}
	for _, inst := range list {
		change := inst.(*xwchange.ApprovedTelemetryTwoChange)
		all = append(all, change)
	}
	return all
}

func GetOneTelemetryTwoChange(id string) *xwchange.TelemetryTwoChange {
	var change *xwchange.TelemetryTwoChange
	changeInst, err := db.GetSimpleDao().GetOne(db.TABLE_XCONF_TELEMETRY_TWO_CHANGE, id)
	if err != nil {
		log.Warn(fmt.Sprintf("no TelemetryTwoChange found for Id: %s", id))
		return nil
	}
	change = changeInst.(*xwchange.TelemetryTwoChange)
	return change
}

func NewApprovedTelemetryTwoChange(change *xwchange.TelemetryTwoChange) *xwchange.ApprovedTelemetryTwoChange {
	return &xwchange.ApprovedTelemetryTwoChange{
		ID:              change.ID,
		EntityID:        change.EntityID,
		EntityType:      change.EntityType,
		ApplicationType: change.ApplicationType,
		Author:          change.Author,
		ApprovedUser:    change.ApprovedUser,
		Operation:       change.Operation,
		OldEntity:       change.OldEntity,
		NewEntity:       change.NewEntity,
	}
}

func SetOneApprovedTelemetryTwoChange(approvedChange *xwchange.ApprovedTelemetryTwoChange) error {
	// create record in DB
	if util.IsBlank(approvedChange.ID) {
		approvedChange.ID = uuid.New().String()
	}
	approvedChange.Updated = util.GetTimestamp(time.Now().UTC())

	approvedChangeBytes, err := json.Marshal(approvedChange)
	if err != nil {
		return err
	}

	return db.GetSimpleDao().SetOne(db.TABLE_XCONF_APPROVED_TELEMETRY_TWO_CHANGE, approvedChange.ID, approvedChangeBytes)
}

func DeleteOneTelemetryTwoChange(id string) error {
	return db.GetSimpleDao().DeleteOne(db.TABLE_XCONF_TELEMETRY_TWO_CHANGE, id)
}

func GetOneApprovedTelemetryTwoChange(id string) *xwchange.ApprovedTelemetryTwoChange {
	var change *xwchange.ApprovedTelemetryTwoChange
	changeInst, err := db.GetSimpleDao().GetOne(db.TABLE_XCONF_APPROVED_TELEMETRY_TWO_CHANGE, id)
	if err != nil {
		log.Warn(fmt.Sprintf("no xwchange.ApprovedTelemetryTwoChange found for Id: %s", id))
		return nil
	}
	change = changeInst.(*xwchange.ApprovedTelemetryTwoChange)
	return change
}

func DeleteOneApprovedTelemetryTwoChange(id string) error {
	return db.GetSimpleDao().DeleteOne(db.TABLE_XCONF_APPROVED_TELEMETRY_TWO_CHANGE, id)
}

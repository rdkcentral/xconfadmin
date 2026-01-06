package applicationtype

import (
	"encoding/json"
	"strings"

	db "github.com/rdkcentral/xconfwebconfig/db"
)

const (
	TABLE_APPLICATION_TYPES = "ApplicationTypes"
)

// Get one application type
func GetOneApplicationType(id string) (*ApplicationType, error) {
	result, err := db.GetCachedSimpleDao().GetOne(TABLE_APPLICATION_TYPES, id)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	appType, err := toApplicationType(result)
	if err != nil {
		return nil, err
	}
	return appType, nil
}

func SetOneApplicationType(appType *ApplicationType) error {
	DeleteOneApplicationType(appType.ID)
	return db.GetCachedSimpleDao().SetOne(TABLE_APPLICATION_TYPES, appType.ID, appType)
}

func DeleteOneApplicationType(id string) error {
	return db.GetCachedSimpleDao().DeleteOne(TABLE_APPLICATION_TYPES, id)
}

func GetApplicationTypeByName(name string) (bool, error) {
	appTypes, err := GetAllApplicationTypeAsList()
	if err != nil {
		return false, err
	}

	for _, appType := range appTypes {
		if strings.EqualFold(appType.Name, name) {
			return true, nil
		}
	}
	return false, nil
}

// Get all application types as list
func GetAllApplicationTypeAsList() ([]*ApplicationType, error) {
	appTypeMap, err := db.GetCachedSimpleDao().GetAllAsMap(TABLE_APPLICATION_TYPES)
	if err != nil {
		return nil, err
	}

	var appTypeList []*ApplicationType
	for _, v := range appTypeMap {
		appType, err := toApplicationType(v)
		if err != nil {
			return nil, err
		}
		appTypeList = append(appTypeList, appType)
	}
	return appTypeList, nil
}

func toApplicationType(v interface{}) (*ApplicationType, error) {
	if appType, ok := v.(*ApplicationType); ok {
		return appType, nil
	}

	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	appType := &ApplicationType{}
	return appType, json.Unmarshal(jsonBytes, appType)
}

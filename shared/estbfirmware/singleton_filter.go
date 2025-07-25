package estbfirmware

import (
	"encoding/json"
	"fmt"
	"strings"

	core "github.com/rdkcentral/xconfadmin/shared"
	"github.com/rdkcentral/xconfadmin/util"

	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
)

const (
	PERCENT_FILTER_SINGLETON_ID     = "PERCENT_FILTER_VALUE"
	ROUND_ROBIN_FILTER_SINGLETON_ID = "DOWNLOAD_LOCATION_ROUND_ROBIN_FILTER_VALUE"
)

type SingletonFilterClass string

const (
	PercentFilterClass        SingletonFilterClass = "com.comcast.xconf.estbfirmware.PercentFilterValue"
	PercentFilterWrapperClass SingletonFilterClass = "com.comcast.xconf.queries.beans.PercentFilterWrapper"
	RoundRobinFilterClass     SingletonFilterClass = "com.comcast.xconf.estbfirmware.DownloadLocationRoundRobinFilterValue"
)

// SingletonFilterValue table - this struct serves as a container for the two subtypes
type SingletonFilterValue struct {
	ID                                    string                                        `json:"id"`
	PercentFilterValue                    *PercentFilterValue                           `json:"-"`
	DownloadLocationRoundRobinFilterValue *coreef.DownloadLocationRoundRobinFilterValue `json:"-"`
}

func (obj *SingletonFilterValue) Clone() (*SingletonFilterValue, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*SingletonFilterValue), nil
}

func NewSingletonFilterValueInf() interface{} {
	return &SingletonFilterValue{}
}

func (sfv *SingletonFilterValue) IsPercentFilterValue() bool {
	return strings.HasSuffix(sfv.ID, PERCENT_FILTER_SINGLETON_ID)
}

func (sfv *SingletonFilterValue) IsDownloadLocationRoundRobinFilterValue() bool {
	return strings.HasSuffix(sfv.ID, ROUND_ROBIN_FILTER_SINGLETON_ID)
}

// // UnmarshalJSON custom unmarshal to handle different subclass of SingletonFilterValue
func (sfv *SingletonFilterValue) UnmarshalJSON(bytes []byte) error {
	type singletonFilterValue SingletonFilterValue

	// Unmarshal just the base class to get the ID
	err := json.Unmarshal(bytes, (*singletonFilterValue)(sfv))
	if err != nil {
		return err
	}

	// Unmarshal the subtype based on the ID
	if sfv.IsPercentFilterValue() {
		var obj PercentFilterValue
		err = json.Unmarshal(bytes, &obj)
		if err != nil {
			return err
		}
		sfv.PercentFilterValue = &obj
	} else if sfv.IsDownloadLocationRoundRobinFilterValue() {
		var obj coreef.DownloadLocationRoundRobinFilterValue
		err = json.Unmarshal(bytes, &obj)
		if err != nil {
			return err
		}
		sfv.DownloadLocationRoundRobinFilterValue = &obj
	} else {
		return fmt.Errorf("Invalid ID for SingletonFilterValue: %v", string(bytes))
	}

	return nil
}

// MarshalJSON custom marshal to handle different subclass of SingletonFilterValue
func (sfv *SingletonFilterValue) MarshalJSON() ([]byte, error) {
	// Unmarshal the subtype
	if sfv.PercentFilterValue != nil && sfv.PercentFilterValue.ID != "" {
		return json.Marshal(sfv.PercentFilterValue)
	} else if sfv.DownloadLocationRoundRobinFilterValue != nil && sfv.DownloadLocationRoundRobinFilterValue.ID != "" {
		return json.Marshal(sfv.DownloadLocationRoundRobinFilterValue)
	} else {
		return nil, fmt.Errorf("Invalid SingletonFilterValue: %v", sfv)
	}
}

func GetRoundRobinIdByApplication(applicationType string) string {
	if core.STB == applicationType {
		return ROUND_ROBIN_FILTER_SINGLETON_ID
	}
	return fmt.Sprintf("%s_%s", strings.ToUpper(applicationType), ROUND_ROBIN_FILTER_SINGLETON_ID)
}

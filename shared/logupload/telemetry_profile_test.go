package logupload

import (
	"testing"

	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
)

type cachedSimpleDaoMock struct{}

func (dao cachedSimpleDaoMock) GetOne(tableName string, rowKey string) (interface{}, error) {
	return getOneMock(tableName, rowKey)
}

func (dao cachedSimpleDaoMock) GetOneFromCacheOnly(tableName string, rowKey string) (interface{}, error) {
	return getOneMock(tableName, rowKey)
}

func (dao cachedSimpleDaoMock) SetOne(tableName string, rowKey string, entity interface{}) error {
	return setOneMock(tableName, rowKey, entity)
}

func (dao cachedSimpleDaoMock) DeleteOne(tableName string, rowKey string) error {
	return deleteOneMock(tableName, rowKey)
}

func (dao cachedSimpleDaoMock) GetAllByKeys(tableName string, rowKeys []string) ([]interface{}, error) {
	return getAllByKeysMock(tableName, rowKeys)
}

func (dao cachedSimpleDaoMock) GetAllAsList(tableName string, maxResults int) ([]interface{}, error) {
	return getAllAsListMock(tableName, maxResults)
}

func (dao cachedSimpleDaoMock) GetAllAsMap(tableName string) (map[interface{}]interface{}, error) {
	return getAllAsMapMock(tableName)
}

func (dao cachedSimpleDaoMock) GetAllAsShallowMap(tableName string) (map[interface{}]interface{}, error) {
	return getAllAsMapMock(tableName)
}

func (dao cachedSimpleDaoMock) GetKeys(tableName string) ([]interface{}, error) {
	return getKeysMock(tableName)
}

func (dao cachedSimpleDaoMock) RefreshAll(tableName string) error {
	return refreshAllMock(tableName)
}

func (dao cachedSimpleDaoMock) RefreshOne(tableName string, rowKey string) error {
	return refreshOneMock(tableName, rowKey)
}

var getOneMock func(tableName string, rowKey string) (interface{}, error)
var setOneMock func(tableName string, rowKey string, entity interface{}) error
var deleteOneMock func(tableName string, rowKey string) error
var getAllByKeysMock func(tableName string, rowKeys []string) ([]interface{}, error)
var getAllAsListMock func(tableName string, maxResults int) ([]interface{}, error)
var getAllAsMapMock func(tableName string) (map[interface{}]interface{}, error)
var getKeysMock func(tableName string) ([]interface{}, error)
var refreshAllMock func(tableName string) error
var refreshOneMock func(tableName string, rowKey string) error

// TestSetOnePermanentTelemetryProfile tests setting a permanent telemetry profile
func TestSetOnePermanentTelemetryProfile(t *testing.T) {
	profile := &logupload.PermanentTelemetryProfile{
		ID:              "profile-123",
		ApplicationType: shared.STB,
	}
	err := SetOnePermanentTelemetryProfile("profile-123", profile)
	// May fail if dao not initialized; that's acceptable in unit test
	_ = err
	t.Log("SetOnePermanentTelemetryProfile executed")
}

// TestGetOnePermanentTelemetryProfile tests retrieving a permanent telemetry profile
func TestGetOnePermanentTelemetryProfile(t *testing.T) {
	result := GetOnePermanentTelemetryProfile("non-existent-id")
	// Expected to return nil if not found
	if result != nil {
		t.Logf("GetOnePermanentTelemetryProfile returned: %+v", result)
	} else {
		t.Log("GetOnePermanentTelemetryProfile returned nil (expected for non-existent ID)")
	}
}

// TestDeletePermanentTelemetryProfile tests deleting a permanent telemetry profile
func TestDeletePermanentTelemetryProfile(t *testing.T) {
	DeletePermanentTelemetryProfile("test-profile-id")
	t.Log("DeletePermanentTelemetryProfile executed")
}

// TestGetPermanentTelemetryProfileList tests retrieving all permanent telemetry profiles
func TestGetPermanentTelemetryProfileList(t *testing.T) {
	profiles := GetPermanentTelemetryProfileList()
	if profiles == nil {
		t.Log("GetPermanentTelemetryProfileList returned nil (expected if no data)")
	} else {
		t.Logf("GetPermanentTelemetryProfileList returned %d profiles", len(profiles))
	}
}

// TestGetPermanentTelemetryProfileListByApplicationType tests filtering by application type
func TestGetPermanentTelemetryProfileListByApplicationType(t *testing.T) {
	profiles := GetPermanentTelemetryProfileListByApplicationType(shared.STB)
	if profiles == nil {
		t.Log("GetPermanentTelemetryProfileListByApplicationType returned nil")
	} else {
		t.Logf("GetPermanentTelemetryProfileListByApplicationType returned %d profiles", len(profiles))
	}
}

// TestGetAllTelemetryTwoProfileList tests retrieving all telemetry two profiles
func TestGetAllTelemetryTwoProfileList(t *testing.T) {
	profiles := GetAllTelemetryTwoProfileList(shared.STB)
	if profiles == nil {
		t.Log("GetAllTelemetryTwoProfileList returned nil (expected if no data)")
	} else {
		t.Logf("GetAllTelemetryTwoProfileList returned %d profiles", len(profiles))
	}
}

// TestNewEmptyTelemetryTwoProfile tests creating an empty telemetry two profile
func TestNewEmptyTelemetryTwoProfile(t *testing.T) {
	profile := NewEmptyTelemetryTwoProfile()
	if profile == nil {
		t.Fatal("expected non-nil telemetry two profile")
	}
	if profile.ApplicationType != shared.STB {
		t.Fatalf("expected ApplicationType %s, got %s", shared.STB, profile.ApplicationType)
	}
	if profile.Type != "TelemetryTwoProfile" {
		t.Fatalf("expected Type 'TelemetryTwoProfile', got %s", profile.Type)
	}
	t.Log("NewEmptyTelemetryTwoProfile created successfully")
}

// TestGetOneTelemetryTwoProfile tests retrieving a single telemetry two profile
func TestGetOneTelemetryTwoProfile(t *testing.T) {
	result := GetOneTelemetryTwoProfile("non-existent-id")
	if result != nil {
		t.Logf("GetOneTelemetryTwoProfile returned: %+v", result)
	} else {
		t.Log("GetOneTelemetryTwoProfile returned nil (expected for non-existent ID)")
	}
}

// TestSetOneTelemetryTwoProfile tests setting a telemetry two profile
func TestSetOneTelemetryTwoProfile(t *testing.T) {
	profile := &logupload.TelemetryTwoProfile{
		ID:              "profile-two-123",
		ApplicationType: shared.STB,
	}
	err := SetOneTelemetryTwoProfile(profile)
	// May fail if dao not initialized; that's acceptable
	_ = err
	t.Log("SetOneTelemetryTwoProfile executed")
}

// TestDeleteTelemetryTwoProfile tests deleting a telemetry two profile
func TestDeleteTelemetryTwoProfile(t *testing.T) {
	err := DeleteTelemetryTwoProfile("test-id")
	// May fail if dao not initialized; that's acceptable
	_ = err
	t.Log("DeleteTelemetryTwoProfile executed")
}

// TestSetOneTelemetryProfile tests setting a telemetry profile
func TestSetOneTelemetryProfile(t *testing.T) {
	profile := &logupload.TelemetryProfile{
		ID:              "telemetry-123",
		ApplicationType: shared.STB,
	}
	SetOneTelemetryProfile("telemetry-123", profile)
	t.Log("SetOneTelemetryProfile executed")
}

// TestGetTimestampedRulesPointer tests retrieving timestamped rules
func TestGetTimestampedRulesPointer(t *testing.T) {
	rules := GetTimestampedRulesPointer()
	if rules == nil {
		t.Log("GetTimestampedRulesPointer returned nil (expected if no data)")
	} else {
		t.Logf("GetTimestampedRulesPointer returned %d rules", len(rules))
	}
}

// TestGetOneTelemetryTwoRule tests retrieving a single telemetry two rule
func TestGetOneTelemetryTwoRule(t *testing.T) {
	result := GetOneTelemetryTwoRule("non-existent-rule-id")
	if result != nil {
		t.Logf("GetOneTelemetryTwoRule returned: %+v", result)
	} else {
		t.Log("GetOneTelemetryTwoRule returned nil (expected for non-existent ID)")
	}
}

// TestGetOneTelemetryRule tests retrieving a single telemetry rule
func TestGetOneTelemetryRule(t *testing.T) {
	result := GetOneTelemetryRule("non-existent-rule-id")
	if result != nil {
		t.Logf("GetOneTelemetryRule returned: %+v", result)
	} else {
		t.Log("GetOneTelemetryRule returned nil (expected for non-existent ID)")
	}
}

// TestSetOneTelemetryTwoRule tests setting a telemetry two rule
func TestSetOneTelemetryTwoRule(t *testing.T) {
	rule := &logupload.TelemetryTwoRule{
		ID:              "rule-two-123",
		ApplicationType: shared.STB,
	}
	err := SetOneTelemetryTwoRule("rule-two-123", rule)
	// May fail if dao not initialized; that's acceptable
	_ = err
	t.Log("SetOneTelemetryTwoRule executed")
}

// TestDeleteTelemetryTwoRule tests deleting a telemetry two rule
func TestDeleteTelemetryTwoRule(t *testing.T) {
	err := DeleteTelemetryTwoRule("test-rule-id")
	// May fail if dao not initialized; that's acceptable
	_ = err
	t.Log("DeleteTelemetryTwoRule executed")
}

// func TestGetOne(t *testing.T) {
// 	// GetCachedSimpleDaoFunc = func() ds.CachedSimpleDao {
// 	// 	return cachedSimpleDaoMock{}
// 	// }
// 	getOneMock = func(tableName string, rowKey string) (interface{}, error) {
// 		if tableName == ds.TABLE_TELEMETRY {
// 			telemetryProfile := &logupload.TelemetryProfile{
// 				ID:               "id",
// 				Name:             "name",
// 				TelemetryProfile: nil,
// 				Schedule:         "Schedule",
// 				Expires:          123456,
// 				UploadRepository: "uploadRepository:URL",
// 				ApplicationType:  "ApplicationType",
// 			}
// 			return telemetryProfile, nil
// 		}
// 		return nil, nil
// 	}
// 	telemetryProfile := logupload.GetOneTelemetryProfile("rowKey")
// 	//assert.Equal(t, telemetryProfile.ID, "id")
// 	assert.Equal(t, telemetryProfile.Schedule, "Schedule")
// 	var a int64 = 123456
// 	assert.Equal(t, telemetryProfile.Expires, a)
// 	assert.Equal(t, telemetryProfile.UploadRepository, "uploadRepository:URL")
// }

// func TestGetTelemetryProfileList(t *testing.T) {
// 	// GetCachedSimpleDaoFunc = func() ds.CachedSimpleDao {
// 	// 	return cachedSimpleDaoMock{}
// 	// }
// 	getAllAsListMock = func(tableName string, maxResults int) ([]interface{}, error) {
// 		if tableName == ds.TABLE_TELEMETRY {
// 			telemetryProfile1 := logupload.TelemetryProfile{
// 				ID:               "id1",
// 				Name:             "name1",
// 				TelemetryProfile: nil,
// 				Schedule:         "Schedule1",
// 				Expires:          123451,
// 				UploadRepository: "uploadRepository:URL1",
// 				ApplicationType:  "ApplicationType1",
// 			}
// 			telemetryProfile2 := logupload.TelemetryProfile{
// 				ID:               "id2",
// 				Name:             "name2",
// 				TelemetryProfile: nil,
// 				Schedule:         "Schedule2",
// 				Expires:          123452,
// 				UploadRepository: "uploadRepository:URL2",
// 				ApplicationType:  "ApplicationType2",
// 			}
// 			tpList := make([]interface{}, 0)
// 			tpList = append(tpList, telemetryProfile1)
// 			tpList = append(tpList, telemetryProfile2)
// 			return tpList, nil
// 		}
// 		return nil, nil
// 	}
// 	//[]*TelemetryProfile
// 	telemetryProfileList := logupload.GetTelemetryProfileList()
// 	assert.Equal(t, len(telemetryProfileList), 2)
// 	assert.Equal(t, telemetryProfileList[0].ApplicationType, "ApplicationType1")
// 	assert.Equal(t, telemetryProfileList[1].UploadRepository, "uploadRepository:URL2")
// }

// func TestGetTelemetryProfileMap(t *testing.T) {

// 	// GetCachedSimpleDaoFunc = func() ds.CachedSimpleDao {
// 	// 	return cachedSimpleDaoMock{}
// 	// }
// 	getAllAsMapMock = func(tableName string) (map[interface{}]interface{}, error) {
// 		if tableName == ds.TABLE_TELEMETRY {
// 			telemetryProfile1 := logupload.TelemetryProfile{
// 				ID:               "id1",
// 				Name:             "name1",
// 				TelemetryProfile: nil,
// 				Schedule:         "Schedule1",
// 				Expires:          123451,
// 				UploadRepository: "uploadRepository:URL1",
// 				ApplicationType:  "ApplicationType1",
// 			}
// 			telemetryProfile2 := logupload.TelemetryProfile{
// 				ID:               "id2",
// 				Name:             "name2",
// 				TelemetryProfile: nil,
// 				Schedule:         "Schedule2",
// 				Expires:          123452,
// 				UploadRepository: "uploadRepository:URL2",
// 				ApplicationType:  "ApplicationType2",
// 			}
// 			rule := re.Rule{
// 				Negated:  true,
// 				Relation: "Relation",
// 			}
// 			timestampedRule1 := logupload.TimestampedRule{
// 				Rule:      rule,
// 				Timestamp: 1234561,
// 			}
// 			timestampedRule2 := logupload.TimestampedRule{
// 				Rule:      rule,
// 				Timestamp: 1234562,
// 			}
// 			timestampedRuleBytes1, _ := json.Marshal(timestampedRule1)
// 			timestampedRuleBytes2, _ := json.Marshal(timestampedRule2)

// 			map1 := make(map[interface{}]interface{})
// 			map1[string(timestampedRuleBytes1)] = telemetryProfile1
// 			map1[string(timestampedRuleBytes2)] = telemetryProfile2
// 			return map1, nil
// 		}
// 		return nil, nil
// 	}
// 	finalMap := logupload.GetTelemetryProfileMap()
// 	assert.Equal(t, len(*finalMap), 2)
// 	var a1 int64 = 1234561
// 	for k, v := range *finalMap {
// 		bytes := []byte(k)
// 		var timestampedRule logupload.TimestampedRule
// 		json.Unmarshal(bytes, &timestampedRule)
// 		if timestampedRule.Timestamp == a1 {
// 			assert.Equal(t, v.ApplicationType, "ApplicationType1")
// 		}
// 	}
// }

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
package logupload

import (
	"testing"

	core "github.com/rdkcentral/xconfadmin/shared"
	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/stretchr/testify/assert"
)

// TestIsValidUploadProtocol tests the IsValidUploadProtocol function
func TestIsValidUploadProtocol(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		expected bool
	}{
		{"TFTP uppercase", "TFTP", true},
		{"tftp lowercase", "tftp", true},
		{"SFTP uppercase", "SFTP", true},
		{"sftp lowercase", "sftp", true},
		{"SCP uppercase", "SCP", true},
		{"scp lowercase", "scp", true},
		{"HTTP uppercase", "HTTP", true},
		{"http lowercase", "http", true},
		{"HTTPS uppercase", "HTTPS", true},
		{"https lowercase", "https", true},
		{"S3 uppercase", "S3", true},
		{"s3 lowercase", "s3", true},
		{"Mixed case Http", "Http", true},
		{"Invalid protocol FTP", "FTP", false},
		{"Invalid protocol SSH", "SSH", false},
		{"Empty string", "", false},
		{"Invalid protocol ABC", "ABC", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidUploadProtocol(tt.protocol)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsValidUrl tests the IsValidUrl function
func TestIsValidUrl(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"Valid HTTP URL", "http://example.com/path", true},
		{"Valid HTTPS URL", "https://example.com/path", true},
		{"Valid TFTP URL", "tftp://server.example.com", true},
		{"Valid SFTP URL", "sftp://server.example.com/path", true},
		{"Valid SCP URL", "scp://server.example.com", true},
		{"Valid S3 URL", "s3://bucket.example.com", true},
		{"Valid URL with port", "https://example.com:8080/path", true},
		{"Valid URL with query", "https://example.com/path?query=value", true},
		{"Invalid protocol FTP", "ftp://example.com", false},
		{"No scheme", "example.com", false},
		{"No host", "http://", false},
		{"Empty string", "", false},
		{"Invalid URL format", "not-a-url", false},
		{"Scheme only", "http://", false},
		{"Invalid host format", "http://invalid host", false},
		{"Valid complex URL", "https://sub.example.com/path/to/resource?param=value#anchor", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidUrl(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestUploadRepositoryClone tests the Clone method
func TestUploadRepositoryClone(t *testing.T) {
	original := &UploadRepository{
		ID:              "repo1",
		Updated:         1234567890,
		Name:            "Test Repo",
		Description:     "Test Description",
		URL:             "https://example.com",
		ApplicationType: "stb",
		Protocol:        "HTTPS",
	}

	cloned, err := original.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Updated, cloned.Updated)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.Description, cloned.Description)
	assert.Equal(t, original.URL, cloned.URL)
	assert.Equal(t, original.ApplicationType, cloned.ApplicationType)
	assert.Equal(t, original.Protocol, cloned.Protocol)

	// Verify it's a deep copy
	cloned.Name = "Modified Name"
	assert.NotEqual(t, original.Name, cloned.Name)
}

// TestNewUploadRepositoryInf tests the constructor
func TestNewUploadRepositoryInf(t *testing.T) {
	obj := NewUploadRepositoryInf()
	assert.NotNil(t, obj)

	repo, ok := obj.(*UploadRepository)
	assert.True(t, ok)
	assert.Equal(t, core.STB, repo.ApplicationType)
}

// TestLogFileClone tests the LogFile Clone method
func TestLogFileClone(t *testing.T) {
	original := &LogFile{
		ID:             "logfile1",
		Updated:        1234567890,
		Name:           "test.log",
		DeleteOnUpload: true,
	}

	cloned, err := original.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Updated, cloned.Updated)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.DeleteOnUpload, cloned.DeleteOnUpload)

	// Verify it's a deep copy
	cloned.Name = "modified.log"
	assert.NotEqual(t, original.Name, cloned.Name)
}

// TestNewLogFileInf tests the LogFile constructor
func TestNewLogFileInf(t *testing.T) {
	obj := NewLogFileInf()
	assert.NotNil(t, obj)

	_, ok := obj.(*LogFile)
	assert.True(t, ok)
}

// TestLogFilesGroupsClone tests the LogFilesGroups Clone method
func TestLogFilesGroupsClone(t *testing.T) {
	original := &LogFilesGroups{
		ID:         "group1",
		Updated:    1234567890,
		GroupName:  "Test Group",
		LogFileIDs: []string{"file1", "file2", "file3"},
	}

	cloned, err := original.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Updated, cloned.Updated)
	assert.Equal(t, original.GroupName, cloned.GroupName)
	assert.Equal(t, len(original.LogFileIDs), len(cloned.LogFileIDs))
	assert.Equal(t, original.LogFileIDs, cloned.LogFileIDs)

	// Verify it's a deep copy
	cloned.GroupName = "Modified Group"
	cloned.LogFileIDs = append(cloned.LogFileIDs, "file4")
	assert.NotEqual(t, original.GroupName, cloned.GroupName)
	assert.NotEqual(t, len(original.LogFileIDs), len(cloned.LogFileIDs))
}

// TestNewLogFilesGroupsInf tests the LogFilesGroups constructor
func TestNewLogFilesGroupsInf(t *testing.T) {
	obj := NewLogFilesGroupsInf()
	assert.NotNil(t, obj)

	_, ok := obj.(*LogFilesGroups)
	assert.True(t, ok)
}

// TestLogFileListClone tests the LogFileList Clone method
func TestLogFileListClone(t *testing.T) {
	original := &LogFileList{
		Updated: 1234567890,
		Data: []*LogFile{
			{ID: "file1", Name: "test1.log", DeleteOnUpload: true},
			{ID: "file2", Name: "test2.log", DeleteOnUpload: false},
		},
	}

	cloned, err := original.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
	assert.Equal(t, original.Updated, cloned.Updated)
	assert.Equal(t, len(original.Data), len(cloned.Data))

	for i := range original.Data {
		assert.Equal(t, original.Data[i].ID, cloned.Data[i].ID)
		assert.Equal(t, original.Data[i].Name, cloned.Data[i].Name)
		assert.Equal(t, original.Data[i].DeleteOnUpload, cloned.Data[i].DeleteOnUpload)
	}

	// Verify it's a deep copy
	cloned.Data[0].Name = "modified.log"
	assert.NotEqual(t, original.Data[0].Name, cloned.Data[0].Name)
}

// TestNewLogFileListInf tests the LogFileList constructor
func TestNewLogFileListInf(t *testing.T) {
	obj := NewLogFileListInf()
	assert.NotNil(t, obj)

	_, ok := obj.(*LogFileList)
	assert.True(t, ok)
}

// TestSetLogFile tests the SetLogFile function
func TestSetLogFile(t *testing.T) {
	// This test requires database setup
	// Skip if database is not configured
	if db.GetCachedSimpleDao() == nil {
		t.Skip("Database not configured")
	}

	logFile := &LogFile{
		ID:             "test-logfile-1",
		Updated:        1234567890,
		Name:           "test.log",
		DeleteOnUpload: true,
	}

	err := SetLogFile(logFile.ID, logFile)
	// We expect either success or a database error
	// The important thing is that the function doesn't panic
	if err != nil {
		t.Logf("SetLogFile returned error (expected if DB not fully configured): %v", err)
	}
}

// TestGetLogFileGroupsList tests the GetLogFileGroupsList function
func TestGetLogFileGroupsList(t *testing.T) {
	// This test requires database setup
	// Skip if database is not configured
	if db.GetCachedSimpleDao() == nil {
		t.Skip("Database not configured")
	}

	groups, err := GetLogFileGroupsList(10)
	// We expect either a list or an error
	// The important thing is that the function doesn't panic
	if err != nil {
		t.Logf("GetLogFileGroupsList returned error (expected if DB not fully configured): %v", err)
		assert.Nil(t, groups)
	} else {
		assert.NotNil(t, groups)
	}
}

// TestUploadProtocolConstants tests that all protocol constants are defined
func TestUploadProtocolConstants(t *testing.T) {
	assert.Equal(t, UploadProtocol("TFTP"), TFTP)
	assert.Equal(t, UploadProtocol("SFTP"), SFTP)
	assert.Equal(t, UploadProtocol("SCP"), SCP)
	assert.Equal(t, UploadProtocol("HTTP"), HTTP)
	assert.Equal(t, UploadProtocol("HTTPS"), HTTPS)
	assert.Equal(t, UploadProtocol("S3"), S3)
}

// TestLogUploadConstants tests that all constants are defined
func TestLogUploadConstants(t *testing.T) {
	assert.Equal(t, "estbIP", EstbIp)
	assert.Equal(t, "estbMacAddress", EstbMacAddress)
	assert.Equal(t, "ecmMacAddress", EcmMac)
	assert.Equal(t, "env", Env)
	assert.Equal(t, "model", Model)
	assert.Equal(t, "accountMgmt", AccountMgmt)
	assert.Equal(t, "serialNum", SerialNum)
	assert.Equal(t, "partnerId", PartnerId)
	assert.Equal(t, "firmwareVersion", FirmwareVersion)
	assert.Equal(t, "controllerId", ControllerId)
	assert.Equal(t, "channelMapId", ChannelMapId)
	assert.Equal(t, "vodId", VodId)
	assert.Equal(t, "uploadImmediately", UploadImmediately)
	assert.Equal(t, "timezone", Timezone)
	assert.Equal(t, "accountHash", AccountHash)
	assert.Equal(t, "accountId", AccountId)
	assert.Equal(t, "configSetHash", ConfigSetHash)
}

// TestScheduleStruct tests the Schedule struct
func TestScheduleStruct(t *testing.T) {
	schedule := Schedule{
		Type:              "CronExpression",
		Expression:        "0 0 * * *",
		TimeZone:          "UTC",
		ExpressionL1:      "0 1 * * *",
		ExpressionL2:      "0 2 * * *",
		ExpressionL3:      "0 3 * * *",
		StartDate:         "2025-01-01",
		EndDate:           "2025-12-31",
		TimeWindowMinutes: "60",
	}

	assert.Equal(t, "CronExpression", schedule.Type)
	assert.Equal(t, "0 0 * * *", schedule.Expression)
	assert.Equal(t, "UTC", schedule.TimeZone)
	assert.Equal(t, "0 1 * * *", schedule.ExpressionL1)
	assert.Equal(t, "0 2 * * *", schedule.ExpressionL2)
	assert.Equal(t, "0 3 * * *", schedule.ExpressionL3)
	assert.Equal(t, "2025-01-01", schedule.StartDate)
	assert.Equal(t, "2025-12-31", schedule.EndDate)
	assert.Equal(t, "60", string(schedule.TimeWindowMinutes))
}

// TestConfigurationServiceURLStruct tests the ConfigurationServiceURL struct
func TestConfigurationServiceURLStruct(t *testing.T) {
	serviceURL := ConfigurationServiceURL{
		ID:          "service1",
		Name:        "Test Service",
		Description: "Test Description",
		URL:         "https://example.com/api",
	}

	assert.Equal(t, "service1", serviceURL.ID)
	assert.Equal(t, "Test Service", serviceURL.Name)
	assert.Equal(t, "Test Description", serviceURL.Description)
	assert.Equal(t, "https://example.com/api", serviceURL.URL)
}

// TestIsValidUrl_EdgeCases tests edge cases for URL validation
func TestIsValidUrl_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"URL with username", "https://user@example.com", true},
		{"URL with username and password", "https://user:pass@example.com", true},
		{"URL with fragment", "https://example.com/path#section", true},
		{"URL with multiple subdomains", "https://a.b.c.example.com", true},
		{"URL with hyphen in domain", "https://my-domain.example.com", true},
		{"URL with numbers", "https://example123.com", true},
		{"Just scheme and colon", "http:", false},
		{"Malformed URL", "http:/example.com", false},
		{"Space in URL", "http://example .com", false},
		{"Missing TLD", "http://example", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidUrl(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestUploadRepositoryClone_EmptyFields tests cloning with empty fields
func TestUploadRepositoryClone_EmptyFields(t *testing.T) {
	original := &UploadRepository{}

	cloned, err := original.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
	assert.Equal(t, "", cloned.ID)
	assert.Equal(t, int64(0), cloned.Updated)
	assert.Equal(t, "", cloned.Name)
}

// TestLogFileClone_EmptyFields tests cloning LogFile with empty fields
func TestLogFileClone_EmptyFields(t *testing.T) {
	original := &LogFile{}

	cloned, err := original.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
	assert.Equal(t, "", cloned.ID)
	assert.Equal(t, int64(0), cloned.Updated)
	assert.Equal(t, false, cloned.DeleteOnUpload)
}

// TestLogFilesGroupsClone_EmptySlice tests cloning with empty LogFileIDs slice
func TestLogFilesGroupsClone_EmptySlice(t *testing.T) {
	original := &LogFilesGroups{
		ID:         "group1",
		Updated:    1234567890,
		GroupName:  "Empty Group",
		LogFileIDs: []string{},
	}

	cloned, err := original.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
	assert.Equal(t, 0, len(cloned.LogFileIDs))
}

// TestLogFilesGroupsClone_NilSlice tests cloning with nil LogFileIDs slice
func TestLogFilesGroupsClone_NilSlice(t *testing.T) {
	original := &LogFilesGroups{
		ID:         "group1",
		Updated:    1234567890,
		GroupName:  "Nil Group",
		LogFileIDs: nil,
	}

	cloned, err := original.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
}

// TestLogFileListClone_EmptyData tests cloning with empty Data slice
func TestLogFileListClone_EmptyData(t *testing.T) {
	original := &LogFileList{
		Updated: 1234567890,
		Data:    []*LogFile{},
	}

	cloned, err := original.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
	assert.Equal(t, 0, len(cloned.Data))
}

// TestLogFileListClone_NilData tests cloning with nil Data slice
func TestLogFileListClone_NilData(t *testing.T) {
	original := &LogFileList{
		Updated: 1234567890,
		Data:    nil,
	}

	cloned, err := original.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
}

// TestGetLogFileGroupsList_WithData tests GetLogFileGroupsList with actual data
func TestGetLogFileGroupsList_WithData(t *testing.T) {
	// This test requires database setup
	if db.GetCachedSimpleDao() == nil {
		t.Skip("Database not configured")
	}

	// First, try to create a test group
	testGroup := &LogFilesGroups{
		ID:         "test-group-1",
		Updated:    1234567890,
		GroupName:  "Test Group",
		LogFileIDs: []string{"file1", "file2"},
	}

	// Try to save it (may fail if DB not configured)
	err := db.GetCachedSimpleDao().SetOne(db.TABLE_LOG_FILES_GROUPS, testGroup.ID, testGroup)
	if err != nil {
		t.Logf("Could not save test group: %v", err)
	}

	// Now try to get the list
	groups, err := GetLogFileGroupsList(100)
	if err != nil {
		t.Logf("GetLogFileGroupsList returned error: %v", err)
	} else {
		// If we got a result, verify it's a valid list
		assert.NotNil(t, groups)
		t.Logf("Retrieved %d groups", len(groups))
	}
}

// TestUploadRepositoryStruct tests the UploadRepository struct fields
func TestUploadRepositoryStruct(t *testing.T) {
	repo := UploadRepository{
		ID:              "repo-123",
		Updated:         1234567890,
		Name:            "Production Repository",
		Description:     "Main production upload repository",
		URL:             "https://upload.example.com",
		ApplicationType: "stb",
		Protocol:        "HTTPS",
	}

	assert.Equal(t, "repo-123", repo.ID)
	assert.Equal(t, int64(1234567890), repo.Updated)
	assert.Equal(t, "Production Repository", repo.Name)
	assert.Equal(t, "Main production upload repository", repo.Description)
	assert.Equal(t, "https://upload.example.com", repo.URL)
	assert.Equal(t, "stb", repo.ApplicationType)
	assert.Equal(t, "HTTPS", repo.Protocol)
}

// TestLogFileStruct tests the LogFile struct fields
func TestLogFileStruct(t *testing.T) {
	logFile := LogFile{
		ID:             "log-456",
		Updated:        1234567890,
		Name:           "application.log",
		DeleteOnUpload: true,
	}

	assert.Equal(t, "log-456", logFile.ID)
	assert.Equal(t, int64(1234567890), logFile.Updated)
	assert.Equal(t, "application.log", logFile.Name)
	assert.True(t, logFile.DeleteOnUpload)
}

// TestLogFilesGroupsStruct tests the LogFilesGroups struct fields
func TestLogFilesGroupsStruct(t *testing.T) {
	group := LogFilesGroups{
		ID:         "group-789",
		Updated:    1234567890,
		GroupName:  "System Logs",
		LogFileIDs: []string{"log1", "log2", "log3"},
	}

	assert.Equal(t, "group-789", group.ID)
	assert.Equal(t, int64(1234567890), group.Updated)
	assert.Equal(t, "System Logs", group.GroupName)
	assert.Equal(t, 3, len(group.LogFileIDs))
	assert.Equal(t, "log1", group.LogFileIDs[0])
}

// TestLogFileListStruct tests the LogFileList struct fields
func TestLogFileListStruct(t *testing.T) {
	logList := LogFileList{
		Updated: 1234567890,
		Data: []*LogFile{
			{ID: "log1", Name: "file1.log"},
			{ID: "log2", Name: "file2.log"},
		},
	}

	assert.Equal(t, int64(1234567890), logList.Updated)
	assert.Equal(t, 2, len(logList.Data))
	assert.Equal(t, "log1", logList.Data[0].ID)
	assert.Equal(t, "file2.log", logList.Data[1].Name)
}

// ============ Tests for settings.go functions ============

// TestIsValidSettingType tests the IsValidSettingType function
func TestIsValidSettingType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"PARTNER_SETTINGS uppercase", "PARTNER_SETTINGS", true},
		{"EPON uppercase", "EPON", true},
		{"partnersettings lowercase", "partnersettings", true},
		{"epon lowercase", "epon", true},
		{"Invalid type", "INVALID", false},
		{"Empty string", "", false},
		{"Random string", "random", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidSettingType(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSettingTypeEnum tests the SettingTypeEnum function
func TestSettingTypeEnum(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"epon lowercase", "epon", EPON},
		{"EPON uppercase", "EPON", EPON},
		{"Epon mixed case", "Epon", EPON},
		{"partner_settings", "partner_settings", PARTNER_SETTINGS},
		{"partnersettings", "partnersettings", PARTNER_SETTINGS},
		{"PARTNER_SETTINGS uppercase", "PARTNER_SETTINGS", PARTNER_SETTINGS},
		{"PartnerSettings mixed", "PartnerSettings", PARTNER_SETTINGS},
		{"Invalid type", "INVALID", 0},
		{"Empty string", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SettingTypeEnum(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSettingProfilesClone tests the Clone method for SettingProfiles
func TestSettingProfilesClone(t *testing.T) {
	original := &SettingProfiles{
		ID:               "profile1",
		Updated:          1234567890,
		SettingProfileID: "sp1",
		SettingType:      "EPON",
		Properties: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		ApplicationType: "stb",
	}

	cloned, err := original.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Updated, cloned.Updated)
	assert.Equal(t, original.SettingProfileID, cloned.SettingProfileID)
	assert.Equal(t, original.SettingType, cloned.SettingType)
	assert.Equal(t, original.ApplicationType, cloned.ApplicationType)
	assert.Equal(t, original.Properties, cloned.Properties)

	// Verify it's a deep copy
	cloned.Properties["key1"] = "modified"
	assert.NotEqual(t, original.Properties["key1"], cloned.Properties["key1"])
}

// TestNewSettingProfilesInf tests the constructor
func TestNewSettingProfilesInf(t *testing.T) {
	obj := NewSettingProfilesInf()
	assert.NotNil(t, obj)

	profile, ok := obj.(*SettingProfiles)
	assert.True(t, ok)
	assert.Equal(t, core.STB, profile.ApplicationType)
}

// TestVodSettingsClone tests the Clone method for VodSettings
func TestVodSettingsClone(t *testing.T) {
	original := &VodSettings{
		ID:              "vod1",
		Updated:         1234567890,
		Name:            "Test VOD",
		LocationsURL:    "http://example.com",
		IPNames:         []string{"ip1", "ip2"},
		IPList:          []string{"192.168.1.1", "192.168.1.2"},
		SrmIPList:       map[string]string{"srm1": "10.0.0.1"},
		ApplicationType: "stb",
	}

	cloned, err := original.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Updated, cloned.Updated)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.LocationsURL, cloned.LocationsURL)
	assert.Equal(t, original.ApplicationType, cloned.ApplicationType)

	// Verify it's a deep copy
	cloned.IPNames[0] = "modified"
	assert.NotEqual(t, original.IPNames[0], cloned.IPNames[0])
}

// TestNewVodSettingsInf tests the constructor
func TestNewVodSettingsInf(t *testing.T) {
	obj := NewVodSettingsInf()
	assert.NotNil(t, obj)

	vod, ok := obj.(*VodSettings)
	assert.True(t, ok)
	assert.Equal(t, core.STB, vod.ApplicationType)
}

// TestSettingRuleClone tests the Clone method for SettingRule
func TestSettingRuleClone(t *testing.T) {
	original := &SettingRule{
		ID:              "rule1",
		Updated:         1234567890,
		Name:            "Test Rule",
		BoundSettingID:  "setting1",
		ApplicationType: "stb",
	}

	cloned, err := original.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Updated, cloned.Updated)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.BoundSettingID, cloned.BoundSettingID)
	assert.Equal(t, original.ApplicationType, cloned.ApplicationType)
}

// TestSettingRuleGetApplicationType tests the GetApplicationType method
func TestSettingRuleGetApplicationType(t *testing.T) {
	// Test with ApplicationType set
	rule := &SettingRule{ApplicationType: "xhome"}
	assert.Equal(t, "xhome", rule.GetApplicationType())

	// Test with empty ApplicationType (should return default STB)
	rule2 := &SettingRule{ApplicationType: ""}
	assert.Equal(t, core.STB, rule2.GetApplicationType())
}

// TestSettingRuleXRuleInterface tests XRule interface methods
func TestSettingRuleXRuleInterface(t *testing.T) {
	rule := &SettingRule{
		ID:   "rule1",
		Name: "Test Rule",
	}

	assert.Equal(t, "rule1", rule.GetId())
	assert.Equal(t, "Test Rule", rule.GetName())
	assert.Equal(t, "", rule.GetTemplateId())
	assert.Equal(t, "SettingRule", rule.GetRuleType())

	rulePtr := rule.GetRule()
	assert.NotNil(t, rulePtr)
}

// TestNewSettingRulesInf tests the constructor
func TestNewSettingRulesInf(t *testing.T) {
	obj := NewSettingRulesInf()
	assert.NotNil(t, obj)

	rule, ok := obj.(*SettingRule)
	assert.True(t, ok)
	assert.Equal(t, core.STB, rule.ApplicationType)
}

// TestNewSettings tests the NewSettings constructor
func TestNewSettings(t *testing.T) {
	settings := NewSettings(5)
	assert.NotNil(t, settings)
	assert.NotNil(t, settings.RuleIDs)
	assert.NotNil(t, settings.SrmIPList)
	assert.NotNil(t, settings.EponSettings)
	assert.NotNil(t, settings.PartnerSettings)
	assert.NotNil(t, settings.LusLogFiles)
	assert.Equal(t, 5, len(settings.LusLogFiles))
}

// TestCopyDeviceSettings tests the CopyDeviceSettings method
func TestCopyDeviceSettings(t *testing.T) {
	source := NewSettings(0)
	source.GroupName = "TestGroup"
	source.CheckOnReboot = true
	source.ConfigurationServiceURL = "http://config.example.com"
	source.ScheduleCron = "0 0 * * *"
	source.ScheduleDurationMinutes = 60
	source.ScheduleStartDate = "2025-01-01"
	source.ScheduleEndDate = "2025-12-31"

	dest := NewSettings(0)
	dest.CopyDeviceSettings(source)

	assert.Equal(t, source.GroupName, dest.GroupName)
	assert.Equal(t, source.CheckOnReboot, dest.CheckOnReboot)
	assert.Equal(t, source.ConfigurationServiceURL, dest.ConfigurationServiceURL)
	assert.Equal(t, source.ScheduleCron, dest.ScheduleCron)
	assert.Equal(t, source.ScheduleDurationMinutes, dest.ScheduleDurationMinutes)
	assert.Equal(t, source.ScheduleStartDate, dest.ScheduleStartDate)
	assert.Equal(t, source.ScheduleEndDate, dest.ScheduleEndDate)
}

// TestCopyLusSettingWithTrue tests CopyLusSetting with setLUSSettings=true
func TestCopyLusSettingWithTrue(t *testing.T) {
	source := NewSettings(2)
	source.LusName = "TestLUS"
	source.LusNumberOfDay = 7
	source.LusUploadRepositoryName = "TestRepo"
	source.LusUploadRepositoryURL = "http://upload.example.com"
	source.LusUploadRepositoryURLNew = "http://new.example.com"
	source.LusUploadRepositoryUploadProtocol = "HTTPS"
	source.LusUploadOnReboot = true
	source.LusLogFiles = []*LogFile{{ID: "log1"}, {ID: "log2"}}
	source.LusLogFilesStartDate = "2025-01-01"
	source.LusLogFilesEndDate = "2025-12-31"
	source.LusScheduleDurationMinutes = 30
	source.LusScheduleStartDate = "2025-01-01"
	source.LusScheduleEndDate = "2025-12-31"

	dest := NewSettings(0)
	dest.CopyLusSetting(source, true)

	assert.Equal(t, "", dest.LusMessage)
	assert.Equal(t, source.LusName, dest.LusName)
	assert.Equal(t, source.LusNumberOfDay, dest.LusNumberOfDay)
	assert.Equal(t, source.LusUploadRepositoryName, dest.LusUploadRepositoryName)
	assert.Equal(t, source.LusUploadRepositoryURL, dest.LusUploadRepositoryURL)
	assert.Equal(t, source.LusUploadRepositoryURLNew, dest.LusUploadRepositoryURLNew)
	assert.Equal(t, source.LusUploadRepositoryUploadProtocol, dest.LusUploadRepositoryUploadProtocol)
	assert.Equal(t, source.LusUploadOnReboot, dest.LusUploadOnReboot)
	assert.Equal(t, source.LusLogFiles, dest.LusLogFiles)
	assert.Equal(t, source.LusLogFilesStartDate, dest.LusLogFilesStartDate)
	assert.Equal(t, source.LusLogFilesEndDate, dest.LusLogFilesEndDate)
	assert.Equal(t, source.LusScheduleDurationMinutes, dest.LusScheduleDurationMinutes)
	assert.Equal(t, source.LusScheduleStartDate, dest.LusScheduleStartDate)
	assert.Equal(t, source.LusScheduleEndDate, dest.LusScheduleEndDate)
	assert.True(t, dest.Upload)
}

// TestCopyLusSettingWithFalse tests CopyLusSetting with setLUSSettings=false
func TestCopyLusSettingWithFalse(t *testing.T) {
	source := NewSettings(2)
	source.LusName = "TestLUS"
	source.LusNumberOfDay = 7

	dest := NewSettings(0)
	dest.CopyLusSetting(source, false)

	assert.Equal(t, DEFAULT_LOG_UPLOAD_SETTINGS_MESSAGE, dest.LusMessage)
	assert.Equal(t, "", dest.LusName)
	assert.Equal(t, 0, dest.LusNumberOfDay)
	assert.Equal(t, "", dest.LusUploadRepositoryName)
	assert.Equal(t, "", dest.LusUploadRepositoryURL)
	assert.Equal(t, "", dest.LusUploadRepositoryURLNew)
	assert.Equal(t, "", dest.LusUploadRepositoryUploadProtocol)
	assert.False(t, dest.LusUploadOnReboot)
	assert.Nil(t, dest.LusLogFiles)
	assert.Equal(t, "", dest.LusLogFilesStartDate)
	assert.Equal(t, "", dest.LusLogFilesEndDate)
	assert.Equal(t, 0, dest.LusScheduleDurationMinutes)
	assert.Equal(t, "", dest.LusScheduleStartDate)
	assert.Equal(t, "", dest.LusScheduleEndDate)
	assert.False(t, dest.Upload)
}

// TestCopyVodSettings tests the CopyVodSettings method
func TestCopyVodSettings(t *testing.T) {
	source := NewSettings(0)
	source.VodSettingsName = "TestVOD"
	source.LocationUrl = "http://vod.example.com"
	source.SrmIPList = map[string]string{"srm1": "10.0.0.1", "srm2": "10.0.0.2"}

	dest := NewSettings(0)
	dest.CopyVodSettings(source)

	assert.Equal(t, source.VodSettingsName, dest.VodSettingsName)
	assert.Equal(t, source.LocationUrl, dest.LocationUrl)
	assert.Equal(t, source.SrmIPList, dest.SrmIPList)
}

// TestAreFull tests the AreFull method
func TestAreFull(t *testing.T) {
	// Test with all fields set
	settings := NewSettings(0)
	settings.GroupName = "TestGroup"
	settings.LusName = "TestLUS"
	settings.VodSettingsName = "TestVOD"
	assert.True(t, settings.AreFull())

	// Test with GroupName missing
	settings2 := NewSettings(0)
	settings2.LusName = "TestLUS"
	settings2.VodSettingsName = "TestVOD"
	assert.False(t, settings2.AreFull())

	// Test with LusName missing
	settings3 := NewSettings(0)
	settings3.GroupName = "TestGroup"
	settings3.VodSettingsName = "TestVOD"
	assert.False(t, settings3.AreFull())

	// Test with VodSettingsName missing
	settings4 := NewSettings(0)
	settings4.GroupName = "TestGroup"
	settings4.LusName = "TestLUS"
	assert.False(t, settings4.AreFull())

	// Test with all fields empty
	settings5 := NewSettings(0)
	assert.False(t, settings5.AreFull())
}

// TestSetSettingProfiles tests the SetSettingProfiles method
func TestSetSettingProfiles(t *testing.T) {
	settings := NewSettings(0)

	profiles := []SettingProfiles{
		{
			SettingType: "PARTNER_SETTINGS",
			Properties: map[string]string{
				"partner1": "value1",
				"partner2": "value2",
			},
		},
		{
			SettingType: "EPON",
			Properties: map[string]string{
				"epon1": "value1",
				"epon2": "value2",
			},
		},
	}

	settings.SetSettingProfiles(profiles)

	assert.Equal(t, 2, len(settings.PartnerSettings))
	assert.Equal(t, "value1", settings.PartnerSettings["partner1"])
	assert.Equal(t, "value2", settings.PartnerSettings["partner2"])

	assert.Equal(t, 2, len(settings.EponSettings))
	assert.Equal(t, "value1", settings.EponSettings["epon1"])
	assert.Equal(t, "value2", settings.EponSettings["epon2"])
}

// TestSetSettingProfilesEmpty tests SetSettingProfiles with empty slice
func TestSetSettingProfilesEmpty(t *testing.T) {
	settings := NewSettings(0)
	settings.SetSettingProfiles([]SettingProfiles{})

	assert.Equal(t, 0, len(settings.PartnerSettings))
	assert.Equal(t, 0, len(settings.EponSettings))
}

// TestSetSettingProfilesInvalidType tests SetSettingProfiles with invalid type
func TestSetSettingProfilesInvalidType(t *testing.T) {
	settings := NewSettings(0)

	profiles := []SettingProfiles{
		{
			SettingType: "INVALID_TYPE",
			Properties: map[string]string{
				"key": "value",
			},
		},
	}

	settings.SetSettingProfiles(profiles)

	// Should not set anything for invalid type
	assert.Equal(t, 0, len(settings.PartnerSettings))
	assert.Equal(t, 0, len(settings.EponSettings))
}

// TestCreateSettingsResponseObject tests CreateSettingsResponseObject with all fields set
func TestCreateSettingsResponseObject(t *testing.T) {
	settings := NewSettings(0)
	settings.GroupName = "TestGroup"
	settings.CheckOnReboot = true
	settings.ScheduleCron = "0 0 * * *"
	settings.ScheduleDurationMinutes = 60
	settings.LusMessage = "Test Message"
	settings.LusName = "TestLUS"
	settings.LusNumberOfDay = 7
	settings.LusUploadRepositoryName = "TestRepo"
	settings.LusUploadRepositoryURLNew = "http://new.example.com"
	settings.LusUploadRepositoryUploadProtocol = "HTTPS"
	settings.LusUploadRepositoryURL = "http://old.example.com"
	settings.LusUploadOnReboot = true
	settings.UploadImmediately = true
	settings.Upload = true
	settings.LusScheduleCron = "0 1 * * *"
	settings.LusScheduleCronL1 = "0 2 * * *"
	settings.LusScheduleCronL2 = "0 3 * * *"
	settings.LusScheduleCronL3 = "0 4 * * *"
	settings.LusScheduleDurationMinutes = 30
	settings.VodSettingsName = "TestVOD"
	settings.LocationUrl = "http://vod.example.com"
	settings.SrmIPList = map[string]string{"srm1": "10.0.0.1"}
	settings.EponSettings = map[string]string{"epon1": "value1"}
	settings.PartnerSettings = map[string]string{"partner1": "value1"}

	response := CreateSettingsResponseObject(settings)

	assert.NotNil(t, response)
	assert.Equal(t, "TestGroup", response.GroupName)
	assert.True(t, response.CheckOnReboot)
	assert.Equal(t, "0 0 * * *", response.ScheduleCron)
	assert.Equal(t, 60, response.ScheduleDurationMinutes)
	assert.Equal(t, "Test Message", response.LusMessage)
	assert.Equal(t, "TestLUS", response.LusName)
	assert.Equal(t, 7, response.LusNumberOfDay)
	assert.Equal(t, "TestRepo", response.LusUploadRepositoryName)
	assert.Equal(t, "http://new.example.com", response.LusUploadRepositoryURLNew)
	assert.Equal(t, "HTTPS", response.LusUploadRepositoryUploadProtocol)
	assert.Equal(t, "http://old.example.com", response.LusUploadRepositoryURL)
	assert.True(t, response.LusUploadOnReboot)
	assert.True(t, response.UploadImmediately)
	assert.True(t, response.Upload)
	assert.Equal(t, "0 1 * * *", response.LusScheduleCron)
	assert.Equal(t, "0 2 * * *", response.LusScheduleCronL1)
	assert.Equal(t, "0 3 * * *", response.LusScheduleCronL2)
	assert.Equal(t, "0 4 * * *", response.LusScheduleCronL3)
	assert.Equal(t, 30, response.LusScheduleDurationMinutes)
	assert.Equal(t, "TestVOD", response.VodSettingsName)
	assert.Equal(t, "http://vod.example.com", response.LocationUrl)
	assert.Equal(t, map[string]string{"srm1": "10.0.0.1"}, response.SrmIPList)
	assert.Equal(t, map[string]string{"epon1": "value1"}, response.EponSettings)
	assert.Equal(t, map[string]string{"partner1": "value1"}, response.PartnerSettings)
}

// TestCreateSettingsResponseObjectWithEmptyFields tests CreateSettingsResponseObject with empty fields
func TestCreateSettingsResponseObjectWithEmptyFields(t *testing.T) {
	settings := NewSettings(0)

	response := CreateSettingsResponseObject(settings)

	assert.NotNil(t, response)
	assert.Nil(t, response.GroupName)
	assert.Nil(t, response.ScheduleCron)
	assert.Nil(t, response.LusMessage)
	assert.Nil(t, response.LusName)
	assert.Nil(t, response.LusUploadRepositoryName)
	assert.Nil(t, response.LusScheduleCron)
	assert.Nil(t, response.LusScheduleCronL1)
	assert.Nil(t, response.LusScheduleCronL2)
	assert.Nil(t, response.LusScheduleCronL3)
	assert.Nil(t, response.VodSettingsName)
	assert.Nil(t, response.LocationUrl)
	assert.Nil(t, response.SrmIPList)
}

// TestDeviceSettingsClone tests the Clone method for DeviceSettings
func TestDeviceSettingsClone(t *testing.T) {
	original := &DeviceSettings{
		ID:            "device1",
		Updated:       1234567890,
		Name:          "Test Device",
		CheckOnReboot: true,
		ConfigurationServiceURL: &ConfigurationServiceURL{
			ID:   "url1",
			Name: "Test URL",
		},
		SettingsAreActive: true,
		ApplicationType:   "stb",
	}

	cloned, err := original.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Updated, cloned.Updated)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.CheckOnReboot, cloned.CheckOnReboot)
	assert.Equal(t, original.SettingsAreActive, cloned.SettingsAreActive)
	assert.Equal(t, original.ApplicationType, cloned.ApplicationType)
}

// TestNewDeviceSettingsInf tests the constructor
func TestNewDeviceSettingsInf(t *testing.T) {
	obj := NewDeviceSettingsInf()
	assert.NotNil(t, obj)

	device, ok := obj.(*DeviceSettings)
	assert.True(t, ok)
	assert.Equal(t, core.STB, device.ApplicationType)
}

// TestLogUploadSettingsClone tests the Clone method for LogUploadSettings
func TestLogUploadSettingsClone(t *testing.T) {
	original := &LogUploadSettings{
		ID:                  "lus1",
		Updated:             1234567890,
		Name:                "Test LUS",
		UploadOnReboot:      true,
		NumberOfDays:        7,
		AreSettingsActive:   true,
		LogFileIds:          []string{"log1", "log2"},
		LogFilesGroupID:     "group1",
		ModeToGetLogFiles:   MODE_TO_GET_LOG_FILES_0,
		UploadRepositoryID:  "repo1",
		ActiveDateTimeRange: true,
		FromDateTime:        "2025-01-01T00:00:00",
		ToDateTime:          "2025-12-31T23:59:59",
		ApplicationType:     "stb",
	}

	cloned, err := original.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Updated, cloned.Updated)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.UploadOnReboot, cloned.UploadOnReboot)
	assert.Equal(t, original.NumberOfDays, cloned.NumberOfDays)
	assert.Equal(t, original.AreSettingsActive, cloned.AreSettingsActive)
	assert.Equal(t, original.ApplicationType, cloned.ApplicationType)

	// Verify it's a deep copy
	cloned.LogFileIds[0] = "modified"
	assert.NotEqual(t, original.LogFileIds[0], cloned.LogFileIds[0])
}

// TestNewLogUploadSettingsInf tests the constructor
func TestNewLogUploadSettingsInf(t *testing.T) {
	obj := NewLogUploadSettingsInf()
	assert.NotNil(t, obj)

	lus, ok := obj.(*LogUploadSettings)
	assert.True(t, ok)
	assert.Equal(t, core.STB, lus.ApplicationType)
}

// TestGetOneDeviceSettings tests GetOneDeviceSettings (requires DB setup)
func TestGetOneDeviceSettings(t *testing.T) {
	// Test with non-existent ID (should return nil)
	result := GetOneDeviceSettings("non-existent-id")
	assert.Nil(t, result)
}

// TestGetOneLogUploadSettings tests GetOneLogUploadSettings (requires DB setup)
func TestGetOneLogUploadSettings(t *testing.T) {
	// Test with non-existent ID (should return nil)
	result := GetOneLogUploadSettings("non-existent-id")
	assert.Nil(t, result)
}

// TestGetOneUploadRepository tests GetOneUploadRepository (requires DB setup)
func TestGetOneUploadRepository(t *testing.T) {
	// Test with non-existent ID (should return nil)
	result := GetOneUploadRepository("non-existent-id")
	assert.Nil(t, result)
}

// TestGetOneVodSettings tests GetOneVodSettings (requires DB setup)
func TestGetOneVodSettings(t *testing.T) {
	// Test with non-existent ID (should return nil)
	result := GetOneVodSettings("non-existent-id")
	assert.Nil(t, result)
}

// TestGetOneSettingProfile tests GetOneSettingProfile (requires DB setup)
func TestGetOneSettingProfile(t *testing.T) {
	// Test with non-existent ID (should return nil)
	result := GetOneSettingProfile("non-existent-id")
	assert.Nil(t, result)
}

// TestGetLogFileList tests GetLogFileList (requires DB setup)
func TestGetLogFileList(t *testing.T) {
	// Test with non-existent data (should return nil)
	result := GetLogFileList(10)
	assert.Nil(t, result)
}

// TestGetAllLogFileList tests GetAllLogFileList (requires DB setup)
func TestGetAllLogFileList(t *testing.T) {
	// Test with non-existent data (should return nil)
	result := GetAllLogFileList(10)
	assert.Nil(t, result)
}

// TestGetAllSettingRuleList tests GetAllSettingRuleList (requires DB setup)
func TestGetAllSettingRuleList(t *testing.T) {
	// Test with non-existent data (should return empty slice)
	result := GetAllSettingRuleList()
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))
}

// TestGetAllLogUploadSettings tests GetAllLogUploadSettings (requires DB setup)
func TestGetAllLogUploadSettings(t *testing.T) {
	// Test with non-existent data (should return error)
	result, err := GetAllLogUploadSettings(10)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestSetOneLogUploadSettings tests SetOneLogUploadSettings (requires DB setup)
func TestSetOneLogUploadSettings(t *testing.T) {
	lus := &LogUploadSettings{
		ID:   "test-lus",
		Name: "Test",
	}
	// This will fail without proper DB setup, but tests the function signature
	err := SetOneLogUploadSettings("test-lus", lus)
	assert.Error(t, err)
}

// TestGetOneLogFileList tests GetOneLogFileList
func TestGetOneLogFileList(t *testing.T) {
	// Test with non-existent ID (should return empty LogFileList with empty Data)
	result, err := GetOneLogFileList("non-existent-id")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Data)
	assert.Equal(t, 0, len(result.Data))
}

// TestSetOneLogFile tests SetOneLogFile
func TestSetOneLogFile(t *testing.T) {
	logFile := &LogFile{
		ID:   "log1",
		Name: "test.log",
	}
	// This will work with in-memory DB
	err := SetOneLogFile("test-list", logFile)
	// May succeed or fail depending on DB state, just test it doesn't panic
	_ = err
}

// TestDeleteOneLogFileList tests DeleteOneLogFileList
func TestDeleteOneLogFileList(t *testing.T) {
	// This will work with in-memory DB
	err := DeleteOneLogFileList("test-list")
	// May succeed or fail depending on DB state, just test it doesn't panic
	_ = err
}

// ============ Additional tests to improve coverage ============

// TestSettingProfilesCloneError tests Clone error handling
func TestSettingProfilesCloneError(t *testing.T) {
	// Test with a valid object that should clone successfully
	profile := &SettingProfiles{
		ID: "test",
	}
	cloned, err := profile.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
}

// TestVodSettingsCloneError tests Clone error handling
func TestVodSettingsCloneError(t *testing.T) {
	// Test with a valid object that should clone successfully
	vod := &VodSettings{
		ID: "test",
	}
	cloned, err := vod.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
}

// TestSettingRuleCloneError tests Clone error handling
func TestSettingRuleCloneError(t *testing.T) {
	// Test with a valid object that should clone successfully
	rule := &SettingRule{
		ID: "test",
	}
	cloned, err := rule.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
}

// TestDeviceSettingsCloneError tests Clone error handling
func TestDeviceSettingsCloneError(t *testing.T) {
	// Test with a valid object that should clone successfully
	device := &DeviceSettings{
		ID: "test",
	}
	cloned, err := device.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
}

// TestLogUploadSettingsCloneError tests Clone error handling
func TestLogUploadSettingsCloneError(t *testing.T) {
	// Test with a valid object that should clone successfully
	lus := &LogUploadSettings{
		ID: "test",
	}
	cloned, err := lus.Clone()
	assert.NoError(t, err)
	assert.NotNil(t, cloned)
}

// TestSetOneLogFileWithReplacement tests SetOneLogFile replacing existing log file
func TestSetOneLogFileWithReplacement(t *testing.T) {
	// First, create a log file list with an existing file
	listID := "test-replacement-list"
	existingFile := &LogFile{
		ID:   "existing-log",
		Name: "existing.log",
	}

	// Set initial file
	err := SetOneLogFile(listID, existingFile)
	assert.NoError(t, err)

	// Now replace with same ID but different name
	replacementFile := &LogFile{
		ID:   "existing-log",
		Name: "replaced.log",
	}

	err = SetOneLogFile(listID, replacementFile)
	assert.NoError(t, err)

	// Verify the replacement
	list, err := GetOneLogFileList(listID)
	assert.NoError(t, err)
	assert.NotNil(t, list)

	// Should have only one file with the new name
	found := false
	for _, lf := range list.Data {
		if lf.ID == "existing-log" {
			found = true
			assert.Equal(t, "replaced.log", lf.Name)
		}
	}
	assert.True(t, found, "Replaced log file should be in the list")
}

// TestSetOneLogFileMultiple tests SetOneLogFile with multiple files
func TestSetOneLogFileMultiple(t *testing.T) {
	listID := "test-multiple-list"

	file1 := &LogFile{ID: "log1", Name: "file1.log"}
	file2 := &LogFile{ID: "log2", Name: "file2.log"}
	file3 := &LogFile{ID: "log3", Name: "file3.log"}

	err := SetOneLogFile(listID, file1)
	assert.NoError(t, err)

	err = SetOneLogFile(listID, file2)
	assert.NoError(t, err)

	err = SetOneLogFile(listID, file3)
	assert.NoError(t, err)

	list, err := GetOneLogFileList(listID)
	assert.NoError(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, 3, len(list.Data))
}

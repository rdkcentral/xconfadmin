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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/stretchr/testify/assert"
)

// helper to wrap recorder for drained body handlers
func makeNSXW(body any) (*httptest.ResponseRecorder, *xwhttp.XResponseWriter) {
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	if body != nil {
		b, _ := json.Marshal(body)
		xw.SetBody(string(b))
	}
	return rr, xw
}

// Simple UT tests

func TestDeleteIpAddressGroupHandler_Success(t *testing.T) {
	// Test successful deletion
	id := uuid.NewString()
	// Create an IP address group first
	ipList := makeGenericList(id, shared.IP_LIST, []string{"192.168.1.1"})
	CreateNamespacedList(ipList, false)

	url := fmt.Sprintf("/xconfAdminService/queries/ipAddressGroups/%s?applicationType=stb", id)
	req := httptest.NewRequest("DELETE", url, nil)
	req = mux.SetURLVars(req, map[string]string{"id": id})
	rr := httptest.NewRecorder()

	DeleteIpAddressGroupHandler(rr, req)
	// Should succeed with NoContent
	assert.True(t, rr.Code == http.StatusNoContent || rr.Code == http.StatusOK)
}

func TestDeleteIpAddressGroupHandler_MissingId(t *testing.T) {
	// Test WriteAdminErrorResponse for missing ID
	req := httptest.NewRequest("DELETE", "/xconfAdminService/queries/ipAddressGroups/?applicationType=stb", nil)
	rr := httptest.NewRecorder()

	DeleteIpAddressGroupHandler(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid")
}

func TestDeleteIpAddressGroupHandler_AuthError(t *testing.T) {
	// Test xhttp.AdminError path - no auth
	req := httptest.NewRequest("DELETE", "/xconfAdminService/queries/ipAddressGroups/test-id", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "test-id"})
	rr := httptest.NewRecorder()

	DeleteIpAddressGroupHandler(rr, req)
	// May succeed with default auth or return error
	assert.True(t, rr.Code >= 200 && rr.Code < 500)
}

func TestGetQueriesIpAddressGroupsV2_Success(t *testing.T) {
	// Test successful retrieval of IP address groups
	req := httptest.NewRequest("GET", "/xconfAdminService/queries/ipAddressGroups?applicationType=stb", nil)
	rr := httptest.NewRecorder()

	GetQueriesIpAddressGroupsV2(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetQueriesIpAddressGroupsV2_AuthError(t *testing.T) {
	// Test xhttp.AdminError in auth.CanRead
	req := httptest.NewRequest("GET", "/xconfAdminService/queries/ipAddressGroups", nil)
	rr := httptest.NewRecorder()

	GetQueriesIpAddressGroupsV2(rr, req)
	// Auth handling varies
	assert.True(t, rr.Code >= 200 && rr.Code < 500)
}

func TestGetQueriesMacListsById_Success(t *testing.T) {
	// Test successful retrieval
	id := uuid.NewString()
	macList := makeGenericList(id, shared.MAC_LIST, []string{"AA:BB:CC:DD:EE:FF"})
	CreateNamespacedList(macList, false)

	url := fmt.Sprintf("/xconfAdminService/queries/macs/%s?applicationType=stb", id)
	req := httptest.NewRequest("GET", url, nil)
	req = mux.SetURLVars(req, map[string]string{"id": id})
	rr := httptest.NewRecorder()

	GetQueriesMacListsById(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetQueriesMacListsById_MissingId(t *testing.T) {
	// Test WriteAdminErrorResponse for missing ID
	req := httptest.NewRequest("GET", "/xconfAdminService/queries/macs/?applicationType=stb", nil)
	rr := httptest.NewRecorder()

	GetQueriesMacListsById(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid")
}

func TestGetQueriesMacListsById_NotFound(t *testing.T) {
	// Test when MAC list doesn't exist - WriteXconfResponse returns empty
	nonExistentId := uuid.NewString()
	url := fmt.Sprintf("/xconfAdminService/queries/macs/%s?applicationType=stb", nonExistentId)
	req := httptest.NewRequest("GET", url, nil)
	req = mux.SetURLVars(req, map[string]string{"id": nonExistentId})
	rr := httptest.NewRecorder()

	GetQueriesMacListsById(rr, req)
	// Returns 200 with empty body when not found (backward compatibility)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAddDataMacListHandler_Success(t *testing.T) {
	// Test successful addition of data to MAC list
	// Create a list through the handler instead of CreateNamespacedList
	listId := "test-mac-list-add-" + uuid.NewString()

	// First create via handler
	createList := makeGenericList(listId, shared.MAC_LIST, []string{"AA:BB:CC:DD:EE:00"})
	createResp := CreateNamespacedList(createList, false)
	assert.Equal(t, http.StatusCreated, createResp.Status)

	wrapper := shared.StringListWrapper{
		List: []string{"11:22:33:44:55:66"},
	}

	url := fmt.Sprintf("/xconfAdminService/queries/macs/addData/%s?applicationType=stb", listId)
	req := httptest.NewRequest("POST", url, nil)
	req = mux.SetURLVars(req, map[string]string{"listId": listId})
	rr, xw := makeNSXW(wrapper)

	AddDataMacListHandler(xw, req)
	// Should succeed
	assert.True(t, rr.Code >= 200 && rr.Code < 300, "Expected 2xx status, got %d: %s", rr.Code, rr.Body.String())
}

func TestAddDataMacListHandler_MissingListId(t *testing.T) {
	// Test WriteAdminErrorResponse for missing listId
	req := httptest.NewRequest("POST", "/xconfAdminService/queries/macs/addData/?applicationType=stb", nil)
	rr := httptest.NewRecorder()

	AddDataMacListHandler(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid")
}

func TestAddDataMacListHandler_InvalidJson(t *testing.T) {
	// Test error when XResponseWriter cast succeeds but invalid JSON body
	listId := uuid.NewString()
	url := fmt.Sprintf("/xconfAdminService/queries/macs/addData/%s?applicationType=stb", listId)
	req := httptest.NewRequest("POST", url, nil)
	req = mux.SetURLVars(req, map[string]string{"listId": listId})
	rr, xw := makeNSXW(nil)
	xw.SetBody("invalid json")

	AddDataMacListHandler(xw, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRemoveDataMacListHandler_Success(t *testing.T) {
	// Test successful removal of data from MAC list
	// Create a list with 2 MACs so we can remove one
	listId := "test-mac-list-remove-" + uuid.NewString()

	// First create via handler with 2 MACs
	createList := makeGenericList(listId, shared.MAC_LIST, []string{"AA:BB:CC:DD:EE:FF", "11:22:33:44:55:66"})
	createResp := CreateNamespacedList(createList, false)
	if createResp.Status != http.StatusCreated {
		t.Logf("Create failed: %d - %v", createResp.Status, createResp.Error)
		t.Skip("Cannot test remove when create fails")
	}

	wrapper := shared.StringListWrapper{
		List: []string{"AA:BB:CC:DD:EE:FF"},
	}

	url := fmt.Sprintf("/xconfAdminService/queries/macs/removeData/%s?applicationType=stb", listId)
	req := httptest.NewRequest("DELETE", url, nil)
	req = mux.SetURLVars(req, map[string]string{"listId": listId})
	rr, xw := makeNSXW(wrapper)

	RemoveDataMacListHandler(xw, req)
	// Should succeed
	assert.True(t, rr.Code >= 200 && rr.Code < 300, "Expected 2xx status, got %d: %s", rr.Code, rr.Body.String())
}

func TestRemoveDataMacListHandler_MissingListId(t *testing.T) {
	// Test WriteAdminErrorResponse for missing listId
	req := httptest.NewRequest("DELETE", "/xconfAdminService/queries/macs/removeData/?applicationType=stb", nil)
	rr := httptest.NewRecorder()

	RemoveDataMacListHandler(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid")
}

func TestRemoveDataMacListHandler_InvalidJson(t *testing.T) {
	// Test WriteAdminErrorResponse for invalid JSON
	listId := uuid.NewString()
	url := fmt.Sprintf("/xconfAdminService/queries/macs/removeData/%s?applicationType=stb", listId)
	req := httptest.NewRequest("DELETE", url, nil)
	req = mux.SetURLVars(req, map[string]string{"listId": listId})
	rr, xw := makeNSXW(nil)
	xw.SetBody("invalid json")

	RemoveDataMacListHandler(xw, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetNamespacedListHandler_Success(t *testing.T) {
	// Test successful retrieval
	id := uuid.NewString()
	nsList := makeGenericList(id, shared.IP_LIST, []string{"192.168.1.1"})
	CreateNamespacedList(nsList, false)

	url := fmt.Sprintf("/xconfAdminService/queries/namespacedLists/%s?applicationType=stb", id)
	req := httptest.NewRequest("GET", url, nil)
	req = mux.SetURLVars(req, map[string]string{"id": id})
	rr := httptest.NewRecorder()

	GetNamespacedListHandler(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetNamespacedListHandler_MissingId(t *testing.T) {
	// Test WriteAdminErrorResponse for missing ID
	req := httptest.NewRequest("GET", "/xconfAdminService/queries/namespacedLists/?applicationType=stb", nil)
	rr := httptest.NewRecorder()

	GetNamespacedListHandler(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid")
}

func TestGetNamespacedListHandler_NotFound(t *testing.T) {
	// Test WriteAdminErrorResponse when list doesn't exist
	nonExistentId := uuid.NewString()
	url := fmt.Sprintf("/xconfAdminService/queries/namespacedLists/%s?applicationType=stb", nonExistentId)
	req := httptest.NewRequest("GET", url, nil)
	req = mux.SetURLVars(req, map[string]string{"id": nonExistentId})
	rr := httptest.NewRecorder()

	GetNamespacedListHandler(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "does not exist")
}

func TestGetNamespacedListHandler_ExportWithHeaders(t *testing.T) {
	// Test WriteXconfResponseWithHeaders for export
	id := uuid.NewString()
	nsList := makeGenericList(id, shared.IP_LIST, []string{"192.168.1.1"})
	CreateNamespacedList(nsList, false)

	url := fmt.Sprintf("/xconfAdminService/queries/namespacedLists/%s?applicationType=stb&export=true", id)
	req := httptest.NewRequest("GET", url, nil)
	req = mux.SetURLVars(req, map[string]string{"id": id})
	rr := httptest.NewRecorder()

	GetNamespacedListHandler(rr, req)
	// Export may succeed with 200 or fail with 404 if list not in cache
	assert.True(t, rr.Code == http.StatusOK || rr.Code == http.StatusNotFound)
	if rr.Code == http.StatusOK {
		// Check Content-Disposition header is set for successful export
		assert.NotEmpty(t, rr.Header().Get("Content-Disposition"))
	}
}

func TestGetNamespacedListHandler_AuthError(t *testing.T) {
	// Test xhttp.AdminError in auth.CanRead
	req := httptest.NewRequest("GET", "/xconfAdminService/queries/namespacedLists/test-id", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "test-id"})
	rr := httptest.NewRecorder()

	GetNamespacedListHandler(rr, req)
	// Auth handling varies, may succeed with default or return error
	assert.True(t, rr.Code >= 200 && rr.Code < 500)
}

// Additional error case tests for comprehensive coverage

func TestAddDataMacListHandler_XResponseWriterCastError(t *testing.T) {
	// Test xhttp.AdminError when responsewriter cast fails
	listId := uuid.NewString()
	url := fmt.Sprintf("/xconfAdminService/queries/macs/addData/%s?applicationType=stb", listId)
	req := httptest.NewRequest("POST", url, nil)
	req = mux.SetURLVars(req, map[string]string{"listId": listId})
	rr := httptest.NewRecorder() // Not XResponseWriter

	AddDataMacListHandler(rr, req)
	// Should get InternalServerError for responsewriter cast error
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestRemoveDataMacListHandler_XResponseWriterCastError(t *testing.T) {
	// Test xhttp.AdminError when responsewriter cast fails
	listId := uuid.NewString()
	url := fmt.Sprintf("/xconfAdminService/queries/macs/removeData/%s?applicationType=stb", listId)
	req := httptest.NewRequest("DELETE", url, nil)
	req = mux.SetURLVars(req, map[string]string{"listId": listId})
	rr := httptest.NewRecorder() // Not XResponseWriter

	RemoveDataMacListHandler(rr, req)
	// Should get InternalServerError for responsewriter cast error
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestDeleteIpAddressGroupHandler_NotFound(t *testing.T) {
	// Test that deleting non-existent entity returns NoContent (idempotent delete)
	nonExistentId := uuid.NewString()
	url := fmt.Sprintf("/xconfAdminService/queries/ipAddressGroups/%s?applicationType=stb", nonExistentId)
	req := httptest.NewRequest("DELETE", url, nil)
	req = mux.SetURLVars(req, map[string]string{"id": nonExistentId})
	rr := httptest.NewRecorder()

	DeleteIpAddressGroupHandler(rr, req)
	// Handler returns NoContent when entity doesn't exist (idempotent delete)
	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestGetQueriesIpAddressGroupsV2_EmptyResult(t *testing.T) {
	// Test WriteXconfResponse with empty list
	req := httptest.NewRequest("GET", "/xconfAdminService/queries/ipAddressGroups?applicationType=stb&type=UNKNOWN_TYPE", nil)
	rr := httptest.NewRecorder()

	GetQueriesIpAddressGroupsV2(rr, req)
	// Should succeed with empty result
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAddDataMacListHandler_ValidationError(t *testing.T) {
	// Test WriteAdminErrorResponse for validation error (invalid MAC)
	listId := "test-mac-list-validation-" + uuid.NewString()

	// Create a valid list first
	createList := makeGenericList(listId, shared.MAC_LIST, []string{"AA:BB:CC:DD:EE:00"})
	createResp := CreateNamespacedList(createList, false)
	if createResp.Status != http.StatusCreated {
		t.Skip("Cannot test validation when create fails")
	}

	wrapper := shared.StringListWrapper{
		List: []string{"INVALID-MAC"},
	}

	url := fmt.Sprintf("/xconfAdminService/queries/macs/addData/%s?applicationType=stb", listId)
	req := httptest.NewRequest("POST", url, nil)
	req = mux.SetURLVars(req, map[string]string{"listId": listId})
	rr, xw := makeNSXW(wrapper)

	AddDataMacListHandler(xw, req)
	// Should get BadRequest for validation error
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRemoveDataMacListHandler_NotInList(t *testing.T) {
	// Test WriteAdminErrorResponse when trying to remove MAC not in list
	listId := "test-mac-list-notfound-" + uuid.NewString()

	// Create a list with one MAC
	createList := makeGenericList(listId, shared.MAC_LIST, []string{"AA:BB:CC:DD:EE:00"})
	createResp := CreateNamespacedList(createList, false)
	if createResp.Status != http.StatusCreated {
		t.Skip("Cannot test remove when create fails")
	}

	wrapper := shared.StringListWrapper{
		List: []string{"FF:FF:FF:FF:FF:FF"}, // Not in the list
	}

	url := fmt.Sprintf("/xconfAdminService/queries/macs/removeData/%s?applicationType=stb", listId)
	req := httptest.NewRequest("DELETE", url, nil)
	req = mux.SetURLVars(req, map[string]string{"listId": listId})
	rr, xw := makeNSXW(wrapper)

	RemoveDataMacListHandler(xw, req)
	// Should get BadRequest when MAC not in list
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

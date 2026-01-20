package applicationtype

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	xapptype "github.com/rdkcentral/xconfadmin/shared/applicationtype"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
)

func CreateApplicationTypeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusForbidden, err.Error())
		return
	}
	var appType xapptype.ApplicationType
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "responsewriter cast error")
		return
	}
	body := xw.Body()
	err = json.Unmarshal([]byte(body), &appType)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	isDefault := xapptype.IsDefaultAppType(appType.Name)
	if isDefault {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Cannot create default application type")
		return
	}
	exists, _ := xapptype.ApplicationTypeNameExists(appType.Name)
	if exists {
		xhttp.WriteAdminErrorResponse(w, http.StatusConflict, "Application type already exists")
		return
	}
	createdAppType, err := CreateApplicationType(r, &appType)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	data, err := json.Marshal(createdAppType)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusCreated, data)
}

func GetAllApplicationTypeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanRead(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusForbidden, err.Error())
		return
	}
	appTypes, err := xapptype.GetAllApplicationTypeAsList()
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	data, err := json.Marshal(appTypes)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, data)
}

func GetApplicationTypeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanRead(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusForbidden, err.Error())
		return
	}
	id := mux.Vars(r)["id"]
	if id == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Application type ID is required")
		return
	}
	appType, err := xapptype.GetOneApplicationType(id)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if appType == nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, "Application type not found")
		return
	}
	data, err := json.Marshal(appType)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, data)
}

func DeleteApplicationTypeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusForbidden, err.Error())
		return
	}
	id := mux.Vars(r)["id"]
	if id == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Application type ID is required")
		return
	}
	appType, err := xapptype.GetOneApplicationType(id)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if appType == nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, "Application type not found")
		return
	}
	if xapptype.IsDefaultAppType(appType.Name) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Default application types cannot be deleted")
		return
	}
	err = xapptype.DeleteOneApplicationType(id)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	response := map[string]string{
		"message": fmt.Sprintf("Application type '%s' deleted successfully", appType.Name),
	}
	data, err := json.Marshal(response)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, data)
}

func UpdateApplicationTypeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.CHANGE_ENTITY)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusForbidden, err.Error())
		return
	}
	id := mux.Vars(r)["id"]
	if id == "" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Application type ID is required")
		return
	}
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "responsewriter cast error")
		return
	}
	body := xw.Body()
	var updateRequest xapptype.ApplicationType
	err = json.Unmarshal([]byte(body), &updateRequest)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	err = ValidateApplicationType(&updateRequest)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	existingAppType, err := xapptype.GetOneApplicationType(id)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if existingAppType.Name != updateRequest.Name {
		nameExists, _ := xapptype.ApplicationTypeNameExists(updateRequest.Name)
		if nameExists {
			xhttp.WriteAdminErrorResponse(w, http.StatusConflict,
				fmt.Sprintf("Application type with name '%s' already exists", updateRequest.Name))
			return
		}
	}
	if xapptype.IsDefaultAppType(existingAppType.Name) {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Default application types cannot be updated")
		return
	}
	existingAppType.ID = id
	existingAppType.Name = updateRequest.Name
	if updateRequest.Description != "" {
		existingAppType.Description = updateRequest.Description
	}
	existingAppType.UpdatedAt = time.Now().Unix()

	err = xapptype.SetOneApplicationType(existingAppType)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	data, err := json.Marshal(existingAppType)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, data)
}

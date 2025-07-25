package queries

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	ccommon "github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	util "github.com/rdkcentral/xconfadmin/util"

	"github.com/rdkcentral/xconfwebconfig/db"

	"github.com/gorilla/mux"
)

func GetPenetrationDataByMacHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CanRead(r, auth.TOOL_ENTITY); err != nil {
		xhttp.AdminError(w, err)
		return
	}

	estbMac := mux.Vars(r)[ccommon.MAC_ADDRESS]
	normalizedMac, err := util.ValidateAndNormalizeMacAddress(estbMac)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	pr, err := db.GetDatabaseClient().GetPenetrationMetrics(normalizedMac)
	if err != nil {
		errorStr := fmt.Sprintf("%v not found", normalizedMac)
		xhttp.WriteAdminErrorResponse(w, http.StatusNotFound, errorStr)
		return
	}
	res, err := json.Marshal(pr)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, res)
}

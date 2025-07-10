package ipmacrule

import (
	"encoding/json"
	"net/http"

	"xconfadmin/adminapi/auth"
	"xconfadmin/common"
	ccommon "xconfadmin/common"
	xhttp "xconfadmin/http"
)

func GetIpMacRuleConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanRead(r, auth.COMMON_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}
	macIpRuleConfig := common.MacIpRuleConfig{
		IpMacIsConditionLimit: ccommon.IpMacIsConditionLimit,
	}
	if b, err := json.Marshal(macIpRuleConfig); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

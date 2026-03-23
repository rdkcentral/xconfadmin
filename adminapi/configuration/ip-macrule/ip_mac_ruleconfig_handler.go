package ipmacrule

import (
	"encoding/json"
	"net/http"

	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	"github.com/rdkcentral/xconfadmin/common"
	ccommon "github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"

	log "github.com/sirupsen/logrus"
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
		log.WithFields(log.Fields{"error": err}).Error("failed to marshal ip mac rule config")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

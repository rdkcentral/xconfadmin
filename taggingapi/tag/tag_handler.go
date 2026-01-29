package tag

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"
)

const (
	RequestBodyReadErrorMsg        = "request body unmarshall error: %s"
	NotSpecifiedErrorMsg           = "%s is not specified"
	EmptyListErrorMsg              = "%s list is empty"
	MaxMemberLimitExceededErrorMsg = "batch size %d exceeds the limit of %d"
	ResponseWriterCastErrorMsg     = "response writer cast error"
	NotFoundErrorMsg               = "%s tag not found"

	TagMemberLimit = 1000
)

func GetTagsByMemberHandler(w http.ResponseWriter, r *http.Request) {
	member, found := mux.Vars(r)[common.Member]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Member)))
		return
	}
	tags, err := GetTagsByMember(member)
	respBytes, err := json.Marshal(tags)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, respBytes)
}

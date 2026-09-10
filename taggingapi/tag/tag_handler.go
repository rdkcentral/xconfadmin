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

	tagType, err := getTagTypeFromRequest(r)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	audit := newOpAudit(w, OpReverseLookup)
	audit.setTagType(tagType)

	tags, err := GetTagsByMember(member, tagType)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	audit.set("num_results", len(tags))

	respBytes, err := json.Marshal(tags)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, respBytes)
}

func GetTagsWithValuesByMemberHandler(w http.ResponseWriter, r *http.Request) {
	member, found := mux.Vars(r)[common.Member]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Member)))
		return
	}

	tagType, err := getTagTypeFromRequest(r)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	audit := newOpAudit(w, OpReverseLookupValues)
	audit.setTagType(tagType)

	tags, err := GetTagsWithValuesByMember(member, tagType)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	audit.set("num_results", len(tags))

	respBytes, err := json.Marshal(tags)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, respBytes)
}

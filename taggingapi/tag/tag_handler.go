package tag

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	"github.com/rdkcentral/xconfadmin/taggingapi/percentage"

	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

	"github.com/gorilla/mux"
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

func GetAllTagsHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	_, ok := queryParams[common.FULL]
	if ok {
		tags, err := GetAllTags()
		if err != nil {
			xhttp.WriteXconfErrorResponse(w, err)
			return
		}
		resp, err := xhttp.ReturnJsonResponse(tags, r)
		if err != nil {
			xhttp.WriteXconfErrorResponse(w, err)
		}
		xhttp.WriteXconfResponse(w, http.StatusOK, resp)
		return
	}

	tagIds, err := GetAllTagIds()
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	resp, err := xhttp.ReturnJsonResponse(tagIds, r)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, resp)
}

func GetTagByIdHandler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}
	tag := GetTagById(id)
	if tag == nil {
		xhttp.WriteXconfResponse(w, http.StatusNotFound, []byte(fmt.Sprintf(NotFoundErrorMsg, id)))
		return
	}

	respBytes, err := json.Marshal(tag)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, respBytes)
}

func DeleteTagHandler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}
	err := DeleteTag(id)
	if err != nil {
		xhttp.WriteXconfResponse(w, http.StatusNotFound, []byte(fmt.Sprintf(NotFoundErrorMsg, id)))
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusNoContent, nil)
}

// DeleteTagFromXconfWithoutPrefixHandler deletes a tag from xConf without the prefix
// Only for testing and clean up purpose, should be removed before deploying to production
func DeleteTagFromXconfWithoutPrefixHandler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}
	tag := GetOneTag(id)
	if tag == nil {
		xhttp.WriteXconfResponse(w, http.StatusNotFound, []byte(fmt.Sprintf(NotFoundErrorMsg, id)))
		return
	}
	err := DeleteOneTag(id)
	if err != nil {
		xhttp.WriteXconfResponse(w, http.StatusNotFound, []byte(fmt.Sprintf(NotFoundErrorMsg, id)))
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusNoContent, nil)
}

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

func AddMembersToTagHandler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteXconfResponse(w, http.StatusInternalServerError, []byte(ResponseWriterCastErrorMsg))
		return
	}
	var members []string
	if err := json.Unmarshal([]byte(xw.Body()), &members); err != nil {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(RequestBodyReadErrorMsg, err.Error())))
		return
	}
	if len(members) == 0 {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(EmptyListErrorMsg, common.Member)))
		return
	}
	if err := CheckBatchSizeExceeded(len(members)); err != nil {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(err.Error()))
		return
	}
	_, err := AddMembersToTag(id, members)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, nil)
}

func RemoveMemberFromTagHandler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}
	member, found := mux.Vars(r)[common.Member]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Member)))
		return
	}

	tag, err := RemoveMemberFromTag(id, member)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	respBytes, err := json.Marshal(tag)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusNoContent, respBytes)
}

func GetTagsByMemberPercentageHandler(w http.ResponseWriter, r *http.Request) {
	member, found := mux.Vars(r)[common.Member]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Member)))
		return
	}

	tags, err := GetTagsByMemberPercentage(member)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	respBytes, err := json.Marshal(tags)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, respBytes)
}

func GetTagMembersHandler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}

	members, err := GetTagMembers(id)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	respBytes, err := json.Marshal(members)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, respBytes)
}

func CalculatePercentageValueHandler(w http.ResponseWriter, r *http.Request) {
	member, found := mux.Vars(r)[common.Member]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Member)))
		return
	}
	calculated := percentage.CalculatePercent(member)
	respBytes, err := json.Marshal(calculated)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, respBytes)

}

func AddMemberPercentageToTagHandler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}

	startRangeStr, found := mux.Vars(r)[common.StartRange]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.StartRange)))
		return
	}

	endRangeStr, found := mux.Vars(r)[common.EndRange]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.EndRange)))
		return
	}

	err := AddAccountRangeToTag(id, startRangeStr, endRangeStr)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	xhttp.WriteXconfResponse(w, http.StatusOK, nil)
}

func CleanPercentageRangeHandler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}

	err := CleanPercentageRange(id)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusNoContent, nil)
}

func RemoveMembersFromTagHandler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}
	var members []string
	body, err := io.ReadAll(r.Body)
	if err != nil {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(RequestBodyReadErrorMsg, err.Error())))
		return
	}
	if err := json.Unmarshal(body, &members); err != nil {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(RequestBodyReadErrorMsg, err.Error())))
		return
	}
	if len(members) == 0 {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(EmptyListErrorMsg, common.Member)))
		return
	}
	if err := CheckBatchSizeExceeded(len(members)); err != nil {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(err.Error()))
		return
	}
	_, err = RemoveMembersFromTag(id, members)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusNoContent, nil)
}

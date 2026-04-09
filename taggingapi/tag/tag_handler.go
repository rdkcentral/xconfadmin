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
	InvalidAccountIdErrorMsg       = "invalid account ID '%s': must contain only digits"

	TagMemberLimit = 1000
)

// getTagType extracts the tag type from the URL path variable.
// Defaults to TagTypeMac if not present (backward compatibility with old routes).
func getTagType(r *http.Request) (string, error) {
	tagType, found := mux.Vars(r)[common.TagType]
	if !found || tagType == "" {
		return TagTypeMac, nil
	}
	if err := ValidateTagType(tagType); err != nil {
		return "", err
	}
	return tagType, nil
}

func GetTagsByMemberHandler(w http.ResponseWriter, r *http.Request) {
	member, found := mux.Vars(r)[common.Member]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Member)))
		return
	}
	tagType, err := getTagType(r)
	if err != nil {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(err.Error()))
		return
	}
	if tagType == TagTypeAccount {
		if err := ValidateAccountId(member); err != nil {
			xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(InvalidAccountIdErrorMsg, member)))
			return
		}
	}
	tags, err := GetTagsByMember(member, tagType)
	respBytes, err := json.Marshal(tags)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, respBytes)
}

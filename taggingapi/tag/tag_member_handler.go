package tag

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"

	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultPageSize = 1000
	MaxPageSize     = 5000
)

func parsePaginationParams(r *http.Request) (*PaginationParams, error) {
	query := r.URL.Query()

	limit := DefaultPageSizeV2
	if limitStr := query.Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, fmt.Errorf("invalid limit parameter: %s", limitStr)
		}
		if parsedLimit > MaxPageSizeV2 {
			return nil, fmt.Errorf("limit %d exceeds maximum %d", parsedLimit, MaxPageSizeV2)
		}
		if parsedLimit < 1 {
			return nil, fmt.Errorf("limit must be positive")
		}
		limit = parsedLimit
	}

	cursor := query.Get("cursor")

	return &PaginationParams{
		Limit:  limit,
		Cursor: cursor,
	}, nil
}

// GetTagMembersHandler - Unified handler supporting both paginated and non-paginated responses
// Non-paginated mode (V1 compatible): Returns []string with up to 100k members, HTTP 206 if truncated
// Paginated mode: Returns paginated envelope when limit/cursor params are present
func GetTagMembersHandler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}

	query := r.URL.Query()
	isPaginatedRequest := query.Has("limit") || query.Has("cursor")

	if isPaginatedRequest {
		params, err := parsePaginationParams(r)
		if err != nil {
			xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(err.Error()))
			return
		}

		response, err := GetMembersPaginated(id, params.Limit, params.Cursor)
		if err != nil {
			xhttp.WriteXconfErrorResponse(w, err)
			return
		}

		respBytes, err := json.Marshal(response)
		if err != nil {
			xhttp.WriteXconfErrorResponse(w, err)
			return
		}

		xhttp.WriteXconfResponse(w, http.StatusOK, respBytes)
	} else {
		// Non-paginated mode: return plain array (V1 compatible)
		members, wasTruncated, err := GetMembersNonPaginated(id)
		if err != nil {
			xhttp.WriteXconfErrorResponse(w, err)
			return
		}

		respBytes, err := json.Marshal(members)
		if err != nil {
			xhttp.WriteXconfErrorResponse(w, err)
			return
		}

		statusCode := http.StatusOK
		if wasTruncated {
			statusCode = http.StatusPartialContent
		}

		xhttp.WriteXconfResponse(w, statusCode, respBytes)
	}
}

// AddMembersToTagHandler - Updated with bucketed implementation
func AddMembersToTagHandler(w http.ResponseWriter, r *http.Request) {
	tagId, found := mux.Vars(r)[common.Tag]
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

	if len(members) > MaxBatchSizeV2 {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest,
			[]byte(fmt.Sprintf("Batch size %d exceeds maximum %d", len(members), MaxBatchSizeV2)))
		return
	}

	err := AddMembersWithXdas(tagId, members)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	xhttp.WriteXconfResponse(w, http.StatusAccepted, nil)
}

// RemoveMembersFromTagHandler - Updated with bucketed implementation
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

	if len(members) > MaxBatchSizeV2 {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest,
			[]byte(fmt.Sprintf("Batch size %d exceeds maximum %d", len(members), MaxBatchSizeV2)))
		return
	}

	err = RemoveMembersWithXdas(id, members)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	xhttp.WriteXconfResponse(w, http.StatusAccepted, nil)
}

// RemoveMemberFromTagHandler - Updated with bucketed implementation
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

	err := RemoveMemberWithXdas(id, member)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	xhttp.WriteXconfResponse(w, http.StatusNoContent, nil)
}

// GetAllTagsHandler returns all tag IDs from V2 storage
func GetAllTagsHandler(w http.ResponseWriter, r *http.Request) {
	tagIds, err := GetAllTagIds()
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	respBytes, err := json.Marshal(tagIds)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	xhttp.WriteXconfResponse(w, http.StatusOK, respBytes)
}

// GetTagByIdHandler retrieves a single tag with its members from V2 storage
func GetTagByIdHandler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}

	members, wasTruncated, err := GetTagById(id)
	if err != nil {
		// Check if tag not found
		if err.Error() == "tag not found" {
			xhttp.WriteXconfResponse(w, http.StatusNotFound, []byte(fmt.Sprintf(NotFoundErrorMsg, id)))
			return
		}
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	// Build response matching V1 format (without updated field)
	response := struct {
		Id      string   `json:"id"`
		Members []string `json:"members"`
	}{
		Id:      id,
		Members: members,
	}

	respBytes, err := json.Marshal(response)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	// Return 206 Partial Content if truncated, otherwise 200 OK
	statusCode := http.StatusOK
	if wasTruncated {
		statusCode = http.StatusPartialContent
	}

	xhttp.WriteXconfResponse(w, statusCode, respBytes)
}

// DeleteTagHandler deletes a tag and all its members from V2 storage asynchronously
func DeleteTagHandler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}

	populatedBuckets, err := getPopulatedBuckets(id)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	if len(populatedBuckets) == 0 {
		xhttp.WriteXconfResponse(w, http.StatusNotFound, []byte(fmt.Sprintf(NotFoundErrorMsg, id)))
		return
	}

	go func(tagId string) {
		if err := DeleteTag(tagId); err != nil {
			log.Errorf("Background deletion failed for tag '%s': %v", tagId, err)
		}
	}(id)

	response := map[string]string{
		"status":  "accepted",
		"message": fmt.Sprintf("Tag '%s' deletion has been queued for processing", id),
		"tag":     id,
	}

	respBytes, err := json.Marshal(response)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	xhttp.WriteXconfResponse(w, http.StatusAccepted, respBytes)
}

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

// GetTagMembersV2Handler - Unified handler supporting both paginated and non-paginated responses
// Non-paginated mode (V1 compatible): Returns []string with up to 100k members, HTTP 206 if truncated
// Paginated mode: Returns paginated envelope when limit/cursor params are present
func GetTagMembersV2Handler(w http.ResponseWriter, r *http.Request) {
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

		response, err := GetMembersV2Paginated(id, params.Limit, params.Cursor)
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
		members, wasTruncated, err := GetMembersV2NonPaginated(id)
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

// AddMembersToTagV2Handler - Updated with bucketed implementation
func AddMembersToTagV2Handler(w http.ResponseWriter, r *http.Request) {
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

// RemoveMembersFromTagV2Handler - Updated with bucketed implementation
func RemoveMembersFromTagV2Handler(w http.ResponseWriter, r *http.Request) {
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

	err = RemoveMembersV2WithXdas(id, members)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	xhttp.WriteXconfResponse(w, http.StatusAccepted, nil)
}

// RemoveMemberFromTagV2Handler - Updated with bucketed implementation
func RemoveMemberFromTagV2Handler(w http.ResponseWriter, r *http.Request) {
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

	err := RemoveMemberV2WithXdas(id, member)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	xhttp.WriteXconfResponse(w, http.StatusNoContent, nil)
}

// GetAllTagsV2Handler returns all tag IDs from V2 storage
func GetAllTagsV2Handler(w http.ResponseWriter, r *http.Request) {
	tagIds, err := GetAllTagIdsV2()
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

// GetTagByIdV2Handler retrieves a single tag with its members from V2 storage
func GetTagByIdV2Handler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}

	members, wasTruncated, err := GetTagByIdV2(id)
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

// DeleteTagV2Handler deletes a tag and all its members from V2 storage
func DeleteTagV2Handler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}

	err := DeleteTagV2(id)
	if err != nil {
		// Check if tag not found
		if err.Error() == "tag not found" {
			xhttp.WriteXconfResponse(w, http.StatusNotFound, []byte(fmt.Sprintf(NotFoundErrorMsg, id)))
			return
		}
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	xhttp.WriteXconfResponse(w, http.StatusNoContent, nil)
}

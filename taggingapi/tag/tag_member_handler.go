package tag

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"

	"github.com/rdkcentral/xconfadmin/adminapi/auth"
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

// inFlightTagDeletions deduplicates concurrent background deletions of the
// same tag on this instance, keyed by "tenantId|tagId". Interim guard until
// tag deletion state is tracked cross-instance (tag registry, topic 3).
var inFlightTagDeletions sync.Map

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
	_, err := auth.CanRead(r, auth.COMMON_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}

	tenantId := xhttp.GetTenantId(r)
	query := r.URL.Query()
	isPaginatedRequest := query.Has("limit") || query.Has("cursor")

	if isPaginatedRequest {
		audit := newOpAudit(w, OpGetMembersPage, tenantId)
		audit.setTag(id)

		params, err := parsePaginationParams(r)
		if err != nil {
			xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(err.Error()))
			return
		}
		audit.set("page_limit", params.Limit)
		audit.set("has_cursor", params.Cursor != "")

		response, stats, err := GetMembersPaginated(tenantId, id, params.Limit, params.Cursor)
		audit.setReadStats(stats)
		if err != nil {
			xhttp.WriteXconfErrorResponse(w, err)
			return
		}
		audit.set("num_results", len(response.Data))
		audit.set("has_more", response.HasMore)

		respBytes, err := json.Marshal(response)
		if err != nil {
			xhttp.WriteXconfErrorResponse(w, err)
			return
		}

		xhttp.WriteXconfResponse(w, http.StatusOK, respBytes)
	} else {
		// Non-paginated mode: return plain array (V1 compatible)
		audit := newOpAudit(w, OpGetMembersFull, tenantId)
		audit.setTag(id)

		members, wasTruncated, stats, err := GetMembersNonPaginated(tenantId, id)
		audit.setReadStats(stats)
		if err != nil {
			xhttp.WriteXconfErrorResponse(w, err)
			return
		}
		audit.set("num_results", len(members))
		audit.set("truncated", wasTruncated)

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
	_, err := auth.CanWrite(r, auth.COMMON_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	tagId, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}

	tagValue := getTagValueFromRequest(r)
	tenantId := xhttp.GetTenantId(r)

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

	audit := newOpAudit(w, OpAddMembers, tenantId)
	audit.setTag(tagId)

	stats, err := AddMembersWithXdas(tenantId, tagId, members, tagValue)
	audit.setWriteStats(stats)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	response := map[string]int{
		"requested": len(members),
		"stored":    stats.CassandraOk,
	}
	respBytes, err := json.Marshal(response)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	xhttp.WriteXconfResponse(w, http.StatusAccepted, respBytes)
}

func getTagValueFromRequest(r *http.Request) string {
	if values, found := r.URL.Query()[common.TagValue]; found {
		if len(values) > 0 {
			return values[0]
		}
	}

	return ""
}

// RemoveMembersFromTagHandler - Updated with bucketed implementation
func RemoveMembersFromTagHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.COMMON_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}

	tenantId := xhttp.GetTenantId(r)

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

	audit := newOpAudit(w, OpRemoveMembers, tenantId)
	audit.setTag(id)

	stats, err := RemoveMembersWithXdas(tenantId, id, members)
	audit.setWriteStats(stats)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	response := map[string]int{
		"requested": len(members),
		"removed":   stats.CassandraOk,
	}
	respBytes, err := json.Marshal(response)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	xhttp.WriteXconfResponse(w, http.StatusAccepted, respBytes)
}

// RemoveMemberFromTagHandler - Updated with bucketed implementation
func RemoveMemberFromTagHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanWrite(r, auth.COMMON_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

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

	tenantId := xhttp.GetTenantId(r)
	audit := newOpAudit(w, OpRemoveMember, tenantId)
	audit.setTag(id)

	stats, err := RemoveMemberWithXdas(tenantId, id, member)
	audit.setWriteStats(stats)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	xhttp.WriteXconfResponse(w, http.StatusNoContent, nil)
}

// GetAllTagsHandler returns all tag IDs from V2 storage
func GetAllTagsHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanRead(r, auth.COMMON_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	tenantId := xhttp.GetTenantId(r)
	audit := newOpAudit(w, OpGetAllTags, tenantId)

	tagIds, err := GetAllTagIds(tenantId)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	audit.set("num_results", len(tagIds))

	respBytes, err := json.Marshal(tagIds)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	xhttp.WriteXconfResponse(w, http.StatusOK, respBytes)
}

// GetTagByIdHandler retrieves a single tag with its members from V2 storage
func GetTagByIdHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.CanRead(r, auth.COMMON_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

	id, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}

	tenantId := xhttp.GetTenantId(r)
	audit := newOpAudit(w, OpGetTag, tenantId)
	audit.setTag(id)

	members, wasTruncated, stats, err := GetTagById(tenantId, id)
	audit.setReadStats(stats)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	audit.set("num_results", len(members))
	audit.set("truncated", wasTruncated)

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
	_, err := auth.CanWrite(r, auth.COMMON_ENTITY)
	if err != nil {
		xhttp.AdminError(w, err)
		return
	}

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

	tenantId := xhttp.GetTenantId(r)
	audit := newOpAudit(w, OpDeleteTag, tenantId)
	audit.setTag(id)

	populatedBuckets, err := getPopulatedBuckets(tenantId, id)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	if len(populatedBuckets) == 0 {
		xhttp.WriteXconfResponse(w, http.StatusNotFound, []byte(fmt.Sprintf(NotFoundErrorMsg, id)))
		return
	}
	audit.set("buckets", len(populatedBuckets))

	deletionKey := tenantId + "|" + id
	if _, alreadyRunning := inFlightTagDeletions.LoadOrStore(deletionKey, true); alreadyRunning {
		audit.set("deletion_state", "already_in_progress")
		response := map[string]string{
			"status":  "accepted",
			"message": fmt.Sprintf("Tag '%s' deletion is already in progress", id),
			"tag":     id,
		}
		respBytes, err := json.Marshal(response)
		if err != nil {
			xhttp.WriteXconfErrorResponse(w, err)
			return
		}
		xhttp.WriteXconfResponse(w, http.StatusAccepted, respBytes)
		return
	}
	audit.set("deletion_state", "accepted")

	auditId := xw.AuditId()
	go func(tagId string) {
		defer inFlightTagDeletions.Delete(deletionKey)
		// DeleteTag logs its own START/PROGRESS/END lines with the audit_id;
		// only the failure needs an extra line here.
		if err := DeleteTag(tenantId, tagId, auditId); err != nil {
			log.WithFields(log.Fields{
				"audit_id": auditId,
				"op":       OpDeleteTag,
				"tenant":   tenantId,
				"tag":      tagId,
			}).Errorf("background deletion failed: %v", err)
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

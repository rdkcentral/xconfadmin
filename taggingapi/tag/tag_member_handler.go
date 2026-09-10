package tag

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultPageSize = 1000
	MaxPageSize     = 5000
)

// inFlightTagDeletions deduplicates concurrent background deletions of the same
// tag on this instance, keyed by tag id. Interim guard until deletion state is
// tracked cross-instance.
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

// GetTagMembersHandler serves both response shapes: a paginated envelope when
// limit/cursor are present, otherwise a plain []string of up to 100k members
// (V1 compatible), with HTTP 206 if truncated.
func GetTagMembersHandler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}

	// The route type scopes the lookup: a tag stored under another type is not
	// visible here, the same way the listing filters it out.
	tagType, err := getTagTypeFromRequest(r)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	query := r.URL.Query()
	isPaginatedRequest := query.Has("limit") || query.Has("cursor")

	if isPaginatedRequest {
		audit := newOpAudit(w, OpGetMembersPage)
		audit.setTag(id)
		audit.setTagType(tagType)

		params, err := parsePaginationParams(r)
		if err != nil {
			xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(err.Error()))
			return
		}
		audit.set("page_limit", params.Limit)
		audit.set("has_cursor", params.Cursor != "")

		response, stats, err := GetMembersPaginated(id, params.Limit, params.Cursor, tagType)
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
		audit := newOpAudit(w, OpGetMembersFull)
		audit.setTag(id)
		audit.setTagType(tagType)

		members, wasTruncated, stats, err := GetMembersNonPaginated(id, tagType)
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

func AddMembersToTagHandler(w http.ResponseWriter, r *http.Request) {
	tagId, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}

	tagType, err := getTagTypeFromRequest(r)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	// Enforced on the typed and untyped routes alike: a tag with a reserved id
	// must never grow. Reads and deletes skip the check, so such a tag (none
	// exist in any environment) could still be drained, just not extended.
	if err := validateTagId(tagId); err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	tagValue := getTagValueFromRequest(r)

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

	audit := newOpAudit(w, OpAddMembers)
	audit.setTag(tagId)
	audit.setTagType(tagType)

	stats, err := AddMembersWithXdas(tagId, members, tagValue, tagType)
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

// getTagTypeFromRequest reads the {tagType} path variable, falling back to the
// ?tagType= query parameter (the untyped list endpoint's filter).
//
// Absent means TagTypeLegacy, deliberately not TagTypeMac: the empty string is
// what tells the service layer to stay permissive on untyped routes. Every typed
// handler resolves through here, so this is also where the account type is gated
// on tag_type_column_enabled (ensureTagTypeSupported).
func getTagTypeFromRequest(r *http.Request) (string, error) {
	tagType, found := mux.Vars(r)[common.TagType]
	if !found {
		tagType = r.URL.Query().Get(common.TagType)
	}
	if err := ValidateTagType(tagType); err != nil {
		return "", err
	}
	if err := ensureTagTypeSupported(tagType); err != nil {
		return "", err
	}
	return tagType, nil
}

// reservedTagIds are path segments in the typed routes, so a tag with one of
// these ids would be shadowed by its route and left unreachable — mux resolves
// /taggingService/tags/account to the typed route. "members" collides the same
// way with /taggingService/tags/members/{member}.
var reservedTagIds = map[string]bool{
	TagTypeMac:          true,
	TagTypeAccount:      true,
	common.Member + "s": true,
}

func validateTagId(tagId string) error {
	if reservedTagIds[strings.ToLower(tagId)] {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest,
			fmt.Sprintf("tag id '%s' is reserved and cannot be used", tagId))
	}
	return nil
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

	if len(members) > MaxBatchSizeV2 {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest,
			[]byte(fmt.Sprintf("Batch size %d exceeds maximum %d", len(members), MaxBatchSizeV2)))
		return
	}

	tagType, err := getTagTypeFromRequest(r)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	audit := newOpAudit(w, OpRemoveMembers)
	audit.setTag(id)
	audit.setTagType(tagType)

	stats, err := RemoveMembersWithXdas(id, members, tagType)
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

	tagType, err := getTagTypeFromRequest(r)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	audit := newOpAudit(w, OpRemoveMember)
	audit.setTag(id)
	audit.setTagType(tagType)

	stats, err := RemoveMemberWithXdas(id, member, tagType)
	audit.setWriteStats(stats)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	xhttp.WriteXconfResponse(w, http.StatusNoContent, nil)
}

// GetAllTagsHandler returns all tag IDs, narrowed by {tagType} on the typed
// routes; an absent type on the legacy route means "everything".
func GetAllTagsHandler(w http.ResponseWriter, r *http.Request) {
	audit := newOpAudit(w, OpGetAllTags)

	// On the typed routes {tagType} narrows the listing; on the legacy route an
	// absent type means "everything", preserving the current response.
	tagType, err := getTagTypeFromRequest(r)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	audit.setTagType(tagType)

	tagIds, err := GetAllTagIds(tagType)
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

// GetTagByIdHandler retrieves a single tag with its members.
func GetTagByIdHandler(w http.ResponseWriter, r *http.Request) {
	id, found := mux.Vars(r)[common.Tag]
	if !found {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.Tag)))
		return
	}

	// Scopes the lookup to the route type — see GetTagMembersHandler.
	tagType, err := getTagTypeFromRequest(r)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	audit := newOpAudit(w, OpGetTag)
	audit.setTag(id)
	audit.setTagType(tagType)

	members, wasTruncated, stats, err := GetTagById(id, tagType)
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

// DeleteTagHandler queues asynchronous deletion of a tag and all its members.
func DeleteTagHandler(w http.ResponseWriter, r *http.Request) {
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

	requestedType, err := getTagTypeFromRequest(r)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	audit := newOpAudit(w, OpDeleteTag)
	audit.setTag(id)

	// Resolved synchronously: a lookup failure must surface as an error response,
	// not as a log line the caller never sees behind an already-sent 202.
	populatedBuckets, storedType, err := getTagMeta(id)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	if len(populatedBuckets) == 0 {
		xhttp.WriteXconfResponse(w, http.StatusNotFound, []byte(fmt.Sprintf(NotFoundErrorMsg, id)))
		return
	}
	audit.set("buckets", len(populatedBuckets))
	audit.setTagType(storedType)

	if requestedType != TagTypeLegacy && tagTypeColumnEnabled() {
		if err := checkTagTypeCompatible(id, storedType, requestedType); err != nil {
			xhttp.WriteXconfErrorResponse(w, err)
			return
		}
	}

	// Keyed on the bare tag id, not type+id: ids are unique across types, so
	// type+id would key one tag two ways and let a typed and an untyped DELETE
	// both pass the guard and delete the same partitions concurrently.
	deletionKey := id
	if existing, alreadyRunning := inFlightTagDeletions.LoadOrStore(deletionKey, storedType); alreadyRunning {
		audit.set("deletion_state", "already_in_progress")
		if inFlightType, ok := existing.(string); ok && inFlightType != TagTypeLegacy {
			audit.set("in_flight_tag_type", inFlightType)
		}
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
		// DeleteTag logs its own progress lines; only failure needs one here.
		if err := DeleteTag(tagId, auditId); err != nil {
			log.WithFields(log.Fields{
				"audit_id": auditId,
				"op":       OpDeleteTag,
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

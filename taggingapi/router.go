package taggingapi

import (
	xhttp "github.com/rdkcentral/xconfadmin/http"
	"github.com/rdkcentral/xconfadmin/taggingapi/tag"

	"github.com/gorilla/mux"
)

var (
	Ws *xhttp.WebconfigServer
)

func WebServerInjection(ws *xhttp.WebconfigServer) {
	Ws = ws
}

func XconfTaggingServiceSetup(server *xhttp.WebconfigServer, r *mux.Router) {
	WebServerInjection(server)
	tag.RegisterTaggingMetrics()
	routeTaggingServiceApis(r, server)
}

func routeTaggingServiceApis(r *mux.Router, s *xhttp.WebconfigServer) {
	paths := []*mux.Router{}

	// Typed routes first: mux matches in registration order, and a request
	// matching no typed route falls through to the untyped subrouter. The tag ids
	// "mac", "account" and "members" are rejected at write time (validateTagId),
	// so the two route sets never collide for a tag that exists.
	typedTaggingPath := r.PathPrefix("/taggingService/tags/{tagType:mac|account}").Subrouter()

	typedTaggingPath.HandleFunc("", tag.GetAllTagsHandler).Methods("GET").Name("Get-all-tags-typed")
	typedTaggingPath.HandleFunc("/{tag}", tag.GetTagByIdHandler).Methods("GET").Name("Get-tag-by-id-typed")
	typedTaggingPath.HandleFunc("/{tag}/members", tag.AddMembersToTagHandler).Methods("PUT").Name("Add-members-to-tag-typed")
	typedTaggingPath.HandleFunc("/{tag}", tag.DeleteTagHandler).Methods("DELETE").Name("Delete-tag-typed")
	typedTaggingPath.HandleFunc("/{tag}/members", tag.RemoveMembersFromTagHandler).Methods("DELETE").Name("Remove-members-from-tag-typed")
	typedTaggingPath.HandleFunc("/{tag}/members/{member}", tag.RemoveMemberFromTagHandler).Methods("DELETE").Name("Remove-member-from-tag-typed")

	typedTaggingPath.HandleFunc("/{tag}/members", tag.GetTagMembersHandler).Methods("GET").Name("Get-tag-members-typed")

	// After /{tag}/members so a tag named "members" wins the overlap; an account
	// member id is numeric and can never be "members".
	typedTaggingPath.HandleFunc("/members/{member}", tag.GetTagsByMemberHandler).Methods("GET").Name("Get-tags-by-member-typed")
	typedTaggingPath.HandleFunc("/members/{member}/values", tag.GetTagsWithValuesByMemberHandler).Methods("GET").Name("Get-tags-with-values-by-member-typed")

	taggingPath := r.PathPrefix("/taggingService/tags").Subrouter()

	taggingPath.HandleFunc("", tag.GetAllTagsHandler).Methods("GET").Name("Get-all-tags")

	taggingPath.HandleFunc("/sync", tag.TriggerTagSyncHandler).Methods("POST").Name("Trigger-tag-sync")
	taggingPath.HandleFunc("/sync/status", tag.TagSyncStatusHandler).Methods("GET").Name("Tag-sync-status")
	taggingPath.HandleFunc("/sync/abort", tag.AbortTagSyncHandler).Methods("POST").Name("Abort-tag-sync")

	taggingPath.HandleFunc("/{tag}", tag.GetTagByIdHandler).Methods("GET").Name("Get-tag-by-id")
	taggingPath.HandleFunc("/{tag}/members", tag.AddMembersToTagHandler).Methods("PUT").Name("Add-members-to-tag")
	taggingPath.HandleFunc("/{tag}", tag.DeleteTagHandler).Methods("DELETE").Name("Delete-tag-v2")
	taggingPath.HandleFunc("/{tag}/members", tag.RemoveMembersFromTagHandler).Methods("DELETE").Name("Remove-members-from-tag")
	taggingPath.HandleFunc("/{tag}/members/{member}", tag.RemoveMemberFromTagHandler).Methods("DELETE").Name("Remove-member-from-tag")

	taggingPath.HandleFunc("/{tag}/members", tag.GetTagMembersHandler).Methods("GET").Name("Get-tag-members")

	taggingPath.HandleFunc("/members/{member}", tag.GetTagsByMemberHandler).Methods("GET").Name("Get-tags-by-member")
	taggingPath.HandleFunc("/members/{member}/values", tag.GetTagsWithValuesByMemberHandler).Methods("GET").Name("Get-tags-with-values-by-member")

	paths = append(paths, typedTaggingPath, taggingPath)

	for _, p := range paths {
		if s.TestOnly() {
			p.Use(s.NoAuthMiddleware)
		} else {
			p.Use(s.XW_XconfServer.SpanMiddleware)
			p.Use(s.AuthValidationMiddleware)
		}
	}
}

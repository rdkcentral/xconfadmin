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

	// Typed routes are registered before the untyped ones because gorilla/mux
	// matches in registration order. The tag ids "mac", "account" and "members"
	// are rejected at write time (validateTagId), which is what keeps the two
	// route sets from ever being ambiguous for a tag that actually exists.
	//
	// A request under this prefix that matches no typed route falls through to
	// the untyped subrouter: mux clears ErrNotFound when a subrouter fails to
	// match and continues scanning.
	typedTaggingPath := r.PathPrefix("/taggingService/tags/{tagType:mac|account}").Subrouter()

	typedTaggingPath.HandleFunc("", tag.GetAllTagsHandler).Methods("GET").Name("Get-all-tags-typed")
	typedTaggingPath.HandleFunc("/{tag}", tag.GetTagByIdHandler).Methods("GET").Name("Get-tag-by-id-typed")
	typedTaggingPath.HandleFunc("/{tag}/members", tag.AddMembersToTagHandler).Methods("PUT").Name("Add-members-to-tag-typed")
	typedTaggingPath.HandleFunc("/{tag}", tag.DeleteTagHandler).Methods("DELETE").Name("Delete-tag-typed")
	typedTaggingPath.HandleFunc("/{tag}/members", tag.RemoveMembersFromTagHandler).Methods("DELETE").Name("Remove-members-from-tag-typed")
	typedTaggingPath.HandleFunc("/{tag}/members/{member}", tag.RemoveMemberFromTagHandler).Methods("DELETE").Name("Remove-member-from-tag-typed")

	typedTaggingPath.HandleFunc("/{tag}/members", tag.GetTagMembersHandler).Methods("GET").Name("Get-tag-members-typed")

	// Registered after /{tag}/members so a tag named "members" resolves to the
	// tag routes. An account member id can never be "members" (numeric only),
	// so tag-wins is the safer resolution of that overlap.
	typedTaggingPath.HandleFunc("/members/{member}", tag.GetTagsByMemberHandler).Methods("GET").Name("Get-tags-by-member-typed")
	typedTaggingPath.HandleFunc("/members/{member}/values", tag.GetTagsWithValuesByMemberHandler).Methods("GET").Name("Get-tags-with-values-by-member-typed")

	taggingPath := r.PathPrefix("/taggingService/tags").Subrouter()

	taggingPath.HandleFunc("", tag.GetAllTagsHandler).Methods("GET").Name("Get-all-tags")
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

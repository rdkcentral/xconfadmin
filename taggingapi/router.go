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
	routeTaggingServiceApis(r, server)
}

func routeTaggingServiceApis(r *mux.Router, s *xhttp.WebconfigServer) {
	paths := []*mux.Router{}

	// Typed routes registered first so {tagType:mac|account} takes priority over untyped /{tag}
	typedTaggingPath := r.PathPrefix("/taggingService/tags/{tagType:mac|account}").Subrouter()

	typedTaggingPath.HandleFunc("", tag.GetAllTagsHandler).Methods("GET").Name("Get-all-tags-typed")
	typedTaggingPath.HandleFunc("/{tag}", tag.GetTagByIdHandler).Methods("GET").Name("Get-tag-by-id-typed")
	typedTaggingPath.HandleFunc("/{tag}/members", tag.AddMembersToTagHandler).Methods("PUT").Name("Add-members-to-tag-typed")
	typedTaggingPath.HandleFunc("/{tag}", tag.DeleteTagHandler).Methods("DELETE").Name("Delete-tag-typed")
	typedTaggingPath.HandleFunc("/{tag}/members", tag.RemoveMembersFromTagHandler).Methods("DELETE").Name("Remove-members-from-tag-typed")
	typedTaggingPath.HandleFunc("/{tag}/members/{member}", tag.RemoveMemberFromTagHandler).Methods("DELETE").Name("Remove-member-from-tag-typed")

	typedTaggingPath.HandleFunc("/{tag}/members", tag.GetTagMembersHandler).Methods("GET").Name("Get-tag-members-typed")

	typedTaggingPath.HandleFunc("/members/{member}", tag.GetTagsByMemberHandler).Methods("GET").Name("Get-tags-by-member-typed")

	// Untyped routes (backward compatible, defaults to mac tag type)
	taggingPath := r.PathPrefix("/taggingService/tags").Subrouter()

	taggingPath.HandleFunc("", tag.GetAllTagsHandler).Methods("GET").Name("Get-all-tags")
	taggingPath.HandleFunc("/{tag}", tag.GetTagByIdHandler).Methods("GET").Name("Get-tag-by-id")
	taggingPath.HandleFunc("/{tag}/members", tag.AddMembersToTagHandler).Methods("PUT").Name("Add-members-to-tag")
	taggingPath.HandleFunc("/{tag}", tag.DeleteTagHandler).Methods("DELETE").Name("Delete-tag-v2")
	taggingPath.HandleFunc("/{tag}/members", tag.RemoveMembersFromTagHandler).Methods("DELETE").Name("Remove-members-from-tag")
	taggingPath.HandleFunc("/{tag}/members/{member}", tag.RemoveMemberFromTagHandler).Methods("DELETE").Name("Remove-member-from-tag")

	taggingPath.HandleFunc("/{tag}/members", tag.GetTagMembersHandler).Methods("GET").Name("Get-tag-members")

	taggingPath.HandleFunc("/members/{member}", tag.GetTagsByMemberHandler).Methods("GET").Name("Get-tags-by-member")

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

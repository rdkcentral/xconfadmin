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

	taggingPath := r.PathPrefix("/taggingService/tags").Subrouter()

	taggingPath.HandleFunc("", tag.GetAllTagsHandler).Methods("GET").Name("Get-all-tags")
	taggingPath.HandleFunc("/{tag}", tag.GetTagByIdHandler).Methods("GET").Name("Get-tag-by-id")
	taggingPath.HandleFunc("/{tag}/members", tag.AddMembersToTagHandler).Methods("PUT").Name("Add-members-to-tag")
	taggingPath.HandleFunc("/{tag}", tag.DeleteTagHandler).Methods("DELETE").Name("Delete-tag-v2")
	taggingPath.HandleFunc("/{tag}/members", tag.RemoveMembersFromTagHandler).Methods("DELETE").Name("Remove-members-from-tag")
	taggingPath.HandleFunc("/{tag}/members/{member}", tag.RemoveMemberFromTagHandler).Methods("DELETE").Name("Remove-member-from-tag")

	taggingPath.HandleFunc("/{tag}/members", tag.GetTagMembersHandler).Methods("GET").Name("Get-tag-members")

	taggingPath.HandleFunc("/members/{member}", tag.GetTagsByMemberHandler).Methods("GET").Name("Get-tags-by-member")

	paths = append(paths, taggingPath)

	for _, p := range paths {
		if s.TestOnly() {
			p.Use(s.NoAuthMiddleware)
		} else {
			p.Use(s.XW_XconfServer.SpanMiddleware)
			p.Use(s.AuthValidationMiddleware)
		}
	}
}

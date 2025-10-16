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

	// New V2 endpoints with improved scalability and pagination
	taggingPath.HandleFunc("/", tag.GetAllTagsV2Handler).Methods("GET").Name("Get-all-tags-v2")
	taggingPath.HandleFunc("/{tag}", tag.GetTagByIdV2Handler).Methods("GET").Name("Get-tag-by-id-v2")
	taggingPath.HandleFunc("/{tag}/members", tag.AddMembersToTagV2Handler).Methods("PUT").Name("Add-members-to-tag-v2")
	taggingPath.HandleFunc("/{tag}", tag.DeleteTagV2Handler).Methods("DELETE").Name("Delete-tag-v2")
	taggingPath.HandleFunc("/{tag}/members", tag.RemoveMembersFromTagV2Handler).Methods("DELETE").Name("Remove-members-from-tag-v2")
	taggingPath.HandleFunc("/{tag}/members/{member}", tag.RemoveMemberFromTagV2Handler).Methods("DELETE").Name("Remove-member-from-tag-v2")

	taggingPath.HandleFunc("/{tag}/members", tag.GetTagMembersV2Handler).Methods("GET").Name("Get-tag-members")

	//will remain the same
	taggingPath.HandleFunc("/members/{member}", tag.GetTagsByMemberHandler).Methods("GET").Name("Get-tags-by-member")

	// Migration endpoint
	taggingPath.HandleFunc("/migrate/v1-to-v2", tag.MigrateV1ToV2Handler).Methods("POST").Name("Migrate-v1-to-v2")

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

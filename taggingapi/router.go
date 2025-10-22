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

	taggingPath.HandleFunc("", tag.GetAllTagsV2Handler).Methods("GET").Name("Get-all-tags")
	taggingPath.HandleFunc("/{tag}", tag.GetTagByIdV2Handler).Methods("GET").Name("Get-tag-by-id")
	taggingPath.HandleFunc("/{tag}/members", tag.AddMembersToTagV2Handler).Methods("PUT").Name("Add-members-to-tag")
	taggingPath.HandleFunc("/{tag}", tag.DeleteTagV2Handler).Methods("DELETE").Name("Delete-tag-v2")
	taggingPath.HandleFunc("/{tag}/members", tag.RemoveMembersFromTagV2Handler).Methods("DELETE").Name("Remove-members-from-tag")
	taggingPath.HandleFunc("/{tag}/members/{member}", tag.RemoveMemberFromTagV2Handler).Methods("DELETE").Name("Remove-member-from-tag")

	taggingPath.HandleFunc("/{tag}/members", tag.GetTagMembersV2Handler).Methods("GET").Name("Get-tag-members")

	taggingPath.HandleFunc("/members/{member}", tag.GetTagsByMemberHandler).Methods("GET").Name("Get-tags-by-member")

	// Migration endpoints
	taggingPath.HandleFunc("/migrate", tag.MigrateV1ToV2Handler).Methods("POST").Name("Migrate-v1-to-v2")
	taggingPath.HandleFunc("/migrate/status", tag.MigrationStatusHandler).Methods("GET").Name("Migration-status")

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

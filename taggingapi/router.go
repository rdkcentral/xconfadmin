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
	taggingPath.HandleFunc("/{tag}", tag.DeleteTagHandler).Methods("DELETE").Name("Delete-tag")
	taggingPath.HandleFunc("/{tag}/noprefix", tag.DeleteTagFromXconfWithoutPrefixHandler).Methods("DELETE").Name("Delete-tag-from-xconf")
	taggingPath.HandleFunc("/{tag}/members", tag.AddMembersToTagHandler).Methods("PUT").Name("Add-members-to-tag")
	taggingPath.HandleFunc("/{tag}/members/{member}", tag.RemoveMemberFromTagHandler).Methods("DELETE").Name("Remove-member-from-tag")
	taggingPath.HandleFunc("/{tag}/members", tag.GetTagMembersHandler).Methods("GET").Name("Get-tag-members")
	taggingPath.HandleFunc("/members/{member}", tag.GetTagsByMemberHandler).Methods("GET").Name("Get-tags-by-member")
	taggingPath.HandleFunc("/{tag}/members", tag.RemoveMembersFromTagHandler).Methods("DELETE").Name("Remove-members-from-tag")
	//taggingPath.HandleFunc("/members/{member}/percentages", tag.GetTagsByMemberPercentageHandler).Methods("GET").Name("Get-tags-by-member-percentage")
	//taggingPath.HandleFunc("/{tag}/members/percentages/ranges/{startRange}/{endRange}", tag.AddMemberPercentageToTagHandler).Methods("PUT").Name("Add-account-percentage-to-tag")
	//taggingPath.HandleFunc("/{tag}/members/percentages/ranges", tag.CleanPercentageRangeHandler).Methods("DELETE").Name("Remove-percentage-members-from-tag")
	taggingPath.HandleFunc("/members/{member}/percentages/calculation", tag.CalculatePercentageValueHandler).Methods("GET").Name("Calculate-percentage-value-for-member")

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

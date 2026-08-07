package taggingapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	xhttp "github.com/rdkcentral/xconfadmin/http"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// resolveRouteName returns the name of the route a request matches, or "" if
// nothing matches. Typed and untyped routes overlap in shape and mux resolves
// the overlap by registration order, so nothing else would catch a reordering
// or a mux upgrade silently changing which handler serves a URL.
func resolveRouteName(t *testing.T, router *mux.Router, method, path string) string {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	var match mux.RouteMatch
	if !router.Match(req, &match) || match.Route == nil {
		return ""
	}
	return match.Route.GetName()
}

func newTaggingRouter(t *testing.T) *mux.Router {
	t.Helper()
	router := mux.NewRouter()
	routeTaggingServiceApis(router, &xhttp.WebconfigServer{})
	return router
}

// The typed routes must serve their own URLs, and the untyped routes must keep
// serving theirs unchanged.
func TestTaggingRoutes_Resolution(t *testing.T) {
	router := newTaggingRouter(t)

	testCases := []struct {
		name     string
		method   string
		path     string
		expected string
	}{
		// Untyped routes — unchanged from before typed routes existed.
		{"list tags", http.MethodGet, "/taggingService/tags", "Get-all-tags"},
		{"get tag", http.MethodGet, "/taggingService/tags/promo2026", "Get-tag-by-id"},
		{"add members", http.MethodPut, "/taggingService/tags/promo2026/members", "Add-members-to-tag"},
		{"delete tag", http.MethodDelete, "/taggingService/tags/promo2026", "Delete-tag-v2"},
		{"remove members", http.MethodDelete, "/taggingService/tags/promo2026/members", "Remove-members-from-tag"},
		{"remove one member", http.MethodDelete, "/taggingService/tags/promo2026/members/AABBCCDDEEFF", "Remove-member-from-tag"},
		{"get members", http.MethodGet, "/taggingService/tags/promo2026/members", "Get-tag-members"},
		{"reverse lookup", http.MethodGet, "/taggingService/tags/members/AABBCCDDEEFF", "Get-tags-by-member"},
		{"reverse lookup values", http.MethodGet, "/taggingService/tags/members/AABBCCDDEEFF/values", "Get-tags-with-values-by-member"},

		// Typed account routes.
		{"account list", http.MethodGet, "/taggingService/tags/account", "Get-all-tags-typed"},
		{"account get tag", http.MethodGet, "/taggingService/tags/account/promo2026", "Get-tag-by-id-typed"},
		{"account add members", http.MethodPut, "/taggingService/tags/account/promo2026/members", "Add-members-to-tag-typed"},
		{"account delete tag", http.MethodDelete, "/taggingService/tags/account/promo2026", "Delete-tag-typed"},
		{"account remove members", http.MethodDelete, "/taggingService/tags/account/promo2026/members", "Remove-members-from-tag-typed"},
		{"account remove one member", http.MethodDelete, "/taggingService/tags/account/promo2026/members/2846573900878987927", "Remove-member-from-tag-typed"},
		{"account get members", http.MethodGet, "/taggingService/tags/account/promo2026/members", "Get-tag-members-typed"},
		{"account reverse lookup", http.MethodGet, "/taggingService/tags/account/members/2846573900878987927", "Get-tags-by-member-typed"},
		{"account reverse lookup values", http.MethodGet, "/taggingService/tags/account/members/2846573900878987927/values", "Get-tags-with-values-by-member-typed"},

		// Typed mac routes.
		{"mac list", http.MethodGet, "/taggingService/tags/mac", "Get-all-tags-typed"},
		{"mac add members", http.MethodPut, "/taggingService/tags/mac/promo2026/members", "Add-members-to-tag-typed"},
		{"mac get tag", http.MethodGet, "/taggingService/tags/mac/promo2026", "Get-tag-by-id-typed"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, resolveRouteName(t, router, tc.method, tc.path))
		})
	}
}

// Only "mac" and "account" activate the typed subrouter. Anything else must fall
// through to the untyped routes, otherwise a tag whose name happened to look
// like a type would become unreachable.
func TestTaggingRoutes_UnknownTypeFallsThroughToUntyped(t *testing.T) {
	router := newTaggingRouter(t)

	testCases := []struct {
		name     string
		method   string
		path     string
		expected string
	}{
		{"device is not a type", http.MethodGet, "/taggingService/tags/device", "Get-tag-by-id"},
		{"generic is not a type", http.MethodGet, "/taggingService/tags/generic/members", "Get-tag-members"},
		{"macaroni is not mac", http.MethodPut, "/taggingService/tags/macaroni/members", "Add-members-to-tag"},
		{"accounts is not account", http.MethodDelete, "/taggingService/tags/accounts", "Delete-tag-v2"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, resolveRouteName(t, router, tc.method, tc.path))
		})
	}
}

// These URLs are ambiguous between the two route sets; typed wins by being
// registered first, which is why "mac", "account" and "members" are rejected as
// tag ids (tag.validateTagId). If a change flips any of these, that reserved-id
// list is what needs revisiting.
func TestTaggingRoutes_AmbiguousUrlsResolveToTyped(t *testing.T) {
	router := newTaggingRouter(t)

	testCases := []struct {
		name     string
		method   string
		path     string
		expected string
		shadowed string
	}{
		{
			name:     "bare type segment",
			method:   http.MethodGet,
			path:     "/taggingService/tags/account",
			expected: "Get-all-tags-typed",
			shadowed: "GET /tags/{tag} for a tag named 'account'",
		},
		{
			name:     "type plus members",
			method:   http.MethodGet,
			path:     "/taggingService/tags/account/members",
			expected: "Get-tag-by-id-typed",
			shadowed: "GET /tags/{tag}/members for a tag named 'account'",
		},
		{
			// The destructive one: a bulk member removal would otherwise be
			// served as an async whole-tag delete of a different tag.
			name:     "delete type plus members",
			method:   http.MethodDelete,
			path:     "/taggingService/tags/account/members",
			expected: "Delete-tag-typed",
			shadowed: "DELETE /tags/{tag}/members for a tag named 'account'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, resolveRouteName(t, router, tc.method, tc.path),
				"this URL shadows: %s", tc.shadowed)
		})
	}
}

// A method with no typed counterpart must still reach the untyped handler rather
// than 405-ing, which is the mux fall-through behavior this design depends on.
func TestTaggingRoutes_MethodMismatchFallsThrough(t *testing.T) {
	router := newTaggingRouter(t)

	// There is no typed "PUT /{tagType}" route, so this must land on the untyped
	// PUT /{tag}/members with tag="account".
	assert.Equal(t, "Add-members-to-tag",
		resolveRouteName(t, router, http.MethodPut, "/taggingService/tags/account/members"))
}

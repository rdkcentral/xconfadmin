package http

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBulkTagMemberPayload(t *testing.T) {
	bulk := []string{
		"/taggingService/tags/my-tag/members",
		"/taggingService/tags/my.tag-2/members",
	}
	for _, path := range bulk {
		r := httptest.NewRequest("PUT", path, nil)
		assert.True(t, isBulkTagMemberPayload(r), path)
	}

	notBulk := []string{
		"/taggingService/tags",
		"/taggingService/tags/my-tag",
		"/taggingService/tags/members/AA:BB:CC:DD:EE:01",
		"/xconfAdminService/firmwarerule",
		"/api/v1/token",
	}
	for _, path := range notBulk {
		r := httptest.NewRequest("POST", path, nil)
		assert.False(t, isBulkTagMemberPayload(r), path)
	}
}

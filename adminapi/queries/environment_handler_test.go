package queries_test

import (
	"net/http"
	"testing"

	"github.com/rdkcentral/xconfadmin/adminapi/queries"
)

func TestUpdateEnvironmentHandler(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		w http.ResponseWriter
		r *http.Request
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries.UpdateEnvironmentHandler(tt.w, tt.r)
		})
	}
}

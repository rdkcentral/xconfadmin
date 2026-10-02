package tag

import (
	"errors"
	"net/http"
	"testing"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/stretchr/testify/assert"
)

// failingReadDbClient fails every read; the embedded interface panics on
// any other database use.
type failingReadDbClient struct {
	db.DatabaseClient
	err error
}

func (c failingReadDbClient) QueryXconfDataRows(string, ...string) ([]map[string]interface{}, error) {
	return nil, c.err
}

// A driver error can name hosts and keyspaces; tag APIs write read errors
// into the response body, so only the generic error may come back.
func TestTagReadErrorDoesNotLeakDriverDetail(t *testing.T) {
	setupTestEnvironment()
	old := db.GetDatabaseClient()
	db.SetDatabaseClient(failingReadDbClient{
		err: errors.New("gocql: no hosts available in the pool: 10.0.0.7:9042 keyspace ApplicationsDiscoveryDataService"),
	})
	t.Cleanup(func() { db.SetDatabaseClient(old) })

	_, _, err := GetTagById("some-tag")
	if assert.Error(t, err) {
		assert.NotContains(t, err.Error(), "10.0.0.7")
		assert.NotContains(t, err.Error(), "gocql")
		assert.ErrorIs(t, err, errTagStoreRead)
		assert.Equal(t, http.StatusInternalServerError, xwcommon.GetXconfErrorStatusCode(err))
	}

	_, err = GetAllTagIds()
	if assert.Error(t, err) {
		assert.NotContains(t, err.Error(), "10.0.0.7")
		assert.ErrorIs(t, err, errTagStoreRead)
		assert.Equal(t, http.StatusInternalServerError, xwcommon.GetXconfErrorStatusCode(err))
	}
}

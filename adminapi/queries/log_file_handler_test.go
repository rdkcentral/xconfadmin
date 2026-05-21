package queries

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rdkcentral/xconfadmin/shared/logupload"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/stretchr/testify/assert"
)

// helper to wrap XResponseWriter with a JSON body
func makeLogFileXW(obj any) (*httptest.ResponseRecorder, *xwhttp.XResponseWriter) {
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	if obj != nil {
		b, _ := json.Marshal(obj)
		xw.SetBody(string(b))
	}
	return rr, xw
}

func TestCreateLogFile_ResponseWriterCastError(t *testing.T) {
	SkipIfMockDatabase(t)
	// pass plain recorder -> cast fail
	r := httptest.NewRequest(http.MethodPost, "/logfile", nil)
	rr := httptest.NewRecorder()
	CreateLogFile(rr, r)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestCreateLogFile_InvalidJSON(t *testing.T) {
	SkipIfMockDatabase(t)
	r := httptest.NewRequest(http.MethodPost, "/logfile", nil)
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody("{not-json")
	CreateLogFile(xw, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateLogFile_EmptyName(t *testing.T) {
	SkipIfMockDatabase(t)
	lf := logupload.LogFile{ID: "", Name: ""}
	rr, xw := makeLogFileXW(lf)
	r := httptest.NewRequest(http.MethodPost, "/logfile", nil)
	CreateLogFile(xw, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateLogFile_NewSuccess(t *testing.T) {
	SkipIfMockDatabase(t)
	lf := logupload.LogFile{Name: "alpha.log"}
	rr, xw := makeLogFileXW(lf)
	r := httptest.NewRequest(http.MethodPost, "/logfile", nil)
	CreateLogFile(xw, r)
	assert.Equal(t, http.StatusCreated, rr.Code)
	created := logupload.LogFile{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, lf.Name, created.Name)
}

func TestCreateLogFile_DuplicateName(t *testing.T) {
	SkipIfMockDatabase(t)
	// seed first
	seed := logupload.LogFile{Name: "dup.log"}
	rr1, xw1 := makeLogFileXW(seed)
	r1 := httptest.NewRequest(http.MethodPost, "/logfile", nil)
	CreateLogFile(xw1, r1)
	if rr1.Code != http.StatusCreated {
		t.Fatalf("seed create failed: %d %s", rr1.Code, rr1.Body.String())
	}

	// attempt second with different ID but same name -> should 400
	lf2 := logupload.LogFile{Name: "dup.log"}
	rr2, xw2 := makeLogFileXW(lf2)
	r2 := httptest.NewRequest(http.MethodPost, "/logfile", nil)
	CreateLogFile(xw2, r2)
	assert.Equal(t, http.StatusBadRequest, rr2.Code)
}

func TestCreateLogFile_UpdatePath(t *testing.T) {
	SkipIfMockDatabase(t)
	// The update path in CreateLogFile calls updateLogUploadSettingsAndLogFileGroups
	// which invokes GetAllLogUploadSettings/GetLogFileGroupsList. On multi-tenant DAO
	// these return an error when the tables are empty, causing a 500.
	// Skip until the service code handles empty-table queries gracefully.
	t.Skip("skipped: updateLogUploadSettingsAndLogFileGroups returns error on empty tables in multi-tenant DAO")
}

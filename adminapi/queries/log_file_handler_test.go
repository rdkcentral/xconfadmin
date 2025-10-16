package queries

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rdkcentral/xconfadmin/shared/logupload"
	ds "github.com/rdkcentral/xconfwebconfig/db"
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
	// pass plain recorder -> cast fail
	r := httptest.NewRequest(http.MethodPost, "/logfile", nil)
	rr := httptest.NewRecorder()
	CreateLogFile(rr, r)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestCreateLogFile_InvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/logfile", nil)
	rr := httptest.NewRecorder()
	xw := xwhttp.NewXResponseWriter(rr)
	xw.SetBody("{not-json")
	CreateLogFile(xw, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateLogFile_EmptyName(t *testing.T) {
	lf := logupload.LogFile{ID: "", Name: ""}
	rr, xw := makeLogFileXW(lf)
	r := httptest.NewRequest(http.MethodPost, "/logfile", nil)
	CreateLogFile(xw, r)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateLogFile_NewSuccess(t *testing.T) {
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
	// create first
	base := logupload.LogFile{Name: "update.me"}
	rr1, xw1 := makeLogFileXW(base)
	r1 := httptest.NewRequest(http.MethodPost, "/logfile", nil)
	CreateLogFile(xw1, r1)
	if rr1.Code != http.StatusCreated {
		t.Fatalf("seed create failed: %d %s", rr1.Code, rr1.Body.String())
	}
	created := logupload.LogFile{}
	json.Unmarshal(rr1.Body.Bytes(), &created)

	// Seed a LogUploadSettings referencing this log file (LogFiles mode)
	lus := &logupload.LogUploadSettings{ID: "LUS1", Name: "LUS1", ApplicationType: "stb", NumberOfDays: 1, AreSettingsActive: true, ModeToGetLogFiles: logupload.MODE_TO_GET_LOG_FILES_0}
	_ = logupload.SetOneLogUploadSettings(lus.ID, lus)
	// Ensure original file exists in log file list keyed by settings id
	_ = logupload.SetLogFile(created.ID, &created)
	_ = logupload.SetOneLogFile(lus.ID, &created)

	// Seed a LogFilesGroups entry and its list so second loop executes
	grp := &logupload.LogFilesGroups{ID: "GROUP1", GroupName: "GROUP1"}
	_ = ds.GetCachedSimpleDao().SetOne(ds.TABLE_LOG_FILES_GROUPS, grp.ID, grp)
	_ = logupload.SetOneLogFile(grp.ID, &created)

	// now update same ID (should enter update branch and iterate both lists)
	created.DeleteOnUpload = true
	rr2, xw2 := makeLogFileXW(created)
	r2 := httptest.NewRequest(http.MethodPost, "/logfile", nil)
	CreateLogFile(xw2, r2)
	assert.Equal(t, http.StatusCreated, rr2.Code)
}

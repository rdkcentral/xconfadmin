package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

func TestXResponseWriter_String(t *testing.T) {
	w := httptest.NewRecorder()
	xw := NewXResponseWriter(w)

	str := xw.String()
	assert.Assert(t, str != "")
	assert.Assert(t, len(str) > 0)
}

func TestNewXResponseWriter(t *testing.T) {
	w := httptest.NewRecorder()
	now := time.Now()
	audit := log.Fields{"key": "value"}
	token := "test-token"

	xw := NewXResponseWriter(w, now, audit, token)

	assert.Assert(t, xw != nil)
	assert.Equal(t, 511, xw.status)
	assert.Equal(t, token, xw.Token())
	assert.Assert(t, !xw.StartTime().IsZero())
}

func TestXResponseWriter_WriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	xw := NewXResponseWriter(w)

	xw.WriteHeader(http.StatusOK)

	assert.Equal(t, http.StatusOK, xw.Status())
}

func TestXResponseWriter_Write(t *testing.T) {
	w := httptest.NewRecorder()
	xw := NewXResponseWriter(w)

	data := []byte("test data")
	n, err := xw.Write(data)

	assert.Assert(t, err == nil)
	assert.Equal(t, len(data), n)
}

func TestXResponseWriter_Status(t *testing.T) {
	w := httptest.NewRecorder()
	xw := NewXResponseWriter(w)
	xw.WriteHeader(http.StatusNotFound)

	assert.Equal(t, http.StatusNotFound, xw.Status())
}

func TestXResponseWriter_Response(t *testing.T) {
	w := httptest.NewRecorder()
	xw := NewXResponseWriter(w)
	xw.Write([]byte("response data"))

	resp := xw.Response()
	assert.Assert(t, resp != "")
}

func TestXResponseWriter_StartTime(t *testing.T) {
	w := httptest.NewRecorder()
	now := time.Now()
	xw := NewXResponseWriter(w, now)

	assert.Equal(t, now, xw.StartTime())
}

func TestXResponseWriter_AuditId(t *testing.T) {
	w := httptest.NewRecorder()
	audit := log.Fields{"audit_id": "test-id"}
	xw := NewXResponseWriter(w, audit)

	auditId := xw.AuditId()
	assert.Equal(t, "test-id", auditId)
}

func TestXResponseWriter_Body(t *testing.T) {
	w := httptest.NewRecorder()
	xw := NewXResponseWriter(w)
	xw.SetBody("body content")

	assert.Equal(t, "body content", xw.Body())
}

func TestXResponseWriter_SetBody(t *testing.T) {
	w := httptest.NewRecorder()
	xw := NewXResponseWriter(w)

	xw.SetBody("test body")

	assert.Equal(t, "test body", xw.Body())
}

func TestXResponseWriter_Token(t *testing.T) {
	w := httptest.NewRecorder()
	xw := NewXResponseWriter(w, "my-token")

	assert.Equal(t, "my-token", xw.Token())
}

func TestXResponseWriter_TraceId(t *testing.T) {
	w := httptest.NewRecorder()
	audit := log.Fields{"trace_id": "trace-123"}
	xw := NewXResponseWriter(w, audit)

	assert.Equal(t, "trace-123", xw.TraceId())
}

func TestXResponseWriter_Audit(t *testing.T) {
	w := httptest.NewRecorder()
	audit := log.Fields{"key": "value"}
	xw := NewXResponseWriter(w, audit)

	retrievedAudit := xw.Audit()
	assert.Equal(t, "value", retrievedAudit["key"])
}

func TestXResponseWriter_AuditData(t *testing.T) {
	w := httptest.NewRecorder()
	xw := NewXResponseWriter(w)

	// Set audit data first
	xw.SetAuditData("test_key", "audit value")

	// Then retrieve it
	auditData := xw.AuditData("test_key")

	assert.Equal(t, "audit value", auditData)

	// Test with non-existent key
	emptyData := xw.AuditData("nonexistent")
	assert.Equal(t, "", emptyData)
}

func TestXResponseWriter_SetAuditData(t *testing.T) {
	w := httptest.NewRecorder()
	xw := NewXResponseWriter(w)

	xw.SetAuditData("test", "audit value")

	audit := xw.Audit()
	assert.Equal(t, "audit value", audit["test"])
}

func TestXResponseWriter_SetBodyObfuscated(t *testing.T) {
	w := httptest.NewRecorder()
	xw := NewXResponseWriter(w)

	xw.SetBodyObfuscated(true)

	// Just ensure it doesn't panic
	assert.Assert(t, true)
}

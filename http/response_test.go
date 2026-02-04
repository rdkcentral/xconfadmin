package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/assert"
)

func TestWriteOkResponse(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	data := map[string]string{"key": "value"}
	WriteOkResponse(w, r, data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Assert(t, w.Body.Len() > 0)
	assert.Equal(t, "application/json", w.Header().Get("Content-type"))
}

func TestWriteOkResponseByTemplate(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	WriteOkResponseByTemplate(w, r, `{"test":"data"}`)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Assert(t, w.Body.Len() > 0)
}

func TestWriteTR181Response(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	WriteTR181Response(w, r, `{"param":"value"}`, "v1.0")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "v1.0", w.Header().Get("ETag"))
	assert.Assert(t, w.Body.Len() > 0)
}

func TestWriteErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()

	err := errors.New("test error")
	WriteErrorResponse(w, http.StatusBadRequest, err)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Assert(t, w.Body.Len() > 0)
}

func TestWriteAdminErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()

	WriteAdminErrorResponse(w, http.StatusNotFound, "not found")

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Assert(t, w.Body.Len() > 0)
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()

	err := errors.New("error message")
	Error(w, err)

	assert.Assert(t, w.Code > 0)
	assert.Assert(t, w.Body.Len() > 0)
}

func TestAdminError(t *testing.T) {
	w := httptest.NewRecorder()

	err := errors.New("bad request")
	AdminError(w, err)

	assert.Assert(t, w.Code > 0)
	assert.Assert(t, w.Body.Len() > 0)
}

func TestWriteXconfResponse(t *testing.T) {
	w := httptest.NewRecorder()

	WriteXconfResponse(w, http.StatusOK, []byte("test data"))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test data", w.Body.String())
}

func TestWriteXconfErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()

	err := errors.New("error")
	WriteXconfErrorResponse(w, err)

	assert.Assert(t, w.Code > 0)
	assert.Assert(t, w.Body.Len() > 0)
}

func TestWriteXconfResponseAsText(t *testing.T) {
	w := httptest.NewRecorder()

	WriteXconfResponseAsText(w, http.StatusOK, []byte("text data"))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
	assert.Equal(t, "text data", w.Body.String())
}

func TestWriteXconfResponseWithHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	headers := map[string]string{
		"X-Custom-Header": "custom-value",
	}

	WriteXconfResponseWithHeaders(w, headers, http.StatusOK, []byte("data"))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "custom-value", w.Header().Get("X-Custom-Header"))
}

func TestWriteXconfResponseHtmlWithHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	headers := map[string]string{
		"X-Test": "test",
	}

	WriteXconfResponseHtmlWithHeaders(w, headers, http.StatusOK, []byte("<html></html>"))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Assert(t, w.Header().Get("Content-Type") != "")
	assert.Equal(t, "test", w.Header().Get("X-Test"))
}

func TestCreateContentDispositionHeader(t *testing.T) {
	header := CreateContentDispositionHeader("test")
	assert.Assert(t, header != nil)
	assert.Assert(t, header["Content-Disposition"] != "")
}

func TestCreateNumberOfItemsHttpHeaders(t *testing.T) {
	headers := CreateNumberOfItemsHttpHeaders(42)
	assert.Equal(t, "42", headers["numberOfItems"])
}

func TestReturnJsonResponse(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("Accept", "application/json")
	data := map[string]string{"key": "value"}

	result, err := ReturnJsonResponse(data, r)

	assert.Assert(t, err == nil)
	assert.Assert(t, len(result) > 0)
}

func TestContextTypeHeader(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)

	contentType := ContextTypeHeader(r)

	assert.Equal(t, "application/json:charset=UTF-8", contentType)
}

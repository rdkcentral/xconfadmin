package http

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/rdkcentral/xconfadmin/common"
	"github.com/rdkcentral/xconfadmin/util"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
)

const (
	OkResponseTemplate = `{"status":200,"message":"OK","data":%v}`

	// TODO, this should be retired
	TR181ResponseTemplate = `{"parameters":%v,"version":"%v"}`
	TYPE_409              = "EntityConflictException"
	TYPE_400              = "ValidationRuntimeException"
	TYPE_404              = "EntityNotFoundException"
	TYPE_500              = "InternalServerErrorException"
	TYPE_501              = "NotImplementedException"
	TYPE_415              = "UnsupportedMediaTypeException"
)

type EntityMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type HttpAdminErrorResponse struct {
	Status  int    `json:"status"`
	Type    string `json:"type,omitempty"`
	Message string `json:"message"`
}

func writeByMarshal(w http.ResponseWriter, status int, o interface{}) {
	addMoracideTagsAsResponseHeaders(w)
	if rbytes, err := util.JSONMarshal(o); err == nil {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(status)
		w.Write(rbytes)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		LogError(w, err)
	}
}

// helper function to write a json response into ResponseWriter
func WriteOkResponse(w http.ResponseWriter, r *http.Request, data interface{}) {
	resp := common.HttpResponse{
		Status:  http.StatusOK,
		Message: http.StatusText(http.StatusOK),
		Data:    data,
	}
	writeByMarshal(w, http.StatusOK, resp)
}

func WriteOkResponseByTemplate(w http.ResponseWriter, r *http.Request, dataStr string) {
	rbytes := []byte(fmt.Sprintf(OkResponseTemplate, dataStr))
	w.Header().Set("Content-type", "application/json")
	addMoracideTagsAsResponseHeaders(w)
	w.WriteHeader(http.StatusOK)
	w.Write(rbytes)
}

func WriteTR181Response(w http.ResponseWriter, r *http.Request, params string, version string) {
	w.Header().Set("Content-type", "application/json")
	w.Header().Set("ETag", version)
	addMoracideTagsAsResponseHeaders(w)
	w.WriteHeader(http.StatusOK)
	rbytes := []byte(fmt.Sprintf(TR181ResponseTemplate, params, version))
	w.Write(rbytes)
}

// this is used to return default tr-181 payload while the cpe is not in the db
func WriteContentTypeAndResponse(w http.ResponseWriter, r *http.Request, rbytes []byte, version string, contentType string) {
	w.Header().Set("Content-type", contentType)
	w.Header().Set("ETag", version)
	addMoracideTagsAsResponseHeaders(w)
	w.WriteHeader(http.StatusOK)
	w.Write(rbytes)
}

// helper function to write a failure json response into ResponseWriter
func WriteErrorResponse(w http.ResponseWriter, status int, err error) {
	errstr := ""
	if err != nil {
		errstr = err.Error()
	}
	resp := common.HttpErrorResponse{
		Status:  status,
		Message: http.StatusText(status),
		Errors:  errstr,
	}
	writeByMarshal(w, status, resp)
}

// helper function to write a failure json response matching xconf java admin response
func WriteAdminErrorResponse(w http.ResponseWriter, status int, errMsg string) {
	typeMsg := ""
	switch status {
	case 400:
		typeMsg = TYPE_400
	case 409:
		typeMsg = TYPE_409
	case 404:
		typeMsg = TYPE_404
	case 500:
		typeMsg = TYPE_500
	case 501:
		typeMsg = TYPE_501
	case 415:
		typeMsg = TYPE_415
	}
	resp := common.HttpAdminErrorResponse{
		Status:  status,
		Type:    typeMsg,
		Message: errMsg,
	}
	writeByMarshal(w, status, resp)
}

func Error(w http.ResponseWriter, err error) {
	status := xwcommon.GetXconfErrorStatusCode(err)
	switch status {
	case http.StatusNoContent, http.StatusNotModified, http.StatusForbidden:
		addMoracideTagsAsResponseHeaders(w)
		w.WriteHeader(status)
	default:
		WriteErrorResponse(w, status, err)
	}
}

func AdminError(w http.ResponseWriter, err error) {
	status := xwcommon.GetXconfErrorStatusCode(err)
	WriteAdminErrorResponse(w, status, err.Error())
}

func WriteResponseBytes(w http.ResponseWriter, rbytes []byte, statusCode int, vargs ...string) {
	if len(vargs) > 0 {
		w.Header().Set("Content-type", vargs[0])
	}
	addMoracideTagsAsResponseHeaders(w)
	w.WriteHeader(statusCode)
	w.Write(rbytes)
}

func WriteXconfResponse(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-type", "application/json")
	addMoracideTagsAsResponseHeaders(w)
	w.WriteHeader(status)
	w.Write(data)
}

func WriteXconfErrorResponse(w http.ResponseWriter, err error) {
	status := xwcommon.GetXconfErrorStatusCode(err)
	w.Header().Set("Content-type", "application/json")
	addMoracideTagsAsResponseHeaders(w)
	w.WriteHeader(status)
	w.Write([]byte(err.Error()))
}

func WriteXconfResponseAsText(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-type", "text/plain")
	addMoracideTagsAsResponseHeaders(w)
	w.WriteHeader(status)
	w.Write(data)
}

func WriteXconfResponseWithHeaders(w http.ResponseWriter, headers map[string]string, status int, data []byte) {
	w.Header().Set("Content-type", "application/json")
	for k, v := range headers {
		w.Header()[k] = []string{v}
	}
	addMoracideTagsAsResponseHeaders(w)
	w.WriteHeader(status)
	w.Write(data)
}

func WriteXconfResponseHtmlWithHeaders(w http.ResponseWriter, headers map[string]string, status int, data []byte) {
	w.Header().Set("Content-type", "text/html; charset=iso-8859-1")
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	addMoracideTagsAsResponseHeaders(w)
	w.WriteHeader(status)
	w.Write(data)
}

func CreateContentDispositionHeader(fileName string) map[string]string {
	return map[string]string{"Content-Disposition": fmt.Sprintf("attachment; filename=%s.json", escapeXml(fileName))}
}

func CreateNumberOfItemsHttpHeaders(size int) map[string]string {
	return map[string]string{"numberOfItems": strconv.Itoa(size)}
}

func escapeXml(str string) string {
	var buffer bytes.Buffer
	xml.EscapeText(&buffer, []byte(str))
	return buffer.String()
}

// ReturnJsonResponse - return JSON response to api
func ReturnJsonResponse(res interface{}, r *http.Request) ([]byte, error) {
	acceptStr := r.Header.Get("Accept")
	if acceptStr == "" {
		if data, err := util.JSONMarshal(res); err != nil {
			return nil, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, fmt.Sprintf("JSON marshal error: %v", err))
		} else {
			return data, nil
		}
	}
	acceptTokens := strings.Split(acceptStr, ",")
	for _, acceptVal := range acceptTokens {
		/* TODO:  Sort out the difference in invocation for export-flavored endpoints Vs others.

		export-flavored endpoints do not show up in developer tools because of the way they are invoked.

		function getFirmwareRuleNamesByTemplate(templateId) {
			return $http.get(API_URL + 'byTemplate/' + templateId + '/names');
		}

		function exportFirmwareRule(id) {
			window.open(API_URL + id + '/?export');
		}

		Also, the header.Accept is set differently for export-flavored endpoints.

		Export flavored endpoints set Header.Accept to
		text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,star/star;q=0.8,application/signed-exchange;v=b3tq=0.9"

		But other endpoints set it to Accept: application/json, text/plain, star/star
		*/

		if strings.Contains(strings.ToLower(acceptVal), "*/*") || strings.Contains(strings.ToLower(acceptVal), "application/json") {
			if data, err := util.JSONMarshal(res); err != nil {
				return nil, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, fmt.Sprintf("JSON marshal error: %v", err))
			} else {
				return data, nil
			}
		}
	}
	return nil, xwcommon.NewRemoteErrorAS(http.StatusNotAcceptable, "At this time only JSON input/output is supported")
}

func ContextTypeHeader(r *http.Request) string {
	return fmt.Sprintf("%s:%s", "application/json", "charset=UTF-8")
}

func addMoracideTagsAsResponseHeaders(w http.ResponseWriter) {
	xw, ok := w.(*XResponseWriter)
	if !ok {
		return
	}
	fields := xw.Audit()
	if fields == nil {
		return
	}

	moracide := xwcommon.FieldsGetString(fields, "resp_moracide_tag")
	if len(moracide) == 0 {
		moracide = xwcommon.FieldsGetString(fields, "req_moracide_tag")
	}
	if len(moracide) > 0 {
		w.Header().Set(xwcommon.HeaderMoracide, moracide)
	}
}

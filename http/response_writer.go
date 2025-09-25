package http

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type XResponseWriter struct {
	http.ResponseWriter
	status         int
	length         int
	response       string
	startTime      time.Time
	body           string
	token          string
	audit          log.Fields
	bodyObfuscated bool
}

func (w *XResponseWriter) String() string {
	ret := fmt.Sprintf("status=%v, length=%v, response=%v, startTime=%v, audit=%v",
		w.status, w.length, w.response, w.startTime, w.audit)
	return ret
}

func NewXResponseWriter(w http.ResponseWriter, vargs ...interface{}) *XResponseWriter {
	var audit log.Fields
	var startTime time.Time
	var token string

	for _, v := range vargs {
		switch ty := v.(type) {
		case time.Time:
			startTime = ty
		case log.Fields:
			audit = ty
		case string:
			token = ty
		}
	}

	if audit == nil {
		audit = make(log.Fields)
	}

	return &XResponseWriter{
		ResponseWriter: w,
		status:         511, // setting default status code as 511
		length:         0,
		response:       "",
		startTime:      startTime,
		token:          token,
		audit:          audit,
	}
}

// interface/override
func (w *XResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *XResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	if err != nil {
		return n, err
	}
	w.length += n
	w.response = string(b)
	return n, nil
}

// get/set
func (w *XResponseWriter) Status() int {
	return w.status
}

func (w *XResponseWriter) Response() string {
	return w.response
}

func (w *XResponseWriter) StartTime() time.Time {
	return w.startTime
}

func (w *XResponseWriter) AuditId() string {
	return w.AuditData("audit_id")
}

func (w *XResponseWriter) Body() string {
	return w.body
}

func (w *XResponseWriter) SetBody(body string) {
	w.body = body
}

func (w *XResponseWriter) Token() string {
	return w.token
}

func (w *XResponseWriter) TraceId() string {
	return w.AuditData("trace_id")
}

func (w *XResponseWriter) Audit() log.Fields {
	// return w.audit
	out := log.Fields{}
	for k, v := range w.audit {
		if k == "body" && w.bodyObfuscated {
			out[k] = "****"
		} else {
			out[k] = v
		}
	}
	return out
}

func (w *XResponseWriter) AuditData(k string) string {
	itf := w.audit[k]
	if itf != nil {
		return itf.(string)
	}
	return ""
}

func (w *XResponseWriter) SetAuditData(k string, v interface{}) {
	w.audit[k] = v
}

func (w *XResponseWriter) SetBodyObfuscated(obfuscated bool) {
	w.bodyObfuscated = obfuscated
}

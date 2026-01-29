package http

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"github.com/rdkcentral/xconfadmin/common"

	xwcommon "github.com/rdkcentral/xconfwebconfig/common"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
)

const (
	defaultConnectTimeout      = 30
	defaultReadTimeout         = 30
	defaultMaxIdleConnsPerHost = 100
	defaultKeepaliveTimeout    = 30
	defaultRetries             = 3
	defaultRetriesInMsecs      = 1000
	HttpGet                    = "GET"
	HttpPost                   = "POST"
	HttpDelete                 = "DELETE"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type HttpClient struct {
	*http.Client
	retries      int
	retryInMsecs int
}

func NewHttpClient(conf *configuration.Config, serviceName string, tlsConfig *tls.Config) *HttpClient {
	confKey := fmt.Sprintf("xconfwebconfig.%v.connect_timeout_in_secs", serviceName)
	connectTimeout := int(conf.GetInt32(confKey, defaultConnectTimeout))

	confKey = fmt.Sprintf("xconfwebconfig.%v.read_timeout_in_secs", serviceName)
	readTimeout := int(conf.GetInt32(confKey, defaultReadTimeout))

	confKey = fmt.Sprintf("xconfwebconfig.%v.max_idle_conns_per_host", serviceName)
	maxIdleConnsPerHost := int(conf.GetInt32(confKey, defaultMaxIdleConnsPerHost))

	confKey = fmt.Sprintf("xconfwebconfig.%v.keepalive_timeout_in_secs", serviceName)
	keepaliveTimeout := int(conf.GetInt32(confKey, defaultKeepaliveTimeout))

	confKey = fmt.Sprintf("xconfwebconfig.%v.retries", serviceName)
	retries := int(conf.GetInt32(confKey, defaultRetries))

	confKey = fmt.Sprintf("xconfwebconfig.%v.retry_in_msecs", serviceName)
	retryInMsecs := int(conf.GetInt32(confKey, defaultRetriesInMsecs))

	return &HttpClient{
		Client: &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   time.Duration(connectTimeout) * time.Second,
					KeepAlive: time.Duration(keepaliveTimeout) * time.Second,
				}).DialContext,
				MaxIdleConns:          0,
				MaxIdleConnsPerHost:   maxIdleConnsPerHost,
				IdleConnTimeout:       time.Duration(keepaliveTimeout) * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				TLSClientConfig:       tlsConfig,
			},
			Timeout: time.Duration(readTimeout) * time.Second,
		},
		retries:      retries,
		retryInMsecs: retryInMsecs,
	}
}

// Do is a wrapper around http.Client.Do
// Inputs: method, url, headers, body as bytes (bbytes), fields for logging (baseFields),
//
//	external service being called (loggerName), attempt # (retry)
//
// Returns: response body as bytes, any err, whether a retry is useful or not, and the status code
func (c *HttpClient) Do(method string, url string, headers map[string]string, bbytes []byte, baseFields log.Fields, loggerName string, retry int) ([]byte, error, bool, int) {
	// verify a response is received
	var respMoracideTagsFound bool
	defer func(found *bool) {
		if !*found {
			log.Debugf("http_client: no moracide tags in response")
		}
	}(&respMoracideTagsFound)

	// statusCode is used in metrics
	statusCode := http.StatusInternalServerError // Default status to return

	var req *http.Request
	var err error
	switch method {
	case "GET":
		req, err = http.NewRequest(method, url, nil)
	case "POST", "PATCH":
		req, err = http.NewRequest(method, url, bytes.NewReader(bbytes))
	case "DELETE":
		req, err = http.NewRequest(method, url, bytes.NewReader(bbytes))
	default:
		return nil, fmt.Errorf("method=%v", method), false, statusCode
	}

	if err != nil {
		return nil, err, true, statusCode
	}

	c.addMoracideTags(headers, baseFields)
	logHeaders := map[string]string{}
	for k, v := range headers {
		req.Header.Set(k, v)
		if k == "Authorization" || k == "X-Client-Secret" || k == "token" {
			logHeaders[k] = "****"
		} else {
			logHeaders[k] = v
		}
	}

	tfields := xwcommon.FilterLogFields(baseFields)
	tfields["logger"] = loggerName

	urlKey := fmt.Sprintf("%v_url", loggerName)
	tfields[urlKey] = url

	methodKey := fmt.Sprintf("%v_method", loggerName)
	tfields[methodKey] = method

	headersKey := fmt.Sprintf("%v_headers", loggerName)
	tfields[headersKey] = logHeaders

	bodyKey := fmt.Sprintf("%v_body", loggerName)
	if len(bbytes) > 0 {
		tfields[bodyKey] = string(bbytes)
	}

	var startMessage string
	if retry > 0 {
		startMessage = fmt.Sprintf("%v retry=%v starts", loggerName, retry)
	} else {
		startMessage = fmt.Sprintf("%v starts", loggerName)
	}
	log.WithFields(tfields).Debug(startMessage)

	delete(tfields, urlKey)
	delete(tfields, methodKey)
	delete(tfields, headersKey)
	delete(tfields, bodyKey)

	var endMessage string
	if retry > 0 {
		endMessage = fmt.Sprintf("%v retry=%v ends", loggerName, retry)
	} else {
		endMessage = fmt.Sprintf("%v ends", loggerName)
	}

	errorKey := fmt.Sprintf("%v_error", loggerName)

	startTime := time.Now()
	res, err := c.Client.Do(req)
	tdiff := time.Since(startTime)
	duration := tdiff.Nanoseconds() / 1000000
	tfields[fmt.Sprintf("%v_duration", loggerName)] = duration

	if res != nil {
		respMoracideTagsFound = c.addMoracideTagsFromResponse(res.Header, baseFields)
	}

	if err != nil {
		// We hit a an err in executing http.Client.Do
		// If err is a timeout, set the status code to 504
		// If err indicates a conn drop, set the status code to 503

		tfields[errorKey] = err.Error()

		uErr, ok := err.(*neturl.Error)
		if ok {
			if uErr.Timeout() {
				log.WithFields(tfields).Debugf("Timeout in %s, status set to 504", loggerName)
				statusCode = http.StatusGatewayTimeout
			} else {
				errStr := uErr.Unwrap().Error()
				if errStr == "EOF" || strings.Contains(errStr, "read: connection reset by peer") {
					log.WithFields(tfields).Debugf("EOF or conn reset in %s, status set to 503", loggerName)
					statusCode = http.StatusServiceUnavailable // Conn drop
				}
			}
		}
		log.WithFields(tfields).Info(endMessage)
		return nil, err, true, statusCode
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	// client.Do succeeded, set the status to response's status code
	statusCode = res.StatusCode

	tfields[fmt.Sprintf("%v_status", loggerName)] = res.StatusCode
	rbytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		tfields[errorKey] = err.Error()
		log.WithFields(tfields).Info(endMessage)
		return nil, err, false, statusCode
	}

	var rbody string
	if loggerName == "satservice" && res.StatusCode == http.StatusOK {
		rbody = "****"
	} else if loggerName == "app" && strings.HasSuffix(url, "/api/v1/common/sat") && res.StatusCode == http.StatusOK {
		rbody = "****"
	} else {
		rbody = string(rbytes)
	}
	if res.StatusCode == http.StatusOK && loggerName == common.AuthProvider &&
		(strings.EqualFold(req.URL.Path, "/idpservice-sp/oauth/token") || strings.EqualFold(req.URL.Path, "/idpservice-sp/keys/jwks")) {
		rbody = "****"
	}

	tfields[fmt.Sprintf("%v_response", loggerName)] = rbody
	log.WithFields(tfields).Debugf("%v ends", loggerName)

	if res.StatusCode >= 400 {
		var errorMessage string
		if len(rbody) > 0 {
			var er ErrorResponse
			if err := json.Unmarshal(rbytes, &er); err == nil {
				errorMessage = er.Message
			}
			if len(errorMessage) == 0 {
				errorMessage = rbody
			}
		} else {
			errorMessage = http.StatusText(res.StatusCode)
		}
		err := common.XconfError{
			Message:    errorMessage,
			StatusCode: res.StatusCode,
		}

		switch res.StatusCode {
		case http.StatusForbidden, http.StatusBadRequest, http.StatusNotFound, 520:
			return rbytes, err, false, statusCode
		}
		return rbytes, err, true, statusCode
	}
	return rbytes, nil, false, statusCode
}

func (c *HttpClient) DoWithRetries(method string, url string, inHeaders map[string]string, bbytes []byte, fields log.Fields, loggerName string) ([]byte, error) {
	var traceId string
	if itf, ok := fields["xmoney_trace_id"]; ok {
		// if traceid is absent in the incoming req, it gets added in logRequestStarts
		// So this if should always fire, but just being safe.
		traceId = itf.(string)
	}

	xmoney := fmt.Sprintf("trace-id=%s;parent-id=0;span-id=0;span-name=%s", traceId, loggerName)
	headers := map[string]string{
		"X-Moneytrace": xmoney,
	}
	for k, v := range inHeaders {
		headers[k] = v
	}

	// var res *http.Response
	var rbytes []byte
	var err error
	var cont bool
	var statusCode int

	startTimeForAllRetries := time.Now()

	extServiceAuditFields := make(map[string]interface{})
	extServiceAuditFields["audit_id"] = fields["audit_id"]
	extServiceAuditFields["trace_id"] = fields["trace_id"]

	i := 0
	// i=0 is NOT considered a retry, so it ends at i=c.webpaRetries
	for i = 0; i <= c.retries; i++ {
		cbytes := make([]byte, len(bbytes))
		copy(cbytes, bbytes)
		if i > 0 {
			time.Sleep(time.Duration(c.retryInMsecs) * time.Millisecond)
		}
		rbytes, err, cont, statusCode = c.Do(method, url, headers, cbytes, extServiceAuditFields, loggerName, i)

		log.WithFields(log.Fields{
			"method": method,
			"url":    url,
			"status": statusCode,
		}).Debug("http client request sent")

		if !cont {
			break
		}
	}

	if WebConfServer != nil && WebConfServer.metricsEnabled && WebConfServer.XW_XconfServer.AppMetrics != nil {
		WebConfServer.XW_XconfServer.AppMetrics.UpdateExternalAPIMetrics(loggerName, method, statusCode, startTimeForAllRetries)
	}

	if err != nil {
		return rbytes, xwcommon.NewRemoteErrorAS(statusCode, err.Error())
	}
	return rbytes, nil
}

// addMoracideTags - if ctx has a moracide tag as a header, add it to the headers
// Also add traceparent, tracestate headers
func (c *HttpClient) addMoracideTags(header map[string]string, fields log.Fields) {
	if itf, ok := fields["out_traceparent"]; ok {
		if ss, ok := itf.(string); ok {
			if len(ss) > 0 {
				header[xwcommon.HeaderTraceparent] = ss
			}
		}
	}
	if itf, ok := fields["out_tracestate"]; ok {
		if ss, ok := itf.(string); ok {
			if len(ss) > 0 {
				header[xwcommon.HeaderTracestate] = ss
			}
		}
	}

	moracide := xwcommon.FieldsGetString(fields, "req_moracide_tag")
	if len(moracide) > 0 {
		header[xwcommon.HeaderMoracide] = moracide
	}
}

func (c *HttpClient) addMoracideTagsFromResponse(header http.Header, fields log.Fields) bool {
	var respMoracideTagsFound bool
	moracide := header.Get(xwcommon.HeaderMoracide)
	if len(moracide) > 0 {
		fields["resp_moracide_tag"] = moracide
		respMoracideTagsFound = true
	}
	return respMoracideTagsFound
}

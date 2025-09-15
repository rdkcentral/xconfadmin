package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	taggingapi_config "github.com/rdkcentral/xconfadmin/taggingapi/config"

	xcommon "github.com/rdkcentral/xconfadmin/common"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	xhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/rdkcentral/xconfwebconfig/tracing"

	"github.com/go-akka/configuration"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	LevelWarn = iota
	LevelInfo
	LevelDebug
	MetricsEnabledDefault     = true
	responseLoggingLowerBound = 1000
	responseLoggingUpperBound = 5000
)

const DEV_PROFILE string = "dev"

var (
	WebConfServer *WebconfigServer
	//ds            *xhttp.XconfServer //TODO
)

// len(response) < lowerBound               ==> convert to json
// lowerBound <= len(response) < upperBound ==> stay string
// upperBound <= len(response)              ==> truncated

type WebconfigServer struct {
	XW_XconfServer *xhttp.XconfServer
	*CanaryMgrConnector
	*XcrpConnector
	IdpServiceConnector
	*XconfConnector
	db.DatabaseClient
	*common.ServerConfig
	*GroupServiceConnector
	*GroupServiceSyncConnector
	*taggingapi_config.TaggingApiConfig
	*tracing.XpcTracer
	tlsConfig          *tls.Config
	notLoggedHeaders   []string
	metricsEnabled     bool
	testOnly           bool
	AppName            string
	ServerOriginId     string
	IdpLoginPath       string
	IdpLogoutPath      string
	IdpLogoutAfterPath string
	IdpCodePath        string
	IdpUrlPath         string
	VerifyStageHost    bool
}

type ExternalConnectors struct {
	xw_ect *xhttp.ExternalConnectors
	IdpServiceConnector
}
type ProcessHook interface {
	Process(*WebconfigServer, ...interface{})
}

func NewExternalConnectors() *ExternalConnectors {
	return &ExternalConnectors{}
}

func NewTlsConfig(conf *configuration.Config) (*tls.Config, error) {
	certFile := conf.GetString("xconfwebconfig.http_client.ca_comodo_cert_file")
	if len(certFile) == 0 {
		log.Warn("http_client.ca_comodo_cert_file not specified")
		return nil, nil
	}
	caCertPEM, err := os.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read comodo cert file %s with error: %+v", certFile, err)
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(caCertPEM)
	if !ok {
		return nil, fmt.Errorf("failed to append cert from pem: %+v", err)
	}

	certFile = conf.GetString("xconfwebconfig.http_client.cert_file")
	if len(certFile) == 0 {
		log.Warn("http_client.cert_file not specified")
		return nil, nil
	}
	privateKeyFile := conf.GetString("xconfwebconfig.http_client.private_key_file")
	if len(privateKeyFile) == 0 {
		log.Warn("http_client.private_key_file not specified")
		return nil, nil
	}

	cert, err := tls.LoadX509KeyPair(certFile, privateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load a public/private key pair: %+v", err)
	}

	return &tls.Config{
		RootCAs:            roots,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
	}, nil
}

// testOnly=true ==> running unit test
func NewWebconfigServer(sc *common.ServerConfig, testOnly bool, dc db.DatabaseClient, ec *ExternalConnectors) *WebconfigServer {
	if ec == nil {
		ec = NewExternalConnectors()
	}
	conf := sc.Config
	var err error

	// appname from config
	appName := strings.Split(conf.GetString("xconfwebconfig.code_git_commit", "xconfadmin-xconf"), "-")[0]

	metricsEnabled := conf.GetBoolean("xconfwebconfig.server.metrics_enabled", MetricsEnabledDefault)

	// configure headers that should not be logged
	ignoredHeaders := conf.GetStringList("xconfwebconfig.log.ignored_headers")
	ignoredHeaders = append(xcommon.DefaultIgnoredHeaders, ignoredHeaders...)
	var notLoggedHeaders []string
	for _, x := range ignoredHeaders {
		notLoggedHeaders = append(notLoggedHeaders, strings.ToLower(x))
	}
	// idp api paths fetching
	idpAuthProvider := conf.GetString("xconfwebconfig.xconf.authprovider", "acl")
	idpAuthServer := conf.GetString("xconfwebconfig.xconf.idp_service_name")
	idpLoginPath := conf.GetString(fmt.Sprintf("xconfwebconfig.%v.idp_login_path", idpAuthServer), idpAuthProvider+"/login")
	idpLogoutPath := conf.GetString(fmt.Sprintf("xconfwebconfig.%v.idp_logout_path", idpAuthServer), idpAuthProvider+"/logout")
	idpUrlPath := conf.GetString(fmt.Sprintf("xconfwebconfig.%v.idp_continue_path", idpAuthServer), idpAuthProvider+"/url")
	idpCodePath := conf.GetString(fmt.Sprintf("xconfwebconfig.%v.idp_code_path", idpAuthServer), idpAuthProvider+"/code")
	idpLogoutAfterPath := conf.GetString(fmt.Sprintf("xconfwebconfig.%v.idp_logout_after_path", idpAuthServer), idpAuthProvider+"/logout/after")
	verifyStageHost := conf.GetBoolean("xconfwebconfig.sat_consumer.verify_stage_host", false)

	// tlsConfig, here we ignore any error
	tlsConfig, err := NewTlsConfig(conf)
	if err != nil && !testOnly {
		panic(err)
	}
	var idpSvc IdpServiceConnector
	if idpAuthProvider != "acl" {
		idpSvc = NewIdpServiceConnector(conf, ec.IdpServiceConnector)
	} else {
		idpSvc = nil
	}
	xpcTracer := tracing.NewXpcTracer(sc.Config)

	WebConfServer = &WebconfigServer{
		tlsConfig:                 tlsConfig,
		notLoggedHeaders:          notLoggedHeaders,
		metricsEnabled:            metricsEnabled,
		testOnly:                  testOnly,
		AppName:                   appName,
		CanaryMgrConnector:        NewCanaryMgrConnector(conf, tlsConfig),
		XcrpConnector:             NewXcrpConnector(conf, tlsConfig),
		IdpServiceConnector:       idpSvc,
		GroupServiceConnector:     NewGroupServiceConnector(conf, tlsConfig),
		GroupServiceSyncConnector: NewGroupServiceSyncConnector(conf, tlsConfig),
		TaggingApiConfig:          taggingapi_config.NewTaggingApiConfig(conf),
		XconfConnector:            NewXconfConnector(conf, "xconf", tlsConfig),
		XW_XconfServer:            xhttp.NewXconfServer(sc, testOnly, ec.xw_ect),
		IdpLoginPath:              idpLoginPath,
		IdpLogoutPath:             idpLogoutPath,
		IdpLogoutAfterPath:        idpLogoutAfterPath,
		IdpCodePath:               idpCodePath,
		IdpUrlPath:                idpUrlPath,
		XpcTracer:                 xpcTracer,
		VerifyStageHost:           verifyStageHost,
	}

	if testOnly {
		WebConfServer.setupMocks()
	}
	return WebConfServer
}

func (s *WebconfigServer) TestingMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		xp := xhttp.NewXResponseWriter(w)
		xw := *xp

		if r.Method == "POST" {
			if r.Body != nil {
				if rbytes, err := ioutil.ReadAll(r.Body); err == nil {
					xw.SetBody(string(rbytes))
				}
			} else {
				xw.SetBody("")
			}
		}
		next.ServeHTTP(&xw, r)
	}
	return http.HandlerFunc(fn)
}

func (s *WebconfigServer) NoAuthMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", AppName())
		w.Header().Set("Xconf-Server", AppName())

		xw := s.logRequestStarts(w, r)
		defer s.logRequestEnds(xw, r)

		next.ServeHTTP(xw, r)
	}
	return http.HandlerFunc(fn)
}

func IsDevProfile() bool {
	activeProfiles := strings.Split(strings.TrimSpace(xcommon.ActiveAuthProfiles), ",")
	if len(activeProfiles) > 0 {
		return DEV_PROFILE == activeProfiles[0]
	}
	defaultProfiles := strings.Split(strings.TrimSpace(xcommon.DefaultAuthProfiles), ",")
	return DEV_PROFILE == defaultProfiles[0]
}

func getPermissions() (permissions []string) {
	if IsDevProfile() {
		permissions = []string{
			xcommon.WRITE_COMMON, xcommon.READ_COMMON,
			xcommon.WRITE_FIRMWARE_ALL, xcommon.READ_FIRMWARE_ALL,
			xcommon.WRITE_DCM_ALL, xcommon.READ_DCM_ALL,
			xcommon.WRITE_TELEMETRY_ALL, xcommon.READ_TELEMETRY_ALL,
			xcommon.READ_CHANGES_ALL, xcommon.WRITE_CHANGES_ALL}
	}
	return permissions
}

func (s *WebconfigServer) AuthValidationMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", AppName())
		w.Header().Set("Xconf-Server", AppName())

		ctx := r.Context()

		// Check for SAT V2 token
		if satToken := getSatTokenFromRequest(r); satToken != "" {
			if subject, capabilities, err := getSubjectAndCapabilitiesFromSatToken(satToken, s.VerifyStageHost); err != nil {
				log.Error(err.Error())
				http.Error(w, "invalid SAT token", http.StatusUnauthorized)
				return
			} else {
				r.Header.Set(AUTH_SUBJECT, subject)

				// Add capabilities to request context
				ctx = context.WithValue(ctx, CTX_KEY_CAPABILITIES, capabilities)
			}
		} else if authToken := getLoginTokenFromRequest(r); authToken != "" {
			if LoginToken, err := ValidateAndGetLoginToken(authToken); err != nil {
				log.Error(err.Error())
				//http.Error(w, "invalid auth token", http.StatusUnauthorized)
				//Cheking and Setting DEV PROFILES to allow and display ALL the TABS
				//ctx = context.WithValue(ctx, CTX_KEY_TOKEN, LoginToken)
				permissions := getPermissions()
				ctx = context.WithValue(ctx, CTX_KEY_PERMISSIONS, permissions)
				//return
			} else {
				//THIS IS LOGIN TOKEN SUCCESS CASE
				r.Header.Set(AUTH_SUBJECT, LoginToken.Subject)

				// Add UI token & permissions to request context
				ctx = context.WithValue(ctx, CTX_KEY_TOKEN, LoginToken)
				permissions := getPermissionsFromLoginToken(LoginToken)
				ctx = context.WithValue(ctx, CTX_KEY_PERMISSIONS, permissions)
			}
		} else if r.Header.Get(RequestID) != "adminui" && !xcommon.SatOn {
			//allowing api request without sat_token if sat is off
			log.Debug("SAT is off, allowing request without SAT token")
		} else {
			http.Error(w, "auth token not found", http.StatusUnauthorized)
			return
		}

		newReq := r.WithContext(ctx)
		xw := s.logRequestStarts(w, newReq)
		defer s.logRequestEnds(xw, newReq)

		next.ServeHTTP(xw, newReq)
	}
	return http.HandlerFunc(fn)
}

func (s *WebconfigServer) MetricsEnabled() bool {
	return s.metricsEnabled
}

func (s *WebconfigServer) TestOnly() bool {
	return s.testOnly
}

func (s *WebconfigServer) TlsConfig() *tls.Config {
	return s.tlsConfig
}

func (s *WebconfigServer) NotLoggedHeaders() []string {
	return s.notLoggedHeaders
}

func getHeadersForLogAsMap(header http.Header, notLoggedHeaders []string) map[string]interface{} {
	loggedHeaders := make(map[string]interface{})
	for k, v := range header {
		if util.CaseInsensitiveContains(notLoggedHeaders, k) {
			continue
		}
		loggedHeaders[k] = v
	}
	return loggedHeaders
}

func (s *WebconfigServer) logRequestStarts(w http.ResponseWriter, r *http.Request) *xhttp.XResponseWriter {
	// extract the token from the header
	authorization := r.Header.Get("Authorization")
	elements := strings.Split(authorization, " ")
	token := ""
	if len(elements) == 2 && elements[0] == "Bearer" {
		token = elements[1]
	}

	var xmTraceId string

	// extract moneytrace from the header
	tracePart := strings.Split(r.Header.Get("X-Moneytrace"), ";")[0]
	if elements := strings.Split(tracePart, "="); len(elements) == 2 {
		if elements[0] == "trace-id" {
			xmTraceId = elements[1]
		}
	}
	if len(xmTraceId) == 0 {
		xmTraceId = uuid.New().String()
		xmoney := fmt.Sprintf("trace-id=%s;parent-id=0;span-id=0;span-name=XConf-Traceid-Injector", xmTraceId)
		log.Debugf("Adding a Money Trace Header %s", xmTraceId)
		r.Header.Add("X-Moneytrace", xmoney)
	}

	// extract auditid from the header
	auditId := r.Header.Get("X-Auditid")
	if len(auditId) == 0 {
		auditId = util.GetAuditId()
	}

	// traceparent handling for E2E tracing
	xpcTrace := tracing.NewXpcTrace(s.XpcTracer, r)
	traceId := xpcTrace.TraceID
	if len(traceId) == 0 {
		traceId = xmTraceId
	}

	fields := log.Fields{
		"audit_id":         auditId,
		"logger":           "request",
		"header":           getHeadersForLogAsMap(r.Header, s.notLoggedHeaders), // logic to remove unwanted headers
		"path":             r.URL.String(),
		"method":           r.Method,
		"remote_ip":        r.RemoteAddr,
		"host_name":        r.Host,
		"auth_subject":     r.Header.Get(AUTH_SUBJECT),
		"xmoney_trace_id":  xmTraceId,
		"traceparent":      xpcTrace.ReqTraceparent,
		"tracestate":       xpcTrace.ReqTracestate,
		"out_traceparent":  xpcTrace.OutTraceparent,
		"out_tracestate":   xpcTrace.OutTracestate,
		"trace_id":         traceId,
		"req_moracide_tag": xpcTrace.ReqMoracideTag,
		"xpc_trace":        xpcTrace,
	}

	xwriter := xhttp.NewXResponseWriter(w, time.Now(), token, fields)

	if r.Method == "POST" || r.Method == "PUT" {
		var body string
		if r.Body != nil {
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fields["error"] = err
				log.Error("request starts")
				return xwriter
			}
			body = string(b)
		}
		xwriter.SetBody(body)
		fields["body"] = body
		// ctx = log.SetContext(ctx, "body", body)

		contentType := r.Header.Get("Content-type")
		if contentType == "application/msgpack" {
			xwriter.SetBodyObfuscated(true)
		}
	}

	tfields := common.FilterLogFields(fields)
	log.WithFields(tfields).Info("request starts")

	return xwriter
}

func (s *WebconfigServer) logRequestEnds(xw *xhttp.XResponseWriter, r *http.Request) {
	tdiff := time.Since(xw.StartTime())
	duration := tdiff.Nanoseconds() / 1000000

	url := r.URL.String()
	response := xw.Response()
	if strings.Contains(url, "/config") || (strings.Contains(url, "/document") && r.Method == "GET") || (url == "/api/v1/token" && r.Method == "POST") {
		response = "****"
	}

	statusCode := xw.Status()
	fields := xw.Audit()

	fields["status"] = statusCode
	fields["duration"] = duration
	fields["response_header"] = getHeadersForLogAsMap(xw.Header(), s.notLoggedHeaders)

	pathTemplate, _ := mux.CurrentRoute(r).GetPathTemplate()
	splPath := false
	if strings.Contains(pathTemplate, "xconf/swu/{applicationType}") {
		splPath = true
	}
	if splPath || statusCode >= http.StatusBadRequest { // >= 400
		// Unmarshal only if json size < lowerBound
		// Truncate only if json size >= upperBound
		// TODO: passwd is exposed if len > lowerBound but < upperBound

		fields["response"] = response
		if len(response) < responseLoggingLowerBound {
			dict := util.Dict{}
			err := json.Unmarshal([]byte(response), &dict)
			if err == nil && len(dict) > 0 {
				if _, ok := dict["password"]; ok {
					dict["password"] = "****"
				}
				fields["response"] = dict
			}
		} else if len(response) >= responseLoggingUpperBound {
			fields["response"] = fmt.Sprintf("%v...TRUNCATED", response[:responseLoggingUpperBound])
		}
	}

	if _, ok := fields["num_results"]; !ok {
		fields["num_results"] = 0
	}

	tfields := common.FilterLogFields(fields)

	s.XpcTracer.SetSpan(fields, s.XpcTracer.MoracideTagPrefix())

	log.WithFields(tfields).Info("request ends")

	if s.metricsEnabled && s.XW_XconfServer.AppMetrics != nil {
		s.XW_XconfServer.AppMetrics.UpdateAPIMetrics(r, xw.Status(), xw.StartTime())
	}
}

func LogError(w http.ResponseWriter, err error) {
	var fields log.Fields
	if xw, ok := w.(*xhttp.XResponseWriter); ok {
		xfields := xw.Audit()
		fields = common.FilterLogFields(xfields)
	} else {
		fields = make(log.Fields)
	}
	fields["error"] = err

	log.WithFields(fields).Error("internal error")
}

// AppName is just a convenience func that returns the AppName, used in metrics
func AppName() string {
	return WebConfServer.AppName
}

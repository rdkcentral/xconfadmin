package tag

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/rdkcentral/xconfadmin/common"
	xhttp "github.com/rdkcentral/xconfadmin/http"
	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// One tag sync run at most per instance; the Cassandra lock guards across
// instances, this guards within one (and carries the cancel for abort).
var (
	activeTagSyncMu     sync.Mutex
	activeTagSyncCancel context.CancelFunc
	activeTagSyncRunId  string
)

// TriggerTagSyncHandler starts a run in the background (DeleteTagHandler
// pattern): validate and lock synchronously, answer 202 with the run id.
// POST /taggingService/tags/sync
func TriggerTagSyncHandler(w http.ResponseWriter, r *http.Request) {
	var opts TagSyncOptions
	body, err := readRequestBody(w, r)
	if err != nil {
		// A body that failed to read must not fall through as "no options
		// sent": that starts a default detect run the caller gets a 202 for,
		// holding the cluster lock against the run they actually asked for.
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf("request body read error: %s", err.Error())))
		return
	}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &opts); err != nil {
			xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(RequestBodyReadErrorMsg, err.Error())))
			return
		}
	}
	if err := validateTagSyncOptions(&opts); err != nil {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(err.Error()))
		return
	}
	tenantId := xhttp.GetTenantId(r)
	if !tagSyncKillSwitchEnabled(tenantId) {
		xhttp.WriteXconfResponse(w, http.StatusConflict, []byte("tag sync is disabled by the TaggingSyncEnabled app setting"))
		return
	}

	activeTagSyncMu.Lock()
	defer activeTagSyncMu.Unlock()
	if activeTagSyncCancel != nil {
		respBytes, _ := json.Marshal(map[string]string{
			"error": "tag sync already running on this instance",
			"runId": activeTagSyncRunId,
		})
		xhttp.WriteXconfResponse(w, http.StatusConflict, respBytes)
		return
	}

	engine, err := PrepareTagSync(opts, tenantId)
	if err != nil {
		var busy *tagSyncBusyError
		if errors.As(err, &busy) {
			respBytes, _ := json.Marshal(map[string]string{
				"error": "tag sync already running",
				"runId": busy.RunId,
				"owner": busy.Owner,
			})
			xhttp.WriteXconfResponse(w, http.StatusConflict, respBytes)
			return
		}
		// A 503 means this instance has not finished starting; say so.
		if xwcommon.GetXconfErrorStatusCode(err) == http.StatusServiceUnavailable {
			xhttp.WriteXconfResponseWithHeaders(w, map[string]string{"Retry-After": "30"},
				http.StatusServiceUnavailable, []byte(err.Error()))
			return
		}
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	// Capture the response fields before Execute starts: once the goroutine
	// runs, engine.run may only be read under the engine mutex.
	runId := engine.run.RunId
	mode := engine.run.Mode
	dryRun := engine.run.Options.DryRun

	ctx, cancel := context.WithCancel(context.Background())
	activeTagSyncCancel = cancel
	activeTagSyncRunId = runId

	go func() {
		defer func() {
			activeTagSyncMu.Lock()
			activeTagSyncCancel = nil
			activeTagSyncRunId = ""
			activeTagSyncMu.Unlock()
			cancel()
		}()
		run := engine.Execute(ctx)
		log.WithFields(log.Fields{
			"audit_id": run.RunId,
			"job":      "tag_sync",
		}).Infof("tag sync background run finished: state=%s", run.State)
	}()

	respBytes, _ := json.Marshal(map[string]interface{}{
		"runId":  runId,
		"mode":   mode,
		"state":  TagSyncStateRunning,
		"dryRun": dryRun,
	})
	xhttp.WriteXconfResponse(w, http.StatusAccepted, respBytes)
}

// TagSyncStatusHandler reports the active run (if any, on any instance) and
// recent run history, straight from the TagSyncState table.
// GET /taggingService/tags/sync/status
func TagSyncStatusHandler(w http.ResponseWriter, r *http.Request) {
	active, history, err := loadTagSyncStatus(newTagSyncDao())
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	respBytes, err := json.Marshal(map[string]interface{}{
		"active":  active,
		"history": history,
		"enabled": tagSyncKillSwitchEnabled(xhttp.GetTenantId(r)),
	})
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, respBytes)
}

// loadTagSyncStatus returns store errors sanitized like the trigger's: the
// driver detail stays in the log and never reaches the client.
func loadTagSyncStatus(dao tagSyncDao) (*TagSyncRun, []*TagSyncRun, error) {
	lock, err := dao.getLock()
	if err != nil {
		return nil, nil, tagSyncStoreError("status lock read", err)
	}
	var active *TagSyncRun
	if lock != nil && !lock.Released && time.Since(lock.HeartbeatAt) < tagSyncLockStaleAfter {
		if active, err = dao.getRun(lock.RunId); err != nil {
			return nil, nil, tagSyncStoreError("status run read", err)
		}
	}
	history, err := dao.listRuns(10)
	if err != nil {
		return nil, nil, tagSyncStoreError("status run list", err)
	}
	return active, history, nil
}

// TagSyncRunStatusHandler returns a single run record, for polling one run
// without the whole history.
// GET /taggingService/tags/sync/status/{runId}
func TagSyncRunStatusHandler(w http.ResponseWriter, r *http.Request) {
	runId := mux.Vars(r)[common.RunId]
	if runId == "" {
		xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte(fmt.Sprintf(NotSpecifiedErrorMsg, common.RunId)))
		return
	}
	run, err := loadTagSyncRun(newTagSyncDao(), runId)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	if run == nil {
		xhttp.WriteXconfResponse(w, http.StatusNotFound,
			[]byte(fmt.Sprintf("tag sync run %s not found; only the %d most recent runs are kept", runId, tagSyncRunHistoryKeep)))
		return
	}

	respBytes, err := json.Marshal(run)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, respBytes)
}

func loadTagSyncRun(dao tagSyncDao, runId string) (*TagSyncRun, error) {
	run, err := dao.getRun(runId)
	if err != nil {
		return nil, tagSyncStoreError("status run read", err)
	}
	return run, nil
}

// AbortTagSyncHandler cancels the run owned by this instance. Runs on other
// instances are stopped with the TaggingSyncEnabled kill switch instead.
// POST /taggingService/tags/sync/abort
func AbortTagSyncHandler(w http.ResponseWriter, r *http.Request) {
	activeTagSyncMu.Lock()
	defer activeTagSyncMu.Unlock()
	if activeTagSyncCancel == nil {
		xhttp.WriteXconfResponse(w, http.StatusNotFound,
			[]byte("no active tag sync on this instance; to stop a run elsewhere set the TaggingSyncEnabled app setting to false"))
		return
	}
	activeTagSyncCancel()
	respBytes, _ := json.Marshal(map[string]string{
		"runId":  activeTagSyncRunId,
		"status": "abort requested",
	})
	xhttp.WriteXconfResponse(w, http.StatusAccepted, respBytes)
}

// readRequestBody works both when the auth middleware has already buffered
// the body into XResponseWriter and when it has not. A read failure comes back
// as an error rather than an empty body, so the caller can tell "no options
// sent" from "options lost in transit".
func readRequestBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	if xw, ok := w.(*xwhttp.XResponseWriter); ok {
		if body := xw.Body(); body != "" {
			return []byte(body), nil
		}
	}
	if r.Body == nil {
		return nil, nil
	}
	return io.ReadAll(r.Body)
}

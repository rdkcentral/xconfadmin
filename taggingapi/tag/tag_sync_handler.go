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

	xhttp "github.com/rdkcentral/xconfadmin/http"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"

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
	if !tagSyncKillSwitchEnabled() {
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

	engine, err := PrepareTagSync(opts)
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
	dao := newTagSyncDao()

	var active *TagSyncRun
	lock, err := dao.getLock()
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	if lock != nil && !lock.Released && time.Since(lock.HeartbeatAt) < tagSyncLockStaleAfter {
		active, err = dao.getRun(lock.RunId)
		if err != nil {
			xhttp.WriteXconfErrorResponse(w, err)
			return
		}
	}

	history, err := dao.listRuns(10)
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}

	respBytes, err := json.Marshal(map[string]interface{}{
		"active":  active,
		"history": history,
		"enabled": tagSyncKillSwitchEnabled(),
	})
	if err != nil {
		xhttp.WriteXconfErrorResponse(w, err)
		return
	}
	xhttp.WriteXconfResponse(w, http.StatusOK, respBytes)
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

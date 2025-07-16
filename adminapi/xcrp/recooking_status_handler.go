package xcrp

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
	owhttp "xconfadmin/http"

	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

const DEFAULT_XCRP_SERVICE_NAME = "xcrp"

type RecookingStatus struct {
	AppName     string    `json:"appName"`
	PartitionId string    `json:"partitionId"`
	State       string    `json:"state"`
	UpdatedTime time.Time `json:"updatedTime"`
}

func GetRecookingStatusHandler(w http.ResponseWriter, r *http.Request) {
	cc, ok := db.GetDatabaseClient().(*db.CassandraClient)
	if !ok {
		http.Error(w, "Database client is not Cassandra client", http.StatusInternalServerError)
		return
	}

	status, updatedTime, err := cc.CheckFinalRecookingStatus(DEFAULT_XCRP_SERVICE_NAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if updatedTime.IsZero() {
		owhttp.WriteErrorResponse(w, http.StatusNotFound, errors.New("no recooking status found"))
		return
	}

	dict := make(util.Dict)

	var responseStatus string
	if status {
		responseStatus = "completed"
		log.Infof("Recooking/precooking completed for %s at %s", DEFAULT_XCRP_SERVICE_NAME, updatedTime)
	} else {
		responseStatus = "in progress"
	}

	dict["status"] = responseStatus
	dict["updatedTime"] = updatedTime

	owhttp.WriteOkResponse(w, r, dict)
}

func GetRecookingStatusDetailsHandler(w http.ResponseWriter, r *http.Request) {
	cc, ok := db.GetDatabaseClient().(*db.CassandraClient)
	if !ok {
		http.Error(w, "Database client is not Cassandra client", http.StatusInternalServerError)
		return
	}
	statuses, err := cc.GetRecookingStatusDetails()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(statuses)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

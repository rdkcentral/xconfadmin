package xcrp

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"xconfadmin/adminapi/auth"
	"xconfadmin/adminapi/lockdown"
	"xconfadmin/common"
	xhttp "xconfadmin/http"
	dao "xconfwebconfig/db"
	xwhttp "xconfwebconfig/http"

	log "github.com/sirupsen/logrus"
)

type State int

func GetXcrpConnector() *xhttp.XcrpConnector {
	return xhttp.WebConfServer.XcrpConnector
}

func PostRecookingLockdownSettingsHandler(w http.ResponseWriter, r *http.Request) {

	if !auth.HasWritePermissionForTool(r) {
		xhttp.WriteAdminErrorResponse(w, http.StatusForbidden, "No write permission: tools")
		return
	}
	xw, ok := w.(*xwhttp.XResponseWriter)
	if !ok {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "responsewriter cast error")
		return
	}
	var recookingLockdownSetting common.RecookingLockdownSettings
	body := xw.Body()
	err := json.Unmarshal([]byte(body), &recookingLockdownSetting)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	fields := xw.Audit()

	var models, partners []string

	if recookingLockdownSetting.Models != nil {
		models = *recookingLockdownSetting.Models
	}
	if recookingLockdownSetting.Partners != nil {
		partners = *recookingLockdownSetting.Partners
	}

	dao.GetCacheManager().ForceSyncChanges()

	var lockdownSettingFromDB *common.LockdownSettings
	lockdownSettingFromDB, err = lockdown.GetLockdownSettings()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isLockdownMode() && *(lockdownSettingFromDB.LockdownModules) == "rfc" {
		xhttp.WriteAdminErrorResponse(w, http.StatusBadRequest, "Lockdown rfc is enabled.")
		return
	}

	lockdownEnabled := true
	lockdownModules := "rfc"
	var timezone *time.Location
	timezone, err = time.LoadLocation(common.DefaultLockdownTimezone)
	if err != nil {
		log.Errorf("Error loading timezone: %s", common.DefaultLockdownTimezone)
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	currentTimeWithDate := time.Now().In(timezone).Format(common.DefaultTimeDateFormatLayout)

	var currentTime time.Time
	currentTime, err = time.Parse(common.DefaultTimeDateFormatLayout, currentTimeWithDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lockdownStartTime := currentTime.Format(common.DefaultTimeFormatLayout)
	lockdownEndTime := currentTime.Add(time.Second * time.Duration(common.LockDuration)).Format(common.DefaultTimeFormatLayout)

	lockdownSettings := common.LockdownSettings{
		LockdownEnabled:   &lockdownEnabled,
		LockdownStartTime: &lockdownStartTime,
		LockdownEndTime:   &lockdownEndTime,
		LockdownModules:   &lockdownModules,
	}
	respEntity := lockdown.SetLockdownSetting(&lockdownSettings)
	if respEntity.Error != nil {
		xhttp.WriteAdminErrorResponse(w, respEntity.Status, respEntity.Error.Error())
		return
	}

	log.Infof("Precook lockdown settings in EDT, lockdownStartTime: %v, lockdownEndTime: %v, lockdownModules: %v, lockdownEnabled: %v", lockdownStartTime, lockdownEndTime, lockdownModules, lockdownEnabled)

	go CheckRecookingStatus(time.Second*time.Duration(common.LockDuration), "rfc", fields)
	err = GetXcrpConnector().PostRecook(models, partners, nil, fields)
	if err != nil {
		xhttp.WriteAdminErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	xhttp.WriteXconfResponse(w, respEntity.Status, nil)
}

// integrate with the lockdown settings api function, since it is not exported,copied the function here
func isLockdownMode() bool {
	if common.GetBooleanAppSetting(common.PROP_LOCKDOWN_ENABLED, false) {
		startTime := common.GetStringAppSetting(common.PROP_LOCKDOWN_STARTTIME)
		endTime := common.GetStringAppSetting(common.PROP_LOCKDOWN_ENDTIME)

		timezone, err := time.LoadLocation(common.DefaultLockdownTimezone)
		if err != nil {
			log.Errorf("Error loading timezone: %s", common.DefaultLockdownTimezone)
			return false
		}

		t := time.Now().In(timezone).Format(common.DefaultTimeDateFormatLayout)
		CurrentDate := time.Now().In(timezone).Format(common.DefaultDateFormatLayout)

		Currenttime, err := time.Parse(common.DefaultTimeDateFormatLayout, t)

		if err != nil {
			log.Errorf("Unable to Parse currenttime: %s", Currenttime)
			return false
		}
		LockdownStartTime, err := time.Parse(common.DefaultTimeDateFormatLayout, CurrentDate+" "+startTime)
		if err != nil {
			log.Errorf("Unable to Parse LockdownStartTime: %s", LockdownStartTime)
			return false
		}
		LockdownEndTime, err := time.Parse(common.DefaultTimeDateFormatLayout, CurrentDate+" "+endTime)
		if err != nil {
			log.Errorf("Unable to Parse LockdownEndTime: %s", LockdownEndTime)
			return false
		}

		if LockdownStartTime.After(LockdownEndTime) || LockdownStartTime.Equal(LockdownEndTime) {
			LockdownStartTime = LockdownStartTime.AddDate(0, 0, -1)
		}

		if (Currenttime.Equal(LockdownStartTime) || Currenttime.After(LockdownStartTime)) && Currenttime.Before(LockdownEndTime) {
			log.Infof("Lockdown Mode is Scheduled Now. Current time=%s, Lockdown StartTime=%s, Lockdown EndTime=%s", t, startTime, endTime)
			return true
		}
		return false
	}
	return false
}

func CheckRecookingStatus(lockDuration time.Duration, module string, fields log.Fields) {
	//utilize the lockDuration to determine when to check the recooking status in background task
	endTime := time.Now().Add(lockDuration).UTC()
	time.Sleep(lockDuration)

	var state bool
	var updatedTime time.Time
	var err error
	state, err = GetXcrpConnector().GetRecookingStatusFromCanaryMgr(module, fields)
	if err != nil {
		log.Errorf("Error checking recooking status from CanaryMgr: %v", err)
		return
	}

	if !state {
		log.Infof("Recooking is not able to be completed at %v, disable the delivery of precook data for now", endTime)
		_, err = common.SetAppSetting(common.PROP_PRECOOK_LOCKDOWN_ENABLED, true)
		if err != nil {
			log.Errorf("Error setting appSetting for precookLockDownEnabled: %v", err)
		}
	} else {
		log.Infof("Recooking is completed at %v, RecookingStatus is checked at %v, deliver precook data", updatedTime, endTime)
		_, err = common.SetAppSetting(common.PROP_PRECOOK_LOCKDOWN_ENABLED, false)
		if err != nil {
			log.Errorf("Error setting appSetting for precookLockDownEnabled: %v", err)
		}
	}

	lockdownSettingFromDB, _ := lockdown.GetLockdownSettings()
	lockdownMoules := *(lockdownSettingFromDB.LockdownModules)
	if lockdownMoules == "rfc" {
		log.Debug("Reached lockdown duration, disable the lockdown for rfc")
		_, err = common.SetAppSetting(common.PROP_LOCKDOWN_ENABLED, false)
		if err != nil {
			log.Errorf("Error setting appSetting for lockdownEnabled: %v", err)
		}
	} else {
		modules := strings.Split(strings.ToLower(lockdownMoules), ",")
		//remove rfc from modules
		var newModules []string
		for _, module := range modules {
			if module != "rfc" {
				newModules = append(newModules, module)
			}
		}
		_, err = common.SetAppSetting(common.PROP_LOCKDOWN_MODULES, strings.Join(newModules, ","))
		log.Debugf("removed rfc from lockdown modules, updated lockdownModules: %v", strings.Join(newModules, ","))
		if err != nil {
			log.Errorf("Error setting appSetting for lockdownModules: %v", err)
		}
	}
}

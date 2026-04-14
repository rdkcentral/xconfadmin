package applicationtype

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/rdkcentral/xconfadmin/adminapi/auth"
	xapptype "github.com/rdkcentral/xconfadmin/shared/applicationtype"
	xwcommon "github.com/rdkcentral/xconfwebconfig/common"
	log "github.com/sirupsen/logrus"
)

const (
	applicationRegexPattern = "^[a-zA-Z]{3,12}$"
)

func CreateApplicationType(r *http.Request, appType *xapptype.ApplicationType) (*xapptype.ApplicationType, error) {
	_, err := auth.CanWrite(r, auth.CHANGE_ENTITY)
	if err != nil {
		return nil, err
	}

	if err := ValidateApplicationType(appType); err != nil {
		return nil, err
	}

	appType.ID = uuid.New().String()
	appType.CreatedBy = auth.GetUserNameOrUnknown(r)
	appType.CreatedAt = time.Now().Unix()

	err = xapptype.SetOneApplicationType(appType)
	if err != nil {
		log.Error(fmt.Sprintf("CreateApplicationType error: %v", err))
		return nil, xwcommon.NewRemoteErrorAS(http.StatusInternalServerError, err.Error())
	}

	log.Info(fmt.Sprintf("Application type created: %s by %s", appType.Name, appType.CreatedBy))
	return appType, nil
}

func ValidateApplicationType(appType *xapptype.ApplicationType) error {
	if appType == nil {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Application type cannot be nil")
	}
	if appType.Name == "" {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Application type name cannot be empty")
	}

	matched, _ := regexp.MatchString(applicationRegexPattern, appType.Name)
	if !matched {
		return xwcommon.NewRemoteErrorAS(http.StatusBadRequest, "Application type name must be 3-12 alphabetic characters")
	}
	return nil
}

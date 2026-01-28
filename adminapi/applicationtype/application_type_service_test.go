package applicationtype

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	xapptype "github.com/rdkcentral/xconfadmin/shared/applicationtype"
	"github.com/stretchr/testify/assert"
)

func TestCreateApplicationType(t *testing.T) {
	appType := &xapptype.ApplicationType{
		Name: "testApp",
	}
	req := httptest.NewRequest(http.MethodPost, "/api/application-types", nil)
	req.Header.Set("X-User-Name", "testuser")
	createdAppType, err := CreateApplicationType(req, appType)

	if err == nil || (err != nil && !strings.Contains(err.Error(), "cache not found") && !strings.Contains(err.Error(), "Table configuration not found")) {
		assert.NotNil(t, createdAppType)
		assert.NoError(t, err)
	} else {
		t.Skip("Skipping: database not configured")
	}
}

func TestValidateApplicationType(t *testing.T) {
	err := ValidateApplicationType(nil)
	assert.Error(t, err)

	err = ValidateApplicationType(&xapptype.ApplicationType{Name: ""})
	assert.Error(t, err)
}

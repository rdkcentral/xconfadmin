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
	req := httptest.NewRequest(http.MethodPost, "/api/applicationtype", nil)
	req.Header.Set("X-User-Name", "testuser")
	createdAppType, err := CreateApplicationType(req, appType)
	if err != nil {
		if strings.Contains(err.Error(), "Table configuration not found") {
			t.Skip("Skipping test: database not configured")
			return
		}
	}
	assert.NotNil(t, createdAppType)
	assert.NoError(t, err)
}

func TestValidateApplicationType(t *testing.T) {
	err := ValidateApplicationType(nil)
	assert.Error(t, err)

	err = ValidateApplicationType(&xapptype.ApplicationType{Name: ""})
	assert.Error(t, err)
}

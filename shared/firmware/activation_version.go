package firmware

type ActivationVersion struct {
	ID                 string   `json:"id"`
	ApplicationType    string   `json:"applicationType,omitempty"`
	Description        string   `json:"description,omitempty"`
	Model              string   `json:"model,omitempty"`
	PartnerId          string   `json:"partnerId,omitempty"`
	RegularExpressions []string `json:"regularExpressions"`
	FirmwareVersions   []string `json:"firmwareVersions"`
}
 

// setApplicationType implements queries.T.
func (obj *ActivationVersion) SetApplicationType(appType string) {
	obj.ApplicationType = appType
}

// getApplicationType implements queries.T.
func (obj* ActivationVersion) GetApplicationType() string {
	return obj.ApplicationType
}

// NewActivationVersion constructor
func NewActivationVersion() *ActivationVersion {
	return &ActivationVersion{
		RegularExpressions: []string{},
		FirmwareVersions:   []string{},
	}
}
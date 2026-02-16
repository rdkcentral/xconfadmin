package applicationtype

import "strings"

type ApplicationType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedBy   string `json:"createdBy"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt,omitempty"`
}

var defaultTypes = []string{"stb", "xhome", "sky"}

func IsDefaultAppType(name string) bool {
	for _, dt := range defaultTypes {
		if strings.EqualFold(name, dt) {
			return true
		}
	}
	return false
}

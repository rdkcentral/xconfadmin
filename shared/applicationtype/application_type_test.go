package applicationtype

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IsDefaultAppType(t *testing.T) {
	assert.True(t, IsDefaultAppType("stb"))
	assert.False(t, IsDefaultAppType("customApp"))
}

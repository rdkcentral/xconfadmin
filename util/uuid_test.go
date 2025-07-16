package util

import (
	"testing"

	"github.com/google/uuid"
	"gotest.tools/assert"
)

func TestUUIDMain(t *testing.T) {
	s := uuid.New().String()
	t.Logf("s=%v\n", s)

	_, err := uuid.Parse(s)
	assert.Assert(t, err, nil)

	s1 := "f9fe049c-134c-4f2c-8300-6a583caf5e6x"
	_, err = uuid.Parse(s1)
	assert.DeepEqual(t, err.Error(), "invalid UUID format")
}

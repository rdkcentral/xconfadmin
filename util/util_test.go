package util

import (
	"bytes"
	"testing"

	"gotest.tools/assert"
)

type TestStruct struct {
	TestVar1 string `json:"var_one"`
	TestVar2 bool   `json:"var_two"`
}

func TestJsonMarshal(t *testing.T) {
	tmp := TestStruct{
		TestVar1: "test&string 1>0",
		TestVar2: false,
	}
	testdata := []byte(`{"var_one":"test&string 1>0","var_two":false}`)

	adata, err := JSONMarshal(tmp)
	assert.NilError(t, err)

	res := bytes.Compare(adata, testdata)
	assert.Equal(t, res, 1)

	list := []string{}
	adata, err = JSONMarshal(list)
	assert.NilError(t, err)
}

func TestValidateCronDayAndMonth(t *testing.T) {
	err := ValidateCronDayAndMonth("0 0 * * *")
	assert.NilError(t, err)

	err = ValidateCronDayAndMonth("0 0 29 1 *")
	assert.NilError(t, err)

	err = ValidateCronDayAndMonth("0 0 1 0 *")
	assert.NilError(t, err)

	err = ValidateCronDayAndMonth("0 0 31 0 *")
	assert.NilError(t, err)

	err = ValidateCronDayAndMonth("0 0 0 0 *")
	assert.ErrorContains(t, err, "CronExpression has unparseable day or month value:")

	err = ValidateCronDayAndMonth("0 0 32 0 *")
	assert.ErrorContains(t, err, "CronExpression has unparseable day or month value:")

	err = ValidateCronDayAndMonth("0 0 1 12 *")
	assert.ErrorContains(t, err, "CronExpression has unparseable day or month value:")
}

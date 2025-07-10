package common

import (
	"errors"
	"net/http"
	"testing"

	"gotest.tools/assert"
)

func Red() error {
	err := XconfError{
		StatusCode: http.StatusNotFound,
		Message:    "data not found",
	}
	return err
}

func Orange() error {
	err := Red()
	if err != nil {
		return err
	}
	return nil
}

func Yellow() error {
	err := Orange()
	if err != nil {
		return err
	}
	return nil
}

func Green() error {
	err := Yellow()
	if err != nil {
		return err
	}
	return nil
}

func Blue() error {
	err := Green()
	if err != nil {
		return err
	}
	return nil
}

func Indigo() error {
	err := Blue()
	if err != nil {
		return err
	}
	return nil
}

func Violet() error {
	err := Indigo()
	if err != nil {
		return err
	}
	return nil
}

func TestUnwrapAll(t *testing.T) {
	err := Violet()
	assert.Assert(t, errors.As(err, XconfErrorType))
	unerr := UnwrapAll(err)
	rhe, _ := unerr.(XconfError)
	assert.Equal(t, rhe.StatusCode, http.StatusNotFound)
	assert.Equal(t, rhe.Message, "data not found")
}

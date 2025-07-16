package common

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	NotOK             = fmt.Errorf("!ok")
	NotFound          = fmt.Errorf("Not found")
	NotFirmwareConfig = fmt.Errorf("Not FirmwareCofig")
	NotFirmwareRule   = fmt.Errorf("Not FirmwareRule")
)

var XconfErrorType = &XconfError{}

type XconfError struct {
	StatusCode int
	Message    string
}

func (e XconfError) Error() string {
	return e.Message
}

func NewXconfError(status int, message string) error {
	return XconfError{StatusCode: status, Message: message}
}

func GetXconfErrorStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if errors.As(err, XconfErrorType) {
		return err.(XconfError).StatusCode
	}
	return http.StatusInternalServerError
}

func UnwrapAll(wrappedErr error) error {
	err := wrappedErr
	for i := 0; i < 10; i++ {
		unerr := errors.Unwrap(err)
		if unerr == nil {
			return err
		}
		err = unerr
	}
	return err
}

func NewError(err error) error {
	return err
}

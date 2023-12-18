/**
 * Copyright 2023 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
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

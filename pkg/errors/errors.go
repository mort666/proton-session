/**
 * Copyright © 2020-2025 Stephen Kapp and Reaper Technologies Limited.
 * All Rights Reserved.
 *
 * @Author: Stephen Kapp
 * @Date: 2025-5-12 23:13:18
 * @Last Modified by: Stephen Kapp
 * @Last Modified time: 2025-5-12 23:13:18
 */

package errors

import (
	"errors"
	"fmt"
)

var (
	ErrorMissingUID          = errors.New("missing UID")
	ErrorMissingAccessToken  = errors.New("missing access token")
	ErrorMissingRefreshToken = errors.New("missing refresh token")
	ErrKeyNotFound           = errors.New("key not found")
	ErrFileNotFound          = errors.New("file not found")
)

var ErrErrorAuthenticating = NewErrorf("authenticating: %w")
var ErrErrorMissingAuthCookie = NewErrorf("auth cookie not found in HTTP headers %s")
var ErrErrorHTTPSatusNotOK = NewErrorf("HTTP status code not OK: %s: %s (code %d with details: %s)")

var ErrUnsupportedOption = errors.New("unsupported option")


type Errorf func(args ...interface{}) error

func NewErrorf(message string) Errorf {
	return func(args ...interface{}) error {
		return fmt.Errorf(message, args...)
	}
}

func New(message string) error {
	return errors.New(message)
}

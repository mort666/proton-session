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
	"encoding/json"
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
var ErrErrorReadingResponseBody = NewErrorf("reading response body: %w")
var ErrErrorUnexpectedServerProof = errors.New("unexpected server proof")
var ErrInvalidProof = errors.New("invalid or unexpected server proof")
var ErrErrorGeneratingProofs = NewErrorf("generating SRP proofs: %w")
var ErrErrorInitSRPAuth = NewErrorf("initializing SRP auth: %w")
var ErrAPIErrIsNotHVErr = errors.New("not HV error")
var ErrErrorUnmarshalApiError = NewErrorf("error unmarshalling apierror response: %w\n\tbody: %s")
var ErrHVRequiredError = NewErrorf("human verification required: %w")
var ErrHVInputTimeoutError = errors.New("timeout while waiting for HV confirmation")


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

func Is(err error, cmp error) bool {
  return errors.Is(err, cmp)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}

func AsType[E error](err error) (E, bool) {
	if err == nil {
		var zero E
		return zero, false
	}

	return errors.AsType[E](err)
}


type ErrDetails []byte

// APIError represents an error returned by the API.
type APIError struct {
	// Status is the HTTP status code of the response that caused the error.
	Status int

	// Code is the error code returned by the API.
	Code Code

	// Message is the error message returned by the API.
	Message string `json:"Error"`

	// Details contains optional error details which are specific to each request.
	// Note that the contents of this field needs to be valid serialized JSON.
	Details ErrDetails `json:"Details,omitempty"`
}

func (err APIError) Error() string {
	return fmt.Sprintf("%v (Code=%v, Status=%v)", err.Message, err.Code, err.Status)
}

func (err APIError) IsHVError() bool {
	return err.Code == HumanVerificationRequired
}

func (err APIError) DetailsToString() string {
	if err.Details == nil {
		return ""
	}

	return string(err.Details)
}

func (d ErrDetails) MarshalJSON() ([]byte, error) {
	return d, nil
}

func (d *ErrDetails) UnmarshalJSON(data []byte) error {
	*d = data
	return nil
}

func (err APIError) GetHVDetails() (*APIHVDetails, error) {
	if !err.IsHVError() {
		return nil, ErrAPIErrIsNotHVErr
	}

	r := new(APIHVDetails)

	if err := json.Unmarshal(err.Details, &r); err != nil {
		return nil, err
	}

	return r, nil
}


type Code int

const (
	SuccessCode                 Code = 1000
	MultiCode                   Code = 1001
	InvalidValue                Code = 2001
	AppVersionMissingCode       Code = 5001
	AppVersionBadCode           Code = 5003
	UsernameInvalid             Code = 6003 // Deprecated, but still used.
	PasswordWrong               Code = 8002
	HumanVerificationRequired   Code = 9001
	PaidPlanRequired            Code = 10004
	AuthRefreshTokenInvalid     Code = 10013
	HumanValidationInvalidToken Code = 12087
)

// APIHVDetails contains information related to the human verification requests.
type APIHVDetails struct {
	Methods []string `json:"HumanVerificationMethods"`
	Token   string   `json:"HumanVerificationToken"`
}

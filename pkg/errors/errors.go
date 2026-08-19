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

type Code int

// Response Codes returned from the ProtonAPI to indicate the outcome of the request
const (
	SuccessCode                 Code = 1000  // Request Successful
	MultiCode                   Code = 1001  // MultiCode Response
	InvalidValue                Code = 2001  // Request included an invalid value
	AppVersionMissingCode       Code = 5001  // Request was missing the required AppVersion header
	AppVersionBadCode           Code = 5003  // An invalid AppVersion header was provided
	UsernameInvalid             Code = 6003  // Deprecated, but still used.
	PasswordWrong               Code = 8002  // Provided User Password is invalid
	HumanVerificationRequired   Code = 9001  // Request failure due to the requirement for human verification
	PaidPlanRequired            Code = 10004 // Request action failed as the action requires the account have a Paid Subscription
	AuthRefreshTokenInvalid     Code = 10013 // Provided Auth Refresh token was invalid
	HumanValidationInvalidToken Code = 12087 // The Human Verification Token when used was invalid
)

// APIHVDetails contains information related to the human verification requests.
type APIHVDetails struct {
	Methods []string `json:"HumanVerificationMethods"`
	Token   string   `json:"HumanVerificationToken"`
}

// ErrorMissingUID indicates that the user UID is missing from the session store
var ErrorMissingUID = errors.New("missing UID")

// ErrorMissingAccessToken indicates that the Access token is missing from the session store
var ErrorMissingAccessToken = errors.New("missing access token")

// ErrorMissingRefreshToken indicates that the Resfresh token is missing from the session store
var ErrorMissingRefreshToken = errors.New("missing refresh token")

// ErrKeyNotFound indicates that a lookup key from a session store is missing
var ErrKeyNotFound = errors.New("key not found")

// ErrFileNotFound generic file not found error
var ErrFileNotFound = errors.New("file not found")

// ErrErrorAuthenticating indicates the wrapped error occured during authentication
var ErrErrorAuthenticating = NewErrorf("authenticating: %w")

// ErrErrorMissingAuthCookie while authenticating this error may occur if the server response
// does not include the expected authentication cookies
var ErrErrorMissingAuthCookie = NewErrorf("auth cookie not found in HTTP headers %s")

// ErrErrorHTTPSatusNotOK indicates that a Non-OK HTTP Status code was received and provides
// additional detail from the response as provided
var ErrErrorHTTPSatusNotOK = NewErrorf("HTTP status code not OK: %s: %s (code %d with details: %s)")

// ErrErrorReadingResponseBody indicates that the wrapped error occured while reading the
// body of the HTTP response
var ErrErrorReadingResponseBody = NewErrorf("reading response body: %w")

// ErrErrorUnexpectedServerProof indicates that the authentication API sent an unexpected server proof
var ErrErrorUnexpectedServerProof = errors.New("unexpected server proof")

// ErrInvalidProof indicates that the authentication API sent an invalid or unexpected server proof
var ErrInvalidProof = errors.New("invalid or unexpected server proof")

// ErrErrorGeneratingProofs indicates that the wrapped error occured during the generation
// of authentication proofs.
var ErrErrorGeneratingProofs = NewErrorf("generating SRP proofs: %w")

// ErrErrorInitSRPAuth indicates that the wrapped error caused a failure to
// initialise the SRP authentication proofs generation code
var ErrErrorInitSRPAuth = NewErrorf("initialising SRP auth: %w")

// ErrAPIErrIsNotHVErr indicates that the returned API error when validated is not an actual
// human verification error even though the API may have returned the 9001 status code
var ErrAPIErrIsNotHVErr = errors.New("not HV error")

// ErrErrorUnmarshalApiError indicates that the wrapped error and associate response body caused
// a JSON UnMarshalling error when building the ApiError object
var ErrErrorUnmarshalApiError = NewErrorf("error unmarshalling apierror response: %w\n\tbody: %s")

// ErrHVRequiredError indicates that the API has triggered a human verification challenge
var ErrHVRequiredError = NewErrorf("human verification required: %w")

// ErrHVInputTimeoutError occurs when the verification token has expired when used
var ErrHVInputTimeoutError = errors.New("timeout while waiting for HV confirmation")

// ErrUnsupportedOption indicates a generic unsupported option was provied to the calling code
var ErrUnsupportedOption = errors.New("unsupported option")

type Errorf func(args ...interface{}) error

// Provides a local wrapper for creation of New Errors that take a message that implements
// format string based parameter inclusion and allows using the '%w' format string directive to Wrap
// a error provided as an arguement. Which can then be reused within the application as a function
// call to return the error.
func NewErrorf(message string) Errorf {
	return func(args ...interface{}) error {
		return fmt.Errorf(message, args...)
	}
}

// Provides a local wrapper for creation of New Errors
func New(message string) error {
	return errors.New(message)
}

// Provides an local wrapper to the standard library errors package
func Is(err error, cmp error) bool {
	return errors.Is(err, cmp)
}

// Provides an local wrapper to the standard library errors package
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Provides an local wrapper to the standard library errors package
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

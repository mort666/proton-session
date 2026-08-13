package proton

import (
	"encoding/json"
	"errors"
	"fmt"
)

var ErrInvalidProof = errors.New("unexpected server proof")
var ErrAPIErrIsNotHVErr = errors.New("not HV error")

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

// NetError represents a network error. It is returned when the API is unreachable.
type NetError struct {
	// Cause is the underlying error that caused the network error.
	Cause error

	// Message is an additional message that describes the network error.
	Message string
}

func NewNetError(err error, message string) *NetError {
	return &NetError{Cause: err, Message: message}
}

func (err *NetError) Error() string {
	return fmt.Sprintf("%s: %v", err.Message, err.Cause)
}

func (err *NetError) Unwrap() error {
	return err.Cause
}

func (err *NetError) Is(target error) bool {
	_, ok := target.(*NetError)
	return ok
}

func Is(err error, cmp error) bool {
  return errors.Is(err, cmp)
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

func As(err error, target any) bool {
	return errors.As(err, target)
}
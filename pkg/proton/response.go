package proton

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-resty/resty/v2"

)

// parseResponse should be used as post-processing of response when request is
// called with resty.SetDoNotParseResponse(off).
//
// In this case the resty is not processing request at all including http
// status check or APIerror parsing. Hence, the returned error would be nil
// even on non-200 reponsenses.
//
// This function also closes the response body.
func parseResponse(res *resty.Response, err error) (*resty.Response, error) {
	if err != nil || res.StatusCode() == 200 {
		return res, err
	}

	method := "NONE"
	route := "N/A"

	if res.Request != nil {
		method = res.Request.Method
		route = res.Request.URL
	}

	apiErr, ok := parseRawAPIError(res.RawBody())
	if !ok {
		apiErr = &APIError{
			Code:    0,
			Message: res.Status(),
		}
	}

	apiErr.Status = res.StatusCode()

	return res, fmt.Errorf(
		"%v %s %s: %w",
		res.StatusCode(), method, route, apiErr,
	)
}

func parseRawAPIError(rawResponse io.ReadCloser) (*APIError, bool) {
	apiErr := APIError{}

	defer func() {
		_ = rawResponse.Close()
	}()

	body, err := io.ReadAll(rawResponse)
	if err != nil {
		return &apiErr, false
	}

	if err := json.Unmarshal(body, &apiErr); err != nil {
		return &apiErr, false
	}

	return &apiErr, true
}


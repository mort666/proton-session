package proton

import (
	"strings"

	"github.com/go-resty/resty/v2"
)

const hvPMTokenHeaderField = "x-pm-human-verification-token"
const hvPMTokenType = "x-pm-human-verification-token-type"

func AddHVToRequest(req *resty.Request, hv *APIHVDetails) *resty.Request {
	if hv == nil {
		return req
	}

	return req.SetHeader(hvPMTokenHeaderField, hv.Token).SetHeader(hvPMTokenType, strings.Join(hv.Methods, ","))
}



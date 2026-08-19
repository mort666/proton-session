/**
 * Copyright © 2020-2026 Stephen Kapp and Reaper Technologies Limited.
 * All Rights Reserved.
 *
 * @Author: Stephen Kapp
 * @Date: 2026-8-19 22:01:08
 * @Last Modified by: Stephen Kapp
 * @Last Modified time: 2026-8-19 22:01:08
 */

package proton

import (
	"strings"

	"github.com/go-resty/resty/v2"
)

const HvPMTokenHeaderField = "x-pm-human-verification-token"
const HvPMTokenType = "x-pm-human-verification-token-type"

func AddHVToRequest(req *resty.Request, hv *APIHVDetails) *resty.Request {
	if hv == nil {
		return req
	}

	return req.SetHeader(HvPMTokenHeaderField, hv.Token).SetHeader(HvPMTokenType, strings.Join(hv.Methods, ","))
}



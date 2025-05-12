/**
 * Copyright © 2020-2025 Stephen Kapp and Reaper Technologies Limited.
 * All Rights Reserved.
 *
 * @Author: Stephen Kapp
 * @Date: 2025-5-12 23:13:18
 * @Last Modified by: Stephen Kapp
 * @Last Modified time: 2025-5-12 23:13:18
 */

package protonsession

import (
	"errors"
)

var (
	ErrorMissingUID          = errors.New("missing UID")
	ErrorMissingAccessToken  = errors.New("missing access token")
	ErrorMissingRefreshToken = errors.New("missing refresh token")
	ErrKeyNotFound           = errors.New("key not found")
	ErrFileNotFound          = errors.New("file not found")
)

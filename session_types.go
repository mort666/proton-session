/**
 * Copyright © 2020-2025 Stephen Kapp and Reaper Technologies Limited.
 * All Rights Reserved.
 *
 * @Author: Stephen Kapp
 * @Date: 2025-5-12 23:09:49
 * @Last Modified by: Stephen Kapp
 * @Last Modified time: 2025-5-12 23:09:49
 */

package protonsession

type SessionConfig = struct {
	UID           string `json:"uid"`
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	SaltedKeyPass string
}

type SessionCredentials struct {
	UID          string `json:"uid"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type SessionStore interface {
	Load() (*SessionConfig, error)
	Save(session *SessionConfig) error
	Delete() error
	List() ([]string, error)
	Switch(account string) error
}

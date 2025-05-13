/**
 * Copyright © 2020-2025 Stephen Kapp and Reaper Technologies Limited.
 * All Rights Reserved.
 *
 * @Author: Stephen Kapp
 * @Date: 2025-5-12 23:09:35
 * @Last Modified by: Stephen Kapp
 * @Last Modified time: 2025-5-12 23:09:35
 */

package protonsession

import (
	"context"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/mort666/go-proton-api"
)

type SessionOptions struct {
	MaxWorkers int
}

type Session struct {
	Client  *proton.Client
	Auth    proton.Auth
	manager *proton.Manager

	MaxWorkers int

	user        proton.User
	UserKeyRing *crypto.KeyRing
}

// Create Session from provided Session credentials. Returns a populated session object
func SessionFromCredentials(ctx context.Context, options []proton.Option, creds *SessionCredentials) (*Session, error) {
	var err error

	// Initialize the client from our cahced credentials
	if creds.UID == "" {
		return nil, ErrorMissingUID
	}

	if creds.AccessToken == "" {
		return nil, ErrorMissingAccessToken
	}

	if creds.RefreshToken == "" {
		return nil, ErrorMissingRefreshToken
	}

	var session Session
	session.MaxWorkers = 10

	session.manager = proton.New(options...)

	session.Client = session.manager.NewClient(creds.UID, creds.AccessToken, creds.RefreshToken)

	session.user, err = session.Client.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// Create Session using the provided AuthToken and RefreshToken, returns a populated session object
func SessionFromRefresh(ctx context.Context, options []proton.Option, creds *SessionCredentials) (*Session, error) {
	var err error

	if creds.UID == "" {
		return nil, ErrorMissingUID
	}

	if creds.RefreshToken == "" {
		return nil, ErrorMissingRefreshToken
	}

	var session Session
	session.MaxWorkers = 10

	session.manager = proton.New(options...)

	session.Client, session.Auth, err = session.manager.NewClientWithRefresh(ctx, creds.UID, creds.RefreshToken)
	if err != nil {
		return nil, err
	}

	session.user, err = session.Client.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// Create Session using provided Login Information, returns pointer to session object
func SessionFromLogin(ctx context.Context, options []proton.Option, username string, password string) (*Session, error) {
	var err error
	session := &Session{}
	session.MaxWorkers = 10
	session.manager = proton.New(options...)

	session.Client, session.Auth, err = session.manager.NewClientWithLogin(ctx, username, []byte(password))
	if err != nil {
		return nil, err
	}

	return session, nil
}

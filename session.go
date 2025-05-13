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

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/rs/zerolog/log"
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

	log.Debug().Msg("session.refresh client")

	session.manager = proton.New(options...)

	log.Debug().Msgf("session.config\n\tuid %s - access_token %s - refresh_token %s", creds.UID, creds.AccessToken, creds.RefreshToken)
	session.Client = session.manager.NewClient(creds.UID, creds.AccessToken, creds.RefreshToken)

	log.Debug().Msg("session.GetUser")
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

	log.Debug().Msg("session.refresh client")

	session.manager = proton.New(options...)

	log.Debug().Msgf("session.config\n\tuid %s - access_token %s - refresh_token %s", creds.UID, creds.AccessToken, creds.RefreshToken)
	session.Client, session.Auth, err = session.manager.NewClientWithRefresh(ctx, creds.UID, creds.RefreshToken)
	if err != nil {
		return nil, err
	}

	log.Debug().Msg("session.GetUser")
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
	log.Debug().Msgf("session.login\n\rusername %s - password %s", username, "<hidden>")
	session.Client, session.Auth, err = session.manager.NewClientWithLogin(ctx, username, []byte(password))
	if err != nil {
		return nil, err
	}

	return session, nil
}

package proton

import (
	"bytes"
	"context"
	"encoding/base64"

	"github.com/ProtonMail/go-srp"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/go-resty/resty/v2"
)

func (m *Manager) NewClient(uid, acc, ref string) *Client {
	return NewClient(uid, m.Client(), WithManager(m)).WithAuth(acc, ref)
}

func (m *Manager) NewClientWithRefresh(ctx context.Context, uid, ref string) (*Client, Auth, error) {
	c := NewClient(uid, m.Client(), WithManager(m))

	authRes, err := m.AuthRefresh(ctx, uid, ref, "")
	if err != nil {
		return nil, Auth{}, err
	}

	return c.WithAuth(authRes.AccessToken, authRes.RefreshToken), authRes, nil
}

func (m *Manager) NewClientWithLogin(ctx context.Context, username string, password []byte) (*Client, Auth, error) {
	return m.NewClientWithLoginWithHVToken(ctx, username, password, nil)
}

func (m *Manager) NewClientWithLoginWithHVToken(ctx context.Context, username string, password []byte, hv *APIHVDetails) (*Client, Auth, error) {
	info, err := m.AuthInfo(ctx, AuthInfoReq{Username: username})
	if err != nil {
		return nil, Auth{}, err
	}

	srpAuth, err := srp.NewAuth(info.Version, username, password, info.Salt, info.Modulus, info.ServerEphemeral)
	if err != nil {
		return nil, Auth{}, err
	}

	proofs, err := srpAuth.GenerateProofs(2048)
	if err != nil {
		return nil, Auth{}, err
	}

	authRes, err := m.auth(ctx, AuthReq{
		Username:        username,
		ClientProof:     base64.StdEncoding.EncodeToString(proofs.ClientProof),
		ClientEphemeral: base64.StdEncoding.EncodeToString(proofs.ClientEphemeral),
		SRPSession:      info.SRPSession,
	}, hv)
	if err != nil {
		return nil, Auth{}, err
	}

	serverProof, err := base64.StdEncoding.DecodeString(authRes.ServerProof)
	if err != nil {
		return nil, Auth{}, err
	}

	if m.verifyProofs {
		if !bytes.Equal(serverProof, proofs.ExpectedServerProof) {
			return nil, Auth{}, ErrInvalidProof
		}
	}

	return NewClient(authRes.UID, m.Client(), WithManager(m)).WithAuth(authRes.AccessToken, authRes.RefreshToken), authRes, nil
}

func (m *Manager) AuthInfo(ctx context.Context, req AuthInfoReq) (AuthInfo, error) {
	var res struct {
		AuthInfo
	}

	if _, err := m.r(ctx).SetBody(req).SetResult(&res).Post("/auth/v4/info"); err != nil {
		return AuthInfo{}, err
	}

	return res.AuthInfo, nil
}

func (m *Manager) AuthModulus(ctx context.Context) (AuthModulus, error) {
	var res AuthModulus

	if _, err := m.r(ctx).SetResult(&res).Get("/auth/v4/modulus"); err != nil {
		return AuthModulus{}, err
	}

	return res, nil
}

func (m *Manager) auth(ctx context.Context, req AuthReq, hv *APIHVDetails) (Auth, error) {
	var res struct {
		Auth
	}

	if _, err := AddHVToRequest(m.r(ctx), hv).SetBody(req).SetResult(&res).Post("/auth/v4"); err != nil {
		return Auth{}, err
	}

	return res.Auth, nil
}

func (m *Manager) AuthRefresh(ctx context.Context, uid, ref, acc string) (Auth, error) {
	state, err := crypto.RandomToken(32)
	if err != nil {
		return Auth{}, err
	}

	req := AuthRefreshReq{
		UID:          uid,
		RefreshToken: ref,
		ResponseType: "token",
		GrantType:    "refresh_token",
		RedirectURI:  "https://protonmail.ch",
		State:        string(state),
		AccessToken:  acc,
	}

	var res struct {
		Auth
	}

	if resp, err := m.r(ctx).SetBody(req).SetResult(&res).Post("/auth/v4/refresh"); err != nil {
		if resp != nil {
			return Auth{}, &resty.ResponseError{Response: resp, Err: err}
		}

		return Auth{}, err
	}

	return res.Auth, nil
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

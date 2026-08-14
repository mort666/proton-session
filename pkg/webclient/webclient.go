/**
 * Copyright © 2020-2026 Stephen Kapp and Reaper Technologies Limited.
 * All Rights Reserved.
 *
 * @Author: Stephen Kapp
 * @Date: 2026-8-14 03:10:04
 * @Last Modified by: Stephen Kapp
 * @Last Modified time: 2026-8-14 03:10:04
 */

package webclient

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/ProtonMail/go-srp"
)

const DefaultWebClientAppVer = "web-account@5.0.407.0" // Setting this here incase version needs updating

type WebClientOption func(*WebApiClient, ...string) error

// WithHttpClient allows providing an alternative http.Client instance.
func WithHttpClient(client *http.Client) WebClientOption {
	return func(opts *WebApiClient, supportedOptions ...string) error {
		opts.HttpClient = client
		return nil
	}
}

// WithAppVersion allows providing an alternative appVersion
func WithAppVersion(appVersion string) WebClientOption {
	return func(opts *WebApiClient, supportedOptions ...string) error {
		opts.AppVersion = appVersion
		return nil
	}
}

// WithUserAgent allows providing an alternative UserAgent
func WithUserAgent(appVersion string) WebClientOption {
	return func(opts *WebApiClient, supportedOptions ...string) error {
		opts.AppVersion = appVersion
		return nil
	}
}

// WithBaseURL allows overriding the API URL.
func WithBaseURL(baseurl string) WebClientOption {
	return func(opts *WebApiClient, supportedOptions ...string) error {
		opts.ApiURLBase = baseurl
		return nil
	}
}

func randomUserAgent() string {
	var seed [32]byte
	_, _ = crand.Read(seed[:])
	generator := rand.NewChaCha8(seed)

	// Pick a random user agent from this list. Because I'm not going to tell
	// Proton shit on where all these funny requests are coming from, given their
	// unhelpfulness in figuring out their authentication flow.
	UserAgents := [...]string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:143.0) Gecko/20100101 Firefox/143.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:143.0) Gecko/20100101 Firefox/143.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Safari/537.36 Edg/151.0.4129.78",
		"Mozilla/5.0 (X11; Linux x86_64; rv:143.0) Gecko/20100101 Firefox/143.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:153.0) Gecko/20100101 Firefox/153.0",
		"Mozilla/5.0 (X11; Linux i686; rv:153.0) Gecko/20100101 Firefox/153.0",
	}
	UserAgent := UserAgents[generator.Uint64()%uint64(len(UserAgents))]

	return UserAgent
}

// WebApiClient is a minimal Proton v4 API client which can handle all the
// oddities of Proton's authentication flow they want to keep hidden
// from the public.
type WebApiClient struct {
	ApiURLBase string
	HttpClient *http.Client
	AppVersion string
	UserAgent  string
	generator  *rand.ChaCha8
}

// newWebApiClient returns an [WebApiClient] with sane defaults matching Proton's
// insane expectations.
func NewWebAPIClient(ctx context.Context, opts ...WebClientOption) (client *WebApiClient, err error) {
	var seed [32]byte
	_, _ = crand.Read(seed[:])
	generator := rand.NewChaCha8(seed)

	apiclient := &WebApiClient{
		ApiURLBase: "https://account.proton.me/api",
		HttpClient: http.DefaultClient,
		AppVersion: DefaultWebClientAppVer,
		UserAgent:  randomUserAgent(),
		generator:  generator,
	}

	for _, opt := range opts {
		opt(apiclient)
	}

	return apiclient, nil
}

// SetHeaders sets the minimal necessary headers for Proton API requests
// to succeed without being blocked by their "security" measures.
// See for example [getMostRecentStableTag] on how the app version must
// be set to a recent version or they block your request. "SeCuRiTy"...
func (c *WebApiClient) SetHeaders(request *http.Request, cookie Cookie) {
	request.Header.Set("Cookie", cookie.String())
	request.Header.Set("User-Agent", c.UserAgent)
	request.Header.Set("x-pm-appversion", c.AppVersion)
	request.Header.Set("x-pm-locale", "en_US")
	request.Header.Set("x-pm-uid", cookie.uid)
}

// SetUserAgent sets the useragent to the user provided value
func (c *WebApiClient) SetAPIBaseURL(baseurl string) {
	c.ApiURLBase = baseurl
}

// SetUserAgent sets the useragent to the user provided value
func (c *WebApiClient) SetUserAgent(useragent string) {
	c.UserAgent = useragent
}

// authenticate performs the full Proton authentication flow
// to obtain an authenticated cookie (uid, token and session ID).
func (c *WebApiClient) Authenticate(ctx context.Context, email, password string,
) (authCookie Cookie, err error) {
	sessionID, err := c.GetSessionID(ctx)
	if err != nil {
		return Cookie{}, fmt.Errorf("getting session ID: %w", err)
	}

	tokenType, accessToken, refreshToken, uid, err := c.GetUnauthSession(ctx, sessionID)
	if err != nil {
		return Cookie{}, fmt.Errorf("getting unauthenticated session data: %w", err)
	}

	cookieToken, err := c.CookieToken(ctx, sessionID, tokenType, accessToken, refreshToken, uid)
	if err != nil {
		return Cookie{}, fmt.Errorf("getting cookie token: %w", err)
	}

	unauthCookie := Cookie{
		uid:       uid,
		token:     cookieToken,
		sessionID: sessionID,
	}
	username, modulusPGPClearSigned, serverEphemeralBase64, saltBase64,
		srpSessionHex, version, err := c.AuthInfo(ctx, email, unauthCookie)
	if err != nil {
		return Cookie{}, fmt.Errorf("getting auth information: %w", err)
	}

	// Prepare SRP proof generator using Proton's official SRP parameters and hashing.
	srpAuth, err := srp.NewAuth(version, username, []byte(password),
		saltBase64, modulusPGPClearSigned, serverEphemeralBase64)
	if err != nil {
		return Cookie{}, fmt.Errorf("initializing SRP auth: %w", err)
	}

	// Generate SRP proofs (A, M1) with the usual 2048-bit modulus.
	const modulusBits = 2048
	proofs, err := srpAuth.GenerateProofs(modulusBits)
	if err != nil {
		return Cookie{}, fmt.Errorf("generating SRP proofs: %w", err)
	}

	authCookie, err = c.Auth(ctx, unauthCookie, email, srpSessionHex, proofs)
	if err != nil {
		return Cookie{}, fmt.Errorf("authentifying: %w", err)
	}

	return authCookie, nil
}

func (c *WebApiClient) GetSessionID(ctx context.Context) (sessionID string, err error) {
	const url = "https://account.proton.me/vpn"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	response, err := c.HttpClient.Do(request)
	if err != nil {
		return "", err
	}
	err = response.Body.Close()
	if err != nil {
		return "", fmt.Errorf("closing response body: %w", err)
	}

	for _, cookie := range response.Cookies() {
		if cookie.Name == "Session-Id" {
			return cookie.Value, nil
		}
	}

	return "", errors.New("session ID not found in cookies")
}

func (c *WebApiClient) GetUnauthSession(ctx context.Context, sessionID string) (
	tokenType, accessToken, refreshToken, uid string, err error,
) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.ApiURLBase+"/auth/v4/sessions", nil)
	if err != nil {
		return "", "", "", "", fmt.Errorf("creating request: %w", err)
	}
	unauthCookie := Cookie{
		sessionID: sessionID,
	}
	c.SetHeaders(request, unauthCookie)

	response, err := c.HttpClient.Do(request)
	if err != nil {
		return "", "", "", "", err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", "", "", "", fmt.Errorf("reading response body: %w", err)
	} else if response.StatusCode != http.StatusOK {
		return "", "", "", "", buildError(response.StatusCode, responseBody)
	}

	var data struct {
		Code         uint     `json:"Code"`         // 1000 on success
		AccessToken  string   `json:"AccessToken"`  // 32-chars lowercase and digits
		RefreshToken string   `json:"RefreshToken"` // 32-chars lowercase and digits
		TokenType    string   `json:"TokenType"`    // "Bearer"
		Scopes       []string `json:"Scopes"`       // should be [] for our usage
		UID          string   `json:"UID"`          // 32-chars lowercase and digits
		LocalID      uint     `json:"LocalID"`      // 0 in my case
	}

	err = json.Unmarshal(responseBody, &data)
	if err != nil {
		return "", "", "", "", fmt.Errorf("decoding response body: %w", err)
	}

	const successCode = 1000
	switch {
	case data.Code != successCode:
		return "", "", "", "", fmt.Errorf("response code %d is not expected success code %d",
			data.Code, successCode)
	case data.AccessToken == "":
		return "", "", "", "", errors.New("access token is empty in response")
	case data.RefreshToken == "":
		return "", "", "", "", errors.New("refresh token is empty in response")
	case data.TokenType == "":
		return "", "", "", "", errors.New("token type is empty in response")
	case data.UID == "":
		return "", "", "", "", errors.New("UID is empty in response")
	}
	// Ignore Scopes and LocalID fields, we don't use them.

	return data.TokenType, data.AccessToken, data.RefreshToken, data.UID, nil
}

func (c *WebApiClient) CookieToken(ctx context.Context, sessionID, tokenType, accessToken,
	refreshToken, uid string,
) (cookieToken string, err error) {
	type requestBodySchema struct {
		GrantType    string `json:"GrantType"`    // "refresh_token"
		Persistent   uint   `json:"Persistent"`   // 0
		RedirectURI  string `json:"RedirectURI"`  // "https://protonmail.com"
		RefreshToken string `json:"RefreshToken"` // 32-chars lowercase and digits
		ResponseType string `json:"ResponseType"` // "token"
		State        string `json:"State"`        // 24-chars letters and digits
		UID          string `json:"UID"`          // 32-chars lowercase and digits
	}
	requestBody := requestBodySchema{
		GrantType:    "refresh_token",
		Persistent:   0,
		RedirectURI:  "https://protonmail.com",
		RefreshToken: refreshToken,
		ResponseType: "token",
		State:        generateLettersDigits(c.generator, 24), //nolint:mnd
		UID:          uid,
	}

	buffer := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buffer)
	if err := encoder.Encode(requestBody); err != nil { //nolint:gosec
		return "", fmt.Errorf("encoding request body: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.ApiURLBase+"/core/v4/auth/cookies", buffer)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	unauthCookie := Cookie{
		uid:       uid,
		sessionID: sessionID,
	}
	c.SetHeaders(request, unauthCookie)
	request.Header.Set("Authorization", tokenType+" "+accessToken)

	response, err := c.HttpClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	} else if response.StatusCode != http.StatusOK {
		return "", buildError(response.StatusCode, responseBody)
	}

	var cookies struct {
		Code           uint   `json:"Code"`           // 1000 on success
		UID            string `json:"UID"`            // should match request UID
		LocalID        uint   `json:"LocalID"`        // 0
		RefreshCounter uint   `json:"RefreshCounter"` // 1
	}
	err = json.Unmarshal(responseBody, &cookies)
	if err != nil {
		return "", fmt.Errorf("decoding response body: %w", err)
	}

	const successCode = 1000
	switch {
	case cookies.Code != successCode:
		return "", fmt.Errorf("response code %d is not expected success code %d",
			cookies.Code, successCode)
	case cookies.UID != requestBody.UID:
		return "", fmt.Errorf("UID %s in response does not match request UID %s",
			cookies.UID, requestBody.UID)
	}
	// Ignore LocalID and RefreshCounter fields, we don't use them.

	for _, cookie := range response.Cookies() {
		if cookie.Name == "AUTH-"+uid {
			return cookie.Value, nil
		}
	}

	return "", errors.New("auth cookie not found")
}

// authInfo fetches SRP parameters for the account.
func (c *WebApiClient) AuthInfo(ctx context.Context, email string, unauthCookie Cookie) (
	username, modulusPGPClearSigned, serverEphemeralBase64, saltBase64, srpSessionHex string,
	version int, err error,
) {
	type requestBodySchema struct {
		Intent   string `json:"Intent"` // "Proton"
		Username string `json:"Username"`
	}
	requestBody := requestBodySchema{
		Intent:   "Proton",
		Username: email,
	}

	buffer := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buffer)
	if err := encoder.Encode(requestBody); err != nil {
		return "", "", "", "", "", 0, fmt.Errorf("encoding request body: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.ApiURLBase+"/core/v4/auth/info", buffer)
	if err != nil {
		return "", "", "", "", "", 0, fmt.Errorf("creating request: %w", err)
	}
	c.SetHeaders(request, unauthCookie)
	request.Header.Set("Content-Type", "application/json")

	response, err := c.HttpClient.Do(request)
	if err != nil {
		return "", "", "", "", "", 0, err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", "", "", "", "", 0, fmt.Errorf("reading response body: %w", err)
	} else if response.StatusCode != http.StatusOK {
		return "", "", "", "", "", 0, buildError(response.StatusCode, responseBody)
	}

	var info struct {
		Code            uint   `json:"Code"`              // 1000 on success
		Modulus         string `json:"Modulus"`           // PGP clearsigned modulus string
		ServerEphemeral string `json:"ServerEphemeral"`   // base64
		Version         *uint  `json:"Version,omitempty"` // 4 as of 2025-10-26
		Salt            string `json:"Salt"`              // base64
		SRPSession      string `json:"SRPSession"`        // hexadecimal
		Username        string `json:"Username"`          // user without @domain.com. Mine has its first letter capitalized.
	}
	err = json.Unmarshal(responseBody, &info)
	if err != nil {
		return "", "", "", "", "", 0, fmt.Errorf("decoding response body: %w", err)
	}

	const successCode = 1000
	switch {
	case info.Code != successCode:
		return "", "", "", "", "", 0, fmt.Errorf("response code %d is not expected success code %d",
			info.Code, successCode)
	case info.Modulus == "":
		return "", "", "", "", "", 0, errors.New("modulus is empty in response")
	case info.ServerEphemeral == "":
		return "", "", "", "", "", 0, errors.New("server ephemeral is empty in response")
	case info.Salt == "":
		return "", "", "", "", "", 0, errors.New("salt is empty in response")
	case info.SRPSession == "":
		return "", "", "", "", "", 0, errors.New("SRP session is empty in response")
	case info.Username == "":
		return "", "", "", "", "", 0, errors.New("username is empty in response")
	case info.Version == nil:
		return "", "", "", "", "", 0, errors.New("version is missing in response")
	}

	version = int(*info.Version) //nolint:gosec
	return info.Username, info.Modulus, info.ServerEphemeral, info.Salt,
		info.SRPSession, version, nil
}

type Cookie struct {
	uid       string
	token     string
	sessionID string
}

func (c *Cookie) String() string {
	s := ""
	if c.token != "" {
		s += fmt.Sprintf("AUTH-%s=%s; ", c.uid, c.token)
	}
	if c.sessionID != "" {
		s += fmt.Sprintf("Session-Id=%s; ", c.sessionID)
	}
	if c.token != "" {
		s += "Tag=default; iaas=W10; Domain=proton.me; Feature=VPNDashboard:A"
	}
	return s
}

// ErrServerProofNotValid indicates the M2 from the server didn't match the expected proof.

// auth performs the SRP proof submission (and optionally TOTP) to obtain tokens.
func (c *WebApiClient) Auth(ctx context.Context, unauthCookie Cookie,
	username, srpSession string, proofs *srp.Proofs,
) (authCookie Cookie, err error) {
	clientEphemeral := base64.StdEncoding.EncodeToString(proofs.ClientEphemeral)
	clientProof := base64.StdEncoding.EncodeToString(proofs.ClientProof)

	type requestBodySchema struct {
		ClientEphemeral string            `json:"ClientEphemeral"`   // base64(A)
		ClientProof     string            `json:"ClientProof"`       // base64(M1)
		Payload         map[string]string `json:"Payload,omitempty"` // not sure
		SRPSession      string            `json:"SRPSession"`        // hexadecimal
		Username        string            `json:"Username"`          // user@protonmail.com
	}
	requestBody := requestBodySchema{
		ClientEphemeral: clientEphemeral,
		ClientProof:     clientProof,
		SRPSession:      srpSession,
		Username:        username,
	}

	buffer := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buffer)
	if err := encoder.Encode(requestBody); err != nil {
		return Cookie{}, fmt.Errorf("encoding request body: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.ApiURLBase+"/core/v4/auth", buffer)
	if err != nil {
		return Cookie{}, fmt.Errorf("creating request: %w", err)
	}
	c.SetHeaders(request, unauthCookie)
	request.Header.Set("Content-Type", "application/json")

	response, err := c.HttpClient.Do(request)
	if err != nil {
		return Cookie{}, err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return Cookie{}, fmt.Errorf("reading response body: %w", err)
	} else if response.StatusCode != http.StatusOK {
		return Cookie{}, buildError(response.StatusCode, responseBody)
	}

	type twoFAStatus uint
	//nolint:unused
	const (
		twoFADisabled twoFAStatus = iota
		twoFAHasTOTP
		twoFAHasFIDO2
		twoFAHasFIDO2AndTOTP
	)
	type twoFAInfo struct {
		Enabled twoFAStatus `json:"Enabled"`
		FIDO2   struct {
			AuthenticationOptions any   `json:"AuthenticationOptions"`
			RegisteredKeys        []any `json:"RegisteredKeys"`
		} `json:"FIDO2"`
		TOTP uint `json:"TOTP"`
	}

	var auth struct {
		Code              uint      `json:"Code"`         // 1000 on success
		LocalID           uint      `json:"LocalID"`      // 7 in my case
		Scopes            []string  `json:"Scopes"`       // this should contain "vpn". Same as `Scope` field value.
		UID               string    `json:"UID"`          // same as `Uid` field value
		UserID            string    `json:"UserID"`       // base64
		EventID           string    `json:"EventID"`      // base64
		PasswordMode      uint      `json:"PasswordMode"` // 1 in my case
		ServerProof       string    `json:"ServerProof"`  // base64(M2)
		TwoFactor         uint      `json:"TwoFactor"`    // 0 if 2FA not required
		TwoFA             twoFAInfo `json:"2FA"`
		TemporaryPassword uint      `json:"TemporaryPassword"` // 0 in my case
	}

	err = json.Unmarshal(responseBody, &auth)
	if err != nil {
		return Cookie{}, fmt.Errorf("decoding response body: %w", err)
	}

	m2, err := base64.StdEncoding.DecodeString(auth.ServerProof)
	if err != nil {
		return Cookie{}, fmt.Errorf("decoding server proof: %w", err)
	}
	if !bytes.Equal(m2, proofs.ExpectedServerProof) {
		return Cookie{}, fmt.Errorf("server proof from server %x is not expected proof %x",
			m2, proofs.ExpectedServerProof)
	}

	const successCode = 1000
	switch {
	case auth.Code != successCode:
		return Cookie{}, fmt.Errorf("response code %d is not expected success code %d",
			auth.Code, successCode)
	case auth.UID != unauthCookie.uid:
		return Cookie{}, fmt.Errorf("UID %s in response does not match request UID %s",
			auth.UID, unauthCookie.uid)
	case auth.TwoFactor != 0:
		return Cookie{}, errors.New("two factor authentication not supported in this client")
	case !slices.Contains(auth.Scopes, "vpn"):
		return Cookie{}, fmt.Errorf("VPN scope not found in scopes %v", auth.Scopes)
	}

	for _, setCookieHeader := range response.Header.Values("Set-Cookie") {
		parts := strings.Split(setCookieHeader, ";")
		for _, part := range parts {
			if strings.HasPrefix(part, "AUTH-"+unauthCookie.uid+"=") {
				authCookie = unauthCookie
				authCookie.token = strings.TrimPrefix(part, "AUTH-"+unauthCookie.uid+"=")
				return authCookie, nil
			}
		}
	}

	return Cookie{}, fmt.Errorf("auth cookie not found in HTTP headers %s", httpHeadersToString(response.Header))
}

// generateLettersDigits mimicing Proton's own random string generator:
// https://github.com/ProtonMail/WebClients/blob/e4d7e4ab9babe15b79a131960185f9f8275512cd/packages/utils/generateLettersDigits.ts
func generateLettersDigits(rng *rand.ChaCha8, length uint) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	return generateFromCharset(rng, length, charset)
}

func generateFromCharset(rng *rand.ChaCha8, length uint, charset string) string {
	result := make([]byte, length)
	randomBytes := make([]byte, length)
	_, _ = rng.Read(randomBytes)
	for i := range length {
		result[i] = charset[int(randomBytes[i])%len(charset)]
	}
	return string(result)
}

func httpHeadersToString(headers http.Header) string {
	var builder strings.Builder
	first := true
	for key, values := range headers {
		for _, value := range values {
			if !first {
				builder.WriteString(", ")
			}
			fmt.Fprintf(&builder, "%s: %s", key, value)
			first = false
		}
	}
	return builder.String()
}

func buildError(httpCode int, body []byte) error {
	prettyCode := http.StatusText(httpCode)
	var protonError struct {
		Code    *int              `json:"Code,omitempty"`
		Error   *string           `json:"Error,omitempty"`
		Details map[string]string `json:"Details"`
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&protonError)
	if err != nil || protonError.Error == nil || protonError.Code == nil {
		return fmt.Errorf("HTTP status code not OK: %s: %s",
			prettyCode, body)
	}

	details := make([]string, 0, len(protonError.Details))
	for key, value := range protonError.Details {
		details = append(details, fmt.Sprintf("%s: %s", key, value))
	}

	return fmt.Errorf("HTTP status code not OK: %s: %s (code %d with details: %s)",
		prettyCode, *protonError.Error, *protonError.Code, strings.Join(details, ", "))
}

func (c *WebApiClient) Do(ctx context.Context, cookie Cookie, url string) (
	*http.Response, error,
) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.SetHeaders(request, cookie)

	response, err := c.HttpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(response.Body)
		return response, buildError(response.StatusCode, b)
	}

	return response, nil
}

func ReplaceParameters(stringWithParams string, params map[string]string) string {
	var paramRegex = regexp.MustCompile(`({.*?})`)
	if len(params) == 0 {
		return stringWithParams
	}

	return paramRegex.ReplaceAllStringFunc(stringWithParams, func(match string) string {
		match = match[1 : len(match)-1]
		return params[match]
	})
}

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

var ErrUnsupportedOption = errors.New("unsupported option")

const (
	SupportedOptionRetries              = "retries"
	SupportedOptionTimeout              = "timeout"
	SupportedOptionAcceptHeaderOverride = "acceptHeaderOverride"
	SupportedOptionURLOverride          = "urlOverride"
)

/**
 * Copyright © 2020-2026 Stephen Kapp and Reaper Technologies Limited.
 * All Rights Reserved.
 *
 * @Author: Stephen Kapp
 * @Date: 2026-8-19 22:01:53
 * @Last Modified by: Stephen Kapp
 * @Last Modified time: 2026-8-19 22:01:53
 */

package proton

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultHostURL is the default host of the API.
	DefaultHostURL = "https://mail.me/api"

	// DefaultAppVersion is the default app version used to communicate with the API.
	// This must be changed (using the WithAppVersion option) for production use.
	DefaultAppVersion = "go-proton-api"

	// DefaultUserAgent is the default user agent used to communicate with the API.
	// See: https://github.com/emersion/hydroxide/issues/252
	DefaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36"
)

type ManagerFunc func() *Manager

func WithManager(rc *Manager) ManagerFunc {
  return func() *Manager {
    return rc
  }
}

type ManagerBuilder struct {
	HostURL      string
	AppVersion   string
	UserAgent    string
	Transport    http.RoundTripper
	VerifyProofs bool
	CookieJar    http.CookieJar
	RetryCount   int
	Logger       resty.Logger
	Debug        bool
	PanicHandler async.PanicHandler
}

func NewManagerBuilder() *ManagerBuilder {
	return &ManagerBuilder{
		HostURL:      DefaultHostURL,
		AppVersion:   DefaultAppVersion,
		UserAgent:    DefaultUserAgent,
		Transport:    http.DefaultTransport,
		VerifyProofs: true,
		CookieJar:    nil,
		RetryCount:   3,
		Logger:       nil,
		Debug:        false,
		PanicHandler: async.NoopPanicHandler{},
	}
}

func DefaultManager(builder *ManagerBuilder) *Manager {
		return &Manager{
		rc: resty.New(),

		errHandlers: make(map[Code][]Handler),

		verifyProofs: builder.VerifyProofs,

		panicHandler: builder.PanicHandler,
	}
}

func (builder *ManagerBuilder) Build() *Manager {

	m := DefaultManager(builder)

	// Set the API host.
	m.Client().SetBaseURL(builder.HostURL)

	// Set the transport.
	m.Client().SetTransport(builder.Transport)

	// Set the cookie jar.
	m.Client().SetCookieJar(builder.CookieJar)

	// Set the logger.
	if builder.Logger != nil {
		m.Client().SetLogger(builder.Logger)
	}

	// Set the debug flag.
	m.Client().SetDebug(builder.Debug)

	// Set app version in header.
	m.Client().OnBeforeRequest(func(_ *resty.Client, req *resty.Request) error {
		req.SetHeader("x-pm-appversion", builder.AppVersion)
		req.SetHeader("User-Agent", builder.UserAgent)
		return nil
	})

	// Set middleware.
	m.Client().OnAfterResponse(catchAPIError)
	m.Client().OnAfterResponse(updateTime)
	m.Client().OnAfterResponse(m.CheckConnUp)
	m.Client().OnError(m.CheckConnDown)
	m.Client().OnError(m.HandleError)

	// Configure retry mechanism.
	m.Client().SetRetryCount(builder.RetryCount)
	m.Client().SetRetryMaxWaitTime(time.Minute)
	m.Client().AddRetryCondition(catchTooManyRequests)
	m.Client().AddRetryCondition(catchDialError)
	m.Client().AddRetryCondition(catchDropError)
	m.Client().SetRetryAfter(catchRetryAfter)

	// Set the data type of API 
	m.Client().SetError(&APIError{})

	return m
}

func updateTime(_ *resty.Client, res *resty.Response) error {
	date, err := time.Parse(time.RFC1123, res.Header().Get("Date"))
	if err != nil {
		return err
	}

	crypto.UpdateTime(date.Unix())

	return nil
}


func catchAPIError(_ *resty.Client, res *resty.Response) error {
	if !res.IsError() {
		return nil
	}

	method := "NONE"
	route := "N/A"

	if res.Request != nil {
		method = res.Request.Method
		route = res.Request.URL
	}

	var err error

	if apiErr, ok := res.Error().(*APIError); ok {
		apiErr.Status = res.StatusCode()
		err = apiErr
	} else {
		statusCode := res.StatusCode()
		statusText := res.Status()

		// Catch error that may slip through when APIError deserialization routine fails for whichever reason.
		if statusCode >= 400 {
			err = &APIError{
				Status:  statusCode,
				Code:    0,
				Message: statusText,
			}
		} else {
			err = fmt.Errorf("%v", res.Status())
		}
	}

	return fmt.Errorf(
		"%v %s %s: %w",
		res.StatusCode(), method, route, err,
	)
}

// nolint:gosec
func catchRetryAfter(_ *resty.Client, res *resty.Response) (time.Duration, error) {
	var log = logrus.WithField("pkg", "gpa")
	// 0 and no error means default behaviour which is exponential backoff with jitter.
	if res.StatusCode() != http.StatusTooManyRequests && res.StatusCode() != http.StatusServiceUnavailable {
		return 0, nil
	}

	// Parse the Retry-After header, or fallback to 10 seconds.
	after, err := strconv.Atoi(res.Header().Get("Retry-After"))
	if err != nil {
		after = 10
	}

	// Add some jitter to the delay.
	after += rand.Intn(10)

	log.WithFields(logrus.Fields{
		"status": res.StatusCode(),
		"url":    res.Request.URL,
		"method": res.Request.Method,
		"after":  after,
	}).Warn("Too many requests, retrying after delay")

	return time.Duration(after) * time.Second, nil
}

func catchTooManyRequests(res *resty.Response, _ error) bool {
	return res.StatusCode() == http.StatusTooManyRequests || res.StatusCode() == http.StatusServiceUnavailable
}

func catchDialError(res *resty.Response, err error) bool {
	return res.RawResponse == nil
}

func catchDropError(_ *resty.Response, err error) bool {
	if netErr := new(net.OpError); As(err, &netErr) {
		return true
	}

	return false
}

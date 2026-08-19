/**
 * Copyright © 2020-2026 Stephen Kapp and Reaper Technologies Limited.
 * All Rights Reserved.
 *
 * @Author: Stephen Kapp
 * @Date: 2026-8-19 22:00:48
 * @Last Modified by: Stephen Kapp
 * @Last Modified time: 2026-8-19 22:00:48
 */

package proton

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/go-resty/resty/v2"
)

type RestyClientFunc func(c *Client) *resty.Client

func WithRestyClient(rc *resty.Client) RestyClientFunc {
	return func(c *Client) *resty.Client {
		return rc
	}
}

// clientID is a unique identifier for a
var ClientID uint64

// AuthHandler is given any new auths that are returned from the API due to an unexpected auth refresh.
type AuthHandler func(Auth)

// Handler is a generic function that can be registered for a certain event (e.g. deauth, API code).
type Handler func()

// Client is the proton
type Client struct {
	client *resty.Client
	// clientID is this client's unique ID.
	ClientID uint64
	m        ManagerFunc
	uid      string
	acc      string
	ref      string
	authLock sync.RWMutex

	authHandlers   []AuthHandler
	deauthHandlers []Handler
	hookLock       sync.RWMutex

	deauthOnce sync.Once
}

func NewClient(uid string, m *resty.Client, manager ManagerFunc) *Client {
	c := &Client{
		client:   m,
		m:        manager,
		uid:      uid,
		ClientID: atomic.AddUint64(&ClientID, 1),
	}

	return c
}

func (c *Client) AddAuthHandler(handler AuthHandler) {
	c.hookLock.Lock()
	defer c.hookLock.Unlock()

	c.authHandlers = append(c.authHandlers, handler)
}

func (c *Client) AddDeauthHandler(handler Handler) {
	c.hookLock.Lock()
	defer c.hookLock.Unlock()

	c.deauthHandlers = append(c.deauthHandlers, handler)
}

func (c *Client) AddPreRequestHook(hook resty.RequestMiddleware) {
	c.hookLock.Lock()
	defer c.hookLock.Unlock()

	c.m().rc.OnBeforeRequest(func(rc *resty.Client, r *resty.Request) error {
		if ClientID, ok := ClientIDFromContext(r.Context()); !ok || ClientID != c.ClientID {
			return nil
		}

		return hook(rc, r)
	})
}

func (c *Client) AddPostRequestHook(hook resty.ResponseMiddleware) {
	c.hookLock.Lock()
	defer c.hookLock.Unlock()

	c.m().rc.OnAfterResponse(func(rc *resty.Client, r *resty.Response) error {
		if ClientID, ok := ClientIDFromContext(r.Request.Context()); !ok || ClientID != c.ClientID {
			return nil
		}

		return hook(rc, r)
	})
}

func (c *Client) Close() {
	c.authLock.Lock()
	defer c.authLock.Unlock()

	c.uid = ""
	c.acc = ""
	c.ref = ""

	c.hookLock.Lock()
	defer c.hookLock.Unlock()

	c.authHandlers = nil
	c.deauthHandlers = nil
}

func (c *Client) WithAuth(acc, ref string) *Client {
	c.acc = acc
	c.ref = ref

	return c
}

func (c *Client) do(ctx context.Context, fn func(*resty.Request) (*resty.Response, error)) error {
	if _, err := c.doRes(ctx, fn); err != nil {
		return err
	}

	return nil
}

func (c *Client) doRes(ctx context.Context, fn func(*resty.Request) (*resty.Response, error)) (*resty.Response, error) {
	c.hookLock.RLock()
	defer c.hookLock.RUnlock()

	res, err := c.exec(ctx, fn)

	if res != nil {
		// If we receive no response, we can't do anything.
		if res.RawResponse == nil {
			return nil, NewNetError(err, "received no response from API")
		}

		// If we receive a net error, we can't do anything.
		if resErr, ok := err.(*resty.ResponseError); ok {
			if netErr := new(net.OpError); As(resErr.Err, &netErr) {
				return nil, NewNetError(netErr, "network error while communicating with API")
			}
		}

		// If we receive a 401, we need to refresh the
		if res.StatusCode() == http.StatusUnauthorized {
			if err := c.authRefresh(ctx); err != nil {
				return nil, fmt.Errorf("failed to refresh auth: %w", err)
			}

			if res, err = c.exec(ctx, fn); err != nil {
				return nil, fmt.Errorf("failed to retry request: %w", err)
			}
		}
	}

	return res, err
}

func (c *Client) exec(ctx context.Context, fn func(*resty.Request) (*resty.Response, error)) (*resty.Response, error) {
	c.authLock.RLock()
	defer c.authLock.RUnlock()

	r := c.m().R(WithClient(ctx, c.ClientID))

	if c.uid != "" {
		r.SetHeader("x-pm-uid", c.uid)
	}

	if c.acc != "" {
		r.SetAuthToken(c.acc)
	}

	return fn(r)
}

func (c *Client) authRefresh(ctx context.Context) error {
	c.authLock.Lock()
	defer c.authLock.Unlock()

	c.hookLock.RLock()
	defer c.hookLock.RUnlock()

	auth, err := c.m().AuthRefresh(ctx, c.uid, c.ref, c.acc)

	if err != nil {
		if respErr, ok := err.(*resty.ResponseError); ok {

			switch respErr.Response.StatusCode() {
			case http.StatusBadRequest, http.StatusUnprocessableEntity:
				c.deauthOnce.Do(func() {
					for _, handler := range c.deauthHandlers {
						handler()
					}
				})

				return fmt.Errorf("failed to refresh auth, de-auth: %w", err)
			case http.StatusConflict, http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusServiceUnavailable:
				return fmt.Errorf("failed to refresh auth, server issues: %w", err)
			default:
				//
			}
		}

		return fmt.Errorf("failed to refresh auth: %w", err)
	}

	c.acc = auth.AccessToken
	c.ref = auth.RefreshToken

	for _, handler := range c.authHandlers {
		handler(auth)
	}

	return nil
}

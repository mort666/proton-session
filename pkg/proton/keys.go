/**
 * Copyright © 2020-2026 Stephen Kapp and Reaper Technologies Limited.
 * All Rights Reserved.
 *
 * @Author: Stephen Kapp
 * @Date: 2026-8-19 22:01:27
 * @Last Modified by: Stephen Kapp
 * @Last Modified time: 2026-8-19 22:01:27
 */

package proton

import (
	"context"

	"github.com/go-resty/resty/v2"
)

func (c *Client) CreateAddressKey(ctx context.Context, req CreateAddressKeyReq) (Key, error) {
	var res struct {
		Key Key
	}

	if err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetBody(req).SetResult(&res).Post("/core/v4/keys/address")
	}); err != nil {
		return Key{}, err
	}

	return res.Key, nil
}

func (c *Client) CreateLegacyAddressKey(ctx context.Context, req CreateAddressKeyReq) (Key, error) {
	var res struct {
		Key Key
	}

	if err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetBody(req).SetResult(&res).Post("/core/v4/keys")
	}); err != nil {
		return Key{}, err
	}

	return res.Key, nil
}

func (c *Client) MakeAddressKeyPrimary(ctx context.Context, keyID string, keyList KeyList) error {
	return c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetBody(struct{ SignedKeyList KeyList }{SignedKeyList: keyList}).Put("/core/v4/keys/" + keyID + "/primary")
	})
}

func (c *Client) DeleteAddressKey(ctx context.Context, keyID string, keyList KeyList) error {
	return c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetBody(struct{ SignedKeyList KeyList }{SignedKeyList: keyList}).Post("/core/v4/keys/" + keyID + "/delete")
	})
}

/**
 * Copyright © 2020-2026 Stephen Kapp and Reaper Technologies Limited.
 * All Rights Reserved.
 *
 * @Author: Stephen Kapp
 * @Date: 2026-8-19 22:00:14
 * @Last Modified by: Stephen Kapp
 * @Last Modified time: 2026-8-19 22:00:14
 */

package proton

import (
	"context"

	"github.com/go-resty/resty/v2"
)

func (c *Client) GetUser(ctx context.Context) (User, error) {
	return c.GetUserWithHV(ctx, nil)
}

func (c *Client) GetUserWithHV(ctx context.Context, hv *APIHVDetails) (User, error) {
	var res struct {
		User User
	}

	if _, err := c.doRes(ctx, func(r *resty.Request) (*resty.Response, error) {
		return AddHVToRequest(r, hv).SetResult(&res).Get("/core/v4/users")
	}); err != nil {
		return User{}, err
	}

	return res.User, nil
}

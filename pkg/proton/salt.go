/**
 * Copyright © 2020-2026 Stephen Kapp and Reaper Technologies Limited.
 * All Rights Reserved.
 *
 * @Author: Stephen Kapp
 * @Date: 2026-8-19 22:00:29
 * @Last Modified by: Stephen Kapp
 * @Last Modified time: 2026-8-19 22:00:29
 */

package proton

import (
	"context"

	"github.com/go-resty/resty/v2"
)

func (c *Client) GetSalts(ctx context.Context) (Salts, error) {
	var res struct {
		KeySalts []Salt
	}

	if err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetResult(&res).Get("/core/v4/keys/salts")
	}); err != nil {
		return nil, err
	}

	return res.KeySalts, nil
}

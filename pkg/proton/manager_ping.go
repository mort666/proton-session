/**
 * Copyright © 2020-2026 Stephen Kapp and Reaper Technologies Limited.
 * All Rights Reserved.
 *
 * @Author: Stephen Kapp
 * @Date: 2026-8-19 22:02:10
 * @Last Modified by: Stephen Kapp
 * @Last Modified time: 2026-8-19 22:02:10
 */

package proton

import "context"

func (m *Manager) Ping(ctx context.Context) error {
	if res, err := m.r(ctx).Get("/tests/ping"); err != nil {
		if res.RawResponse != nil {
			return nil
		}

		return err
	}

	return nil
}

/**
 * Copyright © 2020-2025 Stephen Kapp and Reaper Technologies Limited.
 * All Rights Reserved.
 *
 * @Author: Stephen Kapp
 * @Date: 2025-5-12 23:13:06
 * @Last Modified by: Stephen Kapp
 * @Last Modified time: 2025-5-12 23:13:06
 */

package protonsession

import (
	"context"

	p "github.com/mort666/go-proton-api"
)

type Manager p.Manager

// Salt the provided keypass. The salted keypass is used to Unlock() the account.
func SaltKeyPass(ctx context.Context, client *p.Client, password []byte) ([]byte, error) {
	user, err := client.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	salts, err := client.GetSalts(ctx)
	if err != nil {
		return nil, err
	}

	saltedKeypass, err := salts.SaltForKey(password, user.Keys.Primary().ID)
	if err != nil {
		return nil, err
	}

	return saltedKeypass, nil
}

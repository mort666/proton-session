/**
 * Copyright © 2020-2026 Stephen Kapp and Reaper Technologies Limited.
 * All Rights Reserved.
 *
 * @Author: Stephen Kapp
 * @Date: 2026-8-19 22:02:37
 * @Last Modified by: Stephen Kapp
 * @Last Modified time: 2026-8-19 22:02:37
 */

package proton

import (
	"encoding/base64"
	"fmt"
	"slices"

	"github.com/ProtonMail/go-srp"
)

type Salt struct {
	ID, KeySalt string
}

type Salts []Salt

func (salts Salts) SaltForKey(keyPass []byte, keyID string) ([]byte, error) {
	idx := slices.IndexFunc(salts, func(salt Salt) bool {
		return salt.ID == keyID
	})

	if idx < 0 {
		return nil, fmt.Errorf("no salt found for key %s", keyID)
	}

	keySalt, err := base64.StdEncoding.DecodeString(salts[idx].KeySalt)
	if err != nil {
		return nil, err
	}

	saltedKeyPass, err := srp.MailboxPassword(keyPass, keySalt)
	if err != nil {
		return nil, nil
	}

	return saltedKeyPass[len(saltedKeyPass)-31:], nil
}
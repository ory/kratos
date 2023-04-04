// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/pkg/errors"
)

func UpgradeWebAuthnCredentials(i *Identity, c *Credentials) (err error) {
	if c.Type != CredentialsTypeWebAuthn {
		return nil
	}

	version := c.Version
	if version == 0 {
		if gjson.GetBytes(c.Config, "user_handle").String() == "" {
			id, err := json.Marshal(i.ID[:])
			if err != nil {
				return errors.WithStack(err)
			}

			c.Config, err = sjson.SetRawBytes(c.Config, "user_handle", id)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		var index = -1
		var err error
		gjson.GetBytes(c.Config, "credentials").ForEach(func(_, value gjson.Result) bool {
			index++

			if value.Get("is_passwordless").Exists() {
				return true
			}

			c.Config, err = sjson.SetBytes(c.Config, fmt.Sprintf("credentials.%d.is_passwordless", index), false)
			return err == nil
		})
		if err != nil {
			return errors.WithStack(err)
		}

		c.Version = 1
	}
	return nil
}

// UpgradeCredentials migrates a set of older WebAuthn credentials to newer ones.
func UpgradeCredentials(i *Identity) error {
	for k := range i.Credentials {
		c := i.Credentials[k]
		if err := UpgradeWebAuthnCredentials(i, &c); err != nil {
			return errors.WithStack(err)
		}
		i.Credentials[k] = c
	}
	return nil
}

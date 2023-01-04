// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package credentialmigrate

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
)

func UpgradeWebAuthnCredentials(i *identity.Identity, c *identity.Credentials) (err error) {
	if c.Type != identity.CredentialsTypeWebAuthn {
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
		gjson.GetBytes(c.Config, "credentials").ForEach(func(key, value gjson.Result) bool {
			index++

			if value.Get("is_passwordless").Exists() {
				return true
			}

			c.Config, err = sjson.SetBytes(c.Config, fmt.Sprintf("credentials.%d.is_passwordless", index), false)
			if err != nil {
				return false
			}

			return true
		})
		if err != nil {
			return errors.WithStack(err)
		}

		c.Version = 1
	}
	return nil
}

// UpgradeCredentials migrates a set of older WebAuthn credentials to newer ones.
func UpgradeCredentials(i *identity.Identity) error {
	for k := range i.Credentials {
		c := i.Credentials[k]
		if err := UpgradeWebAuthnCredentials(i, &c); err != nil {
			return errors.WithStack(err)
		}
		i.Credentials[k] = c
	}
	return nil
}

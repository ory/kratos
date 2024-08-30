// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"encoding/json"
	"fmt"
	"strings"

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
		if err := UpgradeCodeCredentials(&c); err != nil {
			return errors.WithStack(err)
		}
		i.Credentials[k] = c
	}
	return nil
}

func UpgradeCodeCredentials(c *Credentials) (err error) {
	if c.Type != CredentialsTypeCodeAuth {
		return nil
	}

	version := c.Version
	if version == 0 {
		addressType := strings.ToLower(strings.TrimSpace(gjson.GetBytes(c.Config, "address_type").String()))

		channel, err := NewCodeChannel(addressType)
		if err != nil {
			// We know that in some cases the address type can be empty. In this case, we default to email
			// as sms is a new addition to the address_type introduced in this PR.
			channel = CodeChannelEmail
		}

		c.Config, err = sjson.DeleteBytes(c.Config, "used_at")
		if err != nil {
			return errors.WithStack(err)
		}

		c.Config, err = sjson.DeleteBytes(c.Config, "address_type")
		if err != nil {
			return errors.WithStack(err)
		}

		for _, id := range c.Identifiers {
			if id == "" {
				continue
			}

			c.Config, err = sjson.SetBytes(c.Config, "addresses.-1", &CredentialsCodeAddress{
				Address: id,
				Channel: channel,
			})
			if err != nil {
				return errors.WithStack(err)
			}
		}

		// This is needed because sjson adds spaces which can trip string comparisons.
		c.Config, err = json.Marshal(json.RawMessage(c.Config))
		if err != nil {
			return errors.WithStack(err)
		}

		c.Version = 1
	}
	return nil
}

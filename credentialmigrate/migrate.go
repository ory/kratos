package credentialmigrate

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/strategy/webauthn"
)

// UpgradeWebAuthnCredential migrates a webauthn credential from an older version to a newer version.
func UpgradeWebAuthnCredential(i *identity.Identity, ic *identity.Credentials, c *webauthn.CredentialsConfig) {
	if ic.Version == 0 {
		if len(c.UserHandle) == 0 {
			c.UserHandle = i.ID[:]
		}

		// We do not set c.IsPasswordless as it defaults to false anyways, which is the correct migration .

		ic.Version = 1
	}
}

func UpgradeWebAuthnCredentials(i *identity.Identity, c *identity.Credentials) error {
	if c.Type != identity.CredentialsTypeWebAuthn {
		return nil
	}

	var cred webauthn.CredentialsConfig
	if err := json.Unmarshal(c.Config, &cred); err != nil {
		return errors.WithStack(err)
	}

	UpgradeWebAuthnCredential(i, c, &cred)

	updatedConf, err := json.Marshal(&cred)
	if err != nil {
		return errors.WithStack(err)
	}

	c.Config = updatedConf
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

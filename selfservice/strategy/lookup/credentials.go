package lookup

import (
	"time"

	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"

	"github.com/ory/x/sqlxx"
)

// CredentialsConfig is the struct that is being used as part of the identity credentials.
type CredentialsConfig struct {
	// LookupSecrets is a list of recovery codes.
	LookupSecrets []LookupSecret `json:"recovery_codes"`
}

func (c *CredentialsConfig) ToNode() *node.Node {
	messages := make([]text.Message, len(c.LookupSecrets))
	formatted := make([]string, len(c.LookupSecrets))
	for k, code := range c.LookupSecrets {
		if time.Time(code.UsedAt).IsZero() {
			messages[k] = *text.NewInfoSelfServiceSettingsLookupSecret(code.Code)
			formatted[k] = code.Code
		} else {
			messages[k] = *text.NewInfoSelfServiceSettingsLookupSecretUsed(time.Time(code.UsedAt).UTC())
			formatted[k] = "used"
		}
	}

	return node.NewTextField(node.LookupCodes, text.NewInfoSelfServiceSettingsLookupSecretList(formatted, messages), node.LookupGroup).
		WithMetaLabel(text.NewInfoSelfServiceSettingsLookupSecretsLabel())
}

type LookupSecret struct {
	// A recovery code
	Code string `json:"code"`

	// UsedAt indicates whether and when a recovery code was used.
	UsedAt sqlxx.NullTime `json:"used_at,omitempty"`
}

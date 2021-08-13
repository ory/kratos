package lookup

import (
	"time"

	"github.com/ory/x/sqlxx"
)

// CredentialsConfig is the struct that is being used as part of the identity credentials.
type CredentialsConfig struct {
	// List of recovery codes
	RecoveryCodes []RecoveryCode `json:"recovery_codes"`
}

func (c *CredentialsConfig) ToReadableList() []string {
	codes := make([]string, len(c.RecoveryCodes))
	for k, code := range c.RecoveryCodes {
		if time.Time(code.UsedAt).IsZero() {
			codes[k] = code.Code
		} else {
			codes[k] = "already used"
		}
	}
	return codes
}

type RecoveryCode struct {
	// A recovery code
	Code string `json:"code"`

	// UsedAt indicates whether and when a recovery code was used.
	UsedAt sqlxx.NullTime `json:"used_at,omitempty"`
}

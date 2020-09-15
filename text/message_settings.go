package text

import (
	"fmt"
	"time"
)

const (
	InfoSelfServiceSettings ID = 1050000 + iota
	InfoSelfServiceSettingsUpdateSuccess
)

const (
	ErrorValidationSettings ID = 4050000 + iota
	ErrorValidationSettingsFlowExpired
)

func NewErrorValidationSettingsFlowExpired(ago time.Duration) *Message {
	return &Message{
		ID:   ErrorValidationSettingsFlowExpired,
		Text: fmt.Sprintf("The settings flow expired %.2f minutes ago, please try again.", ago.Minutes()),
		Type: Error,
		Context: context(map[string]interface{}{
			"expired_at": time.Now().Add(ago),
		}),
	}
}

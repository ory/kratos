package text

import (
	"fmt"
	"time"
)

const (
	InfoSelfServiceSettings ID = 1050000 + iota
	InfoSelfServiceSettingsUpdateSuccess
	InfoSelfServiceSettingsUpdateLinkOidc
	InfoSelfServiceSettingsUpdateUnlinkOidc
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

func NewInfoSelfServiceSettingsUpdateSuccess() *Message {
	return &Message{
		ID:   InfoSelfServiceSettingsUpdateSuccess,
		Text: "Your changes have been saved!",
		Type: Info,
	}
}

func NewInfoSelfServiceSettingsUpdateLinkOIDC(provider string) *Message {
	return &Message{
		ID:   InfoSelfServiceSettingsUpdateLinkOidc,
		Text: fmt.Sprintf("Link %s", provider),
		Type: Info,
		Context: context(map[string]interface{}{
			"provider": provider,
		}),
	}
}

func NewInfoSelfServiceSettingsUpdateUnlinkOIDC(provider string) *Message {
	return &Message{
		ID:   InfoSelfServiceSettingsUpdateUnlinkOidc,
		Text: fmt.Sprintf("Unlink %s", provider),
		Type: Info,
		Context: context(map[string]interface{}{
			"provider": provider,
		}),
	}
}

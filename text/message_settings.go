package text

import (
	"fmt"
	"strings"
	"time"
)

const (
	InfoSelfServiceSettings ID = 1050000 + iota
	InfoSelfServiceSettingsUpdateSuccess
	InfoSelfServiceSettingsUpdateLinkOidc
	InfoSelfServiceSettingsUpdateUnlinkOidc
	InfoSelfServiceSettingsUpdateUnlinkTOTP
	InfoSelfServiceSettingsTOTPQRCode
	InfoSelfServiceSettingsTOTPSecret
	InfoSelfServiceSettingsRevealLookup
	InfoSelfServiceSettingsRegenerateLookup
	InfoSelfServiceSettingsLookupSecrets
	InfoSelfServiceSettingsLookupConfirm
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

func NewInfoSelfServiceSettingsTOTPQRCode() *Message {
	return &Message{
		ID:   InfoSelfServiceSettingsTOTPQRCode,
		Text: "Authenticator app QR code",
		Type: Info,
	}
}

func NewInfoSelfServiceSettingsTOTPSecret(secret string) *Message {
	return &Message{
		ID:   InfoSelfServiceSettingsTOTPSecret,
		Text: secret,
		Type: Info,
		Context: context(map[string]interface{}{
			"secret": secret,
		}),
	}
}
func NewInfoSelfServiceSettingsTOTPSecretLabel() *Message {
	return &Message{
		ID:   InfoSelfServiceSettingsTOTPSecret,
		Text: "This is your authenticator app secret. Use it if you can not scan the QR code.",
		Type: Info,
	}
}

func NewInfoSelfServiceSettingsUpdateSuccess() *Message {
	return &Message{
		ID:   InfoSelfServiceSettingsUpdateSuccess,
		Text: "Your changes have been saved!",
		Type: Info,
	}
}

func NewInfoSelfServiceSettingsUpdateUnlinkTOTP() *Message {
	return &Message{
		ID:   InfoSelfServiceSettingsUpdateUnlinkTOTP,
		Text: "Unlink TOTP Authenticator App",
		Type: Info,
	}
}

func NewInfoSelfServiceSettingsRevealLookup() *Message {
	return &Message{
		ID:   InfoSelfServiceSettingsRevealLookup,
		Text: "Reveal backup recovery codes",
		Type: Info,
	}
}

func NewInfoSelfServiceSettingsRegenerateLookup() *Message {
	return &Message{
		ID:   InfoSelfServiceSettingsRegenerateLookup,
		Text: "Generate new backup recovery codes",
		Type: Info,
	}
}

func NewInfoSelfServiceSettingsLookupConfirm() *Message {
	return &Message{
		ID:   InfoSelfServiceSettingsLookupConfirm,
		Text: "Confirm backup recovery codes",
		Type: Info,
	}
}

func NewInfoSelfServiceSettingsLookupSecrets(secrets []string) *Message {
	return &Message{
		ID:   InfoSelfServiceSettingsLookupSecrets,
		Text: fmt.Sprintf("%s", strings.Join(secrets, " ")),
		Type: Info,
		Context: context(map[string]interface{}{
			"secrets": secrets,
		}),
	}
}

func NewInfoSelfServiceSettingsLookupSecretsLabel() *Message {
	return &Message{
		ID:   InfoSelfServiceSettingsLookupSecrets,
		Text: "These are your back up recovery codes. Please keep them in a safe place!",
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

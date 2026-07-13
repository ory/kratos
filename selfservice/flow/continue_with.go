// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package flow

import (
	"encoding/base64"
	"net/url"

	"github.com/ory/herodot"

	"github.com/gofrs/uuid"

	"github.com/ory/x/urlx"
)

// swagger:model continueWith
type ContinueWith interface {
	//swagger:ignore
	GetAction() string
}

// swagger:enum ContinueWithActionSetOrySessionToken
type ContinueWithActionSetOrySessionToken string

const (
	ContinueWithActionSetOrySessionTokenString ContinueWithActionSetOrySessionToken = "set_ory_session_token" // #nosec G101 -- only a key constant
)

var _ ContinueWith = new(ContinueWithSetOrySessionToken)

// Indicates that a session was issued, and the application should use this token for authenticated requests
//
// swagger:model continueWithSetOrySessionToken
type ContinueWithSetOrySessionToken struct {
	// Action will always be `set_ory_session_token`
	//
	// required: true
	Action ContinueWithActionSetOrySessionToken `json:"action"`

	// Token is the token of the session
	//
	// required: true
	OrySessionToken string `json:"ory_session_token"`
}

func (ContinueWithSetOrySessionToken) AppendTo(url.Values) url.Values {
	return nil
}

func NewContinueWithSetToken(t string) *ContinueWithSetOrySessionToken {
	return &ContinueWithSetOrySessionToken{
		Action:          ContinueWithActionSetOrySessionTokenString,
		OrySessionToken: t,
	}
}

func (c ContinueWithSetOrySessionToken) GetAction() string {
	return string(c.Action)
}

// swagger:enum ContinueWithActionShowVerificationUI
type ContinueWithActionShowVerificationUI string

// #nosec G101 -- only a key constant
const (
	ContinueWithActionShowVerificationUIString ContinueWithActionShowVerificationUI = "show_verification_ui"
)

var _ ContinueWith = new(ContinueWithVerificationUI)

// Indicates, that the UI flow could be continued by showing a verification ui
//
// swagger:model continueWithVerificationUi
type ContinueWithVerificationUI struct {
	// Action will always be `show_verification_ui`
	//
	// required: true
	Action ContinueWithActionShowVerificationUI `json:"action"`
	// Flow contains the ID of the verification flow
	//
	// required: true
	Flow ContinueWithVerificationUIFlow `json:"flow"`
}

func (c ContinueWithVerificationUI) GetAction() string {
	return string(c.Action)
}

// swagger:model continueWithVerificationUiFlow
type ContinueWithVerificationUIFlow struct {
	// The ID of the verification flow
	//
	// required: true
	ID uuid.UUID `json:"id"`

	// The address that should be verified in this flow
	//
	// required: true
	VerifiableAddress string `json:"verifiable_address"`

	// The URL of the verification flow
	//
	// If this value is set, redirect the user's browser to this URL. This value is typically unset for native clients / API flows.
	//
	// required: false
	URL string `json:"url,omitempty"`
}

func NewContinueWithVerificationUI(id uuid.UUID, address, url string) *ContinueWithVerificationUI {
	return &ContinueWithVerificationUI{
		Action: ContinueWithActionShowVerificationUIString,
		Flow: ContinueWithVerificationUIFlow{
			ID:                id,
			VerifiableAddress: address,
			URL:               url,
		},
	}
}

func (c ContinueWithVerificationUI) AppendTo(src *url.URL) *url.URL {
	values := src.Query()
	values.Set("flow", c.Flow.ID.String())
	return urlx.CopyWithQuery(src, values)
}

type FlowWithContinueWith interface {
	Flow
	AddContinueWith(ContinueWith)
	ContinueWith() []ContinueWith
}

// swagger:enum ContinueWithActionShowSettingsUI
type ContinueWithActionShowSettingsUI string

// #nosec G101 -- only a key constant
const (
	ContinueWithActionShowSettingsUIString ContinueWithActionShowSettingsUI = "show_settings_ui"
)

var _ ContinueWith = new(ContinueWithSettingsUI)

// Indicates, that the UI flow could be continued by showing a settings ui
//
// swagger:model continueWithSettingsUi
type ContinueWithSettingsUI struct {
	// Action will always be `show_settings_ui`
	//
	// required: true
	Action ContinueWithActionShowSettingsUI `json:"action"`

	// Flow contains the ID of the settings flow
	//
	// required: true
	Flow ContinueWithSettingsUIFlow `json:"flow"`
}

func (c ContinueWithSettingsUI) GetAction() string {
	return string(c.Action)
}

// swagger:model continueWithSettingsUiFlow
type ContinueWithSettingsUIFlow struct {
	// The ID of the settings flow
	//
	// required: true
	ID uuid.UUID `json:"id"`

	// The URL of the settings flow
	//
	// If this value is set, redirect the user's browser to this URL. This value is typically unset for native clients / API flows.
	//
	// required: false
	URL string `json:"url,omitempty"`
}

func NewContinueWithSettingsUI(f Flow, redirectTo string) *ContinueWithSettingsUI {
	return &ContinueWithSettingsUI{
		Action: ContinueWithActionShowSettingsUIString,
		Flow: ContinueWithSettingsUIFlow{
			ID:  f.GetID(),
			URL: redirectTo,
		},
	}
}

// swagger:enum ContinueWithActionShowRecoveryUI
type ContinueWithActionShowRecoveryUI string

// #nosec G101 -- only a key constant
const (
	ContinueWithActionShowRecoveryUIString ContinueWithActionShowRecoveryUI = "show_recovery_ui"
)

// Indicates, that the UI flow could be continued by showing a recovery ui
//
// swagger:model continueWithRecoveryUi
type ContinueWithRecoveryUI struct {
	// Action will always be `show_recovery_ui`
	//
	// required: true
	Action ContinueWithActionShowRecoveryUI `json:"action"`
	// Flow contains the ID of the recovery flow
	//
	// required: true
	Flow ContinueWithRecoveryUIFlow `json:"flow"`
}

// swagger:model continueWithRecoveryUiFlow
type ContinueWithRecoveryUIFlow struct {
	// The ID of the recovery flow
	//
	// required: true
	ID uuid.UUID `json:"id"`

	// The URL of the recovery flow
	//
	// If this value is set, redirect the user's browser to this URL. This value is typically unset for native clients / API flows.
	//
	// required: false
	URL string `json:"url,omitempty"`
}

func NewContinueWithRecoveryUI(f Flow) *ContinueWithRecoveryUI {
	return &ContinueWithRecoveryUI{
		Action: ContinueWithActionShowRecoveryUIString,
		Flow: ContinueWithRecoveryUIFlow{
			ID: f.GetID(),
		},
	}
}

func (c ContinueWithRecoveryUI) GetAction() string {
	return string(c.Action)
}

// swagger:enum ContinueWithActionRedirectBrowserTo
type ContinueWithActionRedirectBrowserTo string

// #nosec G101 -- only a key constant
const (
	ContinueWithActionRedirectBrowserToString ContinueWithActionRedirectBrowserTo = "redirect_browser_to"
)

// Indicates, that the UI flow could be continued by showing a recovery ui
//
// swagger:model continueWithRedirectBrowserTo
type ContinueWithRedirectBrowserTo struct {
	// Action will always be `redirect_browser_to`
	//
	// required: true
	Action ContinueWithActionRedirectBrowserTo `json:"action"`

	// The URL to redirect the browser to
	//
	// required: true
	RedirectTo string `json:"redirect_browser_to"`
}

func NewContinueWithRedirectBrowserTo(redirectTo string) *ContinueWithRedirectBrowserTo {
	return &ContinueWithRedirectBrowserTo{
		Action:     ContinueWithActionRedirectBrowserToString,
		RedirectTo: redirectTo,
	}
}

func (c ContinueWithRedirectBrowserTo) GetAction() string {
	return string(c.Action)
}

// swagger:enum ContinueWithActionShowPINEntryUI
type ContinueWithActionShowPINEntryUI string

// #nosec G101 -- only an action constant
const (
	ContinueWithActionShowPINEntryUIString ContinueWithActionShowPINEntryUI = "show_pin_entry_ui"
)

var _ ContinueWith = new(ContinueWithDeviceAuthnPINEntryUI)

// Instructs the client to show its PIN entry UI
//
// Returned after a PIN-protected DeviceAuthn enrollment or secret rotation.
// It carries the one-time HPKE-sealed pin_secret, which the device opens with
// the transport private key it generated for this enrollment and binds to the
// user's PIN. The plaintext secret never leaves the device; the server only
// ever stores its at-rest ciphertext.
//
// The enrolled key's client_key_id is not included: it is the SHA-256 fingerprint
// of the device's own public key, which the device derives locally. Non-PIN keys
// need no client action at all and therefore produce no continue_with.
//
// swagger:model continueWithDeviceAuthnPinEntryUi
type ContinueWithDeviceAuthnPINEntryUI struct {
	// Action will always be `show_pin_entry_ui`
	//
	// required: true
	Action ContinueWithActionShowPINEntryUI `json:"action"`

	// Data carries the sealed pin_secret material.
	//
	// required: true
	Data ContinueWithDeviceAuthnPINEntryUIData `json:"data"`
}

// The one-time HPKE-sealed pin_secret material
//
// Contains the encapsulated key and the sealed ciphertext that the device
// opens with the X25519 transport private key it generated for this
// enrollment or rotation. Open it with HPKE (RFC 9180) using the suite
// DHKEM(X25519, HKDF-SHA256), HKDF-SHA256, AES-128-GCM, the ASCII info
// string "ory/deviceauthn/pin-secret/v1", and the key's client_key_id (its
// ASCII hex form) as the AAD.
//
// swagger:model continueWithDeviceAuthnPinEntryUiData
type ContinueWithDeviceAuthnPINEntryUIData struct {
	// The base64-encoded HPKE encapsulated key (the `enc` output of the seal
	// operation).
	//
	// required: true
	Enc string `json:"enc"`

	// The base64-encoded HPKE ciphertext of the sealed pin_secret.
	//
	// required: true
	Ciphertext string `json:"ciphertext"`
}

func NewContinueWithDeviceAuthnPINEntryUI(enc, ciphertext []byte) *ContinueWithDeviceAuthnPINEntryUI {
	return &ContinueWithDeviceAuthnPINEntryUI{
		Action: ContinueWithActionShowPINEntryUIString,
		Data: ContinueWithDeviceAuthnPINEntryUIData{
			Enc:        base64.StdEncoding.EncodeToString(enc),
			Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
		},
	}
}

func (c ContinueWithDeviceAuthnPINEntryUI) GetAction() string {
	return string(c.Action)
}

func ErrorWithContinueWith(err *herodot.DefaultError, continueWith ...ContinueWith) *herodot.DefaultError {
	if err.DetailsField == nil {
		err.DetailsField = map[string]interface{}{}
	}
	err.DetailsField["continue_with"] = continueWith

	return err
}

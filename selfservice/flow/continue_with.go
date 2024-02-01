// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package flow

import (
	"net/url"

	"github.com/ory/herodot"

	"github.com/gofrs/uuid"

	"github.com/ory/x/urlx"
)

// swagger:model continueWith
type ContinueWith any

var _ ContinueWith = new(ContinueWithSetOrySessionToken)

// Indicates that a session was issued, and the application should use this token for authenticated requests
//
// swagger:model continueWithSetOrySessionToken
type ContinueWithSetOrySessionToken struct {
	// action will always be "set_ory_session_token"
	// required: true
	Action string `json:"action"`

	// Token is the token of the session
	OrySessionToken string `json:"ory_session_token"`
}

func (ContinueWithSetOrySessionToken) AppendTo(url.Values) url.Values {
	return nil
}

func NewContinueWithSetToken(t string) *ContinueWithSetOrySessionToken {
	return &ContinueWithSetOrySessionToken{
		Action:          "set_ory_session_token",
		OrySessionToken: t,
	}
}

var _ ContinueWith = new(ContinueWithVerificationUI)

// Indicates, that the UI flow could be continued by showing a verification ui
//
// swagger:model continueWithVerificationUi
type ContinueWithVerificationUI struct {
	// action will always be "show_verification_ui"
	// required: true
	Action string `json:"action"`
	// Flow contains the ID of the verification flow
	Flow ContinueWithVerificationUIFlow `json:"flow"`
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
	// required: false
	URL string `json:"url,omitempty"`
}

func NewContinueWithVerificationUI(f Flow, address, url string) *ContinueWithVerificationUI {
	return &ContinueWithVerificationUI{
		Action: "show_verification_ui",
		Flow: ContinueWithVerificationUIFlow{
			ID:                f.GetID(),
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

// Indicates, that the UI flow could be continued by showing a settings ui
//
// swagger:model continueWithSettingsUi
type ContinueWithSettingsUI struct {
	// action will always be "show_settings_ui"
	// required: true
	Action string `json:"action"`
	// Flow contains the ID of the settings flow
	Flow ContinueWithSettingsUIFlow `json:"flow"`
}

// swagger:model continueWithSettingsUiFlow
type ContinueWithSettingsUIFlow struct {
	// The ID of the settings flow
	//
	// required: true
	ID uuid.UUID `json:"id"`
}

func NewContinueWithSettingsUI(f Flow) *ContinueWithSettingsUI {
	return &ContinueWithSettingsUI{
		Action: "show_settings_ui",
		Flow: ContinueWithSettingsUIFlow{
			ID: f.GetID(),
		},
	}
}

// Indicates, that the UI flow could be continued by showing a recovery ui
//
// swagger:model continueWithRecoveryUi
type ContinueWithRecoveryUI struct {
	// action will always be "show_recovery_ui"
	// required: true
	Action string `json:"action"`
	// Flow contains the ID of the recovery flow
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
	// required: false
	URL string `json:"url,omitempty"`
}

func NewContinueWithRecoveryUI(f Flow) *ContinueWithRecoveryUI {
	return &ContinueWithRecoveryUI{
		Action: "show_recovery_ui",
		Flow: ContinueWithRecoveryUIFlow{
			ID: f.GetID(),
		},
	}
}

func ErrorWithContinueWith(err *herodot.DefaultError, continueWith ...ContinueWith) *herodot.DefaultError {
	if err.DetailsField == nil {
		err.DetailsField = map[string]interface{}{}
	}
	err.DetailsField["continue_with"] = continueWith

	return err
}

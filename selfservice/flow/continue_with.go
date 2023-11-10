// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package flow

import (
	"net/url"

	"github.com/gofrs/uuid"

	"github.com/ory/x/urlx"
)

// swagger:model continueWith
type ContinueWith any

// swagger:enum ContinueWithActionSetOrySessionToken
type ContinueWithActionSetOrySessionToken string

const (
	ContinueWithActionSetOrySessionTokenString ContinueWithActionSetOrySessionToken = "set_ory_session_token" // #nosec G101 -- only a key constant
)

var _ ContinueWith = new(ContinueWithSetOrySessionToken)

// Indicates that a session was issued, and the application should use this token for authenticated requests
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
		Action: ContinueWithActionShowVerificationUIString,
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

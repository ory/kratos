// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package fedcm

import "encoding/json"

type Provider struct {
	// A full path of the IdP config file.
	ConfigURL string `json:"config_url"`

	// The RP's client identifier, issued by the IdP.
	ClientID string `json:"client_id"`

	// A random string to ensure the response is issued for this specific request.
	// Prevents replay attacks.
	Nonce string `json:"nonce"`

	// By specifying one of login_hints values provided by the accounts endpoints,
	// the FedCM dialog selectively shows the specified account.
	LoginHint string `json:"login_hint,omitempty"`

	// By specifying one of domain_hints values provided by the accounts endpoints,
	// the FedCM dialog selectively shows the specified account.
	DomainHint string `json:"domain_hint,omitempty"`

	// Array of strings that specifies the user information ("name", " email",
	// "picture") that RP needs IdP to share with them.
	//
	// Note: Field API is supported by Chrome 132 and later.
	Fields []string `json:"fields,omitempty"`

	// Custom object that allows to specify additional key-value parameters:
	//  - scope: A string value containing additional permissions that RP needs to
	//    request, for example " drive.readonly calendar.readonly"
	//  - nonce: A random string to ensure the response is issued for this specific
	//    request. Prevents replay attacks.
	//
	// Other custom key-value parameters.
	//
	// Note: parameters is supported from Chrome 132.
	Parameters map[string]string `json:"parameters,omitempty"`
}

// CreateFedcmFlowResponse
//
// Contains a list of all available FedCM providers.
//
// swagger:model createFedcmFlowResponse
type CreateFedcmFlowResponse struct {
	Providers []Provider `json:"providers"`
	CSRFToken string     `json:"csrf_token"`
}

// swagger:route GET /self-service/fed-cm/parameters frontend createFedcmFlow
//
// # Get FedCM Parameters
//
// This endpoint returns a list of all available FedCM providers. It is only supported on the Ory Network.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  200: createFedcmFlowResponse
//	  400: errorGeneric
//	  default: errorGeneric

type UpdateFedcmFlowBody struct {
	// Token contains the result of `navigator.credentials.get`.
	//
	// required: true
	Token string `json:"token"`

	// Nonce is the nonce that was used in the `navigator.credentials.get` call. If
	// specified, it must match the `nonce` claim in the token.
	//
	// required: false
	Nonce string `json:"nonce"`

	// CSRFToken is the anti-CSRF token.
	//
	// required: true
	CSRFToken string `json:"csrf_token"`

	// Transient data to pass along to any webhooks.
	//
	// required: false
	TransientPayload json.RawMessage `json:"transient_payload,omitempty" form:"transient_payload"`
}

// swagger:parameters updateFedcmFlow
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type updateFedcmFlow struct {
	// in: body
	// required: true
	Body UpdateFedcmFlowBody
}

// swagger:route POST /self-service/fed-cm/token frontend updateFedcmFlow
//
// # Submit a FedCM token
//
// Use this endpoint to submit a token from a FedCM provider through
// `navigator.credentials.get` and log the user in. The parameters from
// `navigator.credentials.get` must have come from `GET
// /self-service/fed-cm/parameters`.
//
//	Consumes:
//	- application/json
//	- application/x-www-form-urlencoded
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Header:
//	- Set-Cookie
//
//	Responses:
//	  200: successfulNativeLogin
//	  303: emptyResponse
//	  400: loginFlow
//	  410: errorGeneric
//	  422: errorBrowserLocationChangeRequired
//	  default: errorGeneric

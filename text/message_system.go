// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package text

func NewErrorSystemGeneric(reason string) *Message {
	return &Message{
		ID:   ErrorSystemGeneric,
		Text: reason,
		Type: Error,
		Context: context(map[string]any{
			"reason": reason,
		}),
	}
}

func NewErrorSelfServiceNoMethodsAvailable() *Message {
	return &Message{
		ID:   ErrorSelfServiceNoMethodsAvailable,
		Text: "No authentication methods are available for this request. Please contact the site or app owner.",
		Type: Error,
	}
}

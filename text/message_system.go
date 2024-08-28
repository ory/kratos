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

func NewCaptchaContainerMessage() *Message {
	return &Message{
		ID:   InfoNodeLabelCaptcha,
		Text: "Please complete the captcha challenge to continue.",
		Type: Info,
	}
}

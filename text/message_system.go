// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package text

func NewErrorSystemGeneric(reason string) *Message {
	return &Message{
		ID:      ErrorSystemGeneric,
		Text:    reason,
		Type:    Error,
		Context: context(nil),
	}
}

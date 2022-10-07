// Copyright © 2022 Ory Corp

package text

func NewErrorSystemGeneric(reason string) *Message {
	return &Message{
		ID:      ErrorSystemGeneric,
		Text:    reason,
		Type:    Error,
		Context: context(nil),
	}
}

package text

func NewErrorSystemGeneric(reason string) *Message {
	return &Message{
		ID:       ErrorSystemGeneric,
		Text:     reason,
		Type:     Error,
		Context:  context(nil),
		I18nText: reason,
		I18nData: context(nil),
	}
}

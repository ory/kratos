package text

const (
	InfoNodeLabel              ID = 1070000 + iota // 1070000
	InfoNodeLabelInputPassword                     // 1070001
	InfoNodeLabelGenerated                         // 1070002
	InfoNodeLabelSave                              // 1070003
	InfoNodeLabelID                                // 1070004
	InfoNodeLabelSubmit                            // 1070005
)

func NewInfoNodeInputPassword() *Message {
	return &Message{
		ID:   InfoNodeLabelInputPassword,
		Text: "Password",
		Type: Info,
	}
}

func NewInfoNodeLabelGenerated(title string) *Message {
	return &Message{
		ID:   InfoNodeLabelGenerated,
		Text: title,
		Type: Info,
	}
}

func NewInfoNodeLabelSave() *Message {
	return &Message{
		ID:   InfoNodeLabelSave,
		Text: "Save",
		Type: Info,
	}
}

func NewInfoNodeLabelSubmit() *Message {
	return &Message{
		ID:   InfoNodeLabelSubmit,
		Text: "Submit",
		Type: Info,
	}
}

func NewInfoNodeLabelID() *Message {
	return &Message{
		ID:   InfoNodeLabelID,
		Text: "ID",
		Type: Info,
	}
}

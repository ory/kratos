package x

type FormField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required,omitempty"`
	Value    string `json:"value,omitempty"`
}

package text

import "time"

// swagger:model uiTextType
type Type string

const (
	Info  Type = "info"
	Error Type = "error"
)

var Now = time.Now
var Until = time.Until

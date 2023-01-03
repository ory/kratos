// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package text

import "time"

// swagger:enum UITextType
type UITextType string

// aligned with https://github.com/ory/elements/blob/main/src/theme/message.css.ts
const (
	Info    UITextType = "info"
	Error   UITextType = "error"
	Success UITextType = "success"
)

var Now = time.Now
var Until = time.Until

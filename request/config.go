// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package request

import (
	"net/http"
)

type (
	AuthConfig = struct {
		Type   string         `json:"type" koanf:"type"`
		Config map[string]any `json:"config" koanf:"config"`
	}
	ResponseConfig = struct {
		Parse  bool `json:"parse" koanf:"parse"`
		Ignore bool `json:"ignore" koanf:"ignore"`
	}
	Config struct {
		ID                 string            `json:"id" koanf:"id"`
		Method             string            `json:"method" koanf:"method"`
		URL                string            `json:"url" koanf:"url"`
		TemplateURI        string            `json:"body" koanf:"body"`
		Headers            map[string]string `json:"headers" koanf:"headers"`
		Auth               AuthConfig        `json:"auth" koanf:"auth"`
		EmitAnalyticsEvent *bool             `json:"emit_analytics_event" koanf:"emit_analytics_event"`
		CanInterrupt       bool              `json:"can_interrupt" koanf:"can_interrupt"`
		Response           ResponseConfig    `json:"response" koanf:"response"`

		auth   AuthStrategy
		header http.Header
	}
)

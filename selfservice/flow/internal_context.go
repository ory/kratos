// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package flow

import (
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/ory/kratos/identity"
	"github.com/ory/x/sqlxx"
)

// InternalContextKeyCourierBaseURL is the key under which recovery and
// verification flows stash the base URL captured at flow init. Email
// senders read it back at courier-send time to build the link domain. The
// key is top-level (not strategy-scoped) because the value is flow-wide.
const InternalContextKeyCourierBaseURL = "courier_base_url"

// InternalContextKeyRequestBaseURL is the key under which login,
// registration, and settings flows stash the customer-facing base URL
// captured at flow init. The OIDC/SAML strategy reads it back at
// provider-submit time and embeds it in the encrypted state, so the
// social-sign-in callback can be redirected onto the host the user
// started on (Ory CLI/Tunnel, custom domain) instead of the bare
// projects.oryapis host the IdP redirects to. The key is top-level (not
// strategy-scoped) because the value is flow-wide.
const InternalContextKeyRequestBaseURL = "request_base_url"

func PrefixInternalContextKey(t identity.CredentialsType, suffix string) string {
	return string(t) + "_" + suffix
}

// SetRequestBaseURL stashes the base URL captured at flow init (by a
// proxy-aware middleware, via x.WithBaseURL on the request context) into
// the flow's InternalContext under InternalContextKeyRequestBaseURL. Empty
// input is a no-op (preserves the "no rewrite" path). Inputs longer than
// 8192 bytes are rejected so a pathological header cannot bloat the row —
// the same implicit ceiling the courier base URL uses.
func SetRequestBaseURL(f InternalContexter, baseURL string) error {
	if baseURL == "" || len(baseURL) > 8192 {
		return nil
	}
	f.EnsureInternalContext()
	out, err := sjson.SetBytes(f.GetInternalContext(), InternalContextKeyRequestBaseURL, baseURL)
	if err != nil {
		return errors.WithStack(err)
	}
	f.SetInternalContext(sqlxx.JSONRawMessage(out))
	return nil
}

// GetRequestBaseURL returns the base URL captured at flow init, or the
// empty string when nothing was captured (the OIDC/SAML callback then does
// not rewrite — matching plain projects.oryapis traffic).
func GetRequestBaseURL(f InternalContexter) string {
	return gjson.GetBytes(f.GetInternalContext(), InternalContextKeyRequestBaseURL).String()
}

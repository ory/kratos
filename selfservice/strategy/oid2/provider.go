// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oid2

import (
	"context"

	"github.com/ory/x/urlx"
	"net/url"
	"strings"
)

type Provider interface {
	Config() *Configuration
	GetRedirectUrl(ctx context.Context) string
}

func (providerConfig Configuration) Redir(public *url.URL) string {
	return urlx.AppendPaths(public, strings.Replace(RouteCallback, ":provider", providerConfig.ID, 1)).String()
}

// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
	"github.com/ory/x/logrusx"
)

const maxUpstreamBodyLogBytes = 512

var sensitiveFragmentRE = regexp.MustCompile(`(?i)(access_token|refresh_token|id_token|client_secret|authorization)("?)\s*([:=])\s*("[^"]*"|[^,\s&]+)`)

var (
	ErrScopeMissing = herodot.ErrBadRequest.
			WithError("authentication failed because a required scope was not granted").
			WithReasonf(`Unable to finish because one or more permissions were not granted. Please retry and accept all permissions.`)

	ErrIDTokenMissing = herodot.ErrBadRequest.
				WithError("authentication failed because id_token is missing").
				WithReasonf(`Authentication failed because no id_token was returned. Please accept the "openid" permission and try again.`)
)

func logUpstreamError(l *logrusx.Logger, resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil {
		l = l.WithError(err)
	}

	l.WithField("response_code", resp.StatusCode).WithField("response_body", string(body)).Error("The upstream OIDC provider returned a non 200 status code.")
	return errors.WithStack(herodot.ErrUpstreamError().WithReasonf("OpenID Connect provider returned a %d status code but 200 is expected.", resp.StatusCode))
}

func extractOAuth2RetrieveError(err error) (statusCode int, safeBodyFragment string, ok bool) {
	var retrieveErr *oauth2.RetrieveError
	if !errors.As(err, &retrieveErr) {
		return 0, "", false
	}

	safeBodyFragment = safeBodyForLog(retrieveErr.Body, maxUpstreamBodyLogBytes)
	if retrieveErr.Response != nil {
		return retrieveErr.Response.StatusCode, safeBodyFragment, true
	}

	return 0, safeBodyFragment, true
}

func safeBodyForLog(body []byte, limit int) string {
	if len(body) == 0 {
		return ""
	}

	if limit <= 0 {
		limit = maxUpstreamBodyLogBytes
	}
	if len(body) > limit {
		body = body[:limit]
	}

	safe := strings.ReplaceAll(string(body), "\n", " ")
	safe = strings.ReplaceAll(safe, "\r", " ")
	safe = sensitiveFragmentRE.ReplaceAllString(safe, `$1$2$3***`)
	return strings.TrimSpace(safe)
}

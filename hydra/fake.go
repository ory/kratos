// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hydra

import (
	"context"
	"errors"

	"github.com/ory/herodot"
	hydraclientgo "github.com/ory/hydra-client-go/v2"
)

const (
	FakeInvalidLoginChallenge = "2e98454e-031b-4870-9ad6-8517df1ce604"
	FakeValidLoginChallenge   = "5ff59a39-ecc5-467e-bb10-26644c0700ee"
	FakePostLoginURL          = "https://www.ory.sh/fake-post-login"
)

var ErrFakeAcceptLoginRequestFailed = errors.New("failed to accept login request")

type FakeHydra struct {
	Skip       bool
	RequestURL string
}

var _ Hydra = &FakeHydra{}

func NewFake() *FakeHydra {
	return &FakeHydra{
		RequestURL: "https://www.ory.sh",
	}
}

func (h *FakeHydra) AcceptLoginRequest(_ context.Context, params AcceptLoginRequestParams) (string, error) {
	if params.SessionID == "" {
		return "", errors.New("session id must not be empty")
	}
	switch params.LoginChallenge {
	case FakeInvalidLoginChallenge:
		return "", ErrFakeAcceptLoginRequestFailed
	case FakeValidLoginChallenge:
		return FakePostLoginURL, nil
	default:
		panic("unknown fake login_challenge " + params.LoginChallenge)
	}
}

func (h *FakeHydra) GetLoginRequest(_ context.Context, loginChallenge string) (*hydraclientgo.OAuth2LoginRequest, error) {
	switch loginChallenge {
	case FakeInvalidLoginChallenge:
		return nil, herodot.ErrBadRequest.WithReasonf("Unable to get OAuth 2.0 Login Challenge.")
	case FakeValidLoginChallenge:
		return &hydraclientgo.OAuth2LoginRequest{
			RequestUrl: h.RequestURL,
			Skip:       h.Skip,
		}, nil
	default:
		panic("unknown fake login_challenge " + loginChallenge)
	}
}

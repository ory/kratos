// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hydra

import (
	"context"
	"errors"

	hydraclientgo "github.com/ory/hydra-client-go/v2"
	"github.com/ory/kratos/session"
)

const (
	FakeInvalidLoginChallenge = "2e98454e-031b-4870-9ad6-8517df1ce604"
	FakeValidLoginChallenge   = "5ff59a39-ecc5-467e-bb10-26644c0700ee"
	FakePostLoginURL          = "https://www.ory.sh/fake-post-login"
)

var ErrFakeAcceptLoginRequestFailed = errors.New("failed to accept login request")

type FakeHydra struct{}

var _ Hydra = &FakeHydra{}

func NewFake() *FakeHydra {
	return &FakeHydra{}
}

func (h *FakeHydra) AcceptLoginRequest(_ context.Context, loginChallenge string, _ string, _ session.AuthenticationMethods) (string, error) {
	switch loginChallenge {
	case FakeInvalidLoginChallenge:
		return "", ErrFakeAcceptLoginRequestFailed
	case FakeValidLoginChallenge:
		return FakePostLoginURL, nil
	default:
		panic("unknown fake login_challenge " + loginChallenge)
	}
}

func (h *FakeHydra) GetLoginRequest(_ context.Context, loginChallenge string) (*hydraclientgo.OAuth2LoginRequest, error) {
	switch loginChallenge {
	case FakeInvalidLoginChallenge:
		return &hydraclientgo.OAuth2LoginRequest{}, nil
	case FakeValidLoginChallenge:
		return &hydraclientgo.OAuth2LoginRequest{
			RequestUrl: "https://www.ory.sh",
		}, nil
	default:
		panic("unknown fake login_challenge " + loginChallenge)
	}
}

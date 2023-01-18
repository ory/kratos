// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hydra

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"

	hydraclientgo "github.com/ory/hydra-client-go"
	"github.com/ory/kratos/session"
)

const (
	FAKE_GET_LOGIN_REQUEST_RETURN_NIL_NIL = "b805f2d9-2f6d-4745-9d68-a17f48e25774"
	FAKE_ACCEPT_REQUEST_FAIL              = "2e98454e-031b-4870-9ad6-8517df1ce604"
	FAKE_SUCCESS                          = "5ff59a39-ecc5-467e-bb10-26644c0700ee"
)

type FakeHydra struct{}

var _ Hydra = &FakeHydra{}

func NewFakeHydra() *FakeHydra {
	return &FakeHydra{}
}

func (h *FakeHydra) AcceptLoginRequest(ctx context.Context, hlc uuid.UUID, sub string, amr session.AuthenticationMethods) (string, error) {
	switch hlc.String() {
	case FAKE_ACCEPT_REQUEST_FAIL:
		return "", errors.New("failed to accept login request")
	default:
		panic("unknown fake login_challenge " + hlc.String())
	}
}

func (h *FakeHydra) GetLoginRequest(ctx context.Context, hlc uuid.NullUUID) (*hydraclientgo.OAuth2LoginRequest, error) {
	switch hlc.UUID.String() {
	case FAKE_ACCEPT_REQUEST_FAIL:
		return &hydraclientgo.OAuth2LoginRequest{}, nil
	case FAKE_SUCCESS:
		return &hydraclientgo.OAuth2LoginRequest{}, nil
	default:
		panic("unknown fake login_challenge " + hlc.UUID.String())
	}
}

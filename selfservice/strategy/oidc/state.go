// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"bytes"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/x"
)

type State struct {
	FlowID string
	Data   []byte
}

func (s *State) String() string {
	return base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", s.FlowID, s.Data)))
}

func generateState(flowID string) *State {
	return &State{
		FlowID: flowID,
		Data:   x.NewUUID().Bytes(),
	}
}

func (s *State) setCode(code string) {
	s.Data = sha512.New().Sum([]byte(code))
}

func (s *State) codeMatches(code string) bool {
	return subtle.ConstantTimeCompare(s.Data, sha512.New().Sum([]byte(code))) == 1
}

func parseState(s string) (*State, error) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if id, data, ok := bytes.Cut(raw, []byte(":")); !ok {
		return nil, errors.New("state has invalid format")
	} else {
		return &State{FlowID: string(id), Data: data}, nil
	}
}

func storeProvider(f flow.InternalContexter, provider string) error {
	f.EnsureInternalContext()
	bytes, err := sjson.SetBytes(
		f.GetInternalContext(),
		"oidc_provider",
		provider,
	)
	if err != nil {
		return errors.Wrap(err, "failed to store OIDC provider to internal context")
	}
	f.SetInternalContext(bytes)
	return nil
}

func providerFromFlow(f flow.InternalContexter) (provider string) {
	if f.GetInternalContext() == nil {
		return ""
	}
	raw := gjson.GetBytes(f.GetInternalContext(), "oidc_provider")
	if !raw.Exists() {
		return ""
	}
	return raw.String()
}

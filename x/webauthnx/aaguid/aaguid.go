// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

//go:generate curl https://raw.githubusercontent.com/passkeydeveloper/passkey-authenticator-aaguids/main/aaguid.json --output passkey-aaguids.json

package aaguid

import (
	_ "embed"
	"encoding/json"
	"maps"

	"github.com/gofrs/uuid"
)

var (
	//go:embed aaguids.json
	rawAAGUIDs []byte
	//go:embed passkey-aaguids.json
	rawPasskeyAAGUIDs []byte
	aaguids           map[string]AAGUID
)

type AAGUID struct {
	Name      string `json:"name"`
	IconDark  string `json:"icon_dark"`
	IconLight string `json:"icon_light"`
}

func init() {
	err := json.Unmarshal(rawAAGUIDs, &aaguids)
	if err != nil {
		panic(err)
	}

	var passkeyAAGUIDs map[string]AAGUID
	err = json.Unmarshal(rawPasskeyAAGUIDs, &passkeyAAGUIDs)
	if err != nil {
		panic(err)
	}

	maps.Copy(aaguids, passkeyAAGUIDs)
}

func Lookup(id []byte) *AAGUID {
	uid, err := uuid.FromBytes(id)
	if err != nil {
		return nil
	}

	if aaguid, ok := aaguids[uid.String()]; ok {
		return &aaguid
	}
	return nil
}

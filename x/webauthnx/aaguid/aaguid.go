package aaguid

import (
	_ "embed"
	"encoding/json"

	"github.com/gofrs/uuid"
)

var (
	//go:embed combined_aaguid.json
	rawAAGUIDs []byte
	aaguids    map[string]AAGUID
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

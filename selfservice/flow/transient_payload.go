package flow

import (
	"github.com/ory/x/sqlxx"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const internalContextTransientPayloadPath = "transient_payload"

func SetTransientPayloadIntoInternalContext(flow InternalContexter, transientPayload sqlxx.JSONRawMessage) error {
	if flow.GetInternalContext() == nil {
		flow.EnsureInternalContext()
	}
	bytes, err := sjson.SetBytes(
		flow.GetInternalContext(),
		internalContextTransientPayloadPath,
		transientPayload,
	)
	if err != nil {
		return err
	}
	flow.SetInternalContext(bytes)

	return nil
}

func GetTransientPayloadFromInternalContext(flow InternalContexter) (sqlxx.JSONRawMessage, error) {
	if flow.GetInternalContext() == nil {
		flow.EnsureInternalContext()
	}
	raw := gjson.GetBytes(flow.GetInternalContext(), internalContextTransientPayloadPath)
	if !raw.IsObject() {
		return nil, nil
	}

	return sqlxx.JSONRawMessage(raw.Raw), nil
}

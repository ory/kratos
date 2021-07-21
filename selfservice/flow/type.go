package flow

// Type is the flow type.
//
// The flow type can either be `api` or `browser`.
//
// swagger:model selfServiceFlowType
type Type string

const (
	TypeAPI     Type = "api"
	TypeBrowser Type = "browser"
)

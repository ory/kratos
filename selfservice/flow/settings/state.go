package settings

// State represents the state of this flow. It knows two states:
//
// - show_form: No user data has been collected, or it is invalid, and thus the form should be shown.
// - success: Indicates that the settings flow has been updated successfully with the provided data.
//	   Done will stay true when repeatedly checking. If set to true, done will revert back to false only
//	   when a flow with invalid (e.g. "please use a valid phone number") data was sent.
//
// swagger:model selfServiceSettingsFlowState
type State string

const (
	StateShowForm State = "show_form"
	StateSuccess  State = "success"
)

package pin

type submitSelfServiceLoginFlowWithPinMethodBody struct {
	// Method should be set to "pin" when logging in using the pin strategy.
	//
	// required: true
	Method string `json:"method"`

	// Sending the anti-csrf token is only required for browser login flows.
	CSRFToken string `json:"csrf_token"`

	// The user's pin.
	//
	// required: true
	Pin string `json:"pin"`
}

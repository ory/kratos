package text

// This file contains error IDs for all system errors / JSON errors

const (
	ErrIDNeedsPrivilegedSession                        = "session_refresh_required"
	ErrIDSelfServiceFlowExpired                        = "self_service_flow_expired"
	ErrIDSelfServiceFlowDisabled                       = "self_service_flow_disabled"
	ErrIDSelfServiceBrowserLocationChangeRequiredError = "browser_location_change_required"

	ErrIDAlreadyLoggedIn             = "session_already_available"
	ErrIDAddressNotVerified          = "session_verified_address_required"
	ErrIDSessionHasAALAlready        = "session_aal_already_fulfilled"
	ErrIDSessionRequiredForHigherAAL = "session_aal1_required"
	ErrIDHigherAALRequired           = "session_aal2_required"
	ErrNoActiveSession               = "session_inactive"
	ErrIDRedirectURLNotAllowed       = "self_service_flow_return_to_forbidden"
	ErrIDInitiatedBySomeoneElse      = "security_identity_mismatch"

	ErrIDCSRF = "security_csrf_violation"
)

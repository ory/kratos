package text

// This file contains error IDs for all system errors / JSON errors

const (
	ErrIDNeedsPrivilegedSession                        = "needs_privileged_session"
	ErrIDSelfServiceFlowExpired                        = "self_service_flow_expired"
	ErrIDSelfServiceBrowserLocationChangeRequiredError = "browser_location_change_required"

	ErrIDAlreadyLoggedIn             = "has_session_already"
	ErrIDAddressNotVerified          = "no_verified_address"
	ErrIDSessionHasAALAlready        = "session_fulfills_aal"
	ErrIDSessionRequiredForHigherAAL = "aal_needs_session"
	ErrIDHigherAALRequired           = "aal_needs_upgrade"
	ErrNoActiveSession               = "no_active_session"
	ErrIDRedirectURLNotAllowed       = "forbidden_return_to"
	ErrIDInitiatedBySomeoneElse      = "intended_for_someone_else"

	ErrIDCSRF = "csrf_violation"
)

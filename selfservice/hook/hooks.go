// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

const (
	KeySessionIssuer                   = "session"
	KeySessionDestroyer                = "revoke_active_sessions"
	KeySessionImpossibleTravelDetector = "flag_sessions_with_impossible_travel_distance"
	KeyWebHook                         = "web_hook"
	KeyRequireVerifiedAddress          = "require_verified_address"
	KeyVerificationUI                  = "show_verification_ui"
	KeyVerifier                        = "verification"
	KeyVerifyNewAddress                = "verify_new_address"
	KeyNotifyPreviousAddresses         = "notify_previous_addresses"
)

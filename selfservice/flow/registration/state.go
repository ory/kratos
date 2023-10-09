// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration

import "github.com/ory/kratos/selfservice/flow"

// State represents the state of this request:
//
// - choose_method: ask the user to choose a method (e.g. registration with email)
// - sent_email: the email has been sent to the user
// - passed_challenge: the request was successful and the registration challenge was passed.
//
// swagger:model registrationFlowState
type State = flow.State

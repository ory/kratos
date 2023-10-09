// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import "github.com/ory/kratos/selfservice/flow"

// Login Flow State
//
// The state represents the state of the login flow.
//
// - choose_method: ask the user to choose a method (e.g. login account via email)
// - sent_email: the email has been sent to the user
// - passed_challenge: the request was successful and the login challenge was passed.
//
// swagger:model loginFlowState
type State = flow.State

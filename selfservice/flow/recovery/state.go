// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package recovery

import "github.com/ory/kratos/selfservice/flow"

// Recovery Flow State
//
// The state represents the state of the recovery flow.
//
// - choose_method: ask the user to choose a method (e.g. recover account via email)
// - sent_email: the email has been sent to the user
// - passed_challenge: the request was successful and the recovery challenge was passed.
//
// swagger:model recoveryFlowState
type State = flow.State

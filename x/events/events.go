// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"github.com/ory/x/otelx/semconv"
)

const (
	SignupSuccessful       semconv.Event = semconv.EventIdentityCreated
	SignupFailed           semconv.Event = "SignupFailed"
	LoginSuccessful        semconv.Event = semconv.EventSessionIssued
	LoginFailed            semconv.Event = "LoginFailed"
	MFAPrompted            semconv.Event = "MFAPrompted"
	MFASuccessful          semconv.Event = "MFASuccessful"
	MFAFailed              semconv.Event = "MFAFailed"
	RecoveryStarted        semconv.Event = "RecoveryStarted"
	RecoverySuccessful     semconv.Event = "RecoverySuccessful"
	VerificationStarted    semconv.Event = "VerificationStarted"
	VerificationSuccessful semconv.Event = "VerificationSuccessful"
	ProfileUpdated         semconv.Event = "ProfileUpdated"
	ProfileUpdateFailed    semconv.Event = "ProfileUpdateFailed"
	CredentialsAdded       semconv.Event = "CredentialsAdded"
	CredentialsRemoved     semconv.Event = "CredentialsRemoved"
	AccountLinked          semconv.Event = "AccountLinked"
	SignOut                semconv.Event = "SignOut"
)

const (
	AdminIdentityCreated        = "AdminIdentityCreated"
	AdminIdentitiesListed       = "AdminIdentitiesListed"
	AdminIdentityViewed         = "AdminIdentityViewed"
	AdminIdentityUpdated        = "AdminIdentityUpdated"
	AdminAccountRecoveryStarted = "AdminAccountRecoveryStarted"
	AdminSessionTerminated      = "AdminSessionTerminated"
)

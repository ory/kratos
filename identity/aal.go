package identity

func DetermineAAL(cts []CredentialsType) AuthenticatorAssuranceLevel {
	aal := NoAuthenticatorAssuranceLevel

	var firstFactor bool
	var secondFactor bool
	for _, a := range cts {
		switch a {
		case CredentialsTypeRecoveryLink:
			fallthrough
		case CredentialsTypeOIDC:
			fallthrough
		case CredentialsTypePassword:
			firstFactor = true
		case CredentialsTypeTOTP:
			secondFactor = true
		case CredentialsTypeLookup:
			secondFactor = true
		case CredentialsTypeWebAuthn:
			secondFactor = true
		}
	}

	if firstFactor && secondFactor {
		aal = AuthenticatorAssuranceLevel2
	} else if firstFactor {
		aal = AuthenticatorAssuranceLevel1
	}

	// Using only the second factor is not enough for any type of assurance.
	return aal
}

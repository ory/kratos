package identity

func DetermineAAL(cts []CredentialsType) AuthenticatorAssuranceLevel {
	aal := NoAuthenticatorAssuranceLevel

	var firstFactor bool
	var secondFactor bool
	var foundWebAuthn bool
	for _, a := range cts {
		switch a {
		case CredentialsTypeRecoveryLink:
			fallthrough
		case CredentialsTypeOIDC:
			fallthrough
		case "v0.6_legacy_session":
			fallthrough
		case CredentialsTypePassword:
			firstFactor = true
		case CredentialsTypeTOTP:
			secondFactor = true
		case CredentialsTypeLookup:
			secondFactor = true
		case CredentialsTypeWebAuthn:
			secondFactor = true
			foundWebAuthn = true
		}
	}

	if firstFactor && secondFactor {
		aal = AuthenticatorAssuranceLevel2
	} else if firstFactor {
		aal = AuthenticatorAssuranceLevel1
	} else if foundWebAuthn {
		// If none of the above match but WebAuthn is set, we have AAL1
		aal = AuthenticatorAssuranceLevel1
	}

	// Using only the second factor is not enough for any type of assurance.
	return aal
}

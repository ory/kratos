package identity

import "github.com/tidwall/gjson"

func IsCredentialAAL1(credential Credentials, webAuthnIsPasswordless bool) bool {
	switch credential.Type {
	case CredentialsTypeRecoveryLink:
		fallthrough
	case CredentialsTypeOIDC:
		fallthrough
	case "v0.6_legacy_session":
		fallthrough
	case CredentialsTypePassword:
		return true
	case CredentialsTypeWebAuthn:
		for _, c := range gjson.GetBytes(credential.Config, "credentials").Array() {
			if c.Get("is_passwordless").Bool() {
				return webAuthnIsPasswordless
			}
		}
		return false
	}
	return false
}

func IsCredentialAAL2(credential Credentials, webAuthnIsPasswordless bool) bool {
	switch credential.Type {
	case CredentialsTypeTOTP:
		return true
	case CredentialsTypeLookup:
		return true
	case CredentialsTypeWebAuthn:
		creds := gjson.GetBytes(credential.Config, "credentials").Array()
		if len(creds) == 0 {
			// Legacy credential before passwordless -> AAL2
			return true
		}
		for _, c := range gjson.GetBytes(credential.Config, "credentials").Array() {
			if !c.Get("is_passwordless").Bool() {
				return !webAuthnIsPasswordless
			}
		}
	}
	return false
}

func MaximumAAL(c map[CredentialsType]Credentials, conf interface {
	WebAuthnForPasswordless() bool
}) AuthenticatorAssuranceLevel {
	aal := NoAuthenticatorAssuranceLevel

	var firstFactor bool
	var secondFactor bool
	for _, a := range c {
		if IsCredentialAAL1(a, conf.WebAuthnForPasswordless()) {
			firstFactor = true
		} else if IsCredentialAAL2(a, conf.WebAuthnForPasswordless()) {
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

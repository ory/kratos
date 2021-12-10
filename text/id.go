package text

// This file MUST not have any imports to modules that are not in the standard library.
// Otherwise, `make docs/cli` will fail.

type ID int

const (
	InfoSelfServiceLoginRoot      ID = 1010000 + iota // 1010000
	InfoSelfServiceLogin                              // 1010001
	InfoSelfServiceLoginWith                          // 1010002
	InfoSelfServiceLoginReAuth                        // 1010003
	InfoSelfServiceLoginMFA                           // 1010004
	InfoSelfServiceLoginVerify                        // 1010005
	InfoSelfServiceLoginTOTPLabel                     // 1010006
	InfoLoginLookupLabel                              // 1010007
	InfoSelfServiceLoginWebAuthn                      // 1010008
	InfoLoginTOTP                                     // 1010009
	InfoLoginLookup                                   // 1010010
)

const (
	InfoSelfServiceLogout ID = 1020000 + iota
)

const (
	InfoSelfServiceMFA ID = 1030000 + iota
)

const (
	InfoSelfServiceRegistrationRoot ID = 1040000 + iota // 1040000
	InfoSelfServiceRegistration                         // 1040001
	InfoSelfServiceRegistrationWith                     // 1040002
	InfoRegistrationContinue                            // 1040003
)

const (
	InfoSelfServiceSettings ID = 1050000 + iota
	InfoSelfServiceSettingsUpdateSuccess
	InfoSelfServiceSettingsUpdateLinkOidc
	InfoSelfServiceSettingsUpdateUnlinkOidc
	InfoSelfServiceSettingsUpdateUnlinkTOTP
	InfoSelfServiceSettingsTOTPQRCode
	InfoSelfServiceSettingsTOTPSecret
	InfoSelfServiceSettingsRevealLookup
	InfoSelfServiceSettingsRegenerateLookup
	InfoSelfServiceSettingsLookupSecret
	InfoSelfServiceSettingsLookupSecretLabel
	InfoSelfServiceSettingsLookupConfirm
	InfoSelfServiceSettingsRegisterWebAuthn
	InfoSelfServiceSettingsRegisterWebAuthnDisplayName
	InfoSelfServiceSettingsLookupSecretUsed
	InfoSelfServiceSettingsLookupSecretList
	InfoSelfServiceSettingsDisableLookup
	InfoSelfServiceSettingsTOTPSecretLabel
)

const (
	InfoSelfServiceRecovery           ID = 1060000 + iota // 1060000
	InfoSelfServiceRecoverySuccessful                     // 1060001
	InfoSelfServiceRecoveryEmailSent                      // 1060002
)

const (
	InfoNodeLabel              ID = 1070000 + iota // 1070000
	InfoNodeLabelInputPassword                     // 1070001
	InfoNodeLabelGenerated                         // 1070002
	InfoNodeLabelSave                              // 1070003
	InfoNodeLabelID                                // 1070004
	InfoNodeLabelSubmit                            // 1070005
	InfoNodeLabelVerifyOTP                         // 1070006
	InfoNodeLabelEmail                             // 1070007
)

const (
	InfoSelfServiceVerification           ID = 1080000 + iota // 1080000
	InfoSelfServiceVerificationEmailSent                      // 1080001
	InfoSelfServiceVerificationSuccessful                     // 1080002
)

const (
	ErrorValidation ID = 4000000 + iota
	ErrorValidationGeneric
	ErrorValidationRequired
	ErrorValidationMinLength
	ErrorValidationInvalidFormat
	ErrorValidationPasswordPolicyViolation
	ErrorValidationInvalidCredentials
	ErrorValidationDuplicateCredentials
	ErrorValidationTOTPVerifierWrong
	ErrorValidationIdentifierMissing
	ErrorValidationAddressNotVerified
	ErrorValidationNoTOTPDevice
	ErrorValidationLookupAlreadyUsed
	ErrorValidationNoWebAuthnDevice
	ErrorValidationNoLookup
)

const (
	ErrorValidationLogin                       ID = 4010000 + iota // 4010000
	ErrorValidationLoginFlowExpired                                // 4010001
	ErrorValidationLoginNoStrategyFound                            // 4010002
	ErrorValidationRegistrationNoStrategyFound                     // 4010003
	ErrorValidationSettingsNoStrategyFound                         // 4010004
	ErrorValidationRecoveryNoStrategyFound                         // 4010005
	ErrorValidationVerificationNoStrategyFound                     // 4010006
)

const (
	ErrorValidationRegistration ID = 4040000 + iota
	ErrorValidationRegistrationFlowExpired
)

const (
	ErrorValidationSettings ID = 4050000 + iota
	ErrorValidationSettingsFlowExpired
)

const (
	ErrorValidationRecovery                          ID = 4060000 + iota // 4060000
	ErrorValidationRecoveryRetrySuccess                                  // 4060001
	ErrorValidationRecoveryStateFailure                                  // 4060002
	ErrorValidationRecoveryMissingRecoveryToken                          // 4060003
	ErrorValidationRecoveryTokenInvalidOrAlreadyUsed                     // 4060004
	ErrorValidationRecoveryFlowExpired                                   // 4060005
)

const (
	ErrorValidationVerification                          ID = 4070000 + iota // 4070000
	ErrorValidationVerificationTokenInvalidOrAlreadyUsed                     // 4070001
	ErrorValidationVerificationRetrySuccess                                  // 4070002
	ErrorValidationVerificationStateFailure                                  // 4070003
	ErrorValidationVerificationMissingVerificationToken                      // 4070004
	ErrorValidationVerificationFlowExpired                                   // 4070005
)

const (
	ErrorSystem ID = 5000000 + iota
	ErrorSystemGeneric
)

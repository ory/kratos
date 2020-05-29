package text

type ID int

const (
	InfoSelfServiceLogin ID = 1010000 + iota
)

const (
	InfoSelfServiceLogout ID = 1020000 + iota
)

const (
	InfoSelfServiceMFA ID = 1030000 + iota
)

const (
	InfoSelfServiceRegistration ID = 1040000 + iota
)

const (
	InfoSelfServiceSettings ID = 1050000 + iota
	InfoSelfServiceSettingsUpdateSuccess
)

const (
	InfoSelfServiceRecovery ID = 1060000 + iota
	InfoSelfServiceRecoverySuccessful
	InfoSelfServiceRecoveryEmailSent
)

const (
	InfoSelfServiceVerification ID = 1070000 + iota
)

const (
	ErrorValidation ID = 4000000 + iota
	ErrorValidationRequired
)

const (
	ErrorValidationRecovery ID = 4060000 + iota
	ErrorValidationRecoveryRetrySuccess
	ErrorValidationRecoveryStateFailure
	ErrorValidationRecoveryMissingRecoveryToken
	ErrorValidationRecoveryRecoveryTokenInvalidOrAlreadyUsed
)

const (
	ErrorSystem ID = 5000000 + iota
)

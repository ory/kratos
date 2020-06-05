package text

type ID int

const (
	InfoSelfServiceLogout ID = 1020000 + iota
)

const (
	InfoSelfServiceMFA ID = 1030000 + iota
)

const (
	InfoSelfServiceSettings ID = 1050000 + iota
	InfoSelfServiceSettingsUpdateSuccess
)

const (
	ErrorSystem ID = 5000000 + iota
	ErrorSystemGeneric
)

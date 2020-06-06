package form

type (
	richError interface {
		StatusCode() int
		Reason() string
	}
)

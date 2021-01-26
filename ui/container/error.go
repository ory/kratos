package container

type (
	richError interface {
		StatusCode() int
		Reason() string
	}
)

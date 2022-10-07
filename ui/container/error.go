// Copyright © 2022 Ory Corp

package container

type (
	richError interface {
		StatusCode() int
		Reason() string
	}
)

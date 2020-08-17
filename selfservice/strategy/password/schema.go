package password

import (
	"github.com/gobuffalo/packr/v2"
)

var schemas = packr.New(".schema", ".schema")
var loginSchema, registrationSchema []byte

func init() {
	var err error
	loginSchema, err = schemas.Find("login.schema.json")
	if err != nil {
		panic(err)
	}

	registrationSchema, err = schemas.Find("registration.schema.json")
	if err != nil {
		panic(err)
	}
}

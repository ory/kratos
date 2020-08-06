package password

import (
	"github.com/markbates/pkger"

	"github.com/ory/kratos/x/pkgerx"
)

var loginSchema = pkgerx.MustRead(pkger.Open("/selfservice/strategy/password/.schema/login.schema.json"))
var registrationSchema = pkgerx.MustRead(pkger.Open("/selfservice/strategy/password/.schema/registration.schema.json"))

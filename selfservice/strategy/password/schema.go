package password

import (
	"github.com/markbates/pkger"

	"github.com/ory/kratos/x"
)

var _ = pkger.Dir("/selfservice/strategy/password/.schema")

var loginSchema, registrationSchema, settingsSchema []byte

func init() {
	loginSchema = x.MustPkgerRead(pkger.Open("/selfservice/strategy/password/.schema/login.schema.json"))
	registrationSchema = x.MustPkgerRead(pkger.Open("/selfservice/strategy/password/.schema/registration.schema.json"))
	settingsSchema = x.MustPkgerRead(pkger.Open("/selfservice/strategy/password/.schema/settings.schema.json"))
}

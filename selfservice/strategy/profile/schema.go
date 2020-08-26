package profile

import (
	"github.com/markbates/pkger"

	"github.com/ory/kratos/x"
)

var _ = pkger.Dir("/selfservice/strategy/profile/.schema")

var  settingsSchema []byte

func init() {
	settingsSchema = x.MustPkgerRead(pkger.Open("/selfservice/strategy/password/.schema/settings.schema.json"))
}

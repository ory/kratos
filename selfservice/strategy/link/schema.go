package link

import (
	"github.com/markbates/pkger"

	"github.com/ory/kratos/x"
)

var _ = pkger.Dir("/selfservice/strategy/link/.schema")

var emailSchema []byte

func init() {
	emailSchema = x.MustPkgerRead(pkger.Open("/selfservice/strategy/link/.schema/email.schema.json"))
}

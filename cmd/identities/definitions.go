package identities

import (
	"strings"

	"github.com/ory/kratos-client-go"

	"github.com/ory/x/cmdx"
)

type (
	outputIdentity           kratos.Identity
	outputIdentityCollection struct {
		identities []kratos.Identity
	}
)

func (_ *outputIdentity) Header() []string {
	return []string{"ID", "VERIFIED ADDRESSES", "RECOVERY ADDRESSES", "SCHEMA ID", "SCHEMA URL"}
}

func (i *outputIdentity) Columns() []string {
	data := [5]string{
		i.Id,
		cmdx.None,
		cmdx.None,
		cmdx.None,
		cmdx.None,
	}

	addresses := make([]string, 0, len(i.VerifiableAddresses))
	for _, a := range i.VerifiableAddresses {
		if len(a.Value) > 0 {
			addresses = append(addresses, a.Value)
		}
	}
	data[1] = strings.Join(addresses, ", ")

	addresses = addresses[:0]
	for _, a := range i.RecoveryAddresses {
		if len(a.Value) > 0 {
			addresses = append(addresses, a.Value)
		}
	}
	data[2] = strings.Join(addresses, ", ")
	data[3] = i.SchemaId
	data[4] = i.SchemaUrl

	return data[:]
}

func (i *outputIdentity) Interface() interface{} {
	return i
}

func (_ *outputIdentityCollection) Header() []string {
	return []string{"ID", "VERIFIED ADDRESS 1", "RECOVERY ADDRESS 1", "SCHEMA ID", "SCHEMA URL"}
}

func (c *outputIdentityCollection) Table() [][]string {
	rows := make([][]string, len(c.identities))
	for i, ident := range c.identities {
		data := [5]string{
			ident.Id,
			cmdx.None,
			cmdx.None,
			cmdx.None,
			cmdx.None,
		}

		if len(ident.VerifiableAddresses) != 0 {
			data[1] = (ident.VerifiableAddresses)[0].Value
		}

		if len(ident.RecoveryAddresses) != 0 {
			data[2] = (ident.RecoveryAddresses)[0].Value
		}

		data[3] = ident.SchemaId
		data[4] = ident.SchemaUrl

		rows[i] = data[:]
	}
	return rows
}

func (c *outputIdentityCollection) Interface() interface{} {
	return c.identities
}

func (c *outputIdentityCollection) Len() int {
	return len(c.identities)
}

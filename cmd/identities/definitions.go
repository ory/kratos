package identities

import (
	"strings"

	kratos "github.com/ory/kratos-client-go"

	"github.com/ory/x/cmdx"
)

type (
	outputIder               string
	outputIdentity           kratos.Identity
	outputIdentityCollection struct {
		identities []kratos.Identity
	}
	outputIderCollection struct {
		ids []outputIder
	}
)

func (_ outputIdentity) Header() []string {
	return []string{"ID", "VERIFIED ADDRESSES", "RECOVERY ADDRESSES", "SCHEMA ID", "SCHEMA URL"}
}

func (i outputIdentity) Columns() []string {
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

func (i outputIdentity) Interface() interface{} {
	return i
}

func (_ outputIder) Header() []string {
	return []string{"ID"}
}

func (i outputIder) Columns() []string {
	return []string{string(i)}
}

func (i outputIder) Interface() interface{} {
	return i
}

func (_ outputIdentityCollection) Header() []string {
	return outputIdentity{}.Header()
}

func (c outputIdentityCollection) Table() [][]string {
	rows := make([][]string, len(c.identities))
	for i, ident := range c.identities {
		rows[i] = outputIdentity(ident).Columns()
	}
	return rows
}

func (c outputIdentityCollection) Interface() interface{} {
	return c.identities
}

func (c *outputIdentityCollection) Len() int {
	return len(c.identities)
}

func (_ *outputIderCollection) Header() []string {
	return []string{"ID"}
}

func (c *outputIderCollection) Table() [][]string {
	rows := make([][]string, len(c.ids))
	for i, ident := range c.ids {
		rows[i] = []string{string(ident)}
	}
	return rows
}

func (c *outputIderCollection) Interface() interface{} {
	return c.ids
}

func (c *outputIderCollection) Len() int {
	return len(c.ids)
}

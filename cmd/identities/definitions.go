package identities

import (
	"strings"

	"github.com/ory/x/cmdx"

	"github.com/ory/kratos-client-go/models"
)

type (
	outputIdentity           models.Identity
	outputIdentityCollection struct {
		identities []*models.Identity
	}
)

func (_ *outputIdentity) Header() []string {
	return []string{"ID", "VERIFIED ADDRESSES", "RECOVERY ADDRESSES", "SCHEMA ID", "SCHEMA URL"}
}

func (i *outputIdentity) Columns() []string {
	data := [5]string{
		string(i.ID),
		cmdx.None,
		cmdx.None,
		cmdx.None,
		cmdx.None,
	}

	addresses := make([]string, 0, len(i.VerifiableAddresses))
	for _, a := range i.VerifiableAddresses {
		if a.Value != nil {
			addresses = append(addresses, *a.Value)
		}
	}
	data[1] = strings.Join(addresses, ", ")

	addresses = addresses[:0]
	for _, a := range i.RecoveryAddresses {
		if a.Value != nil {
			addresses = append(addresses, *a.Value)
		}
	}
	data[2] = strings.Join(addresses, ", ")

	if i.SchemaID != nil {
		data[3] = *i.SchemaID
	}

	if i.SchemaURL != nil {
		data[4] = *i.SchemaURL
	}

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
			string(ident.ID),
			cmdx.None,
			cmdx.None,
			cmdx.None,
			cmdx.None,
		}

		if len(ident.VerifiableAddresses) != 0 && ident.VerifiableAddresses[0].Value != nil {
			data[1] = *ident.VerifiableAddresses[0].Value
		}

		if len(ident.RecoveryAddresses) != 0 && ident.RecoveryAddresses[0].Value != nil {
			data[2] = *ident.RecoveryAddresses[0].Value
		}

		if ident.SchemaID != nil {
			data[3] = *ident.SchemaID
		}

		if ident.SchemaURL != nil {
			data[4] = *ident.SchemaURL
		}

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

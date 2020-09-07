package identities

import (
	"strings"

	"github.com/ory/kratos/internal/clihelpers"
	"github.com/ory/kratos/internal/httpclient/models"
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

func (i *outputIdentity) Fields() []string {
	data := [5]string{
		string(i.ID),
		clihelpers.None,
		clihelpers.None,
		clihelpers.None,
		i.SchemaURL,
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
			clihelpers.None,
			clihelpers.None,
			clihelpers.None,
			ident.SchemaURL,
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

		rows[i] = data[:]
	}
	return rows
}

func (c *outputIdentityCollection) Interface() interface{} {
	return c.identities
}

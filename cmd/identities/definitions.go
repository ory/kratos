// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identities

import (
	"strings"

	kratos "github.com/ory/kratos/internal/httpclient"

	"github.com/ory/x/cmdx"
)

type (
	outputIdentity           kratos.Identity
	outputIdentityCollection struct {
		Identities       []kratos.Identity `json:"identities"`
		NextPageToken    string            `json:"next_page_token"`
		includePageToken bool
	}
)

func (outputIdentity) Header() []string {
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

func (outputIdentityCollection) Header() []string {
	return outputIdentity{}.Header()
}

func (c outputIdentityCollection) Table() [][]string {
	rows := make([][]string, len(c.Identities))
	for i, ident := range c.Identities {
		rows[i] = outputIdentity(ident).Columns()
	}
	return append(rows,
		[]string{""},
		[]string{"NEXT PAGE TOKEN", c.NextPageToken},
	)
}

func (c outputIdentityCollection) Interface() interface{} {
	if c.includePageToken {
		return c
	}
	return c.Identities
}

func (c *outputIdentityCollection) Len() int {
	return len(c.Identities)
}

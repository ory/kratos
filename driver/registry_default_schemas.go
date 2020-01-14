package driver

import (
	"github.com/ory/kratos/schema"
	"net/url"
)

func (m *RegistryDefault) IdentityTraitsSchemas() schema.Schemas {
	ms := m.c.IdentityTraitsSchemas()
	var ss schema.Schemas

	for i := range ms {
		uri, err := url.Parse(ms[i]["url"])
		if err != nil {
			m.l.Fatalf("Could not parse url %s for schema %s", ms[i]["url"], ms[i]["id"])
		}

		ss = append(ss, schema.Schema{
			ID:     ms[i]["id"],
			URL:    uri,
			RawURL: ms[i]["url"],
		})
	}

	return ss
}

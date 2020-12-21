package driver

import (
	"net/url"

	"github.com/ory/kratos/schema"
)

func (m *RegistryDefault) IdentityTraitsSchemas() schema.Schemas {
	ms := m.c.IdentityTraitsSchemas()
	var ss schema.Schemas

	for _, s := range ms {
		surl, err := url.Parse(s.URL)
		if err != nil {
			m.l.Fatalf("Could not parse url %s for schema %s", s.URL, s.ID)
		}

		ss = append(ss, schema.Schema{
			ID:     s.ID,
			URL:    surl,
			RawURL: s.URL,
		})
	}

	return ss
}

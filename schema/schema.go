package schema

import (
	"net/url"

	"github.com/ory/kratos/driver/configuration"

	"github.com/pkg/errors"

	"github.com/ory/x/urlx"
)

type Schemas []Schema

func (s Schemas) GetByID(id string) (*Schema, error) {
	if id == "" {
		id = configuration.DefaultIdentityTraitsSchemaID
	}

	for _, ss := range s {
		if ss.ID == id {
			return &ss, nil
		}
	}

	return nil, errors.Errorf("unable to find JSON schema with ID: %s", id)
}

type Schema struct {
	ID     string   `json:"id"`
	URL    *url.URL `json:"-"`
	RawURL string   `json:"url"`
}

func (s *Schema) SchemaURL(host *url.URL) *url.URL {
	return urlx.AppendPaths(host, SchemasPath, s.ID)
}

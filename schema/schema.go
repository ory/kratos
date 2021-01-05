package schema

import (
	"context"
	"io/ioutil"
	"net/url"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/herodot"
	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/urlx"
)

type Schemas []Schema
type IdentityTraitsProvider interface {
	IdentityTraitsSchemas(ctx context.Context) Schemas
}

func (s Schemas) GetByID(id string) (*Schema, error) {
	if id == "" {
		id = config.DefaultIdentityTraitsSchemaID
	}

	for _, ss := range s {
		if ss.ID == id {
			return &ss, nil
		}
	}

	return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Unable to find JSON Schema ID: %s", id))
}

var orderedKeyCacheMutex sync.RWMutex
var orderedKeyCache map[string][]string

func init() {
	orderedKeyCache = make(map[string][]string)
}

func computeKeyPositions(schema []byte, dest *[]string, parents []string) {
	switch gjson.GetBytes(schema, "type").String() {
	case "object":
		gjson.GetBytes(schema, "properties").ForEach(func(key, value gjson.Result) bool {
			computeKeyPositions([]byte(value.Raw), dest, append(parents, strings.Replace(key.String(), ".", "\\.", -1)))
			return true
		})
	default:
		*dest = append(*dest, strings.Join(parents, "."))
	}
}

func GetKeysInOrder(schemaRef string) ([]string, error) {
	orderedKeyCacheMutex.RLock()
	keysInOrder, ok := orderedKeyCache[schemaRef]
	orderedKeyCacheMutex.RUnlock()
	if !ok {
		sio, err := jsonschema.LoadURL(schemaRef)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		schema, err := ioutil.ReadAll(sio)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		computeKeyPositions(schema, &keysInOrder, []string{})
		orderedKeyCacheMutex.Lock()
		orderedKeyCache[schemaRef] = keysInOrder
		orderedKeyCacheMutex.Unlock()
	}

	return keysInOrder, nil
}

type Schema struct {
	ID     string   `json:"id"`
	URL    *url.URL `json:"-"`
	RawURL string   `json:"url"`
}

func (s *Schema) SchemaURL(host *url.URL) *url.URL {
	return urlx.AppendPaths(host, SchemasPath, s.ID)
}

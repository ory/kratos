// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/herodot"
	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/pagination"
	"github.com/ory/x/urlx"
)

type Schemas []Schema
type IdentityTraitsProvider interface {
	IdentityTraitsSchemas(ctx context.Context) (Schemas, error)
}

func (s Schemas) GetByID(id string, fallbackTemplate string) (*Schema, error) {
	if id == "" {
		id = config.DefaultIdentityTraitsSchemaID
	}

	for _, ss := range s {
		if ss.ID == id {
			return &ss, nil
		}
	}

	if fallbackTemplate != "" {
		source := fmt.Sprintf(fallbackTemplate, id)

		parsedURL, err := url.Parse(source)
		if err != nil {
			return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Unable to parse identity schema fallback tempalte. Please contact the site's administrator."))
		}

		return &Schema{
			ID:     id,
			URL:    parsedURL,
			RawURL: fmt.Sprintf(fallbackTemplate, id),
		}, nil
	}

	return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Unable to find JSON Schema ID: %s", id))
}

func (s Schemas) Total() int {
	return len(s)
}

func (s Schemas) List(page, perPage int) Schemas {
	if page < 0 {
		page = 0
	}
	if perPage < 1 {
		perPage = 1
	}
	start, end := pagination.Index((page+1)*perPage, page*perPage, len(s))
	return s[start:end]
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
			computeKeyPositions([]byte(value.Raw), dest, append(parents, strings.ReplaceAll(key.String(), ".", "\\.")))
			return true
		})
	default:
		*dest = append(*dest, strings.Join(parents, "."))
	}
}

func GetKeysInOrder(ctx context.Context, schemaRef string) ([]string, error) {
	orderedKeyCacheMutex.RLock()
	keysInOrder, ok := orderedKeyCache[schemaRef]
	orderedKeyCacheMutex.RUnlock()
	if !ok {
		sio, err := jsonschema.LoadURL(ctx, schemaRef)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		schema, err := io.ReadAll(io.LimitReader(sio, 1024*1024))
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
	ID  string   `json:"id"`
	URL *url.URL `json:"-"`
	// RawURL contains the raw URL value as it was passed in the configuration. URL parsing can break base64 encoded URLs.
	RawURL string `json:"url"`
}

func (s *Schema) SchemaURL(host *url.URL) *url.URL {
	return IDToURL(host, s.ID)
}

func IDToURL(host *url.URL, id string) *url.URL {
	return urlx.AppendPaths(host, SchemasPath, base64.RawURLEncoding.EncodeToString([]byte(id)))
}

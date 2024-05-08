// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"cmp"
	"context"
	"encoding/base64"
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

var _ IdentitySchemaList = (*Schemas)(nil)

type Schemas []Schema

type IdentitySchemaProvider interface {
	IdentityTraitsSchemas(ctx context.Context) (IdentitySchemaList, error)
}

type deps interface {
	config.Provider
}

type DefaultIdentitySchemaProvider struct {
	d deps
}

func NewDefaultIdentityTraitsProvider(d deps) *DefaultIdentitySchemaProvider {
	return &DefaultIdentitySchemaProvider{d: d}
}

func (d *DefaultIdentitySchemaProvider) IdentityTraitsSchemas(ctx context.Context) (IdentitySchemaList, error) {
	ms, err := d.d.Config().IdentityTraitsSchemas(ctx)
	if err != nil {
		return nil, err
	}

	var ss Schemas
	for _, s := range ms {
		surl, err := url.Parse(s.URL)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		ss = append(ss, Schema{
			ID:     s.ID,
			URL:    surl,
			RawURL: s.URL,
		})
	}

	return ss, nil
}

type IdentitySchemaList interface {
	GetByID(id string) (*Schema, error)
	Total() int
	List(page, perPage int) Schemas
}

func (s Schemas) GetByID(id string) (*Schema, error) {
	id = cmp.Or(id, config.DefaultIdentityTraitsSchemaID)

	for _, ss := range s {
		if ss.ID == id {
			return &ss, nil
		}
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

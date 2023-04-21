// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"net/url"

	"github.com/ory/kratos/x"

	"github.com/pkg/errors"

	"github.com/ory/kratos/schema"
)

func (m *RegistryDefault) IdentityTraitsSchemas(ctx context.Context) (schema.Schemas, error) {
	ms, err := m.Config().IdentityTraitsSchemas(ctx)
	if err != nil {
		return nil, err
	}

	var ss schema.Schemas
	for i, s := range ms {
		surl, err := url.Parse(s.URL)
		if err != nil {
			return nil, errors.WithStack(x.ErrMisconfiguration.WithReasonf("Unable to parse Identity Schema URL: %d", i))
		}

		ss = append(ss, schema.Schema{
			ID:     s.ID,
			URL:    surl,
			RawURL: s.URL,
		})
	}

	return ss, nil
}

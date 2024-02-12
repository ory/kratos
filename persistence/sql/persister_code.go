// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"
)

type oneTimeCodeProvider interface {
	GetID() uuid.UUID
	Validate() error
	TableName(ctx context.Context) string
	GetHMACCode() string
}

type codeOptions struct {
	IdentityID *uuid.UUID
}

type codeOption func(o *codeOptions)

func withCheckIdentityID(id uuid.UUID) codeOption {
	return func(o *codeOptions) {
		o.IdentityID = &id
	}
}

func useOneTimeCode[P any, U interface {
	*P
	oneTimeCodeProvider
}](ctx context.Context, p *Persister, flowID uuid.UUID, userProvidedCode string, flowTableName string, foreignKeyName string, opts ...codeOption,
) (_ U, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.useOneTimeCode")
	defer otelx.End(span, &err)

	o := new(codeOptions)
	for _, opt := range opts {
		opt(o)
	}

	var target U
	nid := p.NetworkID(ctx)
	if err := p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		//#nosec G201 -- TableName is static
		if err := tx.RawQuery(fmt.Sprintf("UPDATE %s SET submit_count = submit_count + 1 WHERE id = ? AND nid = ?", flowTableName), flowID, nid).Exec(); err != nil {
			return err
		}

		var submitCount int
		// Because MySQL does not support "RETURNING" clauses, but we need the updated `submit_count` later on.
		//#nosec G201 -- TableName is static
		if err := sqlcon.HandleError(tx.RawQuery(fmt.Sprintf("SELECT submit_count FROM %s WHERE id = ? AND nid = ?", flowTableName), flowID, nid).First(&submitCount)); err != nil {
			if errors.Is(err, sqlcon.ErrNoRows) {
				// Return no error, as that would roll back the transaction
				return nil
			}
			return err
		}

		// This check prevents parallel brute force attacks by checking the submit count inside this database
		// transaction. If the flow has been submitted more than 5 times, the transaction is aborted (regardless of
		// whether the code was correct or not) and we thus give no indication whether the supplied code was correct or
		// not. For more explanation see [this comment](https://github.com/ory/kratos/pull/2645#discussion_r984732899).
		if submitCount > 5 {
			return errors.WithStack(code.ErrCodeSubmittedTooOften)
		}

		var codes []U
		codesQuery := tx.Where(fmt.Sprintf("nid = ? AND %s = ?", foreignKeyName), nid, flowID)
		if o.IdentityID != nil {
			codesQuery = codesQuery.Where("identity_id = ?", *o.IdentityID)
		}

		if err := sqlcon.HandleError(codesQuery.All(&codes)); err != nil {
			if errors.Is(err, sqlcon.ErrNoRows) {
				// Return no error, as that would roll back the transaction and reset the submit count.
				return nil
			}

			return err
		}

	secrets:
		for _, secret := range p.r.Config().SecretsSession(ctx) {
			suppliedCode := []byte(p.hmacValueWithSecret(ctx, userProvidedCode, secret))
			for i := range codes {
				c := codes[i]
				if subtle.ConstantTimeCompare([]byte(c.GetHMACCode()), suppliedCode) == 0 {
					// Not the supplied code
					continue
				}
				target = c
				break secrets
			}
		}

		if target.Validate() != nil {
			// Return no error, as that would roll back the transaction
			return nil
		}

		//#nosec G201 -- TableName is static
		return tx.RawQuery(fmt.Sprintf("UPDATE %s SET used_at = ? WHERE id = ? AND nid = ?", target.TableName(ctx)), time.Now().UTC(), target.GetID(), nid).Exec()
	}); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if err := target.Validate(); err != nil {
		return nil, err
	}

	return target, nil
}

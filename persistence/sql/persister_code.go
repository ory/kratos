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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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
) (target U, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.useOneTimeCode")
	defer otelx.End(span, &err)

	o := new(codeOptions)
	for _, opt := range opts {
		opt(o)
	}

	// Before we do anything else, increment the submit count and check if we're
	// being brute-forced. This is a separate statement/transaction to the rest
	// of the operations so that it is correct for all transaction isolation
	// levels.
	submitCount, err := incrementOTPCodeSubmitCount(ctx, p, flowID, flowTableName)
	if err != nil {
		return nil, err
	}
	if submitCount > 5 {
		return nil, errors.WithStack(code.ErrCodeSubmittedTooOften)
	}

	nid := p.NetworkID(ctx)

	if err := p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		var codes []U
		codesQuery := tx.Where(fmt.Sprintf("nid = ? AND %s = ?", foreignKeyName), nid, flowID)
		if o.IdentityID != nil {
			codesQuery = codesQuery.Where("identity_id = ?", *o.IdentityID)
		}

		if err := sqlcon.HandleError(codesQuery.All(&codes)); err != nil {
			if errors.Is(err, sqlcon.ErrNoRows) {
				return errors.WithStack(code.ErrCodeNotFound)
			}
			return err
		}

	secrets:
		for _, secret := range p.r.Config().SecretsSession(ctx) {
			suppliedCode := []byte(hmacValueWithSecret(ctx, userProvidedCode, secret))
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
			// Return no error, as that would roll back the transaction. We re-validate the code after the transaction.
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

func incrementOTPCodeSubmitCount(ctx context.Context, p *Persister, flowID uuid.UUID, flowTableName string) (submitCount int, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.incrementOTPCodeSubmitCount",
		trace.WithAttributes(attribute.Stringer("flow_id", flowID), attribute.String("flow_table_name", flowTableName)))
	defer otelx.End(span, &err)
	defer func() {
		span.SetAttributes(attribute.Int("submit_count", submitCount))
	}()

	nid := p.NetworkID(ctx)

	// The branch below is a marginal performance optimization for databases
	// supporting RETURNING (one query instead of two). There is no real
	// security difference here, but there is an observable difference in
	// behavior.
	//
	// Databases supporting RETURNING will perform the increment+select
	// atomically. That means that always exactly 5 attempts will be allowed for
	// each flow, no matter how many concurrent attempts are made.
	//
	// Databases without support for RETURNING (MySQL) will perform the UPDATE
	// and SELECT in two queries, which are not atomic. The effect is that there
	// will still never be more than 5 attempts for each flow, but there may be
	// fewer before we reject. Under normal operation, this is never a problem
	// because a human will never submit their code as quickly as would be
	// required to trigger this race condition.
	//
	// In a very strict sense of the word, the MySQL implementation is even more
	// secure than the RETURNING implementation. But we're ok either way :)
	if p.c.Dialect.Name() == "mysql" {
		//#nosec G201 -- TableName is static
		qUpdate := fmt.Sprintf("UPDATE %s SET submit_count = submit_count + 1 WHERE id = ? AND nid = ?", flowTableName)
		if err := p.GetConnection(ctx).RawQuery(qUpdate, flowID, nid).Exec(); err != nil {
			return 0, sqlcon.HandleError(err)
		}
		//#nosec G201 -- TableName is static
		qSelect := fmt.Sprintf("SELECT submit_count FROM %s WHERE id = ? AND nid = ?", flowTableName)
		err = sqlcon.HandleError(p.GetConnection(ctx).RawQuery(qSelect, flowID, nid).First(&submitCount))
	} else {
		//#nosec G201 -- TableName is static
		q := fmt.Sprintf("UPDATE %s SET submit_count = submit_count + 1 WHERE id = ? AND nid = ? RETURNING submit_count", flowTableName)
		err = sqlcon.HandleError(p.Connection(ctx).RawQuery(q, flowID, nid).First(&submitCount))
	}
	if errors.Is(err, sqlcon.ErrNoRows) {
		return 0, errors.WithStack(code.ErrCodeNotFound)
	}

	return submitCount, err
}

package sql

import (
	"context"

	"github.com/gobuffalo/pop/v5"
)

type transactionContextKey int

const transactionKey transactionContextKey = 0

func WithTransaction(ctx context.Context, tx *pop.Connection) context.Context {
	return context.WithValue(ctx, transactionKey, tx)
}

func (p *Persister) Transaction(ctx context.Context, callback func(ctx context.Context, connection *pop.Connection) error) error {
	c := ctx.Value(transactionKey)
	if c != nil {
		if conn, ok := c.(*pop.Connection); ok {
			return callback(ctx, conn.WithContext(ctx))
		}
	}

	return p.c.WithContext(ctx).Transaction(func(tx *pop.Connection) error {
		return callback(WithTransaction(ctx, tx), tx)
	})
}

func (p *Persister) GetConnection(ctx context.Context) *pop.Connection {
	c := ctx.Value(transactionKey)
	if c != nil {
		if conn, ok := c.(*pop.Connection); ok {
			return conn.WithContext(ctx)
		}
	}
	return p.c.WithContext(ctx)
}

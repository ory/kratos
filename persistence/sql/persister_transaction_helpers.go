package sql

import (
	"context"

	"github.com/gobuffalo/pop/v5"
)

type transactionContextKey int

const transactionKey transactionContextKey = 0

func (p *Persister) Transaction(ctx context.Context, callback func(ctx context.Context, connection *pop.Connection) error) error {
	c := ctx.Value(transactionKey)
	if c != nil {
		if conn, ok := c.(*pop.Connection); ok {
			return callback(ctx, conn)
		}
	}

	return p.c.Transaction(func(tx *pop.Connection) error {
		return callback(context.WithValue(ctx, transactionKey, tx), tx)
	})
}

func (p *Persister) GetConnection(ctx context.Context) *pop.Connection {
	c := ctx.Value(transactionKey)
	if c != nil {
		if conn, ok := c.(*pop.Connection); ok {
			return conn
		}
	}
	return p.c
}

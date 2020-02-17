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

func (p *Persister) Transaction(ctx context.Context, callback func(connection *pop.Connection) error) error {
	c := ctx.Value(transactionKey)
	if c != nil {
		if conn, ok := c.(*pop.Connection); ok {
			return callback(conn)
		}
	}
	return p.c.Transaction(callback)
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

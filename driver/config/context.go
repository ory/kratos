package config

import "context"

type ContextKey int

const (
	RequestContextKey ContextKey = iota + 1
	ValidContextKey
)

var RootContext = context.WithValue(context.Background(), ValidContextKey, true)

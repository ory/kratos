package tablename

import "context"

func Contextualize(_ context.Context, name string) string {
	return name
}

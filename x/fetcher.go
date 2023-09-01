package x

import "github.com/ory/x/jwksx"

type JWKFetchProvider interface {
	Fetcher() *jwksx.FetcherNext
}

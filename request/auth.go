package request

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type (
	AuthStrategy interface {
		apply(req *http.Request)
	}

	authStrategyFactory func(c json.RawMessage) (AuthStrategy, error)
)

var strategyFactories = map[string]authStrategyFactory{
	"":           newNoopAuthStrategy,
	"api_key":    newApiKeyStrategy,
	"basic_auth": newBasicAuthStrategy,
}

func authStrategy(name string, config json.RawMessage) (AuthStrategy, error) {
	strategyFactory, ok := strategyFactories[name]
	if ok {
		return strategyFactory(config)
	}

	return nil, fmt.Errorf("unsupported auth type: %s", name)

}

package x

import "github.com/ory/kratos/driver/configuration"

func RedactInProd(d configuration.Provider, value interface{}) interface{} {
	if d.IsInsecureDevMode() {
		return value
	}
	return "This value has been redacted to prevent leak of sensitive information to logs. Switch to ORY Kratos Development Mode using --dev to view the original value."
}

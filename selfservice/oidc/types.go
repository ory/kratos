package oidc

type CredentialsConfig struct {
	Subject  string `json:"subject"`
	Provider string `json:"provider"`
}

type RequestMethodConfigProvider struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type RequestMethodConfig struct {
	Error     string                        `json:"error"`
	Providers []RequestMethodConfigProvider `json:"providers"`
}

type request interface {
	Valid() error
	GetID() string
}

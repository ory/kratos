package saml

import (
	"bytes"
	"context"

	"github.com/crewjam/saml/samlsp"
	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/jsonx"
)

type ProviderSAML struct {
	config *Configuration
	reg    registrationStrategyDependencies
}

func NewProviderSAML(
	config *Configuration,
	reg registrationStrategyDependencies,
) *ProviderSAML {
	return &ProviderSAML{
		config: config,
		reg:    reg,
	}
}

// Translate attributes from saml asseryion into kratos claims
func (d *ProviderSAML) Claims(ctx context.Context, config *config.Config, attributeSAML samlsp.Attributes, pid string) (*Claims, error) {

	var c ConfigurationCollection

	conf := config.SelfServiceStrategy(ctx, "saml").Config
	if err := jsonx.
		NewStrictDecoder(bytes.NewBuffer(conf)).
		Decode(&c); err != nil {
		return nil, errors.Wrapf(err, "Unable to decode config %v", string(conf))
	}

	providerConfig, err := c.ProviderConfig(pid)
	if err != nil {
		return nil, err
	}

	claims := &Claims{
		Issuer:        "saml",
		Subject:       attributeSAML.Get(providerConfig.AttributesMap["id"]),
		Name:          attributeSAML.Get(providerConfig.AttributesMap["firstname"]),
		LastName:      attributeSAML.Get(providerConfig.AttributesMap["lastname"]),
		Nickname:      attributeSAML.Get(providerConfig.AttributesMap["nickname"]),
		Gender:        attributeSAML.Get(providerConfig.AttributesMap["gender"]),
		Birthdate:     attributeSAML.Get(providerConfig.AttributesMap["birthdate"]),
		Picture:       attributeSAML.Get(providerConfig.AttributesMap["picture"]),
		Email:         attributeSAML.Get(providerConfig.AttributesMap["email"]),
		Roles:         attributeSAML[providerConfig.AttributesMap["roles"]],
		Groups:        attributeSAML[providerConfig.AttributesMap["groups"]],
		PhoneNumber:   attributeSAML.Get(providerConfig.AttributesMap["phone_number"]),
		EmailVerified: true,
	}

	return claims, nil
}

func (d *ProviderSAML) Config() *Configuration {
	return d.config
}

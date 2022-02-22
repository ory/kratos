package oidc

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/url"
	"time"

	"github.com/form3tech-oss/jwt-go"
	"github.com/pkg/errors"

	"golang.org/x/oauth2"
)

type ProviderApple struct {
	*ProviderGenericOIDC
}

func NewProviderApple(
	config *Configuration,
	reg dependencies,
) *ProviderApple {
	return &ProviderApple{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}

func (a *ProviderApple) newClientSecret() (string, error) {
	// decode the pem format
	block, _ := pem.Decode([]byte(a.config.PrivateKey))
	if block == nil || block.Type != "PRIVATE KEY" {
		return "", errors.New("failed to decode PEM block containing private key")
	}

	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", errors.Wrap(err, "Private key decoding failed")
	}
	privateKey, ok := parsedKey.(*ecdsa.PrivateKey)
	if !ok {
		return "", errors.New("Private key is not ecdsa key")
	}

	now := time.Now()
	expirationTime := time.Now().Add(5 * time.Minute)

	appleToken := jwt.NewWithClaims(jwt.SigningMethodES256,
		jwt.StandardClaims{
			Audience:  []string{a.config.IssuerURL},
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  now.Unix(),
			Issuer:    a.config.TeamId,
			Subject:   a.config.ClientID,
		})
	appleToken.Header["kid"] = a.config.PrivateKeyId

	return appleToken.SignedString(privateKey)
}

func (a *ProviderApple) oauth2(ctx context.Context) (*oauth2.Config, error) {
	// Apple requires a JWT token that acts as a client secret
	secret, err := a.newClientSecret()
	if err != nil {
		return nil, err
	}
	a.config.ClientSecret = secret

	endpoint := oauth2.Endpoint{
		AuthURL:  "https://appleid.apple.com/auth/authorize",
		TokenURL: "https://appleid.apple.com/auth/token",
	}
	return a.oauth2ConfigFromEndpoint(ctx, endpoint), nil
}

func (a *ProviderApple) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return a.oauth2(ctx)
}

func (a *ProviderApple) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	var options []oauth2.AuthCodeOption

	if isForced(r) {
		options = append(options, oauth2.SetAuthURLParam("prompt", "login"))
	}
	if len(a.config.RequestedClaims) != 0 {
		options = append(options, oauth2.SetAuthURLParam("claims", string(a.config.RequestedClaims)))
	}

	// When requesting email or name, Apple requires the form_post response mode.
	// This also means the return url will be called by Apple using POST method.
	for _, scope := range a.config.Scope {
		if scope == "email" || scope == "name" {
			options = append(options, oauth2.SetAuthURLParam("response_mode", "form_post"))
			break
		}
	}

	return options
}

func (a *ProviderApple) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	claims, err := a.ProviderGenericOIDC.Claims(ctx, exchange, query)
	if err != nil {
		return claims, err
	}

	// https://developer.apple.com/documentation/sign_in_with_apple/sign_in_with_apple_js/configuring_your_webpage_for_sign_in_with_apple#3331292
	// First name and last name are passed only once in an additional parameter to the redirect URL.
	// There's no way to make sure they haven't been tampered with. Blame Apple.
	var user struct {
		name *struct {
			firstName *string `json:"first_name"`
			lastName  *string `json:"last_name"`
		} `json:"name"`
	}
	if err = json.Unmarshal([]byte(query.Get("user")), &user); err != nil {
		if name := user.name; name != nil {
			if firstName := name.firstName; firstName != nil && claims.GivenName == "" {
				claims.GivenName = *firstName
			}
			if lastName := name.lastName; lastName != nil && claims.LastName == "" {
				claims.LastName = *lastName
			}
		}
	}

	return claims, nil
}

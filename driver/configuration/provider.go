package configuration

import (
	"encoding/json"
	"net/url"
	"time"
)

type HasherArgon2Config struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

type SelfServiceHook struct {
	Run    string          `json:"run"`
	Config json.RawMessage `json:"config"`
}

type SelfServiceStrategy struct {
	Enabled bool            `json:"enabled"`
	Config  json.RawMessage `json:"config"`
}

type Provider interface {
	AdminListenOn() string
	PublicListenOn() string
	DSN() string

	SessionSecrets() [][]byte

	SelfPublicURL() *url.URL
	SelfAdminURL() *url.URL

	DefaultReturnToURL() *url.URL

	LoginURL() *url.URL
	ErrorURL() *url.URL
	MultiFactorURL() *url.URL

	SessionLifespan() time.Duration
	SelfServiceLoginRequestLifespan() time.Duration
	SelfServiceRegistrationRequestLifespan() time.Duration

	SelfServiceStrategy(strategy string) *SelfServiceStrategy
	SelfServiceLoginBeforeHooks() []SelfServiceHook
	SelfServiceRegistrationBeforeHooks() []SelfServiceHook
	SelfServiceLoginAfterHooks(strategy string) []SelfServiceHook
	SelfServiceRegistrationAfterHooks(strategy string) []SelfServiceHook
	SelfServiceLogoutRedirectURL() *url.URL

	DefaultIdentityTraitsSchemaURL() *url.URL

	WhitelistedReturnToDomains() []url.URL

	RegisterURL() *url.URL

	HashersArgon2() *HasherArgon2Config

	// This will now be solved using redirect hooks
	// SignupDefaultReturnToURL() *url.URL
	// AuthnDefaultReturnToURL() *url.URL
}

func MustValidate(p Provider) {
	_ = p.AdminListenOn()
	_ = p.PublicListenOn()
	_ = p.DSN()
	_ = p.SessionSecrets()
	_ = p.MultiFactorURL()
	_ = p.SelfAdminURL()
	_ = p.SelfPublicURL()
	_ = p.LoginURL()
	_ = p.ErrorURL()
	_ = p.SessionLifespan()
	_ = p.WhitelistedReturnToDomains()
	_ = p.SelfServiceLoginRequestLifespan()
	_ = p.RegisterURL()
	_ = p.HashersArgon2()
	_ = p.DefaultIdentityTraitsSchemaURL()
}

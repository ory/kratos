package config

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/publicsuffix"

	"github.com/duo-labs/webauthn/protocol"

	"github.com/duo-labs/webauthn/webauthn"

	"github.com/ory/x/jsonschemax"

	"github.com/ory/x/watcherx"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/kratos/embedx"

	"github.com/ory/x/tlsx"

	"github.com/google/uuid"

	"github.com/stretchr/testify/require"

	"github.com/inhies/go-bytesize"
	kjson "github.com/knadh/koanf/parsers/json"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/tidwall/gjson"

	"github.com/ory/x/configx"
	"github.com/ory/x/dbal"
	"github.com/ory/x/jsonx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/stringsx"
	"github.com/ory/x/tracing"
)

const (
	DefaultIdentityTraitsSchemaID                            = "default"
	DefaultBrowserReturnURL                                  = "default_browser_return_url"
	DefaultSQLiteMemoryDSN                                   = dbal.SQLiteInMemory
	DefaultPasswordHashingAlgorithm                          = "argon2"
	DefaultCipherAlgorithm                                   = "noop"
	UnknownVersion                                           = "unknown version"
	ViperKeyDSN                                              = "dsn"
	ViperKeyCourierSMTPURL                                   = "courier.smtp.connection_uri"
	ViperKeyCourierTemplatesPath                             = "courier.template_override_path"
	ViperKeyCourierSMTPFrom                                  = "courier.smtp.from_address"
	ViperKeyCourierSMTPFromName                              = "courier.smtp.from_name"
	ViperKeyCourierSMTPHeaders                               = "courier.smtp.headers"
	ViperKeySecretsDefault                                   = "secrets.default"
	ViperKeySecretsCookie                                    = "secrets.cookie"
	ViperKeySecretsCipher                                    = "secrets.cipher"
	ViperKeyDisablePublicHealthRequestLog                    = "serve.public.request_log.disable_for_health"
	ViperKeyPublicBaseURL                                    = "serve.public.base_url"
	ViperKeyPublicPort                                       = "serve.public.port"
	ViperKeyPublicHost                                       = "serve.public.host"
	ViperKeyPublicSocketOwner                                = "serve.public.socket.owner"
	ViperKeyPublicSocketGroup                                = "serve.public.socket.group"
	ViperKeyPublicSocketMode                                 = "serve.public.socket.mode"
	ViperKeyPublicTLSCertBase64                              = "serve.public.tls.cert.base64"
	ViperKeyPublicTLSKeyBase64                               = "serve.public.tls.key.base64"
	ViperKeyPublicTLSCertPath                                = "serve.public.tls.cert.path"
	ViperKeyPublicTLSKeyPath                                 = "serve.public.tls.key.path"
	ViperKeyDisableAdminHealthRequestLog                     = "serve.admin.request_log.disable_for_health"
	ViperKeyAdminBaseURL                                     = "serve.admin.base_url"
	ViperKeyAdminPort                                        = "serve.admin.port"
	ViperKeyAdminHost                                        = "serve.admin.host"
	ViperKeyAdminSocketOwner                                 = "serve.admin.socket.owner"
	ViperKeyAdminSocketGroup                                 = "serve.admin.socket.group"
	ViperKeyAdminSocketMode                                  = "serve.admin.socket.mode"
	ViperKeyAdminTLSCertBase64                               = "serve.admin.tls.cert.base64"
	ViperKeyAdminTLSKeyBase64                                = "serve.admin.tls.key.base64"
	ViperKeyAdminTLSCertPath                                 = "serve.admin.tls.cert.path"
	ViperKeyAdminTLSKeyPath                                  = "serve.admin.tls.key.path"
	ViperKeySessionLifespan                                  = "session.lifespan"
	ViperKeySessionSameSite                                  = "session.cookie.same_site"
	ViperKeySessionDomain                                    = "session.cookie.domain"
	ViperKeySessionName                                      = "session.cookie.name"
	ViperKeySessionPath                                      = "session.cookie.path"
	ViperKeySessionPersistentCookie                          = "session.cookie.persistent"
	ViperKeySessionWhoAmIAAL                                 = "session.whoami.required_aal"
	ViperKeyCookieSameSite                                   = "cookies.same_site"
	ViperKeyCookieDomain                                     = "cookies.domain"
	ViperKeyCookiePath                                       = "cookies.path"
	ViperKeySelfServiceStrategyConfig                        = "selfservice.methods"
	ViperKeySelfServiceBrowserDefaultReturnTo                = "selfservice." + DefaultBrowserReturnURL
	ViperKeyURLsWhitelistedReturnToDomains                   = "selfservice.whitelisted_return_urls"
	ViperKeySelfServiceRegistrationEnabled                   = "selfservice.flows.registration.enabled"
	ViperKeySelfServiceRegistrationUI                        = "selfservice.flows.registration.ui_url"
	ViperKeySelfServiceRegistrationRequestLifespan           = "selfservice.flows.registration.lifespan"
	ViperKeySelfServiceRegistrationAfter                     = "selfservice.flows.registration.after"
	ViperKeySelfServiceRegistrationBeforeHooks               = "selfservice.flows.registration.before.hooks"
	ViperKeySelfServiceLoginUI                               = "selfservice.flows.login.ui_url"
	ViperKeySelfServiceLoginRequestLifespan                  = "selfservice.flows.login.lifespan"
	ViperKeySelfServiceLoginAfter                            = "selfservice.flows.login.after"
	ViperKeySelfServiceLoginBeforeHooks                      = "selfservice.flows.login.before.hooks"
	ViperKeySelfServiceErrorUI                               = "selfservice.flows.error.ui_url"
	ViperKeySelfServiceLogoutBrowserDefaultReturnTo          = "selfservice.flows.logout.after." + DefaultBrowserReturnURL
	ViperKeySelfServiceSettingsURL                           = "selfservice.flows.settings.ui_url"
	ViperKeySelfServiceSettingsAfter                         = "selfservice.flows.settings.after"
	ViperKeySelfServiceSettingsRequestLifespan               = "selfservice.flows.settings.lifespan"
	ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter = "selfservice.flows.settings.privileged_session_max_age"
	ViperKeySelfServiceSettingsRequiredAAL                   = "selfservice.flows.settings.required_aal"
	ViperKeySelfServiceRecoveryAfter                         = "selfservice.flows.recovery.after"
	ViperKeySelfServiceRecoveryEnabled                       = "selfservice.flows.recovery.enabled"
	ViperKeySelfServiceRecoveryUI                            = "selfservice.flows.recovery.ui_url"
	ViperKeySelfServiceRecoveryRequestLifespan               = "selfservice.flows.recovery.lifespan"
	ViperKeySelfServiceRecoveryBrowserDefaultReturnTo        = "selfservice.flows.recovery.after." + DefaultBrowserReturnURL
	ViperKeySelfServiceVerificationEnabled                   = "selfservice.flows.verification.enabled"
	ViperKeySelfServiceVerificationUI                        = "selfservice.flows.verification.ui_url"
	ViperKeySelfServiceVerificationRequestLifespan           = "selfservice.flows.verification.lifespan"
	ViperKeySelfServiceVerificationBrowserDefaultReturnTo    = "selfservice.flows.verification.after." + DefaultBrowserReturnURL
	ViperKeySelfServiceVerificationAfter                     = "selfservice.flows.verification.after"
	ViperKeyDefaultIdentitySchemaURL                         = "identity.default_schema_url"
	ViperKeyIdentitySchemas                                  = "identity.schemas"
	ViperKeyHasherAlgorithm                                  = "hashers.algorithm"
	ViperKeyHasherArgon2ConfigMemory                         = "hashers.argon2.memory"
	ViperKeyHasherArgon2ConfigIterations                     = "hashers.argon2.iterations"
	ViperKeyHasherArgon2ConfigParallelism                    = "hashers.argon2.parallelism"
	ViperKeyHasherArgon2ConfigSaltLength                     = "hashers.argon2.salt_length"
	ViperKeyHasherArgon2ConfigKeyLength                      = "hashers.argon2.key_length"
	ViperKeyHasherArgon2ConfigExpectedDuration               = "hashers.argon2.expected_duration"
	ViperKeyHasherArgon2ConfigExpectedDeviation              = "hashers.argon2.expected_deviation"
	ViperKeyHasherArgon2ConfigDedicatedMemory                = "hashers.argon2.dedicated_memory"
	ViperKeyHasherBcryptCost                                 = "hashers.bcrypt.cost"
	ViperKeyCipherAlgorithm                                  = "ciphers.algorithm"
	ViperKeyLinkLifespan                                     = "selfservice.methods.link.config.lifespan"
	ViperKeyLinkBaseURL                                      = "selfservice.methods.link.config.base_url"
	ViperKeyPasswordHaveIBeenPwnedHost                       = "selfservice.methods.password.config.haveibeenpwned_host"
	ViperKeyPasswordHaveIBeenPwnedEnabled                    = "selfservice.methods.password.config.haveibeenpwned_enabled"
	ViperKeyPasswordMaxBreaches                              = "selfservice.methods.password.config.max_breaches"
	ViperKeyIgnoreNetworkErrors                              = "selfservice.methods.password.config.ignore_network_errors"
	ViperKeyTOTPIssuer                                       = "selfservice.methods.totp.config.issuer"
	ViperKeyWebAuthnRPDisplayName                            = "selfservice.methods.webauthn.config.rp.display_name"
	ViperKeyWebAuthnRPID                                     = "selfservice.methods.webauthn.config.rp.id"
	ViperKeyWebAuthnRPOrigin                                 = "selfservice.methods.webauthn.config.rp.origin"
	ViperKeyWebAuthnRPIcon                                   = "selfservice.methods.webauthn.config.rp.issuer"
	ViperKeyVersion                                          = "version"
)

const (
	HighestAvailableAAL                 = "highest_available"
	Argon2DefaultMemory                 = 128 * bytesize.MB
	Argon2DefaultIterations      uint32 = 1
	Argon2DefaultSaltLength      uint32 = 16
	Argon2DefaultKeyLength       uint32 = 32
	Argon2DefaultDuration               = 500 * time.Millisecond
	Argon2DefaultDeviation              = 500 * time.Millisecond
	Argon2DefaultDedicatedMemory        = 1 * bytesize.GB
	BcryptDefaultCost            uint32 = 12
)

// DefaultSessionCookieName returns the default cookie name for the kratos session.
const DefaultSessionCookieName = "ory_kratos_session"

type (
	Argon2 struct {
		Memory            bytesize.ByteSize `json:"memory"`
		Iterations        uint32            `json:"iterations"`
		Parallelism       uint8             `json:"parallelism"`
		SaltLength        uint32            `json:"salt_length"`
		KeyLength         uint32            `json:"key_length"`
		ExpectedDuration  time.Duration     `json:"expected_duration"`
		ExpectedDeviation time.Duration     `json:"expected_deviation"`
		DedicatedMemory   bytesize.ByteSize `json:"dedicated_memory"`
	}
	Bcrypt struct {
		Cost uint32 `json:"cost"`
	}
	SelfServiceHook struct {
		Name   string          `json:"hook"`
		Config json.RawMessage `json:"config"`
	}
	SelfServiceStrategy struct {
		Enabled bool            `json:"enabled"`
		Config  json.RawMessage `json:"config"`
	}
	Schema struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	PasswordPolicy struct {
		HaveIBeenPwnedHost    string `json:"haveibeenpwned_host"`
		HaveIBeenPwnedEnabled bool   `json:"haveibeenpwned_enabled"`
		MaxBreaches           uint   `json:"max_breaches"`
		IgnoreNetworkErrors   bool   `json:"ignore_network_errors"`
	}
	Schemas []Schema
	Config  struct {
		l              *logrusx.Logger
		p              *configx.Provider
		identitySchema *jsonschema.Schema
		stdOutOrErr    io.Writer
	}

	Provider interface {
		Config(ctx context.Context) *Config
	}
)

func (c *Argon2) MarshalJSON() ([]byte, error) {
	type encoded struct {
		Memory            string `json:"memory"`
		Iterations        uint32 `json:"iterations"`
		Parallelism       uint8  `json:"parallelism"`
		SaltLength        uint32 `json:"salt_length"`
		KeyLength         uint32 `json:"key_length"`
		ExpectedDuration  string `json:"minimal_duration"`
		ExpectedDeviation string `json:"expected_deviation"`
		DedicatedMemory   string `json:"dedicated_memory"`
	}

	return json.Marshal(&encoded{
		Memory:            c.Memory.String(),
		Iterations:        c.Iterations,
		Parallelism:       c.Parallelism,
		SaltLength:        c.SaltLength,
		KeyLength:         c.KeyLength,
		ExpectedDuration:  c.ExpectedDuration.String(),
		ExpectedDeviation: c.ExpectedDeviation.String(),
		DedicatedMemory:   c.DedicatedMemory.String(),
	})
}

var Argon2DefaultParallelism = uint8(runtime.NumCPU() * 2)

const HookGlobal = "global"

func HookStrategyKey(key, strategy string) string {
	if strategy == HookGlobal {
		return fmt.Sprintf("%s.hooks", key)
	} else {
		return fmt.Sprintf("%s.%s.hooks", key, strategy)
	}
}

func (s Schemas) FindSchemaByID(id string) (*Schema, error) {
	for _, sc := range s {
		if sc.ID == id {
			return &sc, nil
		}
	}

	return nil, errors.Errorf("could not find schema with id \"%s\"", id)
}

func MustNew(t *testing.T, l *logrusx.Logger, stdOutOrErr io.Writer, opts ...configx.OptionModifier) *Config {
	p, err := New(context.TODO(), l, stdOutOrErr, opts...)
	require.NoError(t, err)
	return p
}

func New(ctx context.Context, l *logrusx.Logger, stdOutOrErr io.Writer, opts ...configx.OptionModifier) (*Config, error) {
	var c *Config

	opts = append([]configx.OptionModifier{
		configx.WithStderrValidationReporter(),
		configx.OmitKeysFromTracing("dsn", "courier.smtp.connection_uri", "secrets.default", "secrets.cookie", "secrets.cipher", "client_secret"),
		configx.WithImmutables("serve", "profiling", "log"),
		configx.WithLogrusWatcher(l),
		configx.WithLogger(l),
		configx.WithContext(ctx),
		configx.AttachWatcher(func(event watcherx.Event, err error) {
			if c == nil {
				panic(errors.New("the config provider did not initialise correctly in time"))
			}
			if err := c.validateIdentitySchemas(); err != nil {
				l.WithError(err).
					Errorf("The changed identity schema configuration is invalid and could not be loaded. Rolling back to the last working configuration revision. Please address the validation errors before restarting the process.")
			}
		}),
	}, opts...)

	p, err := configx.New([]byte(embedx.ConfigSchema), opts...)
	if err != nil {
		return nil, err
	}

	l.UseConfig(p)

	c = &Config{l: l, p: p, stdOutOrErr: stdOutOrErr}

	if !p.SkipValidation() {
		if err := c.validateIdentitySchemas(); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (p *Config) getIdentitySchemaValidator() (*jsonschema.Schema, error) {
	if p.identitySchema == nil {
		c := jsonschema.NewCompiler()
		err := embedx.AddSchemaResources(c, embedx.IdentityMeta)
		if err != nil {
			return nil, err
		}
		p.identitySchema, err = c.Compile(embedx.IdentityMeta.GetSchemaID())
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return p.identitySchema, nil
}

func (p *Config) validateIdentitySchemas() error {
	j, err := p.getIdentitySchemaValidator()
	if err != nil {
		return err
	}

	for _, s := range p.IdentityTraitsSchemas() {
		resource, err := jsonschema.LoadURL(s.URL)
		if err != nil {
			return errors.WithStack(err)
		}
		defer resource.Close()

		schema, err := io.ReadAll(resource)
		if err != nil {
			return errors.WithStack(err)
		}

		if err = j.Validate(bytes.NewBuffer(schema)); err != nil {
			p.formatJsonErrors(schema, err)
			return errors.WithStack(err)
		}
	}
	return nil
}

func (p *Config) formatJsonErrors(schema []byte, err error) {
	_, _ = fmt.Fprintln(p.stdOutOrErr, "")
	jsonschemax.FormatValidationErrorForCLI(p.stdOutOrErr, schema, err)
}

func (p *Config) Source() *configx.Provider {
	return p.p
}

func (p *Config) CORS(iface string) (cors.Options, bool) {
	switch iface {
	case "admin":
		return p.cors("serve.admin")
	case "public":
		return p.cors("serve.public")
	default:
		panic(fmt.Sprintf("Received unexpected CORS interface: %s", iface))
	}
}

func (p *Config) cors(prefix string) (cors.Options, bool) {
	return p.p.CORS(prefix, cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Cookie"},
		ExposedHeaders:   []string{"Content-Type", "Set-Cookie"},
		AllowCredentials: true,
	})
}

func (p *Config) Set(key string, value interface{}) error {
	return p.p.Set(key, value)
}

func (p *Config) MustSet(key string, value interface{}) {
	if err := p.p.Set(key, value); err != nil {
		p.l.WithError(err).Fatalf("Unable to set \"%s\" to \"%s\".", key, value)
	}
}

func (p *Config) SessionName() string {
	return stringsx.Coalesce(p.p.String(ViperKeySessionName), DefaultSessionCookieName)
}

func (p *Config) HasherArgon2() *Argon2 {
	// warn about usage of default values and point to the docs
	// warning will require https://github.com/ory/viper/issues/19
	return &Argon2{
		Memory:            p.p.ByteSizeF(ViperKeyHasherArgon2ConfigMemory, Argon2DefaultMemory),
		Iterations:        uint32(p.p.IntF(ViperKeyHasherArgon2ConfigIterations, int(Argon2DefaultIterations))),
		Parallelism:       uint8(p.p.IntF(ViperKeyHasherArgon2ConfigParallelism, int(Argon2DefaultParallelism))),
		SaltLength:        uint32(p.p.IntF(ViperKeyHasherArgon2ConfigSaltLength, int(Argon2DefaultSaltLength))),
		KeyLength:         uint32(p.p.IntF(ViperKeyHasherArgon2ConfigKeyLength, int(Argon2DefaultKeyLength))),
		ExpectedDuration:  p.p.DurationF(ViperKeyHasherArgon2ConfigExpectedDuration, Argon2DefaultDuration),
		ExpectedDeviation: p.p.DurationF(ViperKeyHasherArgon2ConfigExpectedDeviation, Argon2DefaultDeviation),
		DedicatedMemory:   p.p.ByteSizeF(ViperKeyHasherArgon2ConfigDedicatedMemory, Argon2DefaultDedicatedMemory),
	}
}

func (p *Config) HasherBcrypt() *Bcrypt {
	// warn about usage of default values and point to the docs
	// warning will require https://github.com/ory/viper/issues/19
	cost := uint32(p.p.IntF(ViperKeyHasherBcryptCost, int(BcryptDefaultCost)))
	if !p.IsInsecureDevMode() && cost < BcryptDefaultCost {
		cost = BcryptDefaultCost
	}

	return &Bcrypt{Cost: cost}
}

func (p *Config) listenOn(key string) string {
	fb := 4433
	if key == "admin" {
		fb = 4434
	}

	port := p.p.IntF("serve."+key+".port", fb)
	if port < 1 {
		p.l.Fatalf("serve.%s.port can not be zero or negative", key)
	}

	return configx.GetAddress(p.p.String("serve."+key+".host"), port)
}

func (p *Config) DefaultIdentityTraitsSchemaURL() *url.URL {
	return p.ParseURIOrFail(ViperKeyDefaultIdentitySchemaURL)
}

func (p *Config) TOTPIssuer() string {
	return p.Source().StringF(ViperKeyTOTPIssuer, p.SelfPublicURL().Hostname())
}

func (p *Config) IdentityTraitsSchemas() Schemas {
	ds := Schema{
		ID:  DefaultIdentityTraitsSchemaID,
		URL: p.DefaultIdentityTraitsSchemaURL().String(),
	}

	if !p.p.Exists(ViperKeyIdentitySchemas) {
		return Schemas{ds}
	}

	var ss Schemas
	out, err := p.p.Marshal(kjson.Parser())
	if err != nil {
		p.l.WithError(err).Fatalf("Unable to dencode values from %s.", ViperKeyIdentitySchemas)
		return Schemas{ds}
	}

	config := gjson.GetBytes(out, ViperKeyIdentitySchemas).Raw
	if len(config) == 0 {
		return Schemas{ds}
	}

	if err := json.NewDecoder(bytes.NewBufferString(config)).Decode(&ss); err != nil {
		p.l.WithError(err).Fatalf("Unable to encode values from %s.", ViperKeyIdentitySchemas)
		return Schemas{ds}
	}

	return append(ss, ds)
}

func (p *Config) AdminListenOn() string {
	return p.listenOn("admin")
}

func (p *Config) PublicListenOn() string {
	return p.listenOn("public")
}

func (p *Config) PublicSocketPermission() *configx.UnixPermission {
	return &configx.UnixPermission{
		Owner: p.p.String(ViperKeyPublicSocketOwner),
		Group: p.p.String(ViperKeyPublicSocketGroup),
		Mode:  os.FileMode(p.p.IntF(ViperKeyPublicSocketMode, 0755)),
	}
}

func (p *Config) AdminSocketPermission() *configx.UnixPermission {
	return &configx.UnixPermission{
		Owner: p.p.String(ViperKeyAdminSocketOwner),
		Group: p.p.String(ViperKeyAdminSocketGroup),
		Mode:  os.FileMode(p.p.IntF(ViperKeyAdminSocketMode, 0755)),
	}
}

func (p *Config) DSN() string {
	dsn := p.p.String(ViperKeyDSN)

	if dsn == "memory" {
		return DefaultSQLiteMemoryDSN
	}

	if len(dsn) > 0 {
		return dsn
	}

	p.l.Fatal("dsn must be set")
	return ""
}

func (p *Config) DisableAPIFlowEnforcement() bool {
	if p.IsInsecureDevMode() && os.Getenv("DEV_DISABLE_API_FLOW_ENFORCEMENT") == "true" {
		p.l.Warn("Because \"DEV_DISABLE_API_FLOW_ENFORCEMENT=true\" and the \"--dev\" flag are set, self-service API flows will no longer check if the interaction is actually a browser flow. This is very dangerous as it allows bypassing of anti-CSRF measures, leaving the deployment highly vulnerable. This option should only be used for automated testing and never come close to real user data anywhere.")
		return true
	}
	return false
}

func (p *Config) SelfServiceFlowRegistrationEnabled() bool {
	return p.p.Bool(ViperKeySelfServiceRegistrationEnabled)
}

func (p *Config) SelfServiceFlowVerificationEnabled() bool {
	return p.p.Bool(ViperKeySelfServiceVerificationEnabled)
}

func (p *Config) SelfServiceFlowRecoveryEnabled() bool {
	return p.p.Bool(ViperKeySelfServiceRecoveryEnabled)
}

func (p *Config) SelfServiceFlowLoginBeforeHooks() []SelfServiceHook {
	return p.selfServiceHooks(ViperKeySelfServiceLoginBeforeHooks)
}

func (p *Config) SelfServiceFlowRegistrationBeforeHooks() []SelfServiceHook {
	return p.selfServiceHooks(ViperKeySelfServiceRegistrationBeforeHooks)
}

func (p *Config) selfServiceHooks(key string) []SelfServiceHook {
	var hooks []SelfServiceHook
	if !p.p.Exists(key) {
		return []SelfServiceHook{}
	}

	out, err := p.p.Marshal(kjson.Parser())
	if err != nil {
		p.l.WithError(err).Fatalf("Unable to decode values from configuration key: %s", key)
	}

	config := gjson.GetBytes(out, key).Raw
	if len(config) == 0 {
		return []SelfServiceHook{}
	}

	if err := jsonx.NewStrictDecoder(bytes.NewBufferString(config)).Decode(&hooks); err != nil {
		p.l.WithError(err).Fatalf("Unable to encode value \"%s\" from configuration key: %s", config, key)
	}

	for k := range hooks {
		if len(hooks[k].Config) == 0 {
			hooks[k].Config = json.RawMessage("{}")
		}
	}

	return hooks
}

func (p *Config) SelfServiceFlowLoginAfterHooks(strategy string) []SelfServiceHook {
	return p.selfServiceHooks(HookStrategyKey(ViperKeySelfServiceLoginAfter, strategy))
}

func (p *Config) SelfServiceFlowSettingsAfterHooks(strategy string) []SelfServiceHook {
	return p.selfServiceHooks(HookStrategyKey(ViperKeySelfServiceSettingsAfter, strategy))
}

func (p *Config) SelfServiceFlowRegistrationAfterHooks(strategy string) []SelfServiceHook {
	return p.selfServiceHooks(HookStrategyKey(ViperKeySelfServiceRegistrationAfter, strategy))
}

func (p *Config) SelfServiceStrategy(strategy string) *SelfServiceStrategy {
	config := "{}"
	out, err := p.p.Marshal(kjson.Parser())
	if err != nil {
		p.l.WithError(err).Warn("Unable to marshal self service strategy configuration.")
	} else if c := gjson.GetBytes(out,
		fmt.Sprintf("%s.%s.config", ViperKeySelfServiceStrategyConfig, strategy)).Raw; len(c) > 0 {
		config = c
	}

	enabledKey := fmt.Sprintf("%s.%s.enabled", ViperKeySelfServiceStrategyConfig, strategy)
	s := &SelfServiceStrategy{
		Enabled: p.p.Bool(enabledKey),
		Config:  json.RawMessage(config),
	}

	// The default value can easily be overwritten by setting e.g. `{"selfservice": "null"}` which means that
	// we need to forcibly set these values here:
	if !p.p.Exists(enabledKey) {
		switch strategy {
		case "password":
			fallthrough
		case "profile":
			fallthrough
		case "link":
			s.Enabled = true
		}
	}

	if len(s.Config) == 0 {
		s.Config = json.RawMessage("{}")
	}

	return s
}

func (p *Config) SecretsDefault() [][]byte {
	secrets := p.p.Strings(ViperKeySecretsDefault)

	if len(secrets) == 0 {
		secrets = []string{uuid.New().String()}
		p.MustSet(ViperKeySecretsDefault, secrets)
	}

	result := make([][]byte, len(secrets))
	for k, v := range secrets {
		result[k] = []byte(v)
	}

	return result
}

func (p *Config) SecretsSession() [][]byte {
	secrets := p.p.Strings(ViperKeySecretsCookie)
	if len(secrets) == 0 {
		return p.SecretsDefault()
	}

	result := make([][]byte, len(secrets))
	for k, v := range secrets {
		result[k] = []byte(v)
	}

	return result
}

func (p *Config) SecretsCipher() [][32]byte {
	secrets := p.p.Strings(ViperKeySecretsCipher)
	var cleanSecrets []string
	for k := range secrets {
		if len(secrets[k]) == 32 {
			cleanSecrets = append(cleanSecrets, secrets[k])
		}
	}
	if len(cleanSecrets) == 0 {
		return [][32]byte{}
	}
	result := make([][32]byte, len(cleanSecrets))
	for n, s := range secrets {
		for k, v := range []byte(s) {
			result[n][k] = byte(v)
		}
	}
	return result
}

func (p *Config) SelfServiceBrowserDefaultReturnTo() *url.URL {
	return p.ParseAbsoluteOrRelativeURIOrFail(ViperKeySelfServiceBrowserDefaultReturnTo)
}

func (p *Config) guessBaseURL(keyHost, keyPort string, defaultPort int) *url.URL {
	port := p.p.IntF(keyPort, defaultPort)

	host := p.p.String(keyHost)
	if host == "0.0.0.0" || len(host) == 0 {
		var err error
		host, err = os.Hostname()
		if err != nil {
			p.l.WithError(err).Warn("Unable to get hostname from system, falling back to 127.0.0.1.")
			host = "127.0.0.1"
		}
	}

	guess := url.URL{Host: fmt.Sprintf("%s:%d", host, port), Scheme: "https", Path: "/"}
	if p.IsInsecureDevMode() {
		guess.Scheme = "http"
	}

	return &guess
}

func (p *Config) baseURL(keyURL, keyHost, keyPort string, defaultPort int) *url.URL {
	switch t := p.p.Get(keyURL).(type) {
	case *url.URL:
		return t
	case url.URL:
		return &t
	case string:
		parsed, err := url.ParseRequestURI(t)
		if err != nil {
			p.l.WithError(err).Errorf("Configuration key %s is not a valid URL. Falling back to optimistically guessing the server's base URL. Please set a value to avoid problems with redirects and cookies.", keyURL)
			return p.guessBaseURL(keyHost, keyPort, defaultPort)
		}
		return parsed
	}

	p.l.Warnf("Configuration key %s was left empty. Optimistically guessing the server's base URL. Please set a value to avoid problems with redirects and cookies.", keyURL)
	return p.guessBaseURL(keyHost, keyPort, defaultPort)
}

func (p *Config) DisablePublicHealthRequestLog() bool {
	return p.p.Bool(ViperKeyDisablePublicHealthRequestLog)
}

func (p *Config) SelfPublicURL() *url.URL {
	return p.baseURL(ViperKeyPublicBaseURL, ViperKeyPublicHost, ViperKeyPublicPort, 4433)
}

func (p *Config) DisableAdminHealthRequestLog() bool {
	return p.p.Bool(ViperKeyDisableAdminHealthRequestLog)
}

func (p *Config) SelfAdminURL() *url.URL {
	return p.baseURL(ViperKeyAdminBaseURL, ViperKeyAdminHost, ViperKeyAdminPort, 4434)
}

func (p *Config) CourierSMTPURL() *url.URL {
	return p.ParseURIOrFail(ViperKeyCourierSMTPURL)
}

func (p *Config) SelfServiceFlowLoginUI() *url.URL {
	return p.ParseAbsoluteOrRelativeURIOrFail(ViperKeySelfServiceLoginUI)
}

func (p *Config) SelfServiceFlowSettingsUI() *url.URL {
	return p.ParseAbsoluteOrRelativeURIOrFail(ViperKeySelfServiceSettingsURL)
}

func (p *Config) SelfServiceFlowErrorURL() *url.URL {
	return p.ParseAbsoluteOrRelativeURIOrFail(ViperKeySelfServiceErrorUI)
}

func (p *Config) SelfServiceFlowRegistrationUI() *url.URL {
	return p.ParseAbsoluteOrRelativeURIOrFail(ViperKeySelfServiceRegistrationUI)
}

func (p *Config) SelfServiceFlowRecoveryUI() *url.URL {
	return p.ParseAbsoluteOrRelativeURIOrFail(ViperKeySelfServiceRecoveryUI)
}

// SessionLifespan returns nil when the value is not set.
func (p *Config) SessionLifespan() time.Duration {
	return p.p.DurationF(ViperKeySessionLifespan, time.Hour*24)
}

func (p *Config) SessionPersistentCookie() bool {
	return p.p.Bool(ViperKeySessionPersistentCookie)
}

func (p *Config) SelfServiceBrowserWhitelistedReturnToDomains() (us []url.URL) {
	src := p.p.Strings(ViperKeyURLsWhitelistedReturnToDomains)
	for k, u := range src {
		if len(u) == 0 {
			continue
		}

		parsed, err := url.ParseRequestURI(u)
		if err != nil {
			p.l.WithError(err).Warnf("Ignoring URL \"%s\" from configuration key \"%s.%d\".", u, ViperKeyURLsWhitelistedReturnToDomains, k)
			continue
		}
		if parsed.Host == "*" {
			p.l.Warnf("Ignoring wildcard \"%s\" from configuration key \"%s.%d\".", u, ViperKeyURLsWhitelistedReturnToDomains, k)
			continue
		}
		eTLD, icann := publicsuffix.PublicSuffix(parsed.Host)
		if len(parsed.Host) > 0 &&
			parsed.Host[:1] == "*" &&
			icann &&
			parsed.Host == fmt.Sprintf("*.%s", eTLD) {
			p.l.Warnf("Ignoring wildcard \"%s\" from configuration key \"%s.%d\".", u, ViperKeyURLsWhitelistedReturnToDomains, k)
			continue
		}

		us = append(us, *parsed)
	}

	return us
}

func (p *Config) SelfServiceFlowLoginRequestLifespan() time.Duration {
	return p.p.DurationF(ViperKeySelfServiceLoginRequestLifespan, time.Hour)
}

func (p *Config) SelfServiceFlowSettingsFlowLifespan() time.Duration {
	return p.p.DurationF(ViperKeySelfServiceSettingsRequestLifespan, time.Hour)
}

func (p *Config) SelfServiceFlowRegistrationRequestLifespan() time.Duration {
	return p.p.DurationF(ViperKeySelfServiceRegistrationRequestLifespan, time.Hour)
}

func (p *Config) SelfServiceFlowLogoutRedirectURL() *url.URL {
	return p.p.RequestURIF(ViperKeySelfServiceLogoutBrowserDefaultReturnTo, p.SelfServiceBrowserDefaultReturnTo())
}

func (p *Config) CourierSMTPFrom() string {
	return p.p.StringF(ViperKeyCourierSMTPFrom, "noreply@kratos.ory.sh")
}

func (p *Config) CourierSMTPFromName() string {
	return p.p.StringF(ViperKeyCourierSMTPFromName, "")
}

func (p *Config) CourierTemplatesRoot() string {
	return p.p.StringF(ViperKeyCourierTemplatesPath, "courier/builtin/templates")
}

func (p *Config) CourierSMTPHeaders() map[string]string {
	return p.p.StringMap(ViperKeyCourierSMTPHeaders)
}

func splitUrlAndFragment(s string) (string, string) {
	i := strings.IndexByte(s, '#')
	if i < 0 {
		return s, ""
	}
	return s[:i], s[i+1:]
}

func (p *Config) ParseAbsoluteOrRelativeURIOrFail(key string) *url.URL {
	u, frag := splitUrlAndFragment(p.p.String(key))
	parsed, err := url.ParseRequestURI(u)
	if err != nil {
		p.l.WithError(errors.WithStack(err)).
			Fatalf("Configuration value from key %s is not a valid URL: %s", key, p.p.String(key))
	}

	if frag != "" {
		parsed.Fragment = frag
	}
	return parsed
}

func (p *Config) ParseURIOrFail(key string) *url.URL {
	parsed := p.ParseAbsoluteOrRelativeURIOrFail(key)
	if parsed.Scheme == "" {
		p.l.WithField("reason", "expected scheme to be set").
			Fatalf("Configuration value from key %s is not a valid URL: %s", key, p.p.String(key))
	}
	return parsed
}

func (p *Config) Tracing() *tracing.Config {
	return p.p.TracingConfig("Ory Kratos")
}

func (p *Config) IsInsecureDevMode() bool {
	return p.Source().Bool("dev")
}

func (p *Config) IsBackgroundCourierEnabled() bool {
	return p.Source().Bool("watch-courier")
}

func (p *Config) CourierExposeMetricsPort() int {
	return p.Source().Int("expose-metrics-port")
}

func (p *Config) MetricsListenOn() string {
	return strings.Replace(p.AdminListenOn(), ":4434", fmt.Sprintf(":%d", p.CourierExposeMetricsPort()), 1)
}

func (p *Config) SelfServiceFlowVerificationUI() *url.URL {
	return p.ParseAbsoluteOrRelativeURIOrFail(ViperKeySelfServiceVerificationUI)
}

func (p *Config) SelfServiceFlowVerificationRequestLifespan() time.Duration {
	return p.p.DurationF(ViperKeySelfServiceVerificationRequestLifespan, time.Hour)
}

func (p *Config) SelfServiceFlowVerificationReturnTo(defaultReturnTo *url.URL) *url.URL {
	return p.p.RequestURIF(ViperKeySelfServiceVerificationBrowserDefaultReturnTo, defaultReturnTo)
}

func (p *Config) SelfServiceFlowVerificationAfterHooks(strategy string) []SelfServiceHook {
	return p.selfServiceHooks(HookStrategyKey(ViperKeySelfServiceVerificationAfter, strategy))
}

func (p *Config) SelfServiceFlowRecoveryReturnTo() *url.URL {
	return p.p.RequestURIF(ViperKeySelfServiceRecoveryBrowserDefaultReturnTo, p.SelfServiceBrowserDefaultReturnTo())
}

func (p *Config) SelfServiceFlowRecoveryRequestLifespan() time.Duration {
	return p.p.DurationF(ViperKeySelfServiceRecoveryRequestLifespan, time.Hour)
}

func (p *Config) SelfServiceLinkMethodLifespan() time.Duration {
	return p.p.DurationF(ViperKeyLinkLifespan, time.Hour)
}

func (p *Config) SelfServiceLinkMethodBaseURL() *url.URL {
	return p.p.RequestURIF(ViperKeyLinkBaseURL, p.SelfPublicURL())
}

func (p *Config) SelfServiceFlowRecoveryAfterHooks(strategy string) []SelfServiceHook {
	return p.selfServiceHooks(HookStrategyKey(ViperKeySelfServiceRecoveryAfter, strategy))
}

func (p *Config) SelfServiceFlowSettingsPrivilegedSessionMaxAge() time.Duration {
	return p.p.DurationF(ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, time.Hour)
}

func (p *Config) SessionSameSiteMode() http.SameSite {
	if !p.p.Exists(ViperKeySessionSameSite) {
		return p.CookieSameSiteMode()
	}

	switch p.p.StringF(ViperKeySessionSameSite, "Lax") {
	case "Lax":
		return http.SameSiteLaxMode
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	}
	return http.SameSiteDefaultMode
}

func (p *Config) SessionDomain() string {
	if !p.p.Exists(ViperKeySessionDomain) {
		return p.CookieDomain()
	}
	return p.p.String(ViperKeySessionDomain)
}

func (p *Config) CookieDomain() string {
	return p.p.String(ViperKeyCookieDomain)
}

func (p *Config) SessionWhoAmIAAL() string {
	return p.p.String(ViperKeySessionWhoAmIAAL)
}

func (p *Config) SelfServiceSettingsRequiredAAL() string {
	return p.p.String(ViperKeySelfServiceSettingsRequiredAAL)
}

func (p *Config) CookieSameSiteMode() http.SameSite {
	switch p.p.StringF(ViperKeyCookieSameSite, "Lax") {
	case "Lax":
		return http.SameSiteLaxMode
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	}
	return http.SameSiteDefaultMode
}

func (p *Config) SessionPath() string {
	if !p.p.Exists(ViperKeySessionPath) {
		return p.CookiePath()
	}
	return p.p.String(ViperKeySessionPath)
}

func (p *Config) CookiePath() string {
	return p.p.String(ViperKeyCookiePath)
}

func (p *Config) SelfServiceFlowLoginReturnTo(strategy string) *url.URL {
	return p.selfServiceReturnTo(ViperKeySelfServiceLoginAfter, strategy)
}

func (p *Config) SelfServiceFlowRegistrationReturnTo(strategy string) *url.URL {
	return p.selfServiceReturnTo(ViperKeySelfServiceRegistrationAfter, strategy)
}

func (p *Config) SelfServiceFlowSettingsReturnTo(strategy string, defaultReturnTo *url.URL) *url.URL {
	return p.p.RequestURIF(
		ViperKeySelfServiceSettingsAfter+"."+strategy+"."+DefaultBrowserReturnURL,
		p.p.RequestURIF(ViperKeySelfServiceSettingsAfter+"."+DefaultBrowserReturnURL,
			defaultReturnTo,
		),
	)
}

func (p *Config) selfServiceReturnTo(key string, strategy string) *url.URL {
	return p.p.RequestURIF(
		key+"."+strategy+"."+DefaultBrowserReturnURL,
		p.p.RequestURIF(key+"."+DefaultBrowserReturnURL,
			p.SelfServiceBrowserDefaultReturnTo(),
		),
	)
}

func (p *Config) ConfigVersion() string {
	return p.p.StringF(ViperKeyVersion, UnknownVersion)
}

func (p *Config) PasswordPolicyConfig() *PasswordPolicy {
	return &PasswordPolicy{
		HaveIBeenPwnedHost:    p.p.StringF(ViperKeyPasswordHaveIBeenPwnedHost, "api.pwnedpasswords.com"),
		HaveIBeenPwnedEnabled: p.p.BoolF(ViperKeyPasswordHaveIBeenPwnedEnabled, true),
		MaxBreaches:           uint(p.p.Int(ViperKeyPasswordMaxBreaches)),
		IgnoreNetworkErrors:   p.p.BoolF(ViperKeyIgnoreNetworkErrors, true),
	}
}

func (p *Config) WebAuthnConfig() *webauthn.Config {
	return &webauthn.Config{
		RPDisplayName: p.p.String(ViperKeyWebAuthnRPDisplayName),
		RPID:          p.p.String(ViperKeyWebAuthnRPID),
		RPOrigin:      p.p.String(ViperKeyWebAuthnRPOrigin),
		RPIcon:        p.p.String(ViperKeyWebAuthnRPIcon),
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			UserVerification: protocol.VerificationDiscouraged,
		},
	}
}

func (p *Config) HasherPasswordHashingAlgorithm() string {
	configValue := p.p.StringF(ViperKeyHasherAlgorithm, DefaultPasswordHashingAlgorithm)
	switch configValue {
	case "bcrypt":
		return configValue
	case "argon2":
		fallthrough
	default:
		return configValue
	}
}

func (p *Config) CipherAlgorithm() string {
	configValue := p.p.StringF(ViperKeyCipherAlgorithm, DefaultCipherAlgorithm)
	switch configValue {
	case "noop":
		return configValue
	case "xchacha20-poly1305":
		return configValue
	case "aes":
		fallthrough
	default:
		return configValue

	}
}

func (p *Config) GetTSLCertificatesForPublic() []tls.Certificate {
	return p.getTSLCertificates(
		"public",
		p.p.String(ViperKeyPublicTLSCertBase64),
		p.p.String(ViperKeyPublicTLSKeyBase64),
		p.p.String(ViperKeyPublicTLSCertPath),
		p.p.String(ViperKeyPublicTLSKeyPath),
	)
}

func (p *Config) GetTSLCertificatesForAdmin() []tls.Certificate {
	return p.getTSLCertificates(
		"admin",
		p.p.String(ViperKeyAdminTLSCertBase64),
		p.p.String(ViperKeyAdminTLSKeyBase64),
		p.p.String(ViperKeyAdminTLSCertPath),
		p.p.String(ViperKeyAdminTLSKeyPath),
	)
}

func (p *Config) getTSLCertificates(daemon, certBase64, keyBase64, certPath, keyPath string) []tls.Certificate {
	cert, err := tlsx.Certificate(certBase64, keyBase64, certPath, keyPath)

	if err == nil {
		p.l.Infof("Setting up HTTPS for %s", daemon)
		return cert
	} else if !errors.Is(err, tlsx.ErrNoCertificatesConfigured) {
		p.l.WithError(err).Fatalf("Unable to load HTTPS TLS Certificate")
	}

	p.l.Infof("TLS has not been configured for %s, skipping", daemon)
	return nil
}

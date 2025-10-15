// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	"cmp"
	"context"
	"crypto/sha512"
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

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid"
	"github.com/inhies/go-bytesize"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
	"golang.org/x/net/publicsuffix"

	"github.com/ory/kratos/x"

	"github.com/ory/herodot"
	"github.com/ory/jsonschema/v3"
	"github.com/ory/jsonschema/v3/httploader"
	"github.com/ory/kratos/embedx"
	"github.com/ory/kratos/request"
	"github.com/ory/x/configx"
	"github.com/ory/x/contextx"
	"github.com/ory/x/crdbx"
	"github.com/ory/x/httpx"
	"github.com/ory/x/jsonschemax"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/pointerx"
	"github.com/ory/x/stringsx"
	"github.com/ory/x/watcherx"
)

const (
	DefaultIdentityTraitsSchemaID                            = "default"
	DefaultBrowserReturnURL                                  = "default_browser_return_url"
	DefaultSQLiteMemoryDSN                                   = "sqlite://file::memory:?_fk=true&cache=shared"
	DefaultPasswordHashingAlgorithm                          = "argon2"
	DefaultCipherAlgorithm                                   = "noop"
	UnknownVersion                                           = "unknown version"
	ViperKeyDSN                                              = "dsn"
	ViperKeyCourierSMTPURL                                   = "courier.smtp.connection_uri"
	ViperKeyCourierSMTPClientCertPath                        = "courier.smtp.client_cert_path"
	ViperKeyCourierSMTPClientKeyPath                         = "courier.smtp.client_key_path"
	ViperKeyCourierTemplatesPath                             = "courier.template_override_path"
	ViperKeyCourierTemplatesRecoveryInvalidEmail             = "courier.templates.recovery.invalid.email"
	ViperKeyCourierTemplatesRecoveryValidEmail               = "courier.templates.recovery.valid.email"
	ViperKeyCourierTemplatesRecoveryCodeInvalidEmail         = "courier.templates.recovery_code.invalid.email"
	ViperKeyCourierTemplatesRecoveryCodeValidEmail           = "courier.templates.recovery_code.valid.email"
	ViperKeyCourierTemplatesVerificationInvalidEmail         = "courier.templates.verification.invalid.email"
	ViperKeyCourierTemplatesVerificationValidEmail           = "courier.templates.verification.valid.email"
	ViperKeyCourierTemplatesVerificationCodeInvalidEmail     = "courier.templates.verification_code.invalid.email"
	ViperKeyCourierTemplatesVerificationCodeValidEmail       = "courier.templates.verification_code.valid.email"
	ViperKeyCourierTemplatesVerificationCodeValidSMS         = "courier.templates.verification_code.valid.sms"
	ViperKeyCourierTemplatesRecoveryCodeValidSMS             = "courier.templates.recovery_code.valid.sms"
	ViperKeyCourierTemplatesLoginCodeValidSMS                = "courier.templates.login_code.valid.sms"
	ViperKeyCourierTemplatesRegistrationCodeValidSMS         = "courier.templates.registration_code.valid.sms"
	ViperKeyCourierDeliveryStrategy                          = "courier.delivery_strategy"
	ViperKeyCourierHTTPRequestConfig                         = "courier.http.request_config"
	ViperKeyCourierTemplatesLoginCodeValidEmail              = "courier.templates.login_code.valid.email"
	ViperKeyCourierTemplatesRegistrationCodeValidEmail       = "courier.templates.registration_code.valid.email"
	ViperKeyCourierSMTP                                      = "courier.smtp"
	ViperKeyCourierSMTPFrom                                  = "courier.smtp.from_address"
	ViperKeyCourierSMTPFromName                              = "courier.smtp.from_name"
	ViperKeyCourierSMTPHeaders                               = "courier.smtp.headers"
	ViperKeyCourierSMTPLocalName                             = "courier.smtp.local_name"
	ViperKeyCourierMessageRetries                            = "courier.message_retries"
	ViperKeyCourierWorkerPullCount                           = "courier.worker.pull_count"
	ViperKeyCourierWorkerPullWait                            = "courier.worker.pull_wait"
	ViperKeyCourierChannels                                  = "courier.channels"
	ViperKeySecretsDefault                                   = "secrets.default"
	ViperKeySecretsCookie                                    = "secrets.cookie"
	ViperKeySecretsCipher                                    = "secrets.cipher"
	ViperKeySecretsPagination                                = "secrets.pagination"
	ViperKeyPublicBaseURL                                    = "serve.public.base_url"
	ViperKeyAdminBaseURL                                     = "serve.admin.base_url"
	ViperKeySessionLifespan                                  = "session.lifespan"
	ViperKeySessionSameSite                                  = "session.cookie.same_site"
	ViperKeySessionSecure                                    = "session.cookie.secure"
	ViperKeySessionDomain                                    = "session.cookie.domain"
	ViperKeySessionName                                      = "session.cookie.name"
	ViperKeySessionPath                                      = "session.cookie.path"
	ViperKeySessionPersistentCookie                          = "session.cookie.persistent"
	ViperKeySessionTokenizerTemplates                        = "session.whoami.tokenizer.templates"
	ViperKeySessionWhoAmIAAL                                 = "session.whoami.required_aal"
	ViperKeySessionWhoAmICaching                             = "feature_flags.cacheable_sessions"
	ViperKeyFeatureFlagFasterSessionExtend                   = "feature_flags.faster_session_extend"
	ViperKeySessionWhoAmICachingMaxAge                       = "feature_flags.cacheable_sessions_max_age"
	ViperKeyUseContinueWithTransitions                       = "feature_flags.use_continue_with_transitions"
	ViperKeyChooseRecoveryAddress                            = "feature_flags.choose_recovery_address"
	ViperKeyUseLegacyShowVerificationUI                      = "feature_flags.legacy_continue_with_verification_ui"
	ViperKeyLegacyOIDCRegistrationGroup                      = "feature_flags.legacy_oidc_registration_node_group"
	ViperKeyUseLegacyRequireVerifiedLoginError               = "feature_flags.legacy_require_verified_login_error"
	ViperKeySessionRefreshMinTimeLeft                        = "session.earliest_possible_extend"
	ViperKeyCookieSameSite                                   = "cookies.same_site"
	ViperKeyCookieDomain                                     = "cookies.domain"
	ViperKeyCookiePath                                       = "cookies.path"
	ViperKeyCookieSecure                                     = "cookies.secure"
	ViperKeySelfServiceStrategyConfig                        = "selfservice.methods"
	ViperKeySelfServiceBrowserDefaultReturnTo                = "selfservice." + DefaultBrowserReturnURL
	ViperKeyURLsAllowedReturnToDomains                       = "selfservice.allowed_return_urls"
	ViperKeySelfServiceRegistrationEnabled                   = "selfservice.flows.registration.enabled"
	ViperKeySelfServiceRegistrationLoginHints                = "selfservice.flows.registration.login_hints"
	ViperKeySelfServiceRegistrationEnableLegacyOneStep       = "selfservice.flows.registration.enable_legacy_one_step"
	ViperKeySelfServiceRegistrationFlowStyle                 = "selfservice.flows.registration.style"
	ViperKeySelfServiceRegistrationUI                        = "selfservice.flows.registration.ui_url"
	ViperKeySelfServiceRegistrationRequestLifespan           = "selfservice.flows.registration.lifespan"
	ViperKeySelfServiceRegistrationAfter                     = "selfservice.flows.registration.after"
	ViperKeySelfServiceRegistrationBeforeHooks               = "selfservice.flows.registration.before.hooks"
	ViperKeySelfServiceLoginUI                               = "selfservice.flows.login.ui_url"
	ViperKeySelfServiceLoginFlowStyle                        = "selfservice.flows.login.style"
	ViperKeySecurityAccountEnumerationMitigate               = "security.account_enumeration.mitigate"
	ViperKeySelfServiceLoginRequestLifespan                  = "selfservice.flows.login.lifespan"
	ViperKeySelfServiceLoginAfter                            = "selfservice.flows.login.after"
	ViperKeySelfServiceLoginBeforeHooks                      = "selfservice.flows.login.before.hooks"
	ViperKeySelfServiceErrorUI                               = "selfservice.flows.error.ui_url"
	ViperKeySelfServiceLogoutBrowserDefaultReturnTo          = "selfservice.flows.logout.after." + DefaultBrowserReturnURL
	ViperKeySelfServiceSettingsURL                           = "selfservice.flows.settings.ui_url"
	ViperKeySelfServiceSettingsAfter                         = "selfservice.flows.settings.after"
	ViperKeySelfServiceSettingsBeforeHooks                   = "selfservice.flows.settings.before.hooks"
	ViperKeySelfServiceSettingsRequestLifespan               = "selfservice.flows.settings.lifespan"
	ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter = "selfservice.flows.settings.privileged_session_max_age"
	ViperKeySelfServiceSettingsRequiredAAL                   = "selfservice.flows.settings.required_aal"
	ViperKeySelfServiceRecoveryAfter                         = "selfservice.flows.recovery.after"
	ViperKeySelfServiceRecoveryBeforeHooks                   = "selfservice.flows.recovery.before.hooks"
	ViperKeySelfServiceRecoveryEnabled                       = "selfservice.flows.recovery.enabled"
	ViperKeySelfServiceRecoveryUse                           = "selfservice.flows.recovery.use"
	ViperKeySelfServiceRecoveryUI                            = "selfservice.flows.recovery.ui_url"
	ViperKeySelfServiceRecoveryRequestLifespan               = "selfservice.flows.recovery.lifespan"
	ViperKeySelfServiceRecoveryBrowserDefaultReturnTo        = "selfservice.flows.recovery.after." + DefaultBrowserReturnURL
	ViperKeySelfServiceRecoveryNotifyUnknownRecipients       = "selfservice.flows.recovery.notify_unknown_recipients"
	ViperKeySelfServiceVerificationEnabled                   = "selfservice.flows.verification.enabled"
	ViperKeySelfServiceVerificationUI                        = "selfservice.flows.verification.ui_url"
	ViperKeySelfServiceVerificationRequestLifespan           = "selfservice.flows.verification.lifespan"
	ViperKeySelfServiceVerificationBrowserDefaultReturnTo    = "selfservice.flows.verification.after." + DefaultBrowserReturnURL
	ViperKeySelfServiceVerificationAfter                     = "selfservice.flows.verification.after"
	ViperKeySelfServiceVerificationBeforeHooks               = "selfservice.flows.verification.before.hooks"
	ViperKeySelfServiceVerificationUse                       = "selfservice.flows.verification.use"
	ViperKeySelfServiceVerificationNotifyUnknownRecipients   = "selfservice.flows.verification.notify_unknown_recipients"
	ViperKeyDefaultIdentitySchemaID                          = "identity.default_schema_id"
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
	ViperKeyDatabaseCleanupSleepTables                       = "database.cleanup.sleep.tables"
	ViperKeyDatabaseCleanupBatchSize                         = "database.cleanup.batch_size"
	ViperKeyLinkLifespan                                     = "selfservice.methods.link.config.lifespan"
	ViperKeyCodeLifespan                                     = "selfservice.methods.code.config.lifespan"
	ViperKeyCodeMaxSubmissions                               = "selfservice.methods.code.config.max_submissions"
	ViperKeyCodeConfigMissingCredentialFallbackEnabled       = "selfservice.methods.code.config.missing_credential_fallback_enabled"
	ViperKeyPasswordHaveIBeenPwnedHost                       = "selfservice.methods.password.config.haveibeenpwned_host"
	ViperKeyPasswordHaveIBeenPwnedEnabled                    = "selfservice.methods.password.config.haveibeenpwned_enabled"
	ViperKeyPasswordMaxBreaches                              = "selfservice.methods.password.config.max_breaches"
	ViperKeyPasswordMinLength                                = "selfservice.methods.password.config.min_password_length"
	ViperKeyPasswordIdentifierSimilarityCheckEnabled         = "selfservice.methods.password.config.identifier_similarity_check_enabled"
	ViperKeyIgnoreNetworkErrors                              = "selfservice.methods.password.config.ignore_network_errors"
	ViperKeyPasswordRegistrationProfileGroup                 = "selfservice.methods.password.config.password_profile_registration_node_group"
	ViperKeyTOTPIssuer                                       = "selfservice.methods.totp.config.issuer"
	ViperKeyOIDCBaseRedirectURL                              = "selfservice.methods.oidc.config.base_redirect_uri"
	ViperKeySAMLBaseRedirectURL                              = "selfservice.methods.saml.config.base_redirect_uri"
	ViperKeyWebAuthnRPDisplayName                            = "selfservice.methods.webauthn.config.rp.display_name"
	ViperKeyWebAuthnRPID                                     = "selfservice.methods.webauthn.config.rp.id"
	ViperKeyWebAuthnRPOrigin                                 = "selfservice.methods.webauthn.config.rp.origin"
	ViperKeyWebAuthnRPOrigins                                = "selfservice.methods.webauthn.config.rp.origins"
	ViperKeyWebAuthnPasswordless                             = "selfservice.methods.webauthn.config.passwordless"
	ViperKeyPasskeyEnabled                                   = "selfservice.methods.passkey.enabled"
	ViperKeyPasskeyRPDisplayName                             = "selfservice.methods.passkey.config.rp.display_name"
	ViperKeyPasskeyRPID                                      = "selfservice.methods.passkey.config.rp.id"
	ViperKeyPasskeyRPOrigins                                 = "selfservice.methods.passkey.config.rp.origins"
	ViperKeyOAuth2ProviderURL                                = "oauth2_provider.url"
	ViperKeyOAuth2ProviderHeader                             = "oauth2_provider.headers"
	ViperKeyOAuth2ProviderOverrideReturnTo                   = "oauth2_provider.override_return_to"
	ViperKeyClientHTTPNoPrivateIPRanges                      = "clients.http.disallow_private_ip_ranges"
	ViperKeyClientHTTPPrivateIPExceptionURLs                 = "clients.http.private_ip_exception_urls"
	ViperKeyWebhookHeaderAllowlist                           = "clients.web_hook.header_allowlist"
	ViperKeyPreviewDefaultReadConsistencyLevel               = "preview.default_read_consistency_level"
	ViperKeyVersion                                          = "version"
	ViperKeyPasswordMigrationHook                            = "selfservice.methods.password.config.migrate_hook"
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
	SelfServiceStrategyCode struct {
		*SelfServiceStrategy
		PasswordlessEnabled bool `json:"passwordless_enabled"`
		MFAEnabled          bool `json:"mfa_enabled"`
	}
	Schema struct {
		ID                    string `json:"id" koanf:"id"`
		URL                   string `json:"url" koanf:"url"`
		SelfserviceSelectable bool   `json:"selfservice_selectable" koanf:"selfservice_selectable"`
	}
	PasswordPolicy struct {
		HaveIBeenPwnedHost               string `json:"haveibeenpwned_host"`
		HaveIBeenPwnedEnabled            bool   `json:"haveibeenpwned_enabled"`
		MaxBreaches                      uint   `json:"max_breaches"`
		IgnoreNetworkErrors              bool   `json:"ignore_network_errors"`
		MinPasswordLength                uint   `json:"min_password_length"`
		IdentifierSimilarityCheckEnabled bool   `json:"identifier_similarity_check_enabled"`
	}
	Schemas                  []Schema
	CourierEmailBodyTemplate struct {
		PlainText string `json:"plaintext"`
		HTML      string `json:"html"`
	}
	CourierEmailTemplate struct {
		Body    *CourierEmailBodyTemplate `json:"body"`
		Subject string                    `json:"subject"`
	}
	CourierSMSTemplate struct {
		Body *CourierSMSTemplateBody `json:"body"`
	}
	CourierSMSTemplateBody struct {
		PlainText string `json:"plaintext"`
	}
	CourierChannel struct {
		ID            string         `json:"id" koanf:"id"`
		Type          string         `json:"type" koanf:"type"`
		SMTPConfig    *SMTPConfig    `json:"smtp_config" koanf:"smtp_config"`
		RequestConfig request.Config `json:"request_config" koanf:"request_config"`
	}
	SMTPConfig struct {
		ConnectionURI  string            `json:"connection_uri" koanf:"connection_uri"`
		ClientCertPath string            `json:"client_cert_path" koanf:"client_cert_path"`
		ClientKeyPath  string            `json:"client_key_path" koanf:"client_key_path"`
		FromAddress    string            `json:"from_address" koanf:"from_address"`
		FromName       string            `json:"from_name" koanf:"from_name"`
		Headers        map[string]string `json:"headers" koanf:"headers"`
		LocalName      string            `json:"local_name" koanf:"local_name"`
	}
	PasswordMigrationHook struct {
		Enabled bool           `json:"enabled" koanf:"enabled"`
		Config  request.Config `json:"config" koanf:"config"`
	}
	Config struct {
		l                  *logrusx.Logger
		p                  *configx.Provider
		c                  contextx.Contextualizer
		identityMetaSchema *jsonschema.Schema
		stdOutOrErr        io.Writer
	}
	Provider interface {
		Config() *Config
	}
	CourierConfigs interface {
		CourierTemplatesRoot(ctx context.Context) string
		CourierTemplatesVerificationInvalid(ctx context.Context) *CourierEmailTemplate
		CourierTemplatesVerificationValid(ctx context.Context) *CourierEmailTemplate
		CourierTemplatesRecoveryInvalid(ctx context.Context) *CourierEmailTemplate
		CourierTemplatesRecoveryValid(ctx context.Context) *CourierEmailTemplate
		CourierTemplatesRecoveryCodeInvalid(ctx context.Context) *CourierEmailTemplate
		CourierTemplatesRecoveryCodeValid(ctx context.Context) *CourierEmailTemplate
		CourierTemplatesVerificationCodeInvalid(ctx context.Context) *CourierEmailTemplate
		CourierTemplatesVerificationCodeValid(ctx context.Context) *CourierEmailTemplate
		CourierTemplatesLoginCodeValid(ctx context.Context) *CourierEmailTemplate
		CourierTemplatesRegistrationCodeValid(ctx context.Context) *CourierEmailTemplate
		CourierSMSTemplatesVerificationCodeValid(ctx context.Context) *CourierSMSTemplate
		CourierSMSTemplatesRecoveryCodeValid(ctx context.Context) *CourierSMSTemplate
		CourierSMSTemplatesLoginCodeValid(ctx context.Context) *CourierSMSTemplate
		CourierSMSTemplatesRegistrationCodeValid(ctx context.Context) *CourierSMSTemplate
		CourierMessageRetries(ctx context.Context) int
		CourierWorkerPullCount(ctx context.Context) int
		CourierWorkerPullWait(ctx context.Context) time.Duration
		CourierChannels(context.Context) ([]*CourierChannel, error)
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

	return nil, errors.Errorf("unable to find identity schema with id: %s", id)
}

func MustNew(t testing.TB, l *logrusx.Logger, ctxer contextx.Contextualizer, opts ...configx.OptionModifier) *Config {
	p, err := New(t.Context(), l, os.Stderr, ctxer, opts...)
	require.NoError(t, err)
	return p
}

func New(ctx context.Context, l *logrusx.Logger, stdOutOrErr io.Writer, ctxer contextx.Contextualizer, opts ...configx.OptionModifier) (*Config, error) {
	var c *Config

	opts = append([]configx.OptionModifier{
		configx.WithStderrValidationReporter(),
		configx.OmitKeysFromTracing("dsn", "courier.smtp.connection_uri", "secrets.default", "secrets.cookie", "secrets.cipher", "client_secret"),
		configx.WithImmutables("serve", "profiling", "log"),
		configx.WithExceptImmutables("serve.public.cors.allowed_origins"),
		configx.WithLogrusWatcher(l),
		configx.WithLogger(l),
		configx.WithContext(ctx),
		configx.AttachWatcher(func(event watcherx.Event, err error) {
			if c == nil {
				panic(errors.New("the config provider did not initialise correctly in time"))
			}
			if err := c.validateIdentitySchemas(ctx); err != nil {
				l.WithError(err).
					Errorf("The changed identity schema configuration is invalid and could not be loaded. Rolling back to the last working configuration revision. Please address the validation errors before restarting the process.")
			}
		}),
	}, opts...)

	p, err := configx.New(ctx, embedx.ConfigSchema, opts...)
	if err != nil {
		return nil, err
	}

	l.UseConfig(p)

	c = NewCustom(l, p, stdOutOrErr, ctxer)

	if !p.SkipValidation() {
		if err := c.validateIdentitySchemas(ctx); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func NewCustom(l *logrusx.Logger, p *configx.Provider, stdOutOrErr io.Writer, ctxt contextx.Contextualizer) *Config {
	l.UseConfig(p)
	return &Config{l: l, p: p, c: ctxt, stdOutOrErr: stdOutOrErr}
}

func (p *Config) getIdentitySchemaValidator(ctx context.Context) (*jsonschema.Schema, error) {
	if p.identityMetaSchema == nil {
		c := jsonschema.NewCompiler()
		err := embedx.AddSchemaResources(c, embedx.IdentityMeta)
		if err != nil {
			return nil, err
		}
		p.identityMetaSchema, err = c.Compile(ctx, embedx.IdentityMeta.GetSchemaID())
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return p.identityMetaSchema, nil
}

type validateIdentitySchemasContextKey int

const validateIdentitySchemasClientKey validateIdentitySchemasContextKey = 1

func SetValidateIdentitySchemaResilientClientOptions(ctx context.Context, options []httpx.ResilientOptions) context.Context {
	return context.WithValue(ctx, validateIdentitySchemasClientKey, options)
}

func (p *Config) validateIdentitySchemas(ctx context.Context) error {
	opts := []httpx.ResilientOptions{
		httpx.ResilientClientWithLogger(p.l),
		httpx.ResilientClientWithMaxRetry(2),
		httpx.ResilientClientWithConnectionTimeout(30 * time.Second),
		// Tracing still works correctly even though we pass a no-op tracer
		// here, because the otelhttp package will preferentially use the
		// tracer from the incoming request context over this one.
		httpx.ResilientClientWithTracer(noop.NewTracerProvider().Tracer("github.com/ory/kratos/driver/config")),
	}

	if o, ok := ctx.Value(validateIdentitySchemasClientKey).([]httpx.ResilientOptions); ok {
		opts = o
	}

	if p.ClientHTTPNoPrivateIPRanges(ctx) {
		opts = append(opts, httpx.ResilientClientDisallowInternalIPs())
	}

	ctx = context.WithValue(ctx, httploader.ContextKey, httpx.NewResilientClient(opts...))

	j, err := p.getIdentitySchemaValidator(ctx)
	if err != nil {
		return err
	}

	ss, err := p.IdentityTraitsSchemas(ctx)
	if err != nil {
		return err
	}

	for _, s := range ss {
		resource, err := jsonschema.LoadURL(ctx, s.URL)
		if err != nil {
			return errors.WithStack(err)
		}
		defer func() { _ = resource.Close() }()

		schema, err := io.ReadAll(io.LimitReader(resource, 1024*1024))
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

func (p *Config) ServePublic(ctx context.Context) *configx.Serve {
	return p.GetProvider(ctx).Serve("serve.public", p.IsInsecureDevMode(ctx), configx.Serve{
		Port: 4433,
	})
}

func (p *Config) ServeAdmin(ctx context.Context) *configx.Serve {
	return p.GetProvider(ctx).Serve("serve.admin", p.IsInsecureDevMode(ctx), configx.Serve{
		Port: 4434,
	})
}

func (p *Config) CORSPublic(ctx context.Context) (cors.Options, bool) {
	return p.GetProvider(ctx).CORS("serve.public", cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Cookie"},
		ExposedHeaders:   []string{"Content-Type", "Set-Cookie"},
		AllowCredentials: true,
	})
}

// Deprecated: use context-based [contextx.WithConfigValue] instead.
func (p *Config) Set(_ context.Context, key string, value interface{}) error {
	return p.p.Set(key, value)
}

// Deprecated: use context-based [contextx.WithConfigValue] instead.
func (p *Config) MustSet(_ context.Context, key string, value interface{}) {
	if err := p.p.Set(key, value); err != nil {
		p.l.WithError(err).Fatalf("Unable to set %q to %q.", key, value)
	}
}

func (p *Config) SessionName(ctx context.Context) string {
	return stringsx.Coalesce(p.GetProvider(ctx).String(ViperKeySessionName), DefaultSessionCookieName)
}

func (p *Config) HasherArgon2(ctx context.Context) *Argon2 {
	// warn about usage of default values and point to the docs
	// warning will require https://github.com/ory/viper/issues/19
	return &Argon2{
		Memory: p.GetProvider(ctx).ByteSizeF(ViperKeyHasherArgon2ConfigMemory, Argon2DefaultMemory),
		//nolint:gosec // disable G115
		Iterations: uint32(p.GetProvider(ctx).IntF(ViperKeyHasherArgon2ConfigIterations, int(Argon2DefaultIterations))),
		//nolint:gosec // disable G115
		Parallelism: uint8(p.GetProvider(ctx).IntF(ViperKeyHasherArgon2ConfigParallelism, int(Argon2DefaultParallelism))),
		//nolint:gosec // disable G115
		SaltLength: uint32(p.GetProvider(ctx).IntF(ViperKeyHasherArgon2ConfigSaltLength, int(Argon2DefaultSaltLength))),
		//nolint:gosec // disable G115
		KeyLength:         uint32(p.GetProvider(ctx).IntF(ViperKeyHasherArgon2ConfigKeyLength, int(Argon2DefaultKeyLength))),
		ExpectedDuration:  p.GetProvider(ctx).DurationF(ViperKeyHasherArgon2ConfigExpectedDuration, Argon2DefaultDuration),
		ExpectedDeviation: p.GetProvider(ctx).DurationF(ViperKeyHasherArgon2ConfigExpectedDeviation, Argon2DefaultDeviation),
		DedicatedMemory:   p.GetProvider(ctx).ByteSizeF(ViperKeyHasherArgon2ConfigDedicatedMemory, Argon2DefaultDedicatedMemory),
	}
}

func (p *Config) HasherBcrypt(ctx context.Context) *Bcrypt {
	cost := uint32(p.GetProvider(ctx).IntF(ViperKeyHasherBcryptCost, int(BcryptDefaultCost))) // #nosec G115 -- if the user configures a cost > MaxUint32, go falls back to MaxUint32
	if !p.IsInsecureDevMode(ctx) && cost < BcryptDefaultCost {
		cost = BcryptDefaultCost
	}

	return &Bcrypt{Cost: cost}
}

func (p *Config) DefaultIdentityTraitsSchemaURL(ctx context.Context) (*url.URL, error) {
	ss, err := p.IdentityTraitsSchemas(ctx)
	if err != nil {
		return nil, err
	}

	search := p.GetProvider(ctx).String(ViperKeyDefaultIdentitySchemaID)
	found, err := ss.FindSchemaByID(search)
	if err != nil {
		return nil, err
	}

	return p.ParseURI(found.URL)
}

func (p *Config) DefaultIdentityTraitsSchemaID(ctx context.Context) string {
	return p.GetProvider(ctx).String(ViperKeyDefaultIdentitySchemaID)
}

func (p *Config) IdentityTraitsSchemaURL(ctx context.Context, schemaID string) (*url.URL, error) {
	ss, err := p.IdentityTraitsSchemas(ctx)
	if err != nil {
		return nil, err
	}

	found, err := ss.FindSchemaByID(schemaID)
	if err != nil {
		// default to default schema
		search := p.GetProvider(ctx).String(ViperKeyDefaultIdentitySchemaID)
		found, err = ss.FindSchemaByID(search)
		if err != nil {
			return nil, err
		}
	}

	return p.ParseURI(found.URL)
}

func (p *Config) TOTPIssuer(ctx context.Context) string {
	return p.GetProvider(ctx).StringF(ViperKeyTOTPIssuer, p.SelfPublicURL(ctx).Hostname())
}

func (p *Config) OIDCRedirectURIBase(ctx context.Context) *url.URL {
	return p.GetProvider(ctx).URIF(ViperKeyOIDCBaseRedirectURL, p.SelfPublicURL(ctx))
}

func (p *Config) SAMLRedirectURIBase(ctx context.Context) *url.URL {
	return p.GetProvider(ctx).URIF(ViperKeySAMLBaseRedirectURL, p.SelfPublicURL(ctx))
}

func (p *Config) IdentityTraitsSchemas(ctx context.Context) (ss Schemas, err error) {
	if err = p.GetProvider(ctx).Unmarshal(ViperKeyIdentitySchemas, &ss); err != nil {
		return ss, nil
	}

	return ss, nil
}

func (p *Config) DSN(ctx context.Context) string {
	pp := p.GetProvider(ctx)
	dsn := pp.String(ViperKeyDSN)

	if dsn == "memory" {
		return DefaultSQLiteMemoryDSN
	}

	if len(dsn) > 0 {
		return dsn
	}

	p.l.Fatal("dsn must be set")
	return ""
}

func (p *Config) DisableAPIFlowEnforcement(ctx context.Context) bool {
	if p.IsInsecureDevMode(ctx) && os.Getenv("DEV_DISABLE_API_FLOW_ENFORCEMENT") == "true" {
		p.l.Warn("Because \"DEV_DISABLE_API_FLOW_ENFORCEMENT=true\" and the \"--dev\" flag are set, self-service API flows will no longer check if the interaction is actually a browser flow. This is very dangerous as it allows bypassing of anti-CSRF measures, leaving the deployment highly vulnerable. This option should only be used for automated testing and never come close to real user data anywhere.")
		return true
	}
	return false
}

func (p *Config) ClientHTTPNoPrivateIPRanges(ctx context.Context) bool {
	return p.GetProvider(ctx).Bool(ViperKeyClientHTTPNoPrivateIPRanges)
}

func (p *Config) ClientHTTPPrivateIPExceptionURLs(ctx context.Context) []string {
	return p.GetProvider(ctx).Strings(ViperKeyClientHTTPPrivateIPExceptionURLs)
}

func (p *Config) SelfServiceFlowRegistrationEnabled(ctx context.Context) bool {
	return p.GetProvider(ctx).Bool(ViperKeySelfServiceRegistrationEnabled)
}

func (p *Config) SelfServiceFlowRegistrationLoginHints(ctx context.Context) bool {
	return p.GetProvider(ctx).Bool(ViperKeySelfServiceRegistrationLoginHints)
}

func (p *Config) SelfServiceFlowRegistrationPasswordMethodProfileGroup(ctx context.Context) string {
	switch g := p.GetProvider(ctx).String(ViperKeyPasswordRegistrationProfileGroup); g {
	case "password":
		return "password"
	default:
		return "default"
	}
}

func (p *Config) SelfServiceLegacyOIDCRegistrationGroup(ctx context.Context) bool {
	return p.GetProvider(ctx).Bool(ViperKeyLegacyOIDCRegistrationGroup)
}

func (p *Config) SelfServiceFlowRegistrationTwoSteps(ctx context.Context) bool {
	// The default in previous versions that legacy one-step would be disabled. If legacy is enabled, it means the
	// user has explicitly set the key to true, in which case we respect it.
	if useOneStep := p.GetProvider(ctx).Bool(ViperKeySelfServiceRegistrationEnableLegacyOneStep); useOneStep {
		p.l.Warnf("Found use of deprecated configuration key %q. Please use key %q instead and delete key %[1]q. Will use value from %[1]q to configure registration style.", ViperKeySelfServiceRegistrationEnableLegacyOneStep, ViperKeySelfServiceRegistrationFlowStyle)
		return false
	}

	// In all other cases, we use the new key which (like the old key) defaults to `profile_first` / two-step registration.
	switch style := p.GetProvider(ctx).String(ViperKeySelfServiceRegistrationFlowStyle); style {
	case "profile_first":
		return true
	default:
		return false
	}
}

func (p *Config) SelfServiceFlowIdentitySchema(ctx context.Context, requestedSchema string) (string, error) {
	if requestedSchema == p.GetProvider(ctx).String(ViperKeyDefaultIdentitySchemaID) {
		return requestedSchema, nil
	}
	schemas, err := p.IdentityTraitsSchemas(ctx)
	if err != nil {
		return "", errors.WithStack(err)
	}
	for _, schema := range schemas {
		if schema.ID == requestedSchema {
			if !schema.SelfserviceSelectable {
				return "", errors.WithStack(herodot.ErrBadRequest.WithReasonf("Requested identity schema %q is not enabled for self-service flows.", requestedSchema))
			}
			return requestedSchema, nil
		}
	}
	return "", errors.WithStack(herodot.ErrBadRequest.WithReasonf("Requested identity schema %q does not exist.", requestedSchema))
}

func (p *Config) SelfServiceFlowVerificationEnabled(ctx context.Context) bool {
	return p.GetProvider(ctx).Bool(ViperKeySelfServiceVerificationEnabled)
}

func (p *Config) UseLegacyShowVerificationUI(ctx context.Context) bool {
	return p.GetProvider(ctx).Bool(ViperKeyUseLegacyShowVerificationUI)
}

func (p *Config) UseLegacyRequireVerifiedLoginError(ctx context.Context) bool {
	return p.GetProvider(ctx).Bool(ViperKeyUseLegacyRequireVerifiedLoginError)
}

func (p *Config) SelfServiceFlowRecoveryEnabled(ctx context.Context) bool {
	return p.GetProvider(ctx).Bool(ViperKeySelfServiceRecoveryEnabled)
}

func (p *Config) SelfServiceFlowRecoveryUse(ctx context.Context) string {
	return p.GetProvider(ctx).String(ViperKeySelfServiceRecoveryUse)
}

func (p *Config) SelfServiceFlowLoginBeforeHooks(ctx context.Context) []SelfServiceHook {
	return p.selfServiceHooks(ctx, ViperKeySelfServiceLoginBeforeHooks)
}

func (p *Config) SelfServiceFlowRecoveryBeforeHooks(ctx context.Context) []SelfServiceHook {
	return p.selfServiceHooks(ctx, ViperKeySelfServiceRecoveryBeforeHooks)
}

func (p *Config) SelfServiceFlowVerificationBeforeHooks(ctx context.Context) []SelfServiceHook {
	return p.selfServiceHooks(ctx, ViperKeySelfServiceVerificationBeforeHooks)
}

func (p *Config) SelfServiceFlowVerificationUse(ctx context.Context) string {
	return p.GetProvider(ctx).String(ViperKeySelfServiceVerificationUse)
}

func (p *Config) SelfServiceFlowVerificationNotifyUnknownRecipients(ctx context.Context) bool {
	return p.GetProvider(ctx).BoolF(ViperKeySelfServiceVerificationNotifyUnknownRecipients, false)
}

func (p *Config) SelfServiceFlowSettingsBeforeHooks(ctx context.Context) []SelfServiceHook {
	return p.selfServiceHooks(ctx, ViperKeySelfServiceSettingsBeforeHooks)
}

func (p *Config) SelfServiceFlowRegistrationBeforeHooks(ctx context.Context) []SelfServiceHook {
	hooks := p.selfServiceHooks(ctx, ViperKeySelfServiceRegistrationBeforeHooks)
	if p.SelfServiceFlowRegistrationTwoSteps(ctx) {
		hooks = append(hooks, SelfServiceHook{"two_step_registration", json.RawMessage("{}")})
	}

	return hooks
}

func (p *Config) selfServiceHooks(ctx context.Context, key string) []SelfServiceHook {
	pp := p.GetProvider(ctx)
	val := pp.Get(key)
	if val == nil {
		return []SelfServiceHook{}
	}

	config, err := json.Marshal(val)
	if err != nil {
		p.l.WithError(err).Fatalf("Unable to decode values from configuration key: %s", key)
	}

	var hooks []SelfServiceHook
	if err := json.Unmarshal(config, &hooks); err != nil {
		p.l.WithError(err).Fatalf("Unable to encode value \"%s\" from configuration key: %s", config, key)
	}

	for k := range hooks {
		if len(hooks[k].Config) == 0 {
			hooks[k].Config = json.RawMessage("{}")
		}
	}

	return hooks
}

func (p *Config) SelfServiceFlowLoginAfterHooks(ctx context.Context, strategy string) []SelfServiceHook {
	return p.selfServiceHooks(ctx, HookStrategyKey(ViperKeySelfServiceLoginAfter, strategy))
}

func (p *Config) SelfServiceFlowSettingsAfterHooks(ctx context.Context, strategy string) []SelfServiceHook {
	return p.selfServiceHooks(ctx, HookStrategyKey(ViperKeySelfServiceSettingsAfter, strategy))
}

func (p *Config) SelfServiceFlowRegistrationAfterHooks(ctx context.Context, strategy string) []SelfServiceHook {
	return p.selfServiceHooks(ctx, HookStrategyKey(ViperKeySelfServiceRegistrationAfter, strategy))
}

func (p *Config) SelfServiceStrategy(ctx context.Context, strategy string) *SelfServiceStrategy {
	pp := p.GetProvider(ctx)
	config := json.RawMessage("{}")
	basePath := fmt.Sprintf("%s.%s", ViperKeySelfServiceStrategyConfig, strategy)

	var err error
	config, err = json.Marshal(pp.GetF(basePath+".config", config))
	if err != nil {
		p.l.WithError(err).Warn("Unable to marshal self-service strategy configuration.")
		config = json.RawMessage("{}")
	}

	// The default value can easily be overwritten by setting e.g. `{"selfservice": "null"}` which means that
	// we need to forcibly set these values here:
	defaultEnabled := false
	switch strategy {
	case "identifier_first":
		defaultEnabled = p.SelfServiceLoginFlowIdentifierFirstEnabled(ctx)
	case "code", "password", "profile":
		defaultEnabled = true
	}

	// Backwards compatibility for the old "passwordless_enabled" key
	// This force-enables the code strategy, if passwordless is enabled, because in earlier versions it was possible to
	// disable the code strategy, but enable passwordless
	enabled := pp.BoolF(basePath+".enabled", defaultEnabled)
	if strategy == "code" {
		enabled = enabled || pp.Bool(basePath+".passwordless_enabled")
	}
	return &SelfServiceStrategy{
		Enabled: enabled,
		Config:  config,
	}
}

func (p *Config) SelfServiceCodeStrategy(ctx context.Context) *SelfServiceStrategyCode {
	pp := p.GetProvider(ctx)
	config := json.RawMessage("{}")
	basePath := ViperKeySelfServiceStrategyConfig + ".code"

	var err error
	config, err = json.Marshal(pp.GetF(basePath+".config", config))
	if err != nil {
		p.l.WithError(err).Warn("Unable to marshal self service strategy configuration.")
		config = json.RawMessage("{}")
	}

	return &SelfServiceStrategyCode{
		SelfServiceStrategy: &SelfServiceStrategy{
			Enabled: pp.BoolF(basePath+".enabled", true),
			Config:  config,
		},
		PasswordlessEnabled: pp.BoolF(basePath+".passwordless_enabled", false),
		MFAEnabled:          pp.BoolF(basePath+".mfa_enabled", false),
	}
}

func (p *Config) SecretsDefault(ctx context.Context) [][]byte {
	pp := p.GetProvider(ctx)
	secrets := pp.Strings(ViperKeySecretsDefault)

	if len(secrets) == 0 {
		secrets = []string{uuid.Must(uuid.NewV4()).String()}
		p.MustSet(ctx, ViperKeySecretsDefault, secrets)
	}

	result := make([][]byte, len(secrets))
	for k, v := range secrets {
		result[k] = []byte(v)
	}

	return result
}

func (p *Config) SecretsSession(ctx context.Context) [][]byte {
	secrets := p.GetProvider(ctx).Strings(ViperKeySecretsCookie)
	if len(secrets) == 0 {
		return p.SecretsDefault(ctx)
	}

	result := make([][]byte, len(secrets))
	for k, v := range secrets {
		result[k] = []byte(v)
	}

	return result
}

func (p *Config) SecretsCipher(ctx context.Context) [][32]byte {
	secrets := p.GetProvider(ctx).Strings(ViperKeySecretsCipher)
	return ToCipherSecrets(secrets)
}

func ToCipherSecrets(secrets []string) [][32]byte {
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
			result[n][k] = v
		}
	}
	return result
}

func (p *Config) SecretsPagination(ctx context.Context) [][32]byte {
	secrets := p.GetProvider(ctx).Strings(ViperKeySecretsPagination)

	encryptionKeys := make([][32]byte, len(secrets))
	for i, key := range secrets {
		encryptionKeys[i] = sha512.Sum512_256([]byte(key))
	}

	return encryptionKeys
}

func (p *Config) SelfServiceBrowserDefaultReturnTo(ctx context.Context) *url.URL {
	return p.ParseAbsoluteOrRelativeURIOrFail(ctx, ViperKeySelfServiceBrowserDefaultReturnTo)
}

func (p *Config) SelfPublicURL(ctx context.Context) *url.URL {
	serve := p.ServePublic(ctx)
	return serve.BaseURL
}

func (p *Config) SelfAdminURL(ctx context.Context) *url.URL {
	serve := p.ServeAdmin(ctx)
	return serve.BaseURL
}

func (p *Config) WebhookHeaderAllowlist(ctx context.Context) []string {
	return p.GetProvider(ctx).Strings(ViperKeyWebhookHeaderAllowlist)
}

func (p *Config) OAuth2ProviderHeader(ctx context.Context) http.Header {
	hh := map[string]string{}
	if err := p.GetProvider(ctx).Unmarshal(ViperKeyOAuth2ProviderHeader, &hh); err != nil {
		p.l.WithError(errors.WithStack(err)).
			Errorf("Configuration value from key %s could not be decoded.", ViperKeyOAuth2ProviderHeader)
		return nil
	}

	h := make(http.Header)
	for k, v := range hh {
		h.Set(k, v)
	}

	return h
}

func (p *Config) OAuth2ProviderOverrideReturnTo(ctx context.Context) bool {
	return p.GetProvider(ctx).Bool(ViperKeyOAuth2ProviderOverrideReturnTo)
}

func (p *Config) OAuth2ProviderURL(ctx context.Context) *url.URL {
	k := ViperKeyOAuth2ProviderURL
	v := p.GetProvider(ctx).String(k)
	if v == "" {
		return nil
	}
	parsed, err := p.ParseAbsoluteOrRelativeURI(v)
	if err != nil {
		p.l.WithError(errors.WithStack(err)).
			Errorf("Configuration value from key %s is not a valid URL: %s", k, v)
		return nil
	}
	return parsed
}

func (p *Config) SelfServiceFlowLoginUI(ctx context.Context) *url.URL {
	return p.ParseAbsoluteOrRelativeURIOrFail(ctx, ViperKeySelfServiceLoginUI)
}

func (p *Config) SelfServiceFlowSettingsUI(ctx context.Context) *url.URL {
	return p.ParseAbsoluteOrRelativeURIOrFail(ctx, ViperKeySelfServiceSettingsURL)
}

func (p *Config) SelfServiceFlowErrorURL(ctx context.Context) *url.URL {
	return p.ParseAbsoluteOrRelativeURIOrFail(ctx, ViperKeySelfServiceErrorUI)
}

func (p *Config) SelfServiceFlowRegistrationUI(ctx context.Context) *url.URL {
	return p.ParseAbsoluteOrRelativeURIOrFail(ctx, ViperKeySelfServiceRegistrationUI)
}

func (p *Config) SelfServiceFlowRecoveryUI(ctx context.Context) *url.URL {
	return p.ParseAbsoluteOrRelativeURIOrFail(ctx, ViperKeySelfServiceRecoveryUI)
}

// SessionLifespan returns time.Hour*24 when the value is not set.
func (p *Config) SessionLifespan(ctx context.Context) time.Duration {
	return p.GetProvider(ctx).DurationF(ViperKeySessionLifespan, time.Hour*24)
}

func (p *Config) SessionPersistentCookie(ctx context.Context) bool {
	return p.GetProvider(ctx).Bool(ViperKeySessionPersistentCookie)
}

func (p *Config) SelfServiceBrowserAllowedReturnToDomains(ctx context.Context) (us []url.URL) {
	src := p.GetProvider(ctx).Strings(ViperKeyURLsAllowedReturnToDomains)
	for k, u := range src {
		if len(u) == 0 {
			continue
		}

		parsed, err := url.ParseRequestURI(u)
		if err != nil {
			p.l.WithError(err).Warnf("Ignoring URL \"%s\" from configuration key \"%s.%d\".", u, ViperKeyURLsAllowedReturnToDomains, k)
			continue
		}
		if parsed.Host == "*" {
			p.l.Warnf("Ignoring wildcard \"%s\" from configuration key \"%s.%d\".", u, ViperKeyURLsAllowedReturnToDomains, k)
			continue
		}
		eTLD, icann := publicsuffix.PublicSuffix(parsed.Host)
		if len(parsed.Host) > 0 &&
			parsed.Host[:1] == "*" &&
			icann &&
			parsed.Host == fmt.Sprintf("*.%s", eTLD) {
			p.l.Warnf("Ignoring wildcard \"%s\" from configuration key \"%s.%d\".", u, ViperKeyURLsAllowedReturnToDomains, k)
			continue
		}

		us = append(us, *parsed)
	}

	return us
}

func (p *Config) SelfServiceFlowLoginRequestLifespan(ctx context.Context) time.Duration {
	return p.GetProvider(ctx).DurationF(ViperKeySelfServiceLoginRequestLifespan, time.Hour)
}

func (p *Config) SelfServiceFlowSettingsFlowLifespan(ctx context.Context) time.Duration {
	return p.GetProvider(ctx).DurationF(ViperKeySelfServiceSettingsRequestLifespan, time.Hour)
}

func (p *Config) SelfServiceFlowRegistrationRequestLifespan(ctx context.Context) time.Duration {
	return p.GetProvider(ctx).DurationF(ViperKeySelfServiceRegistrationRequestLifespan, time.Hour)
}

func (p *Config) SelfServiceFlowLogoutRedirectURL(ctx context.Context) *url.URL {
	return p.GetProvider(ctx).RequestURIF(ViperKeySelfServiceLogoutBrowserDefaultReturnTo, p.SelfServiceBrowserDefaultReturnTo(ctx))
}

func (p *Config) CourierEmailStrategy(ctx context.Context) string {
	return p.GetProvider(ctx).StringF(ViperKeyCourierDeliveryStrategy, "smtp")
}

func (p *Config) CourierEmailRequestConfig(ctx context.Context) json.RawMessage {
	if p.CourierEmailStrategy(ctx) != "http" {
		return nil
	}

	config, err := json.Marshal(p.GetProvider(ctx).Get(ViperKeyCourierHTTPRequestConfig))
	if err != nil {
		p.l.WithError(err).Warn("Unable to marshal mailer request configuration.")
		return nil
	}

	return config
}

func (p *Config) CourierTemplatesRoot(ctx context.Context) string {
	return p.GetProvider(ctx).StringF(ViperKeyCourierTemplatesPath, "courier/builtin/templates")
}

func (p *Config) CourierEmailTemplatesHelper(ctx context.Context, key string) *CourierEmailTemplate {
	courierTemplate := &CourierEmailTemplate{
		Body: &CourierEmailBodyTemplate{
			PlainText: "",
			HTML:      "",
		},
		Subject: "",
	}

	if !p.GetProvider(ctx).Exists(key) {
		return courierTemplate
	}

	config, err := json.Marshal(p.GetProvider(ctx).Get(key))
	if err != nil {
		p.l.WithError(err).Fatalf("Unable to decode values from %s.", key)
		return courierTemplate
	}

	if err := json.Unmarshal(config, courierTemplate); err != nil {
		p.l.WithError(err).Fatalf("Unable to encode values from %s.", key)
		return courierTemplate
	}
	return courierTemplate
}

func (p *Config) CourierSMSTemplatesHelper(ctx context.Context, key string) *CourierSMSTemplate {
	courierTemplate := &CourierSMSTemplate{
		Body: &CourierSMSTemplateBody{
			PlainText: "",
		},
	}

	if !p.GetProvider(ctx).Exists(key) {
		return courierTemplate
	}

	config, err := json.Marshal(p.GetProvider(ctx).Get(key))
	if err != nil {
		p.l.WithError(err).Fatalf("Unable to decode values from %s.", key)
		return courierTemplate
	}

	if err := json.Unmarshal(config, courierTemplate); err != nil {
		p.l.WithError(err).Fatalf("Unable to encode values from %s.", key)
		return courierTemplate
	}
	return courierTemplate
}

func (p *Config) CourierTemplatesVerificationInvalid(ctx context.Context) *CourierEmailTemplate {
	return p.CourierEmailTemplatesHelper(ctx, ViperKeyCourierTemplatesVerificationInvalidEmail)
}

func (p *Config) CourierTemplatesVerificationValid(ctx context.Context) *CourierEmailTemplate {
	return p.CourierEmailTemplatesHelper(ctx, ViperKeyCourierTemplatesVerificationValidEmail)
}

func (p *Config) CourierTemplatesRecoveryInvalid(ctx context.Context) *CourierEmailTemplate {
	return p.CourierEmailTemplatesHelper(ctx, ViperKeyCourierTemplatesRecoveryInvalidEmail)
}

func (p *Config) CourierTemplatesRecoveryValid(ctx context.Context) *CourierEmailTemplate {
	return p.CourierEmailTemplatesHelper(ctx, ViperKeyCourierTemplatesRecoveryValidEmail)
}

func (p *Config) CourierTemplatesRecoveryCodeInvalid(ctx context.Context) *CourierEmailTemplate {
	return p.CourierEmailTemplatesHelper(ctx, ViperKeyCourierTemplatesRecoveryCodeInvalidEmail)
}

func (p *Config) CourierTemplatesRecoveryCodeValid(ctx context.Context) *CourierEmailTemplate {
	return p.CourierEmailTemplatesHelper(ctx, ViperKeyCourierTemplatesRecoveryCodeValidEmail)
}

func (p *Config) CourierTemplatesVerificationCodeInvalid(ctx context.Context) *CourierEmailTemplate {
	return p.CourierEmailTemplatesHelper(ctx, ViperKeyCourierTemplatesVerificationCodeInvalidEmail)
}

func (p *Config) CourierTemplatesVerificationCodeValid(ctx context.Context) *CourierEmailTemplate {
	return p.CourierEmailTemplatesHelper(ctx, ViperKeyCourierTemplatesVerificationCodeValidEmail)
}

func (p *Config) CourierSMSTemplatesVerificationCodeValid(ctx context.Context) *CourierSMSTemplate {
	return p.CourierSMSTemplatesHelper(ctx, ViperKeyCourierTemplatesVerificationCodeValidSMS)
}

func (p *Config) CourierSMSTemplatesRecoveryCodeValid(ctx context.Context) *CourierSMSTemplate {
	return p.CourierSMSTemplatesHelper(ctx, ViperKeyCourierTemplatesRecoveryCodeValidSMS)
}

func (p *Config) CourierSMSTemplatesLoginCodeValid(ctx context.Context) *CourierSMSTemplate {
	return p.CourierSMSTemplatesHelper(ctx, ViperKeyCourierTemplatesLoginCodeValidSMS)
}

func (p *Config) CourierSMSTemplatesRegistrationCodeValid(ctx context.Context) *CourierSMSTemplate {
	return p.CourierSMSTemplatesHelper(ctx, ViperKeyCourierTemplatesRegistrationCodeValidSMS)
}

func (p *Config) CourierTemplatesLoginCodeValid(ctx context.Context) *CourierEmailTemplate {
	return p.CourierEmailTemplatesHelper(ctx, ViperKeyCourierTemplatesLoginCodeValidEmail)
}

func (p *Config) CourierTemplatesRegistrationCodeValid(ctx context.Context) *CourierEmailTemplate {
	return p.CourierEmailTemplatesHelper(ctx, ViperKeyCourierTemplatesRegistrationCodeValidEmail)
}

func (p *Config) CourierMessageRetries(ctx context.Context) int {
	return p.GetProvider(ctx).IntF(ViperKeyCourierMessageRetries, 5)
}

func (p *Config) CourierWorkerPullCount(ctx context.Context) int {
	return p.GetProvider(ctx).Int(ViperKeyCourierWorkerPullCount)
}

func (p *Config) CourierWorkerPullWait(ctx context.Context) time.Duration {
	return p.GetProvider(ctx).Duration(ViperKeyCourierWorkerPullWait)
}

func (p *Config) CourierSMTPHeaders(ctx context.Context) map[string]string {
	return p.GetProvider(ctx).StringMap(ViperKeyCourierSMTPHeaders)
}

func (p *Config) CourierChannels(ctx context.Context) (ccs []*CourierChannel, _ error) {
	if err := p.GetProvider(ctx).Unmarshal(ViperKeyCourierChannels, &ccs); err != nil {
		return nil, errors.WithStack(err)
	}

	// load legacy configs
	channel := CourierChannel{
		ID:   "email",
		Type: p.CourierEmailStrategy(ctx),
	}
	if channel.Type == "smtp" {
		if err := p.GetProvider(ctx).Unmarshal(ViperKeyCourierSMTP, &channel.SMTPConfig); err != nil {
			return nil, errors.WithStack(err)
		}
	} else {
		if err := p.GetProvider(ctx).Unmarshal(ViperKeyCourierHTTPRequestConfig, &channel.RequestConfig); err != nil {
			return nil, errors.WithStack(err)
		}
	}
	ccs = append(ccs, &channel)
	return ccs, nil
}

func splitUrlAndFragment(s string) (string, string) {
	i := strings.IndexByte(s, '#')
	if i < 0 {
		return s, ""
	}
	return s[:i], s[i+1:]
}

func (p *Config) ParseAbsoluteOrRelativeURIOrFail(ctx context.Context, key string) *url.URL {
	parsed, err := p.ParseAbsoluteOrRelativeURI(p.GetProvider(ctx).String(key))
	if err != nil {
		p.l.WithError(errors.WithStack(err)).
			Fatalf("Configuration value from key %s is not a valid URL: %s", key, p.GetProvider(ctx).String(key))
	}
	return parsed
}

func (p *Config) ParseURIOrFail(ctx context.Context, key string) *url.URL {
	parsed, err := p.ParseURI(p.GetProvider(ctx).String(key))
	if err != nil {
		p.l.WithField("reason", "expected scheme to be set").
			Fatalf("Configuration value from key %s is not a valid URL: %s", key, p.GetProvider(ctx).String(key))
	}
	return parsed
}

func (p *Config) ParseAbsoluteOrRelativeURI(rawUrl string) (*url.URL, error) {
	u, frag := splitUrlAndFragment(rawUrl)
	parsed, err := url.ParseRequestURI(u)
	if err != nil {
		return nil, errors.Wrapf(err, "configuration value not a valid URL: %s", rawUrl)
	}

	if frag != "" {
		parsed.Fragment = frag
	}

	return parsed, nil
}

func (p *Config) ParseURI(rawUrl string) (*url.URL, error) {
	parsed, err := p.ParseAbsoluteOrRelativeURI(rawUrl)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme == "" {
		return nil, errors.Errorf("configuration value is not a valid URL: %s", rawUrl)
	}
	return parsed, nil
}

func (p *Config) Tracing(ctx context.Context) *otelx.Config {
	return p.GetProvider(ctx).TracingConfig("Ory Kratos")
}

func (p *Config) IsInsecureDevMode(ctx context.Context) bool {
	return p.GetProvider(ctx).Bool("dev")
}

func (p *Config) IsBackgroundCourierEnabled(ctx context.Context) bool {
	return p.GetProvider(ctx).Bool("watch-courier")
}

func (p *Config) CourierExposeMetricsPort(ctx context.Context) int {
	return p.GetProvider(ctx).Int("expose-metrics-port")
}

func (p *Config) SelfServiceFlowVerificationUI(ctx context.Context) *url.URL {
	return p.ParseAbsoluteOrRelativeURIOrFail(ctx, ViperKeySelfServiceVerificationUI)
}

func (p *Config) SelfServiceFlowVerificationRequestLifespan(ctx context.Context) time.Duration {
	return p.GetProvider(ctx).DurationF(ViperKeySelfServiceVerificationRequestLifespan, time.Hour)
}

func (p *Config) SelfServiceFlowVerificationReturnTo(ctx context.Context, defaultReturnTo *url.URL) *url.URL {
	return p.GetProvider(ctx).RequestURIF(ViperKeySelfServiceVerificationBrowserDefaultReturnTo, defaultReturnTo)
}

func (p *Config) SelfServiceFlowVerificationAfterHooks(ctx context.Context, strategy string) []SelfServiceHook {
	return p.selfServiceHooks(ctx, HookStrategyKey(ViperKeySelfServiceVerificationAfter, strategy))
}

func (p *Config) SelfServiceFlowRecoveryReturnTo(ctx context.Context, defaultReturnTo *url.URL) *url.URL {
	return p.GetProvider(ctx).RequestURIF(ViperKeySelfServiceRecoveryBrowserDefaultReturnTo, defaultReturnTo)
}

func (p *Config) SelfServiceFlowRecoveryRequestLifespan(ctx context.Context) time.Duration {
	return p.GetProvider(ctx).DurationF(ViperKeySelfServiceRecoveryRequestLifespan, time.Hour)
}

func (p *Config) SelfServiceFlowRecoveryNotifyUnknownRecipients(ctx context.Context) bool {
	return p.GetProvider(ctx).BoolF(ViperKeySelfServiceRecoveryNotifyUnknownRecipients, false)
}

func (p *Config) SelfServiceLinkMethodLifespan(ctx context.Context) time.Duration {
	return p.GetProvider(ctx).DurationF(ViperKeyLinkLifespan, time.Hour)
}

func (p *Config) SelfServiceLinkMethodBaseURL(ctx context.Context) *url.URL {
	return cmp.Or(x.BaseURLFromContext(ctx), p.SelfPublicURL(ctx))
}

func (p *Config) SelfServiceCodeMethodLifespan(ctx context.Context) time.Duration {
	return p.GetProvider(ctx).DurationF(ViperKeyCodeLifespan, time.Hour)
}

func (p *Config) SelfServiceCodeMethodMaxSubmissions(ctx context.Context) int {
	return p.GetProvider(ctx).IntF(ViperKeyCodeMaxSubmissions, 5)
}

func (p *Config) SelfServiceCodeMethodMissingCredentialFallbackEnabled(ctx context.Context) bool {
	return p.GetProvider(ctx).Bool(ViperKeyCodeConfigMissingCredentialFallbackEnabled)
}

func (p *Config) DatabaseCleanupSleepTables(ctx context.Context) time.Duration {
	return p.GetProvider(ctx).Duration(ViperKeyDatabaseCleanupSleepTables)
}

func (p *Config) DatabaseCleanupBatchSize(ctx context.Context) int {
	return p.GetProvider(ctx).Int(ViperKeyDatabaseCleanupBatchSize)
}

func (p *Config) SelfServiceFlowRecoveryAfterHooks(ctx context.Context, strategy string) []SelfServiceHook {
	return p.selfServiceHooks(ctx, HookStrategyKey(ViperKeySelfServiceRecoveryAfter, strategy))
}

func (p *Config) SelfServiceFlowSettingsPrivilegedSessionMaxAge(ctx context.Context) time.Duration {
	return p.GetProvider(ctx).DurationF(ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, time.Hour)
}

func (p *Config) SessionSameSiteMode(ctx context.Context) http.SameSite {
	if !p.GetProvider(ctx).Exists(ViperKeySessionSameSite) {
		return p.CookieSameSiteMode(ctx)
	}

	switch p.GetProvider(ctx).StringF(ViperKeySessionSameSite, "Lax") {
	case "Lax":
		return http.SameSiteLaxMode
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	}
	return http.SameSiteDefaultMode
}

func (p *Config) SessionDomain(ctx context.Context) string {
	if !p.GetProvider(ctx).Exists(ViperKeySessionDomain) {
		return p.CookieDomain(ctx)
	}
	return p.GetProvider(ctx).String(ViperKeySessionDomain)
}

func (p *Config) SessionCookieSecure(ctx context.Context) bool {
	if !p.GetProvider(ctx).Exists(ViperKeySessionSecure) {
		return !p.IsInsecureDevMode(ctx)
	}
	return p.GetProvider(ctx).Bool(ViperKeySessionSecure)
}

func (p *Config) CookieDomain(ctx context.Context) string {
	return p.GetProvider(ctx).String(ViperKeyCookieDomain)
}

func (p *Config) SessionWhoAmIAAL(ctx context.Context) string {
	return p.GetProvider(ctx).String(ViperKeySessionWhoAmIAAL)
}

func (p *Config) SessionWhoAmICaching(ctx context.Context) bool {
	return p.GetProvider(ctx).Bool(ViperKeySessionWhoAmICaching)
}

func (p *Config) FeatureFlagFasterSessionExtend(ctx context.Context) bool {
	return p.GetProvider(ctx).Bool(ViperKeyFeatureFlagFasterSessionExtend)
}

func (p *Config) SessionWhoAmICachingMaxAge(ctx context.Context) time.Duration {
	return p.GetProvider(ctx).DurationF(ViperKeySessionWhoAmICachingMaxAge, 0)
}

func (p *Config) UseContinueWithTransitions(ctx context.Context) bool {
	return p.GetProvider(ctx).Bool(ViperKeyUseContinueWithTransitions)
}

func (p *Config) ChooseRecoveryAddress(ctx context.Context) bool {
	return p.GetProvider(ctx).Bool(ViperKeyChooseRecoveryAddress)
}

func (p *Config) SessionRefreshMinTimeLeft(ctx context.Context) time.Duration {
	return p.GetProvider(ctx).DurationF(ViperKeySessionRefreshMinTimeLeft, p.SessionLifespan(ctx))
}

func (p *Config) SelfServiceSettingsRequiredAAL(ctx context.Context) string {
	return p.GetProvider(ctx).String(ViperKeySelfServiceSettingsRequiredAAL)
}

func (p *Config) CookieSameSiteMode(ctx context.Context) http.SameSite {
	switch p.GetProvider(ctx).StringF(ViperKeyCookieSameSite, "Lax") {
	case "Lax":
		return http.SameSiteLaxMode
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	}
	return http.SameSiteDefaultMode
}

func (p *Config) SessionPath(ctx context.Context) string {
	if !p.GetProvider(ctx).Exists(ViperKeySessionPath) {
		return p.CookiePath(ctx)
	}
	return p.GetProvider(ctx).String(ViperKeySessionPath)
}

func (p *Config) CookiePath(ctx context.Context) string {
	return p.GetProvider(ctx).String(ViperKeyCookiePath)
}

func (p *Config) CookieSecure(ctx context.Context) bool {
	if !p.GetProvider(ctx).Exists(ViperKeyCookieSecure) {
		return !p.IsInsecureDevMode(ctx)
	}
	return p.GetProvider(ctx).Bool(ViperKeyCookieSecure)
}

func (p *Config) SelfServiceFlowLoginReturnTo(ctx context.Context, strategy string) *url.URL {
	return p.selfServiceReturnTo(ctx, ViperKeySelfServiceLoginAfter, strategy)
}

func (p *Config) SelfServiceFlowRegistrationReturnTo(ctx context.Context, strategy string) *url.URL {
	return p.selfServiceReturnTo(ctx, ViperKeySelfServiceRegistrationAfter, strategy)
}

func (p *Config) SelfServiceFlowSettingsReturnTo(ctx context.Context, strategy string, defaultReturnTo *url.URL) *url.URL {
	return p.GetProvider(ctx).RequestURIF(
		ViperKeySelfServiceSettingsAfter+"."+strategy+"."+DefaultBrowserReturnURL,
		p.GetProvider(ctx).RequestURIF(ViperKeySelfServiceSettingsAfter+"."+DefaultBrowserReturnURL,
			defaultReturnTo,
		),
	)
}

func (p *Config) selfServiceReturnTo(ctx context.Context, key string, strategy string) *url.URL {
	return p.GetProvider(ctx).RequestURIF(
		key+"."+strategy+"."+DefaultBrowserReturnURL,
		p.GetProvider(ctx).RequestURIF(key+"."+DefaultBrowserReturnURL,
			p.SelfServiceBrowserDefaultReturnTo(ctx),
		),
	)
}

func (p *Config) ConfigVersion(ctx context.Context) string {
	return p.GetProvider(ctx).StringF(ViperKeyVersion, UnknownVersion)
}

func (p *Config) PasswordPolicyConfig(ctx context.Context) *PasswordPolicy {
	return &PasswordPolicy{
		HaveIBeenPwnedHost:               p.GetProvider(ctx).StringF(ViperKeyPasswordHaveIBeenPwnedHost, "api.pwnedpasswords.com"),
		HaveIBeenPwnedEnabled:            p.GetProvider(ctx).BoolF(ViperKeyPasswordHaveIBeenPwnedEnabled, true),
		MaxBreaches:                      uint(p.GetProvider(ctx).Int(ViperKeyPasswordMaxBreaches)), // #nosec G115 -- negative values are prevented by the schema validation
		IgnoreNetworkErrors:              p.GetProvider(ctx).BoolF(ViperKeyIgnoreNetworkErrors, true),
		MinPasswordLength:                uint(p.GetProvider(ctx).IntF(ViperKeyPasswordMinLength, 8)), // #nosec G115 -- negative values are prevented by the schema validation
		IdentifierSimilarityCheckEnabled: p.GetProvider(ctx).BoolF(ViperKeyPasswordIdentifierSimilarityCheckEnabled, true),
	}
}

func (p *Config) WebAuthnForPasswordless(ctx context.Context) bool {
	return p.GetProvider(ctx).BoolF(ViperKeyWebAuthnPasswordless, false)
}

func (p *Config) WebAuthnConfig(ctx context.Context) *webauthn.Config {
	scheme := p.SelfPublicURL(ctx).Scheme
	id := p.GetProvider(ctx).String(ViperKeyWebAuthnRPID)
	origin := p.GetProvider(ctx).String(ViperKeyWebAuthnRPOrigin)
	origins := p.GetProvider(ctx).StringsF(ViperKeyWebAuthnRPOrigins, []string{stringsx.Coalesce(origin, scheme+"://"+id)})
	return &webauthn.Config{
		RPDisplayName: p.GetProvider(ctx).String(ViperKeyWebAuthnRPDisplayName),
		RPID:          id,
		RPOrigins:     origins,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			UserVerification: protocol.VerificationDiscouraged,
		},
		EncodeUserIDAsString: false,
	}
}

func (p *Config) PasskeyConfig(ctx context.Context) *webauthn.Config {
	scheme := p.SelfPublicURL(ctx).Scheme
	id := p.GetProvider(ctx).String(ViperKeyPasskeyRPID)
	origins := p.GetProvider(ctx).StringsF(ViperKeyPasskeyRPOrigins, []string{scheme + "://" + id})
	return &webauthn.Config{
		RPDisplayName: p.GetProvider(ctx).String(ViperKeyPasskeyRPDisplayName),
		RPID:          id,
		RPOrigins:     origins,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			AuthenticatorAttachment: "platform",
			RequireResidentKey:      pointerx.Ptr(true),
			ResidentKey:             protocol.ResidentKeyRequirementRequired,
			UserVerification:        protocol.VerificationPreferred,
		},
		EncodeUserIDAsString: false,
	}
}

func (p *Config) HasherPasswordHashingAlgorithm(ctx context.Context) string {
	configValue := p.GetProvider(ctx).StringF(ViperKeyHasherAlgorithm, DefaultPasswordHashingAlgorithm)
	switch configValue {
	case "bcrypt":
		return configValue
	case "argon2":
		fallthrough
	default:
		return configValue
	}
}

func (p *Config) CipherAlgorithm(ctx context.Context) string {
	configValue := p.GetProvider(ctx).StringF(ViperKeyCipherAlgorithm, DefaultCipherAlgorithm)
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

func (p *Config) GetProvider(ctx context.Context) *configx.Provider {
	return p.c.Config(ctx, p.p)
}

type SessionTokenizeFormat struct {
	TTL             time.Duration `koanf:"ttl" json:"ttl"`
	ClaimsMapperURL string        `koanf:"claims_mapper_url" json:"claims_mapper_url"`
	JWKSURL         string        `koanf:"jwks_url" json:"jwks_url"`
	SubjectSource   string        `koanf:"subject_source" json:"subject_source"`
}

func (p *Config) TokenizeTemplate(ctx context.Context, key string) (_ *SessionTokenizeFormat, err error) {
	var result SessionTokenizeFormat
	path := ViperKeySessionTokenizerTemplates + "." + key
	if !p.GetProvider(ctx).Exists(path) {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Unable to find tokenizer template \"%s\".", key))
	}

	if err := p.GetProvider(ctx).Unmarshal(path, &result); err != nil {
		return nil, errors.WithStack(herodot.ErrMisconfiguration.WithReasonf("Unable to decode tokenizer template \"%s\": %s", key, err))
	}

	return &result, nil
}

func (p *Config) DefaultConsistencyLevel(ctx context.Context) crdbx.ConsistencyLevel {
	return crdbx.ConsistencyLevelFromString(p.GetProvider(ctx).String(ViperKeyPreviewDefaultReadConsistencyLevel))
}

func (p *Config) PasswordMigrationHook(ctx context.Context) *PasswordMigrationHook {
	hook := &PasswordMigrationHook{
		Enabled: p.GetProvider(ctx).BoolF(ViperKeyPasswordMigrationHook+".enabled", false),
	}
	if !hook.Enabled {
		return hook
	}

	_ = p.GetProvider(ctx).Unmarshal(ViperKeyPasswordMigrationHook+".config", &hook.Config)

	return hook
}

func (p *Config) SelfServiceLoginFlowIdentifierFirstEnabled(ctx context.Context) bool {
	switch p.GetProvider(ctx).String(ViperKeySelfServiceLoginFlowStyle) {
	case "identifier_first":
		return true
	default:
		return false
	}
}

func (p *Config) SecurityAccountEnumerationMitigate(ctx context.Context) bool {
	return p.GetProvider(ctx).Bool(ViperKeySecurityAccountEnumerationMitigate)
}

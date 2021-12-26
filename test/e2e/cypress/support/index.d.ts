import { Session } from '@ory/kratos-client'

export interface MailMessage {
  fromAddress: string
  toAddresses: Array<string>
  body: string
  subject: string
}

declare global {
  namespace Cypress {
    interface Chainable {
      /**
       * Delete mails from the email server.
       *
       * @param options
       */
      deleteMail(options: { atLeast?: boolean }): Chainable<void>

      /**
       * Fetch the browser's Ory Session.
       *
       * @param opts
       */
      getSession(opts?: {
        expectAal?: 'aal2' | 'aal1'
        expectMethods?: Array<
          'password' | 'webauthn' | 'lookup_secret' | 'totp'
        >
      }): Chainable<Session>

      /**
       * Expect that the browser has no valid Ory Kratos Cookie Session.
       */
      noSession(): Chainable<Response<any>>

      /**
       * Log a user in
       *
       * @param opts
       */
      login(opts: {
        email: string
        password: string
        expectSession?: boolean
        cookieUrl?: string
      }): Chainable<Response<Session | undefined>>

      /**
       * Sign up a user
       *
       * @param opts
       */
      register(opts: {
        email: string
        password: string
        query?: { [key: string]: string }
        fields?: { [key: string]: any }
      }): Chainable<Response<void>>

      /**
       * Updates a user's settings using an API flow
       *
       * @param opts
       */
      settingsApi(opts: {
        fields?: { [key: string]: any }
      }): Chainable<Response<void>>

      /**
       * Set the "privileged session lifespan" to a large value.
       */
      longPrivilegedSessionTime(): Chainable<void>

      /**
       * Return and optionally delete an email from the fake email server.
       *
       * @param opts
       */
      getMail(opts?: { removeMail: boolean }): Chainable<MailMessage>

      performEmailVerification(opts?: {
        expect?: { email?: string; redirectTo?: string }
      }): Chainable<void>

      /**
       * Sets the Ory Kratos configuration profile.
       *
       * @param profile
       */
      useConfigProfile(profile: string): Chainable<void>

      /**
       * Register a new user with email + password via the API.
       *
       * Recommended to use if you need an Ory Session Token or alternatively the user created.
       *
       * @param opts
       */
      registerApi(opts?: {
        email: string
        password: string
        fields: { [key: string]: string }
      }): Chainable<Session>

      /**
       * Submits a recovery flow via the API
       *
       * @param opts
       */
      recoverApi(opts: { email: string; returnTo?: string }): Chainable<void>

      /**
       * Changes the config so that the login flow lifespan is very short.
       *
       *
       * Useful when testing expiry of login flows.
       * @see longLoginLifespan()
       */
      shortLoginLifespan(): Chainable<void>

      /**
       * Changes the config so that the login flow lifespan is very long.
       *
       * Useful when testing expiry of login flows.
       *
       * @see shortLoginLifespan()
       */
      longLoginLifespan(): Chainable<void>

      /**
       * Change the config so that `https://www.ory.sh/` is a allowed return to URL.
       */
      browserReturnUrlOry(): Chainable<void>

      /**
       * Changes the config so that the registration flow lifespan is very short.
       *
       * Useful when testing expiry of registration flows.
       *
       * @see longRegisterLifespan()
       */
      shortRegisterLifespan(): Chainable<void>

      /**
       * Changes the config so that the registration flow lifespan is very long.
       *
       * Useful when testing expiry of registration flows.
       *
       * @see shortRegisterLifespan()
       */
      longRegisterLifespan(): Chainable<void>

      /**
       * Changes the config so that the settings privileged lifespan is very long.
       *
       * Useful when testing privileged settings flows.
       *
       * @see longPrivilegedSessionTime()
       */
      shortPrivilegedSessionTime(): Chainable<void>

      /**
       * Re-authenticates a user.
       *
       * @param opts
       */
      reauth(opts: {
        expect: { email; success?: boolean }
        type: { email?: string; password?: string }
      }): Chainable<void>

      /**
       * Do not require 2fa for /session/whoami
       */
      sessionRequiresNo2fa(): Chainable<void>

      /**
       * Require 2fa for /session/whoami if available
       */
      sessionRequires2fa(): Chainable<void>

      /**
       * Like sessionRequires2fa but sets this also for settings
       */
      requireStrictAal(): Chainable<void>

      /**
       * Like sessionRequiresNo2fa but sets this also for settings
       */
      useLaxAal(): Chainable<void>

      /**
       * Gets the lookup codes from the settings page
       */
      getLookupSecrets(): Chainable<Array<string>>

      /**
       * Expect the settings to be saved.
       */
      expectSettingsSaved(): Chainable<void>

      clearCookies(
        options?: Partial<Loggable & Timeoutable & { domain: null | string }>
      ): Chainable<null>

      /**
       * A workaround for cypress not being able to clear cookies properly
       */
      clearAllCookies(): Chainable<null>

      /**
       * Submits a password form by clicking the button with method=password
       */
      submitPasswordForm(): Chainable<null>

      /**
       * Submits a profile form by clicking the button with method=profile
       */
      submitProfileForm(): Chainable<null>

      /**
       * Expect a CSRF error to occur
       *
       * @param opts
       */
      shouldHaveCsrfError(opts: { app: string }): Chainable<void>

      /**
       * Expects the app to error if a return_to is used which isn't allowed.
       *
       * @param init
       * @param opts
       */
      shouldErrorOnDisallowedReturnTo(
        init: string,
        opts: { app: string }
      ): Chainable<void>

      /**
       * Expect that the second factor login screen is shown
       */
      shouldShow2FAScreen(): Chainable<void>

      /** Click a webauthn button
       *
       * @param type
       */
      clickWebAuthButton(type: 'login' | 'register'): Chainable<void>

      /**
       * Sign up a user using Social Sign In
       *
       * @param opts
       */
      registerOidc(opts: {
        email?: string
        website?: string
        scopes?: Array<string>
        rememberLogin?: boolean
        rememberConsent?: boolean
        acceptLogin?: boolean
        acceptConsent?: boolean
        expectSession?: boolean
        route?: string
      }): Chainable<void>

      /**
       * Sign in a user using Social Sign In
       *
       * @param opts
       */
      loginOidc(opts: {
        expectSession?: boolean
        url?: string
      }): Chainable<void>

      /**
       * Triggers a Social Sign In flow for the given provider
       *
       * @param provider
       */
      triggerOidc(provider?: string): Chainable<void>

      /**
       * Changes the config so that the recovery privileged lifespan is very long.
       *
       * Useful when testing privileged recovery flows.
       *
       * @see shortPrivilegedRecoveryTime()
       */
      longRecoveryLifespan(): Chainable<void>

      /**
       * Changes the config so that the verification privileged lifespan is very long.
       *
       * Useful when testing privileged verification flows.
       *
       * @see shortPrivilegedVerificationTime()
       */
      longVerificationLifespan(): Chainable<void>

      /**
       * Log a user out
       */
      logout(): Chainable<void>

      /**
       * Deletes all mail in the mail mock server.
       *
       * @param opts
       */
      deleteMail(opts?: { atLeast?: number }): Chainable<void>

      /**
       * Changes the config so that the link lifespan is very long.
       *
       * Useful when testing recovery/verification flows.
       *
       * @see shortLinkLifespan()
       */
      longLinkLifespan(): Chainable<void>

      /**
       * Changes the config so that the link lifespan is very short.
       *
       * Useful when testing recovery/verification flows.
       *
       * @see longLinkLifespan()
       */
      shortLinkLifespan(): Chainable<void>

      /**
       * Expect a recovery email which is expired.
       *
       * @param opts
       */
      recoverEmailButExpired(opts?: {
        expect: { email: string }
      }): Chainable<void>

      /**
       * Expect a verification email which is expired.
       *
       * @param opts
       */
      verifyEmailButExpired(opts?: {
        expect: { password?: string; email: string }
      }): Chainable<string>

      /**
       * Disables verification
       */
      disableVerification(): Chainable<void>

      /**
       * Enables verification
       */
      enableVerification(): Chainable<void>

      /**
       * Enables recovery
       */
      enableRecovery(): Chainable<void>

      /**
       * Disabled recovery
       */
      disableRecovery(): Chainable<void>

      /**
       * Disables registration
       */
      disableRegistration(): Chainable<void>

      /**
       * Enables registration
       */
      enableRegistration(): Chainable<void>

      /**
       * Expect a recovery email which is valid.
       *
       * @param opts
       */
      recoverEmail(opts: {
        expect: { email: string }
        shouldVisit?: boolean
      }): Chainable<string>

      /**
       * Expect a verification email which is valid.
       *
       * @param opts
       */
      verifyEmail(opts: {
        expect: { email: string; password?: string; redirectTo?: string }
        shouldVisit?: boolean
      }): Chainable<string>

      /**
       * Configures a hook which only allows verified email addresses to sign in.
       */
      enableLoginForVerifiedAddressOnly(): Chainable<void>

      /**
       * Sign a user in via the API and return the session.
       *
       * @param opts
       */
      loginApi(opts: {
        email: string
        password: string
      }): Chainable<{ session: Session }>

      /**
       * Same as loginApi but uses dark magic to avoid cookie issues.
       *
       * @param opts
       */
      loginApiWithoutCookies(opts: {
        email: string
        password: string
      }): Chainable<{ session: Session }>

      /**
       * Which app to proxy
       */
      proxy(app: 'react' | 'express'): Chainable<void>

      /**
       * Log a user in on mobile
       *
       * @param opts
       */
      loginMobile(opts: { email: string; password: string }): Chainable<void>

      /**
       * Set the identity schema
       * @param schema
       */
      setIdentitySchema(schema: string): Chainable<void>
    }
  }
}

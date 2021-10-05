import {Session} from "@ory/kratos-client";

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
      deleteMail(options: {
        atLeast?: boolean,
      }): Chainable<void>

      /**
       * Fetch the browser's Ory Session.
       *
       * @param opts
       */
      getSession(opts?:
                   {
                     expectAal: 'aal2' | 'aal1',
                     expectMethods: Array<'password' | 'webauthn' | 'lookup_secret'>,
                   }
      ): Chainable<Session>

      /**
       * Expect that the browser has no valid Ory Kratos Cookie Session.
       */
      noSession(): Chainable<Response<any>>

      /**
       * Log a user in
       *
       * @param opts
       */
      login(opts: { email: string, password: string, expectSession?: boolean, cookieUrl?: string }): Chainable<Response<Session | undefined>>

      /**
       * Sign up a user
       *
       * @param opts
       */
      register(opts: {
        email: string,
        password: string,
        query?: { [key: string]: string },
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

      performEmailVerification(opts?: { expect?: { email?: string, redirectTo?: string } }): Chainable<void>

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
      registerApi(opts?: { email: string, password: string, fields: { [key: string]: string } }): Chainable<Session>

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
        expect: { email },
        type: { email?: string, password?: string }
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
       * Gets the lookup codes from the settings page
       */
      getLookupSecrets(): Chainable<Array<string>>

      /**
       * Expect the settings to be saved.
       */
      expectSettingsSaved(): Chainable<void>

      clearCookies(options?: Partial<Loggable & Timeoutable & { domain: null | string }>): Chainable<null>

      /**
       * A workaround for cypress not being able to clear cookies properly
       */
      clearAllCookies(): Chainable<null>

      /**
       * Submits a password form by clicking the button with method=password
       */
      submitPasswordForm(): Chainable<null>

      /**
       * Expect a CSRF error to occur
       *
       * @param opts
       */
      shouldHaveCsrfError(opts: {app: string }) : Chainable<void>
    }
  }
}

import {Session} from "@ory/kratos-client";
import {gen} from "../helpers";

export interface MailMessage {
  fromAddress: string
  toAddresses: Array<string>
  body:string
  subject:string
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
          expectAal: 'aal2' |'aal1',
          expectMethods: Array<'password' | 'webauthn'>,
        }
      ): Chainable<Session>

      /**
       * Expect that the browser has no valid Ory Kratos Cookie Session.
       */
      noSession(): Chainable<Response<any>>

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

      performEmailVerification(opts?: {expect?: {email?: string, redirectTo?: string}}): Chainable<void>

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
      registerApi(opts?: {email:string, password:string, fields: { [key: string]: string  }}): Chainable<Session>

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
       * @see shortLoginLifespan()
       */
      longLoginLifespan(): Chainable<void>

      /**
       * Change the config so that `https://www.ory.sh/` is a allowed return to URL.
       */
      browserReturnUrlOry(): Chainable<void>
    }
  }
}

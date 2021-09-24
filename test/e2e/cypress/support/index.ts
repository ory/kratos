import './commands'
import {Session} from "@ory/kratos-client";

export interface MailMessage {
  fromAddress: string
  toAddresses: Array<string>
  body:string
  subject:string
}

declare global {
  namespace Cypress {
    interface Chainable {
      deleteMail(options: {
        atLeast?: boolean,
      }): Chainable<Element>

      getSession(opts?:
        {
          expectAal: 'aal2' |'aal1',
          expectMethods: Array<'password' | 'webauthn'>,
        }
      ): Chainable<Session>

      noSession(): Chainable<Response<any>>

      longPrivilegedSessionTime(): Chainable<void>

      getMail(opts?: { removeMail: boolean }): Chainable<MailMessage>

      performEmailVerification(opts?: {expect?: {email?: string, redirectTo?: string}}): Chainable<void>

      useConfigProfile(profile: string): Chainable<void>
    }
  }
}

export {}

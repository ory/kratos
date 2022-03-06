import { appPrefix, gen } from '../../../helpers'
import { routes as express } from '../../../helpers/express'
import { routes as react } from '../../../helpers/react'

const signup = (registration: string, email = gen.email()) => {
  cy.visit(registration)

  cy.get('[name="webauthn_register_displayname"]').type('key1')
  cy.get('[name="traits.email"]').type(email)
  cy.get('[name="traits.website"]').type('https://www.ory.sh')
  cy.clickWebAuthButton('register')
  cy.getSession({
    expectAal: 'aal1',
    expectMethods: ['webauthn']
  }).then((session) => {
    expect(session.identity.traits.email).to.equal(email)
    expect(session.identity.traits.website).to.equal('https://www.ory.sh')
  })
}

context('Passwordless registration', () => {
  before(() => {
    cy.task('resetCRI', {})
  })
  after(() => {
    cy.task('resetCRI', {})
  })
  ;[
    {
      login: react.login,
      registration: express.registration,
      settings: react.settings,
      base: react.base,
      app: 'react' as 'react',
      profile: 'passwordless'
    },
    {
      login: express.login,
      registration: express.registration,
      settings: express.settings,
      base: express.base,
      app: 'express' as 'express',
      profile: 'passwordless'
    }
  ].forEach(({ registration, login, profile, app, base, settings }) => {
    describe(`for app ${app}`, () => {
      let authenticator
      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)
        cy.addVirtualAuthenticator().then((result) => {
          authenticator = result
        })
        cy.longPrivilegedSessionTime()
      })

      beforeEach(() => {
        cy.clearAllCookies()
      })

      after(() => {
        cy.task('sendCRI', {
          query: 'WebAuthn.removeVirtualAuthenticator',
          opts: authenticator
        })
      })

      it('should register after validation errors', () => {
        cy.visit(registration)

        cy.get(appPrefix(app) + '[name="webauthn_register_displayname"]').type(
          'key1'
        )
        cy.get('[name="traits.website"]').type('b')
        cy.clickWebAuthButton('register')

        cy.get('[data-testid="ui/message/4000002"]').should('to.exist')
        cy.get('[data-testid="ui/message/4000001"]').should('to.exist')
        cy.get('[name="traits.website"]').should('have.value', 'b')
        const email = gen.email()
        cy.get('[name="traits.email"]').type(email)
        cy.clickWebAuthButton('register')

        cy.get('[data-testid="ui/message/4000001"]').should('to.exist')
        cy.get('[name="traits.website"]').should('have.value', 'b')
        cy.get('[name="traits.email"]').should('have.value', email)
        cy.get('[name="traits.website"]').clear().type('https://www.ory.sh')
        cy.clickWebAuthButton('register')
        cy.getSession({
          expectAal: 'aal1',
          expectMethods: ['webauthn']
        }).then((session) => {
          expect(session.identity.traits.email).to.equal(email)
          expect(session.identity.traits.website).to.equal('https://www.ory.sh')
        })
      })

      it('should be able to login with registered account', () => {
        const email = gen.email()
        signup(registration, email)
        cy.logout()
        cy.visit(login)
        cy.get('[name="identifier"]').type(email)
        cy.get('[value="webauthn"]').click()
        cy.get('[data-testid="ui/message/1010012"]').should('to.exist')
        cy.get('[name="password"]').should('to.not.exist')
        cy.clickWebAuthButton('login')
        cy.getSession({
          expectAal: 'aal1',
          expectMethods: ['webauthn']
        }).then((session) => {
          expect(session.identity.traits.email).to.equal(email)
          expect(session.identity.traits.website).to.equal('https://www.ory.sh')
        })
      })

      it('should not be able to unlink last security key', () => {
        const email = gen.email()
        signup(registration, email)
        cy.visit(settings)
        cy.get('[name="webauthn_remove"]').should('not.exist')
      })

      it('should be able to link password and use both methods for sign in', () => {
        const email = gen.email()
        const password = gen.password()
        signup(registration, email)
        cy.visit(settings)
        cy.get('[name="webauthn_remove"]').should('not.exist')
        cy.get('[name="password"]').type(password)
        cy.get('[value="password"]').click()
        cy.expectSettingsSaved()
        cy.get('[name="webauthn_remove"]').click()
        cy.expectSettingsSaved()
        cy.logout()
        cy.visit(login)
        cy.get('[name="identifier"]').type(email)
        cy.get('[value="webauthn"]').click()
        cy.get('[data-testid="ui/message/4000015"]').should('to.exist')
        cy.get('[name="identifier"]').should('exist')
        cy.get('[name="password"]').should('exist')
        cy.get('[value="password"]').should('exist')
      })

      it('should be able to refresh', () => {
        const email = gen.email()
        signup(registration, email)
        cy.visit(login + '?refresh=true')
        cy.get('[name="identifier"][type="hidden"]').should('exist')
        cy.get('[name="identifier"][type="input"]').should('not.exist')
        cy.get('[name="password"]').should('not.exist')
        cy.get('[value="password"]').should('not.exist')
        cy.clickWebAuthButton('login')
        cy.getSession({
          expectAal: 'aal1',
          expectMethods: ['webauthn', 'webauthn']
        }).then((session) => {
          expect(session.identity.traits.email).to.equal(email)
          expect(session.identity.traits.website).to.equal('https://www.ory.sh')
        })
      })

      it('should not be able to use for MFA', () => {
        const email = gen.email()
        signup(registration, email)
        cy.visit(login + '?aal=aal2')
        cy.get('[value="webauthn"]').should('not.exist')
        cy.get('[name="webauthn_login_trigger"]').should('not.exist')
      })

      it('should be able to add method later and try a variety of refresh flows', () => {
        const email = gen.email()
        const password = gen.password()
        cy.visit(registration)

        cy.get('[name="traits.email"]').type(email)
        cy.get('[name="password"]').type(password)
        cy.get('[name="traits.website"]').type('https://www.ory.sh')
        cy.get('[value="password"]').click()
        cy.location('pathname').should('not.contain', '/registration')
        cy.getSession({
          expectAal: 'aal1',
          expectMethods: ['password']
        })

        cy.visit(settings)
        cy.get('[name="webauthn_register_displayname"]').type('key2')
        cy.clickWebAuthButton('register')
        cy.expectSettingsSaved()

        cy.visit(login + '?refresh=true')
        cy.get('[name="password"]').should('exist')
        cy.clickWebAuthButton('login')
        cy.location('pathname').should('not.contain', '/login')
        cy.getSession({
          expectAal: 'aal1',
          expectMethods: ['password', 'webauthn', 'webauthn']
        })

        cy.visit(login + '?refresh=true')
        cy.get('[name="password"]').type(password)
        cy.get('[value="password"]').click()
        cy.getSession({
          expectAal: 'aal1',
          expectMethods: ['password', 'webauthn', 'webauthn', 'password']
        })

        cy.logout()
        cy.visit(login)
        cy.get('[name="identifier"]').type(email)
        cy.get('[value="webauthn"]').click()
        cy.clickWebAuthButton('login')
        cy.getSession({
          expectAal: 'aal1',
          expectMethods: ['webauthn']
        })
      })

      it('should not be able to use for MFA even when passwordless is false', () => {
        const email = gen.email()
        signup(registration, email)
        cy.updateConfigFile((config) => {
          config.selfservice.methods.webauthn.config.passwordless = false
          return config
        })
        cy.visit(login + '?aal=aal2')
        cy.get('[value="webauthn"]').should('not.exist')
        cy.get('[name="webauthn_login_trigger"]').should('not.exist')

        cy.visit(settings)
        cy.get('[name="webauthn_remove"]').should('not.exist')
        cy.get('[name="webauthn_register_displayname"]').type('key2')
        cy.clickWebAuthButton('register')
        cy.expectSettingsSaved()

        cy.visit(login + '?aal=aal2&refresh=true')
        cy.clickWebAuthButton('login')
        cy.getSession({
          expectAal: 'aal2',
          expectMethods: ['webauthn', 'webauthn', 'webauthn']
        })
      })
    })
  })
})

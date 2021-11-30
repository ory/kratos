import { appPrefix, gen, website } from '../../../../helpers'
import { routes as react } from '../../../../helpers/react'
import { routes as express } from '../../../../helpers/express'

context('Settings failures with email profile', () => {
  ;[
    {
      route: express.settings,
      base: express.base,
      app: 'express' as 'express',
      profile: 'email'
    },
    {
      route: react.settings,
      base: react.base,
      app: 'react' as 'react',
      profile: 'spa'
    }
  ].forEach(({ route, profile, app, base }) => {
    describe(`for app ${app}`, () => {
      let email = gen.email()
      let password = gen.password()

      const emailSecond = `second-${gen.email()}`
      const passwordSecond = gen.password()

      const up = (value) => `not-${value}`

      before(() => {
        cy.proxy(app)
        cy.useConfigProfile(profile)
        cy.registerApi({
          email: emailSecond,
          password: passwordSecond,
          fields: { 'traits.website': 'https://github.com/ory/kratos' }
        })
        cy.registerApi({
          email,
          password,
          fields: { 'traits.website': website }
        })
      })

      beforeEach(() => {
        cy.longPrivilegedSessionTime()

        cy.visit(base)
        cy.clearAllCookies()

        cy.login({ email, password, cookieUrl: base })
        cy.visit(route)
      })

      describe('profile', () => {
        beforeEach(() => {
          cy.visit(route)
        })

        it('fails with validation errors', () => {
          cy.get('input[name="traits.website"]').clear().type('http://s')
          cy.get('[name="method"][value="profile"]').click()
          cy.get('[data-testid^="ui/message"]').should(
            'contain.text',
            'length must be >= 10'
          )
        })

        it('fails because reauth is another person', () => {
          cy.get('input[name="traits.email"]').clear().type(up(email))
          cy.shortPrivilegedSessionTime()
          cy.get('button[value="profile"]').click()

          cy.reauth({
            expect: { email },
            type: { email: emailSecond, password: passwordSecond }
          })

          // We end up in a new settings flow for the second user
          cy.get('input[name="traits.email"]').should('have.value', emailSecond)

          // Try to log in with updated credentials -> should fail
          cy.clearAllCookies()
          cy.login({
            email: up(email),
            password,
            expectSession: false,
            cookieUrl: base
          })
        })

        it('does not update data because resumable session was removed', () => {
          cy.get('input[name="traits.email"]').clear().type(up(email))
          cy.shortPrivilegedSessionTime()
          cy.get('button[value="profile"]').click()

          cy.clearAllCookies()
          cy.login({ email, password, cookieUrl: base })

          cy.getSession().should((session) => {
            const { identity } = session
            expect(identity.traits.email).to.equal(email)
          })
        })

        it('does not update without re-auth', () => {
          cy.get('input[name="traits.email"]').clear().type(up(email))
          cy.shortPrivilegedSessionTime() // wait for the privileged session to time out
          cy.get('button[value="profile"]').click()

          cy.visit(base)

          cy.getSession().should((session) => {
            const { identity } = session
            expect(identity.traits.email).to.equal(email)
          })
        })

        it('does not resume another failed request', () => {
          // checks here that we're checking settingsRequest.id == cookie.stored.id
          cy.get('input[name="traits.email"]').clear().type(up(email))
          cy.shortPrivilegedSessionTime() // wait for the privileged session to time out
          cy.get('button[value="profile"]').click()
          cy.location('pathname').should('not.contain', '/settings')

          cy.visit(route)
          cy.get('input[name="traits.website"]')
            .clear()
            .type('http://github.com/aeneasr')
          cy.get('button[value="profile"]').click()
          cy.expectSettingsSaved()

          cy.getSession().should((session) => {
            const { identity } = session
            expect(identity.traits.email).to.equal(email) // this is NOT up(email)
            expect(identity.traits.website).to.equal(
              'http://github.com/aeneasr'
            ) // this is NOT up(email)
          })
        })
      })

      describe('password', () => {
        beforeEach(() => {
          cy.longPrivilegedSessionTime()
        })

        afterEach(() => {
          cy.longPrivilegedSessionTime()
        })

        it('fails if password policy is violated', () => {
          cy.get('input[name="password"]').clear().type('12345678')
          cy.get('button[value="password"]').click()
          cy.get('*[data-testid^="ui/message"]').should(
            'contain.text',
            'data breaches'
          )
        })

        it('fails because reauth is another person', () => {
          cy.get('input[name="password"]').clear().type(up(password))

          let firstSession
          cy.getSession().then((session) => {
            firstSession = session
          })

          cy.shortPrivilegedSessionTime() // wait for the privileged session to time out
          cy.get('button[value="password"]').click()

          cy.reauth({
            expect: { email },
            type: { email: emailSecond, password: passwordSecond }
          })

          // We want to ensure that the reauth session is completely different from the one we had in the first place.
          cy.getSession().then((session) => {
            expect(session.authentication_methods).to.have.length(1)
            expect(session.identity.traits.email).to.eq(emailSecond)
            expect(session.id).to.not.eq(firstSession.id)
            expect(session.identity.id).to.not.eq(firstSession.identity.id)
            expect(session.authenticated_at).to.not.eq(
              firstSession.authenticated_at
            )
          })

          // We end up in a new settings flow for the second user
          cy.get('input[name="traits.email"]').should('have.value', emailSecond)

          // Try to log in with updated credentials -> should fail
          cy.clearAllCookies()
          cy.login({
            email,
            password: up(password),
            expectSession: false,
            cookieUrl: base
          })
        })

        it('does not update without re-auth', () => {
          cy.get('input[name="password"]').clear().type(up(password))
          cy.shortPrivilegedSessionTime() // wait for the privileged session to time out
          cy.get('button[value="password"]').click()

          cy.visit(base)
          cy.clearAllCookies()
          cy.login({
            email,
            password: up(password),
            expectSession: false,
            cookieUrl: base
          })
        })

        it('does not update data because resumable session was removed', () => {
          cy.get('input[name="password"]').clear().type(up(password))
          cy.shortPrivilegedSessionTime() // wait for the privileged session to time out
          cy.get('button[value="password"]').click()

          cy.clearAllCookies()
          cy.login({ email, password, cookieUrl: base })
          cy.clearAllCookies()
          cy.login({
            email,
            password: up(password),
            expectSession: false,
            cookieUrl: base
          })
        })

        it('does not resume another queued request', () => {
          const email = gen.email()
          const password = gen.password()
          cy.clearAllCookies()
          cy.register({
            email,
            password,
            fields: { 'traits.website': website }
          })
          cy.visit(route)

          // checks here that we're checking settingsRequest.id == cookie.stored.id
          const invalidPassword = 'invalid-' + gen.password()
          cy.get('input[name="password"]').clear().type(invalidPassword)
          cy.shortPrivilegedSessionTime() // wait for the privileged session to time out
          cy.get('button[value="password"]').click()
          cy.location('pathname').should('include', '/login')

          const validPassword = 'valid-' + gen.password()
          cy.visit(route)
          cy.get('input[name="password"]').clear().type(validPassword)
          cy.get('button[value="password"]').click()

          cy.location('pathname').should('include', '/login')
          cy.reauth({ expect: { email }, type: { password: password } })

          cy.location('pathname').should('include', '/settings')
          cy.get('input[name="password"]').should('exist')

          // This should pass because it is the correct password
          cy.clearAllCookies()
          cy.login({ email, password: validPassword, cookieUrl: base })

          // This should fail because it is the wrong password
          cy.clearAllCookies()
          cy.login({
            email,
            password: invalidPassword,
            expectSession: false,
            cookieUrl: base
          })

          cy.clearAllCookies()
          cy.login({
            email,
            password: password,
            expectSession: false,
            cookieUrl: base
          })
        })
      })

      describe('global errors', () => {
        it('fails when CSRF is incorrect', () => {
          cy.get(appPrefix(app) + 'input[name="password"]').type('12345678')
          cy.shouldHaveCsrfError({ app })
        })

        it('fails when a disallowed return_to url is requested', () => {
          cy.shouldErrorOnDisallowedReturnTo(
            route + '?return_to=https://not-allowed',
            { app }
          )
        })
      })
    })
  })
})

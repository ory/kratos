import { appPrefix, gen, website } from '../../../helpers'
import { routes as express } from '../../../helpers/express'
import { routes as react } from '../../../helpers/react'

context('2FA lookup secrets', () => {
  ;[
    {
      login: react.login,
      settings: react.settings,
      base: react.base,
      app: 'react' as 'react',
      profile: 'spa'
    },
    {
      login: express.login,
      settings: express.settings,
      base: express.base,
      app: 'express' as 'express',
      profile: 'mfa'
    }
  ].forEach(({ settings, login, profile, app, base }) => {
    describe(`for app ${app}`, () => {
      before(() => {
        cy.useConfigProfile(profile)
        cy.proxy(app)
      })

      let email = gen.email()
      let password = gen.password()

      beforeEach(() => {
        cy.visit(base)
        cy.clearAllCookies()
        email = gen.email()
        password = gen.password()
        cy.registerApi({
          email,
          password,
          fields: { 'traits.website': website }
        })
        cy.login({ email, password, cookieUrl: base })

        cy.longPrivilegedSessionTime()
        cy.sessionRequiresNo2fa()
      })

      it('should be able to remove lookup codes', () => {
        cy.sessionRequires2fa()
        cy.longPrivilegedSessionTime()
        cy.visit(settings)
        cy.get(
          appPrefix(app) + 'button[name="lookup_secret_regenerate"]'
        ).click()
        cy.get('button[name="lookup_secret_confirm"]').click()
        cy.expectSettingsSaved()
        cy.visit(settings)

        cy.shortPrivilegedSessionTime()
        cy.get('button[name="lookup_secret_disable"]').click()
        cy.reauth({
          expect: { email },
          type: { email: email, password: password }
        })
        cy.expectSettingsSaved()

        cy.clearAllCookies()
        cy.login({ email: email, password: password, cookieUrl: base })

        cy.visit(login + '?aal=aal2')
        cy.get('h2').should('contain.text', 'Two-Factor Authentication')
        cy.get('*[name="method"][value="totp"]').should('not.exist')
        cy.get('*[name="method"][value="lookup_secret"]').should('not.exist')
        cy.get('*[name="method"][value="password"]').should('not.exist')
      })

      it('should go through several lookup secret lifecycles', () => {
        cy.visit(settings)

        cy.get('[data-testid="node/text/lookup_secret_codes/label"]').should(
          'not.exist'
        )
        cy.get('[data-testid="text-lookup_secret_codes-content"] code').should(
          'not.exist'
        )
        cy.get('button[name="lookup_secret_confirm"]').should('not.exist')
        cy.get('button[name="lookup_secret_regenerate"]').click()
        cy.get('[data-testid="node/text/lookup_secret_codes/label"]').should(
          'contain.text',
          'These are your back up recovery codes.'
        )
        cy.get('[data-testid="text-lookup_secret_codes-content"] code').should(
          'not.be.empty'
        )

        let codes
        cy.getLookupSecrets().should((c) => {
          codes = c
        })

        cy.get('button[name="lookup_secret_confirm"]').click()
        cy.expectSettingsSaved()

        cy.get('button[name="lookup_secret_reveal"]').should('exist')
        cy.get('[data-testid="text-lookup_secret_codes-content"] code').should(
          'not.exist'
        )
        cy.get('button[name="lookup_secret_confirm"]').should('not.exist')
        cy.get('button[name="lookup_secret_regenerate"]').should('not.exist')

        cy.get('button[name="lookup_secret_reveal"]').click()
        cy.getLookupSecrets().should((c) => {
          codes = c
        })

        cy.getSession({
          expectAal: 'aal2',
          expectMethods: ['password', 'lookup_secret']
        })

        // Try to log in with a recovery code now
        cy.visit(login + '?aal=aal2&refresh=true')
        cy.location('pathname').should('contain', 'login')

        cy.get('*[name="method"][value="lookup_secret"]').should('exist')
        cy.get('*[name="method"][value="password"]').should('not.exist')

        // Type an invalid code
        cy.get('input[name="lookup_secret"]').should('exist')
        cy.get('input[name="lookup_secret"]').type('invalid-code')
        cy.get('*[name="method"][value="lookup_secret"]').click()
        cy.get('[data-testid="ui/message/4000016"]').should(
          'contain.text',
          'The backup recovery code is not valid.'
        )

        // Type a valid code
        cy.get('input[name="lookup_secret"]').should('exist')
        cy.get('input[name="lookup_secret"]').should('have.value', '')
        cy.get('input[name="lookup_secret"]').then(($e) => {
          cy.wrap($e).type(codes[0])
        })
        cy.get('*[name="method"][value="lookup_secret"]').click()
        cy.location('pathname').should('not.contain', 'login')

        let authenticatedAt
        cy.getSession({
          expectAal: 'aal2',
          expectMethods: ['password', 'lookup_secret', 'lookup_secret']
        }).then((session) => {
          authenticatedAt = session.authenticated_at
          expect(session.authenticator_assurance_level).to.equal('aal2')
        })

        // Retry auth with the used code
        cy.visit(login + '?aal=aal2&refresh=true')
        cy.location().should((loc) => {
          expect(loc.href).to.include('/login')
        })
        cy.get('input[name="lookup_secret"]').then(($e) => {
          cy.wrap($e).type(codes[0])
        })
        cy.get('*[name="method"][value="lookup_secret"]').click()
        // Use a valid code
        cy.get('[data-testid="ui/message/4000012"]').should(
          'contain.text',
          'This backup recovery code has already been used.'
        )

        cy.get('input[name="lookup_secret"]').then(($e) => {
          cy.wrap($e).type(codes[1])
        })
        cy.get('*[name="method"][value="lookup_secret"]').click()
        cy.location('pathname').should('not.contain', 'login')

        cy.getSession({
          expectAal: 'aal2',
          expectMethods: [
            'password',
            'lookup_secret',
            'lookup_secret',
            'lookup_secret'
          ]
        }).then((session) => {
          expect(session.authenticated_at).to.not.equal(authenticatedAt)
        })

        // Going back to the settings UI we should see that the codes have been "used"
        cy.visit(settings)
        cy.get('button[name="lookup_secret_reveal"]').click()
        cy.getLookupSecrets().should((c) => {
          let newCodes = codes
          newCodes[0] = 'Used'
          newCodes[1] = 'Used'
          expect(c).to.eql(newCodes)
        })

        // Regenerating the codes means the old one become invalid
        cy.get('*[name=lookup_secret_regenerate]').click()
        cy.get('*[name=lookup_secret_confirm]').should('exist')
        let regenCodes
        cy.getLookupSecrets().should((c) => {
          regenCodes = c
        })

        // Confirm it
        cy.get('*[name=lookup_secret_confirm]').click()
        cy.get('*[name="lookup_secret_reveal"]').click()
        cy.getLookupSecrets().should((c) => {
          expect(c).to.eql(regenCodes)
        })

        // Log in and see if we can use the old / new keys
        cy.visit(login + '?aal=aal2&refresh=true')
        cy.location('pathname').should('contain', 'login')

        // Using an old code fails
        cy.get('input[name="lookup_secret"]').then(($e) => {
          cy.wrap($e).type(codes[3])
        })
        cy.get('*[name="method"][value="lookup_secret"]').click()
        cy.get('[data-testid="ui/message/4000016"]').should('exist')

        // Using a new code succeeds
        cy.get('input[name="lookup_secret"]').then(($e) => {
          cy.wrap($e).type(regenCodes[0])
        })
        cy.get('*[name="method"][value="lookup_secret"]').click()
        cy.location('pathname').should('not.contain', 'login')

        // Going back to the settings UI we should see that the codes have been "used"
        cy.visit(settings)
        cy.get('button[name="lookup_secret_reveal"]').click()
        cy.getLookupSecrets().should((c) => {
          let newCodes = regenCodes
          newCodes[0] = 'Used'
          expect(c).to.eql(newCodes)
        })
      })

      it('should end up at login screen if trying to reveal without privileged session', () => {
        cy.shortPrivilegedSessionTime()
        cy.visit(settings)
        cy.get('button[name="lookup_secret_regenerate"]').click()
        cy.reauth({
          expect: { email },
          type: { email: email, password: password }
        })

        let codes
        cy.getLookupSecrets().should((c) => {
          codes = c
        })

        cy.shortPrivilegedSessionTime()
        cy.get('button[name="lookup_secret_confirm"]').click()
        cy.reauth({
          expect: { email },
          type: { email: email, password: password }
        })
        cy.expectSettingsSaved()

        cy.shortPrivilegedSessionTime()
        cy.get('button[name="lookup_secret_reveal"]').click()
        cy.reauth({
          expect: { email },
          type: { email: email, password: password }
        })
        cy.getLookupSecrets().should((c) => {
          expect(c).to.not.be.empty
        })
        cy.getSession({
          expectAal: 'aal2'
        })
      })

      it('should not show lookup as an option if not configured', () => {
        cy.visit(login + '?aal=aal2')
        cy.get('*[name="method"][value="totp"]').should('not.exist')
        cy.get('*[name="method"][value="lookup_secret"]').should('not.exist')
        cy.get('*[name="method"][value="password"]').should('not.exist')
        cy.get('h2').should('contain.text', 'Two-Factor Authentication')
      })
    })
  })
})

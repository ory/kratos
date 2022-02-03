import { gen, website } from '../../../helpers'
import { authenticator } from 'otplib'
import { routes as react } from '../../../helpers/react'
import { routes as express } from '../../../helpers/express'

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
        cy.clearAllCookies()
        email = gen.email()
        password = gen.password()

        cy.register({
          email,
          password,
          fields: { 'traits.website': website }
        })
        cy.longPrivilegedSessionTime()

        cy.useLaxAal()
      })

      it('should be be asked to sign in with 2fa if set up', () => {
        cy.visit(settings)
        cy.requireStrictAal()

        let secret
        cy.get('[data-testid="node/text/totp_secret_key/text"]').then(($e) => {
          secret = $e.text().trim()
        })
        cy.get('input[name="totp_code"]').then(($e) => {
          cy.wrap($e).type(authenticator.generate(secret))
        })
        cy.get('*[name="method"][value="totp"]').click()
        cy.expectSettingsSaved()
        cy.getSession({
          expectAal: 'aal2',
          expectMethods: ['password', 'totp']
        })

        cy.clearAllCookies()
        cy.visit(login)

        cy.get('input[name="password_identifier"]').type(email)
        cy.get('input[name="password"]').type(password)
        cy.submitPasswordForm()

        // MFA is now requested
        cy.location('pathname').should((loc) => {
          expect(loc).to.include('/login')
        })
        cy.shouldShow2FAScreen()

        // If we visit settings page we still end up at 2fa screen
        cy.visit(settings)
        cy.location('pathname').should((loc) => {
          expect(loc).to.include('/login')
        })

        cy.shouldShow2FAScreen()
        cy.get('input[name="totp_code"]').then(($e) => {
          cy.wrap($e).type(authenticator.generate(secret))
        })
        cy.get('*[name="method"][value="totp"]').click()
        cy.location('pathname').should((loc) => {
          expect(loc).to.oneOf(['/welcome', '/'])
        })
        cy.getSession({
          expectAal: 'aal2',
          expectMethods: ['password', 'totp']
        })
      })

      it('signin with 2fa and be redirected', () => {
        if (app !== 'express') {
          return
        }

        cy.visit(settings)
        cy.requireStrictAal()

        let secret
        cy.get('[data-testid="node/text/totp_secret_key/text"]').then(($e) => {
          secret = $e.text().trim()
        })
        cy.get('input[name="totp_code"]').then(($e) => {
          cy.wrap($e).type(authenticator.generate(secret))
        })
        cy.get('*[name="method"][value="totp"]').click()
        cy.expectSettingsSaved()
        cy.getSession({
          expectAal: 'aal2',
          expectMethods: ['password', 'totp']
        })

        cy.clearAllCookies()
        cy.visit(`${login}?return_to=https://www.ory.sh/`)

        cy.get('input[name="password_identifier"]').type(email)
        cy.get('input[name="password"]').type(password)
        cy.submitPasswordForm()

        // MFA is now requested
        cy.location('pathname').should((loc) => {
          expect(loc).to.include('/login')
        })
        cy.shouldShow2FAScreen()

        cy.location('pathname').should((loc) => {
          expect(loc).to.include('/login')
        })

        cy.shouldShow2FAScreen()
        cy.get('input[name="totp_code"]').then(($e) => {
          cy.wrap($e).type(authenticator.generate(secret))
        })
        cy.get('*[name="method"][value="totp"]').click()
        cy.url().should('eq', 'https://www.ory.sh/')
      })

      it('should go through several totp lifecycles', () => {
        cy.visit(settings)

        cy.get('[data-testid="node/text/totp_secret_key/text"]').should('exist')
        cy.get('img[data-testid="node/image/totp_qr"]').should('exist')

        // Set up TOTP
        let secret
        cy.get('[data-testid="node/text/totp_secret_key/text"]').then(($e) => {
          secret = $e.text().trim()
        })
        cy.get('input[name="totp_code"]').then(($e) => {
          cy.wrap($e).type(authenticator.generate(secret))
        })
        cy.get('*[name="method"][value="totp"]').click()
        cy.expectSettingsSaved()
        cy.get('[data-testid="node/text/totp_secret_key/text"]').should(
          'not.exist'
        )
        cy.get('img[data-testid="node/image/totp_qr"]').should('not.exist')
        cy.get('*[name="method"][value="totp"]').should('not.exist')
        cy.get('*[name="totp_unlink"]').should('exist')

        // Let's try to do 2FA
        cy.visit(login + '?aal=aal2&refresh=true')
        cy.location('pathname').should((loc) => {
          expect(loc).to.include('/login')
        })
        cy.get('*[name="method"][value="password"]').should('not.exist')

        // Typing a wrong code leaves us with an error message
        cy.get('*[name="totp_code"]').type('111111')
        cy.get('*[name="method"][value="totp"]').click()

        cy.get('[data-testid="ui/message/4000008"]').should(
          'contain.text',
          'The provided authentication code is invalid, please try again.'
        )
        cy.get('input[name="totp_code"]').then(($e) => {
          cy.wrap($e).type(authenticator.generate(secret))
        })
        cy.get('*[name="method"][value="totp"]').click()
        cy.location('pathname').should('not.contain', '/login')
        cy.getSession({
          expectAal: 'aal2',
          expectMethods: ['password', 'totp', 'totp']
        })

        // Going to settings and unlinking the device
        cy.visit(settings)
        cy.get('*[name="totp_unlink"]').click()
        cy.expectSettingsSaved()
        cy.get('[data-testid="node/text/totp_secret_key/text"]').should('exist')
        cy.get('img[data-testid="node/image/totp_qr"]').should('exist')
        cy.get('*[name="method"][value="totp"]').should('exist')
        cy.get('*[name="totp_unlink"]').should('not.exist')

        // 2FA should be gone
        cy.visit(login + '?aal=aal2&refresh=true')
        cy.location('pathname').should((loc) => {
          expect(loc).to.include('/login')
        })
        cy.get('*[name="method"][value="totp"]').should('not.exist')

        // Linking a new device works
        cy.visit(settings)
        let newSecret
        cy.get('[data-testid="node/text/totp_secret_key/text"]').then(($e) => {
          newSecret = $e.text().trim()
        })
        cy.get('input[name="totp_code"]').then(($e) => {
          cy.wrap($e).type(authenticator.generate(newSecret))
        })
        cy.get('*[name="method"][value="totp"]').click()
        cy.expectSettingsSaved()

        // Old secret no longer works in login
        cy.visit(login + '?aal=aal2&refresh=true')
        cy.location('pathname').should((loc) => {
          expect(loc).to.include('/login')
        })
        cy.get('input[name="totp_code"]').then(($e) => {
          cy.wrap($e).type(authenticator.generate(secret))
        })
        cy.get('*[name="method"][value="totp"]').click()
        cy.get('[data-testid="ui/message/4000008"]').should(
          'contain.text',
          'The provided authentication code is invalid, please try again.'
        )

        // But new one does!
        cy.get('input[name="totp_code"]').then(($e) => {
          cy.wrap($e).type(authenticator.generate(newSecret))
        })
        cy.get('*[name="method"][value="totp"]').click()
        cy.location('pathname').should((loc) => {
          expect(loc).to.not.include('/login')
        })

        cy.getSession({
          expectAal: 'aal2',
          expectMethods: ['password', 'totp', 'totp', 'totp', 'totp']
        })
      })

      it('should not show totp as an option if not configured', () => {
        cy.visit(login + '?aal=aal2')
        cy.location('pathname').should((loc) => {
          expect(loc).to.include('/login')
        })

        cy.get('*[name="method"][value="totp"]').should('not.exist')
        cy.get('*[name="method"][value="password"]').should('not.exist')
        cy.shouldShow2FAScreen()

        cy.get('[data-testid="logout-link"]').click()
        cy.location().should((loc) => {
          expect(loc.href).to.include('/login')
          expect(loc.search).to.not.include('aal')
          expect(loc.search).to.not.include('refresh')
        })
        cy.get('h2').should('contain.text', 'Sign In')
        cy.noSession()
      })

      it('should fail to set up totp if verify code is wrong', () => {
        cy.visit(settings)
        cy.get('input[name="totp_code"]').type('12345678')
        cy.get('*[name="method"][value="totp"]').click()
        cy.get('[data-testid="ui/message/4000008"]').should(
          'contain.text',
          'The provided authentication code is invalid, please try again.'
        )
      })
    })
  })
})

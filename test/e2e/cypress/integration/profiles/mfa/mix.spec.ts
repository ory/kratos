import { APP_URL, appPrefix, gen, website } from '../../../helpers'
import { authenticator } from 'otplib'
import { routes as react } from '../../../helpers/react'
import { routes as express } from '../../../helpers/express'

context('2FA with various methods', () => {
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
        cy.task('resetCRI')
      })
      let email = gen.email()
      let password = gen.password()

      beforeEach(() => {
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
        cy.task('sendCRI', {
          query: 'WebAuthn.disable',
          opts: {}
        })
      })

      it('should set up an use all mfa combinations', () => {
        cy.visit(settings)
        cy.task('sendCRI', {
          query: 'WebAuthn.enable',
          opts: {}
        }).then(() => {
          cy.task('sendCRI', {
            query: 'WebAuthn.addVirtualAuthenticator',
            opts: {
              options: {
                protocol: 'ctap2',
                transport: 'usb',
                hasResidentKey: true,
                hasUserVerification: true,
                isUserVerified: true
              }
            }
          }).then(() => {
            // Set up TOTP
            let secret
            cy.get(
              appPrefix(app) + '[data-testid="node/text/totp_secret_key/text"]'
            ).then(($e) => {
              secret = $e.text().trim()
            })
            cy.get('[name="totp_code"]').then(($e) => {
              cy.wrap($e).type(authenticator.generate(secret))
            })
            cy.get('[name="method"][value="totp"]').click()
            cy.expectSettingsSaved()

            // Set up lookup secrets
            cy.get('[name="lookup_secret_regenerate"]').click()
            let codes
            cy.getLookupSecrets().then((c) => {
              codes = c
            })
            cy.get('[name="lookup_secret_confirm"]').click()
            cy.expectSettingsSaved()

            // Set up WebAuthn
            cy.get('[name="webauthn_register_displayname"]').type('my-key')
            // We need a workaround here. So first we click, then we submit
            cy.clickWebAuthButton('register')
            cy.expectSettingsSaved()

            cy.visit(login + '?aal=aal2&refresh=true')
            cy.get('[name="totp_code"]').then(($e) => {
              cy.wrap($e).type(authenticator.generate(secret))
            })
            cy.get('[name="method"][value="totp"]').click()
            cy.getSession({
              expectAal: 'aal2',
              expectMethods: [
                'password',
                'totp',
                'webauthn',
                'lookup_secret',
                'totp'
              ]
            })

            // Use TOTP
            cy.visit(login + '?aal=aal2&refresh=true')
            cy.clickWebAuthButton('login')
            cy.getSession({
              expectAal: 'aal2',
              expectMethods: [
                'password',
                'totp',
                'webauthn',
                'lookup_secret',
                'totp',
                'webauthn'
              ]
            })

            // Use lookup
            cy.visit(login + '?aal=aal2&refresh=true')
            cy.get('[name="lookup_secret"]').then(($e) => {
              cy.wrap($e).type(codes[1])
            })
            cy.get('[name="method"][value="lookup_secret"]').click()
            cy.getSession({
              expectAal: 'aal2',
              expectMethods: [
                'password',
                'totp',
                'webauthn',
                'lookup_secret',
                'totp',
                'webauthn',
                'lookup_secret'
              ]
            })
          })
        })
      })
    })
  })
})

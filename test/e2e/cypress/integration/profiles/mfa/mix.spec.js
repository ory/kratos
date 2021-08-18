import {APP_URL, gen, website} from '../../../helpers'
import {authenticator} from 'otplib';

context('MFA Profile', () => {
  describe('Test MFA combinations', () => {
    before(() => {
      cy.useConfigProfile('mfa')
    })

    let email = gen.email()
    let password = gen.password()

    beforeEach(() => {
      cy.clearCookies()
      email = gen.email()
      password = gen.password()
      cy.registerApi({email, password, fields: {'traits.website': website}})
      cy.login({email, password})
      cy.longPrivilegedSessionTime()
      cy.task("sendCRI", {
        query: "WebAuthn.disable",
        opts: {},
      })
    })

    it('should set up an use all mfa combinations', () => {
      cy.visit(APP_URL + '/settings')
      cy.task("sendCRI", {
        query: "WebAuthn.enable",
        opts: {},
      }).then(() => {
        cy.task("sendCRI", {
          query: "WebAuthn.addVirtualAuthenticator",
          opts: {
            options: {
              protocol: "ctap2",
              transport: "usb",
              hasResidentKey: true,
              hasUserVerification: true,
              isUserVerified: true,
            },
          },
        }).then(() => {


          // Set up TOTP
          let secret
          cy.get('p[data-testid="text-totp_secret_key-content"]')
            .then(($e) => {
              secret = $e.text().trim();
            })
          cy.get('input[name="totp_code"]').then(($e) => {
            cy.wrap($e).type(authenticator.generate(secret))
          })
          cy.get('*[name="method"][value="totp"]').click()
          cy.get('form .messages .message').should('contain.text', 'Your changes have been saved!')

          // Set up lookup secrets
          cy.get('button[name="lookup_secret_regenerate"]').click()
          let codes
          cy.get('p[data-testid="text-lookup_secret_codes-content"]')
            .then(($e) => {
              codes = $e.text().trim().split(', ')
            })
          cy.get('button[name="lookup_secret_confirm"]').click()
          cy.get('form .messages .message').should('contain.text', 'Your changes have been saved!')

          // Set up WebAuthn
          cy.get('*[name="webauthn_register_displayname"]').type("my-key");
          // We need a workaround here. So first we click, then we submit
          cy.get('*[name="webauthn_register_trigger"]').click();
          cy.get('form .messages .message').should('contain.text', 'Your changes have been saved!')

          cy.visit(APP_URL + '/auth/login?aal=aal2&refresh=true')
          cy.get('input[name="totp_code"]').then(($e) => {
            cy.wrap($e).type(authenticator.generate(secret))
          })
          cy.get('*[name="method"][value="totp"]').click()
          cy.session({
            expectAal: 'aal2',
            expectMethods: ['password', 'totp']
          })

          // Use TOTP
          cy.visit(APP_URL + '/auth/login?aal=aal2&refresh=true')
          cy.get('button[name="webauthn_login_trigger"]').click()
          cy.session({
            expectAal: 'aal2',
            expectMethods: ['password', 'totp', 'webauthn']
          })

          // Use lookup
          cy.visit(APP_URL + '/auth/login?aal=aal2&refresh=true')
          cy.get('input[name="lookup_secret"]').then(($e) => {
            cy.wrap($e).type(codes[1])
          })
          cy.get('*[name="method"][value="lookup_secret"]').click()
          cy.session({
            expectAal: 'aal2',
            expectMethods: ['password', 'totp', 'webauthn', 'lookup_secret']
          })
        })
      })
    })
  })
})

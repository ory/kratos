import { APP_URL, gen, MOBILE_URL, website } from '../../../../helpers'
import { authenticator } from 'otplib'

context('Mobile Profile', () => {
  describe('TOTP 2FA Flow', () => {
    before(() => {
      cy.useConfigProfile('mobile')
    })

    describe('password', () => {
      const email = gen.email()
      const password = gen.password()

      before(() => {
        cy.registerApi({
          email,
          password,
          fields: { 'traits.website': website }
        })
      })

      beforeEach(() => {
        cy.loginMobile({ email, password })
        cy.visit(MOBILE_URL + '/Settings')
      })

      it('should be able to lifecycle through TOTP flows', () => {
        cy.get('*[data-testid="field/totp_qr"]').should('exist')
        cy.get('*[data-testid="field/totp_code"]').should('exist')

        // Set up TOTP with invalid key
        cy.get('*[data-testid="field/totp_code"]').type('111111')
        cy.get('*[data-testid="field/method/totp"]').click()
        cy.get('*[data-testid="field/totp_code"]').should(
          'contain.text',
          'The provided authentication code is invalid, please try again.'
        )

        // Set up TOTP with valid key
        let secret
        cy.get('*[data-testid="field/totp_secret_key/text"]').then(($e) => {
          secret = $e.text().trim()
        })
        cy.get('*[data-testid="field/totp_code"]').then(($e) => {
          cy.wrap($e).type(authenticator.generate(secret))
        })
        cy.get('*[data-testid="field/method/totp"]').click()

        cy.get('*[data-testid="form-messages"]').should(
          'contain.text',
          'Your changes have been saved!'
        )

        // Form should look different now
        cy.get('*[data-testid="field/totp_secret_key/text"]').should(
          'not.exist'
        )
        cy.get('*[data-testid="field/totp_code"]').should('not.exist')
        cy.get('*[data-testid="field/totp_qr"]').should('not.exist')
        cy.get('*[data-testid="field/totp_unlink/true"]').should('exist')

        // Lets sign in
        cy.visit(MOBILE_URL + '/Login?aal=aal2&refresh=true')

        // First use a wrong code
        cy.get('*[data-testid="field/totp_code"]').type('111111')
        cy.get('*[data-testid="field/method/totp"]').click()
        cy.get('*[data-testid="form-messages"]').should(
          'contain.text',
          'The provided authentication code is invalid, please try again.'
        )

        // Use the correct code
        cy.get('*[data-testid="field/totp_code"]').then(($e) => {
          cy.wrap($e).type(authenticator.generate(secret))
        })
        cy.get('*[data-testid="field/method/totp"]').click()

        // We have AAL now
        cy.get('[data-testid="session-content"]').should('contain', 'aal2')
        cy.get('[data-testid="session-content"]').should('contain', 'totp')

        // Go back to settings and unlink
        cy.visit(MOBILE_URL + '/Settings')
        cy.get('*[data-testid="field/totp_unlink/true"]').click()
        cy.get('*[data-testid="field/totp_unlink/true"]').should('not.exist')
        cy.get('*[data-testid="field/totp_qr"]').should('exist')
        cy.get('*[data-testid="field/totp_code"]').should('exist')
      })
    })
  })
})

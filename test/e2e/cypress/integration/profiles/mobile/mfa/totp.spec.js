import {APP_URL, gen, MOBILE_URL, website} from '../../../../helpers'
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
          fields: {'traits.website': website}
        })
      })

      beforeEach(() => {
        cy.loginMobile({email, password})
        cy.visit(MOBILE_URL + '/Settings')
      })

      it('should be able to set up TOTP', () => {

        cy.get('p[data-testid="text-totp_secret_key-content"]').should('exist')
        cy.get('img[data-testid="text-totp_qr"]').should('exist')

        // Set up TOTP
        let secret
        cy.get('p[data-testid="text-totp_secret_key-content"]').then(($e) => {
          secret = $e.text().trim()
        })
        cy.get('input[name="totp_code"]').then(($e) => {
          cy.wrap($e).type(authenticator.generate(secret))
        })
        cy.get('*[name="method"][value="totp"]').click()
        cy.get('form .messages .message').should(
          'contain.text',
          'Your changes have been saved!'
        )
        cy.get('p[data-testid="text-totp_secret_key-content"]').should(
          'not.exist'
        )
        cy.get('img[data-testid="text-totp_qr"]').should('not.exist')
        cy.get('*[name="method"][value="totp"]').should('not.exist')
        cy.get('*[name="totp_unlink"]').should('exist')

      })
    })
  })
})

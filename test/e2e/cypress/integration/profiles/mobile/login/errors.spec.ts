import { gen, MOBILE_URL } from '../../../../helpers'

context('Mobile Profile', () => {
  describe('Login Flow Errors', () => {
    before(() => {
      cy.clearAllCookies()
      cy.useConfigProfile('mobile')
    })

    beforeEach(() => {
      cy.visit(MOBILE_URL + '/Login')
    })

    describe('shows validation errors when invalid signup data is used', () => {
      it('should show an error when the password_identifier is missing', () => {
        cy.get('input[data-testid="password"]').type(gen.password())

        cy.get('div[data-testid="submit-form"]').click()

        cy.get('*[data-testid="field/password_identifier"]').should(
          'contain.text',
          'Property password_identifier is missing.'
        )

        cy.get('*[data-testid="field/password"]').should(
          'not.contain.text',
          'Property password is missing.'
        )
      })

      it('should show an error when the password is missing', () => {
        const email = gen.email()
        cy.get('input[data-testid="password_identifier"]')
          .type(email)
          .should('have.value', email)

        cy.get('div[data-testid="submit-form"]').click()

        cy.get('*[data-testid="field/password"]').should(
          'contain.text',
          'Property password is missing.'
        )
      })

      it('should show fail to sign in', () => {
        cy.get('input[data-testid="password_identifier"]').type(gen.email())
        cy.get('input[data-testid="password"]').type(gen.password())
        cy.get('*[data-testid="submit-form"]').click()
        cy.get('*[data-testid="form-messages"]').should(
          'contain.text',
          'credentials are invalid'
        )
      })
    })
  })
})

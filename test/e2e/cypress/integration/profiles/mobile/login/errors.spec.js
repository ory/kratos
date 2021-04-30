import {gen, MOBILE_URL} from '../../../../helpers'

context('Login Flow Errors', () => {
  beforeEach(() => {
    cy.visit(MOBILE_URL + "/Login")
  })

  describe('shows validation errors when invalid signup data is used', () => {
    it('should show an error when the password_identifier is missing', () => {

      cy.get('input[data-testid="password"]')
        .type(gen.password())

      cy.get('div[data-testid="submit-form"]').click()

      cy.get('*[data-testid="field/password_identifier"]').should(
        'contain.text',
        'length must be >= 1, but got 0'
      )

      cy.get('*[data-testid="field/password"]').should(
        'not.contain.text',
        'length must be >= 1, but got 0'
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
        'length must be >= 1, but got 0'
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

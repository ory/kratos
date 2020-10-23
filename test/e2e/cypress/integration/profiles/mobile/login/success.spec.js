import {gen, MOBILE_URL, website} from '../../../../helpers'

context('Login Flow Success', () => {
  const email = gen.email()
  const password = gen.password()

  beforeEach(() => {
    cy.registerApi({email, password, fields: {'traits.website': website}})
    cy.visit(MOBILE_URL + "/Login")
  })

  it('should sign up and be logged in', () => {
    cy.get('input[data-testid="identifier"]').type(email)
    cy.get('input[data-testid="password"]').type(password)
    cy.get('div[data-testid="submit-form"]').click()

    cy.get('[data-testid="session-content"]').should('contain', email)
    cy.get('[data-testid="session-token"]').should('not.be.empty')
  })
})

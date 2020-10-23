import { MOBILE_URL, gen, website } from '../../../../helpers'

context('Login Flow Success', () => {
  const email = gen.email()
  const password = gen.password()

  beforeEach(() => {
    cy.visit(MOBILE_URL + "/Registration")
  })

  it('should sign up and be logged in', () => {
    cy.get('input[data-testid="traits.email"]').type(email)
    cy.get('input[data-testid="password"]').type(password)
    cy.get('input[data-testid="traits.website"]').type(website)
    cy.get('div[data-testid="submit-form"]').click()

    cy.get('[data-testid="session-content"]').should('contain', email)
    cy.get('[data-testid="session-token"]').should('not.be.empty')
  })
})

import { gen, SPA_URL, website } from '../../../../helpers'

context('Mobile Profile', () => {
  describe('Login Flow Success', () => {
    before(() => {
      cy.useConfigProfile('spa')
    })

    const email = gen.email()
    const password = gen.password()

    before(() => {
      cy.registerApi({ email, password, fields: { 'traits.website': website } })
    })

    beforeEach(() => {
      cy.visit(SPA_URL + '/login')
    })

    it('should sign up and be logged in', () => {
      cy.get('input[name="password_identifier"]').type(email)
      cy.get('input[name="password"]').type(password)
      cy.get('button[value="password"][name="method"]').click()

      cy.get('[data-testid="session-content"]').should('contain', email)
    })
  })
})

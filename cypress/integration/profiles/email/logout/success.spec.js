import { APP_URL, gen, password, website } from '../../../../helpers'

context('Login', () => {
  const email = gen.email()
  const password = gen.password()

  before(() => {
    cy.register({ email, password, fields: { 'traits.website': website } })
  })

  beforeEach(() => {
    cy.visit(APP_URL + '/')
  })

  it('should sign out and be able to sign in again', () => {
    cy.get('a[href*="logout"]').click()

    cy.noSession()

    cy.url().should('include', '/auth/login')
  })
})

import { APP_URL, gen, password, website } from '../../../../helpers'

context('Login', () => {
  const email = gen.email()

  before(() => {
    cy.registerOidc({email, website})
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

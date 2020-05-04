import { APP_URL, gen, website } from '../../../../helpers'

context('Login', () => {
  const email = gen.email()
  const password = gen.password()

  beforeEach(() => {
    cy.clearCookies()
    cy.visit(APP_URL + '/auth/login')
  })

  it('should sign up and then log in', () => {
    cy.get('#login-oidc button[value="hydra"]').click()

    cy.get('input[type="email"]').type('foo@bar.com')
    cy.get('input[type="password"]').type('foobar')
    cy.get('#accept').click()

    cy.get('#openid').click()
    cy.get('#accept').click()
  })
})

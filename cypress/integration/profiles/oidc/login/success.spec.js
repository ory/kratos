import { APP_URL, gen, website } from '../../../../helpers'

context('Login', () => {
  beforeEach(() => {
    cy.clearCookies()
  })

  it('should be able to sign up, sign out, and then sign in', () => {
    const email = gen.email()

    cy.registerOidc({email, website})
    cy.get('a[href*="logout"]').click()
    cy.noSession()
    cy.loginOidc({email})
  })
})

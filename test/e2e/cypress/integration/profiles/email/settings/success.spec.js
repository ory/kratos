import { APP_URL, gen, website } from '../../../../helpers'

context('Settings', () => {
  let email = gen.email()
  let password = gen.password()

  const up = (value) => `not-${value}`
  const down = (value) => value.replace(/not-/, '')

  before(() => {
    cy.register({ email, password, fields: { 'traits.website': website } })
  })

  beforeEach(() => {
    cy.clearCookies()
    cy.login({ email, password })
    cy.visit(APP_URL + '/settings')
  })

  it('shows all settings forms', () => {
    cy.get('#user-profile h3').should('contain.text', 'Profile')
    cy.get('#user-profile input[name="traits.email"]').should(
      'contain.value',
      email
    )
    cy.get('#user-profile input[name="traits.website"]').should(
      'contain.value',
      website
    )

    cy.get('#user-password h3').should('contain.text', 'Password')
    cy.get('#user-password input[name="password"]').should('be.empty')
  })

  describe('password', () => {
    it('modifies the password with privileged session', () => {
      // Once input weak password to test which error message is cleared after updating successfully
      cy.get('#user-password input[name="password"]').clear().type('123')
      cy.get('#user-password button[type="submit"]').click()
      cy.get('.container').should(
        'not.contain.text',
        'Your changes have been saved!'
      )
      cy.get('.container').should(
        'contain.text',
        'The password can not be used'
      )
      cy.get('#user-password input[name="password"]').should('be.empty')

      password = up(password)
      cy.get('#user-password input[name="password"]').clear().type(password)
      cy.get('#user-password button[type="submit"]').click()
      cy.get('.container').should(
        'contain.text',
        'Your changes have been saved!'
      )
      cy.get('.container').should(
        'not.contain.text',
        'The password can not be used'
      )
      cy.get('#user-password input[name="password"]').should('be.empty')
    })

    it('is unable to log in with the old password', () => {
      cy.clearCookies()
      cy.login({
        email: email,
        password: down(password),
        expectSession: false,
      })
    })

    it('modifies the password with an unprivileged session', () => {
      password = up(password)
      cy.get('#user-password input[name="password"]').clear().type(password)
      cy.waitForPrivilegedSessionToExpire() // wait for the privileged session to time out
      cy.get('#user-password button[type="submit"]').click()

      cy.reauth({ expect: { email }, type: { password: down(password) } })

      cy.url().should('include', '/settings')
      cy.get('.container').should(
        'contain.text',
        'Your changes have been saved!'
      )
      cy.get('#user-password input[name="password"]').should('be.empty')
    })
  })

  describe('profile', () => {
    it('modifies an unprotected trait', () => {
      cy.get('#user-profile input[name="traits.website"]')
        .clear()
        .type('https://github.com/ory')
      cy.get('#user-profile button[type="submit"]').click()
      cy.get('.container').should(
        'contain.text',
        'Your changes have been saved!'
      )
      cy.get('#user-profile input[name="traits.website"]').should(
        'contain.value',
        'https://github.com/ory'
      )
    })

    it('modifies a protected trait with privileged session', () => {
      email = up(email)
      cy.get('#user-profile input[name="traits.email"]').clear().type(email)
      cy.get('#user-profile button[type="submit"]').click()
      cy.get('.container').should(
        'contain.text',
        'Your changes have been saved!'
      )
      cy.get('#user-profile input[name="traits.email"]').should(
        'contain.value',
        email
      )
    })

    it('is unable to log in with the old email', () => {
      cy.clearCookies()
      cy.visit(APP_URL + '/auth/login')
      cy.login({ email: down(email), password, expectSession: false })
    })

    it('modifies a protected trait with unprivileged session', () => {
      email = up(email)
      cy.get('#user-profile input[name="traits.email"]').clear().type(email)
      cy.waitForPrivilegedSessionToExpire() // wait for the privileged session to time out
      cy.get('#user-profile button[type="submit"]').click()

      cy.reauth({ expect: { email: down(email) }, type: { password } })

      cy.url().should('include', '/settings')
      cy.get('.container').should(
        'contain.text',
        'Your changes have been saved!'
      )
      cy.get('#user-profile input[name="traits.email"]').should(
        'contain.value',
        email
      )
    })
  })
})

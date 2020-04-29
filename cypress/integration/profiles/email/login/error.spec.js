import { APP_URL, identity } from '../../../../helpers'

context('Login', () => {
  beforeEach(() => {
    cy.visit(APP_URL + '/auth/login')
  })

  it('fails when CSRF cookies are missing', () => {
    cy.clearCookies()

    cy.get('input[name="identifier"]').type('i-do-not-exist')
    cy.get('input[name="password"]').type('invalid-password')

    cy.get('button[type="submit"]').click()

    // FIXME https://github.com/ory/kratos/issues/91
    cy.get('html').should('contain.text', 'CSRF token is missing or invalid')
  })

  describe('shows validation errors when invalid signup data is used', () => {
    it('should show an error when the identifier is missing', () => {
      cy.get('button[type="submit"]').click()
      cy.get('.form-errors .message').should(
        'contain.text',
        'missing properties: identifier'
      )
    })

    it('should show an error when the password is missing', () => {
      cy.get('input[name="identifier"]')
        .type(identity)
        .should('have.value', identity)

      cy.get('button[type="submit"]').click()
      cy.get('.form-errors .message').should(
        'contain.text',
        'missing properties: password'
      )
    })

    it('should show fail to sign in', () => {
      cy.get('input[name="identifier"]').type('i-do-not-exist')
      cy.get('input[name="password"]').type('invalid-password')

      cy.get('button[type="submit"]').click()
      cy.get('.form-errors .message').should(
        'contain.text',
        'credentials are invalid'
      )
    })
  })
})

import { APP_URL, gen } from '../../../../helpers'

context('Login Flow Error', () => {
  beforeEach(() => {
    cy.visit(APP_URL + '/auth/login')
  })

  it('fails when CSRF cookies are missing', () => {
    cy.clearCookies()

    cy.get('input[name="password_identifier"]').type('i-do-not-exist')
    cy.get('input[name="password"]').type('invalid-password')

    cy.get('button[type="submit"]').click()

    // FIXME https://github.com/ory/kratos/issues/91
    cy.get('html').should('contain.text', 'missing or invalid csrf_token value')
  })

  describe('shows validation errors when invalid signup data is used', () => {
    it('should show an error when the identifier is missing', () => {
      cy.get('button[type="submit"]').click()
      cy.get('.messages .message').should(
        'contain.text',
        'length must be >= 1, but got 0'
      )
    })

    it('should show an error when the password is missing', () => {
      const identity = gen.email()
      cy.get('input[name="password_identifier"]')
        .type(identity)
        .should('have.value', identity)

      cy.get('button[type="submit"]').click()
      cy.get('.messages .message').should(
        'contain.text',
        'length must be >= 1, but got 0'
      )
    })

    it('should show fail to sign in', () => {
      cy.get('input[name="password_identifier"]').type('i-do-not-exist')
      cy.get('input[name="password"]').type('invalid-password')

      cy.get('button[type="submit"]').click()
      cy.get('.messages .message').should(
        'contain.text',
        'credentials are invalid'
      )
    })
  })
})

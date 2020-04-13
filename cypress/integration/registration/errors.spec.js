const {APP_URL,identity,  password} = require("../../helpers")

context('Registration', () => {
  beforeEach(() => {
    cy.visit(APP_URL + '/auth/registration')
  })

  describe("show error warnings when invalid signup data is used", () => {
    it('should show an error when the password has leaked before', () => {
      cy.get('input[name="traits.email"]').type(identity).should('have.value', identity)
      cy.get('input[name="password"]').type('123456').should('have.value', '123456')

      cy.get('button[type="submit"]').click()
      cy.get('.form-errors .message').should('contain.text', 'data breaches')
    })

    it('should show an error when the password is to similar', () => {
      cy.get('input[name="traits.email"]').type(identity)
      cy.get('input[name="password"]').type(identity)

      cy.get('button[type="submit"]').click()
      cy.get('.form-errors .message').should('contain.text', 'to similar')
    })

    it('should show an error when the password is empty', () => {
      cy.get('input[name="traits.email"]').type(identity)

      cy.get('button[type="submit"]').click()
      cy.get('.form-errors .message').should('contain.text', 'missing')
    })

    it('should show an error when the email is empty', () => {
      cy.get('input[name="password"]').type(password)

      cy.get('button[type="submit"]').click()
      cy.get('.form-errors .message').should('contain.text', 'valid "email"')
    })

    it('should show an error when the email is not an email', () => {
      cy.get('input[name="password"]').type("not-an-email")

      cy.get('button[type="submit"]').click()
      cy.get('.form-errors .message').should('contain.text', 'valid "email"')
    })

    it('should show a missing indicator if no fields are set', () => {
      cy.get('button[type="submit"]').click()
      cy.get('.form-errors .message').should('contain.text', 'missing properties')
    })
  })
})

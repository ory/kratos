const {APP_URL, password} = require("../../helpers")

context('Registration', () => {
  beforeEach(() => {
    cy.visit(APP_URL + '/auth/registration')
  })

  it('should sign up and receive a verification email', () => {
    const identity = Math.random().toString(36).substring(7) + "@" + Math.random().toString(36).substring(7)
    cy.get('input[name="traits.email"]').type(identity)
    cy.get('input[name="password"]').type(password)

    cy.get('button[type="submit"]').click()
    cy.get('pre').should('contain.text', identity)
    cy.get('.greeting').should('contain.text', "Welcome back")
  })
})

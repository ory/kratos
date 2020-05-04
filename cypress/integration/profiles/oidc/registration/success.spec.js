import {APP_URL} from '../../../../helpers'

context('Register', () => {
  beforeEach(() => {
    cy.clearCookies()
    cy.visit(APP_URL + '/auth/registration')
  })

  it('should be able to sign up with incomplete data and finally be signed in', () => {
    cy.get('button[value="hydra"]').click()

    cy.get('input[type="email"]').type('foo@bar.com')
    cy.get('input[type="password"]').type('foobar')
    cy.get('input[name="remember"]').click()
    cy.get('#accept').click()

    cy.get('#openid').click()
    cy.get('#offline').click()
    cy.get('input[name="remember"]').click()
    cy.get('#accept').click()

    cy.get('#registration-password').should('not.exist');
    cy.get('#registration-oidc input[name="traits.email"]').should('have.value', 'foo@bar.com')
    cy.get('#registration-oidc form > *:last-child').should('have.attr', 'name', 'provider')
    cy.get('.form-errors .message').should('contain.text', 'missing properties: "website"')
    cy.get('#registration-oidc input[name="traits.website"]').type("http://s")

    cy.get('button[value="hydra"]').click()

    cy.get('#registration-password').should('not.exist');
    cy.get('#registration-oidc input[name="traits.email"]').should('have.value', 'foo@bar.com')
    cy.get('#registration-oidc form > *:last-child').should('have.attr', 'name', 'provider')
    cy.get('.form-errors .message').should('contain.text', 'length must be >= 10')
    cy.get('#registration-oidc input[name="traits.website"]').type("https://www.ory.sh/")

    cy.get('button[value="hydra"]').click()
  })
})

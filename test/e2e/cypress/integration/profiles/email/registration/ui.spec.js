import { APP_URL } from '../../../../helpers'

context('Email Profile', () => {
  describe('Registration Flow UI', () => {
    before(() => {
      cy.useConfigProfile('email')
    })

    beforeEach(() => {
      cy.visit(APP_URL + '/auth/registration')
    })

    describe('use ui elements', () => {
      it('should use the json schema titles', () => {
        cy.get('input[name="traits.email"]').siblings('span').should('contain.text', 'Your E-Mail')
        cy.get('input[name="traits.website"]').siblings('span').should('contain.text', 'Your website')
        cy.get('button[value="password"]').should('contain.text', 'Sign up')
    })

    it('clicks the visibility toggle to show the password', () => {
      cy.get('input[name="password"]').type('some password')
      cy.get('input[name="password"]').should('have.prop', 'type', 'password')
      cy.get('.password-visibility-toggle').click()
      cy.get('input[name="password"]').should('have.prop', 'type', 'text')
      cy.get('.password-visibility-toggle').click()
      cy.get('input[name="password"]').should('have.prop', 'type', 'password')
    })

      it('clicks the log in link', () => {
        cy.get('a[href*="auth/login"]').click()
        cy.location('pathname').should('include', 'auth/login')
        cy.location('search').should('not.be.empty', 'request')
      })
    })
  })
})

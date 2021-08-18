import {APP_URL, gen, website} from '../../../helpers'

context('MFA Profile', () => {
  describe('Check UI', () => {
    before(() => {
      cy.useConfigProfile('mfa')
    })

    const email = gen.email()
    const password = gen.password()

    before(() => {
      cy.registerApi({email, password, fields: {'traits.website': website}})
    })

    beforeEach(() => {
      cy.clearCookies()
      cy.login({email, password})
      cy.visit(APP_URL + '/settings')
    })

    it('shows all settings forms', () => {
      cy.get('p').should('contain.text', 'Profile')
      cy.get('input[name="traits.email"]').should('contain.value', email)
      cy.get('input[name="traits.website"]').should('contain.value', website)

      cy.get('p').should('contain.text', 'Password')
      cy.get('p').should('contain.text', 'This is your authenticator app secret')
      cy.get('button').should('contain.text', 'Generate new backup recovery codes')
      cy.get('button').should('contain.text', 'Add security key')
    })
  })
})

const {APP_URL, gen} = require('../../../../helpers')

context('Email Profile', () => {
  describe('Settings Flow UI', () => {
    before(() => {
      cy.useConfigProfile('email')
    })

    beforeEach(() => {
      const identity = gen.identity()
      cy.register({
        ...identity,
        fields: {'traits.website': 'https://www.ory.sh/'},
      })
      cy.login(identity)
      cy.visit(APP_URL)
    })

  describe('use ui elements', () => {
    it('should use the json schema titles', () => {
      cy.get('a[href*="settings"]').click()
      cy.get('input[name="traits.email"]').siblings('span').should('contain.text', 'Your E-Mail')
      cy.get('input[name="traits.website"]').siblings('span').should('contain.text', 'Your website')
      cy.get('input[name="password"]').siblings('span').should('contain.text', 'Password')
      cy.get('button[value="profile"]').should('contain.text', 'Save')
      cy.get('button[value="password"]').should('contain.text', 'Save')
    })

    it('clicks the settings link', () => {
      cy.get('a[href*="settings"]').click()
      cy.location('pathname').should('include', 'settings')
      cy.location('search').should('not.be.empty', 'request')
    })
  })
  })
})

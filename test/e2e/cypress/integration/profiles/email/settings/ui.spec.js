const { APP_URL, gen } = require('../../../../helpers')
context('Settings Flow UI', () => {
  beforeEach(() => {
    const identity = gen.identity()
    cy.register({
      ...identity,
      fields: { 'traits.website': 'https://www.ory.sh/' },
    })
    cy.login(identity)
    cy.visit(APP_URL)
  })

  describe('use ui elements', () => {
    it('clicks the settings link', () => {
      cy.get('a[href*="settings"]').click()
      cy.location('pathname').should('include', 'settings')
      cy.location('search').should('not.be.empty', 'request')
    })
  })
})

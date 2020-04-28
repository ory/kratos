context('Settings', () => {
  beforeEach(() => {
    cy.register({fields: {"traits.website": "https://www.ory.sh/"}})
  })

  describe("use ui elements", () => {
    it('clicks the settings link', () => {
      cy.get('a[href*="settings"]').click()
      cy.location('pathname').should('include', 'settings')
      cy.location('search').should('not.be.empty', 'request')
    })
  })
})

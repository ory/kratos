import { APP_URL, gen, website } from '../../../../helpers'

context('OIDC Profile', () => {
  describe('Register', () => {
    before(() => {
      cy.useConfigProfile('oidc')
    })

    beforeEach(() => {
      cy.clearCookies()
      cy.visit(APP_URL + '/auth/registration')
    })

    const shouldSession = (email) => (session) => {
      const { identity } = session
      expect(identity.id).to.not.be.empty
      expect(identity.schema_id).to.equal('default')
      expect(identity.schema_url).to.equal(`${APP_URL}/schemas/default`)
      expect(identity.traits.website).to.equal(website)
      expect(identity.traits.email).to.equal(email)
    }

    it('should be able to sign up with incomplete data and finally be signed in', () => {
      const email = gen.email()

      cy.registerOidc({ email, expectSession: false })

      cy.get('#registration-password').should('not.exist')
      cy.get('input[name="traits.email"]').should('have.value', email)
      cy.get('form > *:last-child').should('have.attr', 'name', 'provider')
      cy.get('.messages .message').should(
        'contain.text',
        'Property website is missing'
      )

      cy.get('button[name="provider"]')
        .should('have.length', 1)
        .should('have.value', 'hydra')
        .should('contain.text', 'Continue')

      cy.get('input[name="traits.website"]').next('span').contains('Website')
      cy.get('input[name="traits.consent"][type="checkbox"]')
        .next('span')
        .contains('Consent')
      cy.get('input[name="traits.consent"][type="checkbox"]').click()
      cy.get('input[name="traits.newsletter"][type="checkbox"]')
        .next('span')
        .contains('Newsletter')
      cy.get('input[name="traits.newsletter"][type="checkbox"]').click()
      cy.get('input[name="traits.website"]').type('http://s')

      cy.get('button[value="hydra"]').click()

      cy.get('#registration-password').should('not.exist')
      cy.get('input[name="traits.email"]').should('have.value', email)
      cy.get('form > *:last-child').should('have.attr', 'name', 'provider')
      cy.get('.messages .message').should(
        'contain.text',
        'length must be >= 10'
      )
      cy.get('input[name="traits.website"]')
        .should('have.value', 'http://s')
        .clear()
        .type(website)

      cy.get('input[name="traits.consent"]').should('be.checked')
      cy.get('input[name="traits.newsletter"]').should('be.checked')

      cy.get('input[name="traits.website"]').next('span').contains('Website')
      cy.get('input[name="traits.consent"][type="checkbox"]')
        .next('span')
        .contains('Consent')
      cy.get('input[name="traits.newsletter"][type="checkbox"]')
        .next('span')
        .contains('Newsletter')

      cy.get('button[value="hydra"]').click()

      cy.session().should((session) => {
        shouldSession(email)(session)
        expect(session.identity.traits.consent).to.equal(true)
      })
    })

    it('should be able to sign up with complete data', () => {
      const email = gen.email()

      cy.registerOidc({ email, website })
      cy.session().should(shouldSession(email))
    })
    it('should be able to convert a sign up flow to a sign in flow', () => {
      const email = gen.email()

      cy.registerOidc({ email, website })
      cy.get('a[href*="logout"]').click()
      cy.noSession()
      cy.visit(APP_URL + '/auth/registration')
      cy.get('button[value="hydra"]').click()

      cy.session().should(shouldSession(email))
    })

    it('should be able to convert a sign in flow to a sign up flow', () => {
      const email = gen.email()
      cy.visit(APP_URL + '/auth/login')
      cy.get('button[value="hydra"]').click()
      cy.get('#username').clear().type(email)
      cy.get('#remember').click()
      cy.get('#accept').click()
      cy.get('input[name="scope"]').each(($el) => cy.wrap($el).click())
      cy.get('#remember').click()
      cy.get('#accept').click()

      cy.get('.messages .message').should(
        'contain.text',
        'Property website is missing'
      )
      cy.get('input[name="traits.website"]').type('http://s')
      cy.get('button[value="hydra"]').click()

      cy.get('.messages .message').should(
        'contain.text',
        'length must be >= 10'
      )
      cy.get('input[name="traits.website"]')
        .should('have.value', 'http://s')
        .clear()
        .type(website)
      cy.get('button[value="hydra"]').click()

      cy.session().should(shouldSession(email))
    })
  })
})

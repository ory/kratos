import { APP_URL, gen } from '../../../../helpers'

context('Registration Flow Success', () => {
  beforeEach(() => {
    cy.visit(APP_URL + '/auth/registration')
  })

  it('should sign up and be logged in', () => {
    const email = gen.email()
    const password = gen.password()
    const website = 'https://www.ory.sh/'
    cy.get('input[name="traits.email"]').type(email)
    cy.get('input[name="traits.website').type(website)
    cy.get('input[name="password"]').type(password)

    cy.get('button[type="submit"]').click()
    cy.get('pre').should('contain.text', email)
    cy.get('.greeting').should('contain.text', 'Welcome back')

    cy.session().should((session) => {
      const { identity } = session
      expect(identity.id).to.not.be.empty
      expect(identity.verifiable_addresses).to.be.undefined
      expect(identity.schema_id).to.equal('default')
      expect(identity.schema_url).to.equal(`${APP_URL}/schemas/default`)
      expect(identity.traits.website).to.equal(website)
      expect(identity.traits.email).to.equal(email)
    })
  })
})

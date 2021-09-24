import { APP_URL, gen, website } from '../../../../helpers'

context('Email Profile', () => {
  describe('Login Flow Success', () => {
    before(() => {
      cy.useConfigProfile('email')
    })

    const email = gen.email()
    const password = gen.password()

    before(() => {
      cy.registerApi({ email, password, fields: { 'traits.website': website } })
    })

    beforeEach(() => {
      cy.clearCookies()
      cy.visit(APP_URL + '/auth/login')
    })

    it('should sign up and be logged in', () => {
      cy.get('input[name="password_identifier"]').type(email)
      cy.get('input[name="password"]').type(password)
      cy.get('button[type="submit"]').click()

      cy.getSession().should((session) => {
        const { identity } = session
        expect(identity.id).to.not.be.empty
        expect(identity.schema_id).to.equal('default')
        expect(identity.schema_url).to.equal(`${APP_URL}/schemas/default`)
        expect(identity.traits.website).to.equal(website)
        expect(identity.traits.email).to.equal(email)
      })
    })

    it('should sign in with case insensitive identifier', () => {
      cy.get('input[name="password_identifier"]').type(email.toUpperCase())
      cy.get('input[name="password"]').type(password)
      cy.get('button[type="submit"]').click()

      cy.getSession().should((session) => {
        const { identity } = session
        expect(identity.id).to.not.be.empty
        expect(identity.schema_id).to.equal('default')
        expect(identity.schema_url).to.equal(`${APP_URL}/schemas/default`)
        expect(identity.traits.website).to.equal(website)
        expect(identity.traits.email).to.equal(email)
      })
    })
  })
  describe('Login Flow Success with return_to url after flow expires', () => {
    before(() => {
      cy.useConfigProfile('email')
    })

    const email = gen.email()
    const password = gen.password()

    before(() => {
      cy.registerApi({ email, password, fields: { 'traits.website': website } })
    })

    beforeEach(() => {
      cy.shortLoginLifespan()
      cy.browserReturnUrlOry()
      cy.clearCookies()
      cy.visit(
        APP_URL + '/self-service/login/browser?return_to=https://www.ory.sh/'
      )
    })

    it('should redirect to return_to after flow expires', () => {
      cy.wait(105)
      cy.get('input[name="password_identifier"]').type(email.toUpperCase())
      cy.get('input[name="password"]').type(password)

      cy.longLoginLifespan()
      cy.get('button[type="submit"]').click()
      cy.get('.messages .message').should(
        'contain.text',
        'The login flow expired'
      )

      // try again with long lifespan set
      cy.get('input[name="password_identifier"]').type(email.toUpperCase())
      cy.get('input[name="password"]').type(password)
      cy.get('button[type="submit"]').click()

      // check that redirection has happened
      cy.url().should('eq', 'https://www.ory.sh/')
    })
  })
})

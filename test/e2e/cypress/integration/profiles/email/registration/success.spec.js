import { APP_URL, gen } from '../../../../helpers'

context('Email Profile', () => {
  describe('Registration Flow Success', () => {
    before(() => {
      cy.useConfigProfile('email')
    })

    beforeEach(() => {
      cy.visit(APP_URL + '/auth/registration')
    })

    it('should sign up and be logged in', () => {
      const email = gen.email()
      const password = gen.password()
      const website = 'https://www.ory.sh/'
      cy.get('input[name="traits"]').should('not.exist')
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
  describe('Registration Flow Success with return_to url', () => {
    before(() => {
      cy.useConfigProfile('email')
    })

    beforeEach(() => {
      cy.shortRegisterLifespan()
      cy.browserReturnUrlOry()
      cy.visit(
        APP_URL +
          '/self-service/registration/browser?return_to=https://www.ory.sh/'
      )
    })

    it('should redirect to return_to after flow expires', () => {
      //wait for flow to expire
      cy.wait(105)
      const email = gen.email()
      const password = gen.password()
      const website = 'https://www.ory.sh/'
      cy.get('input[name="traits"]').should('not.exist')
      cy.get('input[name="traits.email"]').type(email)
      cy.get('input[name="traits.website').type(website)
      cy.get('input[name="password"]').type(password)

      cy.longRegisterLifespan()
      cy.get('button[type="submit"]').click()
      cy.get('.messages .message').should(
        'contain.text',
        'The registration flow expired'
      )

      // try again with long lifespan set
      cy.get('input[name="traits"]').should('not.exist')
      cy.get('input[name="traits.email"]').type(email)
      cy.get('input[name="traits.website').type(website)
      cy.get('input[name="password"]').type(password)
      cy.get('button[type="submit"]').click()

      cy.url().should('eq', 'https://www.ory.sh/')
    })
  })
})

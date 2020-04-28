import {APP_URL, gen, password} from "../../../../helpers";

context('Registration', () => {
  beforeEach(() => {
    cy.visit(APP_URL + '/auth/registration')
  })

  it('should sign up and be logged in', () => {
    const email = gen.email()
    const website = 'https://www.ory.sh/'
    cy.get('input[name="traits.email"]').type(email)
    cy.get('input[name="traits.website').type(website)
    cy.get('input[name="password"]').type(password)

    cy.get('button[type="submit"]').click()
    cy.get('pre').should('contain.text', email)
    cy.get('.greeting').should('contain.text', "Welcome back")

    cy.session().should(({body: session}) => {
      const {identity} = session
      expect(identity.id).to.not.be.empty
      expect(identity.addresses).to.be.undefined
      expect(identity.traits_schema_id).to.equal('default')
      expect(identity.traits_schema_url).to.equal(`${APP_URL}/.ory/kratos/public/schemas/default`)
      expect(identity.traits.website).to.equal(website)
      expect(identity.traits.email).to.equal(email)
    })
  })
})

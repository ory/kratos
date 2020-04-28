import {APP_URL, gen, password,website} from "../../../../helpers"

context('Login', () => {
  const email = gen.email()
  const password = gen.password()

  before(() => {
    cy.register({email, password, fields: {'traits.website': website}})
  })

  beforeEach(() => {
    cy.clearCookies()
    cy.visit(APP_URL + '/auth/login')
  })

  it('should sign up and be logged in', () => {
    cy.get('input[name="identifier"]').type(email)
    cy.get('input[name="password"]').type(password)
    cy.get('button[type="submit"]').click()

    cy.session().should((session) => {
      const {identity} = session
      expect(identity.id).to.not.be.empty
      expect(identity.traits_schema_id).to.equal('default')
      expect(identity.traits_schema_url).to.equal(`${APP_URL}/.ory/kratos/public/schemas/default`)
      expect(identity.traits.website).to.equal(website)
      expect(identity.traits.email).to.equal(email)
    })
  })
})

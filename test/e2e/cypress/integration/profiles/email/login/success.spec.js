import { APP_URL, gen, password, website } from '../../../../helpers'

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

      cy.session().should((session) => {
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

      cy.session().should((session) => {
        const { identity } = session
        expect(identity.id).to.not.be.empty
        expect(identity.schema_id).to.equal('default')
        expect(identity.schema_url).to.equal(`${APP_URL}/schemas/default`)
        expect(identity.traits.website).to.equal(website)
        expect(identity.traits.email).to.equal(email)
      })
    })
  })
})

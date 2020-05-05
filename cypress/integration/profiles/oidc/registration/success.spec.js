import {APP_URL, gen, website} from '../../../../helpers'

context('Register', () => {
  beforeEach(() => {
    cy.clearCookies()
    cy.visit(APP_URL + '/auth/registration')
  })

  const shouldSession = (email) => (session) => {
    const {identity} = session
    expect(identity.id).to.not.be.empty
    expect(identity.traits_schema_id).to.equal('default')
    expect(identity.traits_schema_url).to.equal(
      `${APP_URL}/.ory/kratos/public/schemas/default`
    )
    expect(identity.traits.website).to.equal(website)
    expect(identity.traits.email).to.equal(email)
  }

  it('should be able to sign up with incomplete data and finally be signed in', () => {
    const email = gen.email()

    cy.registerOidc({email, expectSession: false})

    cy.get('#registration-password').should('not.exist');
    cy.get('#registration-oidc input[name="traits.email"]').should('have.value', email)
    cy.get('#registration-oidc form > *:last-child').should('have.attr', 'name', 'provider')
    cy.get('.form-errors .message').should('contain.text', 'missing properties: "website"')
    cy.get('#registration-oidc input[name="traits.website"]').type("http://s")

    cy.get('button[value="hydra"]').click()

    cy.get('#registration-password').should('not.exist');
    cy.get('#registration-oidc input[name="traits.email"]').should('have.value', email)
    cy.get('#registration-oidc form > *:last-child').should('have.attr', 'name', 'provider')
    cy.get('.form-errors .message').should('contain.text', 'length must be >= 10')
    cy.get('#registration-oidc input[name="traits.website"]').should('have.value', 'http://s').clear().type(website)

    cy.get('button[value="hydra"]').click()

    cy.session().should(shouldSession(email))
  })

  it('should be able to sign up with complete data', () => {
    const email = gen.email()

    cy.registerOidc({email, website})
    cy.session().should(shouldSession(email))
  })

  xit('should be able to sign up, sign out, and then sign in', () => {
    const email = gen.email()

    cy.registerOidc({email, website})
    cy.clearCookies()
    cy.registerOidc({email})
    cy.session().should(shouldSession(email))
  })
})

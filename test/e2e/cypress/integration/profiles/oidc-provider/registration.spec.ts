import { gen } from '../../../helpers'
import * as uuid from 'uuid'
import * as oauth2 from '../../../helpers/oauth2'

context('OpenID Provider', () => {
  const client = {
    auth_endpoint: 'http://localhost:4744/oauth2/auth',
    token_endpoint: 'http://localhost:4744/oauth2/token',
    id: 'dummy-client',
    secret: 'secret',
    token_endpoint_auth_method: 'client_secret_basic',
    grant_types: ['authorization_code', 'refresh_token'],
    response_types: ['code', 'id_token'],
    scopes: ['openid', 'offline', 'email', 'website'],
    callbacks: [
      'http://localhost:5555/callback',
      'https://httpbin.org/anything'
    ]
  }

  it('registration', () => {
    const state = uuid.v4()
    const nonce = uuid.v4()
    const scope = ['offline', 'openid']
    const url = oauth2.getAuthorizeURL(
      client.auth_endpoint,
      '',
      client.id,
      '0',
      nonce,
      'https://httpbin.org/anything',
      'code',
      ['offline', 'openid'],
      state,
      undefined
    )

    cy.visit(url)
    cy.get('[data-testid=cta-link]').click()

    const email = gen.email()
    const password = gen.password()

    cy.get('[name="traits.email"]').type(email)
    cy.get('[name=password]').type(password)
    cy.get('[name="traits.website"]').type('http://example.com')
    cy.get('input[type=checkbox][name="traits.tos"]').click({ force: true })
    cy.get('[name="traits.age"]').type('199')
    cy.get('input[type=checkbox][name="traits.consent"]').click({ force: true })
    cy.get('input[type=checkbox][name="traits.newsletter"]').click({
      force: true
    })
    cy.get('[type=submit]').click()

    cy.get('#openid').click()
    cy.get('#offline').click()
    cy.get('#accept').click()

    cy.location('href')
      .should('match', new RegExp('https://httpbin.org/anything[?]code=.*'))
      .then((body) => {
        cy.get('body')
          .invoke('text')
          .then((text) => {
            const result = JSON.parse(text)
            const tokenParams = {
              code: result.args.code,
              redirect_uri: 'https://httpbin.org/anything',
              scope: scope.join(' ')
            }
            oauth2
              .getToken(
                client.token_endpoint,
                client.id,
                client.secret,
                'authorization_code',
                tokenParams.code,
                tokenParams.redirect_uri,
                tokenParams.scope
              )
              .then((res) => {
                const token = res.body
                expect(token).to.have.property('access_token')
                expect(token).to.have.property('id_token')
                expect(token).to.have.property('refresh_token')
                expect(token).to.have.property('token_type')
                expect(token).to.have.property('expires_in')
                expect(token.scope).to.equal('offline openid')
                let idToken = JSON.parse(decodeURIComponent(escape(window.atob(token.id_token.split('.')[1]))))
                expect(idToken).to.have.property('amr')
                expect(idToken.amr).to.deep.equal(["password"])
              })
          })
      })
  })
})

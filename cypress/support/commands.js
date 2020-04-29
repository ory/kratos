// ***********************************************
// This example commands.js shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************
//
//
// -- This is a parent command --
// Cypress.Commands.add("login", (email, password) => { ... })
//
//
// -- This is a child command --
// Cypress.Commands.add("drag", { prevSubject: 'element'}, (subject, options) => { ... })
//
//
// -- This is a dual command --
// Cypress.Commands.add("dismiss", { prevSubject: 'optional'}, (subject, options) => { ... })
//
//
// -- This will overwrite an existing command --
// Cypress.Commands.overwrite("visit", (originalFn, url, options) => { ... })
import {
  APP_URL,
  assertAddress,
  gen,
  MAIL_API,
  parseHtml,
  pollInterval,
  privilegedLifespan,
} from '../helpers'

Cypress.Commands.add(
  'register',
  ({ email = gen.email(), password = gen.password(), fields = {} } = {}) => {
    cy.visit(APP_URL + '/auth/registration')
    cy.get('input[name="traits.email"]').type(email)
    cy.get('input[name="password"]').type(password)
    Object.keys(fields).forEach((key) => {
      const value = fields[key]
      cy.get(`input[name="${key}"]`).clear().type(value)
    })
    cy.get('button[type="submit"]').click()
  }
)

Cypress.Commands.add('login', ({ email, password, expectSession = true }) => {
  cy.visit(APP_URL + '/auth/login')
  cy.get('input[name="identifier"]').clear().type(email)
  cy.get('input[name="password"]').clear().type(password)
  cy.get('button[type="submit"]').click()
  if (expectSession) {
    cy.session()
  } else {
    cy.noSession()
  }
})

Cypress.Commands.add(
  'reauth',
  ({
    expect: { email },
    type: { email: temail, password: tpassword } = {},
  }) => {
    cy.url().should('include', '/auth/login')
    cy.get('input[name="identifier"]').should('have.value', email)
    if (temail) {
      cy.get('input[name="identifier"]').clear().type(temail)
    }
    if (tpassword) {
      cy.get('input[name="password"]').clear().type(tpassword)
    }
    cy.get('button[type="submit"]').click()
  }
)

Cypress.Commands.add('deleteMail', ({ atLeast = 0 } = {}) => {
  let tries = 0
  let count = 0
  const req = () =>
    cy
      .request('DELETE', `${MAIL_API}/mail`, { pruneCode: 'all' })
      .then(({ body }) => {
        count += parseInt(body)
        if (count < atLeast && tries < 100) {
          cy.log(
            `Expected at least ${atLeast} messages but deleteted only ${count} so far (body: ${body})`
          )
          tries++
          cy.wait(pollInterval)
          return req()
        }

        return Promise.resolve()
      })

  return req()
})

Cypress.Commands.add('session', () =>
  cy
    .request('GET', `${APP_URL}/.ory/kratos/public/sessions/whoami`)
    .then((response) => {
      expect(response.body.sid).to.not.be.empty
      expect(Cypress.moment().isBefore(Cypress.moment(response.body.expires_at))).to
        .be.true

      // Add a grace second for MySQL which does not support millisecs.
      expect(Cypress.moment().isAfter(Cypress.moment(response.body.issued_at).subtract(1,'s'))).to.be
        .true
      expect(Cypress.moment().isAfter(Cypress.moment(response.body.authenticated_at).subtract(1,'s')))
        .to.be.true
      expect(response.body.identity).to.exist
      return response.body
    })
)

Cypress.Commands.add('noSession', () =>
  cy
    .request({
      method: 'GET',
      url: `${APP_URL}/.ory/kratos/public/sessions/whoami`,
      failOnStatusCode: false,
    })
    .then((request) => {
      expect(request.status).to.eq(401)
      return request
    })
)

Cypress.Commands.add('verifyEmail', ({ expect: { email } = {} } = {}) =>
  cy.getMail().then((message) => {
    expect(message.subject.trim()).to.equal('Please verify your email address')
    expect(message.fromAddress.trim()).to.equal('no-reply@ory.kratos.sh')
    expect(message.toAddresses).to.have.length(1)
    expect(message.toAddresses[0].trim()).to.equal(email)

    const link = parseHtml(message.body).querySelector('a')
    expect(link).to.not.be.null
    expect(link.href).to.contain(APP_URL)

    cy.visit(link.href)
    cy.location('pathname').should('not.contain', 'verify')
    cy.session().should(assertAddress({ isVerified: true, email }))
  })
)

// Uses the verification email but waits so that it expires
Cypress.Commands.add(
  'verifyEmailButExpired',
  ({ expect: { email } = {} } = {}) =>
    cy.getMail().then((message) => {
      expect(message.subject.trim()).to.equal(
        'Please verify your email address'
      )
      expect(message.fromAddress.trim()).to.equal('no-reply@ory.kratos.sh')
      expect(message.toAddresses).to.have.length(1)
      expect(message.toAddresses[0].trim()).to.equal(email)

      const link = parseHtml(message.body).querySelector('a')
      cy.session().should((session) => {
        assertAddress({ isVerified: false, email: email })(session)
        cy.wait(
          Cypress.moment(session.identity.addresses[0].expires_at).diff(
            Cypress.moment()
          ) + 100
        )
      })

      cy.visit(link.href)
      cy.location('pathname').should('include', 'verify')
      cy.location('search').should('not.be.empty', 'request')
      cy.get('.form-errors .message').should('contain.text', 'code has expired')

      cy.session().should(assertAddress({ isVerified: false, email: email }))
    })
)

// Uses the verification email but waits so that it expires
Cypress.Commands.add('waitForPrivilegedSessionToExpire', () => {
  cy.session().should((session) => {
    expect(session.authenticated_at).to.not.be.empty
    cy.wait(
      Cypress.moment(session.authenticated_at)
        .add(privilegedLifespan)
        .diff(Cypress.moment()) + 100
    )
  })
})

Cypress.Commands.add('getMail', ({ removeMail = true } = {}) => {
  let tries = 0
  const req = () =>
    cy.request(`${MAIL_API}/mail`).then((response) => {
      expect(response.body).to.have.property('mailItems')
      const count = response.body.mailItems.length
      if (count === 0 && tries < 100) {
        tries++
        cy.wait(pollInterval)
        return req()
      }

      expect(count).to.equal(1)
      if (removeMail) {
        return cy
          .deleteMail({ atLeast: count })
          .then(() => Promise.resolve(response.body.mailItems[0]))
      }

      return Promise.resolve(response.body.mailItems[0])
    })

  return req()
})

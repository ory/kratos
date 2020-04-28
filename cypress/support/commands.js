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
import {APP_URL, MAIL_API} from '../helpers'

Cypress.Commands.add(
  'register',
  ({
     email,
     password,
     fields = {}
   } = {}) => {
    if (!email) {
      email = Math.random().toString(36).substring(7) + "@" + Math.random().toString(36).substring(7)
    }
    if (!password) {
      password = Math.random().toString(36).substring(7)
    }
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

Cypress.Commands.add(
  'login',
  ({
     email,
     password,
     expectSession = true
   }) => {
    cy.visit(APP_URL + '/auth/login')
    cy.get('input[name="identifier"]').clear().type(email)
    cy.get('input[name="password"]').clear().type(password)
    cy.get('button[type="submit"]').click()
    if (expectSession) {
      cy.session()
    } else {
      cy.noSession()
    }
  }
)

Cypress.Commands.add(
  'reauth',
  ({
     expect: {email},
     type: {
       email: temail, password: tpassword
     } = {}
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

Cypress.Commands.add(
  'deleteMail',
  () => {
    cy.request('DELETE', `${MAIL_API}/mail`, {pruneCode: "all"})
  }
)

Cypress.Commands.add(
  'session',
  () => {
    return cy.request('GET', `${APP_URL}/.ory/kratos/public/sessions/whoami`).then((response) => {
      expect(response.body.sid).to.not.be.empty
      expect(Cypress.moment().isBefore(new Date(response.body.expires_at))).to.be.true
      expect(Cypress.moment().isAfter(new Date(response.body.issued_at))).to.be.true
      expect(Cypress.moment().isAfter(new Date(response.body.authenticated_at))).to.be.true
      expect(response.body.identity).to.exist
      return response.body
    })
  }
)

Cypress.Commands.add(
  'noSession',
  () => {
    return cy.request({method: 'GET', url: `${APP_URL}/.ory/kratos/public/sessions/whoami`, failOnStatusCode: false})
      .then((request) => {
        expect(request.status).to.eq(401)
        return request
      })
  }
)

Cypress.Commands.add(
  'getMail',
  () => {
    let tries = 0
    const req = () => cy.request(`${MAIL_API}/mail`)
      .then((response) => {
        expect(response.body).to.have.property('mailItems')
        const count = response.body.mailItems.length
        if (count === 0 && tries < 10) {
          tries++
          cy.wait(1000)
          return req()
        }

        expect(count).to.equal(1)
        return Promise.resolve(response.body)
      })

    return req()
  }
)

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
  assertVerifiableAddress,
  gen,
  KRATOS_ADMIN,
  KRATOS_PUBLIC,
  MAIL_API,
  MOBILE_URL,
  parseHtml,
  pollInterval,
  privilegedLifespan
} from '../helpers'
import dayjs from 'dayjs'

const YAML = require('yamljs')

const configFile = 'test/e2e/kratos.generated.yml'

const mergeFields = (form, fields) => {
  const result = {}
  form.nodes.forEach(({ attributes, type }) => {
    if (type === 'input') {
      result[attributes.name] = attributes.value
    }
  })

  return { ...result, ...fields }
}

const updateConfigFile = (cb) => {
  cy.readFile(configFile).then((contents) => {
    let config = YAML.parse(contents)
    config = cb(config)
    cy.writeFile(configFile, YAML.stringify(config))
  })
  cy.wait(100)
}

let previousProfile = ''
Cypress.Commands.add('useConfigProfile', (profile) => {
  if (profile === previousProfile) {
    return
  }

  cy.readFile(`test/e2e/kratos.${profile}.yml`).then((contents) =>
    cy.writeFile(configFile, contents)
  )
  cy.wait(100)
})

Cypress.Commands.add('shortPrivilegedSessionTime', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.settings.privileged_session_max_age = '1ms'
    return config
  })
})

Cypress.Commands.add('longPrivilegedSessionTime', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.settings.privileged_session_max_age = '1m'
    return config
  })
})
Cypress.Commands.add('longVerificationLifespan', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.verification.lifespan = '1m'
    return config
  })
})
Cypress.Commands.add('shortVerificationLifespan', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.verification.lifespan = '4s'
    return config
  })
})
Cypress.Commands.add('longRecoveryLifespan', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.recovery.lifespan = '1m'
    return config
  })
})

Cypress.Commands.add('shortRecoveryLifespan', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.recovery.lifespan = '4s'
    return config
  })
})

Cypress.Commands.add(
  'register',
  ({
    email = gen.email(),
    password = gen.password(),
    query = {},
    fields = {}
  } = {}) => {
    console.log('Creating user account: ', { email, password })

    // see https://github.com/cypress-io/cypress/issues/408
    cy.visit(APP_URL)
    cy.clearCookies()

    //
    // cy.request({
    //   url: APP_URL + '/self-service/registration/browser',
    //   followRedirect: false,
    //   headers: {
    //     'Accept': 'application/json'
    //   },
    //   qs: query
    // })
    cy.request({
      url: APP_URL + '/self-service/registration/browser',
      followRedirect: false,
      headers: {
        Accept: 'application/json'
      },
      qs: query
    })
      .then(({ body, status }) => {
        expect(status).to.eq(200)
        const form = body.ui
        return cy.request({
          method: form.method,
          body: mergeFields(form, {
            ...fields,
            'traits.email': email,
            password,
            method: 'password'
          }),
          url: form.action,
          followRedirect: false
        })
      })
      .then(({ body }) => {
        expect(body.identity.traits.email).to.contain(email)
      })
  }
)

Cypress.Commands.add(
  'registerApi',
  ({ email = gen.email(), password = gen.password(), fields = {} } = {}) =>
    cy
      .request({
        url: APP_URL + '/self-service/registration/api'
      })
      .then(({ body }) => {
        const form = body.ui
        return cy.request({
          method: form.method,
          body: mergeFields(form, {
            ...fields,
            'traits.email': email,
            password,
            method: 'password'
          }),
          url: form.action
        })
      })
      .then(({ body }) => {
        expect(body.identity.traits.email).to.contain(email)
      })
)

Cypress.Commands.add(
  'registerOidc',
  ({
    email,
    website,
    scopes,
    rememberLogin = true,
    rememberConsent = true,
    acceptLogin = true,
    acceptConsent = true,
    expectSession = true
  }) => {
    cy.visit(APP_URL + '/auth/registration')

    cy.get('button[value="hydra"]').click()

    cy.get('#username').type(email)
    if (rememberLogin) {
      cy.get('#remember').click()
    }
    if (acceptLogin) {
      cy.get('#accept').click()
    } else {
      cy.get('#reject').click()
    }

    if (scopes) {
      scopes.forEach((scope) => {
        cy.get('#' + scope).click()
      })
    } else {
      cy.get('input[name="scope"]').each(($el) => cy.wrap($el).click())
    }

    if (website) {
      cy.get('#website').clear().type(website)
    }

    if (rememberConsent) {
      cy.get('#remember').click()
    }
    if (acceptConsent) {
      cy.get('#accept').click()
    } else {
      cy.get('#reject').click()
    }

    if (expectSession) {
      cy.session()
    } else {
      cy.noSession()
    }
  }
)

Cypress.Commands.add('loginOidc', ({ expectSession = true }) => {
  cy.visit(APP_URL + '/auth/login')
  cy.get('button[value="hydra"]').click()
  if (expectSession) {
    cy.session()
  } else {
    cy.noSession()
  }
})

Cypress.Commands.add('login', ({ email, password, expectSession = true }) => {
  if (expectSession) {
    console.log('Singing in user: ', { email, password })
  } else {
    console.log('Attempting user sign in: ', { email, password })
  }

  // see https://github.com/cypress-io/cypress/issues/408
  cy.visit(APP_URL)
  cy.clearCookies()

  cy.longPrivilegedSessionTime()
  cy.request({
    url: APP_URL + '/self-service/login/browser',
    followRedirect: false,
    failOnStatusCode: false,
    headers: {
      Accept: 'application/json'
    }
  })
    .then(({ body, status }) => {
      expect(status).to.eq(200)
      const form = body.ui
      return cy.request({
        method: form.method,
        body: mergeFields(form, {
          password_identifier: email,
          password,
          method: 'password'
        }),
        headers: {
          Accept: 'application/json'
        },
        url: form.action,
        followRedirect: false,
        failOnStatusCode: false
      })
    })
    .then(({ status }) => {
      console.log('Login sequence completed: ', { email, password })
      if (expectSession) {
        expect(status).to.eq(200)
        return cy.session()
      } else {
        expect(status).to.not.eq(200)
        return cy.noSession()
      }
    })
})

Cypress.Commands.add('loginMobile', ({ email, password }) => {
  cy.visit(MOBILE_URL)
  cy.get('input[data-testid="password_identifier"]').type(email)
  cy.get('input[data-testid="password"]').type(password)
  cy.get('div[data-testid="submit-form"]').click()
})

Cypress.Commands.add('logout', () => {
  cy.get('.logout a').click()
  cy.noSession()
})

Cypress.Commands.add(
  'reauth',
  ({
    expect: { email },
    type: { email: temail, password: tpassword } = {}
  }) => {
    cy.url().should('include', '/auth/login')
    cy.get('input[name="password_identifier"]').should('have.value', email)
    if (temail) {
      cy.get('input[name="password_identifier"]').clear().type(temail)
    }
    if (tpassword) {
      cy.get('input[name="password"]').clear().type(tpassword)
    }
    cy.longPrivilegedSessionTime()
    cy.get('button[value="password"]').click()
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
  cy.request('GET', `${KRATOS_PUBLIC}/sessions/whoami`).then((response) => {
    expect(response.body.id).to.not.be.empty
    expect(dayjs().isBefore(dayjs(response.body.expires_at))).to.be.true

    // Add a grace second for MySQL which does not support millisecs.
    expect(dayjs().isAfter(dayjs(response.body.issued_at).subtract(1, 's'))).to
      .be.true
    expect(
      dayjs().isAfter(dayjs(response.body.authenticated_at).subtract(1, 's'))
    ).to.be.true
    expect(response.body.identity).to.exist
    return response.body
  })
)

Cypress.Commands.add('noSession', () =>
  cy
    .request({
      method: 'GET',
      url: `${KRATOS_PUBLIC}/sessions/whoami`,
      failOnStatusCode: false
    })
    .then((request) => {
      expect(request.status).to.eq(401)
      return request
    })
)
Cypress.Commands.add('getIdentityByEmail', ({ email }) =>
  cy
    .request({
      method: 'GET',
      url: `${KRATOS_ADMIN}/identities`,
      failOnStatusCode: false
    })
    .then((response) => {
      expect(response.status).to.eq(200)
      return response.body.find((identity) => identity.traits.email === email)
    })
)

Cypress.Commands.add(
  'performEmailVerification',
  ({ expect: { email, redirectTo } = {} } = {}) =>
    cy.getMail().then((message) => {
      expect(message.subject.trim()).to.equal(
        'Please verify your email address'
      )
      expect(message.fromAddress.trim()).to.equal('no-reply@ory.kratos.sh')
      expect(message.toAddresses).to.have.length(1)
      expect(message.toAddresses[0].trim()).to.equal(email)

      const link = parseHtml(message.body).querySelector('a')
      expect(link).to.not.be.null
      expect(link.href).to.contain(APP_URL)

      if (redirectTo) {
        cy.request({ url: link.href, followRedirect: false }).should(
          (response) => {
            expect(response.status).to.eq(302)
            expect(response.redirectedToUrl).to.eq(redirectTo)
          }
        )
      } else {
        cy.visit(link.href)
        cy.location('pathname').should('not.contain', 'verify')
      }
    })
)

Cypress.Commands.add(
  'verifyEmail',
  ({ expect: { email, redirectTo } = {} } = {}) =>
    cy.performEmailVerification({ expect: { email, redirectTo } }).then(() => {
      cy.session().should(assertVerifiableAddress({ email, isVerified: true }))
    })
)

// Uses the verification email but waits so that it expires
Cypress.Commands.add(
  'recoverEmailButExpired',
  ({ expect: { email } = {} } = {}) =>
    cy.getMail().then((message) => {
      expect(message.subject.trim()).to.equal('Recover access to your account')
      expect(message.toAddresses[0].trim()).to.equal(email)

      const link = parseHtml(message.body).querySelector('a')
      expect(link).to.not.be.null
      expect(link.href).to.contain(APP_URL)

      cy.longRecoveryLifespan()
      cy.visit(link.href)
    })
)

Cypress.Commands.add('recoverEmail', ({ expect: { email } = {} } = {}) =>
  cy.getMail().then((message) => {
    expect(message.subject.trim()).to.equal('Recover access to your account')
    expect(message.fromAddress.trim()).to.equal('no-reply@ory.kratos.sh')
    expect(message.toAddresses).to.have.length(1)
    expect(message.toAddresses[0].trim()).to.equal(email)

    const link = parseHtml(message.body).querySelector('a')
    expect(link).to.not.be.null
    expect(link.href).to.contain(APP_URL)

    cy.visit(link.href)
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
        assertVerifiableAddress({ isVerified: false, email: email })(session)
        // specified in base...
      })

      cy.longVerificationLifespan()
      cy.visit(link.href)
      cy.location('pathname').should('include', 'verify')
      cy.location('search').should('not.be.empty', 'request')
      cy.get('.messages .message').should(
        'contain.text',
        'verification flow expired'
      )

      cy.session().should(
        assertVerifiableAddress({ isVerified: false, email: email })
      )
    })
)

// Uses the verification email but waits so that it expires
Cypress.Commands.add('waitForPrivilegedSessionToExpire', () => {
  cy.session().should((session) => {
    expect(session.authenticated_at).to.not.be.empty
    cy.wait(
      dayjs(session.authenticated_at).add(privilegedLifespan).diff(dayjs()) +
        100
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

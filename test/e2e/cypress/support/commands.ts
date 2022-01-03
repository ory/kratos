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
import YAML from 'yamljs'
import { Session } from '@ory/kratos-client'

const configFile = 'kratos.generated.yml'

const mergeFields = (form, fields) => {
  const result = {}
  form.nodes.forEach(({ attributes, type }) => {
    if (type === 'input') {
      result[attributes.name] = attributes.value
    }
  })

  return { ...result, ...fields }
}

const updateConfigFile = (cb: (arg: any) => any) => {
  cy.readFile(configFile).then((contents) => {
    cy.writeFile(configFile, YAML.stringify(cb(YAML.parse(contents))))
    cy.wait(200)
  })
}

Cypress.Commands.add('useConfigProfile', (profile: string) => {
  console.log('Switching config profile to:', profile)
  cy.readFile(`kratos.${profile}.yml`).then((contents) =>
    cy.writeFile(configFile, contents)
  )
  cy.wait(200)
})

Cypress.Commands.add('proxy', (app: string) => {
  console.log('Switching proxy profile to:', app)
  cy.writeFile(`proxy.json`, `"${app}"`)
  cy.readFile(`proxy.json`).should('eq', app)
  cy.wait(200)
  cy.visit(APP_URL + '/')
  cy.get(`[data-testid="app-${app}"]`).should('exist')
  cy.clearAllCookies()
  cy.visit(APP_URL + '/')
})

Cypress.Commands.add('shortPrivilegedSessionTime', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.settings.privileged_session_max_age = '1ms'
    return config
  })
})

Cypress.Commands.add('setIdentitySchema', (schema: string) => {
  updateConfigFile((config) => {
    config.identity.default_schema_url = schema
    return config
  })
})

Cypress.Commands.add('longPrivilegedSessionTime', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.settings.privileged_session_max_age = '5m'
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
    config.selfservice.flows.verification.lifespan = '1ms'
    return config
  })
})
Cypress.Commands.add('sessionRequiresNo2fa', ({} = {}) => {
  updateConfigFile((config) => {
    config.session.whoami.required_aal = 'aal1'
    return config
  })
})
Cypress.Commands.add('sessionRequires2fa', ({} = {}) => {
  updateConfigFile((config) => {
    config.session.whoami.required_aal = 'highest_available'
    return config
  })
})
Cypress.Commands.add('shortLinkLifespan', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.methods.link.config.lifespan = '1ms'
    return config
  })
})
Cypress.Commands.add('longLinkLifespan', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.methods.link.config.lifespan = '1m'
    return config
  })
})

Cypress.Commands.add('longRecoveryLifespan', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.recovery.lifespan = '1m'
    return config
  })
})

Cypress.Commands.add('enableLoginForVerifiedAddressOnly', () => {
  updateConfigFile((config) => {
    config.selfservice.flows.login['after'] = {
      password: { hooks: [{ hook: 'require_verified_address' }] }
    }
    return config
  })
})

Cypress.Commands.add('shortLoginLifespan', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.login.lifespan = '100ms'
    return config
  })
})
Cypress.Commands.add('longLoginLifespan', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.login.lifespan = '1h'
    return config
  })
})

Cypress.Commands.add('shortRecoveryLifespan', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.recovery.lifespan = '1ms'
    return config
  })
})

Cypress.Commands.add('requireStrictAal', () => {
  updateConfigFile((config) => {
    config.selfservice.flows.settings.required_aal = 'highest_available'
    config.session.whoami.required_aal = 'highest_available'
    return config
  })
})

Cypress.Commands.add('useLaxAal', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.settings.required_aal = 'aal1'
    config.session.whoami.required_aal = 'aal1'
    return config
  })
})

Cypress.Commands.add('disableVerification', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.verification.enabled = false
    return config
  })
})

Cypress.Commands.add('enableVerification', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.verification.enabled = true
    return config
  })
})

Cypress.Commands.add('enableRecovery', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.recovery.enabled = true
    return config
  })
})

Cypress.Commands.add('disableRecovery', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.recovery.enabled = false
    return config
  })
})

Cypress.Commands.add('disableRegistration', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.registration.enabled = false
    return config
  })
})

Cypress.Commands.add('enableRegistration', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.registration.enabled = true
    return config
  })
})

Cypress.Commands.add('useLaxAal', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.settings.required_aal = 'aal1'
    config.session.whoami.required_aal = 'aal1'
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

Cypress.Commands.add('settingsApi', ({ fields = {} } = {}) =>
  cy
    .request({
      url: APP_URL + '/self-service/settings/api'
    })
    .then(({ body }) => {
      const form = body.ui
      return cy.request({
        method: form.method,
        body: mergeFields(form, {
          ...fields
        }),
        url: form.action
      })
    })
    .then(({ body }) => {
      expect(body.statusCode).to.eq(200)
    })
)

Cypress.Commands.add('loginApi', ({ email, password } = {}) =>
  cy
    .request({
      url: APP_URL + '/self-service/login/api'
    })
    .then(({ body }) => {
      const form = body.ui
      return cy.request({
        method: form.method,
        body: mergeFields(form, {
          password_identifier: email,
          password,
          method: 'password'
        }),
        url: form.action
      })
    })
    .then(({ body }) => {
      expect(body.session.identity.traits.email).to.contain(email)
    })
)

Cypress.Commands.add('loginApiWithoutCookies', ({ email, password } = {}) => {
  cy.task('httpRequest', {
    url: APP_URL + '/self-service/login/api',
    headers: {
      Accept: 'application/json'
    },
    responseType: 'json'
  }).should((body: any) => {
    cy.task('httpRequest', {
      method: body.ui.method,
      json: mergeFields(body.ui, {
        password_identifier: email,
        password,
        method: 'password'
      }),
      headers: {
        Accept: 'application/json'
      },
      responseType: 'json',
      url: body.ui.action
    }).should((body: any) => {
      expect(body.session.identity.traits.email).to.contain(email)
      return body
    })
  })
})

Cypress.Commands.add('recoverApi', ({ email, returnTo }) => {
  let url = APP_URL + '/self-service/recovery/api'
  if (returnTo) {
    url += '?return_to=' + returnTo
  }
  cy.request({ url })
    .then(({ body }) => {
      const form = body.ui
      return cy.request({
        method: form.method,
        body: mergeFields(form, { email, method: 'link' }),
        url: form.action
      })
    })
    .then(({ body }) => {
      expect(body.state).to.contain('sent_email')
    })
})

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
    expectSession = true,
    route = APP_URL + '/registration'
  }) => {
    cy.visit(route)

    cy.triggerOidc()

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
    cy.location('pathname').should('not.include', 'consent')

    if (expectSession) {
      cy.getSession()
    } else {
      cy.noSession()
    }
  }
)

Cypress.Commands.add('shortRegisterLifespan', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.registration.lifespan = '100ms'
    return config
  })
})

Cypress.Commands.add('longRegisterLifespan', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.registration.lifespan = '1h'
    return config
  })
})

Cypress.Commands.add('browserReturnUrlOry', ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.whitelisted_return_urls = ['https://www.ory.sh/']
    return config
  })
})

Cypress.Commands.add(
  'loginOidc',
  ({ expectSession = true, url = APP_URL + '/login' }) => {
    cy.visit(url)
    cy.triggerOidc('hydra')
    cy.location('href').should('not.eq', '/consent')
    if (expectSession) {
      cy.getSession()
    } else {
      cy.noSession()
    }
  }
)

Cypress.Commands.add(
  'login',
  ({ email, password, expectSession = true, cookieUrl = APP_URL }) => {
    if (expectSession) {
      console.log('Singing in user: ', { email, password })
    } else {
      console.log('Attempting user sign in: ', { email, password })
    }

    // see https://github.com/cypress-io/cypress/issues/408
    cy.visit(cookieUrl)
    cy.clearAllCookies()

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
        console.log('Login sequence completed: ', {
          email,
          password,
          expectSession
        })
        if (expectSession) {
          expect(status).to.eq(200)
          return cy.getSession()
        } else {
          expect(status).to.not.eq(200)
          return cy.noSession()
        }
      })
  }
)

Cypress.Commands.add('loginMobile', ({ email, password }) => {
  cy.visit(MOBILE_URL + '/Login')
  cy.get('input[data-testid="password_identifier"]').type(email)
  cy.get('input[data-testid="password"]').type(password)
  cy.get('div[data-testid="submit-form"]').click()
})

Cypress.Commands.add('logout', () => {
  cy.getCookies().should((cookies) => {
    const c = cookies.find(
      ({ name }) => name.indexOf('ory_kratos_session') > -1
    )
    expect(c).to.not.be.undefined
    cy.clearCookie(c.name)
  })
  cy.noSession()
})

Cypress.Commands.add(
  'reauth',
  ({
    expect: { email, success = true },
    type: { email: temail, password: tpassword } = {
      email: undefined,
      password: undefined
    }
  }) => {
    cy.location('pathname').should('contain', '/login')
    cy.get('input[name="password_identifier"]').should('have.value', email)
    if (temail) {
      cy.get('input[name="password_identifier"]').clear().type(temail)
    }
    if (tpassword) {
      cy.get('input[name="password"]').clear().type(tpassword)
    }
    cy.longPrivilegedSessionTime()
    cy.get('button[value="password"]').click()
    if (success) {
      cy.location('pathname').should('not.contain', '/login')
    }
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

Cypress.Commands.add(
  'getSession',
  ({ expectAal = 'aal1', expectMethods = [] } = {}) =>
    cy.request('GET', `${KRATOS_PUBLIC}/sessions/whoami`).then((response) => {
      expect(response.body.id).to.not.be.empty
      expect(dayjs().isBefore(dayjs(response.body.expires_at))).to.be.true

      // Add a grace second for MySQL which does not support millisecs.
      expect(dayjs().isAfter(dayjs(response.body.issued_at).subtract(1, 's')))
        .to.be.true
      expect(
        dayjs().isAfter(dayjs(response.body.authenticated_at).subtract(1, 's'))
      ).to.be.true

      expect(response.body.identity).to.exist

      expect(response.body.authenticator_assurance_level).to.equal(expectAal)
      if (expectMethods.length > 0) {
        expect(response.body.authentication_methods).to.have.lengthOf(
          expectMethods.length
        )
        expectMethods.forEach((value) => {
          expect(
            response.body.authentication_methods.find(
              ({ method }) => method === value
            )
          ).to.exist
        })
      }

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
  ({
    expect: { email, redirectTo } = { email: undefined, redirectTo: undefined }
  } = {}) =>
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

      cy.request({ url: link.href, followRedirect: false }).should(
        (response) => {
          expect(response.status).to.eq(303)
          if (redirectTo) {
            expect(response.redirectedToUrl).to.eq(redirectTo)
          } else {
            expect(response.redirectedToUrl).to.not.contain('verification')
          }
        }
      )
    })
)

Cypress.Commands.add(
  'verifyEmail',
  ({ expect: { email, password, redirectTo } }) =>
    cy.performEmailVerification({ expect: { email, redirectTo } }).then(() => {
      cy.getSession().should((session) =>
        assertVerifiableAddress({ email, isVerified: true })(session)
      )
    })
)

// Uses the verification email but waits so that it expires
Cypress.Commands.add('recoverEmailButExpired', ({ expect: { email } }) => {
  cy.getMail().should((message) => {
    expect(message.subject.trim()).to.equal('Recover access to your account')
    expect(message.toAddresses[0].trim()).to.equal(email)

    const link = parseHtml(message.body).querySelector('a')
    expect(link).to.not.be.null
    expect(link.href).to.contain(APP_URL)

    cy.visit(link.href)
  })
})

Cypress.Commands.add(
  'recoverEmail',
  ({ expect: { email }, shouldVisit = true }) =>
    cy.getMail().should((message) => {
      expect(message.subject.trim()).to.equal('Recover access to your account')
      expect(message.fromAddress.trim()).to.equal('no-reply@ory.kratos.sh')
      expect(message.toAddresses).to.have.length(1)
      expect(message.toAddresses[0].trim()).to.equal(email)

      const link = parseHtml(message.body).querySelector('a')
      expect(link).to.not.be.null
      expect(link.href).to.contain(APP_URL)

      if (shouldVisit) {
        cy.visit(link.href)
      }
      return link.href
    })
)

// Uses the verification email but waits so that it expires
Cypress.Commands.add(
  'verifyEmailButExpired',
  ({ expect: { email, password } }) =>
    cy.getMail().then((message) => {
      expect(message.subject.trim()).to.equal(
        'Please verify your email address'
      )

      expect(message.fromAddress.trim()).to.equal('no-reply@ory.kratos.sh')
      expect(message.toAddresses).to.have.length(1)
      expect(message.toAddresses[0].trim()).to.equal(email)

      const link = parseHtml(message.body).querySelector('a')
      cy.getSession().should((session) => {
        assertVerifiableAddress({
          isVerified: false,
          email: email
        })(session)
      })

      cy.visit(link.href)
      cy.location('pathname').should('include', 'verification')
      cy.get('[data-testid="ui/message/4070005"]').should(
        'contain.text',
        'verification flow expired'
      )

      cy.getSession().should((session) => {
        assertVerifiableAddress({
          isVerified: false,
          email: email
        })(session)
      })
    })
)

// Uses the verification email but waits so that it expires
Cypress.Commands.add('waitForPrivilegedSessionToExpire', () => {
  cy.getSession().should((session: Session) => {
    expect(session.authenticated_at).to.not.be.empty
    cy.wait(
      dayjs(session.authenticated_at).add(privilegedLifespan).diff(dayjs()) +
        100
    )
  })
})

Cypress.Commands.add('getLookupSecrets', () =>
  cy
    .get('[data-testid="node/text/lookup_secret_codes/text"] code')
    .then(($e) => $e.map((_, e) => e.innerText.trim()).toArray())
)
Cypress.Commands.add('expectSettingsSaved', () =>
  cy
    .get('[data-testid="ui/message/1050001"]')
    .should('contain.text', 'Your changes have been saved')
)

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

Cypress.Commands.add('clearAllCookies', () => {
  cy.clearCookies({ domain: null })
})

Cypress.Commands.add('submitPasswordForm', () => {
  cy.get('[name="method"][value="password"]').click()
  cy.get('[name="method"][value="password"]:disabled').should('not.exist')
})

Cypress.Commands.add('submitProfileForm', () => {
  cy.get('[name="method"][value="profile"]').click()
  cy.get('[name="method"][value="profile"]:disabled').should('not.exist')
})

Cypress.Commands.add('clickWebAuthButton', (type: string) => {
  cy.get('*[data-testid="node/script/webauthn_script"]').should('exist')
  cy.wait(500) // Wait for script to load
  cy.get('*[name="webauthn_' + type + '_trigger"]').click()
  cy.wait(500) // Wait webauth to pass
})

Cypress.Commands.add('shouldShow2FAScreen', () => {
  cy.get('h2').should('contain.text', 'Two-Factor Authentication')
  cy.get('[data-testid="ui/message/1010004"]').should(
    'contain.text',
    'Please complete the second authentication challenge.'
  )
})

Cypress.Commands.add(
  'shouldErrorOnDisallowedReturnTo',
  (init: string, { app }: { app: 'express' | 'react' }) => {
    cy.visit(init, { failOnStatusCode: false })
    if (app === 'react') {
      cy.location('href').should('include', init.split('?')[0])
      cy.get('.Toastify').should(
        'contain.text',
        'The return_to address is not allowed.'
      )
    } else {
      cy.location('pathname').should('contain', 'error')
      cy.get('code').should(
        'contain.text',
        'Requested return_to URL \\"https://not-allowed\\" is not whitelisted.'
      )
    }
  }
)

Cypress.Commands.add(
  'shouldHaveCsrfError',
  ({ app }: { app: 'express' | 'react' }) => {
    let initial
    let pathname
    cy.location().should((location) => {
      initial = location.search
      pathname = location.pathname
    })

    cy.getCookies().should((cookies) => {
      const csrf = cookies.find(({ name }) => name.indexOf('csrf') > -1)
      expect(csrf).to.not.be.undefined
      cy.clearCookie(csrf.name)
    })
    cy.submitPasswordForm()

    // We end up at a new flow
    if (app === 'express') {
      cy.location().should((location) => {
        expect(initial).to.not.be.empty
        expect(location.search).to.not.eq(initial)
      })

      cy.location('pathname').should('include', '/error')
      cy.get('code').should('contain.text', 'csrf_token')
    } else {
      cy.location('pathname').should((got) => {
        expect(got).to.eql(pathname)
      })
      cy.get('.Toastify').should(
        'contain.text',
        'A security violation was detected, please fill out the form again.'
      )
    }
  }
)

Cypress.Commands.add('triggerOidc', (provider: string = 'hydra') => {
  cy.get('[name="provider"][value="' + provider + '"]').click()
})

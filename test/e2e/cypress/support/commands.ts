// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import {
  APP_URL,
  assertVerifiableAddress,
  extractRecoveryCode,
  gen,
  KRATOS_ADMIN,
  KRATOS_PUBLIC,
  MAIL_API,
  MOBILE_URL,
  parseHtml,
  pollInterval,
} from "../helpers"

import dayjs from "dayjs"
import YAML from "yamljs"
import { Strategy } from "."
import { OryKratosConfiguration } from "./config"

const configFile = "kratos.generated.yml"

const mergeFields = (form, fields) => {
  const result = {}
  form.nodes.forEach(({ attributes, type }) => {
    if (type === "input") {
      result[attributes.name] = attributes.value
    }
  })

  return { ...result, ...fields }
}

function checkConfigVersion(previous, tries = 0) {
  cy.wait(50)
  cy.request("GET", KRATOS_ADMIN + "/health/config").then(({ body }) => {
    if (previous !== body) {
      return
    } else if (tries > 8) {
      console.warn(
        "Config version did not change after 5 tries, maybe the changes did not have an effect?",
      )
      return
    }
    cy.wait(50)
    checkConfigVersion(previous, tries + 1)
  })
}

const updateConfigFile = (
  cb: (arg: OryKratosConfiguration) => OryKratosConfiguration,
) => {
  cy.request("GET", KRATOS_ADMIN + "/health/config").then(({ body }) => {
    cy.readFile(configFile).then((contents) => {
      cy.writeFile(configFile, YAML.stringify(cb(YAML.parse(contents))))
      cy.wait(500)
    })
    checkConfigVersion(body)
  })
}

Cypress.Commands.add("useConfigProfile", (profile: string) => {
  cy.request("GET", KRATOS_ADMIN + "/health/config").then(({ body }) => {
    console.log("Switching config profile to:", profile)
    cy.readFile(`kratos.${profile}.yml`).then((contents) =>
      cy.writeFile(configFile, contents),
    )
    checkConfigVersion(body)
  })
})

Cypress.Commands.add("proxy", (app: string) => {
  console.log("Switching proxy profile to:", app)
  cy.writeFile(`proxy.json`, `"${app}"`)
  cy.request(APP_URL + "/")
    .its("body", { log: false })
    .then((body) => {
      expect(body.indexOf(`data-testid="app-${app}"`) > -1).to.be.true
    })
})

Cypress.Commands.add("shortPrivilegedSessionTime", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.settings.privileged_session_max_age = "1ms"
    return config
  })
})

Cypress.Commands.add("setIdentitySchema", (schema: string) => {
  updateConfigFile((config) => {
    const id = gen.password()
    config.identity.default_schema_id = id
    config.identity.schemas = [
      ...(config.identity.schemas || []),
      {
        id,
        url: schema,
      },
    ]
    return config
  })
})

Cypress.Commands.add("setDefaultIdentitySchema", (id: string) => {
  updateConfigFile((config) => {
    config.identity.default_schema_id = id
    return config
  })
})

Cypress.Commands.add("longPrivilegedSessionTime", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.settings.privileged_session_max_age = "5m"
    return config
  })
})
Cypress.Commands.add("longVerificationLifespan", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.verification.lifespan = "1m"
    return config
  })
})
Cypress.Commands.add("shortVerificationLifespan", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.verification.lifespan = "1ms"
    return config
  })
})
Cypress.Commands.add("sessionRequiresNo2fa", ({} = {}) => {
  updateConfigFile((config) => {
    config.session.whoami.required_aal = "aal1"
    return config
  })
})
Cypress.Commands.add("sessionRequires2fa", ({} = {}) => {
  updateConfigFile((config) => {
    config.session.whoami.required_aal = "highest_available"
    return config
  })
})
Cypress.Commands.add("shortLinkLifespan", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.methods.link.config.lifespan = "1ms"
    return config
  })
})
Cypress.Commands.add("longLinkLifespan", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.methods.link.config.lifespan = "1m"
    return config
  })
})

Cypress.Commands.add("shortCodeLifespan", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.methods.code.config.lifespan = "1ms"
    return config
  })
})

Cypress.Commands.add("shortLifespan", (strategy: Strategy) => {
  updateConfigFile((config) => {
    config.selfservice.methods[strategy].config.lifespan = "1ms"
    return config
  })
})

Cypress.Commands.add("longLifespan", (strategy: Strategy) => {
  updateConfigFile((config) => {
    config.selfservice.methods[strategy].config.lifespan = "1m"
    return config
  })
})

Cypress.Commands.add("longCodeLifespan", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.methods.code.config.lifespan = "1m"
    return config
  })
})

Cypress.Commands.add("shortCodeLifespan", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.methods.code.config.lifespan = "1ms"
    return config
  })
})

Cypress.Commands.add("longCodeLifespan", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.methods.code.config.lifespan = "1m"
    return config
  })
})

Cypress.Commands.add("longRecoveryLifespan", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.recovery.lifespan = "1m"
    return config
  })
})

Cypress.Commands.add("enableLoginForVerifiedAddressOnly", () => {
  updateConfigFile((config) => {
    config.selfservice.flows.login["after"] = {
      password: { hooks: [{ hook: "require_verified_address" }] },
    }
    return config
  })
})

Cypress.Commands.add("setupHooks", (flow, phase, kind, hooks) => {
  updateConfigFile((config) => {
    config.selfservice.flows[flow][phase][kind] = { hooks }
    return config
  })
})

Cypress.Commands.add("setPostPasswordRegistrationHooks", (hooks) => {
  cy.setupHooks("registration", "after", "password", hooks)
})

Cypress.Commands.add("setPostCodeRegistrationHooks", (hooks) => {
  cy.setupHooks("registration", "after", "code", hooks)
})

Cypress.Commands.add("shortLoginLifespan", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.login.lifespan = "100ms"
    return config
  })
})
Cypress.Commands.add("longLoginLifespan", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.login.lifespan = "1h"
    return config
  })
})

Cypress.Commands.add("shortRecoveryLifespan", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.recovery.lifespan = "1ms"
    return config
  })
})

Cypress.Commands.add("requireStrictAal", () => {
  updateConfigFile((config) => {
    config.selfservice.flows.settings.required_aal = "highest_available"
    config.session.whoami.required_aal = "highest_available"
    return config
  })
})

Cypress.Commands.add("useLaxAal", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.settings.required_aal = "aal1"
    config.session.whoami.required_aal = "aal1"
    return config
  })
})

Cypress.Commands.add("disableVerification", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.verification.enabled = false
    return config
  })
})

Cypress.Commands.add("enableVerification", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.verification.enabled = true
    return config
  })
})

Cypress.Commands.add("enableRecovery", ({} = {}) => {
  updateConfigFile((config) => {
    if (!config.selfservice.flows.recovery) {
      config.selfservice.flows.recovery = {}
    }
    config.selfservice.flows.recovery.enabled = true
    return config
  })
})

Cypress.Commands.add("useRecoveryStrategy", (strategy: Strategy) => {
  updateConfigFile((config) => {
    if (!config.selfservice.flows.recovery) {
      config.selfservice.flows.recovery = {}
    }
    config.selfservice.flows.recovery.use = strategy
    if (!config.selfservice.methods[strategy]) {
      config.selfservice.methods[strategy] = {}
    }
    config.selfservice.methods[strategy].enabled = true
    return config
  })
})

Cypress.Commands.add("disableRecoveryStrategy", (strategy: Strategy) => {
  updateConfigFile((config) => {
    config.selfservice.methods[strategy].enabled = false
    return config
  })
})

Cypress.Commands.add("disableRecovery", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.recovery.enabled = false
    return config
  })
})

Cypress.Commands.add("disableRegistration", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.registration.enabled = false
    return config
  })
})

Cypress.Commands.add("enableRegistration", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.registration.enabled = true
    return config
  })
})

Cypress.Commands.add("useLaxAal", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.settings.required_aal = "aal1"
    config.session.whoami.required_aal = "aal1"
    return config
  })
})

Cypress.Commands.add("updateConfigFile", (cb: (arg: any) => any) => {
  updateConfigFile(cb)
})

Cypress.Commands.add(
  "register",
  ({
    email = gen.email(),
    password = gen.password(),
    query = {},
    fields = {},
  } = {}) => {
    console.log("Creating user account: ", { email, password })

    // see https://github.com/cypress-io/cypress/issues/408
    cy.clearAllCookies()

    cy.request({
      url: APP_URL + "/self-service/registration/browser",
      followRedirect: false,
      headers: {
        Accept: "application/json",
      },
      qs: query,
    })
      .then(({ body, status }) => {
        expect(status).to.eq(200)
        const form = body.ui
        return cy.request({
          method: form.method,
          body: mergeFields(form, {
            ...fields,
            "traits.email": email,
            password,
            method: "password",
          }),
          url: form.action,
          followRedirect: false,
        })
      })
      .then(({ body }) => {
        expect(body.identity.traits.email).to.contain(email)
      })
  },
)

Cypress.Commands.add(
  "registerWithCode",
  ({ email = gen.email(), code = undefined, query = {} } = {}) => {
    console.log("Creating user account: ", { email })

    cy.clearAllCookies()

    cy.request({
      url: APP_URL + "/self-service/registration/browser",
      method: "GET",
      followRedirect: false,
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      qs: query || {},
    }).then(({ body, status }) => {
      expect(status).to.eq(200)
      const form = body.ui
      return cy
        .request({
          headers: {
            Accept: "application/json",
          },
          method: form.method,
          body: mergeFields(form, {
            method: "code",
            "traits.email": email,
            ...(code && { code }),
          }),
          url: form.action,
          followRedirect: false,
        })
        .then(({ body }) => {
          if (!code) {
            expect(
              body.ui.nodes.find(
                (f) =>
                  f.group === "default" && f.attributes.name === "traits.email",
              ).attributes.value,
            ).to.eq(email)
            return cy.getRegistrationCodeFromEmail(email).then((code) => {
              return cy.request({
                headers: {
                  Accept: "application/json",
                },
                method: form.method,
                body: mergeFields(form, {
                  method: "code",
                  "traits.email": email,
                  code,
                }),
                url: form.action,
                followRedirect: false,
              })
            })
          } else {
            expect(body.session).to.contain(email)
          }
        })
    })
  },
)

Cypress.Commands.add(
  "registerApi",
  ({ email = gen.email(), password = gen.password(), fields = {} } = {}) =>
    cy
      .request({
        url: APP_URL + "/self-service/registration/api",
      })
      .then(({ body }) => {
        const form = body.ui
        return cy.request({
          method: form.method,
          body: mergeFields(form, {
            ...fields,
            "traits.email": email,
            password,
            method: "password",
          }),
          url: form.action,
        })
      })
      .then(({ body }) => {
        expect(body.identity.traits.email).to.contain(email)
        return body
      }),
)

Cypress.Commands.add("settingsApi", ({ fields = {} } = {}) =>
  cy
    .request({
      url: APP_URL + "/self-service/settings/api",
    })
    .then(({ body }) => {
      const form = body.ui
      return cy.request({
        method: form.method,
        body: mergeFields(form, {
          ...fields,
        }),
        url: form.action,
      })
    })
    .then(({ body }) => {
      expect(body.statusCode).to.eq(200)
    }),
)

Cypress.Commands.add("loginApi", ({ email, password } = {}) =>
  cy
    .request({
      url: APP_URL + "/self-service/login/api",
    })
    .then(({ body }) => {
      const form = body.ui
      return cy.request({
        method: form.method,
        body: mergeFields(form, {
          identifier: email,
          password,
          method: "password",
        }),
        url: form.action,
      })
    })
    .then(({ body }) => {
      expect(body.session.identity.traits.email).to.contain(email)
      return body
    }),
)

Cypress.Commands.add("loginApiWithoutCookies", ({ email, password } = {}) => {
  cy.task("httpRequest", {
    url: APP_URL + "/self-service/login/api",
    headers: {
      Accept: "application/json",
    },
    responseType: "json",
  }).should((body: any) => {
    cy.task("httpRequest", {
      method: body.ui.method,
      json: mergeFields(body.ui, {
        identifier: email,
        password,
        method: "password",
      }),
      headers: {
        Accept: "application/json",
      },
      responseType: "json",
      url: body.ui.action,
    }).should((body: any) => {
      expect(body.session.identity.traits.email).to.contain(email)
      return body
    })
  })
})

Cypress.Commands.add("recoverApi", ({ email, returnTo }) => {
  let url = APP_URL + "/self-service/recovery/api"
  if (returnTo) {
    url += "?return_to=" + returnTo
  }
  cy.request({ url })
    .then(({ body }) => {
      const form = body.ui
      // label should still exist after request, for more detail: #2591
      expect(form.nodes[1].meta).to.not.be.null
      expect(form.nodes[1].meta.label).to.not.be.null
      expect(form.nodes[1].meta.label.text).to.equal("Email")

      return cy.request({
        method: form.method,
        body: mergeFields(form, { email, method: "link" }),
        url: form.action,
      })
    })
    .then(({ body }) => {
      expect(body.state).to.contain("sent_email")
    })
})

Cypress.Commands.add(
  "verificationApi",
  ({ email, returnTo, strategy = "code" }) => {
    let url = APP_URL + "/self-service/verification/api"
    if (returnTo) {
      url += "?return_to=" + returnTo
    }
    cy.request({ url })
      .then(({ body }) => {
        const form = body.ui
        expect(form.nodes.some((node) => node.meta?.label?.text === "Email")).to
          .be.true

        return cy.request({
          method: form.method,
          body: mergeFields(form, { email, method: strategy }),
          url: form.action,
          headers: {
            Accept: "application/json", // "Emulate" an API client, as kratos responds with a redirect otherwise
          },
        })
      })
      .then(({ body }) => {
        expect(body.state).to.contain("sent_email")
      })
  },
)

Cypress.Commands.add(
  "verificationApiExpired",
  ({ email, returnTo, strategy = "code" }) => {
    cy.shortVerificationLifespan()
    let url = APP_URL + "/self-service/verification/api"
    if (returnTo) {
      url += "?return_to=" + returnTo
    }
    cy.request({ url })
      .then(({ body }) => {
        const form = body.ui
        return cy.request({
          method: form.method,
          body: mergeFields(form, { email, method: strategy }),
          url: form.action,
          failOnStatusCode: false,
        })
      })
      .then((response) => {
        expect(response.status).to.eq(410)
        expect(response.body.error.reason).to.eq(
          "The verification flow has expired. Redirect the user to the verification flow init endpoint to initialize a new verification flow.",
        )
        expect(response.body.error.details.redirect_to).to.eq(
          "http://localhost:4455/self-service/verification/browser",
        )
      })
  },
)

Cypress.Commands.add("verificationBrowser", ({ email, returnTo }) => {
  let url = APP_URL + "/self-service/verification/browser"
  if (returnTo) {
    url += "?return_to=" + returnTo
  }
  cy.request({ url })
    .then(({ body }) => {
      const form = body.ui
      return cy.request({
        method: form.method,
        body: mergeFields(form, { email, method: "link" }),
        url: form.action,
      })
    })
    .then(({ body }) => {
      expect(body.state).to.contain("sent_email")
    })
})
Cypress.Commands.add("addVirtualAuthenticator", () =>
  cy
    .task("sendCRI", {
      query: "WebAuthn.enable",
      opts: {},
    })
    .then(() =>
      cy.task("sendCRI", {
        query: "WebAuthn.addVirtualAuthenticator",
        opts: {
          options: {
            protocol: "ctap2",
            transport: "usb",
            hasResidentKey: true,
            hasUserVerification: true,
            isUserVerified: true,
          },
        },
      }),
    ),
)

Cypress.Commands.add(
  "registerOidc",
  ({
    app,
    email,
    website,
    scopes,
    rememberLogin = true,
    rememberConsent = true,
    acceptLogin = true,
    acceptConsent = true,
    expectSession = true,
    route = APP_URL + "/registration",
  }) => {
    cy.visit(route)

    cy.triggerOidc(app)

    cy.get("#username").type(email)
    if (rememberLogin) {
      cy.get("#remember").click()
    }
    if (acceptLogin) {
      cy.get("#accept").click()
    } else {
      cy.get("#reject").click()
    }

    if (scopes) {
      scopes.forEach((scope) => {
        cy.get("#" + scope).click()
      })
    } else {
      cy.get('input[name="scope"]').each(($el) => cy.wrap($el).click())
    }

    if (website) {
      cy.get("#website").clear().type(website)
    }

    if (rememberConsent) {
      cy.get("#remember").click()
    }

    if (acceptConsent) {
      cy.get("#accept").click()
    } else {
      cy.get("#reject").click()
    }

    cy.location("pathname").should("not.include", "consent")

    if (expectSession) {
      cy.getSession()
    } else {
      cy.noSession()
    }
  },
)

Cypress.Commands.add("shortRegisterLifespan", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.registration.lifespan = "100ms"
    return config
  })
})

Cypress.Commands.add("longRegisterLifespan", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.flows.registration.lifespan = "1h"
    return config
  })
})

Cypress.Commands.add("browserReturnUrlOry", ({} = {}) => {
  updateConfigFile((config) => {
    config.selfservice.allowed_return_urls = [
      "https://www.ory.sh/",
      "https://www.example.org/",
    ]
    return config
  })
})

Cypress.Commands.add("remoteCourierRecoveryTemplates", ({} = {}) => {
  updateConfigFile((config) => {
    config.courier.templates = {
      recovery: {
        invalid: {
          email: {
            body: {
              html: "base64://SGksCgp0aGlzIGlzIGEgcmVtb3RlIGludmFsaWQgcmVjb3ZlcnkgdGVtcGxhdGU=",
              plaintext:
                "base64://SGksCgp0aGlzIGlzIGEgcmVtb3RlIGludmFsaWQgcmVjb3ZlcnkgdGVtcGxhdGU=",
            },
            subject: "base64://QWNjb3VudCBBY2Nlc3MgQXR0ZW1wdGVk",
          },
        },
        valid: {
          email: {
            body: {
              html: "base64://SGksCgp0aGlzIGlzIGEgcmVtb3RlIHRlbXBsYXRlCnBsZWFzZSByZWNvdmVyIGFjY2VzcyB0byB5b3VyIGFjY291bnQgYnkgY2xpY2tpbmcgdGhlIGZvbGxvd2luZyBsaW5rOgo8YSBocmVmPSJ7eyAuUmVjb3ZlcnlVUkwgfX0iPnt7IC5SZWNvdmVyeVVSTCB9fTwvYT4=",
              plaintext:
                "base64://SGksCgp0aGlzIGlzIGEgcmVtb3RlIHRlbXBsYXRlCnBsZWFzZSByZWNvdmVyIGFjY2VzcyB0byB5b3VyIGFjY291bnQgYnkgY2xpY2tpbmcgdGhlIGZvbGxvd2luZyBsaW5rOgp7eyAuUmVjb3ZlcnlVUkwgfX0=",
            },
            subject: "base64://UmVjb3ZlciBhY2Nlc3MgdG8geW91ciBhY2NvdW50",
          },
        },
      },
    }
    return config
  })
})

Cypress.Commands.add("remoteCourierRecoveryCodeTemplates", ({} = {}) => {
  updateConfigFile((config) => {
    config.courier.templates = {
      recovery_code: {
        invalid: {
          email: {
            body: {
              html: "base64://cmVjb3ZlcnlfY29kZV9pbnZhbGlkIFJFTU9URSBURU1QTEFURSBIVE1M", // only
              plaintext:
                "base64://cmVjb3ZlcnlfY29kZV9pbnZhbGlkIFJFTU9URSBURU1QTEFURSBUWFQ=",
            },
            subject:
              "base64://cmVjb3ZlcnlfY29kZV9pbnZhbGlkIFJFTU9URSBURU1QTEFURSBTVUJKRUNU",
          },
        },
        valid: {
          email: {
            body: {
              html: "base://cmVjb3ZlcnlfY29kZV92YWxpZCBSRU1PVEUgVEVNUExBVEUgSFRNTA==",
              plaintext:
                "base64://cmVjb3ZlcnlfY29kZV92YWxpZCBSRU1PVEUgVEVNUExBVEUgVFhU",
            },
            subject:
              "base64://cmVjb3ZlcnlfY29kZV92YWxpZCBSRU1PVEUgVEVNUExBVEUgU1VCSkVDVA==",
          },
        },
      },
    }
    return config
  })
})

Cypress.Commands.add("resetCourierTemplates", (type) => {
  updateConfigFile((config) => {
    if (config?.courier?.templates && type in config.courier.templates) {
      delete config.courier.templates[type]
    }
    return config
  })
})

Cypress.Commands.add(
  "loginOidc",
  ({ app, expectSession = true, url = APP_URL + "/login", preTriggerHook }) => {
    cy.visit(url)
    if (preTriggerHook) {
      preTriggerHook()
    }
    cy.triggerOidc(app, "hydra")
    cy.location("href").should("not.eq", "/consent")
    if (expectSession) {
      // for some reason react flakes here although the login succeeded and there should be a session it fails
      if (app === "react") {
        cy.wait(2000) // adding arbitrary wait here. not sure if there is a better way in this case
      }
      cy.getSession()
    } else {
      cy.noSession()
    }
  },
)

Cypress.Commands.add(
  "login",
  ({ email, password, expectSession = true, cookieUrl = APP_URL }) => {
    if (expectSession) {
      console.log("Singing in user: ", { email, password })
    } else {
      console.log("Attempting user sign in: ", { email, password })
    }

    // see https://github.com/cypress-io/cypress/issues/408
    cy.visit(cookieUrl)
    cy.clearAllCookies()

    cy.longPrivilegedSessionTime()
    cy.request({
      url: APP_URL + "/self-service/login/browser",
      followRedirect: false,
      failOnStatusCode: false,
      headers: {
        Accept: "application/json",
      },
    })
      .then(({ body, status }) => {
        expect(status).to.eq(200)
        const form = body.ui
        return cy.request({
          method: form.method,
          body: mergeFields(form, {
            identifier: email,
            password,
            method: "password",
          }),
          headers: {
            Accept: "application/json",
          },
          url: form.action,
          followRedirect: false,
          failOnStatusCode: false,
        })
      })
      .then(({ status }) => {
        console.log("Login sequence completed: ", {
          email,
          password,
          expectSession,
        })
        if (expectSession) {
          expect(status).to.eq(200)
          return cy.getSession()
        } else {
          expect(status).to.not.eq(200)
          return cy.noSession()
        }
      })
  },
)

Cypress.Commands.add("loginMobile", ({ email, password }) => {
  cy.visit(MOBILE_URL + "/Login")
  cy.get('input[data-testid="identifier"]').type(email)
  cy.get('input[data-testid="password"]').type(password)
  cy.get('div[data-testid="submit-form"]').click()
})

Cypress.Commands.add("logout", () => {
  cy.getCookies().should((cookies) => {
    const c = cookies.find(
      ({ name }) => name.indexOf("ory_kratos_session") > -1,
    )
    expect(c).to.not.be.undefined
    cy.clearCookie(c.name)
  })
  cy.noSession()
})

Cypress.Commands.add(
  "reauthWithOtherAccount",
  ({
    previousUrl,
    expect: { email, success = true },
    type: { email: temail, password: tpassword } = {
      email: undefined,
      password: undefined,
    },
  }) => {
    cy.location("pathname").should("contain", "/login")
    cy.location().then((loc) => {
      const uri = new URLSearchParams(loc.search)
      const flow = uri.get("flow")
      expect(flow).to.not.be.empty
      cy.request({
        url: APP_URL + `/self-service/login/flows?id=${flow}`,
        followRedirect: false,
        failOnStatusCode: false,
        headers: {
          Accept: "application/json",
        },
      }).then(({ body, status }) => {
        expect(status).to.eq(200)
        const form = body.ui
        console.log(form.action)
        return cy
          .request({
            method: form.method,
            body: mergeFields(form, {
              identifier: temail || email,
              password: tpassword,
              method: "password",
            }),
            headers: {
              Accept: "application/json",
              ContentType: "application/json",
            },
            url: form.action,
            followRedirect: false,
            failOnStatusCode: false,
          })
          .then((res) => {
            expect(res.status).to.eq(200)
            cy.visit(previousUrl)
          })
      })
    })
  },
)
Cypress.Commands.add(
  "reauth",
  ({
    expect: { email, success = true },
    type: { email: temail, password: tpassword } = {
      email: undefined,
      password: undefined,
    },
  }) => {
    cy.location("pathname").should("contain", "/login")
    cy.get('input[name="identifier"]').should("have.value", email)
    if (temail) {
      cy.get('input[name="identifier"]').invoke("attr", "value", temail)
    }
    if (tpassword) {
      cy.get('input[name="password"]').clear().type(tpassword)
    }
    cy.longPrivilegedSessionTime()
    cy.get('button[value="password"]').click()
    if (success) {
      cy.location("pathname").should("not.contain", "/login")
    }
  },
)

Cypress.Commands.add("deleteMail", ({ atLeast = 0 } = {}) => {
  let tries = 0
  let count = 0
  const req = () =>
    cy
      .request("DELETE", `${MAIL_API}/mail`, { pruneCode: "all" })
      .then(({ body }) => {
        count += parseInt(body)
        if (count < atLeast && tries < 100) {
          cy.log(
            `Expected at least ${atLeast} messages but deleteted only ${count} so far (body: ${body})`,
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
  "getSession",
  ({ expectAal = "aal1", expectMethods = [] } = {}) => {
    // Do the request once to ensure we have a session (with retry)
    cy.request({
      method: "GET",
      url: `${KRATOS_PUBLIC}/sessions/whoami`,
    })
      .its("status") // adds retry
      .should("eq", 200)

    // Return the session for further propagation
    return cy
      .request("GET", `${KRATOS_PUBLIC}/sessions/whoami`)
      .then((response) => {
        expect(response.body.id).to.not.be.empty
        expect(dayjs().isBefore(dayjs(response.body.expires_at))).to.be.true

        // Add a grace second for MySQL which does not support millisecs.
        expect(dayjs().isAfter(dayjs(response.body.issued_at).subtract(1, "s")))
          .to.be.true
        expect(
          dayjs().isAfter(
            dayjs(response.body.authenticated_at).subtract(1, "s"),
          ),
        ).to.be.true

        expect(response.body.identity).to.exist

        expect(response.body.authenticator_assurance_level).to.equal(expectAal)
        if (expectMethods.length > 0) {
          expect(response.body.authentication_methods).to.have.lengthOf(
            expectMethods.length,
          )
          expectMethods.forEach((value) => {
            expect(
              response.body.authentication_methods.find(
                ({ method }) => method === value,
              ),
            ).to.exist
          })
        }

        return response.body
      })
  },
)

Cypress.Commands.add("noSession", () =>
  cy
    .request({
      method: "GET",
      url: `${KRATOS_PUBLIC}/sessions/whoami`,
      failOnStatusCode: false,
    })
    .then((request) => {
      expect(request.status).to.eq(401)
      return request
    }),
)

Cypress.Commands.add(
  "performEmailVerification",
  ({ expect: { email, redirectTo }, strategy = "code" }) => {
    cy.getMail().then((message) => {
      expect(message.subject).to.equal("Please verify your email address")
      expect(message.fromAddress.trim()).to.equal("no-reply@ory.kratos.sh")
      expect(message.toAddresses).to.have.length(1)
      expect(message.toAddresses[0].trim()).to.equal(email)

      const link = parseHtml(message.body).querySelector("a")
      expect(link).to.not.be.null
      expect(link.href).to.contain(APP_URL)
      const params = new URL(link.href).searchParams

      cy.visit(link.href)
      if (strategy === "code") {
        const code = params.get("code")
        expect(code).to.not.be.null
        cy.get(`button[name="method"][value="code"]`).click()
      }

      if (redirectTo) {
        cy.get(`[data-testid="node/anchor/continue"`)
          .contains("Continue")
          .click()
        cy.url().should("be.equal", redirectTo)
      }
    })
  },
)

Cypress.Commands.add(
  "verifyEmail",
  ({ expect: { email, password, redirectTo }, strategy }) => {
    cy.performEmailVerification({
      expect: { email, redirectTo },
      strategy,
    }).then(() => {
      cy.getSession().should((session) =>
        assertVerifiableAddress({ email, isVerified: true })(session),
      )
    })
  },
)

// Uses the verification email but waits so that it expires
Cypress.Commands.add("recoverEmailButExpired", ({ expect: { email } }) => {
  cy.getMail().should((message) => {
    expect(message.subject).to.equal("Recover access to your account")
    expect(message.toAddresses[0].trim()).to.equal(email)

    const link = parseHtml(message.body).querySelector("a")
    expect(link).to.not.be.null
    expect(link.href).to.contain(APP_URL)

    cy.visit(link.href)
  })
})

Cypress.Commands.add(
  "recoveryEmailWithCode",
  ({ expect: { email, enterCode = true } }) => {
    cy.getMail({ removeMail: true }).should((message) => {
      expect(message.subject).to.equal("Recover access to your account")
      expect(message.toAddresses[0].trim()).to.equal(email)

      const code = extractRecoveryCode(message.body)
      expect(code).to.not.be.undefined
      expect(code.length).to.equal(6)
      if (enterCode) {
        cy.get("input[name='code']").type(code)
      }
    })
  },
)

Cypress.Commands.add(
  "recoverEmail",
  ({ expect: { email }, shouldVisit = true }) =>
    cy.getMail().should((message) => {
      expect(message.subject).to.equal("Recover access to your account")
      expect(message.fromAddress.trim()).to.equal("no-reply@ory.kratos.sh")
      expect(message.toAddresses).to.have.length(1)
      expect(message.toAddresses[0].trim()).to.equal(email)

      const link = parseHtml(message.body).querySelector("a")
      expect(link).to.not.be.null
      expect(link.href).to.contain(APP_URL)

      if (shouldVisit) {
        cy.visit(link.href)
      }
      return link.href
    }),
)

// Uses the verification email but waits so that it expires
Cypress.Commands.add(
  "verifyEmailButExpired",
  ({ expect: { email }, strategy = "code" }) => {
    cy.getMail().then((message) => {
      expect(message.subject).to.equal("Please verify your email address")

      expect(message.fromAddress.trim()).to.equal("no-reply@ory.kratos.sh")
      expect(message.toAddresses).to.have.length(1)
      expect(message.toAddresses[0].trim()).to.equal(email)

      const link = parseHtml(message.body).querySelector("a")
      cy.getSession().should((session) => {
        assertVerifiableAddress({
          isVerified: false,
          email: email,
        })(session)
      })

      cy.visit(link.href)
      if (strategy === "code") {
        cy.get('button[name="method"][value="code"]').click()
      }
      cy.get('[data-testid="ui/message/4070005"]').should(
        "contain.text",
        "verification flow expired",
      )
      cy.location("pathname").should("include", "verification")

      cy.getSession().should((session) => {
        assertVerifiableAddress({
          isVerified: false,
          email: email,
        })(session)
      })
    })
  },
)

Cypress.Commands.add("useVerificationStrategy", (strategy: Strategy) => {
  cy.updateConfigFile((config) => {
    config.selfservice.flows.verification.use = strategy
    return config
  })
})

Cypress.Commands.add("useLookupSecrets", (value: boolean) => {
  cy.updateConfigFile((config) => {
    config.selfservice.methods = {
      ...config.selfservice.methods,
      lookup_secret: {
        enabled: value,
      },
    }
    return config
  })
})

Cypress.Commands.add("getLookupSecrets", () =>
  cy
    .get('[data-testid="node/text/lookup_secret_codes/text"] code')
    .then(($e) => $e.map((_, e) => e.innerText.trim()).toArray()),
)
Cypress.Commands.add("expectSettingsSaved", () => {
  cy.get('[data-testid="ui/message/1050001"]').should(
    "contain.text",
    "Your changes have been saved",
  )
})

Cypress.Commands.add(
  "getMail",
  ({ removeMail = true, expectedCount = 1, email = undefined } = {}) => {
    let tries = 0
    const req = () =>
      cy.request(`${MAIL_API}/mail`).then((response) => {
        expect(response.body).to.have.property("mailItems")
        const count = response.body.mailItems.length
        if (count === 0 && tries < 100) {
          tries++
          cy.wait(pollInterval)
          return req()
        }
        let mailItem: any
        if (email) {
          mailItem = response.body.mailItems.find((m: any) =>
            m.toAddresses.includes(email),
          )
        } else {
          mailItem = response.body.mailItems[0]
        }
        console.log({ mailItems: response.body.mailItems })
        console.log({ mailItem })
        console.log({ email })

        expect(count).to.equal(expectedCount)
        if (removeMail) {
          return cy.deleteMail({ atLeast: count }).then(() => {
            return Promise.resolve(mailItem)
          })
        }

        return Promise.resolve(mailItem)
      })

    return req()
  },
)

Cypress.Commands.add("clearAllCookies", () => {
  cy.clearCookies({ domain: null })
})

Cypress.Commands.add("submitPasswordForm", () => {
  cy.get('[name="method"][value="password"]').click()
  cy.get('[name="method"][value="password"]:disabled').should("not.exist")
})

Cypress.Commands.add("submitProfileForm", () => {
  cy.get('[name="method"][value="profile"]').click()
  cy.get('[name="method"][value="profile"]:disabled').should("not.exist")
})

Cypress.Commands.add("submitCodeForm", () => {
  cy.get('button[name="method"][value="code"]').click()
  cy.get('button[name="method"][value="code"]:disabled').should("not.exist")
})

Cypress.Commands.add("clickWebAuthButton", (type: string) => {
  cy.get('*[data-testid="node/script/webauthn_script"]').should("exist")
  cy.wait(500) // Wait for script to load
  cy.get('*[name="webauthn_' + type + '_trigger"]').click()
  cy.wait(500) // Wait webauth to pass
})

Cypress.Commands.add("shouldShow2FAScreen", () => {
  cy.location().should((loc) => {
    expect(loc.pathname).to.include("/login")
  })
  cy.get("h2").should("contain.text", "Two-Factor Authentication")
  cy.get('[data-testid="ui/message/1010004"]').should(
    "contain.text",
    "Please complete the second authentication challenge.",
  )
})

Cypress.Commands.add(
  "shouldErrorOnDisallowedReturnTo",
  (init: string, { app }: { app: "express" | "react" }) => {
    cy.visit(init, { failOnStatusCode: false })
    if (app === "react") {
      cy.location("href").should("include", init.split("?")[0])
      cy.get(".Toastify").should(
        "contain.text",
        "The return_to address is not allowed.",
      )
    } else {
      cy.location("pathname").should("contain", "error")
      cy.get("div").should(
        "contain.text",
        'Requested return_to URL "https://not-allowed" is not allowed.',
      )
    }
  },
)

Cypress.Commands.add(
  "shouldHaveCsrfError",
  ({ app }: { app: "express" | "react" }) => {
    let initial
    let pathname
    cy.location().should((location) => {
      initial = location.search
      pathname = location.pathname
    })

    cy.getCookies().should((cookies) => {
      const csrf = cookies.find(({ name }) => name.indexOf("csrf") > -1)
      expect(csrf).to.not.be.undefined
      cy.clearCookie(csrf.name)
    })
    cy.submitPasswordForm()

    // We end up at a new flow
    if (app === "express") {
      cy.location().should((location) => {
        expect(initial).to.not.be.empty
        expect(location.search).to.not.eq(initial)
      })

      cy.location("pathname").should("include", "/error")
      cy.get(`div`).should("contain.text", "CSRF")
    } else {
      cy.location("pathname").should((got) => {
        expect(got).to.eql(pathname)
      })
      cy.get(".Toastify").should(
        "contain.text",
        "A security violation was detected, please fill out the form again.",
      )
    }
  },
)

Cypress.Commands.add(
  "triggerOidc",
  (app: "react" | "express", provider: string = "hydra") => {
    let initial, didHaveSearch
    cy.location().then((loc) => {
      didHaveSearch = loc.search.length > 0
      initial = loc.pathname + loc.search
    })
    cy.get('[name="provider"][value="' + provider + '"]').click()
    cy.location().should((loc) => {
      if (app === "express" || didHaveSearch) {
        return
      }
      expect(loc.pathname + loc.search).not.to.eql(initial)
    })
  },
)

Cypress.Commands.add(
  "removeAttribute",
  (selectors: string[], attribute: string) => {
    selectors.forEach((selector) => {
      cy.get(selector).then(($el) => {
        $el.removeAttr(attribute)
      })
    })
  },
)

Cypress.Commands.add(
  "addInputElement",
  (parent: string, attribute: string, value: string) => {
    cy.get(parent).then(($el) => {
      $el.append(`<input type="hidden" name="${attribute}" value="${value}" />`)
    })
  },
)

Cypress.Commands.add(
  "notifyUnknownRecipients",
  (flow: "recovery" | "verification", value: boolean = true) => {
    cy.updateConfigFile((config) => {
      config.selfservice.flows[flow].notify_unknown_recipients = value
      return config
    })
  },
)

Cypress.Commands.add("getCourierMessages", () => {
  return cy.request(KRATOS_ADMIN + "/courier/messages").then((res) => {
    return res.body
  })
})

Cypress.Commands.add(
  "enableVerificationUIAfterRegistration",
  (strategy: "password" | "oidc" | "webauthn") => {
    cy.updateConfigFile((config) => {
      if (!config.selfservice.flows.registration.after[strategy]) {
        config.selfservice.flows.registration.after = {
          [strategy]: { hooks: [] },
        }
      }

      const hooks =
        config.selfservice.flows.registration.after[strategy].hooks || []
      config.selfservice.flows.registration.after[strategy].hooks = [
        ...hooks.filter((h) => h.hook !== "show_verification_ui"),
        { hook: "show_verification_ui" },
      ]
      return config
    })
  },
)

Cypress.Commands.add("getVerificationCodeFromEmail", (email) => {
  return cy
    .getMail({ removeMail: true })
    .should((message) => {
      expect(message.subject).to.equal("Please verify your email address")
      expect(message.toAddresses[0].trim()).to.equal(email)
    })
    .then((message) => {
      const code = extractRecoveryCode(message.body)
      expect(code).to.not.be.undefined
      expect(code.length).to.equal(6)
      return code
    })
})

Cypress.Commands.add("enableRegistrationViaCode", (enable: boolean = true) => {
  cy.updateConfigFile((config) => {
    config.selfservice.methods.code.registration_enabled = enable
    return config
  })
})

Cypress.Commands.add("getRegistrationCodeFromEmail", (email, opts) => {
  return cy
    .getMail({ removeMail: true, email, ...opts })
    .should((message) => {
      expect(message.subject).to.equal("Complete your account registration")
      expect(message.toAddresses[0].trim()).to.equal(email)
    })
    .then((message) => {
      const code = extractRecoveryCode(message.body)
      expect(code).to.not.be.undefined
      expect(code.length).to.equal(6)
      return code
    })
})

Cypress.Commands.add("getLoginCodeFromEmail", (email, opts) => {
  return cy
    .getMail({ removeMail: true, email, ...opts })
    .should((message) => {
      expect(message.subject).to.equal("Login to your account")
      expect(message.toAddresses[0].trim()).to.equal(email)
    })
    .then((message) => {
      const code = extractRecoveryCode(message.body)
      expect(code).to.not.be.undefined
      expect(code.length).to.equal(6)
      return code
    })
})

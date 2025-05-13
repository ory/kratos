// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { Identity } from "@ory/kratos-client"
import {
  APIRequestContext,
  CDPSession,
  expect as baseExpect,
  Page,
  test as base,
} from "@playwright/test"
import { writeFile } from "fs/promises"
import { merge } from "lodash"
import { OryKratosConfiguration } from "../../shared/config"
import { default_config } from "../setup/default_config"
import { APIResponse } from "playwright-core"
import { SessionWithResponse } from "../types"
import { retryOptions } from "../lib/request"
import promiseRetry from "promise-retry"
import { Protocol } from "playwright-core/types/protocol"
import { createIdentityWithPassword } from "../actions/identity"
import { randomBytes } from "crypto"

// from https://stackoverflow.com/questions/61132262/typescript-deep-partial
type DeepPartial<T> = T extends object
  ? {
      [P in keyof T]?: DeepPartial<T[P]>
    }
  : T

type TestFixtures = {
  identity: { oryIdentity: Identity; email: string; password: string }
  configOverride: DeepPartial<OryKratosConfiguration>
  config: OryKratosConfiguration
  virtualAuthenticatorOptions: Partial<Protocol.WebAuthn.VirtualAuthenticatorOptions>
  pageCDPSession: CDPSession
  virtualAuthenticator: Protocol.WebAuthn.addVirtualAuthenticatorReturnValue
}

type WorkerFixtures = {
  kratosAdminURL: string
  kratosPublicURL: string
  mode:
    | "reconfigure_kratos"
    | "reconfigure_ory_network_project"
    | "existing_kratos"
    | "existing_ory_network_project"
}

export const test = base.extend<TestFixtures, WorkerFixtures>({
  configOverride: {},
  config: [
    async ({ request, configOverride }, use) => {
      const configToWrite = merge(default_config, configOverride)

      const revision = randomBytes(16).toString("hex")
      const fileDirectory = __dirname + "/../.."
      await writeFile(
        fileDirectory + "/playwright/kratos.config.json",
        JSON.stringify(
          {
            ...configToWrite,
            // Forces a new hash, even if the config was not changed.
            revision,
          },
          null,
          2,
        ),
      )

      await expect(async () => {
        const resp = await request.get(
          "http://localhost:4434/admin/health/config",
        )
        const updatedRevision = (await resp.body()).toString()
        expect(updatedRevision).toBe(revision)
      }).toPass()

      await use(configToWrite)
    },
    { auto: true },
  ],
  virtualAuthenticatorOptions: undefined,
  pageCDPSession: async ({ page }, use) => {
    const cdpSession = await page.context().newCDPSession(page)
    await use(cdpSession)
    await cdpSession.detach()
  },
  virtualAuthenticator: async (
    { pageCDPSession, virtualAuthenticatorOptions },
    use,
  ) => {
    await pageCDPSession.send("WebAuthn.enable")
    const { authenticatorId } = await pageCDPSession.send(
      "WebAuthn.addVirtualAuthenticator",
      {
        options: {
          protocol: "ctap2",
          transport: "internal",
          hasResidentKey: true,
          hasUserVerification: true,
          isUserVerified: true,
          ...virtualAuthenticatorOptions,
        },
      },
    )
    await use({ authenticatorId })
    await pageCDPSession.send("WebAuthn.removeVirtualAuthenticator", {
      authenticatorId,
    })

    await pageCDPSession.send("WebAuthn.disable")
  },
  identity: async ({ request }, use, i) => {
    const {
      identity: oryIdentity,
      password,
      email,
    } = await createIdentityWithPassword(request)
    i.attach("identity", {
      body: JSON.stringify(oryIdentity, null, 2),
      contentType: "application/json",
    })
    await use({
      oryIdentity,
      email,
      password,
    })
  },
  kratosAdminURL: ["http://localhost:4434", { option: true, scope: "worker" }],
  kratosPublicURL: ["http://localhost:4433", { option: true, scope: "worker" }],
})

export const expect = baseExpect.extend({
  toHaveSession,
  toMatchResponseData,
})

async function toHaveSession(
  requestOrPage: APIRequestContext | Page,
  baseUrl: string,
) {
  let r: APIRequestContext
  if ("request" in requestOrPage) {
    r = requestOrPage.request
  } else {
    r = requestOrPage
  }
  let pass = true

  let responseData: string
  let response: APIResponse = null
  try {
    const result = await promiseRetry(
      () =>
        r
          .get(baseUrl + "/sessions/whoami", {
            failOnStatusCode: false,
          })
          .then<SessionWithResponse>(
            async (res: APIResponse): Promise<SessionWithResponse> => {
              return {
                session: await res.json(),
                response: res,
              }
            },
          ),
      retryOptions,
    )
    pass = !!result.session.active
    responseData = await result.response.text()
    response = result.response
  } catch (e) {
    pass = false
    responseData = JSON.stringify(e.message, undefined, 2)
  }

  const message = () =>
    this.utils.matcherHint("toHaveSession", undefined, undefined, {
      isNot: this.isNot,
    }) +
    `\n
    \n
    Expected: ${this.isNot ? "not" : ""} to have session\n
    Session data received: ${responseData}\n
    Headers: ${JSON.stringify(response?.headers(), null, 2)}\n
    `

  return {
    message,
    pass,
    name: "toHaveSession",
  }
}

async function toMatchResponseData(
  res: APIResponse,
  options: {
    statusCode?: number
    failureHint?: string
  },
) {
  const body = await res.text()
  const statusCode = options.statusCode ?? 200
  const failureHint = options.failureHint ?? ""
  const message = () =>
    this.utils.matcherHint("toMatch", undefined, undefined, {
      isNot: this.isNot,
    }) +
    `\n
    ${failureHint}
    \n
    Expected: ${this.isNot ? "not" : ""} to match\n
    Status Code: ${statusCode}\n
    Body: ${body}\n
    Headers: ${JSON.stringify(res.headers(), null, 2)}\n
    URL: ${JSON.stringify(res.url(), null, 2)}\n
    `

  return {
    message,
    pass: res.status() === statusCode,
    name: "toMatch",
  }
}

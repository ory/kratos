// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { faker } from "@faker-js/faker"
import { Identity } from "@ory/kratos-client"
import { CDPSession, test as base, expect } from "@playwright/test"
import { writeFile } from "fs/promises"
import { merge } from "lodash"
import { OryKratosConfiguration } from "../../shared/config"
import { default_config } from "../setup/default_config"

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
  addVirtualAuthenticator: boolean
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

      const resp = await request.get("http://localhost:4434/health/config")

      const configRevision = await resp.body()

      const fileDirectory = __dirname + "/../.."

      await writeFile(
        fileDirectory + "/playwright/kratos.config.json",
        JSON.stringify(configToWrite, null, 2),
      )
      await expect(async () => {
        const resp = await request.get("http://localhost:4434/health/config")
        const updatedRevision = await resp.body()
        expect(updatedRevision).not.toBe(configRevision)
      }).toPass()

      await use(configToWrite)
    },
    { auto: true },
  ],
  addVirtualAuthenticator: false,
  page: async ({ page, addVirtualAuthenticator }, use) => {
    let cdpSession: CDPSession
    let authenticatorId = ""
    if (addVirtualAuthenticator) {
      cdpSession = await page.context().newCDPSession(page)
      await cdpSession.send("WebAuthn.enable")
      const { authenticatorId: aid } = await cdpSession.send(
        "WebAuthn.addVirtualAuthenticator",
        {
          options: {
            protocol: "ctap2",
            transport: "internal",
            hasResidentKey: true,
            hasUserVerification: true,
            isUserVerified: true,
          },
        },
      )
      authenticatorId = aid
    }
    await use(page)
    if (addVirtualAuthenticator) {
      await cdpSession.send("WebAuthn.removeVirtualAuthenticator", {
        authenticatorId,
      })

      await cdpSession.send("WebAuthn.disable")
      await cdpSession.detach()
    }
  },
  identity: async ({ request }, use, i) => {
    const email = faker.internet.email({ provider: "ory.sh" })
    const password = faker.internet.password()
    const resp = await request.post("http://localhost:4434/admin/identities", {
      data: {
        schema_id: "email",
        traits: {
          email,
          website: faker.internet.url(),
        },

        credentials: {
          password: {
            config: {
              password,
            },
          },
        },
      },
    })
    const oryIdentity = await resp.json()
    i.attach("identity", {
      body: JSON.stringify(oryIdentity, null, 2),
      contentType: "application/json",
    })
    expect(resp.status()).toBe(201)
    await use({
      oryIdentity,
      email,
      password,
    })
  },
  kratosAdminURL: ["http://localhost:4434", { option: true, scope: "worker" }],
  kratosPublicURL: ["http://localhost:4433", { option: true, scope: "worker" }],
})

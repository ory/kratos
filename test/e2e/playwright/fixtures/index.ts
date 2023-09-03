// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { Identity } from "@ory/kratos-client"
import { test as base, expect } from "@playwright/test"
import { OryKratosConfiguration } from "../../cypress/support/config"
import { merge } from "lodash"
import { default_config } from "../setup/default_config"
import { writeFile } from "fs/promises"
import { faker } from "@faker-js/faker"

// from https://stackoverflow.com/questions/61132262/typescript-deep-partial
type DeepPartial<T> = T extends object
  ? {
      [P in keyof T]?: DeepPartial<T[P]>
    }
  : T

type TestFixtures = {
  identity: Identity
  configOverride: DeepPartial<OryKratosConfiguration>
  config: void
}

type WorkerFixtures = {}

export const test = base.extend<TestFixtures, WorkerFixtures>({
  configOverride: {},
  config: [
    async ({ request, configOverride }, use) => {
      const configToWrite = merge(default_config, configOverride)

      const resp = await request.get("http://localhost:4434/health/config")

      const configRevision = await resp.body()

      await writeFile(
        "playwright/kratos.config.json",
        JSON.stringify(configToWrite),
      )
      await expect(async () => {
        const resp = await request.get("http://localhost:4434/health/config")
        const updatedRevision = await resp.body()
        expect(updatedRevision).not.toBe(configRevision)
      }).toPass()

      await use()
    },
    { auto: true },
  ],
  identity: async ({ request }, use) => {
    const resp = await request.post("http://localhost:4434/admin/identities", {
      data: {
        schema_id: "email",
        traits: {
          email: faker.internet.email(undefined, undefined, "ory.sh"),
          website: faker.internet.url(),
        },
      },
    })
    expect(resp.status()).toBe(201)
    await use(await resp.json())
  },
})

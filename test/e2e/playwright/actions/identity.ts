// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { faker } from "@faker-js/faker"
import { APIRequestContext } from "@playwright/test"
import { CreateIdentityBody } from "@ory/kratos-client"
import { generatePhoneNumber, CountryNames } from "phone-number-generator-js"
import { expect } from "../fixtures"

export async function createIdentity(
  request: APIRequestContext,
  data: Partial<CreateIdentityBody>,
) {
  const resp = await request.post("http://localhost:4434/admin/identities", {
    data,
  })
  expect(resp.status()).toBe(201)
  return await resp.json()
}

export async function createIdentityWithPhoneNumber(
  request: APIRequestContext,
) {
  const phone = generatePhoneNumber({
    countryName: CountryNames.Germany,
    withoutCountryCode: false,
  })
  return {
    identity: await createIdentity(request, {
      schema_id: "sms",
      traits: {
        phone,
      },
    }),
    phone,
  }
}

export async function createIdentityWithEmail(request: APIRequestContext) {
  const email = faker.internet.email({ provider: "ory.sh" })
  return {
    identity: await createIdentity(request, {
      schema_id: "email",
      traits: {
        email,
        website: faker.internet.url(),
      },
    }),
    email,
  }
}

export async function createIdentityWithPassword(request: APIRequestContext) {
  const email = faker.internet.email({ provider: "ory.sh" })
  const password = faker.internet.password()
  return {
    identity: await createIdentity(request, {
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
    }),
    email,
    password,
  }
}

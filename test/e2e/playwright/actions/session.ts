// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { APIRequestContext, expect } from "@playwright/test"
import { Session } from "@ory/kratos-client"

export async function hasSession(
  r: APIRequestContext,
  kratosPublicURL: string,
): Promise<void> {
  const resp = await r.get(kratosPublicURL + "/sessions/whoami", {
    failOnStatusCode: true,
  })
  const session = await resp.json()
  expect(session).toBeDefined()
  expect(session.active).toBe(true)
}

export async function getSession(
  r: APIRequestContext,
  kratosPublicURL: string,
): Promise<Session> {
  const resp = await r.get(kratosPublicURL + "/sessions/whoami", {
    failOnStatusCode: true,
  })
  return resp.json()
}

export async function hasNoSession(
  r: APIRequestContext,
  kratosPublicURL: string,
): Promise<void> {
  const resp = await r.get(kratosPublicURL + "/sessions/whoami", {
    failOnStatusCode: false,
  })
  expect(resp.status()).toBe(401)
  return resp.json()
}

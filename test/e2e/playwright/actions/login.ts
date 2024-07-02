// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { APIRequestContext } from "@playwright/test"
import { findCsrfToken } from "../lib/helper"
import { LoginFlow, Session } from "@ory/kratos-client"
import { expectJSONResponse } from "../lib/request"
import { expect } from "../fixtures"

export async function loginWithPassword(
  user: { password: string; traits: { email: string } },
  r: APIRequestContext,
  baseUrl: string,
): Promise<void> {
  const { ui } = await expectJSONResponse<LoginFlow>(
    await r.get(baseUrl + "/self-service/login/browser", {
      headers: {
        Accept: "application/json",
      },
    }),
    {
      message: "Initializing login flow failed",
    },
  )

  const res = await r.post(ui.action, {
    headers: {
      Accept: "application/json",
    },
    data: {
      identifier: user.traits.email,
      password: user.password,
      method: "password",
      csrf_token: findCsrfToken(ui),
    },
  })
  const { session } = await expectJSONResponse<{ session: Session }>(res)
  expect(session?.identity?.traits.email).toEqual(user.traits.email)
  expect(
    res.headersArray().find(
      ({ name, value }) =>
        name.toLowerCase() === "set-cookie" &&
        (value.indexOf("ory_session_") > -1 || // Ory Network
          value.indexOf("ory_kratos_session") > -1), // Locally hosted
    ),
  ).toBeDefined()
}

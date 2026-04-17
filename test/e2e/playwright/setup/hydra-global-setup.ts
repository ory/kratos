// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { request } from "@playwright/test"

/**
 * Creates a Hydra OAuth2 client and exposes its credentials via process.env so
 * that test workers see the generated client_id / client_secret when they load
 * the oidcConfig constants.  Playwright propagates process.env mutations from
 * globalSetup to all worker processes.
 */
async function globalSetup() {
  const context = await request.newContext()
  try {
    const resp = await context.post("http://localhost:4445/admin/clients", {
      data: {
        grant_types: ["authorization_code", "refresh_token"],
        response_types: ["code", "id_token"],
        scope: "openid offline",
        redirect_uris: [
          "http://localhost:4455/self-service/methods/oidc/callback/hydra",
        ],
      },
    })

    if (resp.status() !== 201) {
      throw new Error(
        `Failed to create Hydra OAuth2 client: ${resp.status()} ${await resp.text()}`,
      )
    }

    const client = await resp.json()
    process.env["OIDC_HYDRA_CLIENT_ID"] = client.client_id
    process.env["OIDC_HYDRA_CLIENT_SECRET"] = client.client_secret
  } finally {
    await context.dispose()
  }
}

export default globalSetup

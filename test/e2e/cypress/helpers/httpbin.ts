// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import * as oauth2 from "./oauth2"

export function checkToken(
  client: oauth2.oAuth2Client,
  scope: string[],
  check: (token: any) => void,
) {
  cy.location("href")
    .should("match", new RegExp("https://httpbin.org/anything[?]code=.*"))
    .then((body) => {
      cy.get("body")
        .invoke("text")
        .then((text) => {
          const result = JSON.parse(text)
          const tokenParams = {
            code: result.args.code,
            redirect_uri: "https://httpbin.org/anything",
            scope: scope.join(" "),
          }
          oauth2
            .getToken(
              client.token_endpoint,
              client.id,
              client.secret,
              "authorization_code",
              tokenParams.code,
              tokenParams.redirect_uri,
              tokenParams.scope,
            )
            .then((res) => {
              const token = res.body
              check(token)
            })
        })
    })
}

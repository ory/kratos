// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import * as uuid from "uuid"

export type oAuth2Client = {
  auth_endpoint: string
  token_endpoint: string
  id: string
  secret: string
  token_endpoint_auth_method: string
  grant_types: string[]
  response_types: string[]
  scopes: string[]
  callbacks: string[]
}

export function getDefaultAuthorizeURL(client: oAuth2Client) {
  const state = uuid.v4()
  const nonce = uuid.v4()
  return getAuthorizeURL(
    client.auth_endpoint,
    "",
    client.id,
    undefined,
    nonce,
    "https://httpbin.org/anything",
    "code",
    ["offline", "openid"],
    state,
    undefined,
  )
}

export function getAuthorizeURL(
  auth_endpoint: string,
  audience: string,
  client_id: string,
  max_age: string | undefined,
  nonce: string,
  redirect_uri: string,
  response_type:
    | "code"
    | "id_token"
    | "id_token token"
    | "code id_token"
    | "code token"
    | "code id_token token",
  scopes: string[],
  state: string,
  code_challenge?: string,
): string {
  const r = new URL(auth_endpoint)
  r.searchParams.append("audience", audience)
  r.searchParams.append("client_id", client_id)
  if (max_age !== undefined) {
    r.searchParams.append("max_age", max_age)
  }
  r.searchParams.append("nonce", nonce)
  r.searchParams.append("prompt", "")
  r.searchParams.append("redirect_uri", redirect_uri)
  r.searchParams.append("response_type", response_type)
  r.searchParams.append("scope", scopes.join(" "))
  r.searchParams.append("state", state)

  code_challenge && r.searchParams.append("code_challenge", code_challenge)
  return r.toString()
}

export function getToken(
  token_endpoint: string,
  client_id: string,
  client_secret: string,
  grant_type: "authorization_code",
  code: string,
  redirect_uri: string,
  scope: string,
) {
  let urlEncodedData = ""
  const urlEncodedDataPairs = []
  urlEncodedDataPairs.push(
    encodeURIComponent("grant_type") + "=" + encodeURIComponent(grant_type),
  )
  urlEncodedDataPairs.push(
    encodeURIComponent("code") + "=" + encodeURIComponent(code),
  )
  urlEncodedDataPairs.push(
    encodeURIComponent("redirect_uri") + "=" + encodeURIComponent(redirect_uri),
  )
  urlEncodedDataPairs.push(
    encodeURIComponent("scope") + "=" + encodeURIComponent(scope),
  )

  urlEncodedData = urlEncodedDataPairs.join("&").replace(/%20/g, "+")

  return cy.request({
    method: "POST",
    url: token_endpoint,
    form: true,
    body: urlEncodedData,
    headers: {
      Accept: "application/json",
      Authorization: "Basic " + btoa(client_id + ":" + client_secret),
    },
  })
}

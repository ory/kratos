export function getAuthorizeURL(
  auth_endpoint: string,
  audience: string,
  client_id: string,
  max_age: string,
  nonce: string,
  redirect_uri: string,
  response_type:
    | 'code'
    | 'id_token'
    | 'id_token token'
    | 'code id_token'
    | 'code token'
    | 'code id_token token',
  scopes: string[],
  state: string,
  code_challenge?: string
): string {
  const r = new URL(auth_endpoint)
  r.searchParams.append('audience', audience)
  r.searchParams.append('client_id', client_id)
  r.searchParams.append('max_age', max_age)
  r.searchParams.append('nonce', nonce)
  r.searchParams.append('prompt', '')
  r.searchParams.append('redirect_uri', redirect_uri)
  r.searchParams.append('response_type', response_type)
  r.searchParams.append('scope', scopes.join(' '))
  r.searchParams.append('state', state)

  code_challenge && r.searchParams.append('code_challenge', code_challenge)
  return r.toString()
}

export function getToken(
  token_endpoint: string,
  client_id: string,
  client_secret: string,
  grant_type: 'authorization_code',
  code: string,
  redirect_uri: string,
  scope: string
) {
  let urlEncodedData = ''
  const urlEncodedDataPairs = []
  urlEncodedDataPairs.push(
    encodeURIComponent('grant_type') + '=' + encodeURIComponent(grant_type)
  )
  urlEncodedDataPairs.push(
    encodeURIComponent('code') + '=' + encodeURIComponent(code)
  )
  urlEncodedDataPairs.push(
    encodeURIComponent('redirect_uri') + '=' + encodeURIComponent(redirect_uri)
  )
  urlEncodedDataPairs.push(
    encodeURIComponent('scope') + '=' + encodeURIComponent(scope)
  )

  urlEncodedData = urlEncodedDataPairs.join('&').replace(/%20/g, '+')

  return cy.request({
    method: 'POST',
    url: token_endpoint,
    form: true,
    body: urlEncodedData,
    headers: {
      Accept: 'application/json',
      Authorization: 'Basic ' + btoa(client_id + ':' + client_secret)
    }
  })
}

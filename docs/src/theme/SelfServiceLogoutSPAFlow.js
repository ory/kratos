import React from 'react'
import Mermaid from './Mermaid'

const chart = `
sequenceDiagram

  participant B as Browser
  participant A as Client-Side UI, e.g. ReactJS
  participant K as Ory Kratos

  B->>A: Makes request with Ory Session Cookie
  A->>K: Requests logout URL for given Ory Session Cookie
  K->>A: Returns logout URL
  A-->>A: Renders logout URL in UI / HTML
  A->>B: Returns HTML
  B->>K: Opens logout URL
  alt Logout URL is valid
    K-->>K: Invalidates session
    K->>B: Redirects to post logout return address.
  else Logout URL is invalid
    K->>B: Redirect to error UI.
  end
`

const SelfServiceBrowserFlow = () => <Mermaid chart={chart} />

export default SelfServiceBrowserFlow

import React from 'react'
import Mermaid from './Mermaid'

const chart = ({
  flows = ['login', 'registration', 'settings', '...'],
  interactions = ['"Log in"', '"Sign Up"', '"Update Email"', '...'],
  success = 'Perform flow-specific action (e.g. create user, set session cookie, ...)'
}) => {
  const components =
    flows.length > 1 ? `<${flows.join('|')}>` : `${flows.join('|')}`
  return `
sequenceDiagram
  participant B as AJAX Client
  participant K as Ory Kratos

  B->>K: REST GET /self-service/${components}/browser
  K-->>K: Create and store new ${flows.join(', ')} flow
  K->>B: HTTP 200 OK with flow as application/json payload
  B-->>B: Render form using HTML input elements
  B-->>B: User fills out forms, clicks e.g. ${interactions}
  B->>K: REST POST to e.g. /self-service/${components}?flow=...>
  K-->>K: Validates and processes payload
  alt Form payload is valid
    K->>B: ${success}
  else Form payload invalid
    K-->>K: Update and store flow (e.g. add form validation errors)
    K->>B: Respond with e.g. HTTP 400 Bad Request and updated flow as payload
    B-->>B: Render form and validation errors using HTML input elements
    B-->>K: Repeat flow with input data, submit, validate, ...
  end
`
}

const SelfServiceSpaFlow = (props) => <Mermaid chart={chart(props)} />

export default SelfServiceSpaFlow

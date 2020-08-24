import React from 'react'
import Mermaid from "./Mermaid";

const chart = ({
  flows = ['login', 'registration', 'settings', '...'],
  success = 'Perform flow-specific action (e.g. create user, set session cookie, ...)'
}) => `
sequenceDiagram
  participant B as API Client
  participant K as ORY Kratos

  B->>K: REST GET /self-service/<${flows.join('|')}>/api
  K-->>K: Create and store new ${flows.join(', ')} flow
  K->>B: HTTP 200 OK with flow as application/json payload
  B-->>B: Render form using e.g. Native iOS UI Elements
  B-->>B: User fills out forms, clicks e.g. "Log in", "Sign Up", "Update Email", ...
  B->>K: REST POST to e.g. /self-service/login/methods/password
  K-->>K: Validates and processes payload
  alt Form payload is valid valid
    K->>B: Perform flow-specific action (e.g. create user, return session cookie, ...)
  else Login data invalid
    K-->>K: Update and store flow (e.g. add form validation errors)
    K->>B: Respond with e.g. HTTP 400 Bad Request and updated flow as payload
    B-->>B: Render form and validation errors using e.g. Native iOS UI Elements
    B-->>K: Repeat flow with input data, submit, validate, ...
  end
`

const SelfServiceAPIFlow = (props) => <Mermaid chart={chart(props)}/>

export default SelfServiceAPIFlow

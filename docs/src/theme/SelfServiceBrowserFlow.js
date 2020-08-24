import React from 'react'
import Mermaid from "./Mermaid";

const chart = ({
  flows = ['login', 'registration', 'settings', '...'],
  interactions = ['"Log in"', '"Sign Up"', '"Update Email"', '...'],
  success = 'Perform flow-specific action (e.g. create user, set session cookie, ...)'
}) => `
sequenceDiagram

  participant B as Browser
  participant K as ORY Kratos
  participant A as Flow UI

  B->>K: Follow link to /self-service/<${flows.join('|')}>/browser
  K-->>K: Create and store new ${flows.join(', ')} flow
  K->>B: HTTP 302 Found <selfservice.flows.<${flows.join('|')}>.ui_url>?flow=<flow-id>

  B->>A: Opens <selfservice.flows.<${flows.join('|')}>.ui_url>?flow=<flow-id>
  A-->>K: Fetches data to render forms using /selfservice/<${flows.join('|')}>/flows?id=<flow-id>
  B-->>A: Fills out forms, clicks e.g. ${interactions.join(', ')}
  B->>K: Submits form
  K-->>K: Validates and processes form payloads

  alt Form payload is valid valid
    K->>B: ${success}
  else Login data invalid
    K-->>K: Update and store flow (e.g. add form validation errors)
    K-->>B: HTTP 302 Found <selfservice.flows.<${flows.join('|')}>.ui_url>?flow=<flow-id>
    B->>A: Opens <selfservice.flows.<${flows.join('|')}>?flow=<flow-id>
    A-->>K: Fetches data to render form fields and errors
    B->>K: Repeat flow with input data, submit, validate, ...
  end
`

const SelfServiceBrowserFlow = (props) => <Mermaid chart={chart(props)}/>

export default SelfServiceBrowserFlow

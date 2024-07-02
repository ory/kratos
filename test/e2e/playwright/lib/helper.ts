// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { Message } from "mailhog"
import {
  UiContainer,
  UiNodeAttributes,
  UiNodeInputAttributes,
} from "@ory/kratos-client"
import { expect } from "../fixtures"
import { LoginFlowStyle, OryKratosConfiguration } from "../../shared/config"

export const codeRegex = /(\d{6})/

/**
 * Extracts the recovery or verification code from a mail
 *
 * @param mail the mail to extract the code from
 * @returns the code or null if no code was found
 */
export function extractCode(mail: Message) {
  const result = codeRegex.exec(mail.html || mail.text)
  if (result != null && result.length > 0) {
    return result[0]
  }
  return null
}

export function findCsrfToken(ui: UiContainer) {
  const csrf = ui.nodes
    .filter((node) => isUiNodeInputAttributes(node.attributes))
    // Since we filter all non-input attributes, the following as is ok:
    .map(
      (node): UiNodeInputAttributes => node.attributes as UiNodeInputAttributes,
    )
    .find(({ name }) => name === "csrf_token")?.value
  expect(csrf).toBeDefined()
  return csrf
}

export function isUiNodeInputAttributes(
  attrs: UiNodeAttributes,
): attrs is UiNodeInputAttributes & {
  node_type: "input"
} {
  return attrs.node_type === "input"
}

export const toConfig = ({
  style = "identifier_first",
  mitigateEnumeration = false,
  selfservice,
}: {
  style?: LoginFlowStyle
  mitigateEnumeration?: boolean
  selfservice?: Partial<OryKratosConfiguration["selfservice"]>
}) => ({
  selfservice: {
    default_browser_return_url: "http://localhost:4455/welcome",
    ...selfservice,
    flows: {
      login: {
        ...selfservice?.flows?.login,
        style,
      },
      ...selfservice?.flows,
    },
  },
  security: {
    account_enumeration: {
      mitigate: mitigateEnumeration,
    },
  },
})

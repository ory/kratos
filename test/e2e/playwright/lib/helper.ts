// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import { Message } from "mailhog"

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

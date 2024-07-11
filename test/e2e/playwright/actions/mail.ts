// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import mailhog from "mailhog"
import retry from "async-retry"

const mh = mailhog({
  basePath: "http://localhost:8025/api",
})

type searchProps = {
  query: string
  kind: "to" | "from" | "containing"
  /**
   *
   * @param message an email message
   * @returns decide whether to include the message in the result
   */
  filter?: (message: mailhog.Message) => boolean
}

export function search({ query, kind, filter }: searchProps) {
  return retry(
    async () => {
      const res = await mh.search(query, kind)
      if (res.total === 0) {
        throw new Error("no emails found")
      }
      const result = filter ? res.items.filter(filter) : res.items
      if (result.length === 0) {
        throw new Error("no emails found")
      }
      return result
    },
    {
      retries: 3,
    },
  )
}

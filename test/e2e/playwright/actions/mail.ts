// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import mailhog from "mailhog"
import retry from "async-retry"

const mh = mailhog({
  basePath: "http://localhost:8025/api",
})

export function search(...props: Parameters<typeof mh["search"]>) {
  return retry(
    async () => {
      const res = await mh.search(...props)
      if (res.total === 0) {
        throw new Error("no emails found")
      }
      return res.items
    },
    {
      retries: 3,
    },
  )
}

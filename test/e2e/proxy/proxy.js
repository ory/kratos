// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

import request from "request"
import { resolve } from "path"
import urljoin from "url-join"
import express from "express"
import fs from "fs"

const app = express()

const proxy =
  (base, prefix = null) =>
  (req, res, next) => {
    let url = urljoin(base, req.url)

    // we need to tell Krato we are behind a reverse proxy
    req.headers["x-forwarded-host"] = req.get("host")

    if (prefix) {
      url = urljoin(base, prefix, req.url)
    }
    req
      .pipe(request(url, { followRedirect: false }).on("error", next))
      .pipe(res)
  }

app.use(
  "/self-service/",
  proxy(process.env.KRATOS_PUBLIC_URL, "/self-service/"),
)
app.use("/schemas/", proxy(process.env.KRATOS_PUBLIC_URL, "/schemas/"))
app.use("/.well-known/", proxy(process.env.KRATOS_PUBLIC_URL, "/.well-known/"))

app.use("/", (req, res, next) => {
  const pc = JSON.parse(fs.readFileSync(resolve("../proxy.json")).toString())
  switch (pc) {
    case "react":
      proxy(process.env.KRATOS_UI_REACT_URL)(req, res, next)
      return
    case "react-native":
      proxy(process.env.KRATOS_UI_REACT_NATIVE_URL)(req, res, next)
      return
  }
  proxy(process.env.KRATOS_UI_URL)(req, res, next)
})

const port = parseInt(process.env.PORT) || 4455

let listener = () => {
  console.log(`Listening on http://0.0.0.0:${port}`)
}

app.listen(port, "0.0.0.0", listener)

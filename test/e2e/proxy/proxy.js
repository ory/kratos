const request = require('request')
const urljoin = require('url-join')
const express = require('express')

const app = express()

const proxy = (base, prefix = null) => (req, res, next) => {
  let url = urljoin(base, req.url)
  if (prefix) {
    url = urljoin(base, prefix, req.url)
  }
  req
    .pipe(request(url, {followRedirect: false}).on('error', next))
    .pipe(res)
}

app.use('/self-service/', proxy(process.env.KRATOS_PUBLIC_URL,'/self-service/'))
app.use('/schemas/', proxy(process.env.KRATOS_PUBLIC_URL,'/schemas/'))

app.use('/', proxy(process.env.KRATOS_UI_URL))

const port = parseInt(process.env.PORT) || 4455

let listener = () => {
  console.log(`Listening on http://0.0.0.0:${port}`)
}

app.listen(port, "0.0.0.0", listener)

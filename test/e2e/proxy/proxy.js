const request = require('request')
const urljoin = require('url-join')
const express = require('express')
const fs = require('fs')

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

app.use('/self-service/', proxy(process.env.KRATOS_PUBLIC_URL, '/self-service/'))
app.use('/schemas/', proxy(process.env.KRATOS_PUBLIC_URL, '/schemas/'))
app.use('/.well-known/', proxy(process.env.KRATOS_PUBLIC_URL, '/.well-known/'))

app.use('/', (req, res, next) => {
  const pc = fs.readFileSync(require.resolve('../proxy.json'))
  console.log(pc,JSON.parse(pc.toString()) )
  if (JSON.parse(pc.toString()) === 'react') {
    proxy(process.env.KRATOS_UI_REACT_URL)(req, res, next)
  } else {
    proxy(process.env.KRATOS_UI_URL)(req, res, next)
  }
})

const port = parseInt(process.env.PORT) || 4455

let listener = () => {
  console.log(`Listening on http://0.0.0.0:${port}`)
}

app.listen(port, "0.0.0.0", listener)

---
title: REST API
id: api
---

Welcome to the ORY Kratos HTTP API documentation!

> You are viewing REST API documentation. This documentation is auto-generated
> from a swagger specification which itself is generated from annotations in the
> source code of the project. It is possible that this documentation includes
> bugs and that code samples are incomplete or wrong.
>
> If you find issues in the respective documentation, please do not edit the
> Markdown files directly (as they are generated) but raise an issue on the
> project's GitHub presence instead. This documentation will improve over time
> with your help! If you have ideas how to improve this part of the
> documentation, feel free to share them in a
> [GitHub issue](https://github.com/ory/docs/issues/new) any time.

<a id="ory-kratos-health"></a>

## health

<a id="opIdisInstanceAlive"></a>

### Check alive status

```
GET /health/alive HTTP/1.1
Accept: application/json

```

This endpoint returns a 200 status code when the HTTP server is up running. This
status does currently not include checks whether the database connection is
working.

If the service supports TLS Edge Termination, this endpoint does not require the
`X-Forwarded-Proto` header to be set.

Be aware that if you are running multiple nodes of this service, the health
status will never refer to the cluster state, only to a single instance.

#### Responses

<a id="check-alive-status-responses"></a>

##### Overview

| Status | Meaning                                                                    | Description  | Schema                              |
| ------ | -------------------------------------------------------------------------- | ------------ | ----------------------------------- |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)                    | healthStatus | [healthStatus](#schemahealthstatus) |
| 500    | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError | [genericError](#schemagenericerror) |

##### Examples

###### 200 response

```json
{
  "status": "string"
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-isInstanceAlive">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-isInstanceAlive-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceAlive-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceAlive-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceAlive-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceAlive-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceAlive-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-isInstanceAlive-shell">

```shell
curl -X GET /health/alive \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceAlive-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/health/alive", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceAlive-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/health/alive', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceAlive-java">

```java
// This sample needs improvement.
URL obj = new URL("/health/alive");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceAlive-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/health/alive',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceAlive-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/health/alive',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdisInstanceReady"></a>

### Check readiness status

```
GET /health/ready HTTP/1.1
Accept: application/json

```

This endpoint returns a 200 status code when the HTTP server is up running and
the environment dependencies (e.g. the database) are responsive as well.

If the service supports TLS Edge Termination, this endpoint does not require the
`X-Forwarded-Proto` header to be set.

Be aware that if you are running multiple nodes of this service, the health
status will never refer to the cluster state, only to a single instance.

#### Responses

<a id="check-readiness-status-responses"></a>

##### Overview

| Status | Meaning                                                                  | Description          | Schema                                              |
| ------ | ------------------------------------------------------------------------ | -------------------- | --------------------------------------------------- |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)                  | healthStatus         | [healthStatus](#schemahealthstatus)                 |
| 503    | [Service Unavailable](https://tools.ietf.org/html/rfc7231#section-6.6.4) | healthNotReadyStatus | [healthNotReadyStatus](#schemahealthnotreadystatus) |

##### Examples

###### 200 response

```json
{
  "status": "string"
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-isInstanceReady">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-isInstanceReady-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceReady-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceReady-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceReady-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceReady-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceReady-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-isInstanceReady-shell">

```shell
curl -X GET /health/ready \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceReady-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/health/ready", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceReady-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/health/ready', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceReady-java">

```java
// This sample needs improvement.
URL obj = new URL("/health/ready");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceReady-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/health/ready',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceReady-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/health/ready',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="ory-kratos-administrative-endpoints"></a>

## Administrative Endpoints

<a id="opIdlistIdentities"></a>

### List all identities in the system

```
GET /identities HTTP/1.1
Accept: application/json

```

This endpoint returns a login request's context with, for example, error details
and other information.

Learn how identities work in
[ORY Kratos' User And Identity Model Documentation](https://www.ory.sh/docs/next/kratos/concepts/identity-user-model).

#### Responses

<a id="list-all-identities-in-the-system-responses"></a>

##### Overview

| Status | Meaning                                                                    | Description           | Schema                              |
| ------ | -------------------------------------------------------------------------- | --------------------- | ----------------------------------- |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)                    | A list of identities. | Inline                              |
| 500    | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError          | [genericError](#schemagenericerror) |

<a id="list-all-identities-in-the-system-responseschema"></a>

##### Response Schema

Status Code **200**

| Name                | Type                                                  | Required | Restrictions | Description                                                                                                    |
| ------------------- | ----------------------------------------------------- | -------- | ------------ | -------------------------------------------------------------------------------------------------------------- |
| _anonymous_         | [[Identity](#schemaidentity)]                         | false    | none         | none                                                                                                           |
| » addresses         | [[VerifiableAddress](#schemaverifiableaddress)]       | false    | none         | none                                                                                                           |
| »» expires_at       | string(date-time)                                     | true     | none         | none                                                                                                           |
| »» id               | [UUID](#schemauuid)(uuid4)                            | true     | none         | none                                                                                                           |
| »» value            | string                                                | true     | none         | none                                                                                                           |
| »» verified         | boolean                                               | true     | none         | none                                                                                                           |
| »» verified_at      | string(date-time)                                     | true     | none         | none                                                                                                           |
| »» via              | [VerifiableAddressType](#schemaverifiableaddresstype) | true     | none         | none                                                                                                           |
| » id                | [UUID](#schemauuid)(uuid4)                            | true     | none         | none                                                                                                           |
| » traits            | [Traits](#schematraits)                               | true     | none         | none                                                                                                           |
| » traits_schema_id  | string                                                | true     | none         | TraitsSchemaID is the ID of the JSON Schema to be used for validating the identity's traits.                   |
| » traits_schema_url | string                                                | false    | none         | TraitsSchemaURL is the URL of the endpoint where the identity's traits schema can be fetched from. format: url |

##### Examples

###### 200 response

```json
[
  {
    "addresses": [
      {
        "expires_at": "2020-04-12T08:41:41Z",
        "id": "string",
        "value": "string",
        "verified": true,
        "verified_at": "2020-04-12T08:41:41Z",
        "via": "string"
      }
    ],
    "id": "string",
    "traits": {},
    "traits_schema_id": "string",
    "traits_schema_url": "string"
  }
]
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-listIdentities">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-listIdentities-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-listIdentities-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-listIdentities-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-listIdentities-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-listIdentities-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-listIdentities-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-listIdentities-shell">

```shell
curl -X GET /identities \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-listIdentities-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/identities", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-listIdentities-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/identities', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-listIdentities-java">

```java
// This sample needs improvement.
URL obj = new URL("/identities");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-listIdentities-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/identities',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-listIdentities-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/identities',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdcreateIdentity"></a>

### Create an identity

```
POST /identities HTTP/1.1
Content-Type: application/json
Accept: application/json

```

This endpoint creates an identity. It is NOT possible to set an identity's
credentials (password, ...) using this method! A way to achieve that will be
introduced in the future.

Learn how identities work in
[ORY Kratos' User And Identity Model Documentation](https://www.ory.sh/docs/next/kratos/concepts/identity-user-model).

#### Request body

```json
{
  "addresses": [
    {
      "expires_at": "2020-04-12T08:41:41Z",
      "id": "string",
      "value": "string",
      "verified": true,
      "verified_at": "2020-04-12T08:41:41Z",
      "via": "string"
    }
  ],
  "id": "string",
  "traits": {},
  "traits_schema_id": "string",
  "traits_schema_url": "string"
}
```

<a id="create-an-identity-parameters"></a>

##### Parameters

| Parameter | In   | Type                        | Required | Description |
| --------- | ---- | --------------------------- | -------- | ----------- |
| body      | body | [Identity](#schemaidentity) | true     | none        |

#### Responses

<a id="create-an-identity-responses"></a>

##### Overview

| Status | Meaning                                                                    | Description        | Schema                              |
| ------ | -------------------------------------------------------------------------- | ------------------ | ----------------------------------- |
| 201    | [Created](https://tools.ietf.org/html/rfc7231#section-6.3.2)               | A single identity. | [Identity](#schemaidentity)         |
| 400    | [Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)           | genericError       | [genericError](#schemagenericerror) |
| 500    | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError       | [genericError](#schemagenericerror) |

##### Examples

###### 201 response

```json
{
  "addresses": [
    {
      "expires_at": "2020-04-12T08:41:41Z",
      "id": "string",
      "value": "string",
      "verified": true,
      "verified_at": "2020-04-12T08:41:41Z",
      "via": "string"
    }
  ],
  "id": "string",
  "traits": {},
  "traits_schema_id": "string",
  "traits_schema_url": "string"
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-createIdentity">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-createIdentity-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-createIdentity-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-createIdentity-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-createIdentity-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-createIdentity-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-createIdentity-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-createIdentity-shell">

```shell
curl -X POST /identities \
  -H 'Content-Type: application/json' \  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-createIdentity-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("POST", "/identities", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-createIdentity-node">

```nodejs
const fetch = require('node-fetch');
const input = '{
  "addresses": [
    {
      "expires_at": "2020-04-12T08:41:41Z",
      "id": "string",
      "value": "string",
      "verified": true,
      "verified_at": "2020-04-12T08:41:41Z",
      "via": "string"
    }
  ],
  "id": "string",
  "traits": {},
  "traits_schema_id": "string",
  "traits_schema_url": "string"
}';
const headers = {
  'Content-Type': 'application/json',  'Accept': 'application/json'
}

fetch('/identities', {
  method: 'POST',
  body: input,
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-createIdentity-java">

```java
// This sample needs improvement.
URL obj = new URL("/identities");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("POST");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-createIdentity-python">

```python
import requests

headers = {
  'Content-Type': 'application/json',
  'Accept': 'application/json'
}

r = requests.post(
  '/identities',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-createIdentity-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json',
  'Accept' => 'application/json'
}

result = RestClient.post '/identities',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdgetIdentity"></a>

### Get an identity

```
GET /identities/{id} HTTP/1.1
Accept: application/json

```

Learn how identities work in
[ORY Kratos' User And Identity Model Documentation](https://www.ory.sh/docs/next/kratos/concepts/identity-user-model).

<a id="get-an-identity-parameters"></a>

##### Parameters

| Parameter | In   | Type   | Required | Description                                          |
| --------- | ---- | ------ | -------- | ---------------------------------------------------- |
| id        | path | string | true     | ID must be set to the ID of identity you want to get |

#### Responses

<a id="get-an-identity-responses"></a>

##### Overview

| Status | Meaning                                                                    | Description        | Schema                              |
| ------ | -------------------------------------------------------------------------- | ------------------ | ----------------------------------- |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)                    | A single identity. | [Identity](#schemaidentity)         |
| 400    | [Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)           | genericError       | [genericError](#schemagenericerror) |
| 500    | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError       | [genericError](#schemagenericerror) |

##### Examples

###### 200 response

```json
{
  "addresses": [
    {
      "expires_at": "2020-04-12T08:41:41Z",
      "id": "string",
      "value": "string",
      "verified": true,
      "verified_at": "2020-04-12T08:41:41Z",
      "via": "string"
    }
  ],
  "id": "string",
  "traits": {},
  "traits_schema_id": "string",
  "traits_schema_url": "string"
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-getIdentity">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-getIdentity-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getIdentity-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getIdentity-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getIdentity-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getIdentity-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getIdentity-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-getIdentity-shell">

```shell
curl -X GET /identities/{id} \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getIdentity-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/identities/{id}", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getIdentity-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/identities/{id}', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getIdentity-java">

```java
// This sample needs improvement.
URL obj = new URL("/identities/{id}");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getIdentity-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/identities/{id}',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getIdentity-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/identities/{id}',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdupdateIdentity"></a>

### Update an identity

```
PUT /identities/{id} HTTP/1.1
Content-Type: application/json
Accept: application/json

```

This endpoint updates an identity. It is NOT possible to set an identity's
credentials (password, ...) using this method! A way to achieve that will be
introduced in the future.

The full identity payload (except credentials) is expected. This endpoint does
not support patching.

Learn how identities work in
[ORY Kratos' User And Identity Model Documentation](https://www.ory.sh/docs/next/kratos/concepts/identity-user-model).

#### Request body

```json
{
  "addresses": [
    {
      "expires_at": "2020-04-12T08:41:41Z",
      "id": "string",
      "value": "string",
      "verified": true,
      "verified_at": "2020-04-12T08:41:41Z",
      "via": "string"
    }
  ],
  "id": "string",
  "traits": {},
  "traits_schema_id": "string",
  "traits_schema_url": "string"
}
```

<a id="update-an-identity-parameters"></a>

##### Parameters

| Parameter | In   | Type                        | Required | Description                                             |
| --------- | ---- | --------------------------- | -------- | ------------------------------------------------------- |
| id        | path | string                      | true     | ID must be set to the ID of identity you want to update |
| body      | body | [Identity](#schemaidentity) | true     | none                                                    |

#### Responses

<a id="update-an-identity-responses"></a>

##### Overview

| Status | Meaning                                                                    | Description        | Schema                              |
| ------ | -------------------------------------------------------------------------- | ------------------ | ----------------------------------- |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)                    | A single identity. | [Identity](#schemaidentity)         |
| 400    | [Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)           | genericError       | [genericError](#schemagenericerror) |
| 404    | [Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)             | genericError       | [genericError](#schemagenericerror) |
| 500    | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError       | [genericError](#schemagenericerror) |

##### Examples

###### 200 response

```json
{
  "addresses": [
    {
      "expires_at": "2020-04-12T08:41:41Z",
      "id": "string",
      "value": "string",
      "verified": true,
      "verified_at": "2020-04-12T08:41:41Z",
      "via": "string"
    }
  ],
  "id": "string",
  "traits": {},
  "traits_schema_id": "string",
  "traits_schema_url": "string"
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-updateIdentity">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-updateIdentity-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-updateIdentity-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-updateIdentity-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-updateIdentity-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-updateIdentity-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-updateIdentity-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-updateIdentity-shell">

```shell
curl -X PUT /identities/{id} \
  -H 'Content-Type: application/json' \  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-updateIdentity-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("PUT", "/identities/{id}", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-updateIdentity-node">

```nodejs
const fetch = require('node-fetch');
const input = '{
  "addresses": [
    {
      "expires_at": "2020-04-12T08:41:41Z",
      "id": "string",
      "value": "string",
      "verified": true,
      "verified_at": "2020-04-12T08:41:41Z",
      "via": "string"
    }
  ],
  "id": "string",
  "traits": {},
  "traits_schema_id": "string",
  "traits_schema_url": "string"
}';
const headers = {
  'Content-Type': 'application/json',  'Accept': 'application/json'
}

fetch('/identities/{id}', {
  method: 'PUT',
  body: input,
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-updateIdentity-java">

```java
// This sample needs improvement.
URL obj = new URL("/identities/{id}");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("PUT");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-updateIdentity-python">

```python
import requests

headers = {
  'Content-Type': 'application/json',
  'Accept': 'application/json'
}

r = requests.put(
  '/identities/{id}',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-updateIdentity-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json',
  'Accept' => 'application/json'
}

result = RestClient.put '/identities/{id}',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIddeleteIdentity"></a>

### Delete an identity

```
DELETE /identities/{id} HTTP/1.1
Accept: application/json

```

This endpoint deletes an identity. This can not be undone.

Learn how identities work in
[ORY Kratos' User And Identity Model Documentation](https://www.ory.sh/docs/next/kratos/concepts/identity-user-model).

<a id="delete-an-identity-parameters"></a>

##### Parameters

| Parameter | In   | Type   | Required | Description              |
| --------- | ---- | ------ | -------- | ------------------------ |
| id        | path | string | true     | ID is the identity's ID. |

#### Responses

<a id="delete-an-identity-responses"></a>

##### Overview

| Status         | Meaning                                                                    | Description                                                                                                    | Schema                              |
| -------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | ----------------------------------- |
| 204            | [No Content](https://tools.ietf.org/html/rfc7231#section-6.3.5)            | Empty responses are sent when, for example, resources are deleted. The HTTP status code for empty responses is |
| typically 201. | None                                                                       |
| 404            | [Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)             | genericError                                                                                                   | [genericError](#schemagenericerror) |
| 500            | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError                                                                                                   | [genericError](#schemagenericerror) |

##### Examples

###### 404 response

```json
{
  "error": {
    "code": 404,
    "debug": "The database adapter was unable to find the element",
    "details": {
      "property1": {},
      "property2": {}
    },
    "message": "string",
    "reason": "string",
    "request": "string",
    "status": "string"
  }
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-deleteIdentity">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-deleteIdentity-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-deleteIdentity-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-deleteIdentity-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-deleteIdentity-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-deleteIdentity-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-deleteIdentity-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-deleteIdentity-shell">

```shell
curl -X DELETE /identities/{id} \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-deleteIdentity-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("DELETE", "/identities/{id}", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-deleteIdentity-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/identities/{id}', {
  method: 'DELETE',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-deleteIdentity-java">

```java
// This sample needs improvement.
URL obj = new URL("/identities/{id}");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("DELETE");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-deleteIdentity-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.delete(
  '/identities/{id}',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-deleteIdentity-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.delete '/identities/{id}',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="ory-kratos-common"></a>

## common

<a id="opIdgetSchema"></a>

### getSchema

```
GET /schemas/{id} HTTP/1.1
Accept: application/json

```

Get a traits schema definition

<a id="getschema-parameters"></a>

##### Parameters

| Parameter | In   | Type   | Required | Description                                        |
| --------- | ---- | ------ | -------- | -------------------------------------------------- |
| id        | path | string | true     | ID must be set to the ID of schema you want to get |

#### Responses

<a id="getschema-responses"></a>

##### Overview

| Status | Meaning                                                                    | Description                    | Schema                              |
| ------ | -------------------------------------------------------------------------- | ------------------------------ | ----------------------------------- |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)                    | The raw identity traits schema | Inline                              |
| 404    | [Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)             | genericError                   | [genericError](#schemagenericerror) |
| 500    | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError                   | [genericError](#schemagenericerror) |

<a id="getschema-responseschema"></a>

##### Response Schema

##### Examples

###### 200 response

```json
{}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-getSchema">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-getSchema-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSchema-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSchema-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSchema-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSchema-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSchema-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-getSchema-shell">

```shell
curl -X GET /schemas/{id} \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSchema-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/schemas/{id}", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSchema-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/schemas/{id}', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSchema-java">

```java
// This sample needs improvement.
URL obj = new URL("/schemas/{id}");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSchema-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/schemas/{id}',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSchema-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/schemas/{id}',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdgetSelfServiceBrowserLoginRequest"></a>

### Get the request context of browser-based login user flows

```
GET /self-service/browser/flows/requests/login?request=string HTTP/1.1
Accept: application/json

```

This endpoint returns a login request's context with, for example, error details
and other information.

When accessing this endpoint through ORY Kratos' Public API, ensure that cookies
are set as they are required for CSRF to work. To prevent token scanning
attacks, the public endpoint does not return 404 status codes to prevent
scanning attacks.

More information can be found at
[ORY Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).

<a id="get-the-request-context-of-browser-based-login-user-flows-parameters"></a>

##### Parameters

| Parameter | In    | Type   | Required | Description                     |
| --------- | ----- | ------ | -------- | ------------------------------- |
| request   | query | string | true     | Request is the Login Request ID |

##### Detailed descriptions

**request**: Request is the Login Request ID

The value for this parameter comes from `request` URL Query parameter sent to
your application (e.g. `/login?request=abcde`).

#### Responses

<a id="get-the-request-context-of-browser-based-login-user-flows-responses"></a>

##### Overview

| Status | Meaning                                                                    | Description  | Schema                              |
| ------ | -------------------------------------------------------------------------- | ------------ | ----------------------------------- |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)                    | loginRequest | [loginRequest](#schemaloginrequest) |
| 403    | [Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)             | genericError | [genericError](#schemagenericerror) |
| 404    | [Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)             | genericError | [genericError](#schemagenericerror) |
| 410    | [Gone](https://tools.ietf.org/html/rfc7231#section-6.5.9)                  | genericError | [genericError](#schemagenericerror) |
| 500    | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError | [genericError](#schemagenericerror) |

##### Examples

###### 200 response

```json
{
  "active": "string",
  "expires_at": "2020-04-12T08:41:41Z",
  "forced": true,
  "id": "string",
  "issued_at": "2020-04-12T08:41:41Z",
  "methods": {
    "property1": {
      "config": {
        "action": "string",
        "errors": [
          {
            "message": "string"
          }
        ],
        "fields": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ],
        "method": "string",
        "providers": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ]
      },
      "method": "string"
    },
    "property2": {
      "config": {
        "action": "string",
        "errors": [
          {
            "message": "string"
          }
        ],
        "fields": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ],
        "method": "string",
        "providers": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ]
      },
      "method": "string"
    }
  },
  "request_url": "string"
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-getSelfServiceBrowserLoginRequest">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-getSelfServiceBrowserLoginRequest-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceBrowserLoginRequest-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceBrowserLoginRequest-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceBrowserLoginRequest-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceBrowserLoginRequest-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceBrowserLoginRequest-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-getSelfServiceBrowserLoginRequest-shell">

```shell
curl -X GET /self-service/browser/flows/requests/login?request=string \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceBrowserLoginRequest-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/self-service/browser/flows/requests/login", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceBrowserLoginRequest-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/self-service/browser/flows/requests/login?request=string', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceBrowserLoginRequest-java">

```java
// This sample needs improvement.
URL obj = new URL("/self-service/browser/flows/requests/login?request=string");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceBrowserLoginRequest-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/self-service/browser/flows/requests/login',
  params={
    'request': 'string'},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceBrowserLoginRequest-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/self-service/browser/flows/requests/login',
  params: {
    'request' => 'string'}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdgetSelfServiceBrowserRegistrationRequest"></a>

### Get the request context of browser-based registration user flows

```
GET /self-service/browser/flows/requests/registration?request=string HTTP/1.1
Accept: application/json

```

This endpoint returns a registration request's context with, for example, error
details and other information.

When accessing this endpoint through ORY Kratos' Public API, ensure that cookies
are set as they are required for CSRF to work. To prevent token scanning
attacks, the public endpoint does not return 404 status codes to prevent
scanning attacks.

More information can be found at
[ORY Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).

<a id="get-the-request-context-of-browser-based-registration-user-flows-parameters"></a>

##### Parameters

| Parameter | In    | Type   | Required | Description                            |
| --------- | ----- | ------ | -------- | -------------------------------------- |
| request   | query | string | true     | Request is the Registration Request ID |

##### Detailed descriptions

**request**: Request is the Registration Request ID

The value for this parameter comes from `request` URL Query parameter sent to
your application (e.g. `/registration?request=abcde`).

#### Responses

<a id="get-the-request-context-of-browser-based-registration-user-flows-responses"></a>

##### Overview

| Status | Meaning                                                                    | Description         | Schema                                            |
| ------ | -------------------------------------------------------------------------- | ------------------- | ------------------------------------------------- |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)                    | registrationRequest | [registrationRequest](#schemaregistrationrequest) |
| 403    | [Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)             | genericError        | [genericError](#schemagenericerror)               |
| 404    | [Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)             | genericError        | [genericError](#schemagenericerror)               |
| 410    | [Gone](https://tools.ietf.org/html/rfc7231#section-6.5.9)                  | genericError        | [genericError](#schemagenericerror)               |
| 500    | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError        | [genericError](#schemagenericerror)               |

##### Examples

###### 200 response

```json
{
  "active": "string",
  "expires_at": "2020-04-12T08:41:41Z",
  "id": "string",
  "issued_at": "2020-04-12T08:41:41Z",
  "methods": {
    "property1": {
      "config": {
        "action": "string",
        "errors": [
          {
            "message": "string"
          }
        ],
        "fields": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ],
        "method": "string",
        "providers": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ]
      },
      "method": "string"
    },
    "property2": {
      "config": {
        "action": "string",
        "errors": [
          {
            "message": "string"
          }
        ],
        "fields": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ],
        "method": "string",
        "providers": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ]
      },
      "method": "string"
    }
  },
  "request_url": "string"
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-getSelfServiceBrowserRegistrationRequest">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-getSelfServiceBrowserRegistrationRequest-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceBrowserRegistrationRequest-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceBrowserRegistrationRequest-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceBrowserRegistrationRequest-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceBrowserRegistrationRequest-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceBrowserRegistrationRequest-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-getSelfServiceBrowserRegistrationRequest-shell">

```shell
curl -X GET /self-service/browser/flows/requests/registration?request=string \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceBrowserRegistrationRequest-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/self-service/browser/flows/requests/registration", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceBrowserRegistrationRequest-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/self-service/browser/flows/requests/registration?request=string', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceBrowserRegistrationRequest-java">

```java
// This sample needs improvement.
URL obj = new URL("/self-service/browser/flows/requests/registration?request=string");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceBrowserRegistrationRequest-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/self-service/browser/flows/requests/registration',
  params={
    'request': 'string'},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceBrowserRegistrationRequest-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/self-service/browser/flows/requests/registration',
  params: {
    'request' => 'string'}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdgetSelfServiceBrowserSettingsRequest"></a>

### Get the request context of browser-based settings flows

```
GET /self-service/browser/flows/requests/settings?request=string HTTP/1.1
Accept: application/json

```

When accessing this endpoint through ORY Kratos' Public API, ensure that cookies
are set as they are required for checking the auth session. To prevent scanning
attacks, the public endpoint does not return 404 status codes but instead 403
or 500.

More information can be found at
[ORY Kratos User Settings & Profile Management Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-settings-profile-management).

<a id="get-the-request-context-of-browser-based-settings-flows-parameters"></a>

##### Parameters

| Parameter | In    | Type   | Required | Description                     |
| --------- | ----- | ------ | -------- | ------------------------------- |
| request   | query | string | true     | Request is the Login Request ID |

##### Detailed descriptions

**request**: Request is the Login Request ID

The value for this parameter comes from `request` URL Query parameter sent to
your application (e.g. `/login?request=abcde`).

#### Responses

<a id="get-the-request-context-of-browser-based-settings-flows-responses"></a>

##### Overview

| Status | Meaning                                                                    | Description     | Schema                                    |
| ------ | -------------------------------------------------------------------------- | --------------- | ----------------------------------------- |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)                    | settingsRequest | [settingsRequest](#schemasettingsrequest) |
| 403    | [Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)             | genericError    | [genericError](#schemagenericerror)       |
| 404    | [Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)             | genericError    | [genericError](#schemagenericerror)       |
| 410    | [Gone](https://tools.ietf.org/html/rfc7231#section-6.5.9)                  | genericError    | [genericError](#schemagenericerror)       |
| 500    | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError    | [genericError](#schemagenericerror)       |

##### Examples

###### 200 response

```json
{
  "active": "string",
  "expires_at": "2020-04-12T08:41:41Z",
  "id": "string",
  "identity": {
    "addresses": [
      {
        "expires_at": "2020-04-12T08:41:41Z",
        "id": "string",
        "value": "string",
        "verified": true,
        "verified_at": "2020-04-12T08:41:41Z",
        "via": "string"
      }
    ],
    "id": "string",
    "traits": {},
    "traits_schema_id": "string",
    "traits_schema_url": "string"
  },
  "issued_at": "2020-04-12T08:41:41Z",
  "methods": {
    "property1": {
      "config": {
        "action": "string",
        "errors": [
          {
            "message": "string"
          }
        ],
        "fields": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ],
        "method": "string"
      },
      "method": "string"
    },
    "property2": {
      "config": {
        "action": "string",
        "errors": [
          {
            "message": "string"
          }
        ],
        "fields": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ],
        "method": "string"
      },
      "method": "string"
    }
  },
  "request_url": "string",
  "update_successful": true
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-getSelfServiceBrowserSettingsRequest">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-getSelfServiceBrowserSettingsRequest-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceBrowserSettingsRequest-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceBrowserSettingsRequest-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceBrowserSettingsRequest-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceBrowserSettingsRequest-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceBrowserSettingsRequest-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-getSelfServiceBrowserSettingsRequest-shell">

```shell
curl -X GET /self-service/browser/flows/requests/settings?request=string \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceBrowserSettingsRequest-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/self-service/browser/flows/requests/settings", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceBrowserSettingsRequest-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/self-service/browser/flows/requests/settings?request=string', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceBrowserSettingsRequest-java">

```java
// This sample needs improvement.
URL obj = new URL("/self-service/browser/flows/requests/settings?request=string");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceBrowserSettingsRequest-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/self-service/browser/flows/requests/settings',
  params={
    'request': 'string'},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceBrowserSettingsRequest-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/self-service/browser/flows/requests/settings',
  params: {
    'request' => 'string'}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdgetSelfServiceVerificationRequest"></a>

### Get the request context of browser-based verification flows

```
GET /self-service/browser/flows/requests/verification?request=string HTTP/1.1
Accept: application/json

```

When accessing this endpoint through ORY Kratos' Public API, ensure that cookies
are set as they are required for checking the auth session. To prevent scanning
attacks, the public endpoint does not return 404 status codes but instead 403
or 500.

More information can be found at
[ORY Kratos Email and Phone Verification Documentation](https://www.ory.sh/docs/kratos/selfservice/flows/verify-email-account-activation).

<a id="get-the-request-context-of-browser-based-verification-flows-parameters"></a>

##### Parameters

| Parameter | In    | Type   | Required | Description               |
| --------- | ----- | ------ | -------- | ------------------------- |
| request   | query | string | true     | Request is the Request ID |

##### Detailed descriptions

**request**: Request is the Request ID

The value for this parameter comes from `request` URL Query parameter sent to
your application (e.g. `/verify?request=abcde`).

#### Responses

<a id="get-the-request-context-of-browser-based-verification-flows-responses"></a>

##### Overview

| Status | Meaning                                                                    | Description         | Schema                                            |
| ------ | -------------------------------------------------------------------------- | ------------------- | ------------------------------------------------- |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)                    | verificationRequest | [verificationRequest](#schemaverificationrequest) |
| 403    | [Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)             | genericError        | [genericError](#schemagenericerror)               |
| 404    | [Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)             | genericError        | [genericError](#schemagenericerror)               |
| 500    | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError        | [genericError](#schemagenericerror)               |

##### Examples

###### 200 response

```json
{
  "expires_at": "2020-04-12T08:41:41Z",
  "form": {
    "action": "string",
    "errors": [
      {
        "message": "string"
      }
    ],
    "fields": [
      {
        "disabled": true,
        "errors": [
          {
            "message": "string"
          }
        ],
        "name": "string",
        "pattern": "string",
        "required": true,
        "type": "string",
        "value": {}
      }
    ],
    "method": "string"
  },
  "id": "string",
  "issued_at": "2020-04-12T08:41:41Z",
  "request_url": "string",
  "success": true,
  "via": "string"
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-getSelfServiceVerificationRequest">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-getSelfServiceVerificationRequest-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceVerificationRequest-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceVerificationRequest-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceVerificationRequest-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceVerificationRequest-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceVerificationRequest-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-getSelfServiceVerificationRequest-shell">

```shell
curl -X GET /self-service/browser/flows/requests/verification?request=string \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceVerificationRequest-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/self-service/browser/flows/requests/verification", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceVerificationRequest-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/self-service/browser/flows/requests/verification?request=string', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceVerificationRequest-java">

```java
// This sample needs improvement.
URL obj = new URL("/self-service/browser/flows/requests/verification?request=string");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceVerificationRequest-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/self-service/browser/flows/requests/verification',
  params={
    'request': 'string'},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceVerificationRequest-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/self-service/browser/flows/requests/verification',
  params: {
    'request' => 'string'}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdgetSelfServiceError"></a>

### Get user-facing self-service errors

```
GET /self-service/errors HTTP/1.1
Accept: application/json

```

This endpoint returns the error associated with a user-facing self service
errors.

When accessing this endpoint through ORY Kratos' Public API, ensure that cookies
are set as they are required for CSRF to work. To prevent token scanning
attacks, the public endpoint does not return 404 status codes to prevent
scanning attacks.

More information can be found at
[ORY Kratos User User Facing Error Documentation](https://www.ory.sh/docs/kratos/self-service/flows/user-facing-errors).

<a id="get-user-facing-self-service-errors-parameters"></a>

##### Parameters

| Parameter | In    | Type   | Required | Description |
| --------- | ----- | ------ | -------- | ----------- |
| error     | query | string | false    | none        |

#### Responses

<a id="get-user-facing-self-service-errors-responses"></a>

##### Overview

| Status | Meaning                                                                    | Description                | Schema                                  |
| ------ | -------------------------------------------------------------------------- | -------------------------- | --------------------------------------- |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)                    | User-facing error response | [errorContainer](#schemaerrorcontainer) |
| 403    | [Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)             | genericError               | [genericError](#schemagenericerror)     |
| 404    | [Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)             | genericError               | [genericError](#schemagenericerror)     |
| 500    | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError               | [genericError](#schemagenericerror)     |

##### Examples

###### 200 response

```json
{
  "errors": {},
  "id": "string"
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-getSelfServiceError">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-getSelfServiceError-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceError-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceError-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceError-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceError-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getSelfServiceError-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-getSelfServiceError-shell">

```shell
curl -X GET /self-service/errors \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceError-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/self-service/errors", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceError-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/self-service/errors', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceError-java">

```java
// This sample needs improvement.
URL obj = new URL("/self-service/errors");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceError-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/self-service/errors',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getSelfServiceError-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/self-service/errors',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="ory-kratos-public-endpoints"></a>

## Public Endpoints

<a id="opIdinitializeSelfServiceBrowserLoginFlow"></a>

### Initialize browser-based login user flow

```
GET /self-service/browser/flows/login HTTP/1.1
Accept: application/json

```

This endpoint initializes a browser-based user login flow. Once initialized, the
browser will be redirected to `urls.login_ui` with the request ID set as a query
parameter. If a valid user session exists already, the browser will be
redirected to `urls.default_redirect_url`.

> This endpoint is NOT INTENDED for API clients and only works with browsers
> (Chrome, Firefox, ...).

More information can be found at
[ORY Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).

#### Responses

<a id="initialize-browser-based-login-user-flow-responses"></a>

##### Overview

| Status         | Meaning                                                                    | Description                                                                                                    | Schema                              |
| -------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | ----------------------------------- |
| 302            | [Found](https://tools.ietf.org/html/rfc7231#section-6.4.3)                 | Empty responses are sent when, for example, resources are deleted. The HTTP status code for empty responses is |
| typically 201. | None                                                                       |
| 500            | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError                                                                                                   | [genericError](#schemagenericerror) |

##### Examples

###### 500 response

```json
{
  "error": {
    "code": 404,
    "debug": "The database adapter was unable to find the element",
    "details": {
      "property1": {},
      "property2": {}
    },
    "message": "string",
    "reason": "string",
    "request": "string",
    "status": "string"
  }
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-initializeSelfServiceBrowserLoginFlow">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-initializeSelfServiceBrowserLoginFlow-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserLoginFlow-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserLoginFlow-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserLoginFlow-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserLoginFlow-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserLoginFlow-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-initializeSelfServiceBrowserLoginFlow-shell">

```shell
curl -X GET /self-service/browser/flows/login \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserLoginFlow-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/self-service/browser/flows/login", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserLoginFlow-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/self-service/browser/flows/login', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserLoginFlow-java">

```java
// This sample needs improvement.
URL obj = new URL("/self-service/browser/flows/login");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserLoginFlow-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/self-service/browser/flows/login',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserLoginFlow-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/self-service/browser/flows/login',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdinitializeSelfServiceBrowserLogoutFlow"></a>

### Initialize Browser-Based Logout User Flow

```
GET /self-service/browser/flows/logout HTTP/1.1
Accept: application/json

```

This endpoint initializes a logout flow.

> This endpoint is NOT INTENDED for API clients and only works with browsers
> (Chrome, Firefox, ...).

On successful logout, the browser will be redirected (HTTP 302 Found) to
`urls.default_return_to`.

More information can be found at
[ORY Kratos User Logout Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-logout).

#### Responses

<a id="initialize-browser-based-logout-user-flow-responses"></a>

##### Overview

| Status         | Meaning                                                                    | Description                                                                                                    | Schema                              |
| -------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | ----------------------------------- |
| 302            | [Found](https://tools.ietf.org/html/rfc7231#section-6.4.3)                 | Empty responses are sent when, for example, resources are deleted. The HTTP status code for empty responses is |
| typically 201. | None                                                                       |
| 500            | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError                                                                                                   | [genericError](#schemagenericerror) |

##### Examples

###### 500 response

```json
{
  "error": {
    "code": 404,
    "debug": "The database adapter was unable to find the element",
    "details": {
      "property1": {},
      "property2": {}
    },
    "message": "string",
    "reason": "string",
    "request": "string",
    "status": "string"
  }
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-initializeSelfServiceBrowserLogoutFlow">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-initializeSelfServiceBrowserLogoutFlow-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserLogoutFlow-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserLogoutFlow-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserLogoutFlow-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserLogoutFlow-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserLogoutFlow-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-initializeSelfServiceBrowserLogoutFlow-shell">

```shell
curl -X GET /self-service/browser/flows/logout \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserLogoutFlow-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/self-service/browser/flows/logout", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserLogoutFlow-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/self-service/browser/flows/logout', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserLogoutFlow-java">

```java
// This sample needs improvement.
URL obj = new URL("/self-service/browser/flows/logout");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserLogoutFlow-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/self-service/browser/flows/logout',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserLogoutFlow-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/self-service/browser/flows/logout',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdinitializeSelfServiceBrowserRegistrationFlow"></a>

### Initialize browser-based registration user flow

```
GET /self-service/browser/flows/registration HTTP/1.1
Accept: application/json

```

This endpoint initializes a browser-based user registration flow. Once
initialized, the browser will be redirected to `urls.registration_ui` with the
request ID set as a query parameter. If a valid user session exists already, the
browser will be redirected to `urls.default_redirect_url`.

> This endpoint is NOT INTENDED for API clients and only works with browsers
> (Chrome, Firefox, ...).

More information can be found at
[ORY Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).

#### Responses

<a id="initialize-browser-based-registration-user-flow-responses"></a>

##### Overview

| Status         | Meaning                                                                    | Description                                                                                                    | Schema                              |
| -------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | ----------------------------------- |
| 302            | [Found](https://tools.ietf.org/html/rfc7231#section-6.4.3)                 | Empty responses are sent when, for example, resources are deleted. The HTTP status code for empty responses is |
| typically 201. | None                                                                       |
| 500            | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError                                                                                                   | [genericError](#schemagenericerror) |

##### Examples

###### 500 response

```json
{
  "error": {
    "code": 404,
    "debug": "The database adapter was unable to find the element",
    "details": {
      "property1": {},
      "property2": {}
    },
    "message": "string",
    "reason": "string",
    "request": "string",
    "status": "string"
  }
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-initializeSelfServiceBrowserRegistrationFlow">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-initializeSelfServiceBrowserRegistrationFlow-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserRegistrationFlow-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserRegistrationFlow-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserRegistrationFlow-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserRegistrationFlow-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserRegistrationFlow-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-initializeSelfServiceBrowserRegistrationFlow-shell">

```shell
curl -X GET /self-service/browser/flows/registration \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserRegistrationFlow-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/self-service/browser/flows/registration", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserRegistrationFlow-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/self-service/browser/flows/registration', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserRegistrationFlow-java">

```java
// This sample needs improvement.
URL obj = new URL("/self-service/browser/flows/registration");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserRegistrationFlow-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/self-service/browser/flows/registration',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserRegistrationFlow-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/self-service/browser/flows/registration',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdinitializeSelfServiceSettingsFlow"></a>

### Initialize browser-based settings flow

```
GET /self-service/browser/flows/settings HTTP/1.1
Accept: application/json

```

This endpoint initializes a browser-based settings flow. Once initialized, the
browser will be redirected to `urls.settings_ui` with the request ID set as a
query parameter. If no valid user session exists, a login flow will be
initialized.

> This endpoint is NOT INTENDED for API clients and only works with browsers
> (Chrome, Firefox, ...).

More information can be found at
[ORY Kratos User Settings & Profile Management Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-settings-profile-management).

#### Responses

<a id="initialize-browser-based-settings-flow-responses"></a>

##### Overview

| Status         | Meaning                                                                    | Description                                                                                                    | Schema                              |
| -------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | ----------------------------------- |
| 302            | [Found](https://tools.ietf.org/html/rfc7231#section-6.4.3)                 | Empty responses are sent when, for example, resources are deleted. The HTTP status code for empty responses is |
| typically 201. | None                                                                       |
| 500            | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError                                                                                                   | [genericError](#schemagenericerror) |

##### Examples

###### 500 response

```json
{
  "error": {
    "code": 404,
    "debug": "The database adapter was unable to find the element",
    "details": {
      "property1": {},
      "property2": {}
    },
    "message": "string",
    "reason": "string",
    "request": "string",
    "status": "string"
  }
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-initializeSelfServiceSettingsFlow">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-initializeSelfServiceSettingsFlow-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceSettingsFlow-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceSettingsFlow-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceSettingsFlow-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceSettingsFlow-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceSettingsFlow-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-initializeSelfServiceSettingsFlow-shell">

```shell
curl -X GET /self-service/browser/flows/settings \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceSettingsFlow-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/self-service/browser/flows/settings", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceSettingsFlow-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/self-service/browser/flows/settings', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceSettingsFlow-java">

```java
// This sample needs improvement.
URL obj = new URL("/self-service/browser/flows/settings");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceSettingsFlow-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/self-service/browser/flows/settings',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceSettingsFlow-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/self-service/browser/flows/settings',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdcompleteSelfServiceBrowserSettingsPasswordStrategyFlow"></a>

### Complete the browser-based settings flow for the password strategy

```
POST /self-service/browser/flows/settings/strategies/password HTTP/1.1
Accept: application/json

```

This endpoint completes a browser-based settings flow. This is usually achieved
by POSTing data to this endpoint.

> This endpoint is NOT INTENDED for API clients and only works with browsers
> (Chrome, Firefox, ...) and HTML Forms.

More information can be found at
[ORY Kratos User Settings & Profile Management Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-settings-profile-management).

#### Responses

<a id="complete-the-browser-based-settings-flow-for-the-password-strategy-responses"></a>

##### Overview

| Status         | Meaning                                                                    | Description                                                                                                    | Schema                              |
| -------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | ----------------------------------- |
| 302            | [Found](https://tools.ietf.org/html/rfc7231#section-6.4.3)                 | Empty responses are sent when, for example, resources are deleted. The HTTP status code for empty responses is |
| typically 201. | None                                                                       |
| 500            | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError                                                                                                   | [genericError](#schemagenericerror) |

##### Examples

###### 500 response

```json
{
  "error": {
    "code": 404,
    "debug": "The database adapter was unable to find the element",
    "details": {
      "property1": {},
      "property2": {}
    },
    "message": "string",
    "reason": "string",
    "request": "string",
    "status": "string"
  }
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-completeSelfServiceBrowserSettingsPasswordStrategyFlow">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-completeSelfServiceBrowserSettingsPasswordStrategyFlow-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-completeSelfServiceBrowserSettingsPasswordStrategyFlow-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-completeSelfServiceBrowserSettingsPasswordStrategyFlow-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-completeSelfServiceBrowserSettingsPasswordStrategyFlow-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-completeSelfServiceBrowserSettingsPasswordStrategyFlow-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-completeSelfServiceBrowserSettingsPasswordStrategyFlow-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-completeSelfServiceBrowserSettingsPasswordStrategyFlow-shell">

```shell
curl -X POST /self-service/browser/flows/settings/strategies/password \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-completeSelfServiceBrowserSettingsPasswordStrategyFlow-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("POST", "/self-service/browser/flows/settings/strategies/password", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-completeSelfServiceBrowserSettingsPasswordStrategyFlow-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/self-service/browser/flows/settings/strategies/password', {
  method: 'POST',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-completeSelfServiceBrowserSettingsPasswordStrategyFlow-java">

```java
// This sample needs improvement.
URL obj = new URL("/self-service/browser/flows/settings/strategies/password");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("POST");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-completeSelfServiceBrowserSettingsPasswordStrategyFlow-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.post(
  '/self-service/browser/flows/settings/strategies/password',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-completeSelfServiceBrowserSettingsPasswordStrategyFlow-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.post '/self-service/browser/flows/settings/strategies/password',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdcompleteSelfServiceBrowserSettingsProfileStrategyFlow"></a>

### Complete the browser-based settings flow for profile data

```
POST /self-service/browser/flows/settings/strategies/profile?request=string HTTP/1.1
Content-Type: application/json
Accept: application/json

```

This endpoint completes a browser-based settings flow. This is usually achieved
by POSTing data to this endpoint.

If the provided profile data is valid against the Identity's Traits JSON Schema,
the data will be updated and the browser redirected to `url.settings_ui` for
further steps.

> This endpoint is NOT INTENDED for API clients and only works with browsers
> (Chrome, Firefox, ...) and HTML Forms.

More information can be found at
[ORY Kratos User Settings & Profile Management Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-settings-profile-management).

#### Request body

```json
{
  "request_id": "string",
  "traits": {}
}
```

```yaml
request_id: string
traits: {}
```

<a id="complete-the-browser-based-settings-flow-for-profile-data-parameters"></a>

##### Parameters

| Parameter | In    | Type                                                                                                                                | Required | Description                |
| --------- | ----- | ----------------------------------------------------------------------------------------------------------------------------------- | -------- | -------------------------- |
| request   | query | string                                                                                                                              | true     | Request is the request ID. |
| body      | body  | [completeSelfServiceBrowserSettingsStrategyProfileFlowPayload](#schemacompleteselfservicebrowsersettingsstrategyprofileflowpayload) | true     | none                       |

#### Responses

<a id="complete-the-browser-based-settings-flow-for-profile-data-responses"></a>

##### Overview

| Status         | Meaning                                                                    | Description                                                                                                    | Schema                              |
| -------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | ----------------------------------- |
| 302            | [Found](https://tools.ietf.org/html/rfc7231#section-6.4.3)                 | Empty responses are sent when, for example, resources are deleted. The HTTP status code for empty responses is |
| typically 201. | None                                                                       |
| 500            | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError                                                                                                   | [genericError](#schemagenericerror) |

##### Examples

###### 500 response

```json
{
  "error": {
    "code": 404,
    "debug": "The database adapter was unable to find the element",
    "details": {
      "property1": {},
      "property2": {}
    },
    "message": "string",
    "reason": "string",
    "request": "string",
    "status": "string"
  }
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-completeSelfServiceBrowserSettingsProfileStrategyFlow">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-completeSelfServiceBrowserSettingsProfileStrategyFlow-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-completeSelfServiceBrowserSettingsProfileStrategyFlow-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-completeSelfServiceBrowserSettingsProfileStrategyFlow-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-completeSelfServiceBrowserSettingsProfileStrategyFlow-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-completeSelfServiceBrowserSettingsProfileStrategyFlow-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-completeSelfServiceBrowserSettingsProfileStrategyFlow-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-completeSelfServiceBrowserSettingsProfileStrategyFlow-shell">

```shell
curl -X POST /self-service/browser/flows/settings/strategies/profile?request=string \
  -H 'Content-Type: application/json' \  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-completeSelfServiceBrowserSettingsProfileStrategyFlow-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Content-Type": []string{"application/json"},
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("POST", "/self-service/browser/flows/settings/strategies/profile", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-completeSelfServiceBrowserSettingsProfileStrategyFlow-node">

```nodejs
const fetch = require('node-fetch');
const input = '{
  "request_id": "string",
  "traits": {}
}';
const headers = {
  'Content-Type': 'application/json',  'Accept': 'application/json'
}

fetch('/self-service/browser/flows/settings/strategies/profile?request=string', {
  method: 'POST',
  body: input,
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-completeSelfServiceBrowserSettingsProfileStrategyFlow-java">

```java
// This sample needs improvement.
URL obj = new URL("/self-service/browser/flows/settings/strategies/profile?request=string");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("POST");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-completeSelfServiceBrowserSettingsProfileStrategyFlow-python">

```python
import requests

headers = {
  'Content-Type': 'application/json',
  'Accept': 'application/json'
}

r = requests.post(
  '/self-service/browser/flows/settings/strategies/profile',
  params={
    'request': 'string'},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-completeSelfServiceBrowserSettingsProfileStrategyFlow-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Content-Type' => 'application/json',
  'Accept' => 'application/json'
}

result = RestClient.post '/self-service/browser/flows/settings/strategies/profile',
  params: {
    'request' => 'string'}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdinitializeSelfServiceBrowserVerificationFlow"></a>

### Initialize browser-based verification flow

```
GET /self-service/browser/flows/verification/init/{via} HTTP/1.1
Accept: application/json

```

This endpoint initializes a browser-based verification flow. Once initialized,
the browser will be redirected to `urls.settings_ui` with the request ID set as
a query parameter. If no valid user session exists, a login flow will be
initialized.

> This endpoint is NOT INTENDED for API clients and only works with browsers
> (Chrome, Firefox, ...).

More information can be found at
[ORY Kratos Email and Phone Verification Documentation](https://www.ory.sh/docs/kratos/selfservice/flows/verify-email-account-activation).

<a id="initialize-browser-based-verification-flow-parameters"></a>

##### Parameters

| Parameter | In   | Type   | Required | Description    |
| --------- | ---- | ------ | -------- | -------------- |
| via       | path | string | true     | What to verify |

##### Detailed descriptions

**via**: What to verify

Currently only "email" is supported.

#### Responses

<a id="initialize-browser-based-verification-flow-responses"></a>

##### Overview

| Status         | Meaning                                                                    | Description                                                                                                    | Schema                              |
| -------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | ----------------------------------- |
| 302            | [Found](https://tools.ietf.org/html/rfc7231#section-6.4.3)                 | Empty responses are sent when, for example, resources are deleted. The HTTP status code for empty responses is |
| typically 201. | None                                                                       |
| 500            | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError                                                                                                   | [genericError](#schemagenericerror) |

##### Examples

###### 500 response

```json
{
  "error": {
    "code": 404,
    "debug": "The database adapter was unable to find the element",
    "details": {
      "property1": {},
      "property2": {}
    },
    "message": "string",
    "reason": "string",
    "request": "string",
    "status": "string"
  }
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-initializeSelfServiceBrowserVerificationFlow">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-initializeSelfServiceBrowserVerificationFlow-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserVerificationFlow-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserVerificationFlow-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserVerificationFlow-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserVerificationFlow-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-initializeSelfServiceBrowserVerificationFlow-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-initializeSelfServiceBrowserVerificationFlow-shell">

```shell
curl -X GET /self-service/browser/flows/verification/init/{via} \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserVerificationFlow-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/self-service/browser/flows/verification/init/{via}", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserVerificationFlow-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/self-service/browser/flows/verification/init/{via}', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserVerificationFlow-java">

```java
// This sample needs improvement.
URL obj = new URL("/self-service/browser/flows/verification/init/{via}");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserVerificationFlow-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/self-service/browser/flows/verification/init/{via}',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-initializeSelfServiceBrowserVerificationFlow-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/self-service/browser/flows/verification/init/{via}',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdselfServiceBrowserVerify"></a>

### Complete the browser-based verification flows

```
GET /self-service/browser/flows/verification/{via}/confirm/{code} HTTP/1.1
Accept: application/json

```

This endpoint completes a browser-based verification flow.

> This endpoint is NOT INTENDED for API clients and only works with browsers
> (Chrome, Firefox, ...) and HTML Forms.

More information can be found at
[ORY Kratos Email and Phone Verification Documentation](https://www.ory.sh/docs/kratos/selfservice/flows/verify-email-account-activation).

<a id="complete-the-browser-based-verification-flows-parameters"></a>

##### Parameters

| Parameter | In   | Type   | Required | Description    |
| --------- | ---- | ------ | -------- | -------------- |
| code      | path | string | true     | none           |
| via       | path | string | true     | What to verify |

##### Detailed descriptions

**via**: What to verify

Currently only "email" is supported.

#### Responses

<a id="complete-the-browser-based-verification-flows-responses"></a>

##### Overview

| Status         | Meaning                                                                    | Description                                                                                                    | Schema                              |
| -------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | ----------------------------------- |
| 302            | [Found](https://tools.ietf.org/html/rfc7231#section-6.4.3)                 | Empty responses are sent when, for example, resources are deleted. The HTTP status code for empty responses is |
| typically 201. | None                                                                       |
| 500            | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError                                                                                                   | [genericError](#schemagenericerror) |

##### Examples

###### 500 response

```json
{
  "error": {
    "code": 404,
    "debug": "The database adapter was unable to find the element",
    "details": {
      "property1": {},
      "property2": {}
    },
    "message": "string",
    "reason": "string",
    "request": "string",
    "status": "string"
  }
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-selfServiceBrowserVerify">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-selfServiceBrowserVerify-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-selfServiceBrowserVerify-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-selfServiceBrowserVerify-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-selfServiceBrowserVerify-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-selfServiceBrowserVerify-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-selfServiceBrowserVerify-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-selfServiceBrowserVerify-shell">

```shell
curl -X GET /self-service/browser/flows/verification/{via}/confirm/{code} \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-selfServiceBrowserVerify-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/self-service/browser/flows/verification/{via}/confirm/{code}", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-selfServiceBrowserVerify-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/self-service/browser/flows/verification/{via}/confirm/{code}', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-selfServiceBrowserVerify-java">

```java
// This sample needs improvement.
URL obj = new URL("/self-service/browser/flows/verification/{via}/confirm/{code}");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-selfServiceBrowserVerify-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/self-service/browser/flows/verification/{via}/confirm/{code}',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-selfServiceBrowserVerify-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/self-service/browser/flows/verification/{via}/confirm/{code}',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdwhoami"></a>

### Check who the current HTTP session belongs to

```
GET /sessions/whoami HTTP/1.1
Accept: application/json

```

Uses the HTTP Headers in the GET request to determine (e.g. by using checking
the cookies) who is authenticated. Returns a session object or 401 if the
credentials are invalid or no credentials were sent.

This endpoint is useful for reverse proxies and API Gateways.

#### Responses

<a id="check-who-the-current-http-session-belongs-to-responses"></a>

##### Overview

| Status | Meaning                                                                    | Description  | Schema                              |
| ------ | -------------------------------------------------------------------------- | ------------ | ----------------------------------- |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)                    | session      | [session](#schemasession)           |
| 403    | [Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)             | genericError | [genericError](#schemagenericerror) |
| 500    | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | genericError | [genericError](#schemagenericerror) |

##### Examples

###### 200 response

```json
{
  "authenticated_at": "2020-04-12T08:41:41Z",
  "expires_at": "2020-04-12T08:41:41Z",
  "identity": {
    "addresses": [
      {
        "expires_at": "2020-04-12T08:41:41Z",
        "id": "string",
        "value": "string",
        "verified": true,
        "verified_at": "2020-04-12T08:41:41Z",
        "via": "string"
      }
    ],
    "id": "string",
    "traits": {},
    "traits_schema_id": "string",
    "traits_schema_url": "string"
  },
  "issued_at": "2020-04-12T08:41:41Z",
  "sid": "string"
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-whoami">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-whoami-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-whoami-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-whoami-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-whoami-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-whoami-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-whoami-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-whoami-shell">

```shell
curl -X GET /sessions/whoami \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-whoami-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/sessions/whoami", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-whoami-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/sessions/whoami', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-whoami-java">

```java
// This sample needs improvement.
URL obj = new URL("/sessions/whoami");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-whoami-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/sessions/whoami',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-whoami-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/sessions/whoami',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="ory-kratos-version"></a>

## version

<a id="opIdgetVersion"></a>

### Get service version

```
GET /version HTTP/1.1
Accept: application/json

```

This endpoint returns the service version typically notated using semantic
versioning.

If the service supports TLS Edge Termination, this endpoint does not require the
`X-Forwarded-Proto` header to be set.

Be aware that if you are running multiple nodes of this service, the health
status will never refer to the cluster state, only to a single instance.

#### Responses

<a id="get-service-version-responses"></a>

##### Overview

| Status | Meaning                                                 | Description | Schema                    |
| ------ | ------------------------------------------------------- | ----------- | ------------------------- |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1) | version     | [version](#schemaversion) |

##### Examples

###### 200 response

```json
{
  "version": "string"
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-getVersion">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-getVersion-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getVersion-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getVersion-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getVersion-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getVersion-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getVersion-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-getVersion-shell">

```shell
curl -X GET /version \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getVersion-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/version", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getVersion-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/version', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getVersion-java">

```java
// This sample needs improvement.
URL obj = new URL("/version");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getVersion-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/version',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getVersion-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/version',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

## Schemas

<a id="tocScredentialstype">CredentialsType</a>

#### CredentialsType

<a id="schemacredentialstype"></a>

```json
"string"
```

_CredentialsType represents several different credential types, like password
credentials, passwordless credentials,_

#### Properties

| Name                                                                                                                | Type   | Required | Restrictions | Description |
| ------------------------------------------------------------------------------------------------------------------- | ------ | -------- | ------------ | ----------- |
| CredentialsType represents several different credential types, like password credentials, passwordless credentials, | string | false    | none         | and so on.  |

<a id="tocSerror">Error</a>

#### Error

<a id="schemaerror"></a>

```json
{
  "message": "string"
}
```

#### Properties

| Name    | Type   | Required | Restrictions | Description                              |
| ------- | ------ | -------- | ------------ | ---------------------------------------- |
| message | string | false    | none         | Code FormErrorCode `json:"id,omitempty"` |

<a id="tocSidentity">Identity</a>

#### Identity

<a id="schemaidentity"></a>

```json
{
  "addresses": [
    {
      "expires_at": "2020-04-12T08:41:41Z",
      "id": "string",
      "value": "string",
      "verified": true,
      "verified_at": "2020-04-12T08:41:41Z",
      "via": "string"
    }
  ],
  "id": "string",
  "traits": {},
  "traits_schema_id": "string",
  "traits_schema_url": "string"
}
```

#### Properties

| Name              | Type                                            | Required | Restrictions | Description                                                                                                    |
| ----------------- | ----------------------------------------------- | -------- | ------------ | -------------------------------------------------------------------------------------------------------------- |
| addresses         | [[VerifiableAddress](#schemaverifiableaddress)] | false    | none         | none                                                                                                           |
| id                | [UUID](#schemauuid)                             | true     | none         | none                                                                                                           |
| traits            | [Traits](#schematraits)                         | true     | none         | none                                                                                                           |
| traits_schema_id  | string                                          | true     | none         | TraitsSchemaID is the ID of the JSON Schema to be used for validating the identity's traits.                   |
| traits_schema_url | string                                          | false    | none         | TraitsSchemaURL is the URL of the endpoint where the identity's traits schema can be fetched from. format: url |

<a id="tocSrequestmethodconfig">RequestMethodConfig</a>

#### RequestMethodConfig

<a id="schemarequestmethodconfig"></a>

```json
{
  "action": "string",
  "errors": [
    {
      "message": "string"
    }
  ],
  "fields": [
    {
      "disabled": true,
      "errors": [
        {
          "message": "string"
        }
      ],
      "name": "string",
      "pattern": "string",
      "required": true,
      "type": "string",
      "value": {}
    }
  ],
  "method": "string"
}
```

#### Properties

| Name   | Type                            | Required | Restrictions | Description                                                                                 |
| ------ | ------------------------------- | -------- | ------------ | ------------------------------------------------------------------------------------------- |
| action | string                          | true     | none         | Action should be used as the form action URL `<form action="{{ .Action }}" method="post">`. |
| errors | [[Error](#schemaerror)]         | false    | none         | Errors contains all form errors. These will be duplicates of the individual field errors.   |
| fields | [formFields](#schemaformfields) | true     | none         | Fields contains multiple fields                                                             |
| method | string                          | true     | none         | Method is the form method (e.g. POST)                                                       |

<a id="tocStraits">Traits</a>

#### Traits

<a id="schematraits"></a>

```json
{}
```

#### Properties

_None_

<a id="tocSuuid">UUID</a>

#### UUID

<a id="schemauuid"></a>

```json
"string"
```

#### Properties

| Name        | Type          | Required | Restrictions | Description |
| ----------- | ------------- | -------- | ------------ | ----------- |
| _anonymous_ | string(uuid4) | false    | none         | none        |

<a id="tocSverifiableaddress">VerifiableAddress</a>

#### VerifiableAddress

<a id="schemaverifiableaddress"></a>

```json
{
  "expires_at": "2020-04-12T08:41:41Z",
  "id": "string",
  "value": "string",
  "verified": true,
  "verified_at": "2020-04-12T08:41:41Z",
  "via": "string"
}
```

#### Properties

| Name        | Type                                                  | Required | Restrictions | Description |
| ----------- | ----------------------------------------------------- | -------- | ------------ | ----------- |
| expires_at  | string(date-time)                                     | true     | none         | none        |
| id          | [UUID](#schemauuid)                                   | true     | none         | none        |
| value       | string                                                | true     | none         | none        |
| verified    | boolean                                               | true     | none         | none        |
| verified_at | string(date-time)                                     | true     | none         | none        |
| via         | [VerifiableAddressType](#schemaverifiableaddresstype) | true     | none         | none        |

<a id="tocSverifiableaddresstype">VerifiableAddressType</a>

#### VerifiableAddressType

<a id="schemaverifiableaddresstype"></a>

```json
"string"
```

#### Properties

| Name        | Type   | Required | Restrictions | Description |
| ----------- | ------ | -------- | ------------ | ----------- |
| _anonymous_ | string | false    | none         | none        |

<a id="tocScompleteselfservicebrowsersettingsstrategyprofileflowpayload">completeSelfServiceBrowserSettingsStrategyProfileFlowPayload</a>

#### completeSelfServiceBrowserSettingsStrategyProfileFlowPayload

<a id="schemacompleteselfservicebrowsersettingsstrategyprofileflowpayload"></a>

```json
{
  "request_id": "string",
  "traits": {}
}
```

#### Properties

| Name       | Type   | Required | Restrictions | Description                                                               |
| ---------- | ------ | -------- | ------------ | ------------------------------------------------------------------------- |
| request_id | string | false    | none         | RequestID is request ID. in: query                                        |
| traits     | object | true     | none         | Traits contains all of the identity's traits. type: string format: binary |

<a id="tocSerrorcontainer">errorContainer</a>

#### errorContainer

<a id="schemaerrorcontainer"></a>

```json
{
  "errors": {},
  "id": "string"
}
```

#### Properties

| Name   | Type                | Required | Restrictions | Description |
| ------ | ------------------- | -------- | ------------ | ----------- |
| errors | object              | false    | none         | none        |
| id     | [UUID](#schemauuid) | false    | none         | none        |

<a id="tocSform">form</a>

#### form

<a id="schemaform"></a>

```json
{
  "action": "string",
  "errors": [
    {
      "message": "string"
    }
  ],
  "fields": [
    {
      "disabled": true,
      "errors": [
        {
          "message": "string"
        }
      ],
      "name": "string",
      "pattern": "string",
      "required": true,
      "type": "string",
      "value": {}
    }
  ],
  "method": "string"
}
```

_HTMLForm represents a HTML Form. The container can work with both HTTP Form and
JSON requests_

#### Properties

| Name   | Type                            | Required | Restrictions | Description                                                                                 |
| ------ | ------------------------------- | -------- | ------------ | ------------------------------------------------------------------------------------------- |
| action | string                          | true     | none         | Action should be used as the form action URL `<form action="{{ .Action }}" method="post">`. |
| errors | [[Error](#schemaerror)]         | false    | none         | Errors contains all form errors. These will be duplicates of the individual field errors.   |
| fields | [formFields](#schemaformfields) | true     | none         | Fields contains multiple fields                                                             |
| method | string                          | true     | none         | Method is the form method (e.g. POST)                                                       |

<a id="tocSformfield">formField</a>

#### formField

<a id="schemaformfield"></a>

```json
{
  "disabled": true,
  "errors": [
    {
      "message": "string"
    }
  ],
  "name": "string",
  "pattern": "string",
  "required": true,
  "type": "string",
  "value": {}
}
```

_Field represents a HTML Form Field_

#### Properties

| Name     | Type                    | Required | Restrictions | Description                                                             |
| -------- | ----------------------- | -------- | ------------ | ----------------------------------------------------------------------- |
| disabled | boolean                 | false    | none         | Disabled is the equivalent of `<input disabled="{{.Disabled}}">`        |
| errors   | [[Error](#schemaerror)] | false    | none         | Errors contains all validation errors this particular field has caused. |
| name     | string                  | true     | none         | Name is the equivalent of `<input name="{{.Name}}">`                    |
| pattern  | string                  | false    | none         | Pattern is the equivalent of `<input pattern="{{.Pattern}}">`           |
| required | boolean                 | true     | none         | Required is the equivalent of `<input required="{{.Required}}">`        |
| type     | string                  | true     | none         | Type is the equivalent of `<input type="{{.Type}}">`                    |
| value    | object                  | false    | none         | Value is the equivalent of `<input value="{{.Value}}">`                 |

<a id="tocSformfields">formFields</a>

#### formFields

<a id="schemaformfields"></a>

```json
[
  {
    "disabled": true,
    "errors": [
      {
        "message": "string"
      }
    ],
    "name": "string",
    "pattern": "string",
    "required": true,
    "type": "string",
    "value": {}
  }
]
```

_Fields contains multiple fields_

#### Properties

| Name        | Type                            | Required | Restrictions | Description                     |
| ----------- | ------------------------------- | -------- | ------------ | ------------------------------- |
| _anonymous_ | [[formField](#schemaformfield)] | false    | none         | Fields contains multiple fields |

<a id="tocSgenericerror">genericError</a>

#### genericError

<a id="schemagenericerror"></a>

```json
{
  "error": {
    "code": 404,
    "debug": "The database adapter was unable to find the element",
    "details": {
      "property1": {},
      "property2": {}
    },
    "message": "string",
    "reason": "string",
    "request": "string",
    "status": "string"
  }
}
```

_Error response_

#### Properties

| Name  | Type                                              | Required | Restrictions | Description |
| ----- | ------------------------------------------------- | -------- | ------------ | ----------- |
| error | [genericErrorPayload](#schemagenericerrorpayload) | false    | none         | none        |

<a id="tocSgenericerrorpayload">genericErrorPayload</a>

#### genericErrorPayload

<a id="schemagenericerrorpayload"></a>

```json
{
  "code": 404,
  "debug": "The database adapter was unable to find the element",
  "details": {
    "property1": {},
    "property2": {}
  },
  "message": "string",
  "reason": "string",
  "request": "string",
  "status": "string"
}
```

#### Properties

| Name                       | Type           | Required | Restrictions | Description                                                                            |
| -------------------------- | -------------- | -------- | ------------ | -------------------------------------------------------------------------------------- |
| code                       | integer(int64) | false    | none         | Code represents the error status code (404, 403, 401, ...).                            |
| debug                      | string         | false    | none         | Debug contains debug information. This is usually not available and has to be enabled. |
| details                    | object         | false    | none         | none                                                                                   |
| » **additionalProperties** | object         | false    | none         | none                                                                                   |
| message                    | string         | false    | none         | none                                                                                   |
| reason                     | string         | false    | none         | none                                                                                   |
| request                    | string         | false    | none         | none                                                                                   |
| status                     | string         | false    | none         | none                                                                                   |

<a id="tocShealthnotreadystatus">healthNotReadyStatus</a>

#### healthNotReadyStatus

<a id="schemahealthnotreadystatus"></a>

```json
{
  "errors": {
    "property1": "string",
    "property2": "string"
  }
}
```

#### Properties

| Name                       | Type   | Required | Restrictions | Description                                                        |
| -------------------------- | ------ | -------- | ------------ | ------------------------------------------------------------------ |
| errors                     | object | false    | none         | Errors contains a list of errors that caused the not ready status. |
| » **additionalProperties** | string | false    | none         | none                                                               |

<a id="tocShealthstatus">healthStatus</a>

#### healthStatus

<a id="schemahealthstatus"></a>

```json
{
  "status": "string"
}
```

#### Properties

| Name   | Type   | Required | Restrictions | Description                  |
| ------ | ------ | -------- | ------------ | ---------------------------- |
| status | string | false    | none         | Status always contains "ok". |

<a id="tocSloginrequest">loginRequest</a>

#### loginRequest

<a id="schemaloginrequest"></a>

```json
{
  "active": "string",
  "expires_at": "2020-04-12T08:41:41Z",
  "forced": true,
  "id": "string",
  "issued_at": "2020-04-12T08:41:41Z",
  "methods": {
    "property1": {
      "config": {
        "action": "string",
        "errors": [
          {
            "message": "string"
          }
        ],
        "fields": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ],
        "method": "string",
        "providers": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ]
      },
      "method": "string"
    },
    "property2": {
      "config": {
        "action": "string",
        "errors": [
          {
            "message": "string"
          }
        ],
        "fields": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ],
        "method": "string",
        "providers": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ]
      },
      "method": "string"
    }
  },
  "request_url": "string"
}
```

#### Properties

| Name                       | Type                                            | Required | Restrictions | Description                                                                                                                                                                 |
| -------------------------- | ----------------------------------------------- | -------- | ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| active                     | [CredentialsType](#schemacredentialstype)       | false    | none         | and so on.                                                                                                                                                                  |
| expires_at                 | string(date-time)                               | true     | none         | ExpiresAt is the time (UTC) when the request expires. If the user still wishes to log in, a new request has to be initiated.                                                |
| forced                     | boolean                                         | false    | none         | Forced stores whether this login request should enforce reauthentication.                                                                                                   |
| id                         | [UUID](#schemauuid)                             | true     | none         | none                                                                                                                                                                        |
| issued_at                  | string(date-time)                               | true     | none         | IssuedAt is the time (UTC) when the request occurred.                                                                                                                       |
| methods                    | object                                          | true     | none         | Methods contains context for all enabled login methods. If a login request has been processed, but for example the password is incorrect, this will contain error messages. |
| » **additionalProperties** | [loginRequestMethod](#schemaloginrequestmethod) | false    | none         | none                                                                                                                                                                        |
| request_url                | string                                          | true     | none         | RequestURL is the initial URL that was requested from ORY Kratos. It can be used to forward information contained in the URL's path or query for example.                   |

<a id="tocSloginrequestmethod">loginRequestMethod</a>

#### loginRequestMethod

<a id="schemaloginrequestmethod"></a>

```json
{
  "config": {
    "action": "string",
    "errors": [
      {
        "message": "string"
      }
    ],
    "fields": [
      {
        "disabled": true,
        "errors": [
          {
            "message": "string"
          }
        ],
        "name": "string",
        "pattern": "string",
        "required": true,
        "type": "string",
        "value": {}
      }
    ],
    "method": "string",
    "providers": [
      {
        "disabled": true,
        "errors": [
          {
            "message": "string"
          }
        ],
        "name": "string",
        "pattern": "string",
        "required": true,
        "type": "string",
        "value": {}
      }
    ]
  },
  "method": "string"
}
```

#### Properties

| Name   | Type                                                        | Required | Restrictions | Description |
| ------ | ----------------------------------------------------------- | -------- | ------------ | ----------- |
| config | [loginRequestMethodConfig](#schemaloginrequestmethodconfig) | true     | none         | none        |
| method | [CredentialsType](#schemacredentialstype)                   | true     | none         | and so on.  |

<a id="tocSloginrequestmethodconfig">loginRequestMethodConfig</a>

#### loginRequestMethodConfig

<a id="schemaloginrequestmethodconfig"></a>

```json
{
  "action": "string",
  "errors": [
    {
      "message": "string"
    }
  ],
  "fields": [
    {
      "disabled": true,
      "errors": [
        {
          "message": "string"
        }
      ],
      "name": "string",
      "pattern": "string",
      "required": true,
      "type": "string",
      "value": {}
    }
  ],
  "method": "string",
  "providers": [
    {
      "disabled": true,
      "errors": [
        {
          "message": "string"
        }
      ],
      "name": "string",
      "pattern": "string",
      "required": true,
      "type": "string",
      "value": {}
    }
  ]
}
```

#### Properties

| Name      | Type                            | Required | Restrictions | Description                                                                                 |
| --------- | ------------------------------- | -------- | ------------ | ------------------------------------------------------------------------------------------- |
| action    | string                          | true     | none         | Action should be used as the form action URL `<form action="{{ .Action }}" method="post">`. |
| errors    | [[Error](#schemaerror)]         | false    | none         | Errors contains all form errors. These will be duplicates of the individual field errors.   |
| fields    | [formFields](#schemaformfields) | true     | none         | Fields contains multiple fields                                                             |
| method    | string                          | true     | none         | Method is the form method (e.g. POST)                                                       |
| providers | [[formField](#schemaformfield)] | false    | none         | Providers is set for the "oidc" request method.                                             |

<a id="tocSregistrationrequest">registrationRequest</a>

#### registrationRequest

<a id="schemaregistrationrequest"></a>

```json
{
  "active": "string",
  "expires_at": "2020-04-12T08:41:41Z",
  "id": "string",
  "issued_at": "2020-04-12T08:41:41Z",
  "methods": {
    "property1": {
      "config": {
        "action": "string",
        "errors": [
          {
            "message": "string"
          }
        ],
        "fields": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ],
        "method": "string",
        "providers": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ]
      },
      "method": "string"
    },
    "property2": {
      "config": {
        "action": "string",
        "errors": [
          {
            "message": "string"
          }
        ],
        "fields": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ],
        "method": "string",
        "providers": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ]
      },
      "method": "string"
    }
  },
  "request_url": "string"
}
```

#### Properties

| Name                       | Type                                                          | Required | Restrictions | Description                                                                                                                                                                               |
| -------------------------- | ------------------------------------------------------------- | -------- | ------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| active                     | [CredentialsType](#schemacredentialstype)                     | true     | none         | and so on.                                                                                                                                                                                |
| expires_at                 | string(date-time)                                             | true     | none         | ExpiresAt is the time (UTC) when the request expires. If the user still wishes to log in, a new request has to be initiated.                                                              |
| id                         | [UUID](#schemauuid)                                           | true     | none         | none                                                                                                                                                                                      |
| issued_at                  | string(date-time)                                             | true     | none         | IssuedAt is the time (UTC) when the request occurred.                                                                                                                                     |
| methods                    | object                                                        | true     | none         | Methods contains context for all enabled registration methods. If a registration request has been processed, but for example the password is incorrect, this will contain error messages. |
| » **additionalProperties** | [registrationRequestMethod](#schemaregistrationrequestmethod) | false    | none         | none                                                                                                                                                                                      |
| request_url                | string                                                        | true     | none         | RequestURL is the initial URL that was requested from ORY Kratos. It can be used to forward information contained in the URL's path or query for example.                                 |

<a id="tocSregistrationrequestmethod">registrationRequestMethod</a>

#### registrationRequestMethod

<a id="schemaregistrationrequestmethod"></a>

```json
{
  "config": {
    "action": "string",
    "errors": [
      {
        "message": "string"
      }
    ],
    "fields": [
      {
        "disabled": true,
        "errors": [
          {
            "message": "string"
          }
        ],
        "name": "string",
        "pattern": "string",
        "required": true,
        "type": "string",
        "value": {}
      }
    ],
    "method": "string",
    "providers": [
      {
        "disabled": true,
        "errors": [
          {
            "message": "string"
          }
        ],
        "name": "string",
        "pattern": "string",
        "required": true,
        "type": "string",
        "value": {}
      }
    ]
  },
  "method": "string"
}
```

#### Properties

| Name   | Type                                                                      | Required | Restrictions | Description |
| ------ | ------------------------------------------------------------------------- | -------- | ------------ | ----------- |
| config | [registrationRequestMethodConfig](#schemaregistrationrequestmethodconfig) | false    | none         | none        |
| method | [CredentialsType](#schemacredentialstype)                                 | false    | none         | and so on.  |

<a id="tocSregistrationrequestmethodconfig">registrationRequestMethodConfig</a>

#### registrationRequestMethodConfig

<a id="schemaregistrationrequestmethodconfig"></a>

```json
{
  "action": "string",
  "errors": [
    {
      "message": "string"
    }
  ],
  "fields": [
    {
      "disabled": true,
      "errors": [
        {
          "message": "string"
        }
      ],
      "name": "string",
      "pattern": "string",
      "required": true,
      "type": "string",
      "value": {}
    }
  ],
  "method": "string",
  "providers": [
    {
      "disabled": true,
      "errors": [
        {
          "message": "string"
        }
      ],
      "name": "string",
      "pattern": "string",
      "required": true,
      "type": "string",
      "value": {}
    }
  ]
}
```

#### Properties

| Name      | Type                            | Required | Restrictions | Description                                                                                 |
| --------- | ------------------------------- | -------- | ------------ | ------------------------------------------------------------------------------------------- |
| action    | string                          | true     | none         | Action should be used as the form action URL `<form action="{{ .Action }}" method="post">`. |
| errors    | [[Error](#schemaerror)]         | false    | none         | Errors contains all form errors. These will be duplicates of the individual field errors.   |
| fields    | [formFields](#schemaformfields) | true     | none         | Fields contains multiple fields                                                             |
| method    | string                          | true     | none         | Method is the form method (e.g. POST)                                                       |
| providers | [[formField](#schemaformfield)] | false    | none         | Providers is set for the "oidc" request method.                                             |

<a id="tocSsession">session</a>

#### session

<a id="schemasession"></a>

```json
{
  "authenticated_at": "2020-04-12T08:41:41Z",
  "expires_at": "2020-04-12T08:41:41Z",
  "identity": {
    "addresses": [
      {
        "expires_at": "2020-04-12T08:41:41Z",
        "id": "string",
        "value": "string",
        "verified": true,
        "verified_at": "2020-04-12T08:41:41Z",
        "via": "string"
      }
    ],
    "id": "string",
    "traits": {},
    "traits_schema_id": "string",
    "traits_schema_url": "string"
  },
  "issued_at": "2020-04-12T08:41:41Z",
  "sid": "string"
}
```

#### Properties

| Name             | Type                        | Required | Restrictions | Description |
| ---------------- | --------------------------- | -------- | ------------ | ----------- |
| authenticated_at | string(date-time)           | true     | none         | none        |
| expires_at       | string(date-time)           | true     | none         | none        |
| identity         | [Identity](#schemaidentity) | true     | none         | none        |
| issued_at        | string(date-time)           | true     | none         | none        |
| sid              | [UUID](#schemauuid)         | true     | none         | none        |

<a id="tocSsettingsrequest">settingsRequest</a>

#### settingsRequest

<a id="schemasettingsrequest"></a>

```json
{
  "active": "string",
  "expires_at": "2020-04-12T08:41:41Z",
  "id": "string",
  "identity": {
    "addresses": [
      {
        "expires_at": "2020-04-12T08:41:41Z",
        "id": "string",
        "value": "string",
        "verified": true,
        "verified_at": "2020-04-12T08:41:41Z",
        "via": "string"
      }
    ],
    "id": "string",
    "traits": {},
    "traits_schema_id": "string",
    "traits_schema_url": "string"
  },
  "issued_at": "2020-04-12T08:41:41Z",
  "methods": {
    "property1": {
      "config": {
        "action": "string",
        "errors": [
          {
            "message": "string"
          }
        ],
        "fields": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ],
        "method": "string"
      },
      "method": "string"
    },
    "property2": {
      "config": {
        "action": "string",
        "errors": [
          {
            "message": "string"
          }
        ],
        "fields": [
          {
            "disabled": true,
            "errors": [
              {
                "message": "string"
              }
            ],
            "name": "string",
            "pattern": "string",
            "required": true,
            "type": "string",
            "value": {}
          }
        ],
        "method": "string"
      },
      "method": "string"
    }
  },
  "request_url": "string",
  "update_successful": true
}
```

_Request presents a settings request_

#### Properties

| Name                       | Type                                                  | Required | Restrictions | Description                                                                                                                                                                                                                                                                                                |
| -------------------------- | ----------------------------------------------------- | -------- | ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| active                     | string                                                | false    | none         | Active, if set, contains the registration method that is being used. It is initially not set.                                                                                                                                                                                                              |
| expires_at                 | string(date-time)                                     | true     | none         | ExpiresAt is the time (UTC) when the request expires. If the user still wishes to update the setting, a new request has to be initiated.                                                                                                                                                                   |
| id                         | [UUID](#schemauuid)                                   | true     | none         | none                                                                                                                                                                                                                                                                                                       |
| identity                   | [Identity](#schemaidentity)                           | true     | none         | none                                                                                                                                                                                                                                                                                                       |
| issued_at                  | string(date-time)                                     | true     | none         | IssuedAt is the time (UTC) when the request occurred.                                                                                                                                                                                                                                                      |
| methods                    | object                                                | true     | none         | Methods contains context for all enabled registration methods. If a registration request has been processed, but for example the password is incorrect, this will contain error messages.                                                                                                                  |
| » **additionalProperties** | [settingsRequestMethod](#schemasettingsrequestmethod) | false    | none         | none                                                                                                                                                                                                                                                                                                       |
| request_url                | string                                                | true     | none         | RequestURL is the initial URL that was requested from ORY Kratos. It can be used to forward information contained in the URL's path or query for example.                                                                                                                                                  |
| update_successful          | boolean                                               | true     | none         | UpdateSuccessful, if true, indicates that the settings request has been updated successfully with the provided data. Done will stay true when repeatedly checking. If set to true, done will revert back to false only when a request with invalid (e.g. "please use a valid phone number") data was sent. |

<a id="tocSsettingsrequestmethod">settingsRequestMethod</a>

#### settingsRequestMethod

<a id="schemasettingsrequestmethod"></a>

```json
{
  "config": {
    "action": "string",
    "errors": [
      {
        "message": "string"
      }
    ],
    "fields": [
      {
        "disabled": true,
        "errors": [
          {
            "message": "string"
          }
        ],
        "name": "string",
        "pattern": "string",
        "required": true,
        "type": "string",
        "value": {}
      }
    ],
    "method": "string"
  },
  "method": "string"
}
```

#### Properties

| Name   | Type                                              | Required | Restrictions | Description                                   |
| ------ | ------------------------------------------------- | -------- | ------------ | --------------------------------------------- |
| config | [RequestMethodConfig](#schemarequestmethodconfig) | false    | none         | none                                          |
| method | string                                            | false    | none         | Method contains the request credentials type. |

<a id="tocSverificationrequest">verificationRequest</a>

#### verificationRequest

<a id="schemaverificationrequest"></a>

```json
{
  "expires_at": "2020-04-12T08:41:41Z",
  "form": {
    "action": "string",
    "errors": [
      {
        "message": "string"
      }
    ],
    "fields": [
      {
        "disabled": true,
        "errors": [
          {
            "message": "string"
          }
        ],
        "name": "string",
        "pattern": "string",
        "required": true,
        "type": "string",
        "value": {}
      }
    ],
    "method": "string"
  },
  "id": "string",
  "issued_at": "2020-04-12T08:41:41Z",
  "request_url": "string",
  "success": true,
  "via": "string"
}
```

_Request presents a verification request_

#### Properties

| Name        | Type                                                  | Required | Restrictions | Description                                                                                                                                               |
| ----------- | ----------------------------------------------------- | -------- | ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| expires_at  | string(date-time)                                     | false    | none         | ExpiresAt is the time (UTC) when the request expires. If the user still wishes to verify the address, a new request has to be initiated.                  |
| form        | [form](#schemaform)                                   | false    | none         | HTMLForm represents a HTML Form. The container can work with both HTTP Form and JSON requests                                                             |
| id          | [UUID](#schemauuid)                                   | false    | none         | none                                                                                                                                                      |
| issued_at   | string(date-time)                                     | false    | none         | IssuedAt is the time (UTC) when the request occurred.                                                                                                     |
| request_url | string                                                | false    | none         | RequestURL is the initial URL that was requested from ORY Kratos. It can be used to forward information contained in the URL's path or query for example. |
| success     | boolean                                               | false    | none         | Success, if true, implies that the request was completed successfully.                                                                                    |
| via         | [VerifiableAddressType](#schemaverifiableaddresstype) | false    | none         | none                                                                                                                                                      |

<a id="tocSversion">version</a>

#### version

<a id="schemaversion"></a>

```json
{
  "version": "string"
}
```

#### Properties

| Name    | Type   | Required | Restrictions | Description                       |
| ------- | ------ | -------- | ------------ | --------------------------------- |
| version | string | false    | none         | Version is the service's version. |

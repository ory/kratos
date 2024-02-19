// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

;(function () {
  if (!window) {
    return
  }

  if (!window.PublicKeyCredential) {
    console.log("This browser does not support WebAuthn!")
    return
  }

  if (window.__oryWebAuthnInitialized) {
    return
  }

  function __oryWebAuthnBufferDecode(value) {
    return Uint8Array.from(
      atob(value.replaceAll("-", "+").replaceAll("_", "/")),
      function (c) {
        return c.charCodeAt(0)
      },
    )
  }

  function __oryWebAuthnBufferEncode(value) {
    return btoa(String.fromCharCode.apply(null, new Uint8Array(value)))
      .replaceAll("+", "-")
      .replaceAll("/", "_")
      .replaceAll("=", "")
  }

  function __oryWebAuthnLogin(
    opt,
    resultQuerySelector = '*[name="webauthn_login"]',
    triggerQuerySelector = '*[name="webauthn_login_trigger"]',
  ) {
    if (!window.PublicKeyCredential) {
      alert("This browser does not support WebAuthn!")
    }

    opt.publicKey.challenge = __oryWebAuthnBufferDecode(opt.publicKey.challenge)
    opt.publicKey.allowCredentials = opt.publicKey.allowCredentials.map(
      function (value) {
        return {
          ...value,
          id: __oryWebAuthnBufferDecode(value.id),
        }
      },
    )

    navigator.credentials
      .get(opt)
      .then(function (credential) {
        document.querySelector(resultQuerySelector).value = JSON.stringify({
          id: credential.id,
          rawId: __oryWebAuthnBufferEncode(credential.rawId),
          type: credential.type,
          response: {
            authenticatorData: __oryWebAuthnBufferEncode(
              credential.response.authenticatorData,
            ),
            clientDataJSON: __oryWebAuthnBufferEncode(
              credential.response.clientDataJSON,
            ),
            signature: __oryWebAuthnBufferEncode(credential.response.signature),
            userHandle: __oryWebAuthnBufferEncode(
              credential.response.userHandle,
            ),
          },
        })

        document.querySelector(triggerQuerySelector).closest("form").submit()
      })
      .catch((err) => {
        alert(err)
      })
  }

  function __oryWebAuthnRegistration(
    opt,
    resultQuerySelector = '*[name="webauthn_register"]',
    triggerQuerySelector = '*[name="webauthn_register_trigger"]',
  ) {
    if (!window.PublicKeyCredential) {
      alert("This browser does not support WebAuthn!")
    }

    opt.publicKey.user.id = __oryWebAuthnBufferDecode(opt.publicKey.user.id)
    opt.publicKey.challenge = __oryWebAuthnBufferDecode(opt.publicKey.challenge)

    if (opt.publicKey.excludeCredentials) {
      opt.publicKey.excludeCredentials = opt.publicKey.excludeCredentials.map(
        function (value) {
          return {
            ...value,
            id: __oryWebAuthnBufferDecode(value.id),
          }
        },
      )
    }

    navigator.credentials
      .create(opt)
      .then(function (credential) {
        document.querySelector(resultQuerySelector).value = JSON.stringify({
          id: credential.id,
          rawId: __oryWebAuthnBufferEncode(credential.rawId),
          type: credential.type,
          response: {
            attestationObject: __oryWebAuthnBufferEncode(
              credential.response.attestationObject,
            ),
            clientDataJSON: __oryWebAuthnBufferEncode(
              credential.response.clientDataJSON,
            ),
          },
        })

        document.querySelector(triggerQuerySelector).closest("form").submit()
      })
      .catch((err) => {
        alert(err)
      })
  }

  window.__oryPasskeyLoginAutocompleteInit = async function () {
    const dataEl = document.getElementsByName("passkey_challenge")[0]
    const resultEl = document.getElementsByName("passkey_login")[0]
    const identifierEl = document.getElementsByName("identifier")[0]

    if (!dataEl || !resultEl || !identifierEl) {
      console.debug(
        "__oryPasskeyLoginAutocompleteInit: mandatory fields not found",
      )
      return
    }

    if (
      !window.PublicKeyCredential ||
      !window.PublicKeyCredential.isConditionalMediationAvailable ||
      window.Cypress // Cypress auto-fills the autocomplete, which we don't want
    ) {
      console.log("This browser does not support WebAuthn!")
      return
    }
    const isCMA = await PublicKeyCredential.isConditionalMediationAvailable()
    if (!isCMA) {
      console.log(
        "This browser does not support WebAuthn Conditional Mediation!",
      )
      return
    }

    let opt = JSON.parse(dataEl.value)

    if (opt.publicKey.user && opt.publicKey.user.id) {
      opt.publicKey.user.id = __oryWebAuthnBufferDecode(opt.publicKey.user.id)
    }
    opt.publicKey.challenge = __oryWebAuthnBufferDecode(opt.publicKey.challenge)

    // Allow aborting through a global variable
    window.abortPasskeyConditionalUI = new AbortController()

    navigator.credentials
      .get({
        publicKey: opt.publicKey,
        mediation: "conditional",
        signal: abortPasskeyConditionalUI.signal,
      })
      .then(function (credential) {
        resultEl.value = JSON.stringify({
          id: credential.id,
          rawId: __oryWebAuthnBufferEncode(credential.rawId),
          type: credential.type,
          response: {
            authenticatorData: __oryWebAuthnBufferEncode(
              credential.response.authenticatorData,
            ),
            clientDataJSON: __oryWebAuthnBufferEncode(
              credential.response.clientDataJSON,
            ),
            signature: __oryWebAuthnBufferEncode(credential.response.signature),
            userHandle: __oryWebAuthnBufferEncode(
              credential.response.userHandle,
            ),
          },
        })

        resultEl.closest("form").submit()
      })
      .catch((err) => {
        console.log(err)
      })
  }

  window.__oryPasskeyLogin = function () {
    const dataEl = document.getElementsByName("passkey_challenge")[0]
    const resultEl = document.getElementsByName("passkey_login")[0]

    if (!dataEl || !resultEl) {
      console.debug("__oryPasskeyLogin: mandatory fields not found")
      return
    }
    if (!window.PublicKeyCredential) {
      console.log("This browser does not support WebAuthn!")
      return
    }

    let opt = JSON.parse(dataEl.value)

    if (opt.publicKey.user && opt.publicKey.user.id) {
      opt.publicKey.user.id = __oryWebAuthnBufferDecode(opt.publicKey.user.id)
    }
    opt.publicKey.challenge = __oryWebAuthnBufferDecode(opt.publicKey.challenge)
    if (opt.publicKey.allowCredentials) {
      opt.publicKey.allowCredentials = opt.publicKey.allowCredentials.map(
        function (cred) {
          return {
            ...cred,
            id: __oryWebAuthnBufferDecode(cred.id),
          }
        },
      )
    }

    window.abortPasskeyConditionalUI &&
      window.abortPasskeyConditionalUI.abort(
        "only one credentials.get allowed at a time",
      )

    navigator.credentials
      .get({
        publicKey: opt.publicKey,
      })
      .then(function (credential) {
        resultEl.value = JSON.stringify({
          id: credential.id,
          rawId: __oryWebAuthnBufferEncode(credential.rawId),
          type: credential.type,
          response: {
            authenticatorData: __oryWebAuthnBufferEncode(
              credential.response.authenticatorData,
            ),
            clientDataJSON: __oryWebAuthnBufferEncode(
              credential.response.clientDataJSON,
            ),
            signature: __oryWebAuthnBufferEncode(credential.response.signature),
            userHandle: __oryWebAuthnBufferEncode(
              credential.response.userHandle,
            ),
          },
        })

        resultEl.closest("form").submit()
      })
      .catch((err) => {
        // Calling this again will enable the autocomplete once again.
        console.error(err)
        window.abortPasskeyConditionalUI && __oryPasskeyLoginAutocompleteInit()
      })
  }

  window.__oryPasskeyRegistration = function () {
    const dataEl = document.getElementsByName("passkey_create_data")[0]
    const resultEl = document.getElementsByName("passkey_register")[0]

    if (!dataEl || !resultEl) {
      console.debug("__oryPasskeyRegistration: mandatory fields not found")
      return
    }

    const createData = JSON.parse(dataEl.value)

    // Fetch display name from field value
    const displayNameFieldName = createData.displayNameFieldName
    const displayName = dataEl
      .closest("form")
      .querySelector("[name='" + displayNameFieldName + "']").value

    let opts = createData.credentialOptions
    opts.publicKey.user.name = displayName
    opts.publicKey.user.displayName = displayName
    opts.publicKey.user.id = __oryWebAuthnBufferDecode(opts.publicKey.user.id)
    opts.publicKey.challenge = __oryWebAuthnBufferDecode(
      opts.publicKey.challenge,
    )

    if (opts.publicKey.excludeCredentials) {
      opts.publicKey.excludeCredentials = opts.publicKey.excludeCredentials.map(
        function (value) {
          return {
            ...value,
            id: __oryWebAuthnBufferDecode(value.id),
          }
        },
      )
    }

    navigator.credentials
      .create(opts)
      .then(function (credential) {
        resultEl.value = JSON.stringify({
          id: credential.id,
          rawId: __oryWebAuthnBufferEncode(credential.rawId),
          type: credential.type,
          response: {
            attestationObject: __oryWebAuthnBufferEncode(
              credential.response.attestationObject,
            ),
            clientDataJSON: __oryWebAuthnBufferEncode(
              credential.response.clientDataJSON,
            ),
          },
        })

        resultEl.closest("form").submit()
      })
      .catch((err) => {
        console.error(err)
      })
  }

  function __oryPasskeySettingsRegistration() {
    const dataEl = document.getElementsByName("passkey_create_data")[0]
    const resultEl = document.getElementsByName("passkey_settings_register")[0]

    if (!dataEl || !resultEl) {
      console.debug(
        "__oryPasskeySettingsRegistration: mandatory fields not found",
      )
      return
    }

    let opt = JSON.parse(dataEl.value)

    opt.publicKey.user.id = __oryWebAuthnBufferDecode(opt.publicKey.user.id)
    opt.publicKey.challenge = __oryWebAuthnBufferDecode(opt.publicKey.challenge)

    if (opt.publicKey.excludeCredentials) {
      opt.publicKey.excludeCredentials = opt.publicKey.excludeCredentials.map(
        function (value) {
          return {
            ...value,
            id: __oryWebAuthnBufferDecode(value.id),
          }
        },
      )
    }

    navigator.credentials
      .create(opt)
      .then(function (credential) {
        resultEl.value = JSON.stringify({
          id: credential.id,
          rawId: __oryWebAuthnBufferEncode(credential.rawId),
          type: credential.type,
          response: {
            attestationObject: __oryWebAuthnBufferEncode(
              credential.response.attestationObject,
            ),
            clientDataJSON: __oryWebAuthnBufferEncode(
              credential.response.clientDataJSON,
            ),
          },
        })

        resultEl.closest("form").submit()
      })
      .catch((err) => {
        console.error(err)
      })
  }

  window.__oryWebAuthnLogin = __oryWebAuthnLogin
  window.__oryWebAuthnRegistration = __oryWebAuthnRegistration
  window.__oryPasskeySettingsRegistration = __oryPasskeySettingsRegistration
  window.__oryWebAuthnInitialized = true
})()

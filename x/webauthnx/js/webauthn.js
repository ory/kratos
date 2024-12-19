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
    options,
    resultQuerySelector = '*[name="webauthn_login"]',
    triggerQuerySelector = '*[name="webauthn_login_trigger"]',
  ) {
    if (!window.PublicKeyCredential) {
      alert("This browser does not support WebAuthn!")
    }

    const triggerEl = document.querySelector(triggerQuerySelector)
    let opt = options
    if (!opt) {
      opt = JSON.parse(triggerEl.value)
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

        triggerEl.closest("form").submit()
      })
      .catch((err) => {
        alert(err)
      })
  }

  function __oryWebAuthnRegistration(
    options,
    resultQuerySelector = '*[name="webauthn_register"]',
    triggerQuerySelector = '*[name="webauthn_register_trigger"]',
  ) {
    if (!window.PublicKeyCredential) {
      alert("This browser does not support WebAuthn!")
    }

    const triggerEl = document.querySelector(triggerQuerySelector)
    let opt = options
    if (!opt) {
      opt = JSON.parse(triggerEl.value)
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

        triggerEl.closest("form").submit()
      })
      .catch((err) => {
        alert(err)
      })
  }

  async function __oryPasskeyLoginAutocompleteInit () {
    const dataEl = document.getElementsByName("passkey_challenge")[0]
    const resultEl = document.getElementsByName("passkey_login")[0]
    const identifierEl = document.getElementsByName("identifier")[0]

    if (!dataEl || !resultEl || !identifierEl) {
      console.error(
        "Unable to initialize WebAuthn / Passkey autocomplete because one or more required form fields are missing.",
      )
      return
    }

    if (
      !window.PublicKeyCredential ||
      !window.PublicKeyCredential.isConditionalMediationAvailable ||
      window.Cypress // Cypress auto-fills the autocomplete, which we don't want
    ) {
      console.log("This browser does not support Passkey / WebAuthn!")
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

    // If this is set we already have a request ongoing which we need to abort.
    if (window.abortPasskeyConditionalUI) {
      window.abortPasskeyConditionalUI.abort(
        "Canceling Passkey autocomplete to complete trigger-based passkey login.",
      )
      window.abortPasskeyConditionalUI = undefined
    }

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
        console.trace(err)
        console.log(err)
      })
  }

  function __oryPasskeyLogin () {
    const dataEl = document.getElementsByName("passkey_challenge")[0]
    const resultEl = document.getElementsByName("passkey_login")[0]

    if (!dataEl || !resultEl) {
      console.error(
        "Unable to initialize WebAuthn / Passkey autocomplete because one or more required form fields are missing.",
      )
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

    if (window.abortPasskeyConditionalUI) {
      window.abortPasskeyConditionalUI.abort(
        "Canceling Passkey autocomplete to complete trigger-based passkey login.",
      )
      window.abortPasskeyConditionalUI = undefined
    }

    navigator.credentials
      .get({
        publicKey: opt.publicKey,
      })
      .then(function (credential) {
        console.trace('login',credential)
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
        if (err instanceof DOMException && err.name === "SecurityError") {
          console.error(`A security exception occurred while loading Passkeys / WebAuthn. To troubleshoot, please head over to https://www.ory.sh/docs/troubleshooting/passkeys-webauthn-security-error. The original error message is: ${err.message}`)
        } else {
          console.error("[Ory/Passkey] An unknown error occurred while getting passkey credentials", err)
        }

        console.trace(err)

        // Try re-initializing autocomplete
        return __oryPasskeyLoginAutocompleteInit()
      })
  }

  function __oryPasskeyRegistration () {
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

  // Deprecated naming with underscores - kept for support with Ory Elements v0
  window.__oryWebAuthnLogin = __oryWebAuthnLogin
  window.__oryWebAuthnRegistration = __oryWebAuthnRegistration
  window.__oryPasskeySettingsRegistration = __oryPasskeySettingsRegistration
  window.__oryPasskeyLogin = __oryPasskeyLogin
  window.__oryPasskeyRegistration = __oryPasskeyRegistration
  window.__oryPasskeyLoginAutocompleteInit = __oryPasskeyLoginAutocompleteInit

  // Current naming - use with Ory Elements v1
  window.oryWebAuthnLogin = __oryWebAuthnLogin
  window.oryWebAuthnRegistration = __oryWebAuthnRegistration
  window.oryPasskeySettingsRegistration = __oryPasskeySettingsRegistration
  window.oryPasskeyLogin = __oryPasskeyLogin
  window.oryPasskeyRegistration = __oryPasskeyRegistration
  window.oryPasskeyLoginAutocompleteInit = __oryPasskeyLoginAutocompleteInit

  window.__oryWebAuthnInitialized = true
})()

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

  async function __oryPasskeyLogin() {
    const dataEl = document.getElementsByName("passkey_challenge")[0]
    const resultEl = document.getElementsByName("passkey_login")[0]
    const identifierEl = document.getElementsByName("identifier")[0]

    if (!dataEl || !resultEl || !identifierEl) {
      console.debug("__oryPasskeyLogin: mandatory fields not found")
      return
    }

    if (
      !window.PublicKeyCredential ||
      !window.PublicKeyCredential.isConditionalMediationAvailable
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

    navigator.credentials
      .get({
        publicKey: opt.publicKey,
        mediation: "conditional",
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
        identifierEl.value = credential.id

        document
          .querySelector('*[type="submit"][name="method"][value="passkey"]')
          .click()
      })
      .catch((err) => {
        alert(err)
      })
  }

  function __oryPasskeyRegistration() {
    const dataEl = document.getElementsByName("create_passkey_data")[0]
    const resultEl = document.getElementsByName("passkey_register")[0]

    if (!dataEl || !resultEl) {
      console.debug("__oryPasskeyRegistration: mandatory fields not found")
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

        document
          .querySelector('*[type="submit"][name="method"][value="passkey"]')
          .click()
      })
      .catch((err) => {
        alert(err)
      })
  }

  function __oryPasskeySettingsRegistration() {
    const dataEl = document.getElementsByName("create_passkey_data")[0]
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

        document
          .querySelector('*[type="submit"][name="method"][value="passkey"]')
          .click()
      })
      .catch((err) => console.error(err))
  }

  document.addEventListener("DOMContentLoaded", __oryPasskeyLogin)
  document.addEventListener("DOMContentLoaded", __oryPasskeyRegistration)
  document.addEventListener("DOMContentLoaded", function () {
    for (const el of document.getElementsByName("passkey_register_trigger")) {
      el.addEventListener("click", __oryPasskeySettingsRegistration)
    }
  })

  window.__oryWebAuthnLogin = __oryWebAuthnLogin
  window.__oryWebAuthnRegistration = __oryWebAuthnRegistration
  window.__oryPasskeySettingsRegistration = __oryPasskeySettingsRegistration
  window.__oryWebAuthnInitialized = true
})()

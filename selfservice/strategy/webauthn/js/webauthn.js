if (
  (window && window.__oryWebAuthnLogin && window.__oryWebAuthnRegistration) ||
  (!window && __oryWebAuthnLogin && __oryWebAuthnRegistration)
) {
  // Already registered these functions, do nothing.
} else {
  function __oryWebAuthnBufferDecode(value) {
    return Uint8Array.from(atob(value), function (c) {
      return c.charCodeAt(0)
    });
  }

  function __oryWebAuthnBufferEncode(value) {
    return btoa(String.fromCharCode.apply(null, new Uint8Array(value)))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
  }

  function __oryWebAuthnLogin(opt, resultQuerySelector = '*[name="webauthn_login"]', triggerQuerySelector = '*[name="webauthn_login_trigger"]') {
    if (!window.PublicKeyCredential) {
      alert('This browser does not support WebAuthn!');
    }

    opt.publicKey.challenge = __oryWebAuthnBufferDecode(opt.publicKey.challenge);
    opt.publicKey.allowCredentials = opt.publicKey.allowCredentials.map(function (value) {
      return {
        ...value,
        id: __oryWebAuthnBufferDecode(value.id)
      }
    });

    navigator.credentials.get(opt).then(function (credential) {
      document.querySelector(resultQuerySelector).value = JSON.stringify({
        id: credential.id,
        rawId: __oryWebAuthnBufferEncode(credential.rawId),
        type: credential.type,
        response: {
          authenticatorData: __oryWebAuthnBufferEncode(credential.response.authenticatorData),
          clientDataJSON: __oryWebAuthnBufferEncode(credential.response.clientDataJSON),
          signature: __oryWebAuthnBufferEncode(credential.response.signature),
          userHandle: __oryWebAuthnBufferEncode(credential.response.userHandle),
        },
      })

      document.querySelector(triggerQuerySelector).closest('form').submit()
    }).catch((err) => {
      alert(err)
    })
  }

  function __oryWebAuthnRegistration(opt, resultQuerySelector = '*[name="webauthn_register"]', triggerQuerySelector = '*[name="webauthn_register_trigger"]') {
    if (!window.PublicKeyCredential) {
      alert('This browser does not support WebAuthn!');
    }

    opt.publicKey.user.id = __oryWebAuthnBufferDecode(opt.publicKey.user.id);
    opt.publicKey.challenge = __oryWebAuthnBufferDecode(opt.publicKey.challenge);

    if (opt.publicKey.excludeCredentials) {
      opt.publicKey.excludeCredentials = opt.publicKey.excludeCredentials.map(function (value) {
        return {
          ...value,
          id: __oryWebAuthnBufferDecode(value.id)
        }
      })
    }

    navigator.credentials.create(opt).then(function (credential) {
      document.querySelector(resultQuerySelector).value = JSON.stringify({
        id: credential.id,
        rawId: __oryWebAuthnBufferEncode(credential.rawId),
        type: credential.type,
        response: {
          attestationObject: __oryWebAuthnBufferEncode(credential.response.attestationObject),
          clientDataJSON: __oryWebAuthnBufferEncode(credential.response.clientDataJSON),
        },
      })

      document.querySelector(triggerQuerySelector).closest('form').submit()
    }).catch((err) => {
      alert(err)
    })
  }

  if (window) {
    window.__oryWebAuthnLogin = __oryWebAuthnLogin
    window.__oryWebAuthnRegistration = __oryWebAuthnRegistration
  }
}

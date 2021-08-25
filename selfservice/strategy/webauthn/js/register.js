// noinspection JSAnnotator
if (!window.PublicKeyCredential) {
  alert('This browser does not support WebAuthn!');
} else {
  function bufferDecode(value) {
    return Uint8Array.from(atob(value), c => c.charCodeAt(0));
  }

  function bufferEncode(value) {
    return btoa(String.fromCharCode.apply(null, new Uint8Array(value)))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
  }

  const opt = injectWebAuthnOptions
  opt.publicKey.user.id = bufferDecode(opt.publicKey.user.id);
  opt.publicKey.challenge = bufferDecode(opt.publicKey.challenge);

  if (opt.publicKey.excludeCredentials) {
    opt.publicKey.excludeCredentials = opt.publicKey.excludeCredentials.map(function (value) {
      return {
        ...value,
        id: bufferDecode(value.id)
      }
    })
  }

  navigator.credentials.create(opt).then(function (credential) {
    document.querySelector('*[name="webauthn_register"]').value = JSON.stringify({
      id: credential.id,
      rawId: bufferEncode(credential.rawId),
      type: credential.type,
      response: {
        attestationObject: bufferEncode(credential.response.attestationObject),
        clientDataJSON: bufferEncode(credential.response.clientDataJSON),
      },
    })

    document.querySelector('*[name="webauthn_register_trigger"]').closest('form').submit()
  }).catch((err) => {
    alert(err)
  })
}

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
  opt.publicKey.challenge = bufferDecode(opt.publicKey.challenge);
  opt.publicKey.allowCredentials = opt.publicKey.allowCredentials.map(function (value) {
    return {
      ...value,
      id: bufferDecode(value.id)
    }
  });

  navigator.credentials.get(opt).then(function (credential) {
    document.querySelector('*[name="webauthn_login"]').value  = JSON.stringify({
      id: credential.id,
      rawId: bufferEncode(credential.rawId),
      type: credential.type,
      response: {
        authenticatorData: bufferEncode(credential.response.authenticatorData),
        clientDataJSON: bufferEncode(credential.response.clientDataJSON),
        signature: bufferEncode(credential.response.signature),
        userHandle: bufferEncode(credential.response.userHandle),
      },
    })

    document.querySelector('*[name="webauthn_login_trigger"]').closest('form').submit()
  }).catch((err) => {
    alert(err)
  })
}

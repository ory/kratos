// noinspection JSAnnotator
return (function (e) {
  if (e.value) {
    return true
  }

  if (!window.PublicKeyCredential) {
    alert('This browser does not support WebAuthn!');
    return false;
  }

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
    const val = JSON.stringify({
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
    console.log('setting value!', val)
    document.querySelector('input[name="webauthn_login"]').value = val

    console.log('Submitting!')
    e.closest('form').submit()
    console.log('Done!')
  }).catch((err) => {
    alert(err)
  })

  return false
})(this)

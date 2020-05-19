module.exports = {
  // specify additional options here, especially http(s)
  // see https://github.com/request/request#readme for specifics
  ca: [ /* strings or binaries */],
  cert: [ /* strings or binaries */],
  key: [ /* strings or binaries */],
  passphrase: 'yourpassphrase',
  auth: {
    user: 'yourusername',
    pass: 'yourpassword'
  },
  httpSignature: {
    keyId: 'keyId',
    key: 'yourkey'
  },
  strictSSL: false,
  followAllRedirects: false,
  followRedirect: false,
  headers: {
    'x-custom': 'headers'
  }
};

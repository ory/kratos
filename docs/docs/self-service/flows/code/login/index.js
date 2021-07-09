export const getFlowMethodPasswordWithErrors = {
  browser: {
    label: 'Browser UI',
    image: require('./images/browser-error.png').default,
    alt: 'User Login HTML Form with validation errors'
  },
  missing: {
    label: 'Missing Email',
    language: 'shell',
    code: require('raw-loader!./samples/password.missing.txt').default
  },
  wrong: {
    label: 'Wrong Credentials',
    language: 'shell',
    code: require('raw-loader!./samples/password.wrong.txt').default
  }
}

export const getFlowMethodOidcWithErrors = {
  missing: {
    label: 'Missing ID Token',
    language: 'shell',
    code: require('raw-loader!./samples/oidc.missing.txt').default
  }
}

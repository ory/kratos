export const getFlowMethodPasswordWithErrors = {
  browser: {
    label: 'Browser UI',
    image: require('./images/browser-error.png').default,
    alt: 'User Registration HTML Form with validation errors'
  },
  missing: {
    label: 'Missing Email',
    language: 'shell',
    code: require('raw-loader!./samples/password.missing.txt').default
  },
  wrong: {
    label: 'Password Policy Violation',
    language: 'shell',
    code: require('raw-loader!./samples/password.policy.txt').default
  }
}

export const getFlowMethodOidcWithCompletion = {
  browser: {
    label: 'Browser UI',
    image: require('./images/browser-oidc-invalid.png').default,
    alt: 'User Registration HTML Form with missing or invalid fields when performing an OpenID Connect flow'
  },
  missing: {
    label: 'Missing Website',
    language: 'shell',
    code: require('raw-loader!./samples/oidc.invalid.txt').default
  }
}

export const getFlowMethodOidcWithErrors = {
  missing: {
    label: 'Missing ID Token',
    language: 'shell',
    code: require('raw-loader!./samples/oidc.missing.txt').default
  }
}

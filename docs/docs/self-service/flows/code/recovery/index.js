export const getFlowMethodLinkWithErrors = {
  browser: {
    label: 'Browser UI',
    image: require('./images/browser-email-missing.png').default,
    alt: 'Account Recovery HTML Form with validation errors'
  },
  missing: {
    label: 'Missing Email',
    language: 'shell',
    code: require('raw-loader!./samples/link.missing.txt').default
  }
}

export const getFlowMethodLinkSuccess = {
  browser: {
    label: 'Browser UI',
    image: require('./images/browser-success.png').default,
    alt: 'Account Recovery HTML Form with success message'
  },
  missing: {
    label: 'Email Sent',
    language: 'shell',
    code: require('raw-loader!./samples/link.success.txt').default
  }
}

export const getFlowMethodLinkInvalidChallenge = {
  browser: {
    label: 'Browser UI',
    image: require('./images/browser-invalid-challenge.png').default,
    alt: 'Account Recovery HTML Form with an invalid challenge'
  },
  missing: {
    label: 'Invalid Challenge',
    language: 'shell',
    code: require('raw-loader!./samples/link.invalid-challenge.txt').default
  }
}

export const getFlowMethodLinkChallengeDone = {
  browser: {
    label: 'Browser UI',
    image: require('./images/browser-settings-success.png').default,
    alt: 'Account Recovery HTML Form with an invalid challenge'
  },
  missing: {
    label: 'Update Credentials',
    language: 'shell',
    code: require('raw-loader!./samples/settings.success.txt').default
  }
}

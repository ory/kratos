export const getFlowMethodLinkWithErrors = {
  browser: {
    label: 'Browser UI',
    image: require('./images/browser-missing.png').default,
    alt: 'Email Verification and Account Activation HTML Form with validation errors'
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
    alt: 'Email Verification and Account Activation HTML Form with success message'
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
    alt: 'Email Verification and Account Activation HTML Form with an invalid challenge'
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
    image: require('./images/browser-challenge-completed.png').default,
    alt: 'Email Verification and Account Activation HTML Form with an invalid challenge'
  },
  missing: {
    label: 'Success State',
    language: 'shell',
    code: require('raw-loader!./samples/link.passed.txt').default
  }
}

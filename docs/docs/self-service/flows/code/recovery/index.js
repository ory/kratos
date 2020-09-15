export const getFlow = {
  curl: {
    label: 'Raw HTTP',
    language: 'shell',
    code: require('raw-loader!./samples/get.curl.txt').default
  },
  js: {
    label: 'JavaScript SDK',
    language: 'js',
    code: require('raw-loader!./samples/get.js.txt').default
  },
  go: {
    label: 'Go SDK',
    language: 'go',
    code: require('raw-loader!./samples/get.go.txt').default
  }
}

export const initBrowserFlow = {
  curl: {
    label: 'Raw HTTP',
    language: 'shell',
    code: require('raw-loader!./samples/browser/init.curl.txt').default
  },
  html: {
    label: 'HTML',
    language: 'html',
    code: require('raw-loader!./samples/browser/init.html.txt').default
  },
  jsx: {
    label: 'ReactJS',
    language: 'js',
    code: require('raw-loader!./samples/browser/init.jsx.txt').default
  },
  node: {
    label: 'Angular',
    language: 'html',
    code: require('raw-loader!./samples/browser/init.js.txt').default
  },
  angular: {
    label: 'ExpressJS',
    language: 'js',
    code: require('raw-loader!./samples/browser/init.ng.html.txt').default
  }
}

export const initApiFlow = {
  curl: {
    label: 'Raw HTTP',
    language: 'shell',
    code: require('raw-loader!./samples/api/init.curl.txt').default
  },
  js: {
    label: 'Node',
    language: 'js',
    code: require('raw-loader!./samples/api/init.js.txt').default
  },
  go: {
    label: 'Go',
    language: 'go',
    code: require('raw-loader!./samples/api/init.go.txt').default
  }
}

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

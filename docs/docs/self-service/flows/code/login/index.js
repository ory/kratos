export const initBrowserFlow = {
  curl: {
    label: 'curl',
    language: 'shell',
    code: require('raw-loader!./samples/browser/init.curl.txt').default
  },
  html: {
    label: 'HTML',
    language: 'html',
    code: require('raw-loader!./samples/browser/init.html.txt').default
  },
  node: {
    label: 'NodeJS (ExpressJS, ...)',
    language: 'js',
    code: require('raw-loader!./samples/browser/init.js.txt').default
  }
}

export const initSpaFlow = {
  curl: {
    label: 'curl',
    language: 'shell',
    code: require('raw-loader!./samples/browser/init.curl.ajax.txt').default
  },
  jsx: {
    label: 'AJAX (React, Next, Angular, ...)',
    language: 'js',
    code: require('raw-loader!./samples/browser/init.ajax.txt').default
  }
}

export const getFlow = {
  curl: {
    label: 'curl',
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

export const initApiFlow = {
  curl: {
    label: 'curl',
    language: 'shell',
    code: require('raw-loader!./samples/api/init.curl.txt').default
  },
  js: {
    label: 'TypeScript',
    language: 'ts',
    code: require('raw-loader!./samples/api/init.js.txt').default
  },
  go: {
    label: 'Go',
    language: 'go',
    code: require('raw-loader!../../../../../../examples/go/selfserviceinit/login/main.go')
      .default
  }
}

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

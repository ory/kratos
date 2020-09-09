import mp4 from './images/browser-privileged-update.mp4'
import webm from './images/browser-privileged-update.webm'

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
  },
  curlUnauth: {
    label: 'Without Valid Session',
    language: 'shell',
    code: require('raw-loader!./samples/api/init-unauth.curl.txt').default
  }
}

export const getFlowMethodProfileWithErrors = {
  browser: {
    label: 'Browser UI',
    image: require('./images/browser-profile-invalid.png').default,
    alt: 'User Profile HTML Form with validation errors'
  },
  missing: {
    label: 'Not An Email',
    language: 'shell',
    code: require('raw-loader!./samples/profile.invalid.txt').default
  }
}

export const getFlowMethodPasswordWithErrors = {
  browser: {
    label: 'Browser UI',
    image: require('./images/browser-password-missing.png').default,
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

export const privilegedVideo = {mp4, webm}

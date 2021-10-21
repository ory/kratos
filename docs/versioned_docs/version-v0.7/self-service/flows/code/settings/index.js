import mp4 from './images/browser-privileged-update.mp4'
import webm from './images/browser-privileged-update.webm'

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
    label: 'Missing Password',
    language: 'shell',
    code: require('raw-loader!./samples/password.missing.txt').default
  },
  wrong: {
    label: 'Password Policy Violation',
    language: 'shell',
    code: require('raw-loader!./samples/password.policy.txt').default
  }
}

export const privilegedVideo = { mp4, webm }

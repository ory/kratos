export const apiFlow = {
  curl: {
    label: 'curl',
    language: 'shell',
    code: require('raw-loader!./samples/api/curl.txt').default
  },
  go: {
    label: 'Go',
    language: 'go',
    code: require('raw-loader!../../../../../../examples/go/submit/logout/main.go')
      .default
  }
}
export const browserFlow = {
  node: {
    label: 'NodeJS (ExpressJS, ...)',
    language: 'js',
    code: require('raw-loader!./samples/browser/node.txt').default
  }
}

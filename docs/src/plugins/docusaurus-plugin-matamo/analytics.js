/**
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

import ExecutionEnvironment from '@docusaurus/ExecutionEnvironment'

export default (function () {
  if (!ExecutionEnvironment.canUseDOM) {
    return null
  }

  let lastLocation = window.location.pathname
  return {
    onRouteUpdate({ location }) {
      if (!window._paq) {
        return
      }

      const pagePath = location
        ? location.pathname + location.search + location.hash
        : undefined

      if (lastLocation === location.pathname) {
        return
      }
      lastLocation = location.pathname

      _paq.push(['setCustomUrl', pagePath])
      _paq.push(['setDocumentTitle', document.domain + '/' + document.title])
      _paq.push(['trackPageView'])
    }
  }
})()

const { ROUTES } = require('../config/routes')
const { activeRoleHomeRoute } = require('./active-role')

const DEFAULT_ROUTE_MAP = {
  home: ROUTES.playerHome || ROUTES.home,
  map: ROUTES.map,
  message: ROUTES.message,
  comment: ROUTES.message,
  mine: ROUTES.profile,
  profile: ROUTES.profile,
  avatar: ROUTES.profile,
  search: ROUTES.gameHall,
  metaverse: ROUTES.metaverse
}

let forwardRoute = ''

// 地图城市探索与元宇宙均为二期能力。一期页面的入口都通过本导航工具
// 跳转，在这里集中拦截，避免遗漏某个旧页面入口而重新打开未交付页面。
function deferredFeatureMessage(route) {
  const path = String(route || '').replace(/^\/+/, '').split('?')[0]

  if (path === ROUTES.map || path.indexOf('pages/map/') === 0) {
    return '地图功能暂未开放'
  }

  if (path === ROUTES.metaverse || path.indexOf('pages/metaverse/') === 0) {
    return '元宇宙功能暂未开放'
  }

  return ''
}

function showDeferredFeatureMessage(message) {
  if (!message || typeof wx === 'undefined' || typeof wx.showToast !== 'function') {
    return
  }

  wx.showToast({ title: message, icon: 'none' })
}

function normalizeRoute(route) {
  return route ? `/${String(route).replace(/^\/+/, '')}` : ''
}

function decodeRoutePart(value) {
  try {
    return decodeURIComponent(String(value || '').replace(/\+/g, ' '))
  } catch (error) {
    return String(value || '')
  }
}

function routeIdentity(route) {
  const value = String(route || '').replace(/^\/+/, '').split('#')[0]
  const queryIndex = value.indexOf('?')
  const path = queryIndex >= 0 ? value.slice(0, queryIndex) : value
  const query = queryIndex >= 0 ? value.slice(queryIndex + 1) : ''

  if (!path) {
    return ''
  }

  const queryParts = query
    ? query.split('&').filter(Boolean).map((part) => {
      const separatorIndex = part.indexOf('=')
      const key = separatorIndex >= 0 ? part.slice(0, separatorIndex) : part
      const valuePart = separatorIndex >= 0 ? part.slice(separatorIndex + 1) : ''

      return [decodeRoutePart(key), decodeRoutePart(valuePart)]
    }).sort((left, right) => {
      const keyOrder = left[0].localeCompare(right[0])
      return keyOrder || left[1].localeCompare(right[1])
    })
    : []

  return JSON.stringify([path, queryParts])
}

function isSameRoute(route, currentRoute) {
  const target = routeIdentity(route)
  const current = routeIdentity(currentRoute)

  return Boolean(target && current && target === current)
}

function routeForPage(page) {
  if (!page || !page.route) {
    return ''
  }

  const options = page.options && typeof page.options === 'object' ? page.options : {}
  const query = Object.keys(options).sort().map((key) => (
    `${encodeURIComponent(key)}=${encodeURIComponent(options[key] == null ? '' : options[key])}`
  )).join('&')

  return query ? `${page.route}?${query}` : page.route
}

function getCurrentRoute() {
  if (typeof getCurrentPages !== 'function') {
    return ''
  }

  const pages = getCurrentPages()
  const current = pages[pages.length - 1]

  return routeForPage(current)
}

function findRouteInStack(route) {
  if (typeof getCurrentPages !== 'function') {
    return -1
  }

  const pages = getCurrentPages()

  for (let index = pages.length - 1; index >= 0; index -= 1) {
    const page = pages[index]

    if (page && isSameRoute(route, routeForPage(page))) {
      return index
    }
  }

  return -1
}

function routeForShellKey(key, routeMap = {}) {
  if (routeMap[key]) {
    return routeMap[key]
  }

  if (key === 'home') {
    return activeRoleHomeRoute()
  }

  return DEFAULT_ROUTE_MAP[key] || ''
}

function replaceRoute(url) {
  wx.redirectTo({
    url,
    fail: () => {
      wx.reLaunch({ url })
    }
  })
}

function pushRoute(url) {
  wx.navigateTo({
    url,
    fail: () => replaceRoute(url)
  })
}

function navigateShellBack() {
  if (typeof getCurrentPages !== 'function' || getCurrentPages().length <= 1) {
    return false
  }

  const currentRoute = getCurrentRoute()
  const url = normalizeRoute(currentRoute)

  if (url) {
    forwardRoute = url
  }

  wx.navigateBack({
    delta: 1,
    fail: () => {
      forwardRoute = ''
    }
  })

  return true
}

function navigateShellForward() {
  if (!forwardRoute) {
    return false
  }

  const route = forwardRoute
  forwardRoute = ''
  pushRoute(route)

  return true
}

function navigateShellRoute(route, options = {}) {
  if (!route) {
    return false
  }

  const deferredMessage = deferredFeatureMessage(route)
  if (deferredMessage) {
    showDeferredFeatureMessage(deferredMessage)
    return false
  }

  const currentRoute = options.currentRoute || getCurrentRoute()

  if (options.reuseExisting !== false) {
    if (isSameRoute(route, currentRoute)) {
      if (typeof options.onSameRoute === 'function') {
        options.onSameRoute(route)
      }
      return true
    }

    const stackIndex = findRouteInStack(route)

    if (stackIndex >= 0) {
      const delta = getCurrentPages().length - 1 - stackIndex

      if (delta > 0) {
        const url = normalizeRoute(route)

        wx.navigateBack({
          delta,
          fail: () => replaceRoute(url)
        })
        return true
      }

      return true
    }
  }

  const url = normalizeRoute(route)

  if (!url) {
    return false
  }

  pushRoute(url)

  return true
}

function navigateShellKey(key, options = {}) {
  if (key === 'left') {
    return navigateShellBack()
  }

  if (key === 'right') {
    return navigateShellForward()
  }

  const route = routeForShellKey(key, options.routeMap || {})

  return navigateShellRoute(route, Object.assign({}, options, {
    onSameRoute: options.onSameRoute
      ? () => options.onSameRoute(key, route)
      : null
  }))
}

module.exports = {
  navigateShellBack,
  navigateShellForward,
  navigateShellKey,
  navigateShellRoute,
  normalizeRoute,
  routeForShellKey,
  deferredFeatureMessage
}

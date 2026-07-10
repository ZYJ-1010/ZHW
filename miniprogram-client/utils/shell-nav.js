const { ROUTES } = require('../config/routes')

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

function normalizeRoute(route) {
  return route ? `/${String(route).replace(/^\/+/, '')}` : ''
}

function isSameRoute(route, currentRoute) {
  const target = String(route || '').replace(/^\/+/, '').split('?')[0]
  const current = String(currentRoute || '').replace(/^\/+/, '').split('?')[0]

  return Boolean(target && current && target === current)
}

function getCurrentRoute() {
  if (typeof getCurrentPages !== 'function') {
    return ''
  }

  const pages = getCurrentPages()
  const current = pages[pages.length - 1]

  return current && current.route ? current.route : ''
}

function findRouteInStack(route) {
  if (typeof getCurrentPages !== 'function') {
    return -1
  }

  const pages = getCurrentPages()

  for (let index = pages.length - 1; index >= 0; index -= 1) {
    const page = pages[index]

    if (page && isSameRoute(route, page.route)) {
      return index
    }
  }

  return -1
}

function routeForShellKey(key, routeMap = {}) {
  return routeMap[key] || DEFAULT_ROUTE_MAP[key] || ''
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

  const currentRoute = options.currentRoute || getCurrentRoute()

  if (isSameRoute(route, currentRoute)) {
    if (typeof options.onSameRoute === 'function') {
      options.onSameRoute(route)
    }
    return true
  }

  const url = normalizeRoute(route)

  if (!url) {
    return false
  }

  const stackIndex = findRouteInStack(route)

  if (stackIndex >= 0) {
    const delta = getCurrentPages().length - 1 - stackIndex

    if (delta > 0) {
      wx.navigateBack({
        delta,
        fail: () => replaceRoute(url)
      })
      return true
    }
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
  routeForShellKey
}

const fs = require('fs')
const path = require('path')

const root = path.resolve(__dirname, '..')
const miniRoot = path.join(root, 'miniprogram-client')
const goRoot = path.join(root, 'services', 'go-api')

function read(file) {
  return fs.existsSync(file) ? fs.readFileSync(file, 'utf8') : ''
}

function walk(dir, files = []) {
  if (!fs.existsSync(dir)) {
    return files
  }

  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name)
    if (entry.isDirectory()) {
      walk(full, files)
    } else {
      files.push(full)
    }
  }

  return files
}

function rel(file) {
  return path.relative(root, file).replace(/\\/g, '/')
}

function pagePath(page) {
  return path.join(miniRoot, page)
}

function normalizeApiPath(raw) {
  return String(raw || '')
    .replace(/\$\{[^}]+\}/g, ':id')
    .replace(/\$\{encodeURIComponent\([^}]+\)\}/g, ':id')
    .replace(/\/+/g, '/')
}

function routeToPattern(route) {
  const escaped = route
    .replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    .replace(/\\:id/g, '[^/]+')

  return new RegExp(`^${escaped}$`)
}

function loadAppPages() {
  const app = JSON.parse(read(path.join(miniRoot, 'app.json')))
  return new Set(app.pages || [])
}

function loadBackendRoutes() {
  const files = walk(path.join(goRoot, 'internal', 'appapi'))
    .concat(walk(path.join(goRoot, 'cmd', 'server')))
    .filter((file) => file.endsWith('.go'))
  const exact = []
  const prefixes = []
  const handleRe = /handle\(\s*"([A-Z]+)\s+([^"]+)"/g

  for (const file of files) {
    const source = read(file)
    let match
    while ((match = handleRe.exec(source)) !== null) {
      const item = {
        method: match[1],
        route: match[2],
        file: rel(file)
      }

      if (item.route.endsWith('/')) {
        prefixes.push(item)
      } else {
        exact.push(Object.assign(item, { pattern: routeToPattern(item.route) }))
      }
    }
  }

  return { exact, prefixes }
}

function loadSources(paths) {
  return paths.map((item) => {
    const file = item.startsWith('miniprogram-client/')
      ? path.join(root, item)
      : path.join(miniRoot, item)
    return { file: item, source: read(file) }
  })
}

function includesAny(source, tokens) {
  return tokens.some((token) => source.includes(token))
}

function backendCovers(route, backendRoutes) {
  const pathValue = normalizeApiPath(route.path)
  if (backendRoutes.exact.some((item) => item.method === route.method && item.pattern.test(pathValue))) {
    return true
  }

  return backendRoutes.prefixes.some((item) => (
    item.method === route.method && pathValue.startsWith(item.route)
  ))
}

const featureFlows = [
  {
    name: 'invite-registration-login',
    pages: ['pages/entry/index', 'pages/login/index', 'pages/login/invite/index'],
    frontend: ['services/auth.js', 'api/request.js', 'api/modules/game.js'],
    tokens: ['loginByWechat', 'verifyInvite', 'createInviteEntry'],
    apiRoutes: [
      { method: 'POST', path: '/api/app/invites/precheck' },
      { method: 'POST', path: '/api/app/invites/entries' },
      { method: 'POST', path: '/api/app/auth/wechat-login' }
    ]
  },
  {
    name: 'identity-realname',
    pages: ['pages/login/realname/index', 'pages/profile/system-management/profile-info/index'],
    frontend: ['api/modules/user.js', 'services/user.js', 'services/auth.js'],
    tokens: ['getIdentityStatus', 'verifyPhone', 'bindPhone', 'issueTokenAfterIdentity'],
    apiRoutes: [
      { method: 'GET', path: '/api/app/identity/status' },
      { method: 'POST', path: '/api/app/identity/phone/verify' },
      { method: 'POST', path: '/api/app/auth/issue-token-after-identity' }
    ]
  },
  {
    name: 'role-application',
    pages: ['pages/role/apply/index', 'pages/role/status/index', 'pages/home-other/index'],
    frontend: ['services/role.js', 'api/modules/role.js'],
    tokens: ['submitRoleApplication', 'getExpertApplyConfig', 'getGuideApplyConfig', 'getMyRoleApplications'],
    apiRoutes: [
      { method: 'POST', path: '/api/app/role-applications' },
      { method: 'GET', path: '/api/app/role-applications/my' },
      { method: 'GET', path: '/api/app/role-applications/expert/config' },
      { method: 'GET', path: '/api/app/role-applications/guide/config' }
    ]
  },
  {
    name: 'home-browse-game-detail',
    pages: ['pages/home/player/index', 'pages/game/hall/index', 'pages/game/detail/index'],
    frontend: ['services/home.js', 'services/game.js', 'api/modules/home.js', 'api/modules/game.js'],
    tokens: ['getHome', 'getGameList', 'getGameDetail', 'navigateShellRoute'],
    apiRoutes: [
      { method: 'GET', path: '/api/app/home' },
      { method: 'GET', path: '/api/app/games' },
      { method: 'GET', path: '/api/app/games/:id' }
    ]
  },
  {
    name: 'create-game-audit-config',
    pages: ['pages/game/create/index', 'pages/game/detail/index'],
    frontend: ['services/game.js', 'api/modules/game.js'],
    tokens: ['createGame', 'getCategoryConfig', 'getConditionRuleConfig', 'getProfitTemplates', 'publishGame'],
    apiRoutes: [
      { method: 'GET', path: '/api/app/games/category-config' },
      { method: 'GET', path: '/api/app/games/condition-rule-config' },
      { method: 'GET', path: '/api/app/games/profit-templates' },
      { method: 'POST', path: '/api/app/games' }
    ]
  },
  {
    name: 'join-application-and-review',
    pages: ['pages/game/apply/index', 'pages/game/applications/index', 'pages/game/player-manage/index'],
    frontend: ['services/game.js', 'api/modules/game.js'],
    tokens: ['applyGame', 'getReceivedApplications', 'reviewGameApplication', 'respondGameInvitation'],
    apiRoutes: [
      { method: 'POST', path: '/api/app/games/:id/applications' },
      { method: 'GET', path: '/api/app/game-applications/received' },
      { method: 'POST', path: '/api/app/game-applications/:id/audit' },
      { method: 'POST', path: '/api/app/game-invitations/:id/respond' }
    ]
  },
  {
    name: 'im-and-files',
    pages: ['pages/im/room/index'],
    frontend: ['services/im.js', 'services/file.js', 'api/modules/im.js', 'api/modules/file.js'],
    tokens: ['getMessages', 'sendMessage', 'uploadSingleFile', 'createUploadToken'],
    apiRoutes: [
      { method: 'GET', path: '/api/app/games/:id/chat-room' },
      { method: 'GET', path: '/api/app/games/:id/chat/messages' },
      { method: 'POST', path: '/api/app/games/:id/chat/messages' },
      { method: 'POST', path: '/api/app/files/upload-token' }
    ]
  },
  {
    name: 'delivery-service-review',
    pages: ['pages/game/delivery/index', 'pages/game/review/index', 'pages/game/review-complete/index'],
    frontend: ['services/game.js', 'services/review.js', 'api/modules/game.js', 'api/modules/review.js'],
    tokens: ['confirmService', 'getAvailableReviews', 'submitReview', 'getCompleteConfig'],
    apiRoutes: [
      { method: 'POST', path: '/api/app/games/:id/service-confirm' },
      { method: 'GET', path: '/api/app/reviews/available' },
      { method: 'POST', path: '/api/app/reviews' },
      { method: 'GET', path: '/api/app/reviews/complete-config' }
    ]
  },
  {
    name: 'credit-report-appeal',
    pages: [
      'pages/profile/credit-center/index',
      'pages/profile/system-management/report-center/index',
      'pages/profile/system-management/report-appeals/index',
      'pages/profile/system-management/credit-appeal/index'
    ],
    frontend: ['services/profile.js', 'services/report.js', 'api/modules/profile.js', 'api/modules/report.js'],
    tokens: ['getCreditCenter', 'createReport', 'submitCreditAppeal', 'withdrawAppeal'],
    apiRoutes: [
      { method: 'GET', path: '/api/app/profile/credit-center' },
      { method: 'POST', path: '/api/app/reports' },
      { method: 'POST', path: '/api/app/profile/credit-appeals' },
      { method: 'POST', path: '/api/app/reports/:id/appeal/withdraw' }
    ]
  },
  {
    name: 'message-notification-actions',
    pages: ['pages/message/index', 'pages/message/trade-warning/index', 'pages/message/system-detail/index'],
    frontend: ['services/message.js', 'api/modules/message.js'],
    tokens: ['getMessageCenter', 'handleNotificationAction', 'handleTradeWarning', 'submitSystemNotificationFeedback'],
    apiRoutes: [
      { method: 'GET', path: '/api/app/notifications' },
      { method: 'POST', path: '/api/app/notifications/:id/actions' },
      { method: 'GET', path: '/api/app/messages/trade-warning' },
      { method: 'POST', path: '/api/app/messages/system-notification/feedback' }
    ]
  },
  {
    name: 'profile-assets-orders',
    pages: [
      'pages/profile/index',
      'pages/profile/asset-center/orders/index',
      'pages/profile/asset-center/mall/index',
      'pages/profile/asset-center/points/index'
    ],
    frontend: ['services/profile.js', 'api/modules/profile.js'],
    tokens: ['getProfileHome', 'getPointsOrders', 'getPointsOrderDetail', 'cancelPointsOrder', 'exchangePointsMallGood'],
    apiRoutes: [
      { method: 'GET', path: '/api/app/profile/home' },
      { method: 'GET', path: '/api/app/redemption/orders/my' },
      { method: 'GET', path: '/api/app/redemption/orders/:id' },
      { method: 'POST', path: '/api/app/redemption/orders/:id/cancel' },
      { method: 'POST', path: '/api/app/redemption/orders' }
    ]
  },
  {
    name: 'map-basic-and-play-pages',
    pages: [
      'pages/map/index',
      'pages/map/my-city/index',
      'pages/map/city-atlas/index',
      'pages/map/real-checkin/index',
      'pages/map/blind-route/index',
      'pages/map/footprint-heatmap/index',
      'pages/map/friend-city/index'
    ],
    frontend: ['services/map.js', 'services/location.js', 'api/modules/map.js', 'api/modules/location.js'],
    tokens: ['getIndexConfig', 'getMyCity', 'getPlayPage', 'submitCheckin', 'createBlindRoute', 'getNearbyGames'],
    apiRoutes: [
      { method: 'GET', path: '/api/app/map/index-config' },
      { method: 'GET', path: '/api/app/map/my-city' },
      { method: 'GET', path: '/api/app/map/play-pages' },
      { method: 'POST', path: '/api/app/map/checkins' },
      { method: 'POST', path: '/api/app/map/blind-routes' },
      { method: 'GET', path: '/api/app/games/nearby' }
    ]
  }
]

function checkFeature(feature, appPages, backendRoutes) {
  const failures = []

  for (const page of feature.pages || []) {
    if (!appPages.has(page)) {
      failures.push(`page missing: ${page}`)
    }

    const js = pagePath(`${page}.js`)
    const wxml = pagePath(`${page}.wxml`)
    if (!fs.existsSync(js)) {
      failures.push(`page js missing: ${rel(js)}`)
    }
    if (!fs.existsSync(wxml)) {
      failures.push(`page wxml missing: ${rel(wxml)}`)
    }
  }

  const sources = loadSources(feature.frontend || [])
  const pageSources = (feature.pages || []).map((page) => read(pagePath(`${page}.js`)))
  const mergedSource = sources.map((item) => item.source).concat(pageSources).join('\n')

  for (const item of sources) {
    if (!item.source) {
      failures.push(`frontend file missing or empty: ${item.file}`)
    }
  }

  for (const token of feature.tokens || []) {
    if (!mergedSource.includes(token)) {
      failures.push(`frontend token missing: ${token}`)
    }
  }

  for (const route of feature.apiRoutes || []) {
    if (!backendCovers(route, backendRoutes)) {
      failures.push(`backend route missing: ${route.method} ${route.path}`)
    }
  }

  return failures
}

const appPages = loadAppPages()
const backendRoutes = loadBackendRoutes()
const results = featureFlows.map((feature) => ({
  feature: feature.name,
  failures: checkFeature(feature, appPages, backendRoutes)
}))

const failed = results.filter((item) => item.failures.length > 0)

console.log(`feature_flows=${results.length}`)
console.log(`failed_feature_flows=${failed.length}`)

for (const item of results) {
  if (item.failures.length === 0) {
    console.log(`OK ${item.feature}`)
    continue
  }

  console.log(`FAIL ${item.feature}`)
  for (const failure of item.failures) {
    console.log(`  - ${failure}`)
  }
}

if (failed.length > 0) {
  process.exitCode = 1
}

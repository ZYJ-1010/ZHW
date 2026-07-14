const assert = require('assert')
const path = require('path')
const { isBackendRoleActive, normalizeActiveRole } = require('../utils/active-role')

const shellNavPath = path.resolve(__dirname, '../utils/shell-nav.js')

function runNavigation(pages, target, options = {}) {
  const calls = []

  global.getCurrentPages = () => pages
  global.wx = {
    getStorageSync() {
      return options.activeRole || ''
    },
    setStorageSync() {},
    navigateBack(payload) {
      calls.push({ type: 'navigateBack', payload })
    },
    navigateTo(payload) {
      calls.push({ type: 'navigateTo', payload })
    },
    redirectTo(payload) {
      calls.push({ type: 'redirectTo', payload })
    },
    reLaunch(payload) {
      calls.push({ type: 'reLaunch', payload })
    }
  }

  delete require.cache[shellNavPath]
  const { navigateShellRoute } = require(shellNavPath)
  navigateShellRoute(target, options)

  return calls
}

const staleDetailCalls = runNavigation([
  { route: 'pages/game/detail/index', options: { id: '101' } },
  { route: 'pages/game/hall/index', options: {} }
], 'pages/game/detail/index?id=202', {
  currentRoute: 'pages/game/hall/index'
})

assert.deepStrictEqual(staleDetailCalls.map((call) => call.type), ['navigateTo'])
assert.strictEqual(staleDetailCalls[0].payload.url, '/pages/game/detail/index?id=202')

const homeCardCalls = runNavigation([
  { route: 'pages/game/detail/index', options: { id: '101' } },
  { route: 'pages/home/player/index', options: {} }
], 'pages/game/detail/index?id=303', {
  currentRoute: 'pages/home/player/index'
})

assert.deepStrictEqual(homeCardCalls.map((call) => call.type), ['navigateTo'])
assert.strictEqual(homeCardCalls[0].payload.url, '/pages/game/detail/index?id=303')

const exactDetailCalls = runNavigation([
  { route: 'pages/game/detail/index', options: { id: '202' } },
  { route: 'pages/game/hall/index', options: {} }
], 'pages/game/detail/index?id=202', {
  currentRoute: 'pages/game/hall/index'
})

assert.deepStrictEqual(exactDetailCalls.map((call) => call.type), ['navigateBack'])
assert.strictEqual(exactDetailCalls[0].payload.delta, 1)

const reorderedQueryCalls = runNavigation([
  { route: 'pages/example/index', options: { b: 'two words', a: '1' } },
  { route: 'pages/game/hall/index', options: {} }
], 'pages/example/index?a=1&b=two%20words', {
  currentRoute: 'pages/game/hall/index'
})

assert.deepStrictEqual(reorderedQueryCalls.map((call) => call.type), ['navigateBack'])

const noReuseCalls = runNavigation([
  { route: 'pages/game/detail/index', options: { id: '202' } },
  { route: 'pages/game/hall/index', options: {} }
], 'pages/game/detail/index?id=202', {
  currentRoute: 'pages/game/hall/index',
  reuseExisting: false
})

assert.deepStrictEqual(noReuseCalls.map((call) => call.type), ['navigateTo'])

delete require.cache[shellNavPath]
global.getCurrentPages = () => [{ route: 'pages/message/index', options: {} }]
const homeCalls = []
global.wx = {
  getStorageSync: () => 'guide',
  navigateTo: (payload) => homeCalls.push(payload),
  redirectTo() {},
  reLaunch() {}
}
require(shellNavPath).navigateShellKey('home')
assert.strictEqual(homeCalls[0].url, '/pages/home/guide/index')

assert.strictEqual(isBackendRoleActive({
  roles: ['player', 'guide'],
  roleStatusMap: { player: 'approved', guide: 'approved' }
}, 'guide'), true)
assert.strictEqual(isBackendRoleActive({
  roles: ['player'],
  roleStatusMap: { player: 'approved', guide: 'approved' }
}, 'guide'), false)
assert.strictEqual(isBackendRoleActive({
  roles: ['player', 'expert'],
  roleStatusMap: { player: 'approved', expert: 'active' }
}, 'expert'), true)
assert.strictEqual(normalizeActiveRole('leader'), 'guide')
assert.strictEqual(normalizeActiveRole('master'), 'expert')

console.log('shell-nav regression tests passed')

const assert = require('assert')

const storage = {}
global.wx = {
  getStorageSync(key) {
    return storage[key] || ''
  },
  setStorageSync(key, value) {
    storage[key] = value
  },
  removeStorageSync(key) {
    delete storage[key]
  }
}

storage.enjoy_token = 'player-10002'

const session = require('../utils/auth-session')

assert.strictEqual(session.initializeAuthToken(), 'player-10002')

// 模拟另一个开发者工具运行实例覆盖共享 Storage；当前实例仍使用启动时锁定的 token。
storage.enjoy_token = 'expert-10001'
assert.strictEqual(session.getAuthToken(), 'player-10002')

assert.strictEqual(session.setAuthToken('guide-10003'), 'guide-10003')
assert.strictEqual(session.getAuthToken(), 'guide-10003')
assert.strictEqual(storage.enjoy_token, 'guide-10003')

session.clearAuthToken()
assert.strictEqual(session.getAuthToken(), '')
assert.strictEqual(storage.enjoy_token, undefined)

console.log('auth-session runtime isolation tests passed')

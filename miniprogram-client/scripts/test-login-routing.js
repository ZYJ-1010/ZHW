const assert = require('assert')

global.wx = {
  getAccountInfoSync() {
    return { miniProgram: { envVersion: 'develop' } }
  },
  getStorageSync() {
    return ''
  },
  setStorageSync() {},
  removeStorageSync() {},
  showToast() {}
}

let loginPage = null
global.Page = (definition) => {
  loginPage = definition
}

require('../pages/login/index.js')

assert.ok(loginPage, '登录页应成功注册')
assert.strictEqual(loginPage.isRealnameVerified({ realnameStatus: 'phone_verified' }), false)
assert.strictEqual(loginPage.isRealnameVerified({ identity: { status: 'sms_verified' } }), false)
assert.strictEqual(loginPage.isRealnameVerified({ realnameStatus: 'verified' }), true)

async function assertExistingLoginSkipsNewbie(entryFlow, authPageMode) {
  let existingCount = 0
  let newbieCount = 0
  const context = {
    data: { entryFlow },
    continueExistingAfterLogin() {
      existingCount += 1
    },
    showNewbieTasksAfterLogin() {
      newbieCount += 1
    }
  }

  await loginPage.continueAfterLogin.call(context, { authPageMode })
  assert.strictEqual(existingCount, 1, '已注册账号应直接进入原目标页面')
  assert.strictEqual(newbieCount, 0, '已注册账号不应重复进入新手任务')
}

async function run() {
  await assertExistingLoginSkipsNewbie('existing', 'login')
  await assertExistingLoginSkipsNewbie('register', 'login')
  console.log('login routing regression tests passed')
}

run().catch((error) => {
  console.error(error)
  process.exitCode = 1
})

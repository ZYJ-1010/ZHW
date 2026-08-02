const assert = require('assert')
const fs = require('fs')
const path = require('path')
const Module = require('module')

function capturePage(relativePath, mocks = {}) {
  const filePath = path.resolve(__dirname, '..', relativePath)
  const originalLoad = Module._load
  let definition = null

  global.Page = (pageDefinition) => {
    definition = pageDefinition
  }
  Module._load = function patchedLoad(request, parent, isMain) {
    for (const [key, value] of Object.entries(mocks)) {
      if (request.includes(key)) {
        return value
      }
    }
    return originalLoad.call(this, request, parent, isMain)
  }

  try {
    delete require.cache[filePath]
    require(filePath)
  } finally {
    Module._load = originalLoad
  }

  assert.ok(definition, `${relativePath} 应成功注册页面`)
  return definition
}

function pageContext(definition) {
  const context = {
    data: JSON.parse(JSON.stringify(definition.data || {})),
    setData(patch, callback) {
      this.data = Object.assign({}, this.data, patch)
      if (typeof callback === 'function') {
        callback()
      }
    }
  }

  Object.keys(definition).forEach((key) => {
    if (typeof definition[key] === 'function') {
      context[key] = definition[key]
    }
  })
  return context
}

function deferred() {
  let resolve
  let reject
  const promise = new Promise((resolvePromise, rejectPromise) => {
    resolve = resolvePromise
    reject = rejectPromise
  })
  return { promise, resolve, reject }
}

function flush() {
  return new Promise((resolve) => setImmediate(resolve))
}

async function withoutExpectedWarnings(callback) {
  const originalWarn = console.warn
  console.warn = () => {}
  try {
    return await callback()
  } finally {
    console.warn = originalWarn
  }
}

function installWx(overrides = {}) {
  global.wx = Object.assign({
    getMenuButtonBoundingClientRect: () => ({ top: 24, left: 320, height: 32 }),
    getSystemInfoSync: () => ({ windowWidth: 375, windowHeight: 812 }),
    getStorageSync: () => null,
    setStorageSync() {},
    showToast() {},
    showLoading() {},
    hideLoading() {},
    showShareMenu() {},
    hideShareMenu() {}
  }, overrides)
}

async function testInviteRecordAndMemberRoute() {
  const wxml = fs.readFileSync(path.resolve(__dirname, '../pages/profile/service-center/invite/records/index.wxml'), 'utf8')
  assert.ok(!/wx:else\s+class="invite-card-arrow"/.test(wxml), '邀约卡片箭头不能使用无相邻 wx:if 的 wx:else')

  let requested = null
  const definition = capturePage('pages/profile/service-center/invite/member-detail/index.js', {
    'services/profile': {
      getInviteMemberDetail: async (params) => {
        requested = params
        return { member: { name: '测试成员' } }
      }
    }
  })
  const context = pageContext(definition)
  context.onLoad({ memberId: '77' })
  await flush()

  assert.deepStrictEqual(requested, { memberId: '77' })
  assert.strictEqual(context.data.member.name, '测试成员')
}

async function testGameDetailRefreshShareAndGreet() {
  let detailCalls = 0
  let shareEntryCalls = 0
  let hideShareCalls = 0
  installWx({ hideShareMenu: () => { hideShareCalls += 1 } })

  const definition = capturePage('pages/game/detail/index.js', {
    'services/game': {
      getGameDetail: async () => {
        detailCalls += 1
        return {
          game: {
            id: 8,
            title: '测试局',
            status: 'recruiting',
            creatorUserId: 99,
            mainGuideUserId: 66,
            currentPlayers: 2,
            maxPlayers: 6
          },
          myRelation: { isMember: true, isCreator: false, canEnterIM: false },
          shareComponent: { enabled: false, enableInternal: false, enableWechat: false }
        }
      },
      createInviteEntry: async () => {
        shareEntryCalls += 1
        return { inviteCode: 'ABC001' }
      }
    },
    'services/invite': { saveInviteContext() {} },
    'services/user': { getCurrentUser: async () => ({ realnameStatus: 'verified' }) },
    'utils/shell-nav': { navigateShellRoute() {} },
    'utils/user-message': { toUserMessage: (value) => value },
    'config/routes': { ROUTES: { gameHall: '/pages/game/hall/index', gameDetail: '/pages/game/detail/index' } }
  })
  const context = pageContext(definition)

  context.onLoad({ gameId: '8' })
  await context.onShow()
  await flush()

  assert.strictEqual(detailCalls, 1, '详情页首次进入只能请求一次详情')
  assert.ok(hideShareCalls >= 1, '后台关闭微信分享时应隐藏原生分享菜单')
  assert.strictEqual(shareEntryCalls, 0, '后台关闭分享时不能生成分享入口')
  assert.ok(context.data.bottomTools.some((item) => item.key === 'greet'), '已入局但未开局时应保留打招呼入口用于提示')
  assert.ok(context.data.highlights.includes('主领路人：66'), '主领路人字段不能显示为主行家')

  context.setData({
    shareComponent: { enabled: true, enableWechat: true, enableInternal: true },
    shareEntry: null
  })
  const entries = await Promise.all([context.ensureShareEntry(), context.ensureShareEntry()])
  assert.strictEqual(shareEntryCalls, 1, '并发准备分享入口时只能请求一次')
  assert.strictEqual(entries[0].inviteCode, 'ABC001')
  assert.strictEqual(entries[1].inviteCode, 'ABC001')
}

async function testGameDetailWithdrawAndExitActions() {
  installWx()
  const definition = capturePage('pages/game/detail/index.js', {
    'services/game': {
      getGameDetail: async () => ({}),
      cancelGameApplication: async () => ({}),
      exitGame: async () => ({})
    },
    'services/invite': { saveInviteContext() {} },
    'services/user': { getCurrentUser: async () => ({ realnameStatus: 'verified' }) },
    'utils/shell-nav': { navigateShellRoute() {} },
    'utils/user-message': { toUserMessage: (value) => value },
    'config/routes': { ROUTES: { gameHall: '/pages/game/hall/index', gameDetail: '/pages/game/detail/index' } }
  })
  const context = pageContext(definition)
  let withdrawn = 0
  let exited = 0
  context.confirmWithdrawApplication = () => { withdrawn += 1 }
  context.confirmExitGame = () => { exited += 1 }

  context.setData({ primaryAction: { action: 'withdraw_application', disabled: false, applicationId: 12 } })
  await context.onPrimaryAction()
  assert.strictEqual(withdrawn, 1, '申请中状态必须提供撤回入口')

  context.setData({ primaryAction: { action: 'exit_game', disabled: false } })
  await context.onPrimaryAction()
  assert.strictEqual(exited, 1, '已入局未开局状态必须提供退出入口')
}

async function testCollaborationMilestoneAndCheckinPresentation() {
  installWx()
  const definition = capturePage('pages/game/collaboration/index.js', {
    'services/game': {
      getGameCollaboration: async () => ({
        gameId: 9,
        title: '测试协作局',
        statusText: '进行中',
        milestones: [{ id: 3, title: '确认方案', status: 'pending' }],
        checkins: [{
          id: 4,
          checkinType: 'progress',
          content: '已完成现场确认',
          userName: '成员甲',
          createdAtText: '08-02 10:30'
        }],
        actions: { canManageMilestones: true, canCheckin: true }
      }),
      createMilestone: async () => ({}),
      updateMilestone: async () => ({}),
      createGameCheckin: async () => ({})
    },
    'utils/shell-nav': { navigateShellRoute() {} },
    'utils/user-message': { toUserMessage: (value) => value },
    'utils/adaptive-shell-layout': {
      getLegacyWhiteFrameLayoutStyles: () => ({ frameStyle: '', contentStyle: '', titleStyle: '', backStyle: '', safeBottomRpx: 0 })
    },
    'config/routes': { ROUTES: { gameParticipants: '/pages/game/participants/index', gameDelivery: '/pages/game/delivery/index' } }
  })
  const context = pageContext(definition)
  await context.loadCollaboration('9')

  assert.strictEqual(context.data.milestones[0].title, '确认方案', '协作页应展示里程碑')
  assert.strictEqual(context.data.checkins[0].userName, '成员甲', '协作页打卡应展示提交成员')
  assert.strictEqual(context.data.checkins[0].createdAtText, '08-02 10:30', '协作页打卡应展示提交时间')
  assert.strictEqual(context.data.actions.canManageMilestones, true)
  assert.strictEqual(context.data.actions.canCheckin, true)
}

async function testApplySubmitLock() {
  installWx()
  let applyCalls = 0
  const subscription = deferred()
  const definition = capturePage('pages/game/apply/index.js', {
    'services/game': { applyGame: async () => { applyCalls += 1 } },
    'services/file': {},
    'services/profile': {},
    'utils/shell-nav': { navigateShellKey() {}, navigateShellRoute() {} },
    'config/routes': { ROUTES: { gameDetail: '/pages/game/detail/index', gameApply: '/pages/game/apply/index' } }
  })
  const context = pageContext(definition)
  context.setData({
    eligibilityReady: true,
    gameId: '9',
    selectedRole: 'player',
    applicationConfig: {
      requireIntro: false,
      requireAgreement: false,
      uploadRequired: false,
      minIntroLength: 0,
      texts: {}
    },
    form: { intro: '', message: '', imageFiles: [], attachmentFiles: [], agreed: false }
  })
  context.requestApplicationResultSubscription = () => subscription.promise
  context.uploadApplicationFiles = async () => []
  context.buildApplyReason = () => ''
  context.showInfo = () => {}

  const originalSetTimeout = global.setTimeout
  global.setTimeout = () => 1
  try {
    const firstSubmit = context.onSubmit()
    await context.onSubmit()
    assert.strictEqual(applyCalls, 0, '订阅授权未结束前不能提前提交报名')
    subscription.resolve()
    await firstSubmit
  } finally {
    global.setTimeout = originalSetTimeout
  }

  assert.strictEqual(applyCalls, 1, '连续点击报名只能提交一次')
  assert.strictEqual(context.data.submitting, false)
}

async function testIMLoadsBeforeSocketAndKeepsAvatar() {
  installWx()
  const roomRequest = deferred()
  let socketCalls = 0
  const definition = capturePage('pages/im/room/index.js', {
    'services/im': {
      getChatRoom: () => roomRequest.promise,
      getRoomByGame: async () => ({ roomId: 10, title: '协作局' }),
      getMessages: async () => ({ items: [{ id: 20, gameId: 5, senderUserId: 2, messageType: 'text', content: '你好' }] }),
      connectGameSocket: () => {
        socketCalls += 1
        return { close() {}, isOpen: () => true }
      }
    },
    'services/file': {},
    'services/game': { getGameDetail: async () => ({ game: { title: '协作局' } }) },
    'utils/toast': { info() {} },
    'utils/shell-nav': { navigateShellBack: () => false, navigateShellRoute() {} },
    'config/routes': { ROUTES: { gameDetail: '/pages/game/detail/index', message: '/pages/message/index' } }
  })
  const context = pageContext(definition)

  context.onLoad({ gameId: '5' })
  const showPromise = context.onShow()
  assert.strictEqual(socketCalls, 0, '房间权限和成员数据返回前不能连接实时通道')
  roomRequest.resolve({
    id: 10,
    currentUserId: 1,
    status: 'active',
    memberIds: [1, 2],
    members: [
      { userId: 1, name: '我', avatarUrl: '/avatars/me.png' },
      { userId: 2, name: '成员甲', avatarUrl: '/avatars/member.png' }
    ]
  })
  await showPromise

  assert.strictEqual(socketCalls, 1)
  assert.strictEqual(context.data.messages[1].avatarSrc, '/avatars/member.png', '消息头像应回退使用房间成员头像')
  context.onHide()
}

async function testMyGamesKeepsCoreListsWhenFavoriteFails() {
  installWx()
  const pageConfig = {
    pageTitle: '我的局',
    emptyText: '暂无记录',
    categoryTabs: [],
    statusTabs: [{ key: 'all', text: '全部' }]
  }
  const definition = capturePage('pages/profile/service-center/my-games/index.js', {
    'services/game': {
      getPlayerGameManage: async () => ({
        pageConfig,
        orders: [
          { gameId: 1, title: '参与局', status: 'recruiting' },
          { gameId: 3, title: '已完成局', statusType: 'complete', game: { id: 3, title: '已完成局', status: 'completed' } }
        ]
      }),
      getGameManage: async () => ({ orders: [{ gameId: 2, title: '管理局', status: 'recruiting' }] }),
      getMyFavoriteGames: async () => { throw new Error('收藏接口暂不可用') }
    },
    'utils/toast': { info() {} },
    'utils/shell-nav': { navigateShellRoute() {} },
    'utils/adaptive-shell-layout': {
      getProfileWhiteShellLayoutStyles: () => ({
        capsuleTopRpx: 0,
        capsuleHeightRpx: 0,
        contentStyle: '',
        frameStyle: '',
        titleStyle: '',
        backStyle: ''
      })
    },
    'config/routes': {
      ROUTES: {
        gameDetail: 'pages/game/detail/index',
        gamePlayAgain: 'pages/game/play-again/index'
      }
    }
  })
  const context = pageContext(definition)
  await withoutExpectedWarnings(() => context.loadMyGames())

  assert.strictEqual(context.data.cards.length, 3, '收藏接口失败不能清空参与和管理记录')
  assert.ok(context.data.cards.some((item) => item.category === 'joined'))
  assert.ok(context.data.cards.some((item) => item.category === 'created'))
  const completedCard = context.data.cards.find((item) => item.gameId === 3)
  assert.strictEqual(
    completedCard.actions.find((item) => item.type === 'playAgain').route,
    '/pages/game/play-again/index?gameId=3',
    '正式完成的局必须进入必选的再来一局流程'
  )
}

async function testProfileClearsHiddenBannerAndExpiredAccount() {
  installWx()
  let call = 0
  const definition = capturePage('pages/profile/index.js', {
    'services/profile': {
      getProfileHome: async () => {
        call += 1
        if (call === 1) {
          return {
            user: { nickname: '旧账号', avatarText: '旧' },
            stats: [{ label: '引荐数', value: '9' }],
            assets: [{ label: '可用积分', value: '99' }],
            serviceSections: [],
            vipBanner: { visible: true, text: '旧会员入口' }
          }
        }
        if (call === 2) {
          return {
            user: { nickname: '新账号', avatarText: '新' },
            serviceSections: [],
            vipBanner: { visible: false }
          }
        }
        if (call === 3) {
          throw new Error('request:fail network disconnected')
        }
        const error = new Error('登录已过期')
        error.authExpired = true
        throw error
      }
    },
    'utils/toast': { info() {} },
    'utils/shell-nav': { navigateShellKey() {}, navigateShellRoute() {} },
    'utils/auth-error': {
      AUTH_EXPIRED_MESSAGE: '登录已过期，请重新登录',
      goLogin() {},
      isAuthExpiredError: (error) => error && error.authExpired === true
    },
    'utils/active-role': { getActiveRole: () => 'player' },
    'config/routes': { ROUTES: {} }
  })
  const context = pageContext(definition)

  await context.loadProfileHome()
  assert.strictEqual(context.data.vipBanner.visible, true)
  await context.loadProfileHome()
  assert.strictEqual(context.data.vipBanner, null, '后端隐藏会员入口后不能保留旧账号横幅')
  assert.strictEqual(context.data.stats[0].value, '0', '新响应缺少统计时不能沿用旧账号统计')
  await withoutExpectedWarnings(() => context.loadProfileHome())
  assert.strictEqual(context.data.user.nickname, '未登录', '普通加载失败也不能保留旧账号资料')
  assert.strictEqual(context.data.loadError, '网络连接异常，请检查网络后重试')
  await withoutExpectedWarnings(() => context.loadProfileHome())
  assert.strictEqual(context.data.user.nickname, '未登录', '登录失效后必须清除旧账号资料')
  assert.strictEqual(context.data.assets[0].value, '0')
}

async function run() {
  await testInviteRecordAndMemberRoute()
  await testGameDetailRefreshShareAndGreet()
  await testGameDetailWithdrawAndExitActions()
  await testCollaborationMilestoneAndCheckinPresentation()
  await testApplySubmitLock()
  await testIMLoadsBeforeSocketAndKeepsAvatar()
  await testMyGamesKeepsCoreListsWhenFavoriteFails()
  await testProfileClearsHiddenBannerAndExpiredAccount()
  console.log('core flow regression tests passed')
}

run().catch((error) => {
  console.error(error)
  process.exitCode = 1
})

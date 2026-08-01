const assert = require('assert')
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
    setData(patch) {
      this.data = Object.assign({}, this.data, patch)
    }
  }
  Object.keys(definition).forEach((key) => {
    if (typeof definition[key] === 'function') {
      context[key] = definition[key]
    }
  })
  return context
}

const roleGrowthPayload = {
  activeRoleCode: 'guide',
  activeRoleName: '领路人',
  activeRoleGrowth: {
    roleCode: 'guide',
    roleName: '领路人',
    active: true,
    levelNo: 0,
    levelTitle: '0级领路人',
    score: 0,
    scoreLabel: '综合得分',
    progressPercent: 0,
    description: '尚无可审计的带动记录',
    metrics: []
  },
  roleGrowth: [
    {
      roleCode: 'player',
      roleName: '玩家',
      active: true,
      levelNo: 1,
      levelTitle: 'Lv1 初次体验',
      score: 200,
      scoreLabel: '累计经验',
      progressPercent: 50,
      metrics: [{ metricCode: 'experience', title: '累计经验', rawValue: 200, targetValue: 300, score: 50, unit: '经验' }]
    },
    {
      roleCode: 'guide',
      roleName: '领路人',
      active: true,
      levelNo: 0,
      levelTitle: '0级领路人',
      score: 0,
      scoreLabel: '综合得分',
      progressPercent: 0,
      description: '尚无可审计的带动记录',
      metrics: []
    }
  ],
  profile: { experience: 200 },
  creditScore: 0,
  availablePoints: 0,
  achievements: [],
  footprints: [],
  achievementConfig: { filters: [{ key: 'all', label: '全部' }], locked: [] }
}

async function testAchievementSummaryUsesActiveRole() {
  let requestParams = null
  const definition = capturePage('pages/profile/achievements/index.js', {
    'services/profile': { getGrowth: async (params) => { requestParams = params; return roleGrowthPayload } },
    'utils/toast': { info() {} },
    'utils/active-role': {
      getActiveRole: () => 'guide',
      normalizeActiveRole: (value) => value === 'guide' ? 'guide' : (value === 'expert' ? 'expert' : 'player')
    }
  })
  const context = pageContext(definition)
  await context.loadGrowth()

  assert.deepStrictEqual(requestParams, { roleType: 'guide' })
  assert.strictEqual(context.data.activeRoleCode, 'guide')
  assert.strictEqual(context.data.activeRoleName, '领路人')
  assert.strictEqual(context.data.levelName, '0级领路人')
  assert.strictEqual(context.data.xpText, '0 分')
  assert.strictEqual(context.data.creditText, '0')
  assert.strictEqual(context.data.pointsText, '0')
  assert.strictEqual(context.data.roleProfiles.find((item) => item.roleCode === 'guide').selected, true)
}

async function testProfileHomeRequestsActiveRole() {
  let requestParams = null
  const definition = capturePage('pages/profile/index.js', {
    'services/profile': {
      getProfileHome: async (params) => {
        requestParams = params
        return {
          user: { nickname: '测试用户', role: '领路人', roleLevel: '0级领路人', avatarText: '测' },
          stats: [],
          assets: [],
          serviceSections: []
        }
      }
    },
    'utils/toast': { info() {} },
    'utils/shell-nav': { navigateShellKey() {}, navigateShellRoute() {} },
    'utils/auth-error': {
      AUTH_EXPIRED_MESSAGE: '登录已过期',
      goLogin() {},
      isAuthExpiredError: () => false
    },
    'utils/active-role': { getActiveRole: () => 'guide' }
  })
  const context = pageContext(definition)
  await context.loadProfileHome()

  assert.deepStrictEqual(requestParams, { roleType: 'guide' })
  assert.strictEqual(context.data.user.role, '领路人')
  assert.strictEqual(context.data.user.roleLevel, '0级领路人')
}

async function testFootprintHeroUsesActiveRole() {
  let requestParams = null
  const definition = capturePage('pages/profile/footprint/achievements/index.js', {
    'services/profile': { getGrowth: async (params) => { requestParams = params; return roleGrowthPayload } },
    'utils/shell-nav': { navigateShellKey() {} },
    'utils/active-role': {
      getActiveRole: () => 'guide',
      normalizeActiveRole: (value) => value === 'guide' ? 'guide' : (value === 'expert' ? 'expert' : 'player')
    }
  })
  const context = pageContext(definition)
  await context.loadAchievements()

  assert.deepStrictEqual(requestParams, { roleType: 'guide' })
  assert.strictEqual(context.data.levelTitle, '0级领路人')
  assert.strictEqual(context.data.activeRoleName, '领路人')
  assert.strictEqual(context.data.progress, 0)
  assert.strictEqual(context.data.onlineText, '成长中心')
  assert.strictEqual(context.data.roleGrowth.find((item) => item.roleCode === 'guide').selected, true)
}

async function testCreditCenterKeepsZeroAndAppealState() {
  const navigations = []
  const definition = capturePage('pages/profile/credit-center/index.js', {
    'services/profile': {
      getCreditCenter: async () => ({
        score: 0,
        scoreLabel: '信用分',
        level: '冻结',
        status: 'frozen',
        monthlyDelta: 0,
        isPermanent: true,
        summary: [
          { label: '当前状态', value: '冻结' },
          { label: '正向记录', value: '0条' },
          { label: '扣分记录', value: '1条' }
        ],
        records: [{
          id: 9,
          creditLogId: 9,
          title: '低分评价',
          desc: '低分评价',
          score: 0,
          canAppeal: false,
          appealId: 17,
          appealStatusText: '申诉处理中',
          appealRoute: '/pages/profile/system-management/appeal-detail/index?reportId=17'
        }],
        appealEntry: { enabled: false, count: 0, route: '', text: '暂无可申诉记录' }
      })
    },
    'utils/toast': { info() {} },
    'utils/shell-nav': { navigateShellRoute: (route) => navigations.push(route) }
  })
  const context = pageContext(definition)
  await context.loadCreditCenter()

  assert.strictEqual(context.data.score, 0)
  assert.strictEqual(context.data.statusTone, 'red')
  assert.strictEqual(context.data.monthlyDelta, '0')
  assert.strictEqual(context.data.records[0].score, '0')
  assert.strictEqual(context.data.records[0].tone, 'neutral')
  assert.strictEqual(context.data.appealEntry.enabled, false)
  assert.strictEqual(context.data.appealEntry.route, '')
  context.handleRecordTap({ currentTarget: { dataset: context.data.records[0] } })
  assert.deepStrictEqual(navigations, ['/pages/profile/system-management/appeal-detail/index?reportId=17'])
}

async function run() {
  await testProfileHomeRequestsActiveRole()
  await testAchievementSummaryUsesActiveRole()
  await testFootprintHeroUsesActiveRole()
  await testCreditCenterKeepsZeroAndAppealState()
  console.log('profile growth and credit regression tests passed')
}

run().catch((error) => {
  console.error(error)
  process.exitCode = 1
})

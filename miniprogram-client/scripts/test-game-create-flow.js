const assert = require('assert')

const storage = new Map()
global.wx = {
  getAccountInfoSync() {
    return { miniProgram: { envVersion: 'develop' } }
  },
  getStorageSync(key) {
    return storage.get(key)
  },
  setStorageSync(key, value) {
    storage.set(key, value)
  },
  removeStorageSync(key) {
    storage.delete(key)
  },
  showToast() {},
  navigateBack() {},
  getSystemInfoSync() {
    return { windowWidth: 375 }
  }
}

function clone(value) {
  return JSON.parse(JSON.stringify(value))
}

function setByPath(target, path, value) {
  const parts = String(path).split('.')
  let current = target
  parts.slice(0, -1).forEach((part) => {
    if (!current[part] || typeof current[part] !== 'object') {
      current[part] = {}
    }
    current = current[part]
  })
  current[parts[parts.length - 1]] = value
}

function createContext(definition, overrides = {}) {
  const context = Object.assign({}, definition, overrides)
  context.data = Object.assign(clone(definition.data || {}), clone(overrides.data || {}))
  context.setData = function setData(patch, callback) {
    Object.keys(patch || {}).forEach((key) => setByPath(this.data, key, patch[key]))
    if (typeof callback === 'function') callback()
  }
  return context
}

let createPage = null
global.Page = (definition) => {
  createPage = definition
}
require('../pages/game/create/index.js')
assert.ok(createPage, '发起组局页应成功注册')

const futureGameTime = {
  startDate: '2099-08-01',
  startTime: '10:00',
  endDate: '2099-08-01',
  endTime: '12:00'
}
const futureSignupTime = {
  startDate: '2099-07-31',
  startTime: '10:00',
  endDate: '2099-08-01',
  endTime: '09:00'
}

async function run() {
  const invalidCategoryContext = createContext(createPage, {
    data: {
      gameTypes: [{ key: 'social', name: '社交局', tags: [] }],
      createForm: { capacity: { min: 5, max: 8 }, tags: [] },
      form: {
        theme: '测试局',
        type: 'removed-category',
        capacity: 5,
        participation: 'offline',
        allowedRoles: ['player'],
        intro: '介绍',
        highlights: '亮点',
        description: '描述',
        notice: '',
        audience: ''
      },
      coverImage: 'https://example.com/cover.jpg',
      gameTimeConfirmed: true,
      signupTimeConfirmed: true,
      timeDraft: futureGameTime,
      signupTimeDraft: futureSignupTime,
      locationInfo: { address: '西湖', latitude: 30.2, longitude: 120.1 },
      completionRules: [{ key: 'goal', name: '目标达成', active: true }]
    }
  })
  assert.ok(invalidCategoryContext.getPublishMissingFields().includes('局类型'), '失效的大类不能静默回退到第一类')
  assert.strictEqual(invalidCategoryContext.buildCreateGamePayload(), null, '失效的大类不能生成发布参数')

  const timeContext = createContext(createPage, {
    data: {
      timePanelVisible: false,
      timeDraft: futureGameTime,
      gameTimeConfirmed: true,
      gameTimeSummary: { startText: '原开始', endText: '原结束', durationText: '2小时' },
      scheduleFields: [{ key: 'gameTime', value: '原局时间' }]
    }
  })
  timeContext.openGameTimePanel()
  timeContext.onGameTimePickerChange({ currentTarget: { dataset: { field: 'endTime' } }, detail: { value: '13:00' } })
  assert.strictEqual(timeContext.data.gameTimeConfirmed, false, '编辑中的时间不能继续作为已确认时间发布')
  timeContext.cancelGameTimePanel()
  assert.deepStrictEqual(timeContext.data.timeDraft, futureGameTime, '取消编辑应恢复确认前的时间')
  assert.strictEqual(timeContext.data.gameTimeConfirmed, true, '取消编辑应恢复原确认状态')

  const mediaContext = createContext(createPage, {
    data: { descriptionMedia: [{ id: 'video-1', type: 'video' }] }
  })
  const appended = await mediaContext.appendDescriptionMedia({ id: 'video-2', type: 'video' })
  assert.strictEqual(appended, false, '局描述视频应与后端一致最多保留一个')
  assert.strictEqual(mediaContext.data.descriptionMedia.length, 1)

  let publishCount = 0
  const actionContext = createContext(createPage, {
    data: { draftSaving: false, previewPreparing: false, publishing: false },
    publishGame() {
      publishCount += 1
    }
  })
  actionContext.activePreviewSessionId = 'current-preview'
  storage.set('game_create_preview_action_v1', { type: 'publish', sessionId: 'old-preview', createdAt: Date.now() })
  actionContext.handlePreviewPublishAction()
  await new Promise((resolve) => setTimeout(resolve, 5))
  assert.strictEqual(publishCount, 0, '旧预览留下的发布动作不能误发当前表单')
  assert.strictEqual(storage.has('game_create_preview_action_v1'), false, '失效发布动作应被消费并清理')

  storage.set('game_create_preview_action_v1', { type: 'publish', sessionId: 'current-preview', createdAt: Date.now() })
  actionContext.handlePreviewPublishAction()
  await new Promise((resolve) => setTimeout(resolve, 5))
  assert.strictEqual(publishCount, 1, '仅当前预览会话可以触发发布')

  const gameService = require('../services/game')
  const fileService = require('../services/file')
  const originalGetCategoryConfig = gameService.getCategoryConfig
  const originalGetGameDraft = gameService.getGameDraft
  const originalGetDownloadURL = fileService.getDownloadURL
  gameService.getCategoryConfig = async () => ({
    defaultPrimaryCategory: 'task',
    primaryCategories: [{ key: 'task', name: '任务局', tags: [] }],
    createForm: {
      capacity: { min: 5, max: 8 },
      participationModes: [{ key: 'offline', name: '线下' }],
      completionRules: [{ key: 'goal', name: '目标达成' }]
    }
  })
  const categoryContext = createContext(createPage)
  const categoryLoad = categoryContext.loadCategoryConfig()
  assert.ok(categoryLoad && typeof categoryLoad.then === 'function', '分类加载必须返回 Promise，草稿恢复才能等待配置完成')
  await categoryLoad
  assert.strictEqual(categoryContext.data.form.type, 'task')
  gameService.getGameDraft = async () => ({
    id: 88,
    title: '任务草稿',
    payload: {
      form: {
        theme: '任务草稿任务草稿任务草稿任务草稿任务草稿任务草稿',
        type: 'task',
        capacity: 5,
        participation: 'offline',
        allowedRoles: ['player'],
        intro: '介绍',
        highlights: '亮点',
        description: '描述',
        notice: '',
        audience: ''
      },
      timeDraft: futureGameTime,
      signupTimeDraft: futureSignupTime,
      gameTimeConfirmed: true,
      signupTimeConfirmed: true,
      locationInfo: { address: '西湖', latitude: '30.2', longitude: '120.1' },
      descriptionMedia: [{ id: 'image-9', type: 'image', fileId: 9, url: 'expired' }],
      activeTags: ['project'],
      activeCompletionRules: ['goal']
    }
  })
  fileService.getDownloadURL = async (fileId) => `https://static.example.com/${fileId}`
  const restoreContext = createContext(createPage, {
    data: {
      gameTypes: [
        { key: 'social', name: '社交局', tags: [{ key: 'table', name: '桌游' }] },
        { key: 'task', name: '任务局', tags: [{ key: 'project', name: '项目协作' }] }
      ],
      createForm: { capacity: { min: 5, max: 8 }, tags: [] },
      form: { allowedRoles: ['player'] },
      completionRules: [{ key: 'goal', name: '目标达成', active: false }],
      scheduleFields: clone(createPage.data.scheduleFields)
    }
  })
  const restored = await restoreContext.restoreRequestedDraft('88')
  assert.strictEqual(restored, true)
  assert.strictEqual(restoreContext.data.form.type, 'task')
  assert.strictEqual(restoreContext.data.form.theme.length, 20, '服务器草稿恢复时仍须遵守页面字段长度')
  assert.deepStrictEqual(restoreContext.data.tags.map((item) => [item.key, item.active]), [['project', true]], '草稿标签应按草稿的大类恢复')
  assert.strictEqual(restoreContext.data.descriptionMedia[0].url, 'https://static.example.com/9', '服务器草稿媒体应刷新下载地址')
  assert.strictEqual(typeof restoreContext.data.locationInfo.latitude, 'number', '草稿坐标应恢复为发布契约需要的数值')
  gameService.getGameDraft = originalGetGameDraft
  gameService.getCategoryConfig = originalGetCategoryConfig
  fileService.getDownloadURL = originalGetDownloadURL

  await assert.rejects(() => gameService.getGameDraft('not-a-number'), /草稿编号无效/)
  await assert.rejects(() => gameService.saveGameDraft({ id: 'not-a-number', payload: {} }), /草稿编号无效/)

  const request = require('../api/request')
  const originalPost = request.post
  const capturedRequests = []
  request.post = async (url, data, options) => {
    capturedRequests.push({ url, data, options })
    return { code: 0, data: { id: capturedRequests.length } }
  }
  await gameService.createGame({ title: '同一局' }, 'create-session-a')
  await gameService.createGame({ title: '同一局' }, 'create-session-a')
  await gameService.createGame({ title: '同一局' }, 'create-session-b')
  let idempotencyKeys = capturedRequests.map((item) => item.options.header['Idempotency-Key'])
  assert.strictEqual(idempotencyKeys[0], idempotencyKeys[1], '同一次发布重试应复用幂等标识')
  assert.notStrictEqual(idempotencyKeys[1], idempotencyKeys[2], '新的发布会话不能复用旧幂等标识')
  await gameService.saveGameDraft({
    submissionSessionId: 'draft-session-a',
    title: '同一草稿',
    payload: { savedAt: 1, coverFileId: 9, coverImage: 'old-url', descriptionMedia: [{ id: 'm1', type: 'image', fileId: 10, url: 'old-media-url' }] }
  })
  await gameService.saveGameDraft({
    submissionSessionId: 'draft-session-a',
    title: '同一草稿',
    payload: { savedAt: 2, coverFileId: 9, coverImage: 'new-url', descriptionMedia: [{ id: 'm1', type: 'image', fileId: 10, url: 'new-media-url' }] }
  })
  idempotencyKeys = capturedRequests.map((item) => item.options.header['Idempotency-Key'])
  assert.strictEqual(idempotencyKeys[3], idempotencyKeys[4], '草稿重试不应因保存时间或临时下载地址变化而失去幂等保护')
  request.post = originalPost
  let transmittedHeader = null
  wx.request = (options) => {
    transmittedHeader = options.header
    options.success({ statusCode: 200, data: { code: 0, data: {} } })
  }
  await request.post('/api/app/test-idempotency', {}, { header: { 'Idempotency-Key': 'test-key' } })
  assert.strictEqual(transmittedHeader['Idempotency-Key'], 'test-key', '请求层应把幂等标识发送给后端')

  let previewPage = null
  global.Page = (definition) => {
    previewPage = definition
  }
  delete require.cache[require.resolve('../pages/game/create-preview/index.js')]
  storage.set('game_create_preview_v1', {
    previewSessionId: 'preview-session-1',
    form: { theme: '预览局', allowedRoles: ['expert'] },
    descriptionMedia: [
      { id: 'image-preview', type: 'image', url: 'https://static.example.com/image.jpg' },
      { id: 'video-preview', type: 'video', url: 'https://static.example.com/video.mp4' }
    ]
  })
  require('../pages/game/create-preview/index.js')
  const previewContext = createContext(previewPage)
  previewContext.onLoad()
  assert.deepStrictEqual(previewContext.data.preview.media.map((item) => item.type), ['image', 'video'], '预览应同时保留图片和视频')
  assert.deepStrictEqual(previewContext.data.preview.allowedRoleLabels, ['行家'], '预览应展示本局开放身份')
  previewContext.publish()
  assert.strictEqual(storage.get('game_create_preview_action_v1').sessionId, 'preview-session-1', '预览发布动作应绑定本次预览会话')

  console.log('game create flow regression tests passed')
}

run().catch((error) => {
  console.error(error)
  process.exitCode = 1
})

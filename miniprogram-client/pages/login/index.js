const authService = require('../../services/auth')
const inviteService = require('../../services/invite')
const newbieService = require('../../services/newbie')
const userService = require('../../services/user')
const toast = require('../../utils/toast')
const env = require('../../config/env')
const { ROUTES } = require('../../config/routes')
const { navigateShellRoute } = require('../../utils/shell-nav')
const { getAuthToken } = require('../../utils/auth-session')

const TEST_PHONE = '13888888888'
const TEST_REGISTER_PHONE = '13700000000'
const TEST_CODE = '000000'
const TEST_PASSWORD = 'Test123456'
const LOGIN_WALKTHROUGH_MODES = ['home', 'codeVerify', 'account']
const INVITE_REQUIRED_MESSAGE = '小程序需要邀请才可以进入'
const INVITE_INVALID_MESSAGE = '邀请码无效，请检查邀请链接或联系邀请人'
const INITIAL_RESEND_SECONDS = 30

function getInviteErrorMessage(error) {
  const message = String(error && error.message ? error.message : error || '').trim()
  const lowerMessage = message.toLowerCase()

  if (lowerMessage === 'invite expired' || /邀请码已失效/.test(message)) {
    return '邀请码已失效，请重新获取邀请码'
  }

  if (lowerMessage === 'invite code required' || /邀请码.*(必填|缺失|为空)|需要邀请/.test(message)) {
    return INVITE_REQUIRED_MESSAGE
  }

  if (lowerMessage === 'invalid invite code' || /邀请码.*(无效|不存在|错误)/.test(message)) {
    return INVITE_INVALID_MESSAGE
  }

  if (lowerMessage === 'invite code already bound' || /邀请码.*(已绑定|已使用)/.test(message)) {
    return '该邀请码已绑定，请联系邀请人或管理员处理'
  }

  return ''
}

function isInviteError(error) {
  return Boolean(getInviteErrorMessage(error))
}

function isExpiredInviteError(error) {
  return /邀请码已失效|invite expired/i.test(String(error && error.message ? error.message : error || ''))
}

function getSmsSendErrorMessage(error) {
  const message = String((error && (error.message || error.errMsg)) || '').trim()
  const lowerMessage = message.toLowerCase()

  if (/验证码发送过于频繁|今日验证码发送次数已达上限|手机号格式不正确/.test(message)) {
    return message
  }

  if (/invalid phone|phone invalid|phone required/.test(lowerMessage)) {
    return '手机号格式不正确'
  }

  if (/sms.*(rate|daily|too frequent)|too many requests|\b429\b/.test(lowerMessage)) {
    return '验证码发送过于频繁，请稍后再试'
  }

  if (/request:fail|network|timeout|failed to connect|connection|gateway|\b5\d\d\b/.test(lowerMessage)) {
    return '网络异常，请检查网络后重试'
  }

  return '验证码发送失败，请稍后重试'
}

function safeDecode(value) {
  const text = String(value || '').trim()

  try {
    return decodeURIComponent(text)
  } catch (error) {
    return text
  }
}

function normalizeReturnRoute(value) {
  const rawRoute = safeDecode(value).trim()
  const route = rawRoute.startsWith('pages/') ? `/${rawRoute}` : rawRoute
  if (!route || !route.startsWith('/pages/') || route.startsWith(`/${ROUTES.login}`)) {
    return ''
  }
  return route
}

function decodeInviteScene(scene) {
  const decoded = safeDecode(scene)

  if (!decoded) {
    return {}
  }

  if (!decoded.includes('=') && !decoded.includes('&')) {
    return { inviteCode: decoded }
  }

  return decoded.split('&').reduce((result, item) => {
    const index = item.indexOf('=')
    const key = safeDecode(index >= 0 ? item.slice(0, index) : item)
    const value = safeDecode(index >= 0 ? item.slice(index + 1) : '')

    if (key) {
      result[key] = value
    }

    return result
  }, {})
}

function normalizeInviteEntryType(value) {
  const entryType = String(value || '').trim().toLowerCase()
  const entryTypeMap = {
    p: 'poster',
    poster: 'poster',
    card: 'poster',
    share_card: 'poster',
    '小程序卡片': 'poster',
    '海报': 'poster',
    q: 'qrcode',
    qr: 'qrcode',
    qrcode: 'qrcode',
    code: 'qrcode',
    '二维码': 'qrcode',
    l: 'link',
    link: 'link',
    url: 'link',
    '链接': 'link'
  }

  return entryTypeMap[entryType] || entryType
}

function resolveInviteRouteParams(options = {}) {
  const sceneParams = decodeInviteScene(options.scene)
  const inviteCode = inviteService.normalizeInviteCode(
    options.inviteCode ||
    options.code ||
    sceneParams.inviteCode ||
    sceneParams.code ||
    sceneParams.i
  )
  const entryType = normalizeInviteEntryType(
    options.entryType ||
    options.type ||
    sceneParams.entryType ||
    sceneParams.type ||
    sceneParams.t ||
    sceneParams.entry ||
    sceneParams['邀请入口']
  )

  return {
    inviteCode,
    entryType
  }
}

function getTestLoginDefaults() {
  if (!env.isMock) {
    return {}
  }

  return {
    phone: TEST_PHONE,
    verifyCode: TEST_CODE,
    codeDigits: TEST_CODE.split(''),
    isCodeComplete: true,
    password: TEST_PASSWORD,
    inviteCode: ''
  }
}

const DEFAULT_NEWBIE_TASKS = []

const NEWBIE_TASK_META = {
  complete_identity: {
    type: 'realname',
    rewardText: '+50 经验值',
    actionText: '去完成'
  },
  realname: {
    type: 'realname',
    rewardText: '+50 经验值',
    actionText: '去完成'
  },
  apply_role: {
    type: 'role_apply',
    rewardText: '+50 经验值',
    actionText: '去完成'
  },
  join_or_create_game: {
    type: 'first_game',
    rewardText: '+100 经验值',
    actionText: '去完成'
  },
  first_game: {
    type: 'first_game',
    rewardText: '+100 经验值',
    actionText: '去完成'
  },
  complete_game: {
    type: 'complete_game',
    rewardText: '+100 经验值',
    actionText: '去完成'
  },
  submit_review: {
    type: 'review',
    rewardText: '+30 经验值',
    actionText: '去完成'
  },
  profile: {
    type: 'profile',
    rewardText: '+30 经验值',
    actionText: '去完成'
  }
}

function pickNumber() {
  for (let index = 0; index < arguments.length; index++) {
    const value = arguments[index]

    if (typeof value === 'number' && Number.isFinite(value)) {
      return value
    }

    if (typeof value === 'string' && value.trim() !== '') {
      const numberValue = Number(value)

      if (Number.isFinite(numberValue)) {
        return numberValue
      }
    }
  }

  return null
}

function normalizeRewardText(value) {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return `${value > 0 ? '+' : ''}${value} 经验值`
  }

  const text = String(value || '').trim()

  if (!text) {
    return ''
  }

  if (/经验值|积分|XP/i.test(text)) {
    return text
  }

  if (/^[+-]?\d+(\.\d+)?$/.test(text)) {
    return `${text[0] === '-' || text[0] === '+' ? text : `+${text}`} 经验值`
  }

  return text
}

function pickRewardText(source, meta) {
  const rewardParts = []
  const experience = Number(source.rewardExperience)
  const points = Number(source.rewardPoints)

  if (Number.isFinite(experience) && experience > 0) {
    rewardParts.push(`+${experience} 经验值`)
  }
  if (Number.isFinite(points) && points > 0) {
    rewardParts.push(`+${points} 积分`)
  }
  if (rewardParts.length) {
    return rewardParts.join(' · ')
  }

  const values = [
    source.rewardText,
    source.reward_text,
    source.reward,
    source.pointsText,
    source.points,
    source.experienceText,
    source.experience,
    source.xp,
    meta.rewardText
  ]

  for (let index = 0; index < values.length; index++) {
    const rewardText = normalizeRewardText(values[index])

    if (rewardText) {
      return rewardText
    }
  }

  return ''
}

function pickActionText(source, meta) {
  return source.actionText ||
    source.operationText ||
    source.buttonText ||
    source.ctaText ||
    source.nextActionText ||
    source.statusText ||
    meta.actionText ||
    '去完成'
}

function normalizeNewbieTask(task, index) {
  const source = task && typeof task === 'object' ? task : {}
  const code = String(source.code || source.taskCode || source.type || source.id || '').trim()
  const meta = NEWBIE_TASK_META[code] || NEWBIE_TASK_META[source.type] || {}
  const type = source.type || meta.type || code || 'task'

  return Object.assign({}, meta, source, {
    id: source.id || code || `${type}-${index}`,
    code,
    type,
    rewardText: pickRewardText(source, meta),
    actionText: pickActionText(source, meta)
  })
}

function toNewbieTaskView(task, index) {
  const completed = Boolean(task.completed || task.done || task.finished || task.status === 'completed' || task.status === 'done')
  const actionText = task.actionText || '去完成'

  return Object.assign({}, task, {
    id: task.id || `${task.type || 'task'}-${index}`,
    completed,
    actionText,
    statusText: task.statusText || (completed ? '已完成' : actionText),
    itemClass: completed ? 'done' : '',
    checkClass: completed ? '' : 'muted',
    checkText: completed ? '✓' : '',
    statusClass: completed ? 'newbie-task-status' : 'newbie-task-action'
  })
}

function getNewbieTaskData(summary) {
  const responseTasks = Array.isArray(summary && summary.tasks)
    ? summary.tasks
    : Array.isArray(summary && summary.items)
      ? summary.items
      : Array.isArray(summary && summary.list)
        ? summary.list
        : DEFAULT_NEWBIE_TASKS
  const allTasks = responseTasks.map(normalizeNewbieTask).map(toNewbieTaskView)
  const computedCompletedCount = allTasks.filter((task) => task.completed).length
  const completedCount = pickNumber(summary && summary.completedCount, summary && summary.completed, summary && summary.doneCount)
  const totalCount = pickNumber(summary && summary.totalCount, summary && summary.total)
  const resolvedCompletedCount = completedCount === null ? computedCompletedCount : completedCount
  const resolvedTotalCount = totalCount === null ? allTasks.length : totalCount
  const progressPercent = pickNumber(summary && summary.progressPercent, summary && summary.percent)
  const resolvedProgressPercent = progressPercent === null
    ? resolvedTotalCount ? Math.round((resolvedCompletedCount / resolvedTotalCount) * 100) : 0
    : progressPercent

  return {
    newbieTasks: allTasks,
    newbieCompletedCount: resolvedCompletedCount,
    newbieTotalCount: resolvedTotalCount,
    newbieProgressPercent: Math.max(0, Math.min(100, resolvedProgressPercent))
  }
}

const initialNewbieTaskData = getNewbieTaskData()

Page({
  data: {
    loginMode: 'home',
    accountMode: 'code',
    agreed: false,
    hasWechatLogin: false,
    isLoggingIn: false,
    isSendingCode: false,
    isPhoneLoggingIn: false,
    isPasswordLoggingIn: false,
    isCheckingRealname: false,
    isStartingRealname: false,
    isBindingWechat: false,
    wechatBindPending: false,
    isLoadingNewbieTasks: false,
    newbieTasksLoaded: false,
    newbieTaskLoadFailed: false,
    phone: '',
    maskedPhone: '',
    verifyCode: '',
    codeDigits: ['', '', '', '', '', ''],
    isCodeComplete: false,
    codeInputFocus: false,
    resendSeconds: INITIAL_RESEND_SECONDS,
    canResend: true,
    canRequestCode: false,
    entryFlow: 'register',
    loginNotice: '',
    returnRoute: '',
    password: '',
    inviteCode: '',
    inviteContext: null,
    userInfo: null,
    isUiPreview: false,
    isLoginAuthPreview: false,
    isPostLoginRealnameFlow: false,
    uiPreviewStep: 'invite',
    realnameGuideUrl: '/pages/login/realname/index?ui=1',
    newbieTasks: initialNewbieTaskData.newbieTasks,
    newbieCompletedCount: initialNewbieTaskData.newbieCompletedCount,
    newbieTotalCount: initialNewbieTaskData.newbieTotalCount,
    newbieProgressPercent: initialNewbieTaskData.newbieProgressPercent
  },

  onLoad(options = {}) {
    if (options.mode === 'newbieTasks' && options.auth === '1' && getAuthToken()) {
      this.showAuthenticatedNewbieTasks()
      return
    }
    if (options.walkthrough === 'loginAuth' || options.preview === 'loginAuth') {
      this.enterLoginAuthPreview(options.mode || options.step)
      return
    }

    if (options.ui === '1') {
      this.enterUiPreview(options.mode || 'home', options)
      return
    }

    if (options.normalLogin === '1') {
      this.enterNormalLogin(options.mode || 'home')
      return
    }

    if (options.flow === 'existing') {
      this.enterNormalLogin('home')
      this.setData({
        entryFlow: 'existing',
        loginNotice: options.reason === 'reauth' || options.reason === 'expired' ? '为保障账号安全，请通过短信验证码或密码重新登录。' : '',
        returnRoute: normalizeReturnRoute(options.returnTo)
      })
      return
    }

    if (this.restoreExistingSessionIfNoInvite(options)) {
      return
    }

    const inviteParams = resolveInviteRouteParams(options)
    const inviteCode = inviteParams.inviteCode
    const entryType = inviteParams.entryType
    const inviteContext = inviteCode
      ? inviteService.normalizeInviteContext({
        code: inviteCode,
        entryType
      })
      : inviteService.getInviteContext()

    if (inviteContext && inviteContext.code) {
      inviteService.saveInviteContext(inviteContext)
      this.setData({
        inviteCode: inviteContext.code,
        inviteContext,
        entryFlow: 'register'
      })
    } else {
      this.enterNormalLogin('home')
      return
    }

    if (env.isMock) {
      this.setData(Object.assign({}, getTestLoginDefaults(), {
        inviteCode: inviteContext.code,
        inviteContext
      }))
    }
  },

  onUnload() {
    this.clearCodeTimer()
  },

  restoreExistingSessionIfNoInvite(options = {}) {
    const inviteParams = resolveInviteRouteParams(options)
    if (inviteParams.inviteCode || !getAuthToken()) {
      return false
    }

    wx.reLaunch({
      url: `/${ROUTES.playerHome}`
    })
    return true
  },

  toggleAgreement() {
    this.setData({
      agreed: !this.data.agreed
    })
  },

  openAgreementDetail(key, title) {
    const agreement = encodeURIComponent(key)
    const encodedTitle = encodeURIComponent(title)
    navigateShellRoute(`/pages/profile/system-management/agreement-detail/index?agreement=${agreement}&title=${encodedTitle}&signed=1`)
  },

  openUserAgreement() {
    this.openAgreementDetail('user-service', '用户服务协议')
  },

  openPrivacyAgreement() {
    this.openAgreementDetail('privacy', '隐私政策')
  },

  onPhoneInput(event) {
    const phone = String(event.detail.value || '').trim()
    this.setData({
      phone,
      canRequestCode: this.isValidPhone(phone)
    })
  },

  onCodeInput(event) {
    const code = String(event.detail.value || '').replace(/\D/g, '').slice(0, 6)

    this.setData({
      verifyCode: code,
      codeDigits: this.getCodeDigits(code),
      isCodeComplete: code.length === 6
    })
  },

  onPasswordInput(event) {
    this.setData({
      password: String(event.detail.value || '')
    })
  },

  onInviteCodeInput(event) {
    const inviteCode = this.normalizeInviteCode(event.detail.value)
    const inviteContext = inviteCode
      ? inviteService.normalizeInviteContext(Object.assign({}, this.data.inviteContext || {}, {
        code: inviteCode
      }))
      : null

    if (inviteContext && inviteContext.code) {
      inviteService.saveInviteContext(inviteContext)
    }

    this.setData({
      inviteCode,
      inviteContext
    })
  },

  resolveInviteContext() {
    const currentContext = this.data.inviteContext || {}
    const inviteCode = this.normalizeInviteCode(this.data.inviteCode || currentContext.code)
    const inviteContext = inviteCode
      ? inviteService.normalizeInviteContext(Object.assign({}, currentContext, {
        code: inviteCode
      }))
      : null

    if (!inviteContext || !inviteContext.code) {
      return null
    }

    inviteService.saveInviteContext(inviteContext)
    this.setData({
      inviteCode: inviteContext.code,
      inviteContext
    })

    return inviteContext
  },

  async ensureValidInviteContext(inviteContext) {
    if (!inviteContext || !inviteContext.code) {
      return {
        ok: false,
        message: INVITE_REQUIRED_MESSAGE
      }
    }

    try {
      const result = await inviteService.verifyInviteCode(inviteContext.code, inviteContext.entryType || '')

      if (!result || result.status !== 'valid' || !result.invite) {
        return {
          ok: false,
          message: getInviteErrorMessage(result && result.message) || INVITE_INVALID_MESSAGE,
          expired: Boolean(result && result.expired)
        }
      }

      const normalizedInvite = inviteService.saveInviteContext(Object.assign({}, result.invite, {
        entryType: result.invite.entryType || inviteContext.entryType || ''
      }))

      if (!normalizedInvite) {
        return {
          ok: false,
          message: INVITE_INVALID_MESSAGE
        }
      }

      this.setData({
        inviteCode: normalizedInvite.code,
        inviteContext: normalizedInvite
      })

      return {
        ok: true,
        inviteContext: normalizedInvite
      }
    } catch (error) {
      return {
        ok: false,
        message: getInviteErrorMessage(error) || INVITE_INVALID_MESSAGE,
        expired: isExpiredInviteError(error)
      }
    }
  },

  showInviteError(message) {
    inviteService.clearInviteContext()
    this.setData({
      loginMode: 'home',
      inviteContext: null,
      hasWechatLogin: false
    })
    toast.info(message || INVITE_REQUIRED_MESSAGE)
  },

  showExpiredInviteModal() {
    inviteService.clearInviteContext()
    this.setData({
      loginMode: 'home',
      inviteContext: null,
      hasWechatLogin: false
    })
    wx.showModal({
      title: '邀请码已失效',
      content: '当前邀请码已失效，请重新获取邀请码后再次进入。',
      showCancel: false,
      confirmText: '知道了'
    })
  },

  enterNormalLogin(mode = 'home') {
    const loginMode = LOGIN_WALKTHROUGH_MODES.includes(mode) ? mode : 'home'
    const defaults = getTestLoginDefaults()

    inviteService.clearInviteContext()
    this.clearCodeTimer()
    this.setData(Object.assign({
      isUiPreview: false,
      isLoginAuthPreview: false,
      isPostLoginRealnameFlow: false,
      uiPreviewStep: 'invite',
      realnameGuideUrl: '/pages/login/realname/index?ui=1',
      loginMode,
      accountMode: 'password',
      agreed: false,
      hasWechatLogin: false,
      isLoggingIn: false,
      isSendingCode: false,
      isPhoneLoggingIn: false,
      isPasswordLoggingIn: false,
      isCheckingRealname: false,
      isStartingRealname: false,
      isBindingWechat: false,
      wechatBindPending: false,
      isLoadingNewbieTasks: false,
      newbieTasksLoaded: false,
      newbieTaskLoadFailed: false,
      phone: '',
      maskedPhone: '',
      verifyCode: '',
      codeDigits: this.getCodeDigits(''),
      isCodeComplete: false,
      codeInputFocus: false,
      resendSeconds: INITIAL_RESEND_SECONDS,
      canResend: true,
      canRequestCode: this.isValidPhone(defaults.phone || ''),
      entryFlow: 'register',
      password: '',
      inviteCode: '',
      inviteContext: null
    }, defaults, {
      loginMode,
      inviteCode: '',
      inviteContext: null
    }))
  },

  enterLoginAuthPreview(mode = 'home') {
    const loginMode = LOGIN_WALKTHROUGH_MODES.includes(mode) ? mode : 'home'
    const phone = env.isMock ? TEST_PHONE : ''
    const verifyCode = env.isMock ? TEST_CODE : ''
    const password = env.isMock ? TEST_PASSWORD : ''

    this.clearCodeTimer()
    this.setData({
      isUiPreview: false,
      isLoginAuthPreview: true,
      loginMode,
      accountMode: 'password',
      agreed: true,
      phone,
      maskedPhone: this.maskPhone(phone),
      verifyCode,
      codeDigits: this.getCodeDigits(verifyCode),
      isCodeComplete: verifyCode.length === 6,
      codeInputFocus: false,
      resendSeconds: INITIAL_RESEND_SECONDS,
      canResend: true,
      canRequestCode: this.isValidPhone(phone),
      password,
      hasWechatLogin: false,
      isLoggingIn: false,
      isSendingCode: false,
      isPhoneLoggingIn: false,
      isPasswordLoggingIn: false
    })
  },

  showLoginWalkthroughMode(mode) {
    if (!this.data.isLoginAuthPreview || !LOGIN_WALKTHROUGH_MODES.includes(mode)) {
      return
    }

    const phone = this.data.phone || (env.isMock ? TEST_PHONE : '')
    const verifyCode = this.data.verifyCode || (env.isMock ? TEST_CODE : '')

    this.setData({
      loginMode: mode,
      accountMode: 'password',
      agreed: true,
      phone,
      maskedPhone: this.maskPhone(phone),
      verifyCode,
      codeDigits: this.getCodeDigits(verifyCode),
      isCodeComplete: verifyCode.length === 6,
      codeInputFocus: false,
      password: this.data.password || (env.isMock ? TEST_PASSWORD : '')
    })
  },

  showPreviousLoginWalkthrough() {
    const currentIndex = LOGIN_WALKTHROUGH_MODES.indexOf(this.data.loginMode)
    const nextIndex = Math.max(0, currentIndex - 1)

    this.showLoginWalkthroughMode(LOGIN_WALKTHROUGH_MODES[nextIndex])
  },

  showNextLoginWalkthrough() {
    const currentIndex = LOGIN_WALKTHROUGH_MODES.indexOf(this.data.loginMode)
    const nextIndex = Math.min(LOGIN_WALKTHROUGH_MODES.length - 1, currentIndex + 1)

    this.showLoginWalkthroughMode(LOGIN_WALKTHROUGH_MODES[nextIndex])
  },

  enterUiPreview(mode, options = {}) {
    const stepMap = {
      home: 'invite',
      invite: 'invite',
      realnameModal: 'realnameModal',
      realnameGuide: 'realnameGuide',
      newbieTasks: 'newbieTasks'
    }
    const uiPreviewStep = stepMap[mode] || 'invite'
    const loginMode = 'home'
    const newbieTaskData = getNewbieTaskData()
    const optionInviteCode = this.normalizeInviteCode(options.inviteCode || options.code)

    this.clearCodeTimer()
    this.setData({
      isUiPreview: true,
      isLoginAuthPreview: false,
      isPostLoginRealnameFlow: false,
      uiPreviewStep,
      realnameGuideUrl: '/pages/login/realname/index?ui=1',
      loginMode,
      accountMode: 'password',
      agreed: false,
      phone: env.isMock ? TEST_REGISTER_PHONE : '',
      maskedPhone: '',
      verifyCode: env.isMock ? TEST_CODE : '',
      codeDigits: this.getCodeDigits(env.isMock ? TEST_CODE : ''),
      isCodeComplete: env.isMock,
      codeInputFocus: false,
      resendSeconds: INITIAL_RESEND_SECONDS,
      canResend: true,
      canRequestCode: env.isMock,
      password: env.isMock ? TEST_PASSWORD : '',
      inviteCode: optionInviteCode,
      hasWechatLogin: false,
      isLoggingIn: false,
      isSendingCode: false,
      isPhoneLoggingIn: false,
      isPasswordLoggingIn: false,
      isCheckingRealname: false,
      isStartingRealname: false,
      isLoadingNewbieTasks: false,
      newbieTasksLoaded: false,
      newbieTaskLoadFailed: false,
      newbieTasks: newbieTaskData.newbieTasks,
      newbieCompletedCount: newbieTaskData.newbieCompletedCount,
      newbieTotalCount: newbieTaskData.newbieTotalCount,
      newbieProgressPercent: newbieTaskData.newbieProgressPercent
    })

    if (uiPreviewStep === 'newbieTasks') {
      this.loadNewbieTasks()
    }
  },

  showUiPreviewMode(mode) {
    this.enterUiPreview(mode)
  },

  noop() {},

  handleUiPreviewTap() {
    if (!this.data.isUiPreview) {
      return
    }

    if (this.data.uiPreviewStep === 'invite') {
      return
    }

    if (this.data.uiPreviewStep === 'realnameModal') {
      return
    }

    if (this.data.uiPreviewStep === 'realnameGuide') {
      return
    }
  },

  isRealnameVerified(user) {
    if (!user) {
      return false
    }

    const identity = user.identity || {}
    const nestedUser = user.user || {}
    const status = user.realnameStatus || nestedUser.realnameStatus || identity.status || user.authStatus

    // 手机号验证只证明账号归属，不等同于实名认证。只有后台已经确认实名，
    // 或接口明确告知无需再绑定身份时，才可以跳过实名认证引导。
    return status === 'verified' || status === 'approved' || status === 'passed' || user.needRealname === false
  },

  showRealnameModalAfterLogin() {
    this.clearCodeTimer()
    this.setData({
      isUiPreview: true,
      isLoginAuthPreview: false,
      isPostLoginRealnameFlow: true,
      uiPreviewStep: 'realnameModal',
      realnameGuideUrl: '/pages/login/realname/index',
      loginMode: 'home',
      isCheckingRealname: false,
      isStartingRealname: false,
      isLoadingNewbieTasks: false,
      newbieTasksLoaded: false,
      newbieTaskLoadFailed: false
    })
  },

  showNewbieTasksAfterLogin() {
    const newbieTaskData = getNewbieTaskData()

    this.clearCodeTimer()
    this.setData(Object.assign({
      isUiPreview: true,
      isLoginAuthPreview: false,
      isPostLoginRealnameFlow: true,
      uiPreviewStep: 'newbieTasks',
      realnameGuideUrl: '/pages/login/realname/index',
      loginMode: 'home',
      isCheckingRealname: false,
      isStartingRealname: false,
      isLoadingNewbieTasks: false,
      newbieTasksLoaded: false,
      newbieTaskLoadFailed: false
    }, newbieTaskData))

    this.loadNewbieTasks()
  },

  showAuthenticatedNewbieTasks() {
    const newbieTaskData = getNewbieTaskData()
    this.clearCodeTimer()
    this.setData(Object.assign({
      isUiPreview: true,
      isLoginAuthPreview: false,
      isPostLoginRealnameFlow: true,
      uiPreviewStep: 'newbieTasks',
      realnameGuideUrl: '/pages/login/realname/index',
      loginMode: 'home',
      isCheckingRealname: false,
      isStartingRealname: false,
      isLoadingNewbieTasks: false,
      newbieTasksLoaded: false,
      newbieTaskLoadFailed: false
    }, newbieTaskData))
    this.loadNewbieTasks()
  },

  async continueAfterLogin(loginData = {}) {
    // 登录页既承接新用户邀请码注册，也承接未绑定微信的老用户短信/密码
    // 登录。是否进入新手流程必须以后端本次真实结果为准，不能仅依赖页面
    // 入口，否则老用户从普通登录入口进入时会被反复送回新手任务页。
    if (this.data.entryFlow === 'existing' || loginData.authPageMode === 'login') {
      this.continueExistingAfterLogin()
      return
    }
    if (loginData.requiresIdentityBinding === true) {
      const issued = await this.completePhaseOneSMSIdentity()
      if (issued && issued.token) {
        this.showNewbieTasksAfterLogin()
        return
      }

      this.showRealnameModalAfterLogin()
      return
    }

    // requiresIdentityBinding 表示手机号登录凭证是否还需换发，不代表已经
    // 完成姓名和身份证审核；新注册用户仍需按真实实名状态展示引导。
    if (this.isRealnameVerified(loginData.user)) {
      this.showNewbieTasksAfterLogin()
      return
    }

    try {
      const user = await userService.getCurrentUser()
      if (this.isRealnameVerified(user)) {
        this.showNewbieTasksAfterLogin()
        return
      }
    } catch (error) {
      // Fall through to real-name auth when current-user status cannot be confirmed.
    }

    this.showRealnameModalAfterLogin()
  },

  continueExistingAfterLogin() {
    const target = normalizeReturnRoute(this.data.returnRoute) || `/${ROUTES.playerHome}`
    wx.reLaunch({
      url: target,
      fail: () => wx.redirectTo({ url: target })
    })
  },

  async completePhaseOneSMSIdentity() {
    const phone = String(this.data.phone || '').trim()
    const code = String(this.data.verifyCode || '').trim()

    if (!this.isValidPhone(phone) || code.length !== 6) {
      return null
    }

    try {
      await userService.bindPhone({ phone })
      try {
        await authService.sendPhoneCode({
          phone,
          scene: 'invite_register'
        })
      } catch (error) {
        const message = String(error && error.message || '')
        if (!message.includes('频繁') && !message.includes('已发送')) {
          throw error
        }
      }
      await authService.verifyPhoneCode({
        phone,
        code
      })
      return await authService.issueTokenAfterIdentity()
    } catch (error) {
      toast.info(error.message || '手机号验证失败')
      return null
    }
  },

  async checkRealnameAfterRegister() {
    if (this.data.isCheckingRealname) {
      return
    }

    this.setData({
      isCheckingRealname: true
    })

    try {
      const user = await userService.getCurrentUser()

      if (this.isRealnameVerified(user)) {
        this.showNewbieTasksAfterLogin()
        return
      }

      this.showUiPreviewMode('realnameModal')
    } catch (error) {
      this.showUiPreviewMode('realnameModal')
    } finally {
      this.setData({
        isCheckingRealname: false
      })
    }
  },

  isValidPhone(phone) {
    return /^1\d{10}$/.test(String(phone || '').trim())
  },

  async sendUiPreviewInviteCode() {
    if (this.data.isSendingCode || !this.data.canResend) {
      return
    }

    if (!this.isValidPhone(this.data.phone)) {
      toast.info('请输入正确手机号')
      return
    }

    this.setData({
      isSendingCode: true,
      canResend: false,
      resendSeconds: INITIAL_RESEND_SECONDS
    })

    try {
      await authService.sendPhoneCode({
        phone: this.data.phone,
        scene: this.data.inviteCode ? 'invite_register' : 'login'
      })
      toast.success('验证码已发送')
      this.startCodeTimer()
    } catch (error) {
      toast.info(getSmsSendErrorMessage(error))
      this.setData({
        canResend: true
      })
    } finally {
      this.setData({
        isSendingCode: false
      })
    }
  },

  async handleUiPreviewInviteLogin() {
    const inviteContext = this.resolveInviteContext()
    if (inviteContext) {
      this.setData({
        isUiPreview: false,
        loginMode: 'home',
        agreed: true
      })
      await this.handleInvitePrimaryLogin()
      return
    }

    toast.info(INVITE_REQUIRED_MESSAGE)
  },

  async startRealnameAuth() {
    if (this.data.isUiPreview) {
      if (this.data.uiPreviewStep === 'realnameModal') {
        if (this.data.isPostLoginRealnameFlow) {
          this.setData({
            uiPreviewStep: 'realnameGuide',
            realnameGuideUrl: '/pages/login/realname/index'
          })
        } else {
          this.showUiPreviewMode('realnameGuide')
        }
        return
      }

      navigateShellRoute('/pages/login/realname/index')
      return
    }

    navigateShellRoute('/pages/login/realname/index')
  },

  skipRealnameAuth() {
    if (this.data.isUiPreview) {
      if (this.data.isPostLoginRealnameFlow) {
        this.showNewbieTasksAfterLogin()
      } else {
        this.showUiPreviewMode('newbieTasks')
      }
      return
    }

    wx.reLaunch({
      url: `/${ROUTES.playerHome}`
    })
  },

  async loadNewbieTasks() {
    if (this.data.isLoadingNewbieTasks) {
      return
    }

    this.setData({
      isLoadingNewbieTasks: true,
      newbieTaskLoadFailed: false
    })

    try {
      const summary = await newbieService.getNewbieTasks()
      this.setData(Object.assign(getNewbieTaskData(summary), {
        newbieTasksLoaded: true,
        newbieTaskLoadFailed: false
      }))
    } catch (error) {
      this.setData({
        newbieTasksLoaded: true,
        newbieTaskLoadFailed: true
      })
      toast.info(error.message || '获取新手任务失败')
    } finally {
      this.setData({
        isLoadingNewbieTasks: false
      })
    }
  },

  goHomeFromNewbieTasks() {
    if (this.data.isUiPreview && !this.data.isPostLoginRealnameFlow) {
      navigateShellRoute(`/${ROUTES.playerHome}`)
      return
    }

    wx.reLaunch({
      url: `/${ROUTES.playerHome}`
    })
  },

  goNewbieTask(event) {
    const taskId = event.currentTarget.dataset.id
    const task = this.data.newbieTasks.find((item) => item.id === taskId)

    if (!task || task.completed) {
      return
    }

    if (task.type === 'realname' && this.data.isUiPreview && !this.data.isPostLoginRealnameFlow) {
      this.showUiPreviewMode('realnameGuide')
      return
    }

    if (task.type === 'realname') {
      this.startRealnameAuth()
      return
    }

    if (task.route) {
      navigateShellRoute(task.route)
      return
    }

    if (task.type === 'profile') {
      navigateShellRoute(ROUTES.profileSystemProfileInfo)
      return
    }

    if (task.type === 'role_apply') {
      navigateShellRoute(`/${ROUTES.roleFlow}?mode=roleComparison&single=1&returnTo=${encodeURIComponent(`/${ROUTES.playerHome}`)}`)
      return
    }

    if (task.type === 'first_game') {
      navigateShellRoute(ROUTES.gameCreate)
      return
    }

    if (task.type === 'complete_game') {
      navigateShellRoute(ROUTES.gamePlayerManage)
      return
    }

    if (task.type === 'review') {
      navigateShellRoute(ROUTES.gameReview)
      return
    }

    toast.info('请按任务指引继续')
  },

  async startPhoneLogin() {
    if (this.data.isUiPreview) {
      this.showUiPreviewMode('invite')
      return
    }

    this.setData({ loginMode: 'home' })
  },

  startPasswordLogin() {
    if (this.data.isUiPreview) {
      this.setData({
        isUiPreview: false,
        loginMode: 'account',
        accountMode: 'password',
        agreed: true
      })
      return
    }

    if (!this.data.agreed) {
      toast.info('请先同意用户协议和隐私协议')
      return
    }

    this.setData({
      loginMode: 'account',
      accountMode: 'password'
    })
  },

  switchAccountMode(event) {
    const mode = event.currentTarget.dataset.mode

    if (!mode || mode === this.data.accountMode) {
      return
    }

    this.setData({
      accountMode: mode
    })
  },

  async getVerifyCode() {
    if (this.data.isUiPreview) {
      return
    }

    if (this.data.isSendingCode || !this.data.canResend) {
      return
    }

    if (!this.isValidPhone(this.data.phone)) {
      toast.info('请输入正确手机号')
      return
    }

    await this.sendCodeForPhone(this.data.phone)
  },

  async resendCode() {
    if (this.data.isUiPreview) {
      return
    }

    if (!this.data.canResend || this.data.isSendingCode) {
      return
    }

    if (!this.data.phone) {
      toast.info('请先输入手机号')
      return
    }

    await this.sendCodeForPhone(this.data.phone)
  },

  async sendCodeForPhone(phone) {
    this.clearCodeTimer()

    this.setData({
      isSendingCode: true,
      canResend: false,
      resendSeconds: INITIAL_RESEND_SECONDS,
      verifyCode: env.isMock ? TEST_CODE : '',
      codeDigits: this.getCodeDigits(env.isMock ? TEST_CODE : ''),
      isCodeComplete: env.isMock
    })

    try {
      await authService.sendPhoneCode(phone)
      toast.success('验证码已发送')
      this.startCodeTimer()
    } catch (error) {
      toast.info(getSmsSendErrorMessage(error))
      this.setData({
        canResend: true
      })
    } finally {
      this.setData({
        isSendingCode: false
      })
    }
  },

  startCodeTimer() {
    this.clearCodeTimer()

    this.codeTimer = setInterval(() => {
      const nextSeconds = this.data.resendSeconds - 1

      if (nextSeconds <= 0) {
        this.clearCodeTimer()
        this.setData({
          resendSeconds: INITIAL_RESEND_SECONDS,
          canResend: true
        })
        return
      }

      this.setData({
        resendSeconds: nextSeconds
      })
    }, 1000)
  },

  clearCodeTimer() {
    if (this.codeTimer) {
      clearInterval(this.codeTimer)
      this.codeTimer = null
    }
  },

  focusCodeInput() {
    this.setData({
      codeInputFocus: true
    })
  },

  maskPhone(phone) {
    const value = String(phone || '')

    if (value.length !== 11) {
      return value
    }

    return `${value.slice(0, 3)}****${value.slice(7)}`
  },

  getCodeDigits(code) {
    const chars = String(code || '').split('')
    return Array.from({ length: 6 }, (_, index) => chars[index] || '')
  },

  normalizeInviteCode(value) {
    return inviteService.normalizeInviteCode(value)
  },

  async handleInvitePrimaryLogin() {
    if (this.data.isUiPreview) {
      this.showUiPreviewMode('account')
      return
    }

    const inviteContext = this.resolveInviteContext()

    if (!this.data.agreed) {
      toast.info('请先同意用户协议和隐私协议')
      return
    }

    if (!this.isValidPhone(this.data.phone)) {
      toast.info('请输入正确手机号')
      return
    }

    if (String(this.data.verifyCode || '').trim().length !== 6) {
      toast.info('请输入6位验证码')
      return
    }

    if (this.data.isPhoneLoggingIn) {
      return
    }

    this.setData({
      isPhoneLoggingIn: true
    })

    try {
      let resolvedInvite = inviteContext
      if (inviteContext) {
        const verified = await this.ensureValidInviteContext(inviteContext)
        if (!verified.ok) {
          if (verified.expired) {
            this.showExpiredInviteModal()
            return
          }
          this.showInviteError(verified.message)
          return
        }
        resolvedInvite = verified.inviteContext
      }
      const loginData = await authService.loginByPhone({
        phone: String(this.data.phone || '').trim(),
        code: String(this.data.verifyCode || '').trim(),
        inviteCode: resolvedInvite ? resolvedInvite.code : '',
        entryType: resolvedInvite ? resolvedInvite.entryType || '' : ''
      })
      this.setData({
        userInfo: loginData.user
      })
      inviteService.clearInviteContext()
      toast.success(loginData.authPageMode === 'register' ? '注册成功' : '登录成功')
      const wechatBound = await this.promptWechatBindAfterLogin(loginData)
      this.setData({ wechatBindPending: !wechatBound })
      await this.continueAfterLogin(loginData)
    } catch (error) {
      if (isInviteError(error)) {
        if (isExpiredInviteError(error)) {
          this.showExpiredInviteModal()
          return
        }
        this.showInviteError(getInviteErrorMessage(error))
        return
      }
      toast.info(error.message || '手机号登录/注册失败')
    } finally {
      this.setData({
        isPhoneLoggingIn: false
      })
    }
  },

  async handlePhoneLogin() {
    if (this.data.isUiPreview) {
      this.showUiPreviewMode('account')
      return
    }

    if (!this.data.agreed) {
      toast.info('请先同意用户协议和隐私协议')
      return
    }

    await this.handleInvitePrimaryLogin()
  },

  async handlePasswordLogin() {
    if (this.data.isUiPreview) {
      this.showUiPreviewMode('account')
      return
    }

    if (!this.data.agreed) {
      toast.info('请先同意用户协议和隐私协议')
      return
    }

    if (!this.data.phone || !this.data.password) {
      toast.info('请输入账号和密码')
      return
    }

    if (this.data.isPasswordLoggingIn) {
      return
    }

    this.setData({
      isPasswordLoggingIn: true
    })

    try {
      const loginData = await authService.loginByPassword({
        phone: this.data.phone,
        password: this.data.password
      })

      this.setData({
        userInfo: loginData.user
      })
      inviteService.clearInviteContext()
      toast.success('登录成功')
      const wechatBound = await this.promptWechatBindAfterLogin(loginData)
      this.setData({ wechatBindPending: !wechatBound })
      await this.continueAfterLogin(loginData)
    } catch (error) {
      toast.info(error.message || '账号密码登录失败')
    } finally {
      this.setData({
        isPasswordLoggingIn: false
      })
    }
  },

  async promptWechatBindAfterLogin(loginData) {
    if (loginData && loginData.boundWechat) return true
    return new Promise((resolve) => {
      wx.showModal({
        title: '绑定微信',
        content: '现在绑定后可直接微信快捷登录；也可稍后在“我的-系统设置-账号安全”中绑定。',
        confirmText: '现在绑定',
        cancelText: '暂不绑定',
        success: async (res) => {
          if (!res.confirm) {
            resolve(false)
            return
          }

          this.setData({ isBindingWechat: true })
          try {
            const user = await authService.bindWechatAccount()
            this.setData({
              userInfo: user,
              wechatBindPending: false
            })
            toast.success('微信已绑定')
            resolve(true)
          } catch (error) {
            // Do not treat a failed binding request as "bound". The new-user
            // page retains a visible retry entry so the user can bind later.
            toast.info(error.message || '微信绑定失败，可稍后在新手任务页或系统设置中重试')
            resolve(false)
          } finally {
            this.setData({ isBindingWechat: false })
          }
        },
        fail: () => resolve(false)
      })
    })
  },

  async bindWechatFromNewbieTasks() {
    if (this.data.isBindingWechat) {
      return
    }

    this.setData({ isBindingWechat: true })
    try {
      const user = await authService.bindWechatAccount()
      this.setData({
        userInfo: user,
        wechatBindPending: false
      })
      toast.success('微信已绑定')
    } catch (error) {
      toast.info(error.message || '微信绑定失败，请稍后重试')
    } finally {
      this.setData({ isBindingWechat: false })
    }
  },

  goHome() {
    wx.reLaunch({
      url: `/${ROUTES.gameHall}`
    })
  },

  goEntryForLogin() {
    navigateShellRoute(ROUTES.login)
  },

  goForgot() {
    if (this.data.isUiPreview) {
      return
    }

    navigateShellRoute(ROUTES.loginForgot)
  },

  backToLogin() {
    if (this.data.isUiPreview) {
      this.showUiPreviewMode('home')
      return
    }

    this.clearCodeTimer()
    this.setData({
      loginMode: 'home',
      verifyCode: '',
      codeDigits: this.getCodeDigits(''),
      isCodeComplete: false,
      codeInputFocus: false
    })
  }
})

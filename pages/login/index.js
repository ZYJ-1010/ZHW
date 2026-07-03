const authService = require('../../services/auth')
const inviteService = require('../../services/invite')
const newbieService = require('../../services/newbie')
const userService = require('../../services/user')
const toast = require('../../utils/toast')
const env = require('../../config/env')
const { ROUTES } = require('../../config/routes')
const { navigateShellRoute } = require('../../utils/shell-nav')

const TEST_PHONE = '13888888888'
const TEST_REGISTER_PHONE = '13700000000'
const TEST_CODE = '123456'
const TEST_PASSWORD = 'Test123456'
const LOGIN_WALKTHROUGH_MODES = ['home', 'codeVerify', 'account', 'wechatAuth']

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

function toNewbieTaskView(task, index) {
  const completed = Boolean(task.completed || task.status === 'completed')
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
  const responseTasks = Array.isArray(summary && summary.tasks) ? summary.tasks : []
  const hasResponseTasks = responseTasks.length > 0
  const tasks = (hasResponseTasks ? responseTasks : DEFAULT_NEWBIE_TASKS).map(toNewbieTaskView)
  const computedCompletedCount = tasks.filter((task) => task.completed).length
  const completedCount = hasResponseTasks && typeof summary.completedCount === 'number'
    ? summary.completedCount
    : computedCompletedCount
  const totalCount = hasResponseTasks && typeof summary.totalCount === 'number'
    ? summary.totalCount
    : tasks.length
  const progressPercent = hasResponseTasks && typeof summary.progressPercent === 'number'
    ? summary.progressPercent
    : totalCount ? Math.round((completedCount / totalCount) * 100) : 0

  return {
    newbieTasks: tasks,
    newbieCompletedCount: completedCount,
    newbieTotalCount: totalCount,
    newbieProgressPercent: Math.max(0, Math.min(100, progressPercent))
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
    isLoadingNewbieTasks: false,
    phone: '',
    maskedPhone: '',
    verifyCode: '',
    codeDigits: ['', '', '', '', '', ''],
    isCodeComplete: false,
    codeInputFocus: false,
    resendSeconds: 59,
    canResend: false,
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

    const inviteCode = this.normalizeInviteCode(options.inviteCode || options.code)
    const entryType = String(options.entryType || '').trim()
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
        inviteContext
      })
    } else {
      navigateShellRoute(ROUTES.loginInvite)
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

  toggleAgreement() {
    this.setData({
      agreed: !this.data.agreed
    })
  },

  onPhoneInput(event) {
    this.setData({
      phone: String(event.detail.value || '').trim()
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
      isLoadingNewbieTasks: false,
      phone: '',
      maskedPhone: '',
      verifyCode: '',
      codeDigits: this.getCodeDigits(''),
      isCodeComplete: false,
      codeInputFocus: false,
      resendSeconds: 59,
      canResend: false,
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
      resendSeconds: 60,
      canResend: true,
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
      resendSeconds: 60,
      canResend: true,
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

    return status === 'verified' || status === 'approved' || status === 'passed' || user.needRealname === false || user.requiresIdentityBinding === false
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
      isLoadingNewbieTasks: false
    })
  },

  async continueAfterLogin(loginData = {}) {
    if (loginData.requiresIdentityBinding === true) {
      this.showRealnameModalAfterLogin()
      return
    }

    if (loginData.requiresIdentityBinding === false || this.isRealnameVerified(loginData.user)) {
      wx.reLaunch({
        url: `/${ROUTES.playerHome}`
      })
      return
    }

    try {
      const user = await userService.getCurrentUser()
      if (this.isRealnameVerified(user)) {
        wx.reLaunch({
          url: `/${ROUTES.playerHome}`
        })
        return
      }
    } catch (error) {
      // Fall through to real-name auth when current-user status cannot be confirmed.
    }

    this.showRealnameModalAfterLogin()
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
        this.showUiPreviewMode('newbieTasks')
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
      resendSeconds: 60
    })

    try {
      await authService.sendPhoneCode({
        phone: this.data.phone,
        scene: 'invite_register'
      })
      toast.success('验证码已发送')
      this.startCodeTimer()
    } catch (error) {
      toast.info(error.message || '验证码发送失败')
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
    if (!this.isValidPhone(this.data.phone)) {
      toast.info('请输入正确手机号')
      return
    }

    if (this.data.verifyCode.length !== 6) {
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
      const loginData = await authService.loginByPhone({
        phone: this.data.phone,
        code: this.data.verifyCode,
        inviteCode: this.data.inviteCode
      })

      this.setData({
        userInfo: loginData.user
      })
      inviteService.clearInviteContext()
      toast.success('注册成功')
      await this.checkRealnameAfterRegister()
    } catch (error) {
      toast.info(error.message || '注册失败')
    } finally {
      this.setData({
        isPhoneLoggingIn: false
      })
    }
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

    if (this.data.isStartingRealname) {
      return
    }

    this.setData({
      isStartingRealname: true
    })

    try {
      const result = await userService.startRealnameAuth()
      const url = result && result.url

      if (url) {
        navigateShellRoute(url)
        return
      }

      navigateShellRoute('/pages/login/realname/index')
    } catch (error) {
      toast.info(error.message || '实名认证页面打开失败')
    } finally {
      this.setData({
        isStartingRealname: false
      })
    }
  },

  skipRealnameAuth() {
    if (this.data.isUiPreview) {
      this.showUiPreviewMode('newbieTasks')
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
      isLoadingNewbieTasks: true
    })

    try {
      const summary = await newbieService.getNewbieTasks()
      this.setData(getNewbieTaskData(summary))
    } catch (error) {
      toast.info(error.message || '获取新手任务失败')
    } finally {
      this.setData({
        isLoadingNewbieTasks: false
      })
    }
  },

  goHomeFromNewbieTasks() {
    if (this.data.isUiPreview) {
      navigateShellRoute('/pages/home/index?ui=1&mode=homeAll&single=0')
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

    if (task.type === 'realname' && this.data.isUiPreview) {
      this.showUiPreviewMode('realnameGuide')
      return
    }

    if (task.type === 'realname') {
      navigateShellRoute(ROUTES.roleApply)
      return
    }

    if (task.type === 'profile') {
      navigateShellRoute(ROUTES.profileSystemProfileInfo)
      return
    }

    if (task.type === 'first_game') {
      navigateShellRoute(ROUTES.gameCreate)
      return
    }

    toast.info('请按任务指引继续')
  },

  async startPhoneLogin() {
    if (this.data.isUiPreview) {
      this.showUiPreviewMode('invite')
      return
    }

    if (!this.data.agreed) {
      toast.info('请先同意用户协议和隐私协议')
      return
    }

    const phone = this.data.phone

    if (!phone) {
      toast.info('请先输入手机号')
      return
    }

    this.setData({
      loginMode: 'codeVerify',
      phone,
      maskedPhone: this.maskPhone(phone),
      verifyCode: env.isMock ? TEST_CODE : '',
      codeDigits: this.getCodeDigits(env.isMock ? TEST_CODE : ''),
      isCodeComplete: env.isMock,
      codeInputFocus: true
    })

    await this.sendCodeForPhone(phone)
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

    if (this.data.isSendingCode) {
      return
    }

    if (!this.data.phone) {
      toast.info('请先输入手机号')
      return
    }

    this.setData({
      isSendingCode: true
    })

    try {
      await authService.sendPhoneCode(this.data.phone)
      toast.success('验证码已发送')
    } catch (error) {
      toast.info(error.message || '验证码发送失败')
    } finally {
      this.setData({
        isSendingCode: false
      })
    }
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
      resendSeconds: 59,
      verifyCode: env.isMock ? TEST_CODE : '',
      codeDigits: this.getCodeDigits(env.isMock ? TEST_CODE : ''),
      isCodeComplete: env.isMock
    })

    try {
      await authService.sendPhoneCode(phone)
      toast.success('验证码已发送')
      this.startCodeTimer()
    } catch (error) {
      toast.info(error.message || '验证码发送失败')
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
          resendSeconds: 60,
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

  async handlePhoneLogin() {
    if (this.data.isUiPreview) {
      this.showUiPreviewMode('account')
      return
    }

    if (!this.data.agreed) {
      toast.info('请先同意用户协议和隐私协议')
      return
    }

    if (!this.data.phone || this.data.verifyCode.length !== 6) {
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
      const loginData = await authService.loginByPhone({
        phone: this.data.phone,
        code: this.data.verifyCode,
        inviteCode: this.data.inviteContext ? this.data.inviteContext.code : ''
      })

      this.setData({
        userInfo: loginData.user
      })
      inviteService.clearInviteContext()
      this.clearCodeTimer()
      toast.success('登录成功')
    } catch (error) {
      toast.info(error.message || '手机号登录失败')
    } finally {
      this.setData({
        isPhoneLoggingIn: false
      })
    }
  },

  async handlePasswordLogin() {
    if (this.data.isUiPreview) {
      this.showUiPreviewMode('wechatAuth')
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
    } catch (error) {
      toast.info(error.message || '账号密码登录失败')
    } finally {
      this.setData({
        isPasswordLoggingIn: false
      })
    }
  },

  startWechatAuth() {
    if (this.data.isUiPreview) {
      this.setData({
        isUiPreview: false,
        loginMode: 'wechatAuth',
        agreed: true
      })
      return
    }

    if (!this.data.agreed) {
      toast.info('请先同意用户协议和隐私协议')
      return
    }

    const inviteContext = this.resolveInviteContext()
    if (!inviteContext) {
      toast.info('请先输入邀请码')
      return
    }

    this.setData({
      loginMode: 'wechatAuth'
    })
  },

  async handleWechatLogin() {
    if (this.data.isUiPreview) {
      this.showUiPreviewMode('newbieTasks')
      return
    }

    if (!this.data.agreed) {
      toast.info('请先同意用户协议和隐私协议')
      return
    }

    const inviteContext = this.resolveInviteContext()
    if (!inviteContext) {
      toast.info('请先输入邀请码')
      return
    }

    if (this.data.isLoggingIn || this.data.hasWechatLogin) {
      return
    }

    this.setData({
      isLoggingIn: true
    })

    try {
      const loginData = await authService.loginByWechat({
        inviteCode: inviteContext.code
      })

      this.setData({
        hasWechatLogin: true,
        userInfo: loginData.user
      })
      inviteService.clearInviteContext()
      toast.success('登录成功')
      await this.continueAfterLogin(loginData)
    } catch (error) {
      this.setData({
        loginMode: 'wechatAuth'
      })
      toast.info(error.message || '授权失败')
    } finally {
      this.setData({
        isLoggingIn: false
      })
    }
  },

  goHome() {
    wx.reLaunch({
      url: `/${ROUTES.gameHall}`
    })
  },

  goEntryForLogin() {
    navigateShellRoute(ROUTES.loginInvite)
  },

  goForgot() {
    if (this.data.isUiPreview) {
      return
    }

    navigateShellRoute(ROUTES.loginForgot)
  },

  clearInvite() {
    inviteService.clearInviteContext()
    this.setData({
      inviteContext: null
    })
    toast.info('已切换为普通登录')
  },

  declineAuth() {
    if (this.data.isUiPreview) {
      this.showUiPreviewMode('invite')
      return
    }

    this.setData({
      loginMode: 'home',
      agreed: false
    })
    toast.info('已拒绝授权，可稍后再允许')
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

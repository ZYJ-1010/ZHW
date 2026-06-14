const authService = require('../../services/auth')
const inviteService = require('../../services/invite')
const userService = require('../../services/user')
const newbieService = require('../../services/newbie')
const toast = require('../../utils/toast')
const { ROUTES } = require('../../config/routes')

const DEFAULT_NEWBIE_TASKS = [
  {
    id: 'newbie-realname',
    type: 'realname',
    title: '完成实名认证',
    rewardText: '+50 经验值',
    completed: true,
    actionText: '去完成'
  },
  {
    id: 'newbie-profile',
    type: 'profile',
    title: '完善个人资料',
    rewardText: '+30 经验值',
    completed: false,
    actionText: '去完成'
  },
  {
    id: 'newbie-first-game',
    type: 'first_game',
    title: '发布第一个局',
    rewardText: '+100 经验值',
    completed: false,
    actionText: '去完成'
  }
]

function getNewbieTaskRoute(type) {
  if (type === 'profile') {
    return ROUTES.profile
  }

  if (type === 'first_game') {
    return ROUTES.gameCreate
  }

  if (type === 'realname') {
    return `${ROUTES.login}?ui=1&mode=realnameGuide`
  }

  return ''
}

function toNewbieTaskView(task, index) {
  const completed = Boolean(task.completed || task.status === 'completed')
  const actionText = task.actionText || '去完成'
  const route = task.route || getNewbieTaskRoute(task.type)

  return Object.assign({}, task, {
    id: task.id || `${task.type || 'task'}-${index}`,
    completed,
    route,
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
    uiPreviewStep: 'invite',
    newbieTasks: initialNewbieTaskData.newbieTasks,
    newbieCompletedCount: initialNewbieTaskData.newbieCompletedCount,
    newbieTotalCount: initialNewbieTaskData.newbieTotalCount,
    newbieProgressPercent: initialNewbieTaskData.newbieProgressPercent
  },

  onLoad(options) {
    if (options.ui === '1') {
      this.enterUiPreview(options.mode || 'home')
      return
    }

    const inviteContext = inviteService.getInviteContext()
    const inviteCode = String(options.inviteCode || '').trim().toUpperCase()

    if (inviteContext && inviteContext.code) {
      this.setData({
        inviteContext
      })
      return
    }

    if (inviteCode) {
      this.setData({
        inviteContext: {
          code: inviteCode
        }
      })
    }
  },

  onUnload() {
    this.clearCodeTimer()
    this.clearEntryTimer()
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
    this.setData({
      inviteCode: String(event.detail.value || '').trim().toUpperCase()
    })
  },

  enterUiPreview(mode) {
    const stepMap = {
      home: 'invite',
      invite: 'invite',
      entry: 'entry',
      realnameModal: 'realnameModal',
      realnameGuide: 'realnameGuide',
      newbieTasks: 'newbieTasks'
    }
    const uiPreviewStep = stepMap[mode] || 'invite'
    const loginMode = 'home'
    const newbieTaskData = getNewbieTaskData()

    this.clearCodeTimer()
    this.clearEntryTimer()
    this.setData({
      isUiPreview: true,
      uiPreviewStep,
      loginMode,
      accountMode: 'password',
      agreed: false,
      phone: '',
      maskedPhone: '',
      verifyCode: '',
      codeDigits: this.getCodeDigits(''),
      isCodeComplete: false,
      codeInputFocus: false,
      resendSeconds: 60,
      canResend: true,
      password: '',
      inviteCode: '',
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

    if (uiPreviewStep === 'entry') {
      this.startEntryAutoTimer()
    }

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

    if (this.data.uiPreviewStep === 'entry') {
      this.checkRealnameAfterEntry()
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

    const status = user.realnameStatus || user.authStatus

    return status === 'verified' || status === 'approved' || status === 'passed' || user.needRealname === false
  },

  async checkRealnameAfterEntry() {
    if (this.data.isCheckingRealname) {
      return
    }

    this.setData({
      isCheckingRealname: true
    })

    try {
      const user = await userService.getCurrentUser()

      if (this.isRealnameVerified(user)) {
        wx.redirectTo({
          url: `/${ROUTES.profile}`
        })
        return
      }

      this.showUiPreviewMode('realnameModal')
    } catch (error) {
      toast.info(error.message || '获取实名状态失败')
    } finally {
      this.setData({
        isCheckingRealname: false
      })
    }
  },

  startEntryAutoTimer() {
    this.clearEntryTimer()
    this.entryTimer = setTimeout(() => {
      this.entryTimer = null
      this.checkRealnameAfterEntry()
    }, 900)
  },

  clearEntryTimer() {
    if (this.entryTimer) {
      clearTimeout(this.entryTimer)
      this.entryTimer = null
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
      toast.success('验证码已发送：123456')
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
      toast.success('登录成功')
      this.showUiPreviewMode('entry')
    } catch (error) {
      toast.info(error.message || '登录或注册失败')
    } finally {
      this.setData({
        isPhoneLoggingIn: false
      })
    }
  },

  goRealnameGuide() {
    if (!this.data.isUiPreview) {
      return
    }

    this.showUiPreviewMode('realnameGuide')
  },

  async startRealnameAuth() {
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
        wx.navigateTo({
          url
        })
        return
      }

      toast.info('实名认证页面暂未配置')
    } catch (error) {
      toast.info(error.message || '实名认证页面打开失败')
    } finally {
      this.setData({
        isStartingRealname: false
      })
    }
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
    wx.redirectTo({
      url: `/${ROUTES.home}`
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

    const route = task.route || getNewbieTaskRoute(task.type)

    if (!route) {
      toast.info('任务页面暂未配置')
      return
    }

    wx.navigateTo({
      url: route.charAt(0) === '/' ? route : `/${route}`
    })
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

    const phone = this.data.phone || '13888888888'

    this.setData({
      loginMode: 'codeVerify',
      phone,
      maskedPhone: this.maskPhone(phone),
      verifyCode: '',
      codeDigits: this.getCodeDigits(''),
      isCodeComplete: false,
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
      toast.success('验证码已发送：123456')
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

    await this.sendCodeForPhone(this.data.phone || '13888888888')
  },

  async sendCodeForPhone(phone) {
    this.clearCodeTimer()

    this.setData({
      isSendingCode: true,
      canResend: false,
      resendSeconds: 59,
      verifyCode: '',
      codeDigits: this.getCodeDigits(''),
      isCodeComplete: false
    })

    try {
      await authService.sendPhoneCode(phone)
      toast.success('验证码已发送：123456')
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
      this.goHome()
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
      this.goHome()
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

    this.setData({
      loginMode: 'wechatAuth'
    })
  },

  async handleWechatLogin() {
    if (this.data.isUiPreview) {
      this.showUiPreviewMode('entry')
      return
    }

    if (!this.data.agreed) {
      toast.info('请先同意用户协议和隐私协议')
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
        inviteCode: this.data.inviteContext ? this.data.inviteContext.code : ''
      })

      this.setData({
        hasWechatLogin: true,
        userInfo: loginData.user
      })
      inviteService.clearInviteContext()
      this.goEntryForLogin()
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
    setTimeout(() => {
      wx.redirectTo({
        url: `/${ROUTES.home}`
      })
    }, 500)
  },

  goEntryForLogin() {
    setTimeout(() => {
      wx.redirectTo({
        url: `/${ROUTES.entry}?mode=login`
      })
    }, 300)
  },

  goForgot() {
    if (this.data.isUiPreview) {
      return
    }

    wx.navigateTo({
      url: `/${ROUTES.loginForgot}`
    })
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
      this.showUiPreviewMode('account')
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

const authService = require('../../services/auth')
const inviteService = require('../../services/invite')
const toast = require('../../utils/toast')
const { ROUTES } = require('../../config/routes')

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
    phone: '',
    maskedPhone: '',
    verifyCode: '',
    codeDigits: ['', '', '', '', '', ''],
    isCodeComplete: false,
    codeInputFocus: false,
    resendSeconds: 59,
    canResend: false,
    password: '',
    inviteContext: null,
    userInfo: null,
    isUiPreview: false,
    uiPreviewStep: 'invite'
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
    const previewPhone = '13888888888'

    this.clearCodeTimer()
    this.setData({
      isUiPreview: true,
      uiPreviewStep,
      loginMode,
      accountMode: 'password',
      agreed: false,
      phone: previewPhone,
      maskedPhone: this.maskPhone(previewPhone),
      verifyCode: '',
      codeDigits: this.getCodeDigits(''),
      isCodeComplete: false,
      codeInputFocus: false,
      resendSeconds: 59,
      canResend: false,
      password: '',
      hasWechatLogin: false,
      isLoggingIn: false,
      isSendingCode: false,
      isPhoneLoggingIn: false,
      isPasswordLoggingIn: false
    })
  },

  showUiPreviewMode(mode) {
    this.enterUiPreview(mode)
  },

  handleUiPreviewTap() {
    if (!this.data.isUiPreview) {
      return
    }

    if (this.data.uiPreviewStep === 'invite') {
      this.showUiPreviewMode('entry')
      return
    }

    if (this.data.uiPreviewStep === 'entry') {
      this.showUiPreviewMode('realnameModal')
      return
    }

    if (this.data.uiPreviewStep === 'realnameModal') {
      this.showUiPreviewMode('realnameGuide')
      return
    }

    if (this.data.uiPreviewStep === 'realnameGuide') {
      this.showUiPreviewMode('newbieTasks')
    }
  },

  async startPhoneLogin() {
    if (this.data.isUiPreview) {
      this.showUiPreviewMode('codeVerify')
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
      this.showUiPreviewMode('account')
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
          resendSeconds: 0,
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
      this.showUiPreviewMode('wechatAuth')
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
      wx.redirectTo({
        url: `/${ROUTES.entry}?ui=1&mode=login`
      })
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

const authService = require('../../../services/auth')
const toast = require('../../../utils/toast')
const { ROUTES } = require('../../../config/routes')

Page({
  data: {
    step: 'phone',
    phone: '',
    verifyCode: '',
    password: '',
    confirmPassword: '',
    isSendingCode: false,
    isCheckingCode: false,
    isSubmitting: false
  },

  onPhoneInput(event) {
    this.setData({
      phone: String(event.detail.value || '').trim()
    })
  },

  onCodeInput(event) {
    this.setData({
      verifyCode: String(event.detail.value || '').replace(/\D/g, '').slice(0, 6)
    })
  },

  onPasswordInput(event) {
    this.setData({
      password: String(event.detail.value || '')
    })
  },

  onConfirmPasswordInput(event) {
    this.setData({
      confirmPassword: String(event.detail.value || '')
    })
  },

  async getVerifyCode() {
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

  async goPasswordStep() {
    if (!this.data.phone || this.data.verifyCode.length !== 6) {
      toast.info('请输入手机号和6位验证码')
      return
    }

    if (this.data.isCheckingCode) {
      return
    }

    this.setData({
      isCheckingCode: true
    })

    try {
      await authService.verifyPhoneCode({
        phone: this.data.phone,
        code: this.data.verifyCode
      })
      this.setData({
        step: 'password'
      })
    } catch (error) {
      toast.info(error.message || '手机号或验证码错误')
    } finally {
      this.setData({
        isCheckingCode: false
      })
    }
  },

  goPhoneStep() {
    this.setData({
      step: 'phone'
    })
  },

  async submitReset() {
    if (!this.data.password || !this.data.confirmPassword) {
      toast.info('请输入并确认新密码')
      return
    }

    if (!this.isValidPassword(this.data.password)) {
      toast.info('仅支持字母和数字，长度8-20位')
      return
    }

    if (this.data.password !== this.data.confirmPassword) {
      toast.info('两次输入的新密码不一致')
      return
    }

    if (this.data.isSubmitting) {
      return
    }

    this.setData({
      isSubmitting: true
    })

    try {
      await authService.resetPassword({
        phone: this.data.phone,
        code: this.data.verifyCode,
        password: this.data.password
      })
      this.setData({
        step: 'success'
      })
    } catch (error) {
      toast.info(error.message || '密码找回失败')
    } finally {
      this.setData({
        isSubmitting: false
      })
    }
  },

  isValidPassword(password) {
    return /^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d]{8,20}$/.test(password)
  },

  backToLogin() {
    const pages = getCurrentPages()

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    wx.redirectTo({
      url: `/${ROUTES.login}`
    })
  }
})

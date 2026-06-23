const authService = require('../../../services/auth')
const toast = require('../../../utils/toast')
const env = require('../../../config/env')

const TEST_PHONE = '13888888888'
const TEST_CODE = '123456'
const TEST_PASSWORD = 'Test123456'
const FORGOT_WALKTHROUGH_STEPS = ['phone', 'password', 'success']

Page({
  data: {
    step: 'phone',
    phone: '',
    verifyCode: '',
    password: '',
    confirmPassword: '',
    isUiPreview: false,
    isSendingCode: false,
    isCheckingCode: false,
    isSubmitting: false
  },

  onLoad(options = {}) {
    const previewStepMap = {
      phone: 'phone',
      password: 'password',
      success: 'success'
    }

    if (!env.isMock) {
      if (options.ui === '1') {
        this.setData({
          isUiPreview: true,
          step: previewStepMap[options.step] || 'phone',
          phone: TEST_PHONE,
          verifyCode: TEST_CODE,
          password: TEST_PASSWORD,
          confirmPassword: TEST_PASSWORD
        })
      }

      return
    }

    this.setData({
      isUiPreview: options.ui === '1',
      step: previewStepMap[options.step] || 'phone',
      phone: TEST_PHONE,
      verifyCode: TEST_CODE,
      password: TEST_PASSWORD,
      confirmPassword: TEST_PASSWORD
    })
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

  showForgotWalkthroughStep(step) {
    if (!this.data.isUiPreview || !FORGOT_WALKTHROUGH_STEPS.includes(step)) {
      return
    }

    this.setData({
      step
    })
  },

  showPreviousForgotWalkthrough() {
    const currentIndex = FORGOT_WALKTHROUGH_STEPS.indexOf(this.data.step)
    const nextIndex = Math.max(0, currentIndex - 1)

    this.showForgotWalkthroughStep(FORGOT_WALKTHROUGH_STEPS[nextIndex])
  },

  showNextForgotWalkthrough() {
    const currentIndex = FORGOT_WALKTHROUGH_STEPS.indexOf(this.data.step)
    const nextIndex = Math.min(FORGOT_WALKTHROUGH_STEPS.length - 1, currentIndex + 1)

    this.showForgotWalkthroughStep(FORGOT_WALKTHROUGH_STEPS[nextIndex])
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
    toast.developing()
  }
})

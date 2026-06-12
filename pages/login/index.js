const authService = require('../../services/auth')
const inviteService = require('../../services/invite')
const toast = require('../../utils/toast')
const { ROUTES } = require('../../config/routes')
const { INVITE_STATUS_TEXT, INVITE_TIP } = require('../../config/constants')

Page({
  data: {
    agreed: false,
    hasWechatLogin: false,
    isLoggingIn: false,
    isVerifyingInvite: false,
    inviteCode: '',
    inviteStatus: 'idle',
    inviteStatusText: INVITE_STATUS_TEXT.idle,
    inviteTip: INVITE_TIP.idle,
    inviteInfo: null,
    userInfo: null
  },

  toggleAgreement() {
    this.setData({
      agreed: !this.data.agreed
    })
  },

  async handleWechatLogin() {
    if (!this.data.agreed) {
      toast.info('请先勾选同意')
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
        inviteCode: this.data.inviteCode
      })

      this.setData({
        hasWechatLogin: true,
        userInfo: loginData.user
      })
      toast.success('授权成功')
      this.goHome()
    } catch (error) {
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

  declineAuth() {
    this.setData({
      agreed: false
    })
    toast.info('已暂不授权')
  },

  onInviteInput(event) {
    this.setData({
      inviteCode: String(event.detail.value || '').trim().toUpperCase(),
      inviteStatus: 'idle',
      inviteStatusText: INVITE_STATUS_TEXT.idle,
      inviteTip: INVITE_TIP.idle,
      inviteInfo: null
    })
  },

  async verifyInvite() {
    if (this.data.isVerifyingInvite) {
      return
    }

    this.setData({
      isVerifyingInvite: true
    })

    try {
      const result = await inviteService.verifyInviteCode(this.data.inviteCode)

      this.setData({
        inviteStatus: result.status,
        inviteStatusText: INVITE_STATUS_TEXT[result.status],
        inviteTip: INVITE_TIP[result.status],
        inviteInfo: result.invite
      })

      if (result.status === 'valid') {
        toast.success(result.message)
      } else {
        toast.info(result.message)
      }
    } catch (error) {
      toast.info(error.message || '邀请码验证失败')
    } finally {
      this.setData({
        isVerifyingInvite: false
      })
    }
  }
})

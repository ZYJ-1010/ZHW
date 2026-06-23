const inviteService = require('../../../services/invite')
const toast = require('../../../utils/toast')
const { INVITE_STATUS_TEXT, INVITE_TIP } = require('../../../config/constants')
const { ROUTES } = require('../../../config/routes')
const env = require('../../../config/env')

const TEST_INVITE_CODE = 'ENJOY2026'

function getInviteCodeFromOptions(options) {
  const directCode = String(options.code || options.inviteCode || '').trim()

  if (directCode) {
    return directCode.toUpperCase()
  }

  const scene = decodeURIComponent(String(options.scene || '')).trim()

  if (!scene) {
    return ''
  }

  const params = scene.split('&').reduce((result, item) => {
    const pair = item.split('=')
    result[pair[0]] = pair[1] || ''
    return result
  }, {})

  return String(params.inviteCode || params.code || scene).trim().toUpperCase()
}

Page({
  data: {
    inviteCode: '',
    inviteStatus: 'idle',
    statusText: INVITE_STATUS_TEXT.idle,
    statusMessage: INVITE_TIP.idle,
    invite: null,
    isVerifying: false
  },

  onLoad(options) {
    const inviteCode = getInviteCodeFromOptions(options) || (env.isMock ? TEST_INVITE_CODE : '')

    if (!inviteCode) {
      return
    }

    this.setData({
      inviteCode
    })
    this.verifyInvite()
  },

  onInviteInput(event) {
    this.setData({
      inviteCode: String(event.detail.value || '').trim().toUpperCase(),
      inviteStatus: 'idle',
      statusText: INVITE_STATUS_TEXT.idle,
      statusMessage: INVITE_TIP.idle,
      invite: null
    })
  },

  async verifyInvite() {
    if (this.data.isVerifying) {
      return
    }

    if (!this.data.inviteCode) {
      toast.info('请输入邀请码')
      return
    }

    this.setData({
      isVerifying: true
    })

    try {
      const result = await inviteService.verifyInviteCode(this.data.inviteCode)

      this.setData({
        inviteStatus: result.status,
        statusText: INVITE_STATUS_TEXT[result.status],
        statusMessage: result.message || INVITE_TIP[result.status],
        invite: result.invite
      })

      if (result.status === 'valid') {
        inviteService.saveInviteContext(result.invite)
        toast.success('邀请码已确认')
      }
    } catch (error) {
      this.setData({
        inviteStatus: 'invalid',
        statusText: INVITE_STATUS_TEXT.invalid,
        statusMessage: error.message || INVITE_TIP.invalid,
        invite: null
      })
      toast.info(error.message || '邀请码无效')
    } finally {
      this.setData({
        isVerifying: false
      })
    }
  },

  continueToLogin() {
    if (this.data.inviteStatus !== 'valid' || !this.data.invite) {
      toast.info('请先确认有效邀请码')
      return
    }

    inviteService.saveInviteContext(this.data.invite)
    wx.redirectTo({
      url: `/${ROUTES.login}?ui=1&mode=invite&inviteCode=${this.data.invite.code}`
    })
  },

  goNormalLogin() {
    inviteService.clearInviteContext()
    wx.redirectTo({
      url: `/${ROUTES.login}`
    })
  }
})

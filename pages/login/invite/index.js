const inviteService = require('../../../services/invite')
const toast = require('../../../utils/toast')
const { INVITE_STATUS_TEXT, INVITE_TIP } = require('../../../config/constants')
const { ROUTES } = require('../../../config/routes')
const { navigateShellRoute } = require('../../../utils/shell-nav')

function getInviteCodeFromOptions(options) {
  const directCode = inviteService.normalizeInviteCode(options.code || options.inviteCode)

  if (directCode) {
    return directCode
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

  return inviteService.normalizeInviteCode(params.inviteCode || params.code || scene)
}

function getEntryTypeFromOptions(options) {
  const scene = decodeURIComponent(String(options.scene || '')).trim()
  const sceneParams = scene
    ? scene.split('&').reduce((result, item) => {
      const pair = item.split('=')
      result[pair[0]] = pair[1] || ''
      return result
    }, {})
    : {}

  return String(options.entryType || sceneParams.entryType || '').trim()
}

Page({
  data: {
    inviteCode: '',
    inviteStatus: 'idle',
    statusText: INVITE_STATUS_TEXT.idle,
    statusMessage: INVITE_TIP.idle,
    invite: null,
    entryType: '',
    isVerifying: false
  },

  onLoad(options) {
    const inviteCode = getInviteCodeFromOptions(options || {})
    const entryType = getEntryTypeFromOptions(options || {})

    if (!inviteCode) {
      return
    }

    this.setData({
      inviteCode,
      entryType
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
      const invite = result.invite
      const entryType = invite && invite.entryType ? invite.entryType : this.data.entryType

      this.setData({
        inviteStatus: result.status,
        statusText: INVITE_STATUS_TEXT[result.status],
        statusMessage: result.message || INVITE_TIP[result.status],
        invite,
        entryType
      })

      if (result.status === 'valid') {
        inviteService.saveInviteContext(Object.assign({}, invite, {
          entryType
        }))
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

    const invite = inviteService.saveInviteContext(Object.assign({}, this.data.invite, {
      entryType: this.data.entryType
    }))

    if (!invite || !invite.code) {
      toast.info('邀请码信息异常，请重新确认')
      return
    }

    const query = [
      'ui=1',
      'mode=invite',
      `inviteCode=${encodeURIComponent(invite.code)}`
    ]

    if (invite.entryType) {
      query.push(`entryType=${encodeURIComponent(invite.entryType)}`)
    }

    navigateShellRoute(`/${ROUTES.login}?${query.join('&')}`)
  },

  goNormalLogin() {
    inviteService.clearInviteContext()
    navigateShellRoute(ROUTES.login)
  }
})

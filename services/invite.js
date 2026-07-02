const api = require('../api/request')

async function verifyInviteCode(code) {
  const normalizedCode = String(code || '').trim().toUpperCase()

  if (!normalizedCode) {
    return {
      status: 'idle',
      invite: null,
      message: '邀请码可稍后填写'
    }
  }

  const result = await api.verifyInvite(normalizedCode)

  if (result.code !== 0) {
    return {
      status: 'invalid',
      invite: null,
      message: result.message || '邀请码无效'
    }
  }

  return {
    status: 'valid',
    invite: result.data,
    message: '邀请码已确认'
  }
}

function saveInviteContext(invite) {
  if (!invite || !invite.code) {
    return
  }

  wx.setStorageSync('enjoy_invite_context', invite)
}

function getInviteContext() {
  return wx.getStorageSync('enjoy_invite_context') || null
}

function clearInviteContext() {
  wx.removeStorageSync('enjoy_invite_context')
}

module.exports = {
  verifyInviteCode,
  saveInviteContext,
  getInviteContext,
  clearInviteContext
}

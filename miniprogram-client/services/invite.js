const api = require('../api/request')

const EMPTY_INVITE_CODE_VALUES = ['', 'UNDEFINED', 'NULL']

function normalizeInviteCode(value) {
  const code = String(value || '').trim().toUpperCase()

  return EMPTY_INVITE_CODE_VALUES.includes(code) ? '' : code
}

function normalizeInviteContext(source, fallbackCode = '') {
  if (!source || typeof source !== 'object') {
    return null
  }

  const rawInvite = source.inviteCode && typeof source.inviteCode === 'object'
    ? source.inviteCode
    : source.invite && typeof source.invite === 'object'
      ? source.invite
      : source
  const invite = Object.assign({}, rawInvite)
  const code = normalizeInviteCode(invite.code || invite.inviteCode || fallbackCode)

  if (!code) {
    return null
  }

  const entryType = String(source.entryType || invite.entryType || '').trim()

  return Object.assign({}, invite, {
    code,
    entryType,
    authPageMode: source.authPageMode || invite.authPageMode || '',
    boundWechat: Boolean(source.boundWechat || invite.boundWechat),
    boundUserId: source.boundUserId || invite.boundUserId || invite.boundWechatUserId || 0
  })
}

async function verifyInviteCode(code, entryType = '') {
  const normalizedCode = normalizeInviteCode(code)

  if (!normalizedCode) {
    return {
      status: 'idle',
      invite: null,
      message: '邀请码可稍后填写'
    }
  }

  const result = await api.verifyInvite(normalizedCode, String(entryType || '').trim())

  if (result.code !== 0) {
    return {
      status: 'invalid',
      invite: null,
      message: result.message || '邀请码无效'
    }
  }

  const invite = normalizeInviteContext(result.data, normalizedCode)

  if (!invite) {
    return {
      status: 'invalid',
      invite: null,
      message: '邀请码返回格式异常'
    }
  }

  return {
    status: 'valid',
    invite,
    precheck: result.data,
    message: '邀请码已确认'
  }
}

function saveInviteContext(invite) {
  const normalizedInvite = normalizeInviteContext(invite)

  if (!normalizedInvite) {
    return null
  }

  wx.setStorageSync('enjoy_invite_context', normalizedInvite)
  return normalizedInvite
}

function getInviteContext() {
  return normalizeInviteContext(wx.getStorageSync('enjoy_invite_context') || null)
}

function clearInviteContext() {
  wx.removeStorageSync('enjoy_invite_context')
}

module.exports = {
  verifyInviteCode,
  saveInviteContext,
  getInviteContext,
  clearInviteContext,
  normalizeInviteCode,
  normalizeInviteContext
}

const gameService = require('../../../services/game')
const { ROUTES } = require('../../../config/routes')
const { getSurnameInitials } = require('../../../utils/avatar')

const DEFAULT_REPLAY_CONTEXT = {
  sourceGameId: 'game-replay-001',
  serviceOrderId: 'SO-20260613-001',
  inviter: {
    id: 'guide-wang',
    name: '王引荐',
    roleType: 'guide',
    roleLabel: '领路人'
  },
  previousSession: {
    serviceType: '产品架构咨询',
    completedAtText: '2026-06-13 14:30',
    participantText: '3人（行家+玩家+领路人）'
  },
  invitees: [
    {
      id: 'expert-zhang',
      name: '张专家',
      roleType: 'expert',
      roleLabel: '行家',
      desc: '产品架构咨询'
    },
    {
      id: 'player-wang',
      name: '王总',
      roleType: 'player',
      roleLabel: '玩家',
      desc: '需求方'
    }
  ]
}

const QUICK_MESSAGES = [
  '再来一局？',
  '上次合作很愉快，继续！',
  '有个新需求想聊聊',
  '有空再约一局'
]

const ROLE_CLASS_MAP = {
  expert: 'blue',
  player: 'pink',
  guide: 'orange'
}

function getRoleClass(roleType) {
  return ROLE_CLASS_MAP[roleType] || 'blue'
}

function normalizePreviousSession(session = {}) {
  return {
    serviceType: session.serviceType || session.serviceName || session.title || '产品架构咨询',
    completedAtText: session.completedAtText || session.completedAt || session.finishTimeText || '2026-06-13 14:30',
    participantText: session.participantText || session.membersText || '3人（行家+玩家+领路人）'
  }
}

function normalizeInvitee(invitee = {}, index = 0) {
  const roleType = invitee.roleType || invitee.role || ['expert', 'player'][index] || 'player'
  const roleLabel = invitee.roleLabel || invitee.roleName || (roleType === 'expert' ? '行家' : roleType === 'guide' ? '领路人' : '玩家')

  return {
    id: String(invitee.id || invitee.userId || invitee.memberId || `${roleType}-${index}`),
    name: invitee.name || invitee.nickname || roleLabel,
    avatarText: invitee.avatarText || getSurnameInitials(invitee.name || invitee.nickname || '', roleLabel.slice(0, 2)),
    avatarClass: invitee.avatarClass || getRoleClass(roleType),
    roleType,
    role: roleLabel,
    roleClass: invitee.roleClass || getRoleClass(roleType),
    desc: invitee.desc || invitee.description || invitee.serviceText || '',
    selected: invitee.selected !== false
  }
}

function normalizeReplayContext(context = DEFAULT_REPLAY_CONTEXT) {
  const source = context && typeof context === 'object' ? context : DEFAULT_REPLAY_CONTEXT
  const invitees = (Array.isArray(source.invitees) && source.invitees.length
    ? source.invitees
    : DEFAULT_REPLAY_CONTEXT.invitees).map(normalizeInvitee)
  const selectedInviteeIds = invitees
    .filter((item) => item.selected)
    .map((item) => item.id)

  return {
    sourceGameId: source.sourceGameId || source.gameId || DEFAULT_REPLAY_CONTEXT.sourceGameId,
    serviceOrderId: source.serviceOrderId || DEFAULT_REPLAY_CONTEXT.serviceOrderId,
    inviter: source.inviter || DEFAULT_REPLAY_CONTEXT.inviter,
    previousSession: normalizePreviousSession(source.previousSession || source.session || source),
    invitees,
    selectedInviteeIds,
    selectedCount: selectedInviteeIds.length,
    maxInviteeCount: invitees.length
  }
}

Page({
  data: {
    sourceGameId: DEFAULT_REPLAY_CONTEXT.sourceGameId,
    serviceOrderId: DEFAULT_REPLAY_CONTEXT.serviceOrderId,
    previousSession: normalizePreviousSession(DEFAULT_REPLAY_CONTEXT.previousSession),
    inviter: DEFAULT_REPLAY_CONTEXT.inviter,
    invitees: normalizeReplayContext(DEFAULT_REPLAY_CONTEXT).invitees,
    selectedInviteeIds: normalizeReplayContext(DEFAULT_REPLAY_CONTEXT).selectedInviteeIds,
    selectedCount: normalizeReplayContext(DEFAULT_REPLAY_CONTEXT).selectedCount,
    maxInviteeCount: normalizeReplayContext(DEFAULT_REPLAY_CONTEXT).maxInviteeCount,
    quickMessages: QUICK_MESSAGES,
    selectedMessage: QUICK_MESSAGES[0],
    customMessage: '',
    loading: false,
    submitting: false
  },

  onLoad(options = {}) {
    this.loadReplayContext(options)
  },

  async loadReplayContext(options = {}) {
    const sourceGameId = options.sourceGameId || options.gameId || DEFAULT_REPLAY_CONTEXT.sourceGameId
    const serviceOrderId = options.serviceOrderId || DEFAULT_REPLAY_CONTEXT.serviceOrderId

    this.setData({
      sourceGameId,
      serviceOrderId,
      loading: true
    })

    try {
      const context = await gameService.getReplayConfirmContext({
        sourceGameId,
        serviceOrderId
      })

      this.applyReplayContext(context)
    } catch (error) {
      this.applyReplayContext(Object.assign({}, DEFAULT_REPLAY_CONTEXT, {
        sourceGameId,
        serviceOrderId
      }))
      wx.showToast({
        title: error.message || '上局信息加载失败，已使用测试数据',
        icon: 'none'
      })
    }
  },

  applyReplayContext(context) {
    this.setData({
      ...normalizeReplayContext(context),
      loading: false
    })
  },

  onInviteeTap(event) {
    const id = event.currentTarget.dataset.id

    if (!id) {
      return
    }

    const invitees = this.toggleInviteeSelected(id)
    const selectedInviteeIds = invitees
      .filter((item) => item.selected)
      .map((item) => item.id)

    this.setData({
      invitees,
      selectedInviteeIds,
      selectedCount: selectedInviteeIds.length
    })
  },

  toggleInviteeSelected(id) {
    return (this.data.invitees || []).map((item) => item.id === id
      ? {
        ...item,
        selected: !item.selected
      }
      : item)
  },

  onMessageTap(event) {
    const text = event.currentTarget.dataset.text

    if (!text) {
      return
    }

    this.setData({
      selectedMessage: text
    })
  },

  onCustomMessageInput(event) {
    this.setData({
      customMessage: event.detail.value || ''
    })
  },

  onCancelTap() {
    wx.showToast({
      title: '取消组局待接入',
      icon: 'none'
    })
  },

  onConfirmTap() {
    if (!this.data.selectedCount) {
      wx.showToast({
        title: '请选择邀请对象',
        icon: 'none'
      })
      return
    }

    this.submitReplayInvitation()
  },

  async submitReplayInvitation() {
    if (this.data.submitting) {
      return
    }

    const selectedInvitees = (this.data.invitees || []).filter((item) => item.selected)

    this.setData({
      submitting: true
    })

    try {
      const result = await gameService.createReplayInvitation({
        sourceGameId: this.data.sourceGameId,
        serviceOrderId: this.data.serviceOrderId,
        inviter: this.data.inviter,
        invitees: selectedInvitees.map((item) => ({
          id: item.id,
          roleType: item.roleType
        })),
        message: this.data.customMessage || this.data.selectedMessage
      })

      wx.navigateTo({
        url: this.buildProgressUrl(result),
        fail: () => {
          wx.showToast({
            title: '已发起，进度页待接入',
            icon: 'none'
          })
        }
      })
    } catch (error) {
      wx.showToast({
        title: error.message || '确认发起失败',
        icon: 'none'
      })
    } finally {
      this.setData({
        submitting: false
      })
    }
  },

  buildProgressUrl(result = {}) {
    const query = [
      ['source', 'replay'],
      ['replayInvitationId', result.replayInvitationId || result.invitationId || ''],
      ['sourceGameId', this.data.sourceGameId],
      ['serviceOrderId', this.data.serviceOrderId]
    ]
      .filter((item) => item[1])
      .map((item) => `${item[0]}=${encodeURIComponent(item[1])}`)
      .join('&')

    return `/${ROUTES.gameGuideProgress}${query ? `?${query}` : ''}`
  }
})

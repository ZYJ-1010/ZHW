const gameService = require('../../../services/game')
const { ROUTES } = require('../../../config/routes')
const { getSurnameInitials } = require('../../../utils/avatar')
const { navigateShellRoute } = require('../../../utils/shell-nav')

const EMPTY_REPLAY_CONTEXT = {
  sourceGameId: '',
  inviter: null,
  previousSession: {
    serviceType: '',
    completedAtText: '',
    participantText: ''
  },
  invitees: [],
  quickMessages: []
}

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
    serviceType: session.serviceType || session.serviceName || session.title || '',
    completedAtText: session.completedAtText || session.completedAt || session.finishTimeText || '',
    participantText: session.participantText || session.membersText || ''
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

function normalizeReplayContext(context = EMPTY_REPLAY_CONTEXT) {
  const source = context && typeof context === 'object' ? context : EMPTY_REPLAY_CONTEXT
  const invitees = (Array.isArray(source.invitees) ? source.invitees : []).map(normalizeInvitee)
  const selectedInviteeIds = invitees
    .filter((item) => item.selected)
    .map((item) => item.id)
  const quickMessages = (Array.isArray(source.quickMessages) ? source.quickMessages : [])
    .map((item) => String(item || '').trim())
    .filter(Boolean)

  return {
    sourceGameId: source.sourceGameId || source.gameId || '',
    inviter: source.inviter || null,
    previousSession: normalizePreviousSession(source.previousSession || source.session || source),
    invitees,
    selectedInviteeIds,
    selectedCount: selectedInviteeIds.length,
    maxInviteeCount: invitees.length,
    quickMessages
  }
}

Page({
  data: {
    sourceGameId: '',
    previousSession: normalizePreviousSession(EMPTY_REPLAY_CONTEXT.previousSession),
    inviter: null,
    invitees: [],
    selectedInviteeIds: [],
    selectedCount: 0,
    maxInviteeCount: 0,
    quickMessages: [],
    selectedMessage: '',
    customMessage: '',
    loading: false,
    submitting: false,
    loadError: false
  },

  onLoad(options = {}) {
    this.loadReplayContext(options)
  },

  async loadReplayContext(options = {}) {
    const sourceGameId = options.sourceGameId || options.gameId || ''

    if (!sourceGameId) {
      this.setData({
        loadError: true,
        loading: false
      })
      wx.showToast({
        title: '缺少上局信息',
        icon: 'none'
      })
      return
    }

    this.setData({
      sourceGameId,
      loading: true
    })

    try {
      const context = await gameService.getReplayConfirmContext({
        sourceGameId
      })

      this.applyReplayContext(context)
    } catch (error) {
      this.applyReplayContext(Object.assign({}, EMPTY_REPLAY_CONTEXT, {
        sourceGameId
      }))
      wx.showToast({
        title: error.message || '上局信息加载失败',
        icon: 'none'
      })
    }
  },

  applyReplayContext(context) {
    const nextContext = normalizeReplayContext(context)
    this.setData({
      ...nextContext,
      selectedMessage: nextContext.quickMessages[0] || '',
      loading: false,
      loadError: !nextContext.sourceGameId
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
    const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    navigateShellRoute(ROUTES.gameGuideProgress)
  },

  onConfirmTap() {
    if (this.data.loadError || !this.data.sourceGameId) {
      wx.showToast({
        title: '上局信息未加载，暂不能发起',
        icon: 'none'
      })
      return
    }

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
      const idsByRole = (roleType) => selectedInvitees
        .filter((item) => item.roleType === roleType)
        .map((item) => Number(item.id))
        .filter((id) => Number.isInteger(id) && id > 0)
      const result = await gameService.createReplayInvitation({
        sourceGameId: this.data.sourceGameId,
        playerUserIds: idsByRole('player'),
        expertUserIds: idsByRole('expert'),
        guideUserIds: idsByRole('guide'),
        message: this.data.customMessage || this.data.selectedMessage
      })

      navigateShellRoute(this.buildProgressUrl(result))
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
      ['invitationId', result.primaryInvitationId || ''],
      ['gameId', result.replayGameId || ''],
      ['sourceGameId', this.data.sourceGameId]
    ]
      .filter((item) => item[1])
      .map((item) => `${item[0]}=${encodeURIComponent(item[1])}`)
      .join('&')

    return `/${ROUTES.gameGuideProgress}${query ? `?${query}` : ''}`
  }
})

const gameService = require('../../../services/game')
const { ROUTES } = require('../../../config/routes')
const { navigateShellRoute } = require('../../../utils/shell-nav')
const toast = require('../../../utils/toast')
const { toUserMessage } = require('../../../utils/user-message')

const EMPTY_PROFILE = {
  name: '',
  avatarText: '',
  avatarSrc: '',
  avatarClass: '',
  desc: '',
  roleKey: '',
  roleName: '',
  tags: [],
  badges: [],
  profileSections: []
}

function asArray(value) {
  return Array.isArray(value) ? value : []
}

function findInvitation(data = {}, invitationId = '') {
  const items = asArray(data.activeParties).concat(asArray(data.completedParties))
  return items.find((item) => String(item.invitationId || item.id || '') === String(invitationId)) || null
}

function normalizeRows(rows) {
  return asArray(rows).filter((row) => row && row.label && (row.value || row.actionText))
}

function withMapAction(rows) {
  return normalizeRows(rows).map((row) => {
    if (row.label === '地点' || row.label === '地址' || row.key === 'location') {
      return { ...row, actionText: row.actionText || '地图位置' }
    }
    return row
  })
}

function normalizeDetail(invitation = {}) {
  const display = invitation.detailDisplay || {}
  const status = display.status || {}
  const game = invitation.game || invitation.gameInfo || display.gameInfo || {}
  const profile = Object.assign({}, EMPTY_PROFILE, display.player || {})
  const isPendingPlayer = invitation.status === 'pending' && invitation.targetRole === 'player'

  return {
    hasDetail: Boolean(invitation.invitationId || invitation.id),
    pageTitle: display.pageTitle || '确认组局',
    referralText: display.referralText || '邀请你参与组局',
    status,
    countdownText: status.countdown || invitation.countdownText || invitation.remainingText || '',
    applicantTitle: display.applicantTitle || '行家信息',
    profile,
    relation: display.relation || {},
    showRelation: display.showRelation !== false,
    sessionInfo: withMapAction(display.sessionInfo || invitation.infoRows),
    confirmTitle: display.confirmTitle || '局详情',
    confirmRows: normalizeRows(display.confirmRows),
    noticeTitle: display.noticeTitle || '确认须知',
    noticeBullets: asArray(display.noticeBullets),
    actionTip: display.actionTip || '确认后将通知行家进行最终审核',
    confirmText: display.confirmText || invitation.chatAcceptButtonText || '确认参加',
    declineText: display.declineText || invitation.chatDeclineButtonText || '婉拒',
    canConfirm: display.canConfirm !== false && isPendingPlayer,
    confirmDisabledReason: display.confirmDisabledReason || invitation.confirmDisabledReason || '',
    showActionBar: display.showActionBar !== false && isPendingPlayer,
    readonlyText: invitation.status === 'accepted'
      ? '已确认'
      : invitation.status === 'rejected'
        ? '已婉拒'
        : (display.reviewReadonlyText || status.title || invitation.statusText || '已处理'),
    gameId: invitation.gameId || game.id || game.gameId || '',
    gameTitle: game.title || game.topic || invitation.gameTitle || '',
    targetRole: invitation.targetRole || ''
  }
}

Page({
  data: {
    onlineText: '在线',
    invitationId: '',
    gameId: '',
    actionLoading: false,
    hasDetail: false,
    emptyTitle: '组局信息未加载',
    emptyText: '请从玩家组局确认通知进入',
    ...normalizeDetail()
  },

  onLoad(options = {}) {
    const invitationId = options.invitationId || options.id || ''
    const gameId = options.gameId || ''
    this.skipNextShowRefresh = true
    this.setData({ invitationId, gameId })
    this.loadInvitation(invitationId, gameId)
  },

  onShow() {
    if (this.skipNextShowRefresh) {
      this.skipNextShowRefresh = false
      return
    }
    this.loadInvitation()
  },

  async loadInvitation(invitationId = this.data.invitationId, gameId = this.data.gameId) {
    if (!invitationId) return null
    try {
      const data = await gameService.getGuideProgress({ invitationId, gameId })
      const invitation = findInvitation(data, invitationId)
      if (!invitation || invitation.targetRole !== 'player') {
        this.setData({ hasDetail: false, emptyText: '未找到这条玩家组局确认' })
        return
      }
      const resolvedGameId = invitation.gameId || (invitation.game && (invitation.game.id || invitation.game.gameId)) || gameId
      this.setData({
        ...normalizeDetail(invitation),
        invitationId: invitation.invitationId || invitation.id || invitationId,
        gameId: resolvedGameId
      })
      return invitation
    } catch (error) {
      this.setData({ hasDetail: false, emptyText: toUserMessage(error && error.message, '玩家确认信息加载失败') })
    }
  },

  navigateToGameDetail(gameId = this.data.gameId) {
    if (!gameId) {
      return false
    }
    navigateShellRoute(`${ROUTES.gameDetail}?gameId=${encodeURIComponent(gameId)}`, {
      reuseExisting: false
    })
    return true
  },

  handleMapTap() {
    const query = [
      this.data.gameId ? `gameId=${encodeURIComponent(this.data.gameId)}` : '',
      this.data.gameTitle ? `title=${encodeURIComponent(this.data.gameTitle)}` : ''
    ].filter(Boolean).join('&')
    navigateShellRoute(`/${ROUTES.map}${query ? `?${query}` : ''}`)
  },

  handleConfirmDisabledTap(event = {}) {
    const reason = event.detail && event.detail.reason
    toast.info(reason || this.data.confirmDisabledReason || '当前暂不能确认')
  },

  handleDeclineTap() {
    this.respondInvitation('reject')
  },

  handleConfirmTap() {
    this.respondInvitation('accept')
  },

  async respondInvitation(action) {
    if (this.data.actionLoading || !this.data.invitationId) return
    if (action === 'accept' && !this.data.canConfirm) {
      toast.info(this.data.confirmDisabledReason || '当前暂不能确认')
      return
    }
    this.setData({ actionLoading: true })
    try {
      await gameService.respondGameInvitation(this.data.invitationId, action)
      toast.success(action === 'accept' ? '已确认参加' : '已婉拒')
      if (action === 'accept') {
        wx.reLaunch({ url: `/${ROUTES.playerHome || ROUTES.home}` })
        return
      }
      await this.loadInvitation()
    } catch (error) {
      await this.loadInvitation()
      toast.info(error.message || '处理失败')
    } finally {
      this.setData({ actionLoading: false })
    }
  }
})

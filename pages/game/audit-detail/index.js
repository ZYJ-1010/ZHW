const { ROUTES } = require('../../../config/routes')
const toast = require('../../../utils/toast')

const DETAIL_SCROLL_TAP_STEP_RPX = 360
const DETAIL_SCROLL_HOLD_STEP_RPX = 72
const DETAIL_SCROLL_HOLD_INTERVAL_MS = 80
const DETAIL_SCROLL_HOLD_SUPPRESS_TAP_MS = 120

const GAME_PLAYER = {
  requirementConfirmed: true,
  requirementStatusText: '已确认需求',
  confirmed: true,
  statusText: '玩家已确认',
  avatarText: 'LM',
  name: '李明',
  desc: '某互联网公司 · 产品总监',
  tags: ['B端产品', '金融科技'],
  needText: '需求描述：需要资深产品经理帮忙梳理产品架构，预计咨询时长2小时，预算800元。',
  expectedTime: '期望时间：本周内',
  remark: '玩家备注：时间比较紧，希望能尽快开始'
}

const GAME_GUIDE = {
  iconSrc: '/pages/game/guide-chat/assets/icon-invite.png',
  online: true,
  avatarText: 'WA',
  name: '王引荐',
  recommendation: '李明是我之前合作过的客户，非常靠谱，需求也很明确。他急需产品架构方面的建议，我觉得你的经验很匹配。预算方面也比较充足，建议可以接。'
}

const GAME_INFO = {
  topic: '产品开发梳理（1v3）',
  time: '2026年3月23日（周六）14:00-17:00',
  location: '朝阳区图书馆（1号会议室）',
  activityType: '产品架构梳理咨询',
  serviceDuration: '2小时',
  clientBudget: '¥800'
}

const BACKEND_SETTLEMENT = {
  platformFee: '¥80（10%）',
  guideReward: '¥320（40%）',
  partnerReward: '¥80（10%）',
  expertIncome: '¥320'
}

function buildSessionInfo(gameInfo) {
  return [
    {
      label: '组局主题',
      value: gameInfo.topic,
      iconText: 'H',
      iconClass: 'topic'
    },
    {
      label: '时间',
      value: gameInfo.time,
      iconSrc: '/pages/game/detail/assets/icon-clock.png',
      iconText: '',
      iconClass: 'time'
    },
    {
      label: '地点',
      value: gameInfo.location,
      actionText: '地图位置',
      iconSrc: '/pages/game/detail/assets/icon-location.png',
      iconText: '',
      iconClass: 'place'
    }
  ]
}

function buildConfirmRows(gameInfo, settlement) {
  return [
    { label: '活动类型', value: gameInfo.activityType },
    { label: '服务时长', value: gameInfo.serviceDuration },
    { label: '客户预算', value: gameInfo.clientBudget, highlight: true, divider: true },
    { label: '平台', value: settlement.platformFee },
    { label: '领路人', value: settlement.guideReward },
    { label: '生态合伙人', value: settlement.partnerReward, divider: true },
    { label: '你的收益', value: settlement.expertIncome, success: true, total: true }
  ]
}

Page({
  data: {
    auditId: '',
    actionLoading: false,
    onlineText: '3999人在线',
    detailScrollTop: 0,
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ],
    status: {
      title: '等待你通过',
      quote: '老师！您好，目前在产品需求梳理和开发、设计方面遇到很多卡点，特别想参加该场局，以解决卡点',
      guideName: GAME_GUIDE.name,
      countdown: '23:45:12'
    },
    relation: {
      title: '组局关系图',
      totalCount: 6,
      confirmedCount: 3,
      expert: {
        avatarText: 'ME',
        name: '我（行家）',
        roleText: '服务提供方'
      },
      guide: {
        iconSrc: GAME_GUIDE.iconSrc,
        online: GAME_GUIDE.online,
        avatarText: GAME_GUIDE.avatarText,
        name: GAME_GUIDE.name
      },
      player: {
        avatarText: GAME_PLAYER.avatarText,
        name: GAME_PLAYER.name,
        confirmed: GAME_PLAYER.confirmed,
        statusText: GAME_PLAYER.statusText
      }
    },
    game: {
      player: GAME_PLAYER,
      info: GAME_INFO,
      guide: GAME_GUIDE
    },
    backendSettlement: BACKEND_SETTLEMENT,
    sessionInfo: buildSessionInfo(GAME_INFO),
    confirmRows: buildConfirmRows(GAME_INFO, BACKEND_SETTLEMENT),
    optionalActions: [
      {
        key: 'time',
        name: '提议具体时间',
        iconSrc: '/pages/game/audit-detail/assets/option-time.png'
      },
      {
        key: 'chat',
        name: '与玩家沟通',
        iconSrc: '/pages/game/audit-detail/assets/option-chat.png'
      }
    ],
    noticeBullets: [
      '确认后请准时参加，如需取消请提前24小时通知',
      '双方确认后组局正式生效，领路人将获得积分奖励',
      '请保持专业态度，维护平台信誉'
    ]
  },

  onLoad(options = {}) {
    this.setData({
      auditId: options.auditId || options.id || ''
    })
  },

  handleMapTap() {
    toast.info('地图位置待接入')
  },

  handleUploadImageTap() {
    toast.info('上传图片功能开发中')
  },

  handleUploadFileTap() {
    toast.info('上传文件功能开发中')
  },

  handleOptionalActionTap(event) {
    const key = event.currentTarget.dataset.key

    if (key === 'chat') {
      this.navigateToRoute(ROUTES.imRoom)
      return
    }

    toast.info('提议具体时间功能开发中')
  },

  handleDeclineTap() {
    toast.info('已婉拒该组局审核')
  },

  handleApproveTap() {
    if (this.data.actionLoading) {
      return
    }

    this.setData({
      actionLoading: true
    })

    this.approveTimer = setTimeout(() => {
      this.approveTimer = null
      this.setData({
        actionLoading: false
      })
      toast.success('已确认通过')
    }, 500)
  },

  handleShellNavTap(event) {
    const key = event.detail && event.detail.key

    if (key === 'up' || key === 'down') {
      if (!this.suppressNextNavTap) {
        this.scrollDetail(key, DETAIL_SCROLL_TAP_STEP_RPX)
      }
      return
    }

    if (key === 'left' || key === 'right') {
      toast.info('功能正在开发中')
      return
    }

    this.handleShellAction(key)
  },

  handleShellNavLongPress(event) {
    const key = event.detail && event.detail.key

    if (key !== 'up' && key !== 'down') {
      return
    }

    this.suppressNextNavTap = true
    this.stopDetailScrollHold(false)
    this.scrollDetail(key, DETAIL_SCROLL_HOLD_STEP_RPX)

    this.detailScrollHoldTimer = setInterval(() => {
      this.scrollDetail(key, DETAIL_SCROLL_HOLD_STEP_RPX)
    }, DETAIL_SCROLL_HOLD_INTERVAL_MS)
  },

  handleShellNavTouchEnd() {
    this.stopDetailScrollHold(true)
  },

  handleShellAction(key) {
    if (key === 'home') {
      this.scrollDetailToTop()
      return
    }

    if (key === 'search') {
      toast.info('搜索功能开发中')
      return
    }

    if (key === 'comment' || key === 'message') {
      this.navigateToRoute(ROUTES.message)
      return
    }

    if (key === 'avatar' || key === 'mine') {
      this.navigateToRoute(ROUTES.profile)
      return
    }

    const routeMap = {
      metaverse: ROUTES.metaverse,
      map: ROUTES.map
    }

    this.navigateToRoute(routeMap[key])
  },

  navigateToRoute(route) {
    if (!route || route === ROUTES.gameAuditDetail) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  },

  handleDetailScroll(event) {
    const scrollTop = event.detail && event.detail.scrollTop

    if (typeof scrollTop === 'number') {
      this.detailScrollTopValue = scrollTop
    }
  },

  scrollDetail(direction, stepRpx = DETAIL_SCROLL_TAP_STEP_RPX) {
    const current = Number(this.detailScrollTopValue || this.data.detailScrollTop || 0)
    const distance = this.rpxToPx(stepRpx)
    const nextTop = direction === 'up'
      ? Math.max(0, current - distance)
      : current + distance

    this.detailScrollTopValue = nextTop
    this.setData({
      detailScrollTop: nextTop
    })
  },

  scrollDetailToTop() {
    this.detailScrollTopValue = 0
    this.setData({
      detailScrollTop: 0
    })
  },

  stopDetailScrollHold(resetTapSuppress) {
    if (this.detailScrollHoldTimer) {
      clearInterval(this.detailScrollHoldTimer)
      this.detailScrollHoldTimer = null
    }

    if (resetTapSuppress && this.suppressNextNavTap) {
      if (this.detailScrollSuppressTimer) {
        clearTimeout(this.detailScrollSuppressTimer)
      }

      this.detailScrollSuppressTimer = setTimeout(() => {
        this.suppressNextNavTap = false
        this.detailScrollSuppressTimer = null
      }, DETAIL_SCROLL_HOLD_SUPPRESS_TAP_MS)
    }
  },

  clearDetailScrollTimers() {
    this.stopDetailScrollHold(false)

    if (this.detailScrollSuppressTimer) {
      clearTimeout(this.detailScrollSuppressTimer)
      this.detailScrollSuppressTimer = null
    }

    if (this.approveTimer) {
      clearTimeout(this.approveTimer)
      this.approveTimer = null
    }

    this.suppressNextNavTap = false
  },

  rpxToPx(value) {
    if (!wx.getSystemInfoSync) {
      return value / 2
    }

    const system = wx.getSystemInfoSync()
    const windowWidth = system && system.windowWidth ? system.windowWidth : 375

    return Math.round((value * windowWidth) / 750)
  },

  onUnload() {
    this.clearDetailScrollTimers()
  }
})

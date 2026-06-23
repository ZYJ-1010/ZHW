const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
const { getSurnameInitials } = require('../../../utils/avatar')

const DETAIL_SCROLL_TAP_STEP_RPX = 360
const DETAIL_SCROLL_HOLD_STEP_RPX = 72
const DETAIL_SCROLL_HOLD_INTERVAL_MS = 80
const DETAIL_SCROLL_HOLD_SUPPRESS_TAP_MS = 120

Page({
  data: {
    gameId: '',
    invitationId: '',
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
      title: '等待你确认',
      quote: '这位专家我合作过，非常专业，相信能解决你的问题！',
      guideName: '王引荐',
      countdown: '23:45:12'
    },
    expert: {
      name: '张专家',
      avatarText: getSurnameInitials('张专家', 'ZH'),
      desc: '资深产品经理 · 10年经验',
      rating: '4.9',
      serviceText: '服务50+客户',
      tags: ['产品架构', 'MVP规划', '用户增长', 'B端产品'],
      intro: '擅长从0到1的产品架构设计，曾主导多个千万级用户产品。提供产品咨询、架构梳理、团队搭建建议等服务。'
    },
    party: {
      confirmedText: '3/6人已确认',
      player: {
        name: '李娜',
        avatarText: getSurnameInitials('李娜', 'LI'),
        role: '玩家',
        state: '待确认'
      },
      expert: {
        name: '王强',
        avatarText: getSurnameInitials('王强', 'WA'),
        role: '行家',
        state: '待确认'
      }
    },
    sessionInfo: [
      {
        label: '组局主题',
        value: '产品开发 (1v3)',
        iconText: 'H',
        iconClass: 'topic'
      },
      {
        label: '时间',
        value: '2026年3月23日 (周六) 14:00-17:00',
        iconText: 'T',
        iconClass: 'time',
        iconSrc: '/pages/game/detail/assets/icon-clock.png'
      },
      {
        label: '地点',
        value: '朝阳区图书馆 (3号会议室)',
        actionText: '地图位置',
        iconText: 'P',
        iconClass: 'place',
        iconSrc: '/pages/game/detail/assets/icon-location.png'
      }
    ],
    detailRows: [
      { label: '服务类型', value: '产品架构梳理咨询' },
      { label: '咨询时长', value: '2小时' },
      { label: '预算金额', value: '¥800', highlight: true },
      { label: '预计时间', value: '本周内', last: true }
    ],
    costRows: [
      { label: '服务费用', value: '¥800' },
      { label: '合计支付', value: '¥800', highlight: true, total: true, last: true }
    ],
    protectionText: '资金由平台托管，服务完成后支付给服务方',
    noticeBullets: [
      '确认后请准时参加，如需取消请提前24小时通知',
      '双方确认后组局正式生效，领路人将获得积分奖励'
    ]
  },

  onLoad(options = {}) {
    this.setData({
      gameId: options.gameId || options.id || '',
      invitationId: options.invitationId || ''
    })
  },

  onMapTap() {
    this.showInfo('地图位置待接入')
  },

  onDecline() {
    this.handleInvitationRespond('reject')
  },

  onConfirmAttend() {
    this.handleInvitationRespond('accept')
  },

  async handleInvitationRespond(action) {
    if (this.data.actionLoading) {
      return
    }

    this.setData({
      actionLoading: true
    })

    try {
      await gameService.respondGameInvitation(this.data.invitationId, action)
      this.showInfo(action === 'accept' ? '已确认参加，等待最终审核' : '已婉拒')
    } catch (error) {
      this.showInfo((error && error.message) || '处理失败，请重试')
    } finally {
      this.setData({
        actionLoading: false
      })
    }
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
      this.showInfo('功能正在开发中')
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
      this.showInfo('搜索功能开发中')
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
    if (!route || route === ROUTES.gameDetail) {
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

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  },

  onUnload() {
    this.clearDetailScrollTimers()
  }
})

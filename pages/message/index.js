const { ROUTES } = require('../../config/routes')
const messageService = require('../../services/message')

const NAV_ITEMS = [
  { name: '我的', key: 'mine' },
  { name: '元宇宙', key: 'metaverse' },
  { name: '地图', key: 'map' },
  { name: '消息', key: 'message' },
  { name: '首页', key: 'home' }
]

const QUICK_ACTIONS = [
  { key: 'join', label: '组局加入', iconSrc: '/pages/message/assets/i53@3x.png', tone: 'blue', unreadCount: 1 },
  { key: 'system', label: '系统通知', iconSrc: '/pages/message/assets/i54@3x.png', tone: 'green' },
  { key: 'achievement', label: '成就解锁', iconSrc: '/pages/message/assets/i55@3x.png', tone: 'yellow' },
  { key: 'warning', label: '预警通知', iconSrc: '/pages/message/assets/i56@3x.png', tone: 'red' },
  { key: 'friend', label: '好友', iconSrc: '/pages/message/assets/i57@3x.png', tone: 'cyan', unreadCount: 3 }
]

const TABS = [
  { key: 'all', label: '全部消息' },
  { key: 'unread', label: '未读 (3)' },
  { key: 'trade', label: '交易通知' }
]

const MESSAGE_SECTIONS = [
  {
    key: 'system',
    title: '系统通知',
    items: [
      {
        id: 'platform-notice',
        routeKey: 'system',
        iconSrc: '/pages/message/assets/i58@3x.png',
        tone: 'blue',
        title: '平台公告',
        timeText: '2小时前',
        desc: '关于组局功能升级的通知：新增“智能匹配”功能，可自动推荐合适的组局对象...'
      },
      {
        id: 'audit-result',
        iconSrc: '/pages/message/assets/i59@3x.png',
        tone: 'purple',
        title: '活动审核结果',
        timeText: '昨天',
        desc: '你发布的活动“AI技术分享会”已通过审核，将于明天10:00开始展示或者前往组局中心手动发布',
        tagText: '审核通过',
        tagTone: 'success'
      }
    ]
  },
  {
    key: 'group',
    title: '组局动态',
    moreText: '查看全部',
    items: [
      {
        id: 'group-confirm',
        iconSrc: '/pages/message/assets/i60@3x.png',
        tone: 'orange',
        unread: true,
        title: '组局确认通知',
        timeText: '10:23',
        desc: '张伟 发起组局邀请你参与“周末篮球局”，需要你确认是否参加',
        highlightText: '张伟',
        actions: [
          { key: 'accept', text: '确认参加', primary: true },
          { key: 'reject', text: '婉拒' }
        ]
      },
      {
        id: 'group-success',
        iconSrc: '/pages/message/assets/i54@3x.png',
        tone: 'green',
        title: '组局已成局',
        timeText: '昨天',
        desc: '你引荐的 李娜 与 王强 已成功组局“产品经理交流会”',
        summaryText: '✓ 引荐成功',
        subText: '获得积分 +50'
      },
      {
        id: 'pay-success',
        iconSrc: '/pages/message/assets/i61@3x.png',
        tone: 'orangeLight',
        title: '支付成功通知',
        timeText: '昨天',
        desc: '你成功支付了“早起星人挑战”押金 ¥100.00，资金已进入押金池托管。'
      },
      {
        id: 'join-apply',
        avatarText: '小',
        title: '小红 申请加入你的局',
        timeText: '10:30',
        desc: '局：【武康路】复古胶片摄影局...',
        actions: [
          { key: 'decline', text: '拒绝' },
          { key: 'chat', text: '通过并私聊', primary: true, orange: true }
        ]
      }
    ]
  },
  {
    key: 'achievement',
    title: '新增成就',
    items: [
      {
        id: 'achievement-unlock',
        iconSrc: '/pages/message/assets/i62@3x.png',
        tone: 'yellow',
        title: '解锁新成就！',
        timeText: '3月30日',
        desc: '恭喜你解锁了“魔都探险家”成就，获得 200 积分奖励！'
      }
    ]
  },
  {
    key: 'warning',
    title: '预警提醒',
    items: [
      {
        id: 'delivery-warning',
        routeKey: 'warning',
        iconSrc: '/pages/message/assets/i65@3x.png',
        tone: 'red',
        alert: true,
        title: '待交付订单提醒',
        timeText: '2小时前',
        descParts: [
          { text: '你有1个组局服务订单将于 ' },
          { text: '2小时后', danger: true },
          { text: ' 到期交付，请及时处理' }
        ],
        metaText: '订单号：GD2024032201',
        linkText: '立即处理'
      },
      {
        id: 'activity-soon',
        iconSrc: '/pages/message/assets/i64@3x.png',
        tone: 'yellow',
        title: '活动即将开始',
        timeText: '30分钟后',
        desc: '你参与的组局“周末徒步”将于今天14:00开始，地点：奥林匹克森林公园南门',
        actions: [
          { key: 'route', text: '查看路线', primary: true },
          { key: 'contact', text: '联系发起人' }
        ]
      }
    ]
  },
  {
    key: 'friends',
    title: '好友消息',
    items: [
      {
        id: 'friend-liming',
        routeKey: 'friend',
        avatarText: 'LM',
        online: true,
        title: '李明',
        timeText: '12:30',
        desc: '好的，那我们就周六下午2点在咖啡店见，我带上项目资料...',
        unreadCount: 3
      }
    ]
  }
]

function normalizeQuickActions(actions) {
  const source = Array.isArray(actions) && actions.length ? actions : QUICK_ACTIONS

  return source.map((item) => {
    const unreadCount = Number(item.unreadCount || 0)

    return Object.assign({}, item, {
      unreadCount,
      unread: Boolean(item.unread) || unreadCount > 0
    })
  })
}

function normalizeTabs(tabs) {
  const source = Array.isArray(tabs) && tabs.length ? tabs : TABS

  return source.map((item) => {
    const unreadCount = Number(item.unreadCount || 0)

    return Object.assign({}, item, {
      unreadCount,
      unread: Boolean(item.unread) || unreadCount > 0
    })
  })
}

function normalizeSections(sections) {
  const source = Array.isArray(sections) && sections.length ? sections : MESSAGE_SECTIONS

  return source.map((section) => Object.assign({}, section, {
    items: Array.isArray(section.items) ? section.items : []
  }))
}

function normalizeMessageCenter(source = {}, fallbackActiveTab = 'all') {
  return {
    pageTitle: source.pageTitle || '消息中心',
    onlineText: source.onlineText || '3999人在线',
    quickActions: normalizeQuickActions(source.quickActions),
    tabs: normalizeTabs(source.tabs),
    activeTab: source.activeTab || fallbackActiveTab || 'all',
    sections: normalizeSections(source.sections)
  }
}

Page({
  data: {
    pageTitle: '消息中心',
    onlineText: '3999人在线',
    navItems: NAV_ITEMS,
    quickActions: normalizeQuickActions(QUICK_ACTIONS),
    tabs: normalizeTabs(TABS),
    activeTab: 'all',
    sections: MESSAGE_SECTIONS,
    loading: false,
    errorText: ''
  },

  onLoad(options = {}) {
    this.loadMessageCenter({
      tab: options.tab || 'all'
    })
  },

  async loadMessageCenter(params = {}) {
    const tab = params.tab || this.data.activeTab || 'all'

    this.setData({
      loading: true,
      errorText: ''
    })

    try {
      const center = await messageService.getMessageCenter({ tab })
      const normalized = normalizeMessageCenter(center, tab)

      this.setData(Object.assign({}, normalized, {
        loading: false,
        errorText: ''
      }))
    } catch (error) {
      const errorText = error && error.message ? error.message : '消息中心加载失败'

      this.setData({
        loading: false,
        errorText
      })
      this.showInfo(errorText)
    }
  },

  onTabTap(event) {
    const { key } = event.currentTarget.dataset

    if (!key || key === this.data.activeTab) {
      return
    }

    this.setData({
      activeTab: key
    })
    this.loadMessageCenter({
      tab: key
    })
  },

  onQuickTap(event) {
    const { key } = event.currentTarget.dataset
    const routeMap = {
      system: ROUTES.messageSystemDetail || 'pages/message/system-detail/index',
      warning: ROUTES.messageTradeWarning || 'pages/message/trade-warning/index',
      friend: ROUTES.messageMy || 'pages/message/my/index'
    }
    const route = routeMap[key]

    if (route) {
      wx.navigateTo({
        url: `/${route}`
      })
      return
    }

    this.showInfo('功能正在开发中')
  },

  onMessageTap(event) {
    const { routeKey } = event.currentTarget.dataset
    const routeMap = {
      system: ROUTES.messageSystemDetail || 'pages/message/system-detail/index',
      warning: ROUTES.messageTradeWarning || 'pages/message/trade-warning/index',
      friend: ROUTES.messageMy || 'pages/message/my/index'
    }
    const route = routeMap[routeKey]

    if (!route) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  },

  onActionTap(event) {
    const { action } = event.currentTarget.dataset

    this.showInfo(`${action || '操作'}待接入`)
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}
    const routeMap = {
      home: ROUTES.playerHome || ROUTES.home,
      map: ROUTES.map,
      message: ROUTES.message,
      mine: ROUTES.profile,
      avatar: ROUTES.profile,
      metaverse: ROUTES.metaverse
    }
    const route = routeMap[key]

    if (!route || route === ROUTES.message) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})

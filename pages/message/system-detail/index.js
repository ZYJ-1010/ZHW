const { ROUTES } = require('../../../config/routes')
const messageService = require('../../../services/message')

const DEFAULT_SYSTEM_NOTIFICATION_DETAIL = {
  pageTitle: '系统通知',
  onlineText: '3999人在线',
  article: {
    tagText: '重要更新',
    title: '组局功能全新升级：智能匹配系统上线',
    author: '官方运营团队',
    publishedAtText: '2026-03-20',
    readText: '阅读 1.2k',
    blocks: [
      { id: 'lead', type: 'paragraph', text: '亲爱的用户：', lead: true },
      { id: 'intro', type: 'paragraph', text: '为了提升组局效率和匹配精准度，我们于今日正式上新智能匹配功能，根据你的行业标签、兴趣爱好、地理位置等多维度信息，自动推荐最合适的组局对象。' },
      {
        id: 'update-content',
        type: 'updateBox',
        icon: '★',
        title: '主要更新内容',
        points: [
          'AI智能推荐：基于行为分析的个性化推荐',
          '匹配度评分：直观展示双方契合程度',
          '一键邀约：简化组局发起流程'
        ]
      },
      { id: 'message-center', type: 'paragraph', text: '同时，我们对消息触达中心进行了优化，新增消息分类和优先级标记，确保你不会错过任何重要组局信息。' },
      { id: 'cover', type: 'cover', imageUrl: '/pages/message/system-detail/assets/system-update-cover.png', caption: '智能匹配界面示意图' },
      { id: 'closing', type: 'paragraph', text: '如有任何问题，欢迎联系客服团队。感谢你的支持与信任！' },
      { id: 'signature', type: 'signature', teamText: '产品团队', dateText: '2026年3月20日' }
    ]
  },
  feedback: {
    question: '这篇文章对你有帮助吗？',
    useful: { icon: '👍', label: '有用', countText: '128' },
    useless: { icon: '👎', label: '没用', countText: '10' }
  }
}

function normalizeSystemNotificationDetail(data = {}) {
  const source = Object.assign({}, DEFAULT_SYSTEM_NOTIFICATION_DETAIL, data)
  const article = Object.assign({}, DEFAULT_SYSTEM_NOTIFICATION_DETAIL.article, data.article || {})
  const feedback = normalizeFeedback(data.feedback)

  return {
    pageTitle: source.pageTitle || DEFAULT_SYSTEM_NOTIFICATION_DETAIL.pageTitle,
    onlineText: source.onlineText || DEFAULT_SYSTEM_NOTIFICATION_DETAIL.onlineText,
    article: Object.assign({}, article, {
      blocks: normalizeArticleBlocks(article.blocks)
    }),
    feedback,
    messageId: source.messageId || source.notificationId || source.id || '',
    loading: false,
    errorText: ''
  }
}

function normalizeArticleBlocks(blocks) {
  return Array.isArray(blocks) && blocks.length
    ? blocks
    : DEFAULT_SYSTEM_NOTIFICATION_DETAIL.article.blocks
}

function normalizeFeedback(feedback = {}) {
  const useful = Object.assign({}, DEFAULT_SYSTEM_NOTIFICATION_DETAIL.feedback.useful, feedback.useful || {})
  const useless = Object.assign({}, DEFAULT_SYSTEM_NOTIFICATION_DETAIL.feedback.useless, feedback.useless || {})

  return {
    question: feedback.question || DEFAULT_SYSTEM_NOTIFICATION_DETAIL.feedback.question,
    useful: Object.assign({}, useful, {
      countText: useful.countText || String(useful.count || 0)
    }),
    useless: Object.assign({}, useless, {
      countText: useless.countText || String(useless.count || 0)
    })
  }
}

Page({
  data: {
    pageTitle: '系统通知',
    onlineText: '3999人在线',
    navItems: [
      { name: '我的', key: 'mine' },
      { name: '元宇宙', key: 'metaverse' },
      { name: '地图', key: 'map' },
      { name: '消息', key: 'message' },
      { name: '首页', key: 'home' }
    ],
    article: DEFAULT_SYSTEM_NOTIFICATION_DETAIL.article,
    feedback: DEFAULT_SYSTEM_NOTIFICATION_DETAIL.feedback,
    messageId: '',
    loading: false,
    errorText: ''
  },

  onLoad(options = {}) {
    this.loadSystemNotificationDetail(options)
  },

  async loadSystemNotificationDetail(options = {}) {
    this.setData({
      loading: true,
      errorText: ''
    })

    try {
      const detail = await messageService.getSystemNotificationDetail({
        messageId: options.messageId || options.notificationId || options.id || ''
      })

      this.setData(normalizeSystemNotificationDetail(detail))
    } catch (error) {
      this.setData(Object.assign({}, normalizeSystemNotificationDetail(DEFAULT_SYSTEM_NOTIFICATION_DETAIL), {
        errorText: error.message || '获取系统通知失败'
      }))
      this.showInfo(error.message || '获取系统通知失败')
    }
  },

  onFeedbackTap(event) {
    const { value } = event.currentTarget.dataset
    const feedback = value === 'useful' ? this.data.feedback.useful : this.data.feedback.useless
    this.showInfo(`已记录${feedback.label}反馈`)
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

    if (!route || route === ROUTES.messageSystemDetail) {
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

const homeService = require('../../services/home')
const toast = require('../../utils/toast')
const { ROUTES } = require('../../config/routes')

const ROLE_COMPARISON_PREVIEW_PAGE = {
  name: '角色权益对比',
  mode: 'roleComparison',
  roleBadge: '权益',
  title: '角色权益对比',
  subtitle: '选择适合你的角色，开启不同玩法',
  roles: [
    { key: 'player', name: '玩家', level: 'Lv.1+', icon: '🎮', active: true },
    { key: 'guide', name: '领路人', level: 'Lv.5+', icon: '🧭', active: false },
    { key: 'expert', name: '行家', level: 'Lv.20+', icon: '💎', active: false }
  ],
  benefits: [
    { name: '发起组局', player: '✓', leader: '—', expert: '✓' },
    { name: '加入组局', player: '✓', leader: '✓', expert: '✓' },
    { name: '创建路线', player: '✓', leader: '—', expert: '✓' },
    { name: '分润收益', player: '—', leader: '基础会员40%', expert: '高级会员40%' },
    { name: '服务交易', player: '—', leader: '—', expert: '✓' },
    { name: '数据看板', player: '—', leader: '✓', expert: '✓' },
    { name: '信用背书', player: '—', leader: '✓', expert: '✓' }
  ],
  primary: '立即申请角色'
}

const HOME_CONVERTED_PAGES = [
  {
    name: '启动动画',
    mode: 'launchAnimation',
    brand: '眞好玩',
    onlineText: '2999+人在线',
    title: '真好玩',
    subtitle: 'Zhen Hao Wan',
    actionText: 'GO',
    featureDots: ['pink', 'purple', 'cyan']
  },
  {
    name: '领路人申请进度',
    mode: 'leaderProgress',
    roleBadge: '领路人',
    title: '申请进度',
    subtitle: '平台正在审核您的申请资料',
    heroTitle: '领路人申请',
    status: '审核中',
    progress: 66,
    timelineTitle: '申请进度',
    timeline: [
      { title: '提交申请', desc: '已成功提交领路人申请', time: '2024.06.08 10:30', state: 'done', hasLine: true },
      { title: '资料初审', desc: '平台已完成基础资料核验', time: '2024.06.08 11:15', state: 'done', hasLine: true },
      { title: '深度审核', desc: '正在对您的资质进行深层验证，请耐心等待反馈', time: '进行中...', state: 'active', hasLine: true },
      { title: '结果通知', desc: '审核完成后将通过通知中心告知您', time: '待完成', state: 'pending', hasLine: false }
    ],
    detailsTitle: '申请凭证',
    details: [
      { label: '申请身份', value: '领路人', highlight: true },
      { label: '申请时间', value: '2024.06.08 10:30' },
      { label: '申请编号', value: 'LR20240608001' },
      { label: '当前状态', value: '深度审核中', highlight: true },
      { label: '预计完成时间', value: '2024.06.12 18:00' }
    ],
    helperText: '如有疑问请联系 App 内在线客服咨询协助',
    actions: ['返回玩家首页', '查看权益对比']
  },
  {
    name: '领路人审核通过',
    mode: 'roleAuditPassed',
    roleBadge: '审核结果',
    roleType: 'leader',
    title: '审核结果',
    headline: '恭喜审核通过！',
    roleText: '你已成为「领路人」',
    desc: ['你的申请已通过平台审核', '现在可以开始邀约玩家进入组局'],
    certNo: 'ZHW-00115-2026',
    certTime: '2026.06.10 14:30',
    giftTitle: '新手礼包',
    gifts: [
      { icon: '＋', name: '每月添加行家15位', tag: '限时' },
      { icon: '📌', name: '首页推荐7天', tag: '流量' },
      { icon: '◆', name: '赠送200经验值', tag: '奖励' }
    ],
    actions: [
      { icon: '网', name: '关系网开启' },
      { icon: '邀', name: '邀请玩家' },
      { icon: '人', name: '完善资料' }
    ],
    primary: '开启领路人之旅'
  },
  {
    name: '行家审核通过',
    mode: 'roleAuditPassed',
    roleBadge: '审核结果',
    roleType: 'expert',
    title: '审核结果',
    headline: '恭喜审核通过！',
    roleText: '你已成为「行家」',
    desc: ['你的申请已通过平台审核', '现在可以开始创建新局、交付服务'],
    certNo: 'ZHW-00126-2026',
    certTime: '2026.06.10 14:30',
    giftTitle: '新手礼包',
    gifts: [
      { icon: '证', name: '每月添加行家30位', tag: '限时' },
      { icon: '📌', name: '首页推荐7天', tag: '流量' },
      { icon: '◆', name: '赠送200经验值', tag: '奖励' }
    ],
    actions: [
      { icon: '图', name: '关系网开启' },
      { icon: '邀', name: '邀请玩家' },
      { icon: '证', name: '完善资料' }
    ],
    primary: '开启行家之旅'
  }
]

const EXPERT_APPLY_PREVIEW_PAGES = [
  {
    name: '申请行家内页',
    mode: 'expertApplyForm',
    roleBadge: '申请',
    navTitle: '申请行家',
    saveText: '保存',
    title: '申请成为行家',
    icon: '💎',
    tagline: '我懂玩家需要什么！我申请成为行家',
    reviewHint: '审核预计 1-3 个工作日',
    primary: '提交行家申请',
    skillOptions: [
      { name: '摄影', active: true },
      { name: '户外', active: false },
      { name: '美食', active: false },
      { name: '文化', active: false },
      { name: '手工', active: false },
      { name: '运动', active: false },
      { name: '音乐', active: false },
      { name: '+ 自定义', active: false }
    ],
    fields: [
      {
        type: 'chips',
        label: '选择技能领域',
        required: true
      },
      {
        type: 'input',
        label: '技能标签',
        required: true,
        placeholder: '如：人像摄影、风光摄影、夜景拍摄',
        helper: '添加具体标签，让用户更容易找到你'
      },
      {
        type: 'select',
        label: '从业年限',
        required: true,
        placeholder: '请选择从业年限'
      },
      {
        type: 'textarea',
        label: '个人简介',
        required: true,
        placeholder: '不少于 50 字，突出你的专业优势'
      },
      {
        type: 'upload',
        label: '资质证明',
        required: true,
        icon: '📎',
        title: '点击上传作品集及凭证',
        helper: '支持 JPG、PNG、PDF，最多 5 张'
      }
    ],
    priceRows: [
      { label: '服务定价（元/小时）', required: true, value: '业务：___________' },
      { label: '服务成本', required: false, value: '业务：___________' }
    ],
    priceHint: '平台将收取 10% 服务费'
  },
  {
    name: '申请行家浏览页',
    mode: 'expertApplyOverview',
    roleBadge: '申请',
    navTitle: '申请行家',
    saveText: '保存',
    title: '申请成为行家',
    icon: '💎',
    tagline: '我懂玩家需要什么！我申请成为行家',
    reviewHint: '审核预计 1-3 个工作日',
    primary: '下一步',
    requirementsTitle: '申请条件',
    requirements: [
      { title: '玩家等级达到 Lv.20', status: '当前等级: Lv.21 ✓ 已满足' },
      { title: '完成实名认证', status: '认证状态: 已通过 ✓ 已满足' },
      { title: '完成企业认证', status: '认证状态: 已通过 ✓ 已满足' },
      { title: '发起过 5次以上组局', status: '当前: 5 次 ✓ 已满足' },
      { title: '信用分 ≥ 90 分', status: '当前: 92 分 ✓ 已满足' },
      { title: '会员等级≥ 高级会员', status: '当前: 高级会员 ✓ 已满足' }
    ],
    planTask: {
      title: '提交行家计划书',
      desc: '需描述你的资源、能力和项目说明书',
      action: '去填写 ›'
    },
    benefitsTitle: '行家特权',
    benefits: [
      { icon: '💰', text: '有权益的行家可发起有偿局并可获得相应收入' },
      { icon: '⭐', text: '专属行家标识与优先推荐位' },
      { icon: '📊', text: '数据看板：查看服务数据与收益分析' }
    ]
  }
]

const EXPERT_APPLY_WALKTHROUGH_PAGES = [
  EXPERT_APPLY_PREVIEW_PAGES[1],
  EXPERT_APPLY_PREVIEW_PAGES[0]
]

const HOME_PREVIEW_PAGES = [
  {
    name: '玩家首页',
    mode: 'playerHome',
    brand: '真好玩',
    onlineText: '3999人在线',
    title: 'HELLO, 玩家!',
    subtitle: '2026.03.30 | 开启你的今日副本',
    roleTabs: [
      { name: '🎮 玩家', active: true },
      { name: '🎯 行家', active: false },
      { name: '🌐 领路人', active: false }
    ],
    featuredGame: {
      title: 'AI赋能系统搭建交流局',
      tag: '任务局',
      price: '￥0/人',
      action: '加入',
      location: '📍黄浦区 · 8.2km · 3/8人',
      time: '⏰2026年5月1日 14:00--16:00',
      joinedText: '+3位玩家已入局',
      imageText: 'AI',
      actions: ['分享', '关注', '引荐', '打招呼']
    },
    nearbyTabs: [
      { name: '全部', active: true },
      { name: '附近', active: false }
    ],
    nearby: [
      {
        title: '苏州河“记忆碎片”采集',
        tag: '探索局',
        price: '￥0/人',
        action: '加入',
        location: '📍静安区 · 3.2km · 5/8人',
        time: '⏰2026年5月1日 20:00--22:00',
        joinedText: '+5位玩家已入局',
        imageText: '河',
        actions: ['分享', '关注', '引荐', '打招呼']
      }
    ],
    rankingTabs: [
      { name: '玩家', active: true },
      { name: '行家', active: false },
      { name: '领路人', active: false }
    ],
    rankingList: [
      { rank: '01', avatar: '🏆', name: '领域专家 PRO', desc: '本周组局 12 · MVP 5次', xp: '2,450' },
      { rank: '02', avatar: '⭐', name: '社交达人', desc: '本周组局 8 次', xp: '1,890' },
      { rank: '03', avatar: '🌍', name: '探险家', desc: '本周组局 6 次', xp: '1,560' }
    ],
    myRank: {
      rank: '52',
      avatar: 'A',
      name: '我（Alex）',
      desc: '上周排名 65',
      xp: '520'
    },
    achievements: [
      { icon: '🏆', title: '百场王者', status: '▲ 等级' },
      { icon: '🏆', title: '引航王者', status: '▲ 等级' },
      { icon: '🔒', title: '隐藏徽章', status: '▲ 未解锁' },
      { icon: '🌍', title: '地球漫游者', status: '▲ 进度20%' }
    ],
    playerCard: {
      role: '玩家 Lv.5',
      title: '探险家',
      name: 'Alex Chen',
      xp: '580/1000 XP',
      progress: 58,
      next: '距离下一等级还需 420 经验值',
      stats: [
        { value: '12', label: '参与局数' },
        { value: '3', label: '本月MVP' },
        { value: '98%', label: '参与率' }
      ]
    },
    onlineCard: {
      title: '地球online',
      desc: '探索城市副本 · 解锁地图成就',
      leftTag: '附近 12 个组局',
      rightTag: '已打卡 8 处'
    },
    entries: [
      { title: '发起组局', desc: '创建你的带局房间', icon: '📍', theme: 'pink' },
      { title: '局前大厅', desc: '准备就绪加入一局', icon: '🎲', theme: 'cyan' }
    ],
    friendGames: [
      {
        title: '盲盒路线：3小时点亮天际线',
        tag: '探索局',
        price: '￥29/人',
        action: '加入',
        location: '📍梧桐山 · 1.5km · 3/8人',
        time: '2026年5月1日 14:00--16:00',
        joinedText: '+3位玩家已入局',
        imageText: '线',
        actions: ['分享', '关注', '引荐', '打招呼']
      }
    ],
    metaverse: {
      title: '进入元宇宙',
      desc: '共创数字街区｜全球联机互动',
      tags: ['3D空间', 'NFT徽章'],
      badge: '+99'
    },
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ],
    joinText: '加入'
  },
  {
    name: '行家首页',
    mode: 'dashboard',
    roleBadge: '行家',
    title: '行家工作台',
    subtitle: '管理邀约与审核',
    desc: '处理玩家申请，维护高质量局内体验',
    stats: [
      { value: '5', label: '待审核' },
      { value: '18', label: '已交付' },
      { value: '4.9', label: '评分' }
    ],
    cards: [
      { title: '玩家入局申请', meta: '3 条新申请等待处理', tag: '审核' },
      { title: '今日交付提醒', meta: '2 个组局需要确认', tag: '处理' }
    ],
    actions: ['审核列表', '交付确认', '评价管理']
  },
  {
    name: '领路人首页',
    mode: 'dashboard',
    roleBadge: '领路人',
    title: '领路人中心',
    subtitle: '邀请与撮合',
    desc: '为玩家匹配更合适的局和行家',
    stats: [
      { value: '9', label: '待接收' },
      { value: '32', label: '成功引荐' },
      { value: '6', label: '活跃城市' }
    ],
    cards: [
      { title: '发起邀请', meta: '为好友推荐合适局', tag: '邀请' },
      { title: '接收确认', meta: '4 个邀约等待反馈', tag: '查看' }
    ],
    actions: ['邀请玩家', '引荐记录', '收益概览']
  },
  ...HOME_CONVERTED_PAGES,
  {
    name: '无行家权限提示页',
    mode: 'permission',
    roleBadge: '权限提示',
    title: '暂未开通行家权限',
    subtitle: '申请通过后可审核玩家并交付组局',
    desc: '完善资料并提交申请，平台会尽快完成审核。',
    benefits: ['展示专业能力', '获得组局交付机会', '提升个人可信度'],
    primary: '申请成为行家',
    secondary: '先逛逛首页'
  },
  {
    name: '无领路人权限提示页',
    mode: 'permission',
    roleBadge: '权限提示',
    title: '暂未开通领路人权限',
    subtitle: '成为领路人后可发起邀请与引荐',
    desc: '你可以先完成认证，再申请领路人权限。',
    benefits: ['发起邀请', '撮合玩家与行家', '查看引荐收益'],
    primary: '申请成为领路人',
    secondary: '了解权益'
  },
  {
    name: '权益对比页',
    mode: 'compare',
    roleBadge: '权益',
    title: '角色权益对比',
    subtitle: '选择适合你的参与方式',
    plans: [
      { name: '玩家', desc: '加入组局、认识同好', highlights: ['浏览大厅', '申请入局', '评价体验'] },
      { name: '行家', desc: '审核申请、完成交付', highlights: ['审核玩家', '交付确认', '获得评分'] },
      { name: '领路人', desc: '邀请推荐、完成撮合', highlights: ['发起邀请', '引荐记录', '收益概览'] }
    ]
  },
  ...EXPERT_APPLY_WALKTHROUGH_PAGES,
  {
    name: '申请领路人操作页',
    mode: 'apply',
    roleBadge: '申请',
    title: '申请成为领路人',
    subtitle: '完善邀请与撮合资料',
    desc: '说明你的城市资源、社交圈层和可推荐方向。',
    formItems: ['所在城市', '可推荐人群', '常用联系方式', '引荐说明'],
    primary: '提交领路人申请'
  },
  {
    name: '审核进度页',
    mode: 'review',
    roleBadge: '审核',
    title: '审核进行中',
    subtitle: '资料已提交，等待平台审核',
    status: '预计 1-3 个工作日完成',
    steps: ['提交资料', '平台审核', '结果通知']
  },
  {
    name: '审核通过页',
    mode: 'review',
    roleBadge: '通过',
    title: '审核已通过',
    subtitle: '你的角色权限已开通',
    status: '现在可以进入对应工作台',
    steps: ['资料确认', '权限开通', '开始使用']
  },
  {
    name: '审核驳回页',
    mode: 'review',
    roleBadge: '驳回',
    title: '审核未通过',
    subtitle: '请根据原因修改资料后再次提交',
    status: '常见原因：资料不完整或经历说明不足',
    steps: ['查看原因', '修改资料', '重新提交']
  },
  {
    name: '申请状态提示页',
    mode: 'review',
    roleBadge: '状态',
    title: '已有申请记录',
    subtitle: '当前申请仍在处理中',
    status: '请勿重复提交，结果会通过消息通知',
    steps: ['已提交', '审核中', '待通知']
  },
  {
    name: '申请进度查看页',
    mode: 'review',
    roleBadge: '进度',
    title: '申请进度',
    subtitle: '查看当前角色申请节点',
    status: '当前节点：平台审核',
    steps: ['基础资料', '角色信息', '平台审核', '完成']
  },
  {
    name: '首页调色说明页',
    mode: 'theme',
    roleBadge: '视觉',
    title: '首页调色说明',
    subtitle: '首页模块色彩与层级参考',
    colors: [
      { name: '主渐变', value: '#8A5CF6 / #E44F9B' },
      { name: '成功绿', value: '#22C79A' },
      { name: '页面底色', value: '#F7F7F8' },
      { name: '主文字', value: '#1A1A2E' }
    ]
  }
]

const HOME_PREVIEW_LOOKUP_PAGES = [
  ...HOME_PREVIEW_PAGES,
  ROLE_COMPARISON_PREVIEW_PAGE
]

const HOME_PREVIEW_GROUPS = {
  expertApplyWalkthrough: EXPERT_APPLY_WALKTHROUGH_PAGES
}

Page({
  data: {
    isHomePreview: false,
    homePreviewIndex: 0,
    homePreviewNo: 1,
    homePreviewTotal: HOME_PREVIEW_PAGES.length,
    homePreviewPages: HOME_PREVIEW_PAGES,
    currentHomePreview: HOME_PREVIEW_PAGES[0],
    homePreviewSingle: false,
    previewWindowWidth: 375,
    roleComparisonReturnTo: '',
    roleComparisonShell: {
      onlineText: '3999人在线',
      navItems: [
        { name: '我的', active: false },
        { name: '元宇宙', active: false },
        { name: '地图', active: false },
        { name: '消息', active: false },
        { name: '首页', active: true }
      ]
    },
    loading: true,
    user: {
      nickname: '',
      todayCreditScore: '',
      level: '',
      member: {
        planName: ''
      },
      needRealname: true
    },
    hero: {
      roleName: '',
      dateLabel: '',
      subtitle: '',
      onlineText: ''
    },
    notices: [],
    quickActions: [],
    recommendedGames: [],
    playerSummary: {
      displayName: '',
      roleLabel: '',
      title: '',
      xpText: '',
      progressPercent: 0,
      nextLevelText: '',
      stats: []
    },
    rankingList: [],
    achievementList: [],
    friendGames: [],
    metaverseEntry: {
      title: '',
      desc: '',
      actionText: '',
      route: ''
    },
    nearbySummary: {
      cityName: '',
      count: 0
    }
  },

  onLoad(options = {}) {
    if (options.ui === '1') {
      this.setData({
        roleComparisonReturnTo: decodeURIComponent(options.returnTo || '')
      })
      this.enterHomePreview(options.mode || '', options.single === '1')
      return
    }

    this.loadHome()
  },

  enterHomePreview(mode, single = false) {
    const previewPages = HOME_PREVIEW_GROUPS[mode] || HOME_PREVIEW_PAGES
    const index = previewPages.findIndex((page) => page.name === mode || page.mode === mode)
    const lookupIndex = index >= 0
      ? index
      : HOME_PREVIEW_LOOKUP_PAGES.findIndex((page) => page.name === mode || page.mode === mode)
    const currentHomePreview = lookupIndex >= 0
      ? HOME_PREVIEW_LOOKUP_PAGES[lookupIndex]
      : previewPages[0]
    const previewWindowWidth = wx.getSystemInfoSync ? wx.getSystemInfoSync().windowWidth : 375

    this.setData({
      isHomePreview: true,
      loading: false,
      previewWindowWidth,
      homePreviewSingle: single,
      homePreviewPages: previewPages,
      homePreviewIndex: index >= 0 ? index : 0,
      homePreviewNo: index >= 0 ? index + 1 : 1,
      homePreviewTotal: previewPages.length,
      currentHomePreview
    })
  },

  handleHomePreviewTap(event) {
    if (!this.data.isHomePreview) {
      return
    }

    if (this.data.currentHomePreview && this.data.currentHomePreview.mode === 'roleComparison') {
      return
    }

    if (this.data.homePreviewSingle) {
      return
    }

    const datasetDirection = Number(event.currentTarget.dataset.direction)
    const touch = event.changedTouches && event.changedTouches[0]
    const x = touch ? touch.clientX : event.detail.x
    const direction = datasetDirection || (x < this.data.previewWindowWidth / 2 ? -1 : 1)
    const previewPages = this.data.homePreviewPages || HOME_PREVIEW_PAGES
    const total = previewPages.length
    const nextIndex = (this.data.homePreviewIndex + direction + total) % total

    this.setData({
      homePreviewIndex: nextIndex,
      homePreviewNo: nextIndex + 1,
      currentHomePreview: previewPages[nextIndex]
    })
  },

  handleRoleCompareBackTap() {
    const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []

    if (pages.length > 1 && typeof wx.navigateBack === 'function') {
      wx.navigateBack()
      return
    }

    if (this.data.roleComparisonReturnTo && typeof wx.redirectTo === 'function') {
      wx.redirectTo({
        url: this.data.roleComparisonReturnTo.startsWith('/')
          ? this.data.roleComparisonReturnTo
          : `/${this.data.roleComparisonReturnTo}`
      })
    }
  },

  async loadHome() {
    try {
      const home = await homeService.getHome()
      this.setData({
        loading: false,
        user: home.user,
        hero: home.hero,
        notices: home.notices,
        quickActions: home.quickActions,
        recommendedGames: home.recommendedGames,
        playerSummary: home.playerSummary,
        rankingList: home.rankingList,
        achievementList: home.achievementList,
        friendGames: home.friendGames,
        metaverseEntry: home.metaverseEntry,
        nearbySummary: home.nearbySummary
      })
    } catch (error) {
      this.setData({
        loading: false
      })
      toast.info(error.message || '首页加载失败')
    }
  },

  goAction(event) {
    const route = event.currentTarget.dataset.route

    if (!route) {
      return
    }

    toast.developing()
  }
})

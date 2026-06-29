const homeService = require('../../services/home')
const { ROUTES } = require('../../config/routes')

const HOME_SCROLL_TAP_STEP_RPX = 360
const HOME_SCROLL_HOLD_STEP_RPX = 72
const HOME_SCROLL_HOLD_INTERVAL_MS = 80
const HOME_SCROLL_HOLD_SUPPRESS_TAP_MS = 120

const ROLE_TAGS = [
  { key: 'player', icon: '🎮', name: '玩家' },
  { key: 'expert', icon: '🎯', name: '行家' },
  { key: 'guide', icon: '🌐', name: '领路人' }
]

const ROLE_NAMES = {
  player: '玩家',
  expert: '行家',
  guide: '领路人'
}

const PLAYER_ROLE_HOME = {
  roleName: '玩家',
  roleEmoji: '🎮',
  deviceBadge: '⚡',
  profileName: 'Alex Chen',
  identity: '活跃达人',
  levelText: '玩家 Lv.5',
  scoreText: '580/1000 XP',
  progress: 58,
  nextLevelText: '距离下一等级还需 420 经验值',
  sectionTitle: '附近正在发生',
  sectionMore: '',
  stats: [
    { value: '12', label: '参与局数' },
    { value: '3', label: '本月MVP' },
    { value: '98%', label: '参与率' }
  ],
  entries: [
    { title: '发起组局', desc: '创建你的带局房间', icon: '📍', theme: 'pink', route: ROUTES.gameCreate },
    { title: '局前大厅', desc: '准备就绪加入一局', routeIcon: true, theme: 'cyan', route: ROUTES.gameHall }
  ],
  onlineCard: {
    title: '地球online',
    desc: '探索城市副本 · 解锁地图成就',
    tags: ['附近 12 个组局', '已打卡 8 处']
  },
  mainSection: {
    title: '附近正在发生',
    moreText: '',
    desc: ''
  },
  mainTabs: [
    { key: 'all', name: '全部', active: true },
    { key: 'nearby', name: '附近', active: false }
  ],
  mainGames: [],
  skills: [],
  review: null,
  recommendation: null,
  network: null
}

const ROLE_PERMISSION_PROMPTS = {
  expert: {
    roleType: 'expert',
    title: '我懂玩家需要什么！我申请成为行家',
    primary: '申请成为行家',
    secondary: '查看权益对比'
  },
  guide: {
    roleType: 'guide',
    title: '我愿意带领更多人一起玩！我申请成为领路人',
    primary: '申请成为领路人',
    secondary: '查看权益对比'
  }
}

const DEFAULT_ROLE_HOME = {
  roleName: '行家',
  roleEmoji: '🎯',
  profileName: '摄影咖 小李',
  identity: '认证DM',
  levelText: '行家 ⭐️',
  scoreText: '650/1000 XP',
  progress: 65,
  nextLevelText: '距离下一等级还需 350 经验值',
  stats: [
    { value: '12', label: '参与局数' },
    { value: '3', label: '本月MVP' },
    { value: '98%', label: '被选率' }
  ],
  entries: [
    { title: '发起组局', desc: '创建新的一局', icon: '📍', theme: 'pink', route: ROUTES.gameCreate },
    { title: '局前大厅', desc: '准备就绪，等待开局', routeIcon: true, theme: 'cyan', route: ROUTES.gameHall }
  ],
  onlineCard: {
    title: '地球online',
    desc: '探索城市副本 · 解锁地图成就',
    tags: ['附近 12 个组局', '已打卡 8 处']
  },
  mainSection: {
    title: '即将带局',
    moreText: '查看全部 →',
    desc: ''
  },
  mainTabs: [],
  mainGames: [],
  skills: [],
  review: null,
  recommendation: null,
  network: null
}

const GUIDE_ACTION_ENTRIES = [
  { id: 'lobby', title: '局前大厅', desc: '准备加入一局', routeIcon: true, theme: 'pink', route: 'pages/game/hall/index' },
  { id: 'invite', title: '我的邀约', desc: '管理连接的玩家', routeIcon: true, theme: 'cyan', route: 'pages/game/applications/index' }
]

const GUIDE_ROLE_HOME = {
  ...DEFAULT_ROLE_HOME,
  roleName: '领路人',
  roleEmoji: '🌐',
  profileName: '摄影咖 萧飒',
  identity: '百场辅助',
  levelText: '领路人 🐑',
  scoreText: '580/1000 XP',
  progress: 58,
  nextLevelText: '距离下一等级还需 420 经验值',
  stats: [
    { value: '99', label: '连接玩家' },
    { value: '99%', label: '玩家再玩率' },
    { value: '98%', label: '玩家完局率' }
  ],
  entries: GUIDE_ACTION_ENTRIES,
  mainSection: {
    title: '附近正在发生',
    moreText: '查看全部',
    desc: ''
  },
  mainTabs: [
    { key: 'all', name: '全部', active: true },
    { key: 'nearby', name: '附近', active: false }
  ],
  mainGames: [],
  recommendation: null,
  network: null
}

Component({
  options: {
    addGlobalClass: true,
    styleIsolation: 'shared'
  },

  properties: {
    roleType: {
      type: String,
      value: 'expert'
    }
  },

  data: {
    homeReady: false,
    onlineText: '3999人在线',
    homeScrollTop: 0,
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ],
    currentRoleType: 'expert',
    hero: {
      title: 'HELLO, 行家!',
      date: '2026.05.15 | 开启你的今日副本'
    },
    playerCard: {
      role: '行家 ⭐️',
      title: '认证DM',
      name: '摄影咖 小李',
      xp: '650/1000 XP',
      progress: 65,
      next: '距离下一等级还需 350 经验值',
      deviceIcon: '🎯',
      deviceBadge: '⚡',
      stats: DEFAULT_ROLE_HOME.stats
    },
    roleTags: ROLE_TAGS.map((item) => ({
      ...item,
      active: item.key === 'expert'
    })),
    selectedRoleTag: 'expert',
    rolePermissionPrompt: {
      visible: false,
      roleType: '',
      title: '',
      primary: '',
      secondary: ''
    },
    entries: DEFAULT_ROLE_HOME.entries,
    onlineCard: DEFAULT_ROLE_HOME.onlineCard,
    mainSection: DEFAULT_ROLE_HOME.mainSection,
    mainTabs: [],
    mainFilter: 'all',
    mainGames: [],
    skills: [],
    review: null,
    recommendation: null,
    network: null,
    rankingSection: {
      icon: '🏆',
      title: '本周玩霸榜',
      moreText: '查看全部榜单'
    },
    rankingActiveRole: 'expert',
    rankingTabs: [
      { key: 'player', name: '玩家', active: false },
      { key: 'expert', name: '行家', active: true },
      { key: 'guide', name: '领路人', active: false }
    ],
    rankingBoards: {},
    rankingList: [
      { rank: '01', avatarFallback: '👨🏾‍🎓', name: '领域专家 PRO', desc: '本周服务玩家 90 位', xp: '2,450', xpUnit: 'XP' },
      { rank: '02', avatarFallback: '👩🏻‍🎤', name: '社交达人', desc: '本周服务玩家 10 位', xp: '1,890', xpUnit: 'XP' },
      { rank: '03', avatarFallback: '👨🏿‍🚀', name: '探险家', desc: '本周服务玩家 1 位', xp: '1,560', xpUnit: 'XP' }
    ],
    myRank: {
      rank: '52',
      avatarFallback: '👩🏻‍💻',
      name: '我（Alex）',
      desc: '上周排名 65 ↑',
      xp: '520',
      xpUnit: 'XP'
    },
    showMyRank: true,
    achievementSection: {
      icon: '💎',
      title: '我的成就'
    },
    achievements: [
      { id: 'hundred', icon: '🏆', title: '百场王者', status: '▲ 等级', tone: 'gold', unlocked: true },
      { id: 'pilot', icon: '🏆', title: '引航王者', status: '▲ 等级', tone: 'gold', unlocked: true },
      { id: 'earth', icon: '🌍', title: '地球漫游者', status: '▲ 进度20%', tone: 'blue', unlocked: true, progressPercent: 20, showProgress: true },
      { id: 'hidden', icon: '🔒', title: '隐藏徽章', status: '▲ 未解锁', tone: 'locked', unlocked: false }
    ],
    metaverse: {
      title: '进入元宇宙',
      desc: '共创数字街区｜全球联机互动',
      tags: ['3D空间', 'NFT徽章'],
      avatars: [
        { text: '👨🏾‍🎓' },
        { text: '👩🏻‍🎤' },
        { text: '👨🏿‍🚀' }
      ],
      badge: '+99'
    }
  },

  lifetimes: {
    attached() {
      const roleType = this.normalizeRoleType(this.properties.roleType)

      this.applyRoleFallback(roleType)
      this.loadRoleHome(roleType)
    },

    detached() {
      this.clearHomeScrollTimers()
    }
  },

  observers: {
    roleType(value) {
      const roleType = this.normalizeRoleType(value)

      if (roleType === this.data.currentRoleType) {
        return
      }

      this.applyRoleFallback(roleType)
      this.loadRoleHome(roleType)
    }
  },

  methods: {
    applyRoleFallback(roleTypeValue) {
      const roleType = this.normalizeRoleType(roleTypeValue)
      const fallback = this.formatRoleDashboard({}, roleType)
      const rankingState = this.formatRankingState({}, roleType)

      this.setData({
        currentRoleType: roleType,
        homeReady: true,
        selectedRoleTag: roleType,
        hero: this.formatHero({}, fallback, roleType),
        playerCard: this.formatPlayerCard(fallback),
        roleTags: this.formatRoleTags(roleType),
        rolePermissionPrompt: this.formatRolePermissionPrompt('', roleType),
        entries: fallback.entries,
        onlineCard: fallback.onlineCard,
        mainSection: fallback.mainSection,
        mainTabs: fallback.mainTabs,
        mainFilter: fallback.mainTabs.length ? fallback.mainTabs[0].key : 'all',
        mainGames: fallback.mainGames,
        skills: fallback.skills,
        review: fallback.review,
        recommendation: fallback.recommendation,
        network: fallback.network,
        rankingActiveRole: rankingState.activeKey,
        rankingTabs: rankingState.tabs,
        rankingList: rankingState.list,
        myRank: rankingState.myRank,
        showMyRank: rankingState.showMyRank
      })
    },

    async loadRoleHome(roleTypeValue) {
      const roleType = this.normalizeRoleType(roleTypeValue || this.data.currentRoleType || this.properties.roleType)

      try {
        const home = await homeService.getHome({ roleType })
        const dashboard = this.formatRoleDashboard(home.roleDashboard || {}, roleType)
        const hero = home.hero || {}
        const rankingState = this.formatRankingState(home || {}, roleType)
        const achievementState = this.formatAchievementState(home || {})

        this.setData({
          currentRoleType: roleType,
          selectedRoleTag: roleType,
          onlineText: hero.onlineText || this.data.onlineText,
          hero: this.formatHero(hero, dashboard, roleType),
          playerCard: this.formatPlayerCard(dashboard),
          roleTags: this.formatRoleTags(roleType),
          rolePermissionPrompt: this.formatRolePermissionPrompt('', roleType),
          entries: dashboard.entries,
          onlineCard: dashboard.onlineCard,
          mainSection: dashboard.mainSection,
          mainTabs: dashboard.mainTabs,
          mainFilter: dashboard.mainTabs.length ? dashboard.mainTabs[0].key : 'all',
          mainGames: dashboard.mainGames,
          skills: dashboard.skills,
          review: dashboard.review,
          recommendation: dashboard.recommendation,
          network: dashboard.network,
          rankingSection: rankingState.section,
          rankingActiveRole: rankingState.activeKey,
          rankingTabs: rankingState.tabs,
          rankingBoards: rankingState.boards,
          rankingList: rankingState.list,
          myRank: rankingState.myRank,
          achievementSection: achievementState.section,
          achievements: achievementState.list,
          metaverse: this.formatMetaverseEntry(home.metaverseEntry || home.metaverse || {})
        })
      } catch (error) {
        const fallback = this.formatRoleDashboard({}, roleType)

        this.setData({
          selectedRoleTag: roleType,
          hero: this.formatHero({}, fallback, roleType),
          playerCard: this.formatPlayerCard(fallback),
          roleTags: this.formatRoleTags(roleType),
          rolePermissionPrompt: this.formatRolePermissionPrompt('', roleType),
          entries: fallback.entries,
          onlineCard: fallback.onlineCard,
          mainSection: fallback.mainSection,
          mainTabs: fallback.mainTabs,
          mainGames: fallback.mainGames,
          skills: fallback.skills,
          review: fallback.review,
          recommendation: fallback.recommendation,
          network: fallback.network
        })
      }
    },

    formatRoleDashboard(dashboard = {}, roleType = 'expert') {
      const roleName = ROLE_NAMES[roleType] || '行家'
      const defaultHome = this.getDefaultRoleHome(roleType)
      const source = {
        ...defaultHome,
        ...dashboard,
        roleName: dashboard.roleName || roleName
      }

      return {
        ...source,
        stats: Array.isArray(dashboard.stats) && dashboard.stats.length ? dashboard.stats : source.stats,
        entries: Array.isArray(dashboard.entries) && dashboard.entries.length
          ? dashboard.entries
          : this.formatEntries(dashboard.quickActions || source.entries, roleType),
        onlineCard: this.formatRoleOnlineCard(dashboard.onlineCard || source.onlineCard),
        mainSection: this.formatMainSection(dashboard, source),
        mainTabs: this.formatMainTabs(dashboard.mainTabs || dashboard.filterTabs || source.mainTabs),
        mainGames: this.formatGameCards(dashboard.mainGames || dashboard.events || source.mainGames),
        skills: Array.isArray(dashboard.skills) ? dashboard.skills : source.skills,
        review: this.formatReview(dashboard, source),
        recommendation: dashboard.recommendation || source.recommendation,
        network: this.formatNetwork(dashboard.network || source.network)
      }
    },

    getDefaultRoleHome(roleType) {
      if (roleType === 'player') {
        return PLAYER_ROLE_HOME
      }

      if (roleType === 'guide') {
        return GUIDE_ROLE_HOME
      }

      return DEFAULT_ROLE_HOME
    },

    formatNetwork(network) {
      if (!network) {
        return null
      }

      const summary = this.pickFirstValue(
        network.summary,
        network.summaryText,
        this.formatConnectedSummary(network.connectedCount || network.playerCount)
      )
      const items = Array.isArray(network.items) && network.items.length
        ? network.items
        : [
          { id: 'connected', icon: '●', name: '已连接', desc: summary || '156 位玩家' },
          { id: 'income', icon: '¥', name: '本周收益', desc: network.income || '¥1,240' },
          { id: 'location', icon: '📍', name: '核心区域', desc: network.location || '镇海区' }
        ]
      const buttons = Array.isArray(network.buttons) && network.buttons.length
        ? network.buttons
        : [
          { text: network.actionText || network.moreText || '查看全部', primary: true, route: network.route },
          { text: network.secondaryText || '管理连接', route: network.manageRoute }
        ]

      return {
        ...network,
        title: network.title || '我的关系网络',
        status: network.status || '实时连接中',
        hubTitle: network.hubTitle || network.centerTitle || '我',
        hubDesc: network.hubDesc || network.centerDesc || '领路人',
        summary,
        income: network.income || '本周收益 ¥1,240',
        location: network.location || '📍 镇海区',
        locationText: network.locationText || String(network.location || '镇海区').replace(/^📍\s*/, ''),
        items: items.slice(0, 5),
        buttons: buttons.slice(0, 2)
      }
    },

    formatConnectedSummary(count) {
      if (count == null || count === '') {
        return ''
      }

      return `● 已连接 ${count} 位玩家`
    },

    formatReview(dashboard = {}, source = DEFAULT_ROLE_HOME) {
      const sourceReview = source.review || null
      const review = dashboard.review || sourceReview

      if (!review) {
        return null
      }

      const latestReview = this.pickLatestReview(review)
      const count = this.formatStatValue(this.pickFirstValue(
        review.count,
        review.total,
        review.totalCount,
        review.reviewCount,
        dashboard.reviewCount,
        dashboard.reviewTotal,
        dashboard.reviewTotalCount
      ))
      const moreText = count ? `全部${count}条` : this.normalizeReviewMoreText(review.moreText)

      return {
        ...review,
        ...latestReview,
        count,
        moreText,
        avatarUrl: this.pickFirstValue(
          latestReview.avatarUrl,
          latestReview.avatar,
          latestReview.userAvatar,
          latestReview.playerAvatar,
          review.avatarUrl,
          review.avatar,
          review.userAvatar,
          review.playerAvatar
        ),
        avatarFallback: this.formatReviewAvatarFallback(latestReview, review),
        playerLevel: this.pickFirstValue(
          latestReview.playerLevel,
          latestReview.levelText,
          latestReview.userLevel,
          latestReview.authorLevel,
          latestReview.author,
          latestReview.playerName,
          latestReview.nickname,
          review.playerLevel,
          review.levelText,
          review.userLevel,
          review.authorLevel,
          review.author,
          review.playerName,
          review.nickname,
          '萌新玩家'
        ),
        timeText: this.pickFirstValue(
          latestReview.timeText,
          latestReview.createdAtText,
          latestReview.reviewTimeText,
          latestReview.createdAt,
          review.timeText,
          review.createdAtText,
          review.reviewTimeText,
          review.createdAt
        ),
        ratingText: this.formatReviewRating(this.pickFirstValue(
          latestReview.rating,
          latestReview.score,
          latestReview.star,
          latestReview.stars,
          review.rating,
          review.score,
          review.star,
          review.stars
        )),
        content: this.pickFirstValue(
          latestReview.content,
          latestReview.comment,
          latestReview.text,
          latestReview.reviewContent,
          review.content,
          review.comment,
          review.text,
          review.reviewContent
        ),
        route: review.route || dashboard.reviewRoute || 'pages/game/review/index'
      }
    },

    pickLatestReview(review = {}) {
      const list = Array.isArray(review.items) && review.items.length
        ? review.items
        : Array.isArray(review.list) && review.list.length
          ? review.list
          : Array.isArray(review.reviews) && review.reviews.length
            ? review.reviews
            : []

      return review.latest || review.latestReview || list[0] || {}
    },

    normalizeReviewMoreText(moreText = '') {
      return String(moreText || '').replace(/\s*→$/, '').trim()
    },

    formatReviewAvatarFallback(latestReview = {}, review = {}) {
      return this.pickFirstValue(
        latestReview.avatarFallback,
        latestReview.avatarText,
        review.avatarFallback,
        review.avatarText,
        '🎮'
      )
    },

    formatReviewRating(value) {
      if (value == null || value === '') {
        return ''
      }

      if (typeof value === 'string' && value.includes('★')) {
        return value
      }

      const count = Number(value)

      if (!Number.isFinite(count) || count <= 0) {
        return String(value)
      }

      return '★'.repeat(Math.max(0, Math.min(5, Math.round(count))))
    },

    formatMainSection(dashboard = {}, source = DEFAULT_ROLE_HOME) {
      const section = dashboard.mainSection || {}
      const fallback = source.mainSection || DEFAULT_ROLE_HOME.mainSection
      const count = this.formatStatValue(this.pickFirstValue(
        section.count,
        dashboard.sectionCount,
        dashboard.count,
        fallback.count
      ))
      const numericCount = Number(count)
      const hasCount = count !== '' && Number.isFinite(numericCount)
      const moreEnabled = hasCount && numericCount > 2
      const moreText = section.moreText || dashboard.sectionMore || fallback.moreText
      const normalizedMoreText = moreText.replace(/\s*→$/, '')

      return {
        ...fallback,
        ...section,
        title: section.title || dashboard.sectionTitle || fallback.title,
        moreText: moreEnabled ? `${normalizedMoreText} →` : normalizedMoreText,
        count,
        desc: section.desc || dashboard.sectionDesc || fallback.desc,
        moreEnabled
      }
    },

    formatHero(hero = {}, dashboard = {}, roleType = 'expert') {
      const roleName = dashboard.roleName || hero.roleName || ROLE_NAMES[roleType] || '行家'
      const date = [hero.dateLabel, hero.subtitle]
        .filter((item) => typeof item === 'string' && item.trim())
        .join(' | ')

      return {
        title: hero.title || `HELLO, ${roleName}!`,
        date: date || '2026.05.15 | 开启你的今日副本'
      }
    },

    formatPlayerCard(dashboard) {
      return {
        role: dashboard.levelText || dashboard.roleName,
        title: dashboard.identity || dashboard.profileTitle || '',
        name: dashboard.profileName || 'Alex Chen',
        xp: dashboard.xpText || this.formatXpText(dashboard) || dashboard.scoreText || '650/1000 XP',
        progress: this.normalizeProgress(
          this.pickFirstValue(dashboard.progressPercent, dashboard.progress),
          65
        ),
        next: this.formatRoleNextLevelText(dashboard),
        deviceIcon: dashboard.roleEmoji || '🎯',
        deviceBadge: dashboard.deviceBadge || '⚡',
        stats: Array.isArray(dashboard.stats) && dashboard.stats.length ? dashboard.stats.slice(0, 3) : DEFAULT_ROLE_HOME.stats
      }
    },

    formatRoleNextLevelText(dashboard = {}) {
      if (dashboard.nextLevelText) {
        return dashboard.nextLevelText
      }

      const scoreText = dashboard.xpText || this.formatXpText(dashboard) || dashboard.scoreText || ''
      const scoreMatch = String(scoreText).match(/(\d+)\s*\/\s*(\d+)/)

      if (scoreMatch) {
        const current = Number(scoreMatch[1])
        const target = Number(scoreMatch[2])
        const remain = target - current

        if (Number.isFinite(remain) && remain >= 0) {
          return `距离下一等级还需 ${remain} 经验值`
        }
      }

      return dashboard.nextLevelText || '距离下一等级还需 350 经验值'
    },

    formatXpText(dashboard = {}) {
      const experience = dashboard.experience
      const nextLevelExperience = dashboard.nextLevelExperience

      if (experience == null || nextLevelExperience == null) {
        return ''
      }

      return `${experience}/${nextLevelExperience} XP`
    },

    formatEntries(entries, roleType) {
      const source = Array.isArray(entries) && entries.length ? entries : DEFAULT_ROLE_HOME.entries
      const normalizedSource = roleType === 'guide'
        ? this.normalizeGuideEntries(source)
        : source

      return normalizedSource.slice(0, 2).map((item, index) => ({
        title: item.title,
        desc: item.desc,
        icon: item.icon,
        routeIcon: Boolean(item.routeIcon || (roleType === 'expert' && index === 1)),
        theme: item.theme || (index % 2 === 0 ? 'pink' : 'cyan'),
        route: item.route
      }))
    },

    normalizeGuideEntries(entries) {
      const source = Array.isArray(entries) ? entries : []
      const hasExpectedEntries = source.some((item) => item && (item.title === '局前大厅' || item.title === '我的邀约'))

      if (hasExpectedEntries) {
        return source
      }

      return GUIDE_ACTION_ENTRIES
    },

    formatRoleOnlineCard(onlineCard = {}) {
      const ownedGameCount = this.formatStatValue(
        this.pickFirstValue(onlineCard.ownedGameCount, onlineCard.myGameCount, onlineCard.gameCount)
      )
      const checkinCount = this.formatStatValue(
        this.pickFirstValue(onlineCard.checkinCount, onlineCard.checkedInCount, onlineCard.checkInCount)
      )
      const tags = Array.isArray(onlineCard.tags) && onlineCard.tags.length
        ? onlineCard.tags
        : [
          `我的组局 ${ownedGameCount || '0'} 个`,
          `已打卡 ${checkinCount || '0'} 处`
        ]

      return {
        title: onlineCard.title || '地球online',
        desc: onlineCard.desc || '管理我的组局足迹与城市打卡',
        tags
      }
    },

    formatMainTabs(tabs) {
      if (!Array.isArray(tabs) || !tabs.length) {
        return []
      }

      return tabs.map((item, index) => ({
        key: item.key || item.value || (index === 0 ? 'all' : `tab_${index}`),
        name: item.name || item.label || item.title,
        active: index === 0 || Boolean(item.active)
      }))
    },

    formatGameCards(games) {
      if (!Array.isArray(games) || !games.length) {
        return []
      }

      return games.map((item, index) => ({
        id: item.id,
        scope: item.scope || item.distanceScope || (index === 0 ? 'nearby' : 'city'),
        title: item.title,
        startTime: item.time || item.startTime || '',
        day: item.day || item.dayText || '',
        date: item.dateText || item.timeText || item.date || '',
        venue: item.locationText || item.location || this.formatExpertVenue(item.meta),
        members: item.memberText || item.members || this.formatExpertMembers(item.meta),
        tag: this.formatGameTag(item),
        sessionTags: this.formatSessionTags(item),
        coverSrc: item.coverUrl || item.coverSrc || item.cover || (index % 2 ? '/components/game-card/assets/cover-sunset.png' : '/components/game-card/assets/cover-city.png'),
        price: item.priceText || item.price || item.income || '',
        income: item.income || item.priceText || item.price || '',
        action: item.actionText || item.action || '查看',
        location: item.locationText || this.formatGameLocation(item) || item.meta || item.location || '',
        time: item.timeText || item.dateText || item.time || '',
        joinedText: item.joinedText || item.peopleText || this.formatJoinedText(item.joinedCount) || '+3位玩家已入局',
        playerAvatars: this.formatSessionAvatars(item),
        actions: this.formatGameActions(item.actions)
      }))
    },

    formatSessionTags(item = {}) {
      const tags = Array.isArray(item.tags) ? item.tags : []
      const typeTag = item.typeText || item.gameTypeText || item.categoryText || tags[0] || ''
      const statusTag = item.statusText || item.status || item.tagText || tags[1] || ''

      return [typeTag, statusTag].filter(Boolean).slice(0, 2)
    },

    formatGameLocation(item = {}) {
      const parts = [item.cityName || item.locationName, item.distanceText, item.memberText]
        .filter((value) => value != null && value !== '')

      return parts.length ? `📍${parts.join(' · ')}` : ''
    },

    formatSessionAvatars(item = {}) {
      const source = item.avatarUrls || item.participantAvatars || item.playerAvatars || item.players || item.participants || []

      if (!Array.isArray(source) || source.length === 0) {
        return [
          { text: '🙋‍♀️' },
          { text: '🧑‍💼' },
          { text: '👱' }
        ]
      }

      return source.slice(0, 3).map((avatar, index) => {
        if (typeof avatar === 'string') {
          return {
            imageUrl: avatar,
            text: `${index + 1}`
          }
        }

        return {
          imageUrl: avatar.avatarUrl || avatar.imageUrl || avatar.url || '',
          text: avatar.avatarFallback || avatar.nickname || avatar.name || `${index + 1}`
        }
      })
    },

    formatExpertVenue(meta) {
      if (typeof meta !== 'string') {
        return ''
      }

      const parts = meta.split('·').map((item) => item.trim()).filter(Boolean)
      return parts.slice(0, 2).join(' · ')
    },

    formatExpertMembers(meta) {
      if (typeof meta !== 'string') {
        return ''
      }

      const parts = meta.split('·').map((item) => item.trim())
      return parts[2] || parts[parts.length - 1] || ''
    },

    formatGameTag(item = {}) {
      const primaryTag = item.typeText || item.statusText || item.tagText || item.tag || item.status || item.typeName

      if (primaryTag) {
        return primaryTag
      }

      if (Array.isArray(item.tags) && item.tags.length) {
        return item.tags[0]
      }

      return ''
    },

    formatJoinedText(joinedCount) {
      if (joinedCount == null || joinedCount === '') {
        return ''
      }

      return `+${joinedCount}位玩家已入局`
    },

    formatGameActions(actions) {
      const actionMap = {
        share: '分享',
        follow: '关注',
        refer: '引荐',
        greet: '打招呼'
      }

      if (!Array.isArray(actions) || actions.length === 0) {
        return ['分享', '关注', '引荐', '打招呼']
      }

      return actions.map((action) => actionMap[action] || action)
    },

    formatMetaverseEntry(entry = {}) {
      const current = this.data.metaverse

      return {
        ...current,
        title: entry.actionText || entry.title || current.title,
        desc: entry.summary || entry.description || entry.desc || current.desc,
        tags: Array.isArray(entry.tags) && entry.tags.length ? entry.tags.slice(0, 2) : current.tags,
        avatars: this.formatMetaverseAvatars(entry.avatars || entry.users || current.avatars),
        badge: this.formatMetaverseBadge(entry, current.badge)
      }
    },

    formatMetaverseAvatars(avatars) {
      if (!Array.isArray(avatars) || avatars.length === 0) {
        return this.data.metaverse.avatars
      }

      return avatars.slice(0, 3).map((item, index) => {
        if (typeof item === 'string') {
          return { text: item }
        }

        return {
          text: item.text || item.avatarFallback || item.initial || item.nickname || `${index + 1}`,
          imageUrl: item.imageUrl || item.avatarUrl || item.url || ''
        }
      })
    },

    formatMetaverseBadge(entry = {}, fallback = '') {
      const joinedCount = this.pickFirstValue(entry.joinedCount, entry.onlineCount, entry.participantCount)

      if (joinedCount != null && joinedCount !== '') {
        return `+${joinedCount}`
      }

      return entry.badge || entry.badgeText || entry.onlineText || fallback
    },

    formatRankingState(home = {}, roleType = 'expert') {
      const section = home.rankingSection || {}
      const requestedActiveKey = this.normalizeRoleType(section.defaultTab || section.activeKey || home.rankingActiveRole || roleType)
      const tabs = this.formatRankingTabs(section.tabs || home.rankingTabs, requestedActiveKey)
      const boards = this.formatRankingBoards(home, this.data.rankingBoards)
      const activeKey = requestedActiveKey || roleType
      const activeBoard = boards[activeKey] || {}
      const display = this.formatRankingDisplay(
        activeBoard.list || this.data.rankingList,
        activeBoard.myRank || this.data.myRank
      )

      return {
        section: {
          ...this.data.rankingSection,
          icon: section.icon || this.data.rankingSection.icon,
          title: section.title || section.name || this.data.rankingSection.title,
          moreText: section.moreText || section.moreLabel || section.actionText || this.data.rankingSection.moreText
        },
        activeKey,
        tabs: tabs.map((item) => ({
          ...item,
          active: item.key === activeKey
        })),
        boards,
        list: display.list,
        myRank: display.myRank,
        showMyRank: display.showMyRank
      }
    },

    formatRankingTabs(tabs, activeKey) {
      const sourceTabs = Array.isArray(tabs) && tabs.length > 0 ? tabs : this.data.rankingTabs

      return sourceTabs.map((item, index) => {
        const fallbackKey = ROLE_TAGS[index] ? ROLE_TAGS[index].key : `tab_${index}`
        const key = this.normalizeRoleType(item.key || item.value || item.roleType || item.name || fallbackKey)

        return {
          key,
          name: item.name || item.label || item.title || key,
          active: key === activeKey
        }
      })
    },

    formatRankingBoards(home = {}, fallbackBoards = {}) {
      const sourceBoards = home.rankingBoards || home.rankings || home.rankingByRole
      const boards = {}

      if (sourceBoards && typeof sourceBoards === 'object' && !Array.isArray(sourceBoards)) {
        Object.keys(sourceBoards).forEach((key) => {
          const roleKey = this.normalizeRoleType(key)
          boards[roleKey] = this.formatRankingBoard(sourceBoards[key], fallbackBoards[roleKey])
        })
      }

      Object.keys(fallbackBoards || {}).forEach((key) => {
        if (!boards[key]) {
          boards[key] = fallbackBoards[key]
        }
      })

      return boards
    },

    formatRankingBoard(board, fallback = {}) {
      const sourceList = Array.isArray(board)
        ? board
        : (board && (board.list || board.rankingList || board.items || board.records))
      const sourceMyRank = board && !Array.isArray(board)
        ? board.myRank || board.currentUserRank || board.me
        : null

      return {
        list: this.formatRankingList(sourceList, fallback.list || []),
        myRank: this.formatRankingItem(sourceMyRank, fallback.myRank || this.data.myRank)
      }
    },

    formatRankingList(list, fallbackList) {
      if (!Array.isArray(list) || list.length === 0) {
        return fallbackList
      }

      return list.map((item, index) => this.formatRankingItem(item, fallbackList[index]))
    },

    formatRankingDisplay(list = [], myRank = {}) {
      const rankingList = Array.isArray(list) ? list : []
      const currentRank = myRank || {}

      if (!this.shouldInlineMyRank(rankingList, currentRank)) {
        return {
          list: rankingList,
          myRank: currentRank,
          showMyRank: true
        }
      }

      return {
        list: this.mergeMyRankIntoRankingList(rankingList, currentRank).slice(0, 4),
        myRank: currentRank,
        showMyRank: false
      }
    },

    shouldInlineMyRank(list = [], myRank = {}) {
      const rankNumber = this.getRankNumber(myRank.rank)

      return Array.isArray(list) && list.length > 0 && rankNumber > 0 && rankNumber <= 3
    },

    mergeMyRankIntoRankingList(list = [], myRank = {}) {
      const hasCurrentUser = list.some((item) => this.isSameRankingUser(item, myRank))
      const merged = hasCurrentUser ? list : list.concat(myRank)

      return merged.slice().sort((left, right) => {
        const leftRank = this.getRankNumber(left.rank)
        const rightRank = this.getRankNumber(right.rank)

        return (leftRank || 9999) - (rightRank || 9999)
      })
    },

    isSameRankingUser(item = {}, myRank = {}) {
      if (item.isMe || item.isSelf || item.isCurrentUser) {
        return true
      }

      const idKeys = ['id', 'userId', 'memberId', 'profileId', 'openId']
      const hasSameId = idKeys.some((key) => item[key] && myRank[key] && `${item[key]}` === `${myRank[key]}`)

      if (hasSameId) {
        return true
      }

      return item.name && myRank.name && item.rank && myRank.rank && item.name === myRank.name && item.rank === myRank.rank
    },

    formatRankingItem(item, fallback = {}) {
      if (!item || Object.keys(item).length === 0) {
        return fallback
      }

      const avatarCandidate = item.avatarUrl || item.avatarSrc || item.avatarImage || item.avatar
      const avatarUrl = this.isImagePath(avatarCandidate) ? avatarCandidate : fallback.avatarUrl || ''
      const name = item.name || item.nickname || item.displayName || fallback.name || ''

      return {
        id: item.id || fallback.id,
        rank: this.formatRankNo(this.pickFirstValue(item.rank, item.rankNo, item.position, fallback.rank)),
        avatarUrl,
        avatarFallback: item.avatarFallback || item.initials || (!avatarUrl && avatarCandidate) || this.makeAvatarFallback(name) || fallback.avatarFallback,
        name,
        desc: item.desc || item.description || item.summary || item.rankText || fallback.desc || '',
        xp: this.formatRankingXp(this.pickFirstValue(item.xp, item.xpText, item.experience, item.weeklyXp, fallback.xp)),
        xpUnit: item.xpUnit || fallback.xpUnit || 'XP'
      }
    },

    formatRankNo(rank) {
      if (rank == null || rank === '') {
        return ''
      }

      const numericRank = Number(rank)

      if (Number.isInteger(numericRank) && numericRank > 0 && numericRank < 10) {
        return `0${numericRank}`
      }

      return `${rank}`
    },

    getRankNumber(rank) {
      if (rank == null || rank === '') {
        return 0
      }

      const numericRank = Number(rank)

      return Number.isFinite(numericRank) ? numericRank : 0
    },

    formatRankingXp(xp) {
      if (xp == null || xp === '') {
        return ''
      }

      return `${xp}`.replace(/\s*XP$/i, '').trim()
    },

    formatAchievementState(source = {}) {
      const section = source.achievementSection || source.achievementsSection || {}
      const list = source.achievements || source.achievementList || source.badges || []

      return {
        section: {
          ...this.data.achievementSection,
          icon: section.icon || this.data.achievementSection.icon,
          title: section.title || section.name || this.data.achievementSection.title
        },
        list: this.formatAchievements(list, this.data.achievements)
      }
    },

    formatAchievements(list, fallbackList) {
      if (!Array.isArray(list) || list.length === 0) {
        return fallbackList
      }

      return list
        .map((item, index) => this.formatAchievement(item, fallbackList[index]))
        .sort((prev, next) => Number(prev.unlocked === false) - Number(next.unlocked === false))
    },

    formatAchievement(item, fallback = {}) {
      const title = item.title || item.name || item.displayName || fallback.title || ''
      const status = item.statusText || item.statusLabel || item.status || this.formatAchievementStatus(item) || fallback.status || ''
      const unlocked = item.unlocked != null ? Boolean(item.unlocked) : fallback.unlocked !== false
      const progressPercent = this.normalizeProgress(
        this.pickFirstValue(item.progressPercent, item.progress),
        fallback.progressPercent || 0
      )

      return {
        id: item.id || item.code || fallback.id || title,
        icon: item.icon || item.iconText || fallback.icon || '🏆',
        title,
        status: status ? (status.startsWith('▲') ? status : `▲ ${status}`) : '',
        tone: this.resolveAchievementTone(item, fallback, unlocked),
        unlocked,
        progressPercent,
        showProgress: progressPercent > 0 && unlocked
      }
    },

    resolveAchievementTone(item, fallback = {}, unlocked = true) {
      if (!unlocked) {
        return 'locked'
      }

      const text = `${item.code || ''} ${item.title || item.name || ''}`

      if (item.progressPercent != null || item.progress != null || /earth|地球/.test(text)) {
        return 'blue'
      }

      return fallback.tone || 'gold'
    },

    formatAchievementStatus(item) {
      const progress = this.pickFirstValue(item.progressPercent, item.progress)

      if (progress != null && progress !== '') {
        return `进度${progress}%`
      }

      if (item.unlocked === false) {
        return '未解锁'
      }

      return ''
    },

    normalizeRoleType(value) {
      if (value === 'guide' || value === '领路人') {
        return 'guide'
      }

      if (value === 'player' || value === '玩家') {
        return 'player'
      }

      return 'expert'
    },

    formatRoleTags(activeRole) {
      return ROLE_TAGS.map((item) => ({
        ...item,
        active: item.key === activeRole
      }))
    },

    formatRolePermissionPrompt(roleType, currentRoleType = this.data.currentRoleType) {
      if (!roleType || roleType === currentRoleType || roleType === 'player') {
        return {
          visible: false,
          roleType: '',
          title: '',
          primary: '',
          secondary: ''
        }
      }

      const prompt = ROLE_PERMISSION_PROMPTS[roleType]

      if (!prompt) {
        return {
          visible: false,
          roleType: '',
          title: '',
          primary: '',
          secondary: ''
        }
      }

      return {
        ...prompt,
        visible: true
      }
    },

    pickFirstValue(...values) {
      return values.find((value) => value != null && value !== '')
    },

    formatStatValue(value) {
      if (value == null || value === '') {
        return ''
      }

      return `${value}`
    },

    normalizeProgress(progress, fallback) {
      const percent = Number(progress)

      if (!Number.isFinite(percent)) {
        return fallback
      }

      return Math.max(0, Math.min(100, percent))
    },

    isImagePath(value) {
      return typeof value === 'string' && (/^(https?:)?\/\//.test(value) || value.startsWith('/'))
    },

    makeAvatarFallback(name) {
      if (!name) {
        return ''
      }

      return name.trim().slice(0, 2)
    },

    handleShellNavTap(event) {
      const key = event.detail && event.detail.key

      if (key === 'up' || key === 'down') {
        if (!this.suppressNextNavTap) {
          this.scrollRoleHome(key, HOME_SCROLL_TAP_STEP_RPX)
        }
        return
      }

      if (key === 'home') {
        this.scrollRoleHomeToTop()
        return
      }

      wx.showToast({
        title: '功能正在开发中',
        icon: 'none'
      })
    },

    handleShellNavLongPress(event) {
      const key = event.detail && event.detail.key

      if (key !== 'up' && key !== 'down') {
        return
      }

      this.suppressNextNavTap = true
      this.stopHomeScrollHold(false)
      this.scrollRoleHome(key, HOME_SCROLL_HOLD_STEP_RPX)

      this.homeScrollHoldTimer = setInterval(() => {
        this.scrollRoleHome(key, HOME_SCROLL_HOLD_STEP_RPX)
      }, HOME_SCROLL_HOLD_INTERVAL_MS)
    },

    handleShellNavTouchEnd() {
      this.stopHomeScrollHold(true)
    },

    handleHomeScroll(event) {
      const scrollTop = event.detail && event.detail.scrollTop

      if (typeof scrollTop === 'number') {
        this.homeScrollTopValue = scrollTop
      }
    },

    scrollRoleHome(direction, stepRpx = HOME_SCROLL_TAP_STEP_RPX) {
      const current = Number(this.homeScrollTopValue || this.data.homeScrollTop || 0)
      const distance = this.rpxToPx(stepRpx)
      const nextTop = direction === 'up'
        ? Math.max(0, current - distance)
        : current + distance

      this.homeScrollTopValue = nextTop
      this.setData({
        homeScrollTop: nextTop
      })
    },

    scrollRoleHomeToTop() {
      this.homeScrollTopValue = 0
      this.setData({
        homeScrollTop: 0
      })
    },

    stopHomeScrollHold(resetTapSuppress) {
      if (this.homeScrollHoldTimer) {
        clearInterval(this.homeScrollHoldTimer)
        this.homeScrollHoldTimer = null
      }

      if (resetTapSuppress && this.suppressNextNavTap) {
        if (this.homeScrollSuppressTimer) {
          clearTimeout(this.homeScrollSuppressTimer)
        }

        this.homeScrollSuppressTimer = setTimeout(() => {
          this.suppressNextNavTap = false
          this.homeScrollSuppressTimer = null
        }, HOME_SCROLL_HOLD_SUPPRESS_TAP_MS)
      }
    },

    clearHomeScrollTimers() {
      this.stopHomeScrollHold(false)

      if (this.homeScrollSuppressTimer) {
        clearTimeout(this.homeScrollSuppressTimer)
        this.homeScrollSuppressTimer = null
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

    handleMainTabTap(event) {
      const key = event.currentTarget.dataset.key || 'all'

      this.setData({
        mainFilter: key,
        mainTabs: this.data.mainTabs.map((item) => ({
          ...item,
          active: item.key === key
        }))
      })
    },

    handleRoleTagTap(event) {
      const key = this.normalizeRoleType(event.currentTarget.dataset.key || this.data.currentRoleType)
      const currentRoleType = this.data.currentRoleType

      if (key === 'player') {
        this.setData({
          selectedRoleTag: currentRoleType,
          roleTags: this.formatRoleTags(currentRoleType),
          rolePermissionPrompt: this.formatRolePermissionPrompt('', currentRoleType)
        })
        return
      }

      if (currentRoleType === 'guide' && key === 'expert') {
        this.setData({
          selectedRoleTag: currentRoleType,
          roleTags: this.formatRoleTags(currentRoleType),
          rolePermissionPrompt: this.formatRolePermissionPrompt('', currentRoleType)
        })
        return
      }

      this.setData({
        selectedRoleTag: key,
        roleTags: this.formatRoleTags(key),
        rolePermissionPrompt: this.formatRolePermissionPrompt(key)
      })
    },

    handleRoleApplyTap() {
      const prompt = this.data.rolePermissionPrompt || {}
      const roleType = prompt.roleType || this.data.selectedRoleTag
      const normalizedRoleType = this.normalizeRoleType(roleType)
      let url = `/${ROUTES.roleApply}?roleType=${roleType}`

      if (normalizedRoleType === 'expert') {
        const returnRoute = this.data.currentRoleType === 'guide' ? ROUTES.guideHome : ROUTES.expertHome
        url = `/${ROUTES.home}?ui=1&mode=expertApplyOverview&single=1&returnTo=${encodeURIComponent(returnRoute)}`
      } else if (normalizedRoleType === 'guide') {
        url = `/${ROUTES.homeOther}?page=guideApply&single=1&roleType=guide`
      }

      if (normalizedRoleType === 'expert' && typeof wx.redirectTo === 'function') {
        wx.redirectTo({ url })
        return
      }

      if (typeof wx.navigateTo === 'function') {
        wx.navigateTo({ url })
        return
      }

      wx.showToast({
        title: '功能正在开发中',
        icon: 'none'
      })
    },

    handleRoleCompareTap() {
      if (typeof wx.navigateTo === 'function') {
        const returnRoute = this.data.currentRoleType === 'guide' ? ROUTES.guideHome : ROUTES.expertHome

        wx.navigateTo({
          url: `/${ROUTES.home}?ui=1&mode=roleComparison&single=1&returnTo=${encodeURIComponent(returnRoute)}`
        })
        return
      }

      wx.showToast({
        title: '权益对比正在开发中',
        icon: 'none'
      })
    },

    handleRankingTabTap(event) {
      const key = this.normalizeRoleType(event.currentTarget.dataset.key || this.data.rankingActiveRole)
      const board = this.data.rankingBoards[key] || {}
      const patch = {
        rankingActiveRole: key,
        rankingTabs: this.data.rankingTabs.map((item) => ({
          ...item,
          active: item.key === key
        }))
      }

      if (board.list || board.myRank) {
        const display = this.formatRankingDisplay(board.list || this.data.rankingList, board.myRank || this.data.myRank)

        patch.rankingList = display.list
        patch.myRank = display.myRank
        patch.showMyRank = display.showMyRank
      }

      this.setData(patch)
    },

    handleReviewMoreTap() {
      const route = this.data.review && this.data.review.route

      if (route && typeof wx.navigateTo === 'function') {
        wx.navigateTo({
          url: route.startsWith('/') ? route : `/${route}`
        })
        return
      }

      wx.showToast({
        title: '功能正在开发中',
        icon: 'none'
      })
    },

    handleActionTap(event) {
      const route = event && event.currentTarget && event.currentTarget.dataset && event.currentTarget.dataset.route

      if (route && typeof wx.navigateTo === 'function') {
        wx.navigateTo({
          url: route.startsWith('/') ? route : `/${route}`
        })
        return
      }

      wx.showToast({
        title: '功能正在开发中',
        icon: 'none'
      })
    }
  }
})

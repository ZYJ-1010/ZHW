const homeService = require('../../services/home')
const roleService = require('../../services/role')
const { ROUTES } = require('../../config/routes')
const { navigateShellKey, navigateShellRoute } = require('../../utils/shell-nav')

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
  deviceBadge: '',
  profileName: '',
  identity: '',
  levelText: '',
  scoreText: '',
  progress: 0,
  nextLevelText: '',
  sectionTitle: '',
  sectionMore: '',
  stats: [],
  entries: [],
  onlineCard: {
    title: '',
    desc: '',
    tags: []
  },
  mainSection: {
    title: '',
    moreText: '',
    desc: ''
  },
  mainTabs: [],
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

const EMPTY_ROLE_HOME = {
  roleName: '行家',
  roleEmoji: '🎯',
  profileName: '',
  identity: '',
  levelText: '',
  scoreText: '',
  progress: 0,
  nextLevelText: '',
  stats: [],
  entries: [],
  onlineCard: {
    title: '',
    desc: '',
    tags: []
  },
  mainSection: {
    title: '',
    moreText: '',
    desc: ''
  },
  mainTabs: [],
  mainGames: [],
  skills: [],
  review: null,
  recommendation: null,
  network: null
}

const GUIDE_ACTION_ENTRIES = []

const GUIDE_ROLE_HOME = {
  ...EMPTY_ROLE_HOME,
  roleName: '领路人',
  roleEmoji: '🌐',
  profileName: '',
  identity: '',
  levelText: '',
  scoreText: '',
  progress: 0,
  nextLevelText: '',
  stats: [],
  entries: GUIDE_ACTION_ENTRIES,
  mainSection: {
    title: '',
    moreText: '',
    desc: ''
  },
  mainTabs: [],
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
    onlineText: '在线',
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
      title: '',
      date: ''
    },
    playerCard: {
      role: '',
      title: '',
      name: '',
      xp: '',
      progress: 0,
      next: '',
      deviceIcon: '',
      deviceBadge: '',
      stats: []
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
    roleBenefitConfig: {
      permissionPrompts: ROLE_PERMISSION_PROMPTS
    },
    entries: EMPTY_ROLE_HOME.entries,
    onlineCard: EMPTY_ROLE_HOME.onlineCard,
    mainSection: EMPTY_ROLE_HOME.mainSection,
    mainTabs: [],
    mainFilter: 'all',
    mainGames: [],
    skills: [],
    review: null,
    recommendation: null,
    network: null,
    rankingSection: {
      icon: '',
      title: '',
      moreText: '',
      route: ROUTES.profileFootprintAchievements
    },
    rankingActiveRole: 'expert',
    rankingTabs: [],
    rankingBoards: {},
    rankingList: [],
    myRank: {
      rank: '',
      avatarFallback: '',
      name: '',
      desc: '',
      xp: '',
      xpUnit: ''
    },
    showMyRank: false,
    achievementSection: {
      icon: '',
      title: ''
    },
    achievements: [],
    metaverse: {
      title: '',
      desc: '',
      tags: [],
      avatars: [],
      badge: ''
    }
  },

  lifetimes: {
    attached() {
      const roleType = this.normalizeRoleType(this.properties.roleType)

      this.applyRoleFallback(roleType)
      this.loadRoleBenefitConfig()
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
        const dashboard = this.formatRoleDashboard(
          home.roleDashboard || this.homePayloadToRoleDashboard(home, roleType),
          roleType
        )
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

    homePayloadToRoleDashboard(home = {}, roleType = 'expert') {
      const summary = home.playerSummary || {}
      const nearbySummary = home.nearbySummary || {}
      const nearbySection = home.nearbySection || {}
      const roleName = summary.roleName || summary.roleLabel || ROLE_NAMES[roleType] || ''
      const expToNext = summary.expToNextLevel
      const nearbyGameCount = nearbySummary.nearbyGameCount
      const checkedInCount = nearbySummary.checkedInCount
      const mainTabs = Array.isArray(nearbySection.tabs) ? nearbySection.tabs : []
      const mainGames = Array.isArray(home.nearbyGames) && home.nearbyGames.length
        ? home.nearbyGames
        : Array.isArray(home.recommendedGames) ? home.recommendedGames : []

      return {
        roleName,
        roleEmoji: summary.roleEmoji || this.roleEmoji(roleType),
        deviceBadge: summary.deviceBadge || '',
        profileName: summary.displayName || summary.profileName || '',
        identity: summary.identity || summary.profileTitle || '',
        levelText: summary.roleLabel || summary.levelText || '',
        scoreText: this.formatXpText(summary),
        progress: summary.progressPercent || summary.progress || 0,
        nextLevelText: expToNext == null ? '' : `距离下一等级还需 ${expToNext} 经验值`,
        stats: Array.isArray(summary.stats) ? summary.stats : [],
        entries: Array.isArray(home.quickActions) ? home.quickActions : [],
        onlineCard: {
          title: home.onlineCardTitle || '',
          desc: home.onlineCardDesc || '',
          tags: [
            nearbyGameCount == null ? '' : `附近 ${nearbyGameCount} 个组局`,
            checkedInCount == null ? '' : `已打卡 ${checkedInCount} 处`
          ].filter(Boolean)
        },
        mainSection: {
          title: nearbySection.title || '',
          moreText: nearbySection.moreText || '',
          desc: nearbySection.desc || ''
        },
        mainTabs,
        mainGames,
        skills: Array.isArray(home.skills) ? home.skills : [],
        review: home.review || null,
        recommendation: home.recommendation || null,
        network: home.network || null
      }
    },

    roleEmoji(roleType) {
      if (roleType === 'player') {
        return '🎮'
      }
      if (roleType === 'guide') {
        return '🌐'
      }
      return '🎯'
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

      return EMPTY_ROLE_HOME
    },

    homeRoute() {
      const roleType = this.normalizeRoleType(this.data.currentRoleType || this.properties.roleType)

      if (roleType === 'guide') {
        return ROUTES.guideHome || ROUTES.playerHome || ROUTES.home
      }

      if (roleType === 'expert') {
        return ROUTES.expertHome || ROUTES.playerHome || ROUTES.home
      }

      return ROUTES.playerHome || ROUTES.home
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
          { id: 'connected', icon: '●', name: '已连接', desc: summary || '' },
          { id: 'income', icon: '¥', name: '本周收益', desc: network.income || '' },
          { id: 'location', icon: '📍', name: '核心区域', desc: network.location || '' }
        ]
      const buttons = Array.isArray(network.buttons) && network.buttons.length
        ? network.buttons
        : [
          { text: network.actionText || network.moreText || '查看全部', primary: true, route: network.route },
          { text: network.secondaryText || '管理连接', route: network.manageRoute }
        ]

      return {
        ...network,
        title: network.title || '',
        status: network.status || '',
        hubTitle: network.hubTitle || network.centerTitle || '',
        hubDesc: network.hubDesc || network.centerDesc || '',
        summary,
        income: network.income || '',
        location: network.location || '',
        locationText: network.locationText || String(network.location || '').replace(/^📍\s*/, ''),
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

    formatReview(dashboard = {}, source = EMPTY_ROLE_HOME) {
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
          ''
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

    formatMainSection(dashboard = {}, source = EMPTY_ROLE_HOME) {
      const section = dashboard.mainSection || {}
      const fallback = source.mainSection || EMPTY_ROLE_HOME.mainSection
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
        route: section.route || dashboard.sectionRoute || fallback.route || ROUTES.gameHall,
        count,
        desc: section.desc || dashboard.sectionDesc || fallback.desc,
        moreEnabled
      }
    },

    formatHero(hero = {}, dashboard = {}, roleType = 'expert') {
      const roleName = dashboard.roleName || hero.roleName || ROLE_NAMES[roleType] || ''
      const date = [hero.dateLabel, hero.subtitle]
        .filter((item) => typeof item === 'string' && item.trim())
        .join(' | ')

      return {
        title: hero.title || (roleName ? `HELLO, ${roleName}!` : ''),
        date
      }
    },

    formatPlayerCard(dashboard) {
      return {
        role: dashboard.levelText || dashboard.roleName,
        title: dashboard.identity || dashboard.profileTitle || '',
        name: dashboard.profileName || '',
        xp: dashboard.xpText || this.formatXpText(dashboard) || dashboard.scoreText || '',
        progress: this.normalizeProgress(
          this.pickFirstValue(dashboard.progressPercent, dashboard.progress),
          0
        ),
        next: this.formatRoleNextLevelText(dashboard),
        deviceIcon: dashboard.roleEmoji || '',
        deviceBadge: dashboard.deviceBadge || '',
        stats: Array.isArray(dashboard.stats) && dashboard.stats.length ? dashboard.stats.slice(0, 3) : []
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

      return dashboard.nextLevelText || ''
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
      const source = Array.isArray(entries) && entries.length ? entries : []
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
          ownedGameCount ? `我的组局 ${ownedGameCount} 个` : '',
          checkinCount ? `已打卡 ${checkinCount} 处` : ''
        ]
          .filter(Boolean)

      return {
        title: onlineCard.title || '',
        desc: onlineCard.desc || '',
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
        id: this.resolveGameCardId(item),
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
        actions: this.formatGameActions(item.actions),
        route: this.resolveGameCardRoute(item)
      }))
    },

    resolveGameCardId(item = {}) {
      return this.pickFirstValue(item.gameId, item.id, item.gameID, item.game_id)
    },

    resolveGameCardRoute(item = {}) {
      const configuredRoute = item.detailRoute || item.detailUrl || item.route || item.url || item.path
      const gameId = this.resolveGameCardId(item)

      if (configuredRoute) {
        const route = String(configuredRoute).replace(/^\/+/, '')

        return route === ROUTES.gameDetail && gameId != null && gameId !== ''
          ? `${ROUTES.gameDetail}?id=${encodeURIComponent(gameId)}`
          : route
      }

      return gameId != null && gameId !== ''
        ? `${ROUTES.gameDetail}?id=${encodeURIComponent(gameId)}`
        : ROUTES.gameHall
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
          moreText: section.moreText || section.moreLabel || section.actionText || this.data.rankingSection.moreText,
          route: section.route || section.moreRoute || section.actionRoute || this.data.rankingSection.route || ROUTES.profileFootprintAchievements
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

      const config = this.data.roleBenefitConfig || {}
      const prompts = config.permissionPrompts || ROLE_PERMISSION_PROMPTS
      return this.formatRolePermissionPromptWithPrompts(roleType, currentRoleType, prompts)
    },

    formatRolePermissionPromptWithPrompts(roleType, currentRoleType = this.data.currentRoleType, prompts = ROLE_PERMISSION_PROMPTS) {
      if (!roleType || roleType === currentRoleType || roleType === 'player') {
        return {
          visible: false,
          roleType: '',
          title: '',
          primary: '',
          secondary: ''
        }
      }

      const prompt = prompts[roleType] || ROLE_PERMISSION_PROMPTS[roleType]

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

    async loadRoleBenefitConfig() {
      try {
        const config = await roleService.getRoleBenefitConfig()
        const prompts = (config && config.permissionPrompts) || ROLE_PERMISSION_PROMPTS
        const selectedRole = this.data.selectedRoleTag || this.data.currentRoleType

        this.setData({
          roleBenefitConfig: {
            ...config,
            permissionPrompts: prompts
          },
          rolePermissionPrompt: this.formatRolePermissionPromptWithPrompts(selectedRole, this.data.currentRoleType, prompts)
        })
      } catch (error) {
        this.setData({
          roleBenefitConfig: {
            permissionPrompts: ROLE_PERMISSION_PROMPTS
          }
        })
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

      if (key === 'left' || key === 'right') {
        navigateShellKey(key, {
          currentRoute: this.homeRoute()
        })
        return
      }

      if (navigateShellKey(key, {
        currentRoute: this.homeRoute(),
        onSameRoute: () => this.scrollRoleHomeToTop()
      })) {
        return
      }

      wx.showToast({
        title: '请选择可用入口',
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

      navigateShellRoute(url, {
        currentRoute: this.homeRoute()
      })
    },

    handleRoleCompareTap() {
      const returnRoute = this.data.currentRoleType === 'guide' ? ROUTES.guideHome : ROUTES.expertHome

      navigateShellRoute(`${ROUTES.home}?ui=1&mode=roleComparison&single=1&returnTo=${encodeURIComponent(returnRoute)}`, {
        currentRoute: this.homeRoute()
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

      if (route) {
        navigateShellRoute(route, {
          currentRoute: this.homeRoute()
        })
        return
      }

      wx.showToast({
        title: '暂无更多评价',
        icon: 'none'
      })
    },

    handleActionTap(event) {
      const route = event && event.currentTarget && event.currentTarget.dataset && event.currentTarget.dataset.route

      if (route) {
        navigateShellRoute(route, {
          currentRoute: this.homeRoute()
        })
        return
      }

      wx.showToast({
        title: '请选择可用入口',
        icon: 'none'
      })
    }
  }
})

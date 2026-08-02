const homeService = require('../../services/home')
const roleService = require('../../services/role')
const gameService = require('../../services/game')
const locationService = require('../../services/location')
const locationAccess = require('../../utils/location-access')
const { ROUTES } = require('../../config/routes')
const { navigateShellKey, navigateShellRoute } = require('../../utils/shell-nav')
const { setActiveRole } = require('../../utils/active-role')
const { isRoleApplyResultViewed, markRoleApplyResultViewed } = require('../../utils/role-apply-result-view')
const { AUTH_EXPIRED_MESSAGE, goLogin, isAuthExpiredError } = require('../../utils/auth-error')

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
  dateLabel: '',
  subtitle: '',
  roleEmoji: '🎮',
  deviceBadge: '⚡',
  profileName: '',
  identity: '',
  levelText: '',
  scoreText: '',
  progress: 0,
  nextLevelText: '',
  sectionTitle: '',
  sectionMore: '',
  stats: [],
  entries: [
    { title: '发起组局', desc: '创建你的带局房间', icon: '📍', theme: 'pink', route: ROUTES.gameCreate },
    { title: '局前大厅', desc: '准备就绪加入一局', icon: '', routeIcon: true, theme: 'cyan', route: ROUTES.gameHall }
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
  friendSection: {
    icon: '🎲',
    title: '朋友在玩',
    count: 0,
    moreText: '查看全部',
    route: ROUTES.gameHall
  },
  friendGames: [],
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

const DEFAULT_ROLE_STATUS_STATE = {
  player: 'approved',
  expert: 'none',
  guide: 'none'
}

const ROLE_PENDING_STATUSES = ['pending', 'reviewing', 'auditing', 'pending_audit']

const ROLE_AUDIT_DEFAULT_ICONS = {
  pending: '⏳',
  approved: '✅',
  rejected: '😔'
}

function extractResponseData(result) {
  if (
    result
    && typeof result === 'object'
    && Object.prototype.hasOwnProperty.call(result, 'data')
    && (
      Object.prototype.hasOwnProperty.call(result, 'code')
      || Object.prototype.hasOwnProperty.call(result, 'message')
      || Object.prototype.hasOwnProperty.call(result, 'requestId')
    )
  ) {
    return result.data || {}
  }

  return result || {}
}

function hasOwnField(source, key) {
  return Boolean(source && typeof source === 'object' && Object.prototype.hasOwnProperty.call(source, key))
}

function normalizeGameCardAction(action = '', label = '') {
  const key = String(action || '').trim().toLowerCase()
  const text = String(label || action || '').trim()

  if (key === 'primary') return 'primary'
  if (key === 'share' || text === '分享') return 'share'
  if (key === 'follow' || key === 'favorite' || text === '关注' || text === '收藏') return 'follow'
  if (key === 'refer' || key === 'referral' || text === '引荐') return 'refer'
  if (key === 'greet' || key === 'say_hi' || text === '打招呼') return 'greet'

  return text
}

function isJoinActionText(text = '') {
  return /加入|报名|申请/.test(String(text || ''))
}

function isDisabledActionText(text = '') {
  return /已满|结束|完成|关闭/.test(String(text || ''))
}

const EMPTY_ROLE_HOME = {
  roleName: '行家',
  dateLabel: '',
  subtitle: '',
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
  friendSection: {
    icon: '🎲',
    title: '朋友在玩',
    count: 0,
    moreText: '查看全部',
    route: ROUTES.gameHall
  },
  friendGames: [],
  skills: [],
  review: null,
  recommendation: null,
  network: null
}

const GUIDE_ACTION_ENTRIES = [
  { title: '局前大厅', desc: '准备加入一局', icon: '🎮', routeIcon: true, theme: 'pink', route: ROUTES.gameHall },
  { title: '我的邀约', desc: '管理连接的玩家', icon: '📍', routeIcon: true, theme: 'cyan', route: 'pages/profile/service-center/invite/overview/index' }
]

const GUIDE_ROLE_HOME = {
  ...EMPTY_ROLE_HOME,
  roleName: '领路人',
  roleEmoji: '🌐',
  deviceBadge: '⚡',
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
  onlineCard: {
    title: '地球online',
    desc: '查看附近组局与城市关系',
    tags: []
  },
  recommendation: null,
  network: null
}

function normalizeBackendOnlineText(value) {
  if (value == null) {
    return ''
  }

  return String(value).trim()
}

function resolveBackendOnlineText(home = {}) {
  const hero = resolveHomeHero(home)

  return normalizeBackendOnlineText(hero.onlineText)
}

function resolveHomeHero(home = {}) {
  const hero = home && home.hero ? home.hero : {}
  const nestedHero = home && home.data && home.data.hero ? home.data.hero : {}

  return {
    ...nestedHero,
    ...hero,
    onlineText: normalizeBackendOnlineText(hero.onlineText) || normalizeBackendOnlineText(nestedHero.onlineText),
    dateLabel: normalizeBackendOnlineText(hero.dateLabel) || normalizeBackendOnlineText(nestedHero.dateLabel),
    subtitle: normalizeBackendOnlineText(hero.subtitle) || normalizeBackendOnlineText(nestedHero.subtitle),
    roleName: normalizeBackendOnlineText(hero.roleName) || normalizeBackendOnlineText(nestedHero.roleName),
    title: normalizeBackendOnlineText(hero.title) || normalizeBackendOnlineText(nestedHero.title)
  }
}

function createEmptyRoleAuditPrompt() {
  return {
    visible: false,
    roleType: '',
    status: '',
    title: '',
    icon: '',
    submittedText: '',
    detailText: '',
    helperText: '',
    done: false,
    resultStatus: '',
    steps: []
  }
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
    loadError: '',
    loadErrorActionText: '重试',
    loadErrorAuthExpired: false,
    onlineText: '',
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
    blueBadge: {
      enabled: false,
      label: '',
      ruleText: '',
      contactText: ''
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
    roleStatusConfig: null,
    roleStatusState: DEFAULT_ROLE_STATUS_STATE,
    roleApplicationList: [],
    roleAuditPrompt: createEmptyRoleAuditPrompt(),
    entries: EMPTY_ROLE_HOME.entries,
    onlineCard: EMPTY_ROLE_HOME.onlineCard,
    mainSection: EMPTY_ROLE_HOME.mainSection,
    mainTabs: [],
    mainFilter: 'all',
    mainGames: [],
    mainVisibleGames: [],
    mainEmptyText: '暂无组局',
    locationGuideVisible: false,
    friendSection: EMPTY_ROLE_HOME.friendSection,
    friendGames: [],
    friendEmptyText: '暂无朋友组局',
    skills: [],
    skillTreeSlots: [],
    skillTreeExtras: [],
    allSkills: [],
    skillsExpanded: false,
    skillMoreCount: 0,
    skillPreviewCount: 3,
    review: null,
    recommendation: null,
    network: null,
    rankingSection: {
      icon: '',
      title: '',
      desc: '',
      moreText: '',
      route: ROUTES.homeRanking
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

      this.roleHomeDisposed = false
      this.roleHomeRequestSeq = 0
      this.skipNextShowReload = true
      this.applyRoleFallback(roleType)
      this.loadRoleBenefitConfig()
      this.loadRoleHome(roleType)
    },

    detached() {
      this.roleHomeDisposed = true
      this.clearHomeScrollTimers()
    }
  },

  pageLifetimes: {
    show() {
      if (this.skipNextShowReload) {
        this.skipNextShowReload = false
        return
      }

      this.refreshRoleHomeStatus()
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
    refreshRoleHomeStatus() {
      if (this.roleHomeDisposed) {
        return
      }

      const roleType = this.normalizeRoleType(this.data.currentRoleType || this.properties.roleType)

      this.loadRoleHome(roleType)
    },

    applyRoleFallback(roleTypeValue) {
      const roleType = this.normalizeRoleType(roleTypeValue)
      const fallback = this.formatRoleDashboard({}, roleType)
      const rankingState = this.formatRankingState({}, roleType)
      const skillState = this.buildSkillDisplayState(fallback.skills, fallback.skillPreviewCount, false)
      const roleStatusState = Object.assign({}, DEFAULT_ROLE_STATUS_STATE)
      const mainFilter = fallback.mainTabs.length ? fallback.mainTabs[0].key : 'all'
      const mainVisibleState = this.formatMainVisibleState(fallback.mainGames, mainFilter)

      this.setData({
        currentRoleType: roleType,
        homeReady: true,
        loadError: '',
        loadErrorActionText: '重试',
        loadErrorAuthExpired: false,
        onlineText: '',
        selectedRoleTag: roleType,
        roleStatusState,
        roleApplicationList: [],
        roleAuditPrompt: createEmptyRoleAuditPrompt(),
        hero: this.formatHero({}, fallback, roleType),
        playerCard: this.formatPlayerCard(fallback),
        blueBadge: { enabled: false, label: '', ruleText: '', contactText: '' },
        roleTags: this.formatRoleTags(roleType, roleStatusState),
        rolePermissionPrompt: this.formatRolePermissionPrompt('', roleStatusState, roleType),
        entries: fallback.entries,
        onlineCard: fallback.onlineCard,
        mainSection: fallback.mainSection,
        mainTabs: fallback.mainTabs,
        mainFilter,
        mainGames: fallback.mainGames,
        mainVisibleGames: mainVisibleState.games,
        mainEmptyText: mainVisibleState.emptyText,
        friendSection: fallback.friendSection,
        friendGames: fallback.friendGames,
        friendEmptyText: '暂无朋友组局',
        ...skillState,
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
      const requestSeq = (this.roleHomeRequestSeq || 0) + 1

      this.roleHomeRequestSeq = requestSeq

      try {
        const [homePayload, roleInfo, roleStatusConfig] = await Promise.all([
          homeService.getHome({ roleType }),
          this.loadRoleInfoSnapshot(),
          this.loadRoleStatusConfigSnapshot()
        ])

        if (this.roleHomeDisposed || requestSeq !== this.roleHomeRequestSeq) {
          return
        }

        const home = homePayload || {}
        this.assertRoleHomePayload(home, roleType)
        const hero = resolveHomeHero(home)
        const dashboardSource = home.roleDashboard
          ? Object.assign({}, home.roleDashboard, {
            friendSection: home.roleDashboard.friendSection || home.friendSection,
            friendGames: home.roleDashboard.friendGames || home.friendGames
          })
          : this.homePayloadToRoleDashboard(home, roleType)
        const dashboard = this.formatRoleDashboard(
          dashboardSource,
          roleType
        )
        const rankingState = this.formatRankingState(home || {}, roleType)
        const achievementState = this.formatAchievementState(home || {})
        const roleStatusState = this.formatRoleStatusState(home || {}, roleInfo || {})
        const roleApplicationList = this.formatRoleApplicationList(home || {}, roleInfo || {})
        const selectedRole = roleType === 'player'
          ? this.resolveSelectedRole(roleType, roleStatusState, roleApplicationList)
          : roleType
        setActiveRole(roleType)
        const mainFilter = dashboard.mainTabs.length ? dashboard.mainTabs[0].key : 'all'
        const mainVisibleState = this.formatMainVisibleState(dashboard.mainGames, mainFilter)

        this.setData({
          currentRoleType: roleType,
          loadError: '',
          loadErrorActionText: '重试',
          loadErrorAuthExpired: false,
          selectedRoleTag: selectedRole,
          roleStatusConfig,
          roleStatusState,
          roleApplicationList,
          onlineText: resolveBackendOnlineText(home),
          hero: this.formatHero(hero, dashboard, roleType),
          playerCard: this.formatPlayerCard(dashboard),
          blueBadge: home.expertBlueBadge || ((home.user || {}).expertBlueBadge) || { enabled: false, label: '', ruleText: '', contactText: '' },
          roleTags: this.formatRoleTags(selectedRole, roleStatusState, roleApplicationList),
          rolePermissionPrompt: this.formatRolePermissionPrompt(selectedRole, roleStatusState, roleType, roleApplicationList),
          roleAuditPrompt: this.formatRoleAuditPrompt(selectedRole, roleStatusState, roleApplicationList, roleStatusConfig),
          entries: dashboard.entries,
          onlineCard: dashboard.onlineCard,
          mainSection: dashboard.mainSection,
          mainTabs: dashboard.mainTabs,
          mainFilter,
          mainGames: dashboard.mainGames,
          mainVisibleGames: mainVisibleState.games,
          mainEmptyText: mainVisibleState.emptyText,
          friendSection: dashboard.friendSection,
          friendGames: dashboard.friendGames,
          friendEmptyText: '暂无朋友组局',
          ...this.buildSkillDisplayState(dashboard.skills, dashboard.skillPreviewCount, false),
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
        if (this.roleHomeDisposed || requestSeq !== this.roleHomeRequestSeq) {
          return
        }

        const roleStatusState = Object.assign({}, DEFAULT_ROLE_STATUS_STATE)
        const authExpired = isAuthExpiredError(error)

        this.setData({
          currentRoleType: roleType,
          loadError: authExpired ? AUTH_EXPIRED_MESSAGE : '网络异常，请重试',
          loadErrorActionText: authExpired ? '去登录' : '重试',
          loadErrorAuthExpired: authExpired,
          selectedRoleTag: roleType,
          roleStatusConfig: null,
          roleStatusState,
          roleApplicationList: [],
          roleAuditPrompt: createEmptyRoleAuditPrompt(),
          onlineText: '',
          roleTags: this.formatRoleTags(roleType, roleStatusState),
          rolePermissionPrompt: this.formatRolePermissionPrompt('', roleStatusState, roleType)
        })
      }
    },

    async loadRoleInfoSnapshot() {
      return extractResponseData(await roleService.getMyRoles())
    },

    async loadRoleStatusConfigSnapshot() {
      const config = extractResponseData(await roleService.getRoleStatusPageConfig())

      if (!this.isValidRoleStatusConfig(config)) {
        throw new Error('角色状态配置格式错误')
      }

      return config
    },

    assertRoleHomePayload(home = {}, roleType = 'player') {
      const summary = home.playerSummary || {}
      const commonChecks = [
        ['hero', this.isPlainObject(home.hero)],
        ['playerSummary', this.isPlainObject(home.playerSummary)],
        ['playerSummary.displayName', this.hasText(summary.displayName || summary.profileName)],
        ['playerSummary.roleLabel', this.hasText(summary.roleLabel || summary.levelText)],
        ['playerSummary.scoreText', this.hasText(summary.scoreText || summary.xpText)],
        ['playerSummary.stats', Array.isArray(summary.stats)],
        ['onlineCard', this.isPlainObject(home.onlineCard) && this.hasText(home.onlineCard.title) && this.hasText(home.onlineCard.desc) && Array.isArray(home.onlineCard.tags)],
        ['nearbySummary', this.isPlainObject(home.nearbySummary)],
        ['nearbySection', this.isPlainObject(home.nearbySection) && this.hasText(home.nearbySection.title) && Array.isArray(home.nearbySection.tabs)],
        ['rankingSection', this.isPlainObject(home.rankingSection) && this.hasText(home.rankingSection.title) && Array.isArray(home.rankingSection.tabs)],
        ['rankingBoards', this.isPlainObject(home.rankingBoards)],
        ['achievementSection', this.isPlainObject(home.achievementSection) && this.hasText(home.achievementSection.title)],
        ['achievements', Array.isArray(home.achievements)]
      ]
      const roleChecks = []

      if (roleType === 'expert') {
        roleChecks.push(['skills', Array.isArray(home.skills)])
      }

      if (roleType === 'guide') {
        roleChecks.push(['network', this.isPlainObject(home.network)])
      }

      const missing = commonChecks.concat(roleChecks)
        .filter(([, ok]) => !ok)
        .map(([key]) => key)

      if (missing.length > 0) {
        throw new Error(`invalid role home payload: ${missing.join(', ')}`)
      }
    },

    isPlainObject(value) {
      return Boolean(value && typeof value === 'object' && !Array.isArray(value))
    },

    hasText(value) {
      return typeof value === 'string' && value.trim() !== ''
    },

    homePayloadToRoleDashboard(home = {}, roleType = 'expert') {
      const defaultHome = this.getDefaultRoleHome(roleType)
      const hero = resolveHomeHero(home)
      const summary = home.playerSummary || {}
      const nearbySummary = home.nearbySummary || {}
      const nearbySection = home.nearbySection || {}
      const roleName = summary.roleName || hero.roleName || defaultHome.roleName || ROLE_NAMES[roleType] || ''
      const nearbyGameCount = nearbySummary.nearbyGameCount
      const checkedInCount = nearbySummary.checkedInCount
      const mainTabs = Array.isArray(nearbySection.tabs) && nearbySection.tabs.length
        ? nearbySection.tabs
        : defaultHome.mainTabs
      const mainGames = this.mergeMainGameGroups([
        { scope: 'nearby', items: home.nearbyGames },
        { scope: 'city', items: home.recommendedGames }
      ])
      const onlineTags = [
        nearbyGameCount == null ? '' : `附近 ${nearbyGameCount} 个组局`,
        checkedInCount == null ? '' : `已打卡 ${checkedInCount} 处`
      ].filter(Boolean)
      const experience = Number(summary.experience)
      const nextLevelExperience = Number(summary.nextLevelExperience)
      const expToNextLevel = Number(summary.expToNextLevel)
      const hasExperienceProgress = Number.isFinite(experience)
        && experience >= 0
        && Number.isFinite(nextLevelExperience)
        && nextLevelExperience > 0
      // 玩家保留“当前经验 / 下一等级经验”的进度表达；行家、领路人
      // 使用服务端返回的综合得分，不能再误标为 XP。
      const experienceText = roleType === 'player' && hasExperienceProgress
        ? `${experience}/${nextLevelExperience} 经验`
        : (summary.xpText || summary.scoreText || '')
      const nextLevelText = summary.nextLevelText || (
        Number.isFinite(expToNextLevel) && expToNextLevel >= 0
          ? `距离下一级还需 ${expToNextLevel} 经验值`
          : ''
      )
      const legacyLevelText = String(summary.roleLabel || summary.levelText || '').trim()
        .replace(new RegExp(`^${roleName}\\s*[·•|]\\s*`), '')
      const levelNumber = Number(summary.level)
      const fallbackBadge = Number.isFinite(levelNumber)
        ? (roleType === 'expert' ? `${levelNumber}星` : (roleType === 'guide' ? `${levelNumber}级` : `Lv${levelNumber}`))
        : legacyLevelText
      const fallbackTitle = legacyLevelText
        .replace(new RegExp(`^${fallbackBadge}\\s*`), '')
        .trim()

      return {
        roleName,
        roleType,
        dateLabel: hero.dateLabel || defaultHome.dateLabel || '',
        subtitle: hero.subtitle || defaultHome.subtitle || '',
        roleEmoji: summary.roleEmoji || this.roleEmoji(roleType),
        deviceBadge: summary.deviceBadge || defaultHome.deviceBadge || (roleType === 'expert' ? '💎' : ''),
        profileName: summary.displayName || summary.profileName || '',
        identity: summary.levelTitle || fallbackTitle || summary.identity || summary.profileTitle || '',
        levelText: summary.levelBadge || fallbackBadge,
        xpText: experienceText,
        scoreText: summary.scoreText || '',
        progress: this.pickFirstValue(summary.progressPercent, summary.progress, 0),
        nextLevelText,
        stats: Array.isArray(summary.stats) ? summary.stats : [],
        entries: Array.isArray(home.quickActions) && home.quickActions.length ? home.quickActions : defaultHome.entries,
        onlineCard: home.onlineCard || {
          title: home.onlineCardTitle || defaultHome.onlineCard.title || '',
          desc: home.onlineCardDesc || defaultHome.onlineCard.desc || '',
          tags: onlineTags.length ? onlineTags : defaultHome.onlineCard.tags
        },
        mainSection: {
          title: nearbySection.title || defaultHome.mainSection.title || '',
          moreText: nearbySection.moreText || defaultHome.mainSection.moreText || '',
          desc: nearbySection.desc || defaultHome.mainSection.desc || ''
        },
        mainTabs,
        mainGames,
        friendSection: this.formatFriendSection(home.friendSection),
        friendGames: this.formatGameCards(home.friendGames),
        skills: Array.isArray(home.skills) ? home.skills : [],
        skillPreviewCount: home.skillDisplay && home.skillDisplay.homePreviewCount,
        review: home.review || null,
        recommendation: home.recommendation || null,
        network: home.network || home.roleNetwork || (roleType === 'guide' ? this.formatGuideNetwork(home) : null)
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
        roleType,
        deviceBadge: dashboard.deviceBadge || source.deviceBadge || (roleType === 'expert' ? '💎' : ''),
        dateLabel: dashboard.dateLabel || defaultHome.dateLabel || '',
        subtitle: dashboard.subtitle || defaultHome.subtitle || '',
        stats: Array.isArray(dashboard.stats) && dashboard.stats.length ? dashboard.stats : source.stats,
        entries: Array.isArray(dashboard.entries) && dashboard.entries.length
          ? dashboard.entries
          : this.formatEntries(dashboard.quickActions || source.entries, roleType),
        onlineCard: this.formatRoleOnlineCard(dashboard.onlineCard || source.onlineCard),
        mainSection: this.formatMainSection(dashboard, source),
        mainTabs: this.formatMainTabs(dashboard.mainTabs || dashboard.filterTabs || source.mainTabs),
        mainGames: this.formatGameCards(dashboard.mainGames || dashboard.events || source.mainGames),
        friendSection: this.formatFriendSection(dashboard.friendSection || source.friendSection),
        friendGames: this.formatGameCards(dashboard.friendGames || source.friendGames),
        skills: this.formatSkills(dashboard.skills, source.skills),
        skillPreviewCount: Number(dashboard.skillPreviewCount || ((dashboard.skillDisplay || {}).homePreviewCount) || 3),
        review: this.formatReview(dashboard, source),
        recommendation: dashboard.recommendation || source.recommendation,
        network: this.formatNetwork(dashboard.network || source.network)
      }
    },

    formatSkills(skills) {
      return (Array.isArray(skills) ? skills : []).map((item, index) => {
        const sourceItem = item || {}
        const hasSourceValue = Boolean(sourceItem.title || sourceItem.name || sourceItem.label || sourceItem.icon || sourceItem.iconText || sourceItem.emoji)
        const locked = Boolean(
          sourceItem.locked ||
          sourceItem.unlocked === false ||
          sourceItem.tone === 'locked' ||
          !hasSourceValue
        )

        return {
          ...sourceItem,
          icon: sourceItem.icon || sourceItem.iconText || sourceItem.emoji || '',
          title: sourceItem.title || sourceItem.name || sourceItem.label || '',
          tone: locked ? 'locked' : (sourceItem.tone || (index === 0 ? 'green' : 'orange')),
          nodeStyle: locked ? '' : String(sourceItem.nodeStyle || ''),
          locked,
          unlocked: !locked
        }
      })
    },

    buildSkillDisplayState(skills, previewCount, expanded) {
      const allSkills = Array.isArray(skills) ? skills : []
      const safePreviewCount = Math.max(1, Number(previewCount) || 3)
      const skillsExpanded = Boolean(expanded) && allSkills.length > safePreviewCount
      const visibleSkills = skillsExpanded ? allSkills : allSkills.slice(0, safePreviewCount)
      const positions = ['start', 'middle', 'top', 'bottom']
      const skillTreeSlots = positions.map((treePosition, index) => {
        const skill = visibleSkills[index]
        if (skill) {
          return Object.assign({}, skill, { treePosition })
        }
        return {
          key: `locked-${treePosition}`,
          icon: '🔒',
          title: '待解锁',
          tone: 'locked',
          locked: true,
          treePosition
        }
      })
      return {
        allSkills,
        skills: visibleSkills,
        skillTreeSlots,
        skillTreeExtras: visibleSkills.slice(4),
        skillsExpanded,
        skillMoreCount: Math.max(0, allSkills.length - safePreviewCount),
        skillPreviewCount: safePreviewCount
      }
    },

    handleSkillMoreTap() {
      this.setData(this.buildSkillDisplayState(this.data.allSkills, this.data.skillPreviewCount, !this.data.skillsExpanded))
    },

    handleBlueBadgeTap() {
      const badge = this.data.blueBadge || {}
      wx.showModal({
        title: badge.label || '蓝标认证',
        content: [badge.ruleText, badge.contactText].filter(Boolean).join('\n'),
        showCancel: false,
        confirmText: '我知道了'
      })
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
          { id: 'industry-pending', icon: '🌐', name: '领域待完善', desc: '请完善领路人行业' },
          { id: 'all', icon: '+', name: '查看全部', desc: `${Number(network.connectedCount || network.playerCount) || 0}人`, dashed: true }
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
        location: network.location || '',
        locationText: network.locationText || String(network.location || '').replace(/^📍\s*/, ''),
        items: items.slice(0, 5),
        buttons: buttons.slice(0, 2)
      }
    },

    formatGuideNetwork(home = {}) {
      const visualization = home.visualization || {}
      const network = visualization.network || home.networkGraph || null

      if (!network) {
        return null
      }

      const stats = Array.isArray(network.stats) ? network.stats : []
      const statMap = stats.reduce((result, item) => {
        if (item && item.key) {
          result[item.key] = item
        }
        return result
      }, {})
      const relationCount = this.pickFirstValue((statMap.relations || {}).value, home.nearbySummary && home.nearbySummary.relationCount, 0)
      const strongCount = this.pickFirstValue((statMap.strongRelations || {}).value, 0)
      const nodeCount = this.pickFirstValue((statMap.onlineNodes || {}).value, network.nodes && network.nodes.length, 0)

      return {
        title: '我的关系网络',
        status: '实时连接中',
        summary: `已连接 ${relationCount} 位玩家`,
        locationText: '核心区',
        items: [
          { id: 'relations', icon: '👑', name: '累计连接', desc: `${relationCount}人` },
          { id: 'strong', icon: '🎓', name: '强关系', desc: `${strongCount}人` },
          { id: 'nodes', icon: '👶', name: '动态节点', desc: `${nodeCount}个` },
          { id: 'more', icon: '+', name: '更多', desc: '待加入', dashed: true }
        ],
        buttons: [
          { text: '管理我的连接', primary: true, route: ROUTES.relationNetwork }
        ]
      }
    },

    formatCentAmount(value) {
      const amount = Number(value || 0) / 100

      if (!Number.isFinite(amount)) {
        return '0'
      }

      return amount % 1 === 0 ? `${amount}` : amount.toFixed(2)
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
        ''
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
      const roleName = hero.roleName || dashboard.roleName || ROLE_NAMES[roleType] || ''
      const date = [hero.dateLabel || dashboard.dateLabel, hero.subtitle || dashboard.subtitle]
        .filter((item) => typeof item === 'string' && item.trim())
        .join(' | ')

      return {
        title: hero.title || (roleName ? `HELLO, ${roleName}!` : ''),
        date
      }
    },

    formatPlayerCard(dashboard) {
      const roleName = String(dashboard.roleName || '').trim()
      // 首页摘要遵循原型：绿色胶囊承载“身份 + 等级”，其右侧只放后台配置的称号。
      // 不能再把身份放进称号，否则会出现“玩家”单独显示在胶囊里的错误层级。
      const levelText = String(dashboard.levelText || '').trim()
      const title = String(
        dashboard.identity || dashboard.profileTitle || dashboard.levelTitle || ''
      ).trim()
      const role = [roleName, levelText].filter(Boolean).join(' ')

      return {
        role,
        title,
        name: dashboard.profileName || '',
        xp: dashboard.xpText || dashboard.scoreText || '',
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
      return dashboard.nextLevelText || ''
    },

    formatEntries(entries, roleType) {
      const source = Array.isArray(entries) && entries.length ? entries : []
      const normalizedSource = roleType === 'guide'
        ? this.normalizeGuideEntries(source)
        : source

      return normalizedSource.slice(0, 2).map((item, index) => {
        const iconSrc = item.iconSrc || item.iconUrl || ''

        return {
          title: item.title,
          desc: item.desc,
          icon: item.icon,
          iconSrc,
          routeIcon: Boolean(!iconSrc && (item.routeIcon || (roleType === 'expert' && index === 1))),
          theme: item.theme || item.tone || (index % 2 === 0 ? 'pink' : 'cyan'),
          route: item.route
        }
      })
    },

    normalizeGuideEntries(entries) {
      const source = Array.isArray(entries) ? entries : []
      const hasExpectedEntries = source.some((item) => item && (item.title === '局前大厅' || item.title === '我的邀约'))

      if (hasExpectedEntries) {
        return source.map((item) => {
          if (!item || (item.title !== '局前大厅' && item.title !== '我的邀约')) {
            return item
          }

          const fallback = GUIDE_ACTION_ENTRIES.find((entry) => entry.title === item.title) || {}
          const normalized = {
            ...fallback,
            ...item,
            theme: item.theme || item.tone || fallback.theme
          }

          if (item.title === '局前大厅') {
            return {
              ...normalized,
              iconSrc: '',
              routeIcon: true,
              route: normalized.route || fallback.route
            }
          }

          return {
            ...normalized,
            iconSrc: '',
            routeIcon: true,
            route: normalized.route || fallback.route
          }
        })
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

    formatFriendSection(section = {}) {
      const fallback = EMPTY_ROLE_HOME.friendSection || {}
      const source = section || {}
      const countText = this.formatStatValue(this.pickFirstValue(source.count, fallback.count, 0))
      const numericCount = Number(countText)
      const count = countText !== '' && (!Number.isFinite(numericCount) || numericCount > 0)
        ? countText
        : ''

      return {
        ...fallback,
        ...source,
        icon: source.icon || fallback.icon || '',
        title: source.title || fallback.title || '朋友在玩',
        count,
        moreText: source.moreText || fallback.moreText || '查看全部',
        route: source.route || fallback.route || ROUTES.gameHall
      }
    },

    formatGameCards(games) {
      if (!Array.isArray(games) || !games.length) {
        return []
      }

      return games.map((item, index) => {
        const playerAvatars = this.formatSessionAvatars(item)
        const scope = item.scope || item.distanceScope || (index === 0 ? 'nearby' : 'city')
        const action = hasOwnField(item, 'actionText') ? String(item.actionText || '') : String(item.action || '')

        return {
          id: this.resolveGameCardId(item),
          scope,
          scopes: Array.isArray(item.scopes) && item.scopes.length
            ? item.scopes
            : [scope],
          title: item.title,
          startTime: item.startTime || item.time || '',
          day: item.dayText || item.day || '',
          date: item.scheduleText || item.dateText || item.timeText || item.date || '',
          venue: item.expertVenueText || item.venueText || item.venue || this.formatExpertVenue(item.meta),
          members: item.expertMembersText || item.memberText || item.members || this.formatExpertMembers(item.meta),
          tag: this.formatGameTag(item),
          sessionTags: this.formatSessionTags(item),
          coverSrc: item.coverUrl || item.coverSrc || item.cover || '',
          price: item.playerPriceText || item.priceText || item.price || '',
          action,
          primaryActionText: hasOwnField(item, 'primaryActionText') ? String(item.primaryActionText || '') : action,
          primaryActionType: item.primaryActionType || '',
          primaryActionDisabled: Boolean(item.primaryActionDisabled),
          primaryActionToast: item.primaryActionToast || '',
          primaryActionRoute: item.primaryActionRoute || '',
          status: item.status || '',
          statusText: item.statusText || '',
          relationRole: item.relationRole || '',
          isCreator: Boolean(item.isCreator),
          isMember: Boolean(item.isMember),
          canApply: Boolean(item.canApply),
          canAudit: Boolean(item.canAudit),
          pendingApplications: Number(item.pendingApplications || 0),
          location: item.locationText || this.formatGameLocation(item) || item.meta || item.location || '',
          time: item.playerTimeText || item.timeText || item.dateText || item.time || '',
          joinedText: item.joinedText || item.peopleText || this.formatJoinedText(item.joinedCount) || '',
          playerAvatars,
          avatarUrls: playerAvatars.map((avatar) => avatar.imageUrl).filter(Boolean),
          avatarFallbacks: playerAvatars
            .filter((avatar) => !avatar.imageUrl && avatar.text)
            .map((avatar) => avatar.text),
          actions: this.formatGameActions(item.actions, item.shareComponent),
          shareActionLabel: item.shareComponent && item.shareComponent.label || '分享',
          shareComponentVariant: item.shareComponent && item.shareComponent.variant === 'icon_button' ? 'icon_button' : 'channel_sheet',
          route: this.resolveGameCardRoute(item)
        }
      })
    },

    mergeMainGameGroups(groups = []) {
      const merged = []
      const byKey = {}

      groups.forEach((group) => {
        const items = Array.isArray(group.items) ? group.items : []
        items.forEach((item, index) => {
          if (!item) {
            return
          }
          const scope = item.scope || item.distanceScope || group.scope || ''
          const key = this.resolveGameCardId(item) || `${scope}-${index}-${item.title || ''}`
          const existing = byKey[key]

          if (existing) {
            const scopes = new Set(existing.scopes || [])
            if (scope) {
              scopes.add(scope)
            }
            existing.scopes = Array.from(scopes)
            return
          }

          const next = {
            ...item,
            scope,
            scopes: scope ? [scope] : []
          }
          byKey[key] = next
          merged.push(next)
        })
      })

      return merged
    },

    formatMainVisibleState(games = [], filter = 'all') {
      const list = Array.isArray(games) ? games : []
      const key = filter || 'all'
      const visibleGames = key === 'all'
        ? list
        : list.filter((item) => {
          if (!item) {
            return false
          }
          const scopes = Array.isArray(item.scopes) ? item.scopes : []
          return item.scope === key || scopes.indexOf(key) >= 0
        })

      return {
        games: visibleGames,
        emptyText: this.formatMainEmptyText(key)
      }
    },

    formatMainEmptyText(filter = 'all') {
      const textMap = {
        all: '暂无组局',
        nearby: '附近暂无组局',
        city: '同城暂无组局',
        friend: '朋友暂无组局'
      }

      return textMap[filter] || '暂无组局'
    },

    resolveGameCardId(item = {}) {
      return this.pickFirstValue(item.gameId, item.id, item.gameID, item.game_id)
    },

    currentGameCardRoleType() {
      return this.normalizeRoleType(this.data.currentRoleType || this.properties.roleType)
    },

    appendRoleTypeToGameDetailRoute(route) {
      const normalizedRoute = String(route || '').replace(/^\/+/, '')

      if (!normalizedRoute || normalizedRoute.split('?')[0] !== ROUTES.gameDetail) {
        return normalizedRoute
      }

      const roleType = this.currentGameCardRoleType()
      const routeParts = normalizedRoute.split('?')
      const queryParts = routeParts[1] ? routeParts[1].split('&').filter(Boolean) : []
      const hasRoleParam = queryParts.some((part) => {
        const key = decodeURIComponent(String(part).split('=')[0] || '')
        return key === 'roleType' || key === 'role' || key === 'applicationRole'
      })

      if (!hasRoleParam) {
        queryParts.push(`roleType=${encodeURIComponent(roleType)}`)
      }

      return `${routeParts[0]}?${queryParts.join('&')}`
    },

    resolveGameCardRoute(item = {}) {
      const configuredRoute = item.detailRoute || item.detailUrl || item.route || item.url || item.path
      const gameId = this.resolveGameCardId(item)

      if (configuredRoute) {
        const route = String(configuredRoute).replace(/^\/+/, '')

        return route === ROUTES.gameDetail && gameId != null && gameId !== ''
          ? this.appendRoleTypeToGameDetailRoute(`${ROUTES.gameDetail}?id=${encodeURIComponent(gameId)}`)
          : this.appendRoleTypeToGameDetailRoute(route)
      }

      return gameId != null && gameId !== ''
        ? this.appendRoleTypeToGameDetailRoute(`${ROUTES.gameDetail}?id=${encodeURIComponent(gameId)}`)
        : ROUTES.gameHall
    },

    formatSessionTags(item = {}) {
      const sessionTags = Array.isArray(item.sessionTags) ? item.sessionTags : []
      const tags = sessionTags.length ? sessionTags : (Array.isArray(item.tags) ? item.tags : [])
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
      const source = Array.isArray(item.playerAvatars) ? item.playerAvatars : []

      if (!Array.isArray(source) || source.length === 0) {
        return []
      }

      return source.map((avatar) => {
        if (typeof avatar === 'string') {
          const imageUrl = this.isImagePath(avatar) ? avatar : ''

          return {
            imageUrl
          }
        }

        return {
          imageUrl: avatar.avatarUrl || avatar.imageUrl || avatar.url || '',
          text: avatar.avatarText || avatar.text || '',
          name: avatar.name || avatar.displayName || ''
        }
      }).filter((avatar) => avatar.imageUrl || avatar.text).slice(0, 3)
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

    formatGameActions(actions, shareComponent = {}) {
      const actionMap = {
        share: '分享',
        follow: '关注',
        refer: '引荐',
        greet: '打招呼'
      }

      if (!Array.isArray(actions) || actions.length === 0) {
        return shareComponent.enabled === false ? ['关注', '打招呼'] : [shareComponent.label || '分享', '关注', '打招呼']
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
        return []
      }

      return avatars.slice(0, 3).map((item) => {
        if (typeof item === 'string') {
          return { text: item }
        }

        return {
          text: item.text || item.avatarFallback || item.initial || item.nickname || '',
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
      const boards = this.formatRankingBoards(home)
      const activeKey = requestedActiveKey || roleType
      const activeBoard = boards[activeKey] || {}
      const display = this.formatRankingDisplay(
        activeBoard.list || [],
        activeBoard.myRank || {}
      )

      return {
        section: {
          icon: section.icon || '',
          title: this.formatRankingTitle(section.title || section.name || ''),
          desc: section.desc || section.description || '',
          moreText: section.moreText || section.moreLabel || section.actionText || '',
          route: section.route || section.moreRoute || section.actionRoute || ROUTES.homeRanking
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

    formatRankingTitle(value) {
      const title = `${value || ''}`.trim()

      return title === '玩霸榜' ? '本周玩霸榜' : title
    },

    formatRankingTabs(tabs, activeKey) {
      const sourceTabs = Array.isArray(tabs) && tabs.length > 0 ? tabs : []

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

    formatRankingBoards(home = {}) {
      const sourceBoards = home.rankingBoards || home.rankings || home.rankingByRole
      const boards = {}

      if (sourceBoards && typeof sourceBoards === 'object' && !Array.isArray(sourceBoards)) {
        Object.keys(sourceBoards).forEach((key) => {
          const roleKey = this.normalizeRoleType(key)
          boards[roleKey] = this.formatRankingBoard(sourceBoards[key])
        })
      }

      return boards
    },

    formatRankingBoard(board) {
      const sourceList = Array.isArray(board)
        ? board
        : (board && (board.list || board.rankingList || board.items || board.records))
      const sourceMyRank = board && !Array.isArray(board)
        ? board.myRank || board.currentUserRank || board.me
        : null

      return {
        list: this.formatRankingList(sourceList),
        myRank: this.formatRankingItem(sourceMyRank)
      }
    },

    formatRankingList(list) {
      if (!Array.isArray(list) || list.length === 0) {
        return []
      }

      return list.map((item) => this.formatRankingItem(item))
    },

    formatRankingDisplay(list = [], myRank = {}) {
      const rankingList = Array.isArray(list) ? list : []
      const currentRank = myRank || {}
      const hasCurrentRank = this.hasRankingItem(currentRank)
      let displayList = rankingList.slice(0, 3)

      if (hasCurrentRank && this.shouldInlineMyRank(displayList, currentRank)) {
        displayList = this.mergeMyRankIntoRankingList(displayList, currentRank).slice(0, 3)
      }

      const currentInDisplay = hasCurrentRank && displayList.some((item) => this.isSameRankingUser(item, currentRank))

      return {
        list: displayList,
        myRank: currentRank,
        showMyRank: hasCurrentRank && !currentInDisplay
      }
    },

    hasRankingItem(item = {}) {
      return Boolean(item && (
        item.id ||
        item.userId ||
        item.memberId ||
        item.profileId ||
        item.openId ||
        item.rank ||
        item.name ||
        item.avatarUrl ||
        item.avatarFallback
      ))
    },

    shouldInlineMyRank(list = [], myRank = {}) {
      const rankNumber = this.getRankNumber(myRank.rank)
      const existsInList = Array.isArray(list) && list.some((item) => this.isSameRankingUser(item, myRank))

      return existsInList || (Array.isArray(list) && list.length > 0 && rankNumber > 0 && rankNumber <= 3)
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
      const rawName = item.name || item.nickname || item.displayName || fallback.rawName || fallback.name || ''
      const isMe = Boolean(this.pickFirstValue(item.isMe, item.isSelf, item.isCurrentUser, fallback.isMe, false))
      const name = this.formatRankingDisplayName(rawName, isMe)

      return {
        id: item.id || fallback.id,
        userId: this.pickFirstValue(item.userId, item.userID, item.user_id, fallback.userId),
        memberId: this.pickFirstValue(item.memberId, item.memberID, item.member_id, fallback.memberId),
        profileId: this.pickFirstValue(item.profileId, item.profileID, item.profile_id, fallback.profileId),
        openId: this.pickFirstValue(item.openId, item.openID, item.open_id, fallback.openId),
        isMe,
        rank: this.formatRankNo(this.pickFirstValue(item.rank, item.rankNo, item.position, fallback.rank)),
        avatarUrl,
        avatarFallback: item.avatarFallback || item.initials || (!avatarUrl && avatarCandidate) || fallback.avatarFallback || '',
        rawName,
        name,
        desc: item.desc || item.description || item.summary || item.rankText || fallback.desc || '',
        xp: this.formatRankingXp(this.pickFirstValue(item.xp, item.xpText, item.experience, item.weeklyXp, fallback.xp)),
        xpUnit: item.xpUnit || fallback.xpUnit || ''
      }
    },

    formatRankingDisplayName(name = '', isMe = false) {
      const text = `${name || ''}`.trim()

      if (!isMe || !text || /^我[（(]/.test(text)) {
        return text
      }

      return `我（${text}）`
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
      const progressValue = this.pickFirstValue(item.progressPercent, item.progress)
      const hasProgress = (progressValue != null && progressValue !== '') || /进度|progress/i.test(status)
      const locked = item.locked != null
        ? Boolean(item.locked)
        : Boolean(item.hidden || item.secret || (unlocked === false && !hasProgress))
      const progressPercent = this.normalizeProgress(
        progressValue,
        fallback.progressPercent || 0
      )

      return {
        id: item.id || item.code || fallback.id || title,
        icon: item.icon || item.iconText || fallback.icon || this.resolveAchievementIcon(item, title),
        title,
        status: status ? (status.startsWith('▲') ? status : `▲ ${status}`) : '',
        tone: this.resolveAchievementTone(item, fallback, locked),
        unlocked,
        locked,
        progressPercent,
        showProgress: hasProgress && !locked
      }
    },

    resolveAchievementTone(item, fallback = {}, locked = false) {
      if (locked) {
        return 'locked'
      }

      const text = `${item.code || ''} ${item.title || item.name || ''}`

      if (item.progressPercent != null || item.progress != null || /earth|地球/.test(text)) {
        return 'blue'
      }

      return fallback.tone || 'gold'
    },

    resolveAchievementIcon(item, title = '') {
      const text = `${item.id || ''} ${item.code || ''} ${title || item.name || ''}`

      if (/earth|globe|地球/.test(text)) {
        return '🌍'
      }

      return '🏆'
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

    formatRoleStatusState(home = {}, roleInfo = {}) {
      const user = home.user || {}
      const roleUser = roleInfo.user || {}
      const state = Object.assign({}, DEFAULT_ROLE_STATUS_STATE)

      this.mergeRoleStatusMap(state, user.roleStatusMap || home.roleStatusMap)
      this.mergeRoleStatusMap(state, roleUser.roleStatusMap || roleInfo.roleStatusMap)
      this.mergeRoleStatusList(state, user.roles || home.roles)
      this.mergeRoleStatusList(state, roleUser.roles || roleInfo.roles)
      this.collectRoleApplicationSources(home, roleInfo).forEach((source) => {
        this.mergeRoleStatusList(state, source, true)
      })

      return state
    },

    mergeRoleStatusMap(state, roleStatusMap) {
      if (!roleStatusMap || typeof roleStatusMap !== 'object' || Array.isArray(roleStatusMap)) {
        return
      }

      Object.keys(roleStatusMap).forEach((key) => {
        const roleType = this.normalizeRoleType(key)

        if (roleType) {
          state[roleType] = this.normalizeRoleStatus(roleStatusMap[key])
        }
      })
    },

    mergeRoleStatusList(state, roles, isApplicationList = false) {
      if (!Array.isArray(roles)) {
        return
      }

      const list = isApplicationList
        ? roles.slice().sort((left, right) => this.getRoleApplicationSortTime(right) - this.getRoleApplicationSortTime(left))
        : roles

      list.forEach((item) => {
        if (typeof item === 'string') {
          const roleType = this.normalizeRoleType(item)

          if (roleType && !isApplicationList && state[roleType] === 'none') {
            state[roleType] = 'approved'
          }
          return
        }

        const roleType = this.resolveRoleTypeFromItem(item)
        const status = this.normalizeRoleStatus(
          item.status || item.applyStatus || item.applicationStatus || item.roleStatus || item.role_status
        )

        if (roleType && status !== 'none') {
          state[roleType] = isApplicationList
            ? this.mergeApplicationRoleStatus(state[roleType], status)
            : status
        }
      })
    },

    getRoleApplicationSortTime(item = {}) {
      const value = this.pickFirstValue(
        item.updatedAt,
        item.updated_at,
        item.reviewedAt,
        item.reviewed_at,
        item.createdAt,
        item.created_at,
        item.submittedAt,
        item.submitted_at,
        item.applyTime
      )
      const time = Date.parse(value)

      return Number.isFinite(time) ? time : 0
    },

    mergeApplicationRoleStatus(currentStatus, nextStatus) {
      const current = this.normalizeRoleStatus(currentStatus)
      const next = this.normalizeRoleStatus(nextStatus)

      if (next === 'none') {
        return current
      }

      if (current === 'none') {
        return next
      }

      return current
    },

    formatRoleApplicationList(home = {}, roleInfo = {}) {
      const seen = {}

      return this.collectRoleApplicationSources(home, roleInfo)
        .reduce((items, source) => items.concat(source), [])
        .map((item) => this.formatRoleApplicationItem(item))
        .filter((item) => {
          if (!item.roleType) {
            return false
          }

          const key = item.applicationId
            ? `id:${item.applicationId}`
            : `${item.roleType}:${item.status}:${item.submittedAt}`

          if (seen[key]) {
            return false
          }

          seen[key] = true
          return true
        })
    },

    collectRoleApplicationSources(...payloads) {
      const sources = []

      payloads.forEach((payload) => {
        if (!payload || typeof payload !== 'object') {
          return
        }

        const user = payload.user || {}
        ;[
          payload.roleApplications,
          payload.roleApplicationList,
          payload.applications,
          user.roleApplications,
          user.roleApplicationList,
          user.applications
        ].forEach((source) => {
          if (Array.isArray(source)) {
            sources.push(source)
          }
        })
      })

      return sources
    },

    formatRoleApplicationItem(item = {}) {
      return {
        roleType: this.resolveRoleTypeFromItem(item),
        status: this.normalizeRoleStatus(
          item.status || item.applyStatus || item.applicationStatus || item.roleStatus || item.role_status
        ),
        statusText: item.statusText || item.statusLabel || item.applyStatusText || '',
        submittedAt: item.submittedAt || item.createdAt || item.applyTime || item.created_at || '',
        reviewedAt: item.reviewedAt || item.reviewed_at || '',
        expectedReviewAt:
          item.expectedReviewAt ||
          item.expectedReviewedAt ||
          item.estimatedReviewAt ||
          item.estimatedReviewedAt ||
          item.expectedReviewTime ||
          '',
        applicationId: item.applicationId || item.id || item.application_id || '',
        rejectReason: item.rejectReason || item.reject_reason || ''
      }
    },

    resolveRoleTypeFromItem(item = {}) {
      const roleType = this.pickFirstValue(
        item.roleType,
        item.role_type,
        item.roleCode,
        item.role_code,
        item.role,
        item.type,
        item.key,
        item.name
      )

      return roleType ? this.normalizeRoleType(roleType) : ''
    },

    normalizeRoleStatus(status) {
      if (status == null || status === '') {
        return 'none'
      }

      const value = String(status).trim()
      const statusMap = {
        active: 'approved',
        enabled: 'approved',
        passed: 'approved',
        success: 'approved',
        waiting: 'pending',
        reviewing: 'pending',
        auditing: 'pending',
        pending_audit: 'pending',
        rejected_audit: 'rejected',
        reject: 'rejected',
        disabled: 'disabled',
        none: 'none',
        unavailable: 'none',
        available: 'none',
        审核中: 'pending',
        待处理: 'pending',
        已通过: 'approved',
        已驳回: 'rejected',
        未通过: 'rejected'
      }

      const backendStatusMap = this.data.roleStatusConfig
        && this.data.roleStatusConfig.statusMap
        && typeof this.data.roleStatusConfig.statusMap === 'object'
        ? this.data.roleStatusConfig.statusMap
        : {}

      return backendStatusMap[value] || statusMap[value] || value
    },

    isRoleApproved(status) {
      return this.normalizeRoleStatus(status) === 'approved'
    },

    isRolePending(status) {
      return ROLE_PENDING_STATUSES.includes(this.normalizeRoleStatus(status))
    },

    isRoleTerminal(status) {
      const normalizedStatus = this.normalizeRoleStatus(status)

      return normalizedStatus === 'approved' || normalizedStatus === 'rejected'
    },

    findRoleApplication(roleType, applications = this.data.roleApplicationList, status = '') {
      const normalizedRoleType = this.normalizeRoleType(roleType)
      const normalizedStatus = status ? this.normalizeRoleStatus(status) : ''

      return (applications || []).find((item) => {
        if (item.roleType !== normalizedRoleType) {
          return false
        }

        return normalizedStatus ? this.normalizeRoleStatus(item.status) === normalizedStatus : true
      }) || null
    },

    isRoleResultViewed(roleType, status, application = null) {
      const normalizedRoleType = this.normalizeRoleType(roleType)
      const normalizedStatus = this.normalizeRoleStatus(status)

      if (normalizedStatus !== 'approved' && normalizedStatus !== 'rejected') {
        return false
      }

      return isRoleApplyResultViewed({
        roleType: normalizedRoleType,
        status: normalizedStatus,
        application: application || { id: `${normalizedRoleType}-${normalizedStatus}` }
      })
    },

    getUnviewedRoleResultApplication(roleType, applications = this.data.roleApplicationList, roleStatusState = this.data.roleStatusState) {
      const normalizedRoleType = this.normalizeRoleType(roleType)
      const stateStatus = this.normalizeRoleStatus((roleStatusState || {})[normalizedRoleType])

      if (this.isRolePending(stateStatus)) {
        return null
      }

      const application = this.isRoleTerminal(stateStatus)
        ? this.findRoleApplication(normalizedRoleType, applications, stateStatus)
        : (
          this.findRoleApplication(normalizedRoleType, applications, 'approved') ||
          this.findRoleApplication(normalizedRoleType, applications, 'rejected')
        )

      if (!application) {
        return null
      }

      const status = this.normalizeRoleStatus(
        this.isRoleTerminal(stateStatus)
          ? stateStatus
          : (application && application.status)
      )

      if (!this.isRoleTerminal(status) || this.isRolePending(status)) {
        return null
      }

      if (this.isRoleResultViewed(normalizedRoleType, status, application)) {
        return null
      }

      return Object.assign({}, application || {}, {
        roleType: normalizedRoleType,
        status
      })
    },

    hasUnviewedRoleResult(roleType, applications = this.data.roleApplicationList, roleStatusState = this.data.roleStatusState) {
      return Boolean(this.getUnviewedRoleResultApplication(roleType, applications, roleStatusState))
    },

    roleHomeRoute(roleType) {
      const roleRoutes = {
        player: ROUTES.playerHome,
        expert: ROUTES.expertHome,
        guide: ROUTES.guideHome
      }

      return roleRoutes[this.normalizeRoleType(roleType)] || ROUTES.playerHome
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

    formatRoleTags(
      activeRole,
      roleStatusState = this.data.roleStatusState,
      applications = this.data.roleApplicationList
    ) {
      return ROLE_TAGS.map((item) => {
        const resultApplication = this.getUnviewedRoleResultApplication(item.key, applications, roleStatusState)

        return {
          ...item,
          name: this.formatRoleTagName(item, roleStatusState, applications),
          status: this.normalizeRoleStatus(roleStatusState[item.key]),
          resultStatus: resultApplication ? this.normalizeRoleStatus(resultApplication.status) : '',
          active: item.key === activeRole
        }
      })
    },

    formatRoleTagName(item, roleStatusState = this.data.roleStatusState, applications = this.data.roleApplicationList) {
      if (this.isRolePending(roleStatusState[item.key])) {
        return '审核中'
      }

      const resultApplication = this.getUnviewedRoleResultApplication(item.key, applications, roleStatusState)

      if (resultApplication && resultApplication.status === 'approved') {
        return '已通过'
      }

      if (resultApplication && resultApplication.status === 'rejected') {
        return '已驳回'
      }

      return item.name
    },

    formatRolePermissionPrompt(
      roleType,
      roleStatusState = this.data.roleStatusState,
      currentRoleType = this.data.currentRoleType,
      applications = this.data.roleApplicationList
    ) {
      if (
        !roleType ||
        roleType === currentRoleType ||
        roleType === 'player' ||
        this.isRolePending(roleStatusState[roleType]) ||
        this.isRoleApproved(roleStatusState[roleType]) ||
        this.hasUnviewedRoleResult(roleType, applications, roleStatusState)
      ) {
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
      return this.formatRolePermissionPromptWithPrompts(roleType, roleStatusState, currentRoleType, prompts, applications)
    },

    formatRolePermissionPromptWithPrompts(
      roleType,
      roleStatusState = this.data.roleStatusState,
      currentRoleType = this.data.currentRoleType,
      prompts = ROLE_PERMISSION_PROMPTS,
      applications = this.data.roleApplicationList
    ) {
      if (
        !roleType ||
        roleType === currentRoleType ||
        roleType === 'player' ||
        this.isRolePending(roleStatusState[roleType]) ||
        this.isRoleApproved(roleStatusState[roleType]) ||
        this.hasUnviewedRoleResult(roleType, applications, roleStatusState)
      ) {
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

    resolveSelectedRole(activeRole, roleStatusState = this.data.roleStatusState, applications = []) {
      if (this.isRolePending(roleStatusState[activeRole])) {
        return activeRole
      }

      if (this.hasUnviewedRoleResult(activeRole, applications, roleStatusState)) {
        return activeRole
      }

      const pendingApplication = applications.find((item) => this.isRolePending(item.status))

      if (pendingApplication && pendingApplication.roleType) {
        return pendingApplication.roleType
      }

      const resultApplication = applications.find((item) => (
        this.isRoleTerminal(item.status) &&
        !this.isRoleResultViewed(item.roleType, item.status, item)
      ))

      if (resultApplication && resultApplication.roleType) {
        return resultApplication.roleType
      }

      const pendingRole = ROLE_TAGS.find((item) => this.isRolePending(roleStatusState[item.key]))

      if (pendingRole) {
        return pendingRole.key
      }

      const resultRole = ROLE_TAGS.find((item) => this.hasUnviewedRoleResult(item.key, applications, roleStatusState))

      return resultRole ? resultRole.key : activeRole
    },

    formatRoleAuditPrompt(
      roleType,
      roleStatusState = this.data.roleStatusState,
      applications = this.data.roleApplicationList,
      roleStatusConfig = this.data.roleStatusConfig
    ) {
      const normalizedRoleType = this.normalizeRoleType(roleType)
      const status = roleStatusState[normalizedRoleType]
      const resultApplication = this.getUnviewedRoleResultApplication(normalizedRoleType, applications, roleStatusState)
      const resultStatus = resultApplication ? this.normalizeRoleStatus(resultApplication.status) : ''

      if (!this.isRolePending(status) && !resultApplication) {
        return createEmptyRoleAuditPrompt()
      }

      if (!this.isValidRoleStatusConfig(roleStatusConfig)) {
        return createEmptyRoleAuditPrompt()
      }

      const texts = this.getRoleStatusConfigTexts(roleStatusConfig)
      const roleMeta = this.getRoleStatusConfigMeta(roleStatusConfig, normalizedRoleType)
      const roleName = roleMeta.roleName || ROLE_NAMES[normalizedRoleType] || ''
      const displayStatus = resultStatus || 'pending'
      const application = resultApplication || applications.find((item) => item.roleType === normalizedRoleType) || {}

      return {
        visible: true,
        roleType: normalizedRoleType,
        roleName,
        status: displayStatus,
        title: this.formatRoleAuditTitle(roleName, displayStatus, texts),
        icon: this.formatRoleAuditIcon(roleStatusConfig, displayStatus),
        submittedText: resultStatus
          ? this.formatRoleAuditResultText(application, texts)
          : this.formatRoleAuditSubmittedText(application, texts),
        detailText: this.formatRoleAuditDetailText(displayStatus, texts),
        helperText: this.formatRoleAuditHelper(displayStatus, texts),
        done: resultStatus === 'approved',
        resultStatus,
        steps: this.formatRoleAuditSteps(displayStatus, texts)
      }
    },

    isValidRoleStatusConfig(config = {}) {
      return !!(
        config
        && typeof config === 'object'
        && config.texts
        && typeof config.texts === 'object'
        && config.roleMeta
        && typeof config.roleMeta === 'object'
      )
    },

    getRoleStatusConfigTexts(config = this.data.roleStatusConfig) {
      return (config && config.texts && typeof config.texts === 'object') ? config.texts : {}
    },

    getRoleStatusConfigMeta(config = this.data.roleStatusConfig, roleType = 'expert') {
      const roleMeta = (config && config.roleMeta && typeof config.roleMeta === 'object') ? config.roleMeta : {}
      const normalizedRoleType = this.normalizeRoleType(roleType)

      return (roleMeta[normalizedRoleType] && typeof roleMeta[normalizedRoleType] === 'object')
        ? roleMeta[normalizedRoleType]
        : {}
    },

    formatRoleAuditTitle(roleName = '', status = 'pending', texts = {}) {
      if (status === 'approved') {
        return `${roleName}身份申请${this.compactAuditTitle(texts.approvedTitle, '审核通过')}`
      }

      if (status === 'rejected') {
        return `${roleName}身份申请${this.compactAuditTitle(texts.rejectedTitle, '审核未通过')}`
      }

      return `${roleName}身份申请${texts.pendingTitle || '审核中'}`
    },

    compactAuditTitle(text = '', fallback = '') {
      const value = String(text || '')
        .replace(/^恭喜/, '')
        .replace(/[！!。.]$/, '')
        .trim()

      return value || fallback
    },

    formatRoleAuditIcon(config = {}, status = 'pending') {
      const icons = config && config.icons && typeof config.icons === 'object' ? config.icons : {}

      return icons[status] || ROLE_AUDIT_DEFAULT_ICONS[status] || ''
    },

    formatRoleAuditResultText(application = {}, texts = {}) {
      const reviewedAt = this.formatCompactDateTime(application.reviewedAt)
      const resultTitle = texts.resultPageTitle || '审核结果'

      return reviewedAt ? `${resultTitle}已于 ${reviewedAt} 更新` : (texts.backendRecordFallback || resultTitle)
    },

    formatRoleAuditDetailText(status = 'pending', texts = {}) {
      const title = status === 'pending' ? texts.pendingPageTitle : texts.resultPageTitle

      return title ? `查看${title} →` : (status === 'pending' ? '查看详细进度 →' : '查看审核结果 →')
    },

    formatRoleAuditHelper(status = 'pending', texts = {}) {
      if (status === 'approved') {
        return texts.approvedAuditDesc || ''
      }

      if (status === 'rejected') {
        return texts.rejectedSubtitle || texts.rejectedDesc || ''
      }

      return texts.pendingHelper || texts.pendingDesc || ''
    },

    formatRoleAuditSteps(status = 'pending', texts = {}) {
      if (status === 'approved') {
        return [
          { key: 'submitted', name: texts.submittedFallback || '已提交', state: 'done' },
          { key: 'reviewing', name: '已审核', state: 'done' },
          { key: 'approved', name: '已通过', state: 'done' }
        ]
      }

      if (status === 'rejected') {
        return [
          { key: 'submitted', name: texts.submittedFallback || '已提交', state: 'done' },
          { key: 'reviewing', name: '已审核', state: 'done' },
          { key: 'rejected', name: '已驳回', state: 'active rejected' }
        ]
      }

      return [
        { key: 'submitted', name: texts.submittedFallback || '已提交', state: 'done' },
        { key: 'reviewing', name: texts.pendingTitle || '审核中', state: 'active' },
        { key: 'approved', name: '已通过', state: 'waiting' }
      ]
    },

    formatRoleAuditSubmittedText(application = {}, texts = {}) {
      const submittedAt = this.formatCompactDateTime(application.submittedAt)
      const expectedReviewAt = this.formatCompactDateTime(application.expectedReviewAt)

      if (!submittedAt && !expectedReviewAt) {
        return texts.pendingDesc || ''
      }

      if (!submittedAt) {
        return `身份申请已提交 预计 ${expectedReviewAt} 完成审核`
      }

      if (!expectedReviewAt) {
        return `你于 ${submittedAt} 提交了身份申请`
      }

      return `你于 ${submittedAt} 提交了身份申请 预计 ${expectedReviewAt} 完成审核`
    },

    formatCompactDateTime(value) {
      if (!value || typeof value !== 'string') {
        return ''
      }

      return value
        .replace(/-/g, '.')
        .replace('T', ' ')
        .replace(/\+.*$/, '')
        .replace(/Z$/, '')
        .replace(/(\d{2}:\d{2}):\d{2}(?:\.\d+)?$/, '$1')
        .trim()
    },

    async loadRoleBenefitConfig() {
      try {
        const config = extractResponseData(await roleService.getRoleBenefitConfig())
        const prompts = (config && config.permissionPrompts) || ROLE_PERMISSION_PROMPTS
        const selectedRole = this.data.selectedRoleTag || this.data.currentRoleType

        this.setData({
          roleBenefitConfig: {
            ...config,
            permissionPrompts: prompts
          },
          rolePermissionPrompt: this.formatRolePermissionPromptWithPrompts(
            selectedRole,
            this.data.roleStatusState,
            this.data.currentRoleType,
            prompts
          )
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

    handleHomeRetryTap() {
      if (this.data.loadErrorAuthExpired) {
        goLogin(this.homeRoute())
        return
      }

      const roleType = this.normalizeRoleType(this.data.currentRoleType || this.properties.roleType)

      this.setData({
        loadError: '',
        loadErrorActionText: '重试',
        loadErrorAuthExpired: false
      })
      this.loadRoleHome(roleType)
    },

    handlePlayerNameTap() {
      const name = this.data.playerCard && this.data.playerCard.name

      if (!name) {
        return
      }

      wx.showToast({
        title: name,
        icon: 'none'
      })
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
      if (this.roleHomeDisposed) {
        return
      }

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
      if (this.roleHomeDisposed) {
        return
      }

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
      if (key === 'nearby') {
        this.loadHomeNearbyGames()
        return
      }
      const mainVisibleState = this.formatMainVisibleState(this.data.mainGames, key)

      this.setData({
        mainFilter: key,
        mainTabs: this.data.mainTabs.map((item) => ({
          ...item,
          active: item.key === key
        })),
        mainVisibleGames: mainVisibleState.games,
        mainEmptyText: mainVisibleState.emptyText
      })
    },

    async loadHomeNearbyGames() {
      try {
        const location = await locationAccess.getPreciseLocation()
        const data = await locationService.getNearbyGames({ latitude: location.latitude, longitude: location.longitude, radiusMeters: 5000 })
        const nearbyGames = this.formatGameCards(data.items || data.games || [])
        const cityGames = (this.data.mainGames || []).filter((item) => item.scope === 'city' || (item.scopes || []).indexOf('city') >= 0)
        const mainGames = this.mergeMainGameGroups([{ scope: 'nearby', items: nearbyGames }, { scope: 'city', items: cityGames }])
        const mainVisibleState = this.formatMainVisibleState(mainGames, 'nearby')
        this.setData({ mainGames, mainVisibleGames: mainVisibleState.games, mainEmptyText: mainVisibleState.emptyText, mainFilter: 'nearby', mainTabs: this.data.mainTabs.map((item) => ({ ...item, active: item.key === 'nearby' })), locationGuideVisible: false })
      } catch (error) {
        this.setData({ locationGuideVisible: true })
      }
    },

    async handleHomeLocationGuideTap(event) {
      const action = event.currentTarget.dataset.action
      if (action === 'setting') return locationAccess.showDeniedGuide({ onManual: () => this.loadHomeManualLocation(), onFallback: () => this.loadHomeCityFallback() })
      if (action === 'manual') return this.loadHomeManualLocation()
      if (action === 'city') return this.loadHomeCityFallback()
    },

    async loadHomeManualLocation() {
      try {
        const location = await locationAccess.chooseManualLocation()
        const data = await locationService.getNearbyGames({ latitude: location.latitude, longitude: location.longitude, radiusMeters: 5000 })
        const mainGames = this.formatGameCards(data.items || data.games || [])
        const mainVisibleState = this.formatMainVisibleState(mainGames, 'nearby')
        this.setData({ mainGames, mainVisibleGames: mainVisibleState.games, mainEmptyText: mainVisibleState.emptyText, mainFilter: 'nearby', locationGuideVisible: false })
      } catch (error) { wx.showToast({ title: error.message || '未选择位置', icon: 'none' }) }
    },

    async loadHomeCityFallback() {
      this.setData({ locationGuideVisible: false })
      const mainVisibleState = this.formatMainVisibleState(this.data.mainGames, 'city')
      this.setData({ mainFilter: 'city', mainVisibleGames: mainVisibleState.games, mainEmptyText: mainVisibleState.emptyText, mainTabs: this.data.mainTabs.map((item) => ({ ...item, active: item.key === 'city' })) })
    },

    handleRoleTagTap(event) {
      const key = this.normalizeRoleType(event.currentTarget.dataset.key || this.data.currentRoleType)
      const currentRoleType = this.data.currentRoleType
      const roleStatusState = this.data.roleStatusState || DEFAULT_ROLE_STATUS_STATE
      const applications = this.data.roleApplicationList || []
      const currentRoute = this.homeRoute()

      if (key === currentRoleType) {
        this.setData({
          selectedRoleTag: currentRoleType,
          roleTags: this.formatRoleTags(currentRoleType, roleStatusState, applications),
          rolePermissionPrompt: this.formatRolePermissionPrompt('', roleStatusState, currentRoleType),
          roleAuditPrompt: this.formatRoleAuditPrompt(currentRoleType, roleStatusState, applications)
        })
        return
      }

      if (this.isRoleApproved(roleStatusState[key]) && !this.hasUnviewedRoleResult(key, applications, roleStatusState)) {
        navigateShellRoute(this.roleHomeRoute(key), {
          currentRoute
        })
        return
      }

      this.setData({
        selectedRoleTag: key,
        roleTags: this.formatRoleTags(key, roleStatusState, applications),
        rolePermissionPrompt: this.formatRolePermissionPrompt(key, roleStatusState, currentRoleType, applications),
        roleAuditPrompt: this.formatRoleAuditPrompt(key, roleStatusState, applications)
      })
    },

    handleRoleApplyTap() {
      const prompt = this.data.rolePermissionPrompt || {}
      const roleType = prompt.roleType || this.data.selectedRoleTag
      const normalizedRoleType = this.normalizeRoleType(roleType)
      const returnRoute = this.homeRoute()
      let url = `/${ROUTES.roleApply}?roleType=${roleType}`

      if (normalizedRoleType === 'expert') {
        url = `/${ROUTES.roleFlow}?mode=expertApplyOverview&single=1&roleType=expert&returnTo=${encodeURIComponent(returnRoute)}`
      } else if (normalizedRoleType === 'guide') {
        url = `/${ROUTES.roleFlow}?mode=roleApplyOverview&single=1&roleType=guide&returnTo=${encodeURIComponent(returnRoute)}`
      }

      navigateShellRoute(url, {
        currentRoute: this.homeRoute()
      })
    },

    handleRoleCompareTap() {
      const prompt = this.data.rolePermissionPrompt || {}
      const roleType = this.normalizeRoleType(prompt.roleType || this.data.selectedRoleTag || this.data.currentRoleType || this.properties.roleType)
      const returnRoute = this.homeRoute()

      navigateShellRoute(`${ROUTES.roleFlow}?mode=roleComparison&single=1&roleType=${roleType}&returnTo=${encodeURIComponent(returnRoute)}`, {
        currentRoute: this.homeRoute()
      })
    },

    handleRoleAuditDetailTap() {
      const prompt = this.data.roleAuditPrompt || {}
      const roleType = prompt.roleType || this.data.selectedRoleTag || 'guide'
      const status = this.normalizeRoleStatus(prompt.resultStatus || prompt.status)

      if (status === 'approved') {
        const applications = this.data.roleApplicationList || []
        const application = applications.find((item) => (
          this.normalizeRoleType(item && item.roleType) === this.normalizeRoleType(roleType)
          && this.normalizeRoleStatus(item && item.status) === 'approved'
        )) || {}

        markRoleApplyResultViewed({ roleType, status, application })
        setActiveRole(roleType)
        navigateShellRoute(this.roleHomeRoute(roleType), {
          currentRoute: this.homeRoute(),
          reuseExisting: false
        })
        return
      }

      navigateShellRoute(`${ROUTES.roleStatus}?roleType=${roleType}&status=${status || 'pending'}`, {
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

    handleTaskCenterTap() {
      navigateShellRoute(ROUTES.profileTaskCenter, {
        currentRoute: this.homeRoute()
      })
    },

    handleActionTap(event) {
      const datasetRoute = event && event.currentTarget && event.currentTarget.dataset && event.currentTarget.dataset.route
      const datasetTitle = event && event.currentTarget && event.currentTarget.dataset && event.currentTarget.dataset.title
      const detail = event && event.detail || {}
      const detailRoute = detail.route || (detail.item && detail.item.route)

      const rawRoute = String(datasetRoute || detailRoute || '').replace(/^\/+/, '')
      const rawTitle = String(datasetTitle || '').trim()
      if (rawRoute === ROUTES.metaverse || rawTitle.indexOf('元宇宙') !== -1) {
        wx.showToast({ title: '元宇宙玩法暂未开放', icon: 'none' })
        return
      }
      if (rawRoute === ROUTES.map || rawTitle.indexOf('地图') !== -1 || rawTitle.indexOf('地球') !== -1) {
        wx.showToast({ title: '地图玩法暂未开放', icon: 'none' })
        return
      }
      const route = this.resolveHomeActionRoute(datasetRoute || detailRoute, datasetTitle)

      if (route) {
        navigateShellRoute(this.appendRoleTypeToGameDetailRoute(route), {
          currentRoute: this.homeRoute()
        })
        return
      }

      if (detail.item) {
        this.navigateToGameCardRoute(ROUTES.gameDetail, detail.item, 'id')
        return
      }

      wx.showToast({
        title: '请选择可用入口',
        icon: 'none'
      })
    },

    resolveHomeActionRoute(route, title) {
      const normalizedRoute = String(route || '').replace(/^\/+/, '')
      if (normalizedRoute) {
        return normalizedRoute
      }

      const normalizedTitle = String(title || '').trim()
      const fallbackRoutes = {
        '发起组局': ROUTES.gameCreate,
        '局前大厅': ROUTES.gameHall,
        '我的邀约': 'pages/profile/service-center/invite/overview/index',
        '查看全部榜单': ROUTES.homeRanking
      }

      return fallbackRoutes[normalizedTitle] || ''
    },

    handleGameCardAction(event) {
      const detail = event && event.detail || {}
      const item = detail.item || {}
      const action = normalizeGameCardAction(detail.action, detail.label)

      if (action === 'primary') {
        this.handleGameCardPrimaryAction(item, detail.label)
        return
      }
      if (action === 'share') {
        this.navigateToGameCardRoute(ROUTES.gameShare, item, 'id')
        return
      }
      if (action === 'follow') {
        this.followGameCard(item)
        return
      }
      if (action === 'refer') {
        this.navigateToGameCardRoute(ROUTES.gameInvite, item)
        return
      }
      if (action === 'greet') {
        this.navigateToGameCardRoute(ROUTES.gameDetail, item, 'gameId', { intent: 'greet' })
        return
      }

      wx.showToast({ title: '暂无可执行操作', icon: 'none' })
    },

    handleGameCardPrimaryAction(item = {}, label = '') {
      const text = String(label || item.primaryActionText || item.action || '').trim()
      const actionType = String(item.primaryActionType || '').trim().toLowerCase()

      if (item.primaryActionDisabled) {
        wx.showToast({
          title: item.primaryActionToast || text || '当前组局暂不可操作',
          icon: 'none'
        })
        return
      }

      if (item.primaryActionRoute) {
        navigateShellRoute(this.appendRoleTypeToGameDetailRoute(item.primaryActionRoute), {
          currentRoute: this.homeRoute()
        })
        return
      }

      if (isDisabledActionText(text)) {
        wx.showToast({ title: '当前组局暂不可加入', icon: 'none' })
        return
      }

      if (actionType === 'apply' || actionType === 'join' || actionType === 'view' || actionType === 'detail' || isJoinActionText(text)) {
        this.navigateToGameCardRoute(ROUTES.gameDetail, item, 'id')
        return
      }

      this.navigateToGameCardRoute(ROUTES.gameDetail, item, 'id')
    },

    navigateToGameCardRoute(route, item = {}, idParam = 'gameId', extraParams = {}) {
      const gameId = this.resolveGameCardId(item)

      if (gameId == null || gameId === '') {
        wx.showToast({ title: '缺少组局信息', icon: 'none' })
        return
      }

      const query = Object.keys(extraParams || {}).reduce((items, key) => {
        const value = extraParams[key]
        if (value !== undefined && value !== null && value !== '') {
          items.push(`${encodeURIComponent(key)}=${encodeURIComponent(value)}`)
        }
        return items
      }, [`${idParam}=${encodeURIComponent(gameId)}`])
      const targetRoute = this.appendRoleTypeToGameDetailRoute(
        `${route}?${query.join('&')}`
      )
      navigateShellRoute(targetRoute, { currentRoute: this.homeRoute() })
    },

    async followGameCard(item = {}) {
      const gameId = this.resolveGameCardId(item)

      if (gameId == null || gameId === '') {
        wx.showToast({ title: '缺少组局信息', icon: 'none' })
        return
      }

      try {
        await gameService.favoriteGame(gameId)
        wx.showToast({ title: '已关注该组局', icon: 'none' })
      } catch (error) {
        wx.showToast({ title: error.message || '关注失败', icon: 'none' })
      }
    }
  }
})

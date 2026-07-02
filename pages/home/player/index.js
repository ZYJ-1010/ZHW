const homeService = require('../../../services/home')
const roleService = require('../../../services/role')
const { ROUTES } = require('../../../config/routes')
const toast = require('../../../utils/toast')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')

const HOME_SCROLL_TAP_STEP_RPX = 360
const HOME_SCROLL_HOLD_STEP_RPX = 72
const HOME_SCROLL_HOLD_INTERVAL_MS = 80
const HOME_SCROLL_HOLD_SUPPRESS_TAP_MS = 120

const ROLE_TYPE_MAP = {
  player: 'player',
  玩家: 'player',
  expert: 'expert',
  master: 'expert',
  行家: 'expert',
  guide: 'guide',
  leader: 'guide',
  领路人: 'guide'
}

const ROLE_TAGS = [
  { key: 'player', icon: '🎮', name: '玩家' },
  { key: 'expert', icon: '🎯', name: '行家' },
  { key: 'guide', icon: '🌐', name: '领路人' }
]

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

const ROLE_PENDING_STATUSES = ['pending', 'reviewing', 'auditing', 'pending_audit']

const DEFAULT_ROLE_STATUS_STATE = {
  player: 'approved',
  expert: 'none',
  guide: 'none'
}

const ROLE_AUDIT_PROMPTS = {
  expert: {
    roleType: 'expert',
    roleName: '行家',
    title: '行家身份申请审核中',
    icon: '⏳',
    detailText: '查看详细进度 →',
    helperText: '审核期间你的玩家身份不受影响，所有功能不受影响'
  },
  guide: {
    roleType: 'guide',
    roleName: '领路人',
    title: '领路人身份申请审核中',
    icon: '⏳',
    detailText: '查看详细进度 →',
    helperText: '审核期间你的玩家身份不受影响，所有功能不受影响'
  }
}

const ROLE_AUDIT_STEPS = [
  { key: 'submitted', name: '已提交', state: 'done' },
  { key: 'reviewing', name: '审核中', state: 'active' },
  { key: 'approved', name: '已通过', state: 'waiting' }
]

function createEmptyRoleAuditPrompt() {
  return {
    visible: false,
    roleType: '',
    title: '',
    icon: '',
    submittedText: '',
    detailText: '',
    helperText: '',
    done: false,
    steps: []
  }
}

Page({
  data: {
    onlineText: '在线',
    roleType: 'player',
    canvasStyle: '',
    contentStyle: '',
    dockStyle: '',
    homeScrollTop: 0,
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ],
    hero: {
      title: 'HELLO, 玩家!',
      date: '2026.03.30 | 开启你的今日副本'
    },
    nearbySection: {
      title: '附近正在发生'
    },
    nearbyTabs: [
      { key: 'all', name: '全部', active: true },
      { key: 'nearby', name: '附近', active: false }
    ],
    nearbyFilter: 'all',
    nearbyGames: [
      {
        id: 'home-nearby-001',
        title: '苏州河“记忆碎片”采集',
        scope: 'nearby',
        tag: '探索局',
        coverSrc: '/components/game-card/assets/cover-city.png',
        price: '￥0/人',
        action: '加入',
        location: '📍静安区 · 3.2km · 5/8人',
        time: '⏰2026年5月1日 20:00--22:00',
        joinedText: '+5位玩家已入局',
        imageText: '河',
        imageTone: 'small',
        actions: ['分享', '关注', '引荐', '打招呼']
      },
      {
        id: 'home-nearby-002',
        title: 'AI赋能系统搭建交流局',
        scope: 'city',
        tag: '任务局',
        coverSrc: '/components/game-card/assets/cover-sunset.png',
        price: '￥0/人',
        action: '加入',
        location: '📍黄浦区 · 8.2km · 3/8人',
        time: '⏰2026年5月1日 14:00--16:00',
        joinedText: '+3位玩家已入局',
        imageText: 'AI',
        imageTone: '',
        actions: ['分享', '关注', '引荐', '打招呼']
      }
    ],
    rankingSection: {
      icon: '🏆',
      title: '本周玩霸榜',
      moreText: '查看全部榜单',
      route: ROUTES.profileFootprintAchievements
    },
    rankingActiveRole: 'player',
    rankingTabs: [
      { key: 'player', name: '玩家', active: true },
      { key: 'expert', name: '行家', active: false },
      { key: 'guide', name: '领路人', active: false }
    ],
    rankingBoards: {
      player: {
        list: [
          { rank: '01', avatarFallback: '👨🏾‍🎓', name: '领域专家 PRO', desc: '本周组局 12 · MVP 5次', xp: '2,450', xpUnit: 'XP' },
          { rank: '02', avatarFallback: '👩🏻‍🎤', name: '社交达人', desc: '本周组局 8 次', xp: '1,890', xpUnit: 'XP' },
          { rank: '03', avatarFallback: '👨🏿‍🚀', name: '探险家', desc: '本周组局 6 次', xp: '1,560', xpUnit: 'XP' }
        ],
        myRank: {
          rank: '52',
          avatarFallback: '👩🏻‍💻',
          name: '我（Alex）',
          desc: '上周排名 65 ↑',
          xp: '520',
          xpUnit: 'XP'
        }
      }
    },
    rankingList: [
      { rank: '01', avatarFallback: '👨🏾‍🎓', name: '领域专家 PRO', desc: '本周组局 12 · MVP 5次', xp: '2,450', xpUnit: 'XP' },
      { rank: '02', avatarFallback: '👩🏻‍🎤', name: '社交达人', desc: '本周组局 8 次', xp: '1,890', xpUnit: 'XP' },
      { rank: '03', avatarFallback: '👨🏿‍🚀', name: '探险家', desc: '本周组局 6 次', xp: '1,560', xpUnit: 'XP' }
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
    playerCard: {
      role: '玩家 Lv.5',
      title: '活跃达人',
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
    roleTags: ROLE_TAGS.map((item) => ({
      ...item,
      active: item.key === 'player'
    })),
    selectedRoleTag: 'player',
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
    roleStatusState: DEFAULT_ROLE_STATUS_STATE,
    roleApplicationList: [],
    roleAuditPrompt: createEmptyRoleAuditPrompt(),
    onlineCard: {
      title: '地球online',
      desc: '探索城市副本 · 解锁地图成就',
      tags: ['附近 12 个组局', '已打卡 8 处']
    },
    entries: [
      { title: '发起组局', desc: '创建你的带局房间', icon: '📍', theme: 'pink', route: ROUTES.gameCreate },
      { title: '局前大厅', desc: '准备就绪加入一局', icon: '', routeIcon: true, theme: 'cyan', route: ROUTES.gameHall }
    ],
    friendSection: {
      icon: '🎲',
      title: '朋友在玩',
      count: 2,
      moreText: '查看全部',
      route: ROUTES.gameHall
    },
    friendGames: [
      {
        id: 'home-friend-001',
        title: '盲盒路线：3小时点亮天际线',
        tag: '探索局',
        coverSrc: '/components/game-card/assets/cover-sunset.png',
        price: '￥29/人',
        action: '加入',
        location: '📍梧桐山 · 1.5km · 3/8人',
        time: '2026年5月1日 14:00--16:00',
        joinedText: '+3位玩家已入局',
        imageText: '线',
        imageTone: 'warm',
        actions: ['分享', '关注', '引荐', '打招呼']
      }
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
      badge: '+99',
      route: 'pages/placeholder/metaverse/index'
    }
  },

  onLoad() {
    this.loadRoleBenefitConfig()
    this.loadPlayerHome()
  },

  async loadRoleBenefitConfig() {
    try {
      const config = await roleService.getRoleBenefitConfig()
      const prompts = (config && config.permissionPrompts) || ROLE_PERMISSION_PROMPTS
      const selectedRole = this.data.selectedRoleTag || 'player'

      this.setData({
        roleBenefitConfig: {
          ...config,
          permissionPrompts: prompts
        },
        rolePermissionPrompt: this.formatRolePermissionPromptWithPrompts(selectedRole, this.data.roleStatusState, prompts)
      })
    } catch (error) {
      this.setData({
        roleBenefitConfig: {
          permissionPrompts: ROLE_PERMISSION_PROMPTS
        }
      })
    }
  },

  onHide() {
    this.clearHomeScrollTimers()
  },

  onUnload() {
    this.clearHomeScrollTimers()
  },

  async loadPlayerHome() {
    try {
      const home = await homeService.getHome()
      const hero = home && home.hero ? home.hero : {}
      const currentUser = home && home.user ? home.user : {}
      const playerSummary = home && home.playerSummary ? home.playerSummary : {}
      const nearbySummary = home && home.nearbySummary ? home.nearbySummary : {}
      const nearbySection = home && home.nearbySection ? home.nearbySection : {}
      const friendSection = home && home.friendSection ? home.friendSection : {}
      const earth = home && home.earth ? home.earth : {}
      const visualization = home && home.visualization ? home.visualization : {}
      const achievementState = this.formatAchievementState(home || {})
      const rankingState = this.formatRankingState(home || {})
      const roleStatusState = this.formatRoleStatusState(home || {})
      const roleApplicationList = this.formatRoleApplicationList(home || {})
      const roleName = hero.roleName || '玩家'
      const date = this.formatHeroDate(hero.dateLabel, hero.subtitle) || this.data.hero.date
      const activeRole = this.normalizeRoleType(
        playerSummary.currentRole || playerSummary.roleType || hero.currentRole || hero.roleType || roleName
      )
      const selectedRole = this.resolveSelectedRole(activeRole, roleStatusState, roleApplicationList)
      const nearbyGames = this.formatGameCards(home.nearbyGames || home.recommendedGames, this.data.nearbyGames)
      const friendGames = this.formatGameCards(home.friendGames, this.data.friendGames)

      this.setData({
        onlineText: hero.onlineText || this.data.onlineText,
        roleType: activeRole,
        selectedRoleTag: selectedRole,
        hero: {
          ...this.data.hero,
          title: `HELLO, ${roleName}!`,
          date
        },
        playerCard: this.formatPlayerCard(playerSummary),
        roleStatusState,
        roleApplicationList,
        roleTags: this.formatRoleTags(selectedRole, roleStatusState),
        rolePermissionPrompt: this.formatRolePermissionPrompt(selectedRole, roleStatusState),
        roleAuditPrompt: this.formatRoleAuditPrompt(selectedRole, roleStatusState, roleApplicationList, currentUser),
        onlineCard: this.formatOnlineCard(nearbySummary, earth),
        nearbySection: this.formatNearbySection(nearbySection),
        nearbyTabs: this.formatNearbyTabs(nearbySection.tabs || nearbySection.filters),
        nearbyGames,
        friendSection: this.formatFriendSection(friendSection, friendGames),
        friendGames,
        rankingSection: rankingState.section,
        rankingActiveRole: rankingState.activeKey,
        rankingTabs: rankingState.tabs,
        rankingBoards: rankingState.boards,
        rankingList: rankingState.list,
        myRank: rankingState.myRank,
        achievementSection: achievementState.section,
        achievements: achievementState.list,
        metaverse: this.formatMetaverseEntry(home.metaverseEntry || home.metaverse || {}, visualization)
      })
    } catch (error) {
      // 首页静态内容可兜底展示，接口失败时不打断用户浏览。
    }
  },

  formatPlayerCard(playerSummary) {
    const current = this.data.playerCard

    if (!playerSummary || Object.keys(playerSummary).length === 0) {
      return current
    }

    return {
      role: playerSummary.roleLabel || this.formatRoleLabel(playerSummary, current.role),
      title: playerSummary.title || current.title,
      name: playerSummary.displayName || playerSummary.nickname || current.name,
      xp: playerSummary.xpText || this.formatXpText(playerSummary) || current.xp,
      progress: this.normalizeProgress(playerSummary.progressPercent, current.progress),
      next: playerSummary.nextLevelText || this.formatNextLevelText(playerSummary) || current.next,
      stats: this.formatPlayerStats(playerSummary, current.stats)
    }
  },

  formatRoleLabel(playerSummary, fallback) {
    if (playerSummary.level == null) {
      return fallback
    }

    const roleType = this.normalizeRoleType(playerSummary.currentRole || playerSummary.roleType)
    const roleTag = ROLE_TAGS.find((item) => item.key === roleType)
    const roleName = roleTag ? roleTag.name : '玩家'

    return `${roleName} Lv.${playerSummary.level}`
  },

  formatXpText(playerSummary) {
    const experience = playerSummary.experience
    const nextLevelExperience = playerSummary.nextLevelExperience

    if (experience == null || nextLevelExperience == null) {
      return ''
    }

    return `${experience}/${nextLevelExperience} XP`
  },

  formatNextLevelText(playerSummary) {
    const expToNextLevel = playerSummary.expToNextLevel

    if (expToNextLevel == null) {
      return ''
    }

    return `距离下一等级还需 ${expToNextLevel} 经验值`
  },

  formatPlayerStats(playerSummary, fallbackStats) {
    if (Array.isArray(playerSummary.stats) && playerSummary.stats.length > 0) {
      return playerSummary.stats.map((item) => ({
        value: this.formatStatValue(item.value),
        label: item.label
      }))
    }

    return [
      { value: this.formatStatValue(playerSummary.joinCount), label: '参与局数' },
      { value: this.formatStatValue(playerSummary.monthlyMvpCount), label: '本月MVP' },
      { value: this.formatParticipationRate(playerSummary.participationRate), label: '参与率' }
    ].map((item, index) => ({
      value: item.value || fallbackStats[index].value,
      label: item.label
    }))
  },

  formatOnlineCard(nearbySummary, earth = {}) {
    const current = this.data.onlineCard
    const nodes = Array.isArray(earth.nodes) ? earth.nodes : []
    const heatPoints = Array.isArray(earth.heatPoints) ? earth.heatPoints : []
    const nearbyGameCount = this.pickFirstValue(
      nearbySummary.nearbyGameCount,
      nearbySummary.count,
      nearbySummary.gameCount,
      earth.nearbyCount
    )
    const checkedInCount = this.pickFirstValue(
      nearbySummary.checkedInCount,
      nearbySummary.checkinCount,
      nearbySummary.visitedCount,
      heatPoints.length
    )
    const onlineNodeCount = this.pickFirstValue(
      nearbySummary.onlineNodeCount,
      earth.onlineCount,
      nodes.length
    )

    return {
      ...current,
      desc: onlineNodeCount ? `动态连接 ${this.formatStatValue(onlineNodeCount)} 个节点` : current.desc,
      tags: [
        `附近 ${this.formatStatValue(nearbyGameCount) || '12'} 个组局`,
        `已打卡 ${this.formatStatValue(checkedInCount) || '8'} 处`
      ],
      route: nearbySummary.route || earth.route || ROUTES.map
    }
  },

  formatNearbySection(section = {}) {
    return {
      ...this.data.nearbySection,
      title: section.title || section.name || this.data.nearbySection.title
    }
  },

  formatNearbyTabs(tabs) {
    const currentActive = this.data.nearbyFilter || 'all'
    const sourceTabs = Array.isArray(tabs) && tabs.length > 0 ? tabs : this.data.nearbyTabs

    return sourceTabs.map((item, index) => {
      const key = item.key || item.value || (index === 0 ? 'all' : `tab_${index}`)

      return {
        key,
        name: item.name || item.label || item.title || key,
        active: key === currentActive
      }
    })
  },

  formatFriendSection(section = {}, friendGames = this.data.friendGames) {
    const current = this.data.friendSection
    const count = this.pickFirstValue(
      section.count,
      section.friendCount,
      section.total,
      Array.isArray(friendGames) ? friendGames.length : current.count
    )

    return {
      ...current,
      icon: section.icon || current.icon,
      title: section.title || section.name || current.title,
      count,
      moreText: section.moreText || section.moreLabel || section.actionText || current.moreText,
      route: section.route || section.moreRoute || section.actionRoute || current.route || ROUTES.gameHall
    }
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

  formatMetaverseEntry(entry = {}, visualization = {}) {
    const current = this.data.metaverse
    const network = visualization.network || {}
    const networkNodes = Array.isArray(network.nodes) ? network.nodes : []
    const networkEdges = Array.isArray(network.edges) ? network.edges : []
    const tags = Array.isArray(entry.tags) && entry.tags.length > 0
      ? entry.tags
      : networkEdges.length > 0
        ? ['关系网', `${networkEdges.length} 条连接`]
        : this.splitMetaverseTags(entry.desc || entry.subtitle) || current.tags

    return {
      ...current,
      title: entry.actionText || entry.title || current.title,
      desc: entry.summary || entry.description || entry.desc || (networkNodes.length ? `已连接 ${networkNodes.length} 个动态节点` : current.desc),
      tags,
      avatars: this.formatMetaverseAvatars(entry.avatars || entry.users || networkNodes || current.avatars),
      badge: this.formatMetaverseBadge(entry, current.badge, networkNodes.length),
      route: entry.route || current.route
    }
  },

  formatMetaverseBadge(entry = {}, fallback = '', nodeCount = 0) {
    const joinedCount = this.pickFirstValue(entry.joinedCount, entry.onlineCount, entry.participantCount)

    if (joinedCount != null && joinedCount !== '') {
      return `+${joinedCount}`
    }

    if (nodeCount > 0) {
      return `+${nodeCount}`
    }

    return entry.badge || entry.badgeText || entry.onlineText || fallback
  },

  splitMetaverseTags(text) {
    if (!text || typeof text !== 'string') {
      return null
    }

    return text
      .split(/[·｜|、,，]/)
      .map((item) => item.trim())
      .filter(Boolean)
      .slice(0, 2)
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

  formatRankingState(home = {}) {
    const section = home.rankingSection || {}
    const requestedActiveKey = this.normalizeRoleType(
      section.defaultTab || section.activeKey || home.rankingActiveRole || this.data.rankingActiveRole
    )
    const tabs = this.formatRankingTabs(section.tabs || home.rankingTabs, requestedActiveKey)
    const boards = this.formatRankingBoards(home, this.data.rankingBoards)
    const activeKey = boards[requestedActiveKey] ? requestedActiveKey : (tabs[0] && tabs[0].key) || 'player'
    const activeBoard = boards[activeKey] || {}
    const display = this.formatRankingDisplay(
      activeBoard.list || this.data.rankingList,
      activeBoard.myRank || this.data.myRank
    )

    return {
      section: this.formatRankingSection(section),
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

  formatRankingSection(section = {}) {
    return {
      ...this.data.rankingSection,
      icon: section.icon || this.data.rankingSection.icon,
      title: section.title || section.name || this.data.rankingSection.title,
      moreText: section.moreText || section.moreLabel || section.actionText || this.data.rankingSection.moreText,
      route: section.route || section.moreRoute || section.actionRoute || this.data.rankingSection.route || ROUTES.profileFootprintAchievements
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

    if (Array.isArray(sourceBoards)) {
      sourceBoards.forEach((board, index) => {
        const roleKey = this.normalizeRoleType(
          board.key || board.value || board.roleType || board.name || (ROLE_TAGS[index] && ROLE_TAGS[index].key)
        )
        boards[roleKey] = this.formatRankingBoard(board, fallbackBoards[roleKey])
      })
    }

    if (Array.isArray(home.rankingList) && home.rankingList.length > 0 && !boards.player) {
      boards.player = this.formatRankingBoard(
        {
          list: home.rankingList,
          myRank: home.myRank || home.currentUserRank
        },
        fallbackBoards.player
      )
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
      desc: item.desc || item.description || item.summary || item.rankText || this.formatRankingDesc(item) || fallback.desc || '',
      xp: this.formatRankingXp(this.pickFirstValue(item.xp, item.xpText, item.experience, item.weeklyXp, fallback.xp)),
      xpUnit: item.xpUnit || fallback.xpUnit || 'XP'
    }
  },

  formatRankingDesc(item) {
    const gameCount = this.pickFirstValue(item.weeklyGameCount, item.gameCount, item.joinCount)
    const mvpCount = this.pickFirstValue(item.weeklyMvpCount, item.mvpCount)
    const parts = []

    if (gameCount != null && gameCount !== '') {
      parts.push(`本周组局 ${gameCount} 次`)
    }

    if (mvpCount != null && mvpCount !== '') {
      parts.push(`MVP ${mvpCount}次`)
    }

    return parts.join(' · ')
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

  isImagePath(value) {
    return typeof value === 'string' && (/^(https?:)?\/\//.test(value) || value.startsWith('/'))
  },

  makeAvatarFallback(name) {
    if (!name) {
      return ''
    }

    return name.trim().slice(0, 2)
  },

  formatGameCards(games, fallbackGames) {
    if (!Array.isArray(games) || games.length === 0) {
      return this.ensureGameCardRoutes(fallbackGames)
    }

    return games.map((item) => ({
      id: this.resolveGameCardId(item),
      scope: item.scope || item.distanceScope || 'city',
      title: item.title,
      tag: item.typeText || item.statusText || item.tag || '',
      type: item.type || item.statusType || '',
      coverSrc: item.coverUrl || item.coverSrc || item.cover,
      avatarUrls: item.participantAvatars || item.avatarUrls || item.participantAvatarUrls || [],
      price: item.priceText || item.price || '',
      action: item.actionText || item.action || '加入',
      location: item.locationText || this.formatGameLocation(item),
      time: item.timeText || item.time || '',
      joinedText: item.joinedText || this.formatJoinedText(item.joinedCount),
      actions: this.formatGameActions(item.actions),
      route: this.resolveGameCardRoute(item)
    }))
  },

  ensureGameCardRoutes(games) {
    if (!Array.isArray(games)) {
      return []
    }

    return games.map((item) => ({
      ...item,
      route: this.resolveGameCardRoute(item)
    }))
  },

  resolveGameCardRoute(item = {}) {
    const configuredRoute = item.detailRoute || item.detailUrl || item.route || item.url || item.path

    if (configuredRoute) {
      return this.normalizeGameCardRoute(configuredRoute, this.resolveGameCardId(item))
    }

    const gameId = this.resolveGameCardId(item)

    if (gameId != null && gameId !== '') {
      return `${ROUTES.gameDetail}?id=${encodeURIComponent(gameId)}`
    }

    return ROUTES.gameHall
  },

  resolveGameCardId(item = {}) {
    return this.pickFirstValue(item.gameId, item.id, item.gameID, item.game_id)
  },

  normalizeGameCardRoute(route, gameId) {
    const normalizedRoute = String(route || '').replace(/^\/+/, '')

    if (normalizedRoute === ROUTES.gameDetail && gameId != null && gameId !== '') {
      return `${ROUTES.gameDetail}?id=${encodeURIComponent(gameId)}`
    }

    return normalizedRoute
  },

  formatGameLocation(item) {
    const parts = [item.cityName || item.locationName, item.distanceText, item.memberText]
      .filter((value) => value != null && value !== '')

    return parts.length ? `📍${parts.join(' · ')}` : ''
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

  pickFirstValue(...values) {
    return values.find((value) => value != null && value !== '')
  },

  formatStatValue(value) {
    if (value == null || value === '') {
      return ''
    }

    return `${value}`
  },

  formatParticipationRate(value) {
    if (value == null || value === '') {
      return ''
    }

    if (typeof value === 'number') {
      return value <= 1 ? `${Math.round(value * 100)}%` : `${value}%`
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

  formatRoleStatusState(home = {}) {
    const user = home.user || {}
    const state = Object.assign({}, DEFAULT_ROLE_STATUS_STATE)

    this.mergeRoleStatusMap(state, user.roleStatusMap || home.roleStatusMap)
    this.mergeRoleStatusList(state, user.roles || home.roles)
    this.mergeRoleStatusList(state, home.roleApplications || home.roleApplicationList || home.applications, true)

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

    roles.forEach((item) => {
      if (typeof item === 'string') {
        const roleType = this.normalizeRoleType(item)

        if (roleType && !isApplicationList && state[roleType] === 'none') {
          state[roleType] = 'approved'
        }
        return
      }

      const roleType = this.normalizeRoleType(item.roleType || item.role_type || item.type || item.key || item.name)
      const status = this.normalizeRoleStatus(
        item.status || item.applyStatus || item.applicationStatus || item.roleStatus || item.role_status
      )

      if (roleType && status !== 'none') {
        state[roleType] = status
      }
    })
  },

  formatRoleApplicationList(home = {}) {
    const source = home.roleApplications || home.roleApplicationList || home.applications

    if (!Array.isArray(source)) {
      return []
    }

    return source
      .map((item) => ({
        roleType: this.normalizeRoleType(item.roleType || item.role_type || item.type || item.key || item.name),
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
      }))
      .filter((item) => item.roleType)
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
      available: 'none'
    }

    return statusMap[value] || value
  },

  isRolePending(status) {
    return ROLE_PENDING_STATUSES.includes(this.normalizeRoleStatus(status))
  },

  resolveSelectedRole(activeRole, roleStatusState = this.data.roleStatusState, applications = []) {
    if (this.isRolePending(roleStatusState[activeRole])) {
      return activeRole
    }

    const pendingApplication = applications.find((item) => this.isRolePending(item.status))

    if (pendingApplication && pendingApplication.roleType) {
      return pendingApplication.roleType
    }

    const pendingRole = ROLE_TAGS.find((item) => this.isRolePending(roleStatusState[item.key]))

    return pendingRole ? pendingRole.key : activeRole
  },

  normalizeRoleType(roleType) {
    return ROLE_TYPE_MAP[roleType] || this.data.roleType || 'player'
  },

  formatRoleTags(activeRole, roleStatusState = this.data.roleStatusState) {
    return ROLE_TAGS.map((item) => ({
      ...item,
      name: this.isRolePending(roleStatusState[item.key]) ? '审核中' : item.name,
      status: this.normalizeRoleStatus(roleStatusState[item.key]),
      active: item.key === activeRole
    }))
  },

  formatRolePermissionPrompt(roleType, roleStatusState = this.data.roleStatusState) {
    if (this.isRolePending(roleStatusState[roleType])) {
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
    return this.formatRolePermissionPromptWithPrompts(roleType, roleStatusState, prompts)
  },

  formatRolePermissionPromptWithPrompts(roleType, roleStatusState = this.data.roleStatusState, prompts = ROLE_PERMISSION_PROMPTS) {
    if (this.isRolePending(roleStatusState[roleType])) {
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

  formatRoleAuditPrompt(roleType, roleStatusState = this.data.roleStatusState, applications = this.data.roleApplicationList, user = {}) {
    const status = roleStatusState[roleType]
    const template = ROLE_AUDIT_PROMPTS[roleType]

    if (!template || !this.isRolePending(status)) {
      return createEmptyRoleAuditPrompt()
    }

    const application = applications.find((item) => item.roleType === roleType) || {}

    return {
      ...template,
      visible: true,
      submittedText: this.formatRoleAuditSubmittedText(application, user, template.roleName),
      done: this.normalizeRoleStatus(status) === 'approved',
      steps: ROLE_AUDIT_STEPS
    }
  },

  formatRoleAuditSubmittedText(application = {}) {
    const submittedAt = this.formatCompactDateTime(application.submittedAt)
    const expectedReviewAt = this.formatCompactDateTime(application.expectedReviewAt)

    if (!submittedAt && !expectedReviewAt) {
      return '身份申请已提交，请耐心等待审核'
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

  formatHeroDate(dateLabel, subtitle) {
    return [dateLabel, subtitle]
      .filter((item) => typeof item === 'string' && item.trim())
      .map((item) => item.trim())
      .join(' | ')
  },

  handleShellNavTap(event) {
    const key = event.detail && event.detail.key

    if (key === 'up' || key === 'down') {
      if (!this.suppressNextNavTap) {
        this.scrollPlayerHome(key, HOME_SCROLL_TAP_STEP_RPX)
      }
      return
    }

    if (key === 'left' || key === 'right') {
      navigateShellKey(key, {
        currentRoute: ROUTES.playerHome
      })
      return
    }

    if (navigateShellKey(key, {
      currentRoute: ROUTES.playerHome,
      onSameRoute: () => this.scrollPlayerHomeToTop()
    })) {
      return
    }

    if (key === 'home') {
      this.scrollPlayerHomeToTop()
      return
    }

    this.handleActionTap()
  },

  handleShellNavLongPress(event) {
    const key = event.detail && event.detail.key

    if (key !== 'up' && key !== 'down') {
      return
    }

    this.suppressNextNavTap = true
    this.stopHomeScrollHold(false)
    this.scrollPlayerHome(key, HOME_SCROLL_HOLD_STEP_RPX)

    this.homeScrollHoldTimer = setInterval(() => {
      this.scrollPlayerHome(key, HOME_SCROLL_HOLD_STEP_RPX)
    }, HOME_SCROLL_HOLD_INTERVAL_MS)
  },

  handleShellNavTouchEnd() {
    this.stopHomeScrollHold(true)
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

  handleHomeScroll(event) {
    const scrollTop = event.detail && event.detail.scrollTop

    if (typeof scrollTop === 'number') {
      this.homeScrollTopValue = scrollTop
    }
  },

  scrollPlayerHome(direction, stepRpx = HOME_SCROLL_TAP_STEP_RPX) {
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

  scrollPlayerHomeToTop() {
    this.homeScrollTopValue = 0
    this.setData({
      homeScrollTop: 0
    })
  },

  rpxToPx(value) {
    if (!wx.getSystemInfoSync) {
      return value / 2
    }

    const system = wx.getSystemInfoSync()
    const windowWidth = system && system.windowWidth ? system.windowWidth : 375

    return Math.round((value * windowWidth) / 750)
  },

  handleNearbyTabTap(event) {
    const key = event.currentTarget.dataset.key || 'all'

    this.setData({
      nearbyFilter: key,
      nearbyTabs: this.data.nearbyTabs.map((item) => ({
        ...item,
        active: item.key === key
      }))
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

  handleRoleTagTap(event) {
    const key = this.normalizeRoleType(event.currentTarget.dataset.key || 'player')

    this.setData({
      selectedRoleTag: key,
      roleTags: this.formatRoleTags(key, this.data.roleStatusState),
      rolePermissionPrompt: this.formatRolePermissionPrompt(key, this.data.roleStatusState),
      roleAuditPrompt: this.formatRoleAuditPrompt(key, this.data.roleStatusState, this.data.roleApplicationList)
    })
  },

  handleRoleApplyTap() {
    const prompt = this.data.rolePermissionPrompt || {}
    const roleType = prompt.roleType || this.data.selectedRoleTag
    const normalizedRoleType = this.normalizeRoleType(roleType)
    let url = `/${ROUTES.roleApply}?roleType=${roleType}`

    if (normalizedRoleType === 'expert') {
      url = `/${ROUTES.home}?ui=1&mode=expertApplyOverview&single=1&returnTo=${encodeURIComponent(ROUTES.playerHome)}`
    } else if (normalizedRoleType === 'guide') {
      url = `/${ROUTES.homeOther}?page=guideApply&single=1&roleType=guide`
    }

    navigateShellRoute(url, {
      currentRoute: ROUTES.playerHome
    })
  },

  handleRoleCompareTap() {
    navigateShellRoute(`${ROUTES.home}?ui=1&mode=roleComparison&single=1&returnTo=${encodeURIComponent(ROUTES.playerHome)}`, {
      currentRoute: ROUTES.playerHome
    })
  },

  handleRoleAuditDetailTap() {
    const prompt = this.data.roleAuditPrompt || {}
    const roleType = prompt.roleType || this.data.selectedRoleTag || 'guide'

    navigateShellRoute(`${ROUTES.homeOther}?page=pendingCards&roleType=${roleType}`, {
      currentRoute: ROUTES.playerHome
    })
  },

  handleActionTap(event) {
    const detail = event && event.detail ? event.detail : {}
    const item = detail.item || detail
    const currentDataset = event && event.currentTarget && event.currentTarget.dataset
      ? event.currentTarget.dataset
      : {}
    const targetDataset = event && event.target && event.target.dataset ? event.target.dataset : {}
    const route = currentDataset.route || targetDataset.route || detail.route || (item && item.route)

    if (route) {
      navigateShellRoute(route, {
        currentRoute: ROUTES.playerHome
      })
      return
    }

    toast.info('请选择可用入口')
  }
})

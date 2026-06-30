const homeService = require('../../../services/home')
const { ROUTES } = require('../../../config/routes')

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
    onlineText: '',
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
    nearbyGames: [],
    rankingSection: {
      icon: '🏆',
      title: '本周玩霸榜',
      moreText: '查看全部榜单'
    },
    rankingActiveRole: 'player',
    rankingTabs: [
      { key: 'player', name: '玩家', active: true },
      { key: 'expert', name: '行家', active: false },
      { key: 'guide', name: '领路人', active: false }
    ],
    rankingBoards: {},
    rankingList: [],
    myRank: {},
    showMyRank: false,
    achievementSection: {
      icon: '💎',
      title: '我的成就'
    },
    achievements: [],
    playerCard: {
      role: '',
      title: '',
      name: '',
      xp: '',
      progress: 0,
      next: '',
      stats: []
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
    roleStatusState: DEFAULT_ROLE_STATUS_STATE,
    roleApplicationList: [],
    roleAuditPrompt: createEmptyRoleAuditPrompt(),
    onlineCard: {
      title: '地球online',
      desc: '探索城市副本 · 解锁地图成就',
      tags: []
    },
    entries: [
      { title: '发起组局', desc: '创建你的带局房间', icon: '📍', theme: 'pink', route: ROUTES.gameCreate },
      { title: '局前大厅', desc: '准备就绪加入一局', icon: '', routeIcon: true, theme: 'cyan', route: ROUTES.gameHall }
    ],
    friendSection: {
      icon: '🎲',
      title: '朋友在玩',
      count: 0,
      moreText: '查看全部'
    },
    friendGames: [],
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
    this.loadPlayerHome()
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
      const nearbyGames = this.formatGameCards(home.nearbyGames || home.recommendedGames)
      const friendGames = this.formatGameCards(home.friendGames)

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
        onlineCard: this.formatOnlineCard(nearbySummary),
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
        showMyRank: rankingState.showMyRank,
        achievementSection: achievementState.section,
        achievements: achievementState.list,
        metaverse: this.formatMetaverseEntry(home.metaverseEntry || home.metaverse || {})
      })
    } catch (error) {
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
      value: item.value || (fallbackStats[index] && fallbackStats[index].value) || '',
      label: item.label
    }))
  },

  formatOnlineCard(nearbySummary) {
    const current = this.data.onlineCard
    const nearbyGameCount = this.pickFirstValue(
      nearbySummary.nearbyGameCount,
      nearbySummary.count,
      nearbySummary.gameCount
    )
    const checkedInCount = this.pickFirstValue(
      nearbySummary.checkedInCount,
      nearbySummary.checkinCount,
      nearbySummary.visitedCount
    )

    const tags = []
    const nearbyGameText = this.formatStatValue(nearbyGameCount)
    const checkedInText = this.formatStatValue(checkedInCount)

    if (nearbyGameText) {
      tags.push(`附近 ${nearbyGameText} 个组局`)
    }

    if (checkedInText) {
      tags.push(`已打卡 ${checkedInText} 处`)
    }

    return {
      ...current,
      tags
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
      Array.isArray(friendGames) ? friendGames.length : 0
    )

    return {
      ...current,
      icon: section.icon || current.icon,
      title: section.title || section.name || current.title,
      count,
      moreText: section.moreText || section.moreLabel || section.actionText || current.moreText
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
      list: this.formatAchievements(list)
    }
  },

  formatAchievements(list) {
    if (!Array.isArray(list) || list.length === 0) {
      return []
    }

    return list
      .map((item) => this.formatAchievement(item))
      .sort((prev, next) => Number(prev.unlocked === false) - Number(next.unlocked === false))
  },

  formatAchievement(item) {
    const title = item.title || item.name || item.displayName || ''
    const status = item.statusText || item.statusLabel || item.status || this.formatAchievementStatus(item) || ''
    const unlocked = item.unlocked != null ? Boolean(item.unlocked) : item.locked !== true

    const progressPercent = this.normalizeProgress(
      this.pickFirstValue(item.progressPercent, item.progress),
      0
    )

    return {
      id: item.id || item.code || title,
      icon: item.icon || item.iconText || '🏆',
      title,
      status: status ? (status.startsWith('▲') ? status : `▲ ${status}`) : '',
      tone: this.resolveAchievementTone(item, unlocked),
      unlocked,
      progressPercent,
      showProgress: progressPercent > 0 && unlocked
    }
  },

  resolveAchievementTone(item, unlocked = true) {
    if (!unlocked) {
      return 'locked'
    }

    const text = `${item.code || ''} ${item.title || item.name || ''}`

    if (item.progressPercent != null || item.progress != null || /earth|地球/.test(text)) {
      return 'blue'
    }

    return item.tone || 'gold'
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

  formatMetaverseEntry(entry = {}) {
    const current = this.data.metaverse
    const tags = Array.isArray(entry.tags) && entry.tags.length > 0
      ? entry.tags
      : this.splitMetaverseTags(entry.desc || entry.subtitle) || current.tags

    return {
      ...current,
      title: entry.actionText || entry.title || current.title,
      desc: entry.summary || entry.description || entry.desc || current.desc,
      tags,
      avatars: this.formatMetaverseAvatars(entry.avatars || entry.users || current.avatars),
      badge: this.formatMetaverseBadge(entry, current.badge),
      route: entry.route || current.route
    }
  },

  formatMetaverseBadge(entry = {}, fallback = '') {
    const joinedCount = this.pickFirstValue(entry.joinedCount, entry.onlineCount, entry.participantCount)

    if (joinedCount != null && joinedCount !== '') {
      return `+${joinedCount}`
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
    const boards = this.formatRankingBoards(home)
    const activeKey = boards[requestedActiveKey] ? requestedActiveKey : (tabs[0] && tabs[0].key) || 'player'
    const activeBoard = boards[activeKey] || {}
    const display = this.formatRankingDisplay(
      activeBoard.list,
      activeBoard.myRank
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
      moreText: section.moreText || section.moreLabel || section.actionText || this.data.rankingSection.moreText
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

  formatRankingBoards(home = {}) {
    const sourceBoards = home.rankingBoards || home.rankings || home.rankingByRole
    const boards = {}

    if (sourceBoards && typeof sourceBoards === 'object' && !Array.isArray(sourceBoards)) {
      Object.keys(sourceBoards).forEach((key) => {
        const roleKey = this.normalizeRoleType(key)
        boards[roleKey] = this.formatRankingBoard(sourceBoards[key])
      })
    }

    if (Array.isArray(sourceBoards)) {
      sourceBoards.forEach((board, index) => {
        const roleKey = this.normalizeRoleType(
          board.key || board.value || board.roleType || board.name || (ROLE_TAGS[index] && ROLE_TAGS[index].key)
        )
        boards[roleKey] = this.formatRankingBoard(board)
      })
    }

    if (Array.isArray(home.rankingList) && home.rankingList.length > 0 && !boards.player) {
      boards.player = this.formatRankingBoard(
        {
          list: home.rankingList,
          myRank: home.myRank || home.currentUserRank
        }
      )
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

    return list.map((item) => this.formatRankingItem(item)).filter((item) => item.name || item.rank)
  },

  formatRankingDisplay(list = [], myRank = {}) {
    const rankingList = Array.isArray(list) ? list : []
    const currentRank = myRank || {}
    const hasCurrentRank = Boolean(currentRank.rank || currentRank.name || currentRank.avatarUrl || currentRank.avatarFallback)

    if (!this.shouldInlineMyRank(rankingList, currentRank)) {
      return {
        list: rankingList,
        myRank: currentRank,
        showMyRank: hasCurrentRank
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

  formatRankingItem(item) {
    if (!item || Object.keys(item).length === 0) {
      return {}
    }

    const avatarCandidate = item.avatarUrl || item.avatarSrc || item.avatarImage || item.avatar
    const avatarUrl = this.isImagePath(avatarCandidate) ? avatarCandidate : ''
    const name = item.name || item.nickname || item.displayName || ''

    return {
      id: item.id,
      rank: this.formatRankNo(this.pickFirstValue(item.rank, item.rankNo, item.position)),
      avatarUrl,
      avatarFallback: item.avatarFallback || item.initials || (!avatarUrl && avatarCandidate) || this.makeAvatarFallback(name),
      name,
      desc: item.desc || item.description || item.summary || item.rankText || this.formatRankingDesc(item) || '',
      xp: this.formatRankingXp(this.pickFirstValue(item.xp, item.xpText, item.experience, item.weeklyXp)),
      xpUnit: item.xpUnit || 'XP'
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

  formatGameCards(games) {
    if (!Array.isArray(games) || games.length === 0) {
      return []
    }

    return games.map((item) => ({
      id: item.id,
      scope: item.scope || item.distanceScope || 'city',
      title: item.title,
      tag: item.typeText || item.statusText || item.tag || '',
      type: item.type || item.statusType || '',
      coverSrc: item.coverUrl || item.coverSrc || item.cover,
      avatarUrls: item.participantAvatars || item.avatarUrls || item.participantAvatarUrls || [],
      price: item.priceText || item.price || '',
      action: item.actionText || item.action || '',
      location: item.locationText || this.formatGameLocation(item),
      time: item.timeText || item.time || '',
      joinedText: item.joinedText || this.formatJoinedText(item.joinedCount),
      actions: this.formatGameActions(item.actions)
    }))
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
      return []
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
      wx.showToast({
        title: '功能正在开发中',
        icon: 'none'
      })
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

    const display = this.formatRankingDisplay(board.list, board.myRank)

    patch.rankingList = display.list
    patch.myRank = display.myRank
    patch.showMyRank = display.showMyRank

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
      wx.navigateTo({
        url: `/${ROUTES.home}?ui=1&mode=roleComparison&single=1&returnTo=${encodeURIComponent(ROUTES.playerHome)}`
      })
      return
    }

    wx.showToast({
      title: '权益对比正在开发中',
      icon: 'none'
    })
  },

  handleRoleAuditDetailTap() {
    const prompt = this.data.roleAuditPrompt || {}
    const roleType = prompt.roleType || this.data.selectedRoleTag || 'guide'

    wx.navigateTo({
      url: `/${ROUTES.homeOther}?page=pendingCards&roleType=${roleType}`
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
})

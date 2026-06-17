const homeService = require('../../../services/home')

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

Page({
  data: {
    onlineText: '3999人在线',
    roleType: 'player',
    toolbarStyle: '',
    canvasStyle: '',
    contentStyle: '',
    dockStyle: '',
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
      moreText: '查看全部榜单'
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
          { rank: '01', avatarUrl: '/pages/home/player/assets/ranking-avatar-01.png', avatarFallback: 'PRO', name: '领域专家 PRO', desc: '本周组局 12 · MVP 5次', xp: '2,450', xpUnit: 'XP' },
          { rank: '02', avatarUrl: '/pages/home/player/assets/ranking-avatar-02.png', avatarFallback: '星', name: '社交达人', desc: '本周组局 8 次', xp: '1,890', xpUnit: 'XP' },
          { rank: '03', avatarUrl: '/pages/home/player/assets/ranking-avatar-03.png', avatarFallback: '探', name: '探险家', desc: '本周组局 6 次', xp: '1,560', xpUnit: 'XP' }
        ],
        myRank: {
          rank: '52',
          avatarUrl: '/pages/home/player/assets/ranking-avatar-me.png',
          avatarFallback: 'A',
          name: '我（Alex）',
          desc: '上周排名 65 ↑',
          xp: '520',
          xpUnit: 'XP'
        }
      }
    },
    rankingList: [
      { rank: '01', avatarUrl: '/pages/home/player/assets/ranking-avatar-01.png', avatarFallback: 'PRO', name: '领域专家 PRO', desc: '本周组局 12 · MVP 5次', xp: '2,450', xpUnit: 'XP' },
      { rank: '02', avatarUrl: '/pages/home/player/assets/ranking-avatar-02.png', avatarFallback: '星', name: '社交达人', desc: '本周组局 8 次', xp: '1,890', xpUnit: 'XP' },
      { rank: '03', avatarUrl: '/pages/home/player/assets/ranking-avatar-03.png', avatarFallback: '探', name: '探险家', desc: '本周组局 6 次', xp: '1,560', xpUnit: 'XP' }
    ],
    myRank: {
      rank: '52',
      avatarUrl: '/pages/home/player/assets/ranking-avatar-me.png',
      avatarFallback: 'A',
      name: '我（Alex）',
      desc: '上周排名 65 ↑',
      xp: '520',
      xpUnit: 'XP'
    },
    achievements: [
      { icon: '🏆', title: '百场王者', status: '▲ 等级' },
      { icon: '🏆', title: '引航王者', status: '▲ 等级' },
      { icon: '🔒', title: '隐藏徽章', status: '▲ 未解锁' },
      { icon: '🌍', title: '地球漫游者', status: '▲ 进度20%' }
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
    onlineCard: {
      title: '地球online',
      desc: '探索城市副本 · 解锁地图成就',
      tags: ['附近 12 个组局', '已打卡 8 处']
    },
    entries: [
      { title: '发起组局', desc: '创建你的带局房间', icon: '📍', theme: 'pink' },
      { title: '局前大厅', desc: '准备就绪加入一局', icon: '', routeIcon: true, theme: 'cyan' }
    ],
    friendSection: {
      icon: '🎲',
      title: '朋友在玩',
      count: 2,
      moreText: '查看全部'
    },
    friendGames: [
      {
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
      avatars: ['A', 'L', 'M'],
      badge: '+99'
    }
  },

  onLoad() {
    this.alignToolbarToCapsule()
    this.loadPlayerHome()
  },

  async loadPlayerHome() {
    try {
      const home = await homeService.getHome()
      const hero = home && home.hero ? home.hero : {}
      const playerSummary = home && home.playerSummary ? home.playerSummary : {}
      const nearbySummary = home && home.nearbySummary ? home.nearbySummary : {}
      const nearbySection = home && home.nearbySection ? home.nearbySection : {}
      const friendSection = home && home.friendSection ? home.friendSection : {}
      const rankingState = this.formatRankingState(home || {})
      const roleName = hero.roleName || '玩家'
      const date = this.formatHeroDate(hero.dateLabel, hero.subtitle) || this.data.hero.date
      const activeRole = this.normalizeRoleType(
        playerSummary.currentRole || playerSummary.roleType || hero.currentRole || hero.roleType || roleName
      )
      const nearbyGames = this.formatGameCards(home.nearbyGames || home.recommendedGames, this.data.nearbyGames)
      const friendGames = this.formatGameCards(home.friendGames, this.data.friendGames)

      this.setData({
        onlineText: hero.onlineText || this.data.onlineText,
        roleType: activeRole,
        hero: {
          ...this.data.hero,
          title: `HELLO, ${roleName}!`,
          date
        },
        playerCard: this.formatPlayerCard(playerSummary),
        roleTags: this.formatRoleTags(activeRole),
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
        myRank: rankingState.myRank
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

    return {
      ...current,
      tags: [
        `附近 ${this.formatStatValue(nearbyGameCount) || '12'} 个组局`,
        `已打卡 ${this.formatStatValue(checkedInCount) || '8'} 处`
      ]
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
      moreText: section.moreText || section.moreLabel || section.actionText || current.moreText
    }
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

    return {
      section: this.formatRankingSection(section),
      activeKey,
      tabs: tabs.map((item) => ({
        ...item,
        active: item.key === activeKey
      })),
      boards,
      list: activeBoard.list || this.data.rankingList,
      myRank: activeBoard.myRank || this.data.myRank
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
      return fallbackGames
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
      action: item.actionText || item.action || '加入',
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

  normalizeRoleType(roleType) {
    return ROLE_TYPE_MAP[roleType] || this.data.roleType || 'player'
  },

  formatRoleTags(activeRole) {
    return ROLE_TAGS.map((item) => ({
      ...item,
      active: item.key === activeRole
    }))
  },

  formatHeroDate(dateLabel, subtitle) {
    return [dateLabel, subtitle]
      .filter((item) => typeof item === 'string' && item.trim())
      .map((item) => item.trim())
      .join(' | ')
  },

  alignToolbarToCapsule() {
    if (!wx.getMenuButtonBoundingClientRect || !wx.getSystemInfoSync) {
      return
    }

    const menuButton = wx.getMenuButtonBoundingClientRect()
    const system = wx.getSystemInfoSync()
    const ratio = 750 / system.windowWidth
    const iconCenterOffset = 29
    const capsuleCenterTop = (menuButton.top + menuButton.height / 2) * ratio
    const toolbarTop = capsuleCenterTop - iconCenterOffset
    const toolbarRight = (system.windowWidth - menuButton.left + 10) * ratio

    this.setData({
      toolbarStyle: `top: ${toolbarTop}rpx; right: ${toolbarRight}rpx;`
    })
  },

  handleShellNavTap() {},

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
      patch.rankingList = board.list || this.data.rankingList
      patch.myRank = board.myRank || this.data.myRank
    }

    this.setData(patch)
  },

  handleActionTap() {
    wx.showToast({
      title: '功能正在开发中',
      icon: 'none'
    })
  }
})

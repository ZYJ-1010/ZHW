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
        price: '￥0/人',
        action: '加入',
        location: '📍静安区 · 3.2km · 5/8人',
        time: '⏰2026年5月1日 20:00--22:00',
        joinedText: '+5位玩家已入局',
        imageText: '河',
        imageTone: 'small',
        tagTone: 'explore',
        actions: ['分享', '关注', '引荐', '打招呼']
      },
      {
        title: 'AI赋能系统搭建交流局',
        scope: 'city',
        tag: '任务局',
        price: '￥0/人',
        action: '加入',
        location: '📍黄浦区 · 8.2km · 3/8人',
        time: '⏰2026年5月1日 14:00--16:00',
        joinedText: '+3位玩家已入局',
        imageText: 'AI',
        imageTone: '',
        tagTone: '',
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
        imageTone: 'warm',
        tagTone: 'explore',
        actions: ['分享', '关注', '引荐', '打招呼']
      }
    ],
    metaverse: {
      title: '进入元宇宙',
      desc: '共创数字街区｜全球联机互动',
      tags: ['3D空间', 'NFT徽章'],
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
      const roleName = hero.roleName || '玩家'
      const date = this.formatHeroDate(hero.dateLabel, hero.subtitle) || this.data.hero.date
      const activeRole = this.normalizeRoleType(
        playerSummary.currentRole || playerSummary.roleType || hero.currentRole || hero.roleType || roleName
      )

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
        onlineCard: this.formatOnlineCard(nearbySummary)
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

  handleActionTap() {
    wx.showToast({
      title: '功能正在开发中',
      icon: 'none'
    })
  }
})

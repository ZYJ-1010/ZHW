const homeService = require('../../../services/home')
const { toUserMessage } = require('../../../utils/user-message')

const DEFAULT_TABS = [
  { key: 'player', name: '玩家' },
  { key: 'expert', name: '行家' },
  { key: 'guide', name: '领路人' }
]

function formatRankNo(rank) {
  const value = Number(rank)
  if (Number.isInteger(value) && value > 0 && value < 10) {
    return `0${value}`
  }
  return rank ? String(rank) : ''
}

function avatarText(item = {}) {
  const fallback = item.avatarFallback || item.initials || ''
  if (fallback) {
    return fallback
  }
  const name = item.name || item.nickname || item.displayName || ''
  return String(name).trim().slice(0, 2).toUpperCase() || '榜'
}

function formatMember(item = {}) {
  const name = item.name || item.nickname || item.displayName || '用户'
  return {
    id: item.userId || item.id || item.memberId || `${item.rank || 'rank'}-${name}`,
    rank: formatRankNo(item.rank),
    avatarUrl: item.avatarUrl || item.avatarSrc || '',
    avatarText: avatarText(item),
    name,
    desc: item.desc || item.description || item.summary || '',
    score: item.xp || item.experience || item.score || 0,
    scoreUnit: item.xpUnit || item.scoreUnit || '分',
    isMe: item.isMe === true || item.isSelf === true || item.isCurrentUser === true
  }
}

Page({
  data: {
    loading: true,
    error: '',
    title: '玩霸榜',
    desc: '',
    tabs: DEFAULT_TABS,
    activeKey: 'player',
    boards: {},
    members: [],
    myRank: null
  },

  onLoad(options = {}) {
    this.setData({
      activeKey: options.type || 'player'
    })
    this.loadRanking()
  },

  async loadRanking() {
    try {
      const home = await homeService.getHome({ roleType: 'player' })
      const section = home.rankingSection || {}
      const tabs = Array.isArray(section.tabs) && section.tabs.length > 0 ? section.tabs : DEFAULT_TABS
      const boards = home.rankingBoards || {}
      const activeKey = this.data.activeKey || (tabs[0] && tabs[0].key) || 'player'

      this.setData({
        loading: false,
        error: '',
        title: section.title || '玩霸榜',
        desc: section.desc || section.description || '',
        tabs: tabs.map((item) => ({
          key: item.key || item.value || item.name,
          name: item.name || item.label || item.title || item.key
        })),
        boards,
        activeKey
      })
      this.applyBoard(activeKey)
    } catch (error) {
      this.setData({
        loading: false,
        error: toUserMessage(error && error.message, '榜单加载失败')
      })
    }
  },

  applyBoard(key) {
    const board = this.data.boards[key] || {}
    const list = Array.isArray(board.list) ? board.list : []
    const myRank = board.myRank && Object.keys(board.myRank).length > 0 ? formatMember(board.myRank) : null

    this.setData({
      activeKey: key,
      members: list.map(formatMember),
      myRank
    })
  },

  onTabTap(event) {
    const key = event.currentTarget.dataset.key
    if (!key || key === this.data.activeKey) {
      return
    }
    this.applyBoard(key)
  },

  onRetry() {
    this.setData({ loading: true, error: '' })
    this.loadRanking()
  }
})

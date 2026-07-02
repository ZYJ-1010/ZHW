const { ROUTES } = require('../../../config/routes')
const mapService = require('../../../services/map')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')

const NAV_ITEMS = [
  { name: '我的', key: 'mine' },
  { name: '元宇宙', key: 'metaverse' },
  { name: '地图', key: 'map' },
  { name: '消息', key: 'message' },
  { name: '首页', key: 'home' }
]

function decorateStory(story, tagEmojis, badgeEmojis) {
  return {
    ...story,
    tagEmoji: tagEmojis[story.tag] || '',
    highlightEmoji: tagEmojis[story.highlightTag] || '',
    badgeEmoji: badgeEmojis[story.badgeTitle] || ''
  }
}

function sortStories(stories, sortMode) {
  return [...stories].sort((a, b) => {
    const left = new Date(a.sortDate || a.date).getTime()
    const right = new Date(b.sortDate || b.date).getTime()

    return sortMode === 'asc' ? left - right : right - left
  })
}

function getActionStyle(actionText) {
  if (!actionText) {
    return ''
  }

  const width = Math.max(106, Math.min(166, actionText.length * 22 + 48))

  return `width: ${width}rpx;`
}

function buildStoryGroups(groups, sortMode, tagEmojis, badgeEmojis) {
  const pinnedStories = []
  const sortedGroups = (groups || []).reduce((result, group) => {
    const normalStories = []

    const stories = group.stories || []

    stories.forEach((story) => {
      const decoratedStory = decorateStory(story, tagEmojis || {}, badgeEmojis || {})
      decoratedStory.actionStyle = getActionStyle(story.actionText)

      if (story.type === 'first') {
        pinnedStories.push(decoratedStory)
        return
      }

      normalStories.push(decoratedStory)
    })

    if (normalStories.length) {
      result.push({
        year: group.year,
        stories: sortStories(normalStories, sortMode)
      })
    }

    return result
  }, [])

  const orderedGroups = [...sortedGroups].sort((a, b) => (
    sortMode === 'asc' ? a.year - b.year : b.year - a.year
  ))
  const lastGroup = orderedGroups[orderedGroups.length - 1]

  if (lastGroup && pinnedStories.length) {
    lastGroup.stories = lastGroup.stories.concat(pinnedStories)
  }

  return orderedGroups
}

Page({
  data: {
    onlineText: '',
    navItems: NAV_ITEMS,
    showMapOverlay: false,
    pageTitle: '',
    profileName: '',
    routeTip: '',
    switchMapText: '',
    journeyTitle: '',
    journeyDesc: '',
    participantLabel: '',
    endingTitle: '',
    endingDesc: '',
    storySortMode: 'desc',
    summaryStats: [],
    mapLegends: [],
    mapStats: [],
    tagEmojis: {},
    badgeEmojis: {},
    rawStoryGroups: [],
    storyGroups: []
  },

  onLoad() {
    this.loadMyCity()
  },

  async loadMyCity() {
    try {
      const config = await mapService.getMyCity()
      const storyGroups = config.storyGroups || []
      const tagEmojis = config.tagEmojis || {}
      const badgeEmojis = config.badgeEmojis || {}

      this.setData({
        onlineText: config.onlineText || '',
        pageTitle: config.pageTitle || '',
        profileName: config.profileName || '',
        routeTip: config.routeTip || '',
        switchMapText: config.switchMapText || '',
        journeyTitle: config.journeyTitle || '',
        journeyDesc: config.journeyDesc || '',
        participantLabel: config.participantLabel || '',
        endingTitle: config.endingTitle || '',
        endingDesc: config.endingDesc || '',
        summaryStats: config.summaryStats || [],
        mapLegends: config.mapLegends || [],
        mapStats: config.mapStats || [],
        tagEmojis,
        badgeEmojis,
        rawStoryGroups: storyGroups,
        storyGroups: buildStoryGroups(storyGroups, this.data.storySortMode, tagEmojis, badgeEmojis)
      })
    } catch (error) {
      wx.showToast({
        title: error.message || '城市故事暂时不可用',
        icon: 'none'
      })
    }
  },

  handleBackTap() {
    wx.navigateBack({
      delta: 1,
      fail: () => {
        navigateShellRoute(ROUTES.map, {
          currentRoute: ROUTES.mapMyCity
        })
      }
    })
  },

  handleSwitchMapTap() {
    this.setData({
      showMapOverlay: !this.data.showMapOverlay
    })
  },

  handleFilterTap() {
    const nextSortMode = this.data.storySortMode === 'desc' ? 'asc' : 'desc'

    this.setData({
      storySortMode: nextSortMode,
      storyGroups: buildStoryGroups(
        this.data.rawStoryGroups,
        nextSortMode,
        this.data.tagEmojis,
        this.data.badgeEmojis
      )
    })
  },

  handleStoryTap(event) {
    const story = this.findStory(event.currentTarget.dataset.id)

    wx.showToast({
      title: story ? story.title : '暂无故事详情',
      icon: 'none'
    })
  },

  handleStoryActionTap(event) {
    const story = this.findStory(event.currentTarget.dataset.id)

    navigateShellRoute(`${ROUTES.map}?story=${encodeURIComponent(story ? story.id : '')}`, {
      currentRoute: ROUTES.mapMyCity
    })
  },

  findStory(id) {
    for (const group of this.data.rawStoryGroups) {
      const story = (group.stories || []).find((item) => item.id === id)

      if (story) {
        return story
      }
    }

    return null
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}
    navigateShellKey(key, {
      currentRoute: ROUTES.mapMyCity
    })
  }
})

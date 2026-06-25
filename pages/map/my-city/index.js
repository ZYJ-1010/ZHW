const { ROUTES } = require('../../../config/routes')

const NAV_ITEMS = [
  { name: '我的', key: 'mine' },
  { name: '元宇宙', key: 'metaverse' },
  { name: '地图', key: 'map' },
  { name: '消息', key: 'message' },
  { name: '首页', key: 'home' }
]

const SUMMARY_STATS = [
  { value: '28', label: '故事总数', tone: 'blue' },
  { value: '6', label: '覆盖城市', tone: 'violet' },
  { value: '52', label: '参与组局数', tone: 'pink' }
]

const MAP_LEGENDS = [
  { label: '第一次', tone: 'pink' },
  { label: '夜游', tone: 'violet' },
  { label: '社交局', tone: 'blue' },
  { label: '最难忘', tone: 'gold' }
]

const MAP_STATS = [
  { label: '里程', value: '40244', unit: '公里' },
  { label: '次数', value: '33', unit: '次' },
  { label: '国家/地区', value: '1', unit: '个' },
  { label: '城市', value: '10', unit: '个' }
]

const TAG_EMOJIS = {
  最难忘: '👑',
  桌游局: '🎲',
  微醺局: '🍷',
  脑暴局: '💡',
  篮球局: '🏀',
  第一次: '🌱',
  起点: '🌱',
  摄影局: '📷'
}

const BADGE_EMOJIS = {
  创业伙伴: '🤝',
  深度密友: '💬',
  合伙人: '🤝',
  室友: '🏠',
  固定局友: '📌'
}

const STORY_GROUPS = [
  {
    year: 2024,
    stories: [
      {
        id: 'countdown-night',
        tag: '最难忘',
        tagTone: 'gold',
        title: '跨年夜的倒计时',
        date: '12.31',
        sortDate: '2024-12-31',
        location: '',
        cover: '/pages/map/my-city/assets/story-river-cover.png',
        coverLocation: '上海 · 外滩',
        desc: '和刚认识的摄影局朋友们一起在外滩等待新年钟声。江风吹得发抖，但倒数的那一刻，所有的陌生人都变成了朋友。',
        participants: ['a', 'b', 'c'],
        badgeTitle: '',
        badgeDesc: '',
        actionText: '',
        actionTone: ''
      },
      {
        id: 'script-rain',
        tag: '桌游局',
        subTag: '新手场',
        tagTone: 'violet',
        title: '暴雨中的剧本杀',
        date: '2024.09.20',
        sortDate: '2024-09-20',
        location: '杭州 · 西湖区 · 14:00-22:00',
        desc: '原定5人的局因为暴雨只来了3人，却因此有了最深入的交谈。认识了做AI的@阿杰，现在我们是创业合伙人。',
        participants: ['a', 'b', 'c', 'd'],
        badgeTitle: '创业伙伴',
        badgeDesc: '已共同发起 3 个项目',
        actionText: '查看项目',
        actionTone: 'violet'
      },
      {
        id: 'truth-night',
        tag: '微醺局',
        subTag: '深夜场',
        tagTone: 'orange',
        title: '周五晚上的坦白局',
        date: '2024.11.03',
        sortDate: '2024-11-03',
        location: '北京 · 三里屯 · 21:00',
        desc: '"你最后悔的事是什么？"那个问题让陌生人变成了知己。和@Lucy约定每月一次深度对话。',
        participants: ['a', 'b'],
        badgeTitle: '深度密友',
        badgeDesc: '',
        actionText: '约下次',
        actionTone: 'orange'
      },
      {
        id: 'business-canvas',
        tag: '脑暴局',
        subTag: '创始人专场',
        tagTone: 'gold',
        title: '凌晨的商业模式画布',
        date: '2024.12.15',
        sortDate: '2024-12-15',
        location: '深圳 · 科技园 · 通宵',
        desc: '从晚上8点到早上6点，8个人在黑板上画满了想法。这个局让我找到了技术合伙人@老K。',
        participants: ['a', 'b', 'c', 'd'],
        badgeTitle: '合伙人',
        badgeDesc: '公司估值 500w',
        actionText: '查看公司',
        actionTone: 'gold'
      },
      {
        id: 'court-weekly',
        tag: '篮球局',
        subTag: '每周固定',
        tagTone: 'cyan',
        title: '东华球场的汗水',
        date: '每周六',
        sortDate: '2024-06-15',
        location: '上海 · 东华大学 · 16:00',
        desc: '最纯粹的快乐。这里没有身份，只有队友。通过球局认识了现在的室友@阿强。',
        participants: ['a', 'b', 'c'],
        badgeTitle: '室友',
        badgeDesc: '合租 6 个月',
        actionText: '加入球局',
        actionTone: 'cyan'
      },
      {
        id: 'first-use',
        type: 'first',
        tag: '起点',
        subTag: '',
        tagTone: 'green',
        highlightTag: '第一次',
        title: '第一次使用真好玩',
        date: '03.12',
        sortDate: '2024-03-12',
        location: '广州 · 天河公园',
        desc: '抱着试试看的心态参加了第一次飞盘局，从此打开了城市探索的新世界。',
        participants: [],
        badgeTitle: '',
        badgeDesc: '',
        actionText: '',
        actionTone: ''
      }
    ]
  },
  {
    year: 2023,
    stories: [
      {
        id: 'sunrise-photo',
        tag: '摄影局',
        subTag: '第1次参与',
        tagTone: 'pink',
        title: '外滩 sunrise 拍摄',
        date: '2023.03.12',
        sortDate: '2023-03-12',
        location: '上海 · 外滩观景台 · 06:00',
        desc: '为了拍日出早上5点起床，认识了同样疯狂的@小林和@大为。后来我们组成了固定摄影小队，每周六早扫街。',
        participants: ['a', 'b', 'c'],
        badgeTitle: '固定局友',
        badgeDesc: '已持续组队 8 个月',
        actionText: '再组一局',
        actionTone: 'blue'
      }
    ]
  }
]

function decorateStory(story) {
  return {
    ...story,
    tagEmoji: TAG_EMOJIS[story.tag] || '',
    highlightEmoji: TAG_EMOJIS[story.highlightTag] || '',
    badgeEmoji: BADGE_EMOJIS[story.badgeTitle] || ''
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

function buildStoryGroups(sortMode) {
  const pinnedStories = []
  const sortedGroups = STORY_GROUPS.reduce((groups, group) => {
    const normalStories = []

    group.stories.forEach((story) => {
      const decoratedStory = decorateStory(story)
      decoratedStory.actionStyle = getActionStyle(story.actionText)

      if (story.type === 'first') {
        pinnedStories.push(decoratedStory)
        return
      }

      normalStories.push(decoratedStory)
    })

    if (normalStories.length) {
      groups.push({
        year: group.year,
        stories: sortStories(normalStories, sortMode)
      })
    }

    return groups
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
    onlineText: '3999人在线',
    navItems: NAV_ITEMS,
    showMapOverlay: false,
    profileName: '我的信息',
    storySortMode: 'desc',
    summaryStats: SUMMARY_STATS,
    mapLegends: MAP_LEGENDS,
    mapStats: MAP_STATS,
    storyGroups: buildStoryGroups('desc')
  },

  handleBackTap() {
    wx.navigateBack({
      delta: 1,
      fail: () => {
        wx.navigateTo({
          url: `/${ROUTES.map}`
        })
      }
    })
  },

  handleSwitchMapTap() {
    wx.showToast({
      title: '轨迹类型切换待接入',
      icon: 'none'
    })
  },

  handleFilterTap() {
    const nextSortMode = this.data.storySortMode === 'desc' ? 'asc' : 'desc'

    this.setData({
      storySortMode: nextSortMode,
      storyGroups: buildStoryGroups(nextSortMode)
    })
  },

  handleStoryTap(event) {
    const story = this.findStory(event.currentTarget.dataset.id)

    wx.showToast({
      title: story ? `${story.title}待接入` : '故事详情待接入',
      icon: 'none'
    })
  },

  handleStoryActionTap(event) {
    const story = this.findStory(event.currentTarget.dataset.id)

    wx.showToast({
      title: story ? `${story.actionText}待接入` : '故事动作待接入',
      icon: 'none'
    })
  },

  findStory(id) {
    for (const group of STORY_GROUPS) {
      const story = group.stories.find((item) => item.id === id)

      if (story) {
        return story
      }
    }

    return null
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}
    const routeMap = {
      home: ROUTES.playerHome,
      map: ROUTES.map,
      message: ROUTES.message,
      mine: ROUTES.profile,
      metaverse: ROUTES.metaverse
    }
    const route = routeMap[key]

    if (!route) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  }
})

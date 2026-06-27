const ASSET_BASE = '/pages/profile/system-management/service-case-detail/assets'

function buildStars(activeCount = 5) {
  return Array.from({ length: 5 }, (_, index) => ({
    key: `star-${index}`,
    active: index < activeCount
  }))
}

const CASE_DETAIL_MAP = {
  'case-stranger': {
    caseInfo: {
      title: '陌生人破冰局 · 氛围带动',
      date: '2026-05-25',
      playersText: '5人局',
      totalPlayers: 5,
      ratingText: '5.0分',
      iconText: '🎭',
      tone: 'blue'
    },
    rating: {
      score: '5.0',
      tags: ['氛围超棒', '零冷场', '快速破冰', '推荐']
    },
    detailSections: [
      {
        label: '服务亮点：',
        text: '通过"两真一假"游戏快速打破僵局，15分钟内让8位互不相识的玩家建立初步信任。'
      },
      {
        label: '个性化设计：',
        text: '根据玩家性格特点分组，内向玩家优先参与低压力环节，外向玩家带动全场节奏。'
      },
      {
        label: '玩家反馈：',
        text: '"像认识了很久的朋友"、"完全没想到能这么快融入"、"行家太会带气氛了"。'
      }
    ],
    players: [
      { name: '小明', desc: '首次参与 · 内向型', last: false },
      { name: '小红', desc: '老玩家 · 外向型', last: false },
      { name: '阿杰', desc: '首次参与 · 观察型', last: true }
    ]
  },
  'case-script': {
    caseInfo: {
      title: '沉浸式剧本杀局 · 细节控',
      date: '2026-06-04',
      playersText: '6人局',
      totalPlayers: 6,
      ratingText: '4.0分',
      iconText: '🔍',
      tone: 'purple'
    },
    rating: {
      score: '4.0',
      tags: ['准备充分', '节奏稳定', '道具细致', '安全感']
    },
    detailSections: [
      {
        label: '服务亮点：',
        text: '提前确认玩家偏好、角色接受度和到场时间，让不同经验的玩家都能自然进入剧情。'
      },
      {
        label: '个性化设计：',
        text: '根据角色强弱和玩家性格安排提示卡，并在关键节点控制节奏，减少等待和打断。'
      },
      {
        label: '玩家反馈：',
        text: '"细节很到位"、"每个人都有被照顾到"、"推进节奏比较舒服"。'
      }
    ],
    players: [
      { name: '林同学', desc: '新手玩家 · 代入型', last: false },
      { name: '阿哲', desc: '老玩家 · 推理型', last: false },
      { name: '小北', desc: '剧情偏好 · 观察型', last: true }
    ]
  },
  'case-board-game': {
    caseInfo: {
      title: '桌游竞技局 · 策略引导',
      date: '2026-05-18',
      playersText: '4人局',
      totalPlayers: 4,
      ratingText: '4.8分',
      iconText: '🎯',
      tone: 'orange'
    },
    rating: {
      score: '4.8',
      tags: ['策略清晰', '新手友好', '节奏紧凑', '复盘到位']
    },
    detailSections: [
      {
        label: '服务亮点：',
        text: '开局前快速确认玩家经验差异，并将复杂策略拆成阶段目标，降低新手理解成本。'
      },
      {
        label: '个性化设计：',
        text: '对新手玩家提供关键节点提示，对熟练玩家保留决策空间，让竞技感和参与感同时在线。'
      },
      {
        label: '玩家反馈：',
        text: '"第一次玩也能跟上"、"提示很克制不破坏竞技体验"、"结束复盘很有帮助"。'
      }
    ],
    players: [
      { name: '周舟', desc: '策略新手 · 学习型', last: false },
      { name: 'Allen', desc: '桌游老手 · 竞技型', last: false },
      { name: '可可', desc: '轻策略偏好 · 协作型', last: true }
    ]
  }
}

function decodeOption(value) {
  if (!value) {
    return ''
  }

  try {
    return decodeURIComponent(value)
  } catch (error) {
    return value
  }
}

function buildCaseDetail(caseId, iconOptions = {}) {
  const detail = CASE_DETAIL_MAP[caseId] || CASE_DETAIL_MAP['case-stranger']
  const score = Number(detail.rating.score)
  const starCount = Number.isFinite(score) ? Math.max(1, Math.min(5, Math.round(score))) : 5

  return {
    caseInfo: {
      ...detail.caseInfo,
      iconText: iconOptions.iconText || detail.caseInfo.iconText,
      tone: iconOptions.tone || detail.caseInfo.tone,
      stars: buildStars(starCount)
    },
    rating: {
      ...detail.rating,
      stars: buildStars(starCount)
    },
    detailSections: detail.detailSections,
    players: detail.players
  }
}

Page({
  data: {
    assets: {
      calendar: `${ASSET_BASE}/icon-calendar.svg`,
      users: `${ASSET_BASE}/icon-users.svg`,
      caseFile: `${ASSET_BASE}/icon-case-file.svg`,
      star: `${ASSET_BASE}/icon-star.svg`,
      avatar: `${ASSET_BASE}/avatar-default.png`
    },
    ...buildCaseDetail('case-stranger')
  },

  onLoad(options = {}) {
    this.setData(buildCaseDetail(options.caseId, {
      iconText: decodeOption(options.iconText),
      tone: decodeOption(options.tone)
    }))
  }
})

const EMPTY_SUMMARY = {
  label: '',
  amount: '',
  activeCount: 0,
  pendingSettlementCount: 0,
  completedCount: 0,
  stats: []
}

const ROLE_THEMES = {
  guide: {
    themeClass: 'role-guide',
    iconClass: 'icon-guide'
  },
  expert: {
    themeClass: 'role-expert',
    iconClass: 'icon-income-trend'
  },
  player: {
    themeClass: 'role-player',
    iconClass: 'icon-player'
  },
  default: {
    themeClass: 'role-guide',
    iconClass: 'icon-guide'
  }
}

function firstDefined(...values) {
  return values.find((value) => value !== undefined && value !== null && value !== '')
}

function normalizeRoleType(value) {
  const text = String(value || '').trim().toLowerCase()

  if (text === 'expert' || text === '行家') {
    return 'expert'
  }

  if (text === 'player' || text === '玩家') {
    return 'player'
  }

  return 'guide'
}

function getCount(value, fallback) {
  const count = Number(value)

  return Number.isFinite(count) ? count : fallback
}

function normalizeStats(stats) {
  if (!Array.isArray(stats)) {
    return []
  }

  return stats
    .map((item, index) => {
      if (typeof item === 'string') {
        return {
          key: `stat-${index}`,
          text: item
        }
      }

      const stat = item || {}
      const label = firstDefined(stat.label, stat.title, '')
      const value = firstDefined(stat.value, stat.count, stat.text, '')
      const suffix = firstDefined(stat.suffix, stat.unit, '')
      const text = firstDefined(stat.text, `${label} ${value}${suffix}`)

      return {
        key: firstDefined(stat.key, stat.id, `stat-${index}`),
        text
      }
    })
    .filter((item) => item.text)
}

Component({
  properties: {
    summary: {
      type: Object,
      value: {}
    },
    roleType: {
      type: String,
      value: 'guide'
    },
    label: {
      type: String,
      value: ''
    },
    amount: {
      type: String,
      value: ''
    },
    activeCount: {
      type: null,
      value: null
    },
    pendingSettlementCount: {
      type: null,
      value: null
    },
    completedCount: {
      type: null,
      value: null
    },
    background: {
      type: String,
      value: ''
    },
    iconType: {
      type: String,
      value: ''
    }
  },

  data: {
    resolvedCard: EMPTY_SUMMARY
  },

  observers: {
    'summary, roleType, label, amount, activeCount, pendingSettlementCount, completedCount, background, iconType': function () {
      this.updateCard()
    }
  },

  lifetimes: {
    attached() {
      this.updateCard()
    }
  },

  methods: {
    updateCard() {
      const props = this.properties
      const summary = props.summary || {}
      const roleType = normalizeRoleType(props.roleType)
      const theme = ROLE_THEMES[roleType] || ROLE_THEMES.default
      const iconType = normalizeRoleType(props.iconType || roleType)
      const iconTheme = ROLE_THEMES[iconType] || theme
      const background = props.background || summary.background || summary.bg || ''
      const resolvedCard = {
        label: firstDefined(props.label, summary.label, summary.title, EMPTY_SUMMARY.label),
        amount: firstDefined(props.amount, summary.amountText, summary.serviceIncomeText, summary.monthlyIncomeText, summary.amount, EMPTY_SUMMARY.amount),
        activeCount: getCount(firstDefined(props.activeCount, summary.activeCount, summary.ongoingCount, summary.processingCount), EMPTY_SUMMARY.activeCount),
        pendingSettlementCount: getCount(firstDefined(props.pendingSettlementCount, summary.pendingSettlementCount, summary.settlementCount, summary.waitingSettlementCount), EMPTY_SUMMARY.pendingSettlementCount),
        completedCount: getCount(firstDefined(props.completedCount, summary.completedCount, summary.completeCount, summary.doneCount), EMPTY_SUMMARY.completedCount),
        themeClass: theme.themeClass,
        iconClass: iconTheme.iconClass,
        iconSrc: firstDefined(summary.iconSrc, summary.iconUrl, iconTheme.iconSrc, ''),
        cardStyle: background ? `background: ${background};` : ''
      }

      const customStats = normalizeStats(summary.stats || summary.statList)
      resolvedCard.stats = customStats

      this.setData({
        resolvedCard
      })
    }
  }
})

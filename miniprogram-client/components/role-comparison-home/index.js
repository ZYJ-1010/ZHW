const homeService = require('../../services/home')
const { ROUTES } = require('../../config/routes')
const { navigateShellRoute } = require('../../utils/shell-nav')

const NAV_ITEMS = [
  { name: '我的', active: false },
  { name: '元宇宙', active: false },
  { name: '地图', active: false },
  { name: '消息', active: false },
  { name: '首页', active: true }
]

const ROLE_PRIMARY_TEXT = {
  guide: '申请成为领路人',
  expert: '申请成为行家',
  player: '立即申请角色'
}

function normalizeOnlineText(value) {
  if (value == null) {
    return ''
  }

  return String(value).trim()
}

function normalizeRoleType(value) {
  const roleType = String(value || '').trim()

  if (roleType === 'guide' || roleType === 'leader' || roleType === '领路人') {
    return 'guide'
  }

  if (roleType === 'player' || roleType === '玩家') {
    return 'player'
  }

  return 'expert'
}

function resolveHomeHero(home = {}) {
  const hero = home && home.hero ? home.hero : {}
  const nestedHero = home && home.data && home.data.hero ? home.data.hero : {}

  return {
    ...nestedHero,
    ...hero,
    onlineText: normalizeOnlineText(hero.onlineText) || normalizeOnlineText(nestedHero.onlineText)
  }
}

function formatComparison(comparison = {}, targetRole = 'expert') {
  const roleType = normalizeRoleType(targetRole)
  const roles = Array.isArray(comparison.roles) ? comparison.roles : []
	const benefits = Array.isArray(comparison.benefits) ? comparison.benefits : []
  const primaryText = comparison.primaryOverrideText ||
    (comparison.primaryDisabled && comparison.primaryDisabledText
      ? comparison.primaryDisabledText
      : (ROLE_PRIMARY_TEXT[roleType] || comparison.primary || '立即申请角色'))

  return {
    ...comparison,
    primary: primaryText,
    benefits: benefits.map((benefit) => ({
      ...benefit,
      // 兼容历史后台配置的 leader 字段，前端展示统一使用 guide。
      guide: benefit.guide == null ? benefit.leader : benefit.guide
    })),
    roles: roles.map((role) => {
      const key = normalizeRoleType(role.key || role.roleType || role.name)

      return {
        ...role,
        key,
        active: key === roleType
      }
    })
  }
}

Component({
  options: {
    addGlobalClass: true,
    styleIsolation: 'shared'
  },

  properties: {
    comparison: {
      type: Object,
      value: {}
    },
    returnTo: {
      type: String,
      value: ''
    },
    targetRole: {
      type: String,
      value: 'expert'
    }
  },

  data: {
    onlineText: '',
    navItems: NAV_ITEMS,
    themeRole: 'expert',
    displayComparison: formatComparison({}, 'expert')
  },

  lifetimes: {
    attached() {
      this.updateComparison()
      this.loadOnlineText()
    }
  },

  observers: {
    'comparison,targetRole'() {
      this.updateComparison()
    }
  },

  methods: {
    updateComparison() {
      const roleType = normalizeRoleType(this.properties.targetRole)

      this.setData({
        themeRole: roleType,
        displayComparison: formatComparison(this.properties.comparison || {}, roleType)
      })
    },

    async loadOnlineText() {
      this.setData({
        onlineText: ''
      })

      try {
        const home = await homeService.getHome({ roleType: 'player' })
        const hero = resolveHomeHero(home)

        this.setData({
          onlineText: hero.onlineText
        })
      } catch (error) {
        this.setData({
          onlineText: ''
        })
      }
    },

    handleBackTap() {
      const returnTo = String(this.properties.returnTo || '').trim()

      if (returnTo) {
        navigateShellRoute(returnTo)
        return
      }

      if (typeof getCurrentPages === 'function') {
        const pages = getCurrentPages()

        if (pages.length > 1 && typeof wx.navigateBack === 'function') {
          wx.navigateBack()
          return
        }
      }

      if (typeof wx.reLaunch === 'function') {
        wx.reLaunch({
          url: `/${ROUTES.playerHome}`
        })
      }
    },

    handleApplyTap() {
      if (this.data.displayComparison && this.data.displayComparison.primaryDisabled) {
        return
      }

      this.triggerEvent('apply', {
        roleType: normalizeRoleType(this.properties.targetRole)
      })
    },

    handleRoleSelectTap(event = {}) {
      const roleType = normalizeRoleType(event.currentTarget && event.currentTarget.dataset && event.currentTarget.dataset.role)

      // 玩家身份是默认身份，无需进入角色申请资料流。
      if (roleType === 'player') {
        return
      }

      this.triggerEvent('rolechange', { roleType })
    }
  }
})

const profileService = require('../../services/profile')

const ASSET_BASE = '/pages/profile/member/assets'

const LEVEL_KEYS = ['basic', 'advanced', 'premium']

function normalizeList(value) {
  return Array.isArray(value) ? value : []
}

function buildLevels(levelKey, levelConfig = {}) {
  if (Array.isArray(levelConfig.levels) && levelConfig.levels.length) {
    return levelConfig.levels
  }

  const activeIndex = Number(levelConfig.activeIndex)
  const activeKey = levelConfig.key || levelKey

  return LEVEL_KEYS.map((key, index) => ({
    key,
    active: Number.isInteger(activeIndex) ? index === activeIndex : key === activeKey
  }))
}

function getBenefitItem(items, key, fallbackIndex) {
  return items.find((item) => item && item.key === key) || items[fallbackIndex] || {}
}

function normalizeMemberCenterConfig(data = {}, levelKey = 'basic') {
  const level = data.level || {}
  const benefits = data.benefitsSection || {}
  const benefitItems = normalizeList(benefits.items)
  const referral = getBenefitItem(benefitItems, 'referral', 0)
  const profit = getBenefitItem(benefitItems, 'profit', 1)
  const radar = data.radarSection || {}
  const audience = data.audienceSection || {}
  const openRules = data.openRulesSection || {}
  const purchase = data.purchase || {}
  const latestNotice = data.latestNotice || null
  const roleLink = data.roleLink || {}
  const normalizedLevelKey = level.key || data.levelKey || data.key || levelKey

  return {
    key: normalizedLevelKey,
    memberLevel: level.name || data.memberLevel || '',
    cardClass: data.cardClass || normalizedLevelKey || '',
    benefitSectionTitle: benefits.title || '',
    benefitClass: referral.theme || referral.className || '',
    primaryBenefitTitle: referral.title || '',
    profitBenefitTitle: profit.title || '',
    profitRate: profit.value || profit.rate || data.profitRate || '',
    levels: buildLevels(levelKey, level),
    referralBenefits: normalizeList(referral.points || referral.items || data.referralBenefits),
    radarTitle: radar.title || '',
    radarSubtitle: radar.subtitle || '',
    radarButtonText: radar.buttonText || '',
    radarPeople: normalizeList(radar.people || radar.nodes),
    audienceTitle: audience.title || '',
    audience: normalizeList(audience.items || data.audience),
    openRulesTitle: openRules.title || '',
    openRules: {
      step: openRules.stepText || openRules.step || '',
      notes: normalizeList(openRules.notes)
    },
    roleLinkText: roleLink.text || '',
    roleLinkRoute: roleLink.route || '',
    latestNotice,
    purchase: {
      buttonText: purchase.buttonText || '',
      priceText: purchase.priceText || '',
      route: purchase.route || '',
      agreementPrefix: purchase.agreementPrefix || '',
      agreementName: purchase.agreementName || '',
      agreementRoute: purchase.agreementRoute || '',
      agreementSuffix: purchase.agreementSuffix || '',
      highlightText: purchase.highlightText || ''
    },
    assets: {
      referral: `${ASSET_BASE}/i86@3x.png`,
      profit: `${ASSET_BASE}/i87@3x.png`
    }
  }
}

function buildDisplayData(levelKey) {
  return normalizeMemberCenterConfig({}, levelKey)
}

Component({
  properties: {
    levelKey: {
      type: String,
      value: 'basic'
    }
  },

  data: {
    ...buildDisplayData('basic'),
    loading: false
  },

  observers: {
    levelKey(levelKey) {
      this.loadMemberConfig(levelKey)
    }
  },

  lifetimes: {
    attached() {
      this.loadMemberConfig(this.properties.levelKey)
    }
  },

  methods: {
    async loadMemberConfig(levelKey = 'basic') {
      this.setData({
        ...buildDisplayData(levelKey),
        loading: true
      })

      try {
        const data = await profileService.getMemberCenterConfig({
          level: levelKey
        })

        this.setData({
          ...normalizeMemberCenterConfig(data, levelKey),
          loading: false
        })
      } catch (error) {
        this.setData({
          loading: false
        })
        wx.showToast({
          title: error.message || '会员中心配置加载失败',
          icon: 'none'
        })
      }
    },

    handleMatch() {
      wx.navigateTo({
        url: '/pages/profile/member/radar/index'
      })
    },

    handleUnlockRole() {
      wx.navigateTo({
        url: this.data.roleLinkRoute ? `/${this.data.roleLinkRoute}` : '/pages/role/apply/index'
      })
    },

    handleOpenMember() {
      if (this.data.purchase && this.data.purchase.route) {
        wx.navigateTo({
          url: this.data.purchase.route.startsWith('/')
            ? this.data.purchase.route
            : `/${this.data.purchase.route}`
        })
        return
      }

      wx.showToast({
        title: '会员支付待接入',
        icon: 'none'
      })
    },

    handleAgreement() {
      const route = this.data.purchase && this.data.purchase.agreementRoute

      if (route) {
        wx.navigateTo({
          url: route.startsWith('/') ? route : `/${route}`
        })
        return
      }

      wx.showToast({
        title: '服务协议待补充',
        icon: 'none'
      })
    }
  }
})

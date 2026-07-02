const profileApi = require('../../api/modules/profile')
const { navigateShellRoute } = require('../../utils/shell-nav')

const ASSET_BASE = '/pages/profile/member/assets'

const COMMON_RULES = {
  step: '',
  notes: []
}

const COMMON_RADAR_PEOPLE = [
  { id: 'me', className: 'me', avatarUrl: '/pages/home/player/assets/ranking-avatar-me.png', text: '我' },
  { id: 'expert', className: 'expert', avatarUrl: '/pages/home/player/assets/ranking-avatar-01.png', text: '行' },
  { id: 'guide', className: 'guide', avatarUrl: '/pages/home/player/assets/ranking-avatar-02.png', text: '领' },
  { id: 'player', className: 'player', avatarUrl: '/pages/home/player/assets/ranking-avatar-03.png', text: '玩' },
  { id: 'nearby', className: 'nearby', avatarUrl: '', text: '局' },
  { id: 'friend', className: 'friend', avatarUrl: '', text: '友' },
  { id: 'resource', className: 'resource', avatarUrl: '', text: '资' }
]

const MEMBER_CONFIGS = {
  basic: {
    key: 'basic',
    memberLevel: '',
    cardClass: 'basic',
    benefitClass: 'purple',
    primaryBenefitTitle: '',
    price: '',
    profitRate: '',
    noticeLevel: '',
    noticeAvatar: '',
    noticeName: '',
    agreementKey: '',
    agreementTitle: '',
    levels: [
      { key: 'basic', active: true },
      { key: 'advanced', active: false },
      { key: 'premium', active: false }
    ],
    referralBenefits: [],
    audience: []
  },
  advanced: {
    key: 'advanced',
    memberLevel: '',
    cardClass: 'advanced',
    benefitClass: 'advanced',
    primaryBenefitTitle: '',
    price: '',
    profitRate: '',
    noticeLevel: '',
    noticeAvatar: '',
    noticeName: '',
    agreementKey: '',
    agreementTitle: '',
    levels: [
      { key: 'basic', active: false },
      { key: 'advanced', active: true },
      { key: 'premium', active: false }
    ],
    referralBenefits: [],
    audience: []
  },
  premium: {
    key: 'premium',
    memberLevel: '',
    cardClass: 'premium',
    benefitClass: 'premium',
    primaryBenefitTitle: '',
    price: '',
    profitRate: '',
    noticeLevel: '',
    noticeAvatar: '',
    noticeName: '',
    agreementKey: '',
    agreementTitle: '',
    levels: [
      { key: 'basic', active: false },
      { key: 'advanced', active: false },
      { key: 'premium', active: true }
    ],
    referralBenefits: [],
    audience: []
  }
}

function normalizePlan(plan = {}, fallback = {}) {
  const key = plan.key || plan.code || fallback.key || 'basic'

  return {
    ...fallback,
    ...plan,
    key,
    memberLevel: plan.memberLevel || plan.name || fallback.memberLevel || '',
    price: String(plan.price || fallback.price || ''),
    profitRate: plan.profitRate || fallback.profitRate || '',
    noticeLevel: plan.noticeLevel || fallback.noticeLevel || '',
    noticeAvatar: plan.noticeAvatar || fallback.noticeAvatar || '',
    noticeName: plan.noticeName || fallback.noticeName || '',
    agreementKey: plan.agreementKey || fallback.agreementKey || 'user-service',
    agreementTitle: plan.agreementTitle || fallback.agreementTitle || '服务协议',
    referralBenefits: Array.isArray(plan.referralBenefits) ? plan.referralBenefits : (fallback.referralBenefits || []),
    audience: Array.isArray(plan.audience) ? plan.audience : (fallback.audience || []),
    levels: buildLevels(key),
    openRules: normalizeRules(plan.openRules || fallback.openRules)
  }
}

function normalizeRules(rules = {}) {
  return {
    step: rules.step || COMMON_RULES.step,
    notes: Array.isArray(rules.notes) ? rules.notes : COMMON_RULES.notes
  }
}

function buildLevels(activeKey) {
  return ['basic', 'advanced', 'premium'].map((key) => ({
    key,
    active: key === activeKey
  }))
}

function plansToConfigMap(plans = []) {
  return plans.reduce((result, plan) => {
    const key = plan.key || plan.code
    if (key) {
      result[key] = normalizePlan(plan, MEMBER_CONFIGS[key] || {})
    }
    return result
  }, {})
}

function buildDisplayData(levelKey, configs) {
  const map = configs || MEMBER_CONFIGS
  const config = map[levelKey] || map.basic || MEMBER_CONFIGS.basic

  return {
    ...config,
    assets: {
      referral: `${ASSET_BASE}/i86@3x.png`,
      profit: `${ASSET_BASE}/i87@3x.png`
    },
    radarPeople: COMMON_RADAR_PEOPLE,
    openRules: normalizeRules(config.openRules)
  }
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
    planConfigs: MEMBER_CONFIGS
  },

  observers: {
    levelKey(levelKey) {
      this.setData(buildDisplayData(levelKey, this.data.planConfigs))
    }
  },

  lifetimes: {
    attached() {
      this.loadPlans()
    }
  },

  methods: {
    async loadPlans() {
      try {
        const data = await profileApi.getMemberPlans()
        const planConfigs = plansToConfigMap(data.items || data.plans || [])
        const nextConfigs = Object.keys(planConfigs).length ? planConfigs : MEMBER_CONFIGS
        this.setData({
          planConfigs: nextConfigs,
          ...buildDisplayData(this.properties.levelKey, nextConfigs)
        })
      } catch (error) {
        console.warn('get membership plans failed', error)
        this.setData(buildDisplayData(this.properties.levelKey, this.data.planConfigs))
      }
    },

    handleMatch() {
      navigateShellRoute('/pages/profile/member/radar/index')
    },

    handleUnlockRole() {
      navigateShellRoute('/pages/role/apply/index')
    },

    handleOpenMember() {
      const target = this.properties.levelKey === 'premium'
        ? '/pages/profile/member/premium/index'
        : '/pages/profile/member/advanced/index'

      navigateShellRoute(target)
    },

    handleAgreement() {
      const agreementKey = this.data.agreementKey || 'user-service'
      const title = this.data.agreementTitle || '服务协议'

      navigateShellRoute(`/pages/profile/system-management/agreement-detail/index?agreement=${encodeURIComponent(agreementKey)}&title=${encodeURIComponent(title)}&signed=0`)
    }
  }
})

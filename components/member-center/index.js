const ASSET_BASE = '/pages/profile/member/assets'

const COMMON_RULES = {
  step: '1. 选择会员等级 → 2. 在线支付 → 3. 即时生效',
  notes: [
    '支持微信支付、支持银行卡支付',
    '升级后原有权益自动叠加，不重复收费',
    '如需帮助，请联系客服：400-XXX-XXXX'
  ]
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
    memberLevel: '基础会员',
    cardClass: 'basic',
    benefitClass: 'purple',
    primaryBenefitTitle: '业务引荐权益',
    price: '515',
    profitRate: '40%',
    noticeLevel: '高级会员',
    levels: [
      { key: 'basic', active: true },
      { key: 'advanced', active: false },
      { key: 'premium', active: false }
    ],
    referralBenefits: [
      '可引荐平台业务',
      '享受引荐收益'
    ],
    audience: [
      '有人脉、善对接的社交达人、资源型人才；',
      '希望不做销售、不投重金，只靠人脉赚钱；',
      '有高客单价产品/资源，想初步了解；',
      '连接供需，促成交易'
    ]
  },
  advanced: {
    key: 'advanced',
    memberLevel: '高级会员',
    cardClass: 'advanced',
    benefitClass: 'advanced',
    primaryBenefitTitle: '业务被引荐权益',
    price: '10000',
    profitRate: '40%',
    noticeLevel: '高级会员',
    levels: [
      { key: 'basic', active: false },
      { key: 'advanced', active: true },
      { key: 'premium', active: false }
    ],
    referralBenefits: [
      '您的产品或服务进入引荐池',
      '平台引荐人主动为您引荐',
      '享受被引荐带来的40%收益分成',
      '提供推广数据报表和分析'
    ],
    audience: [
      '有高客单价产品/资源、有技术的专业人士',
      '希望获得额外收入来源'
    ]
  },
  premium: {
    key: 'premium',
    memberLevel: '尊享会员',
    cardClass: 'premium',
    benefitClass: 'premium',
    primaryBenefitTitle: '渠道引荐权益',
    price: '39800',
    profitRate: '10%',
    noticeLevel: '高级会员',
    levels: [
      { key: 'basic', active: false },
      { key: 'advanced', active: false },
      { key: 'premium', active: true }
    ],
    referralBenefits: [
      '可发展和管理下级推广团队',
      '团队引荐收益的10%作为管理奖励',
      '提供团队管理工具和数据看板'
    ],
    audience: [
      '有人脉、爱分享、想裂变的推广能手',
      '团队管理者、团长、行业主理人',
      '拥有推广资源的个人/机构',
      '希望建立推广体系的创业者'
    ]
  }
}

function buildDisplayData(levelKey) {
  const config = MEMBER_CONFIGS[levelKey] || MEMBER_CONFIGS.basic

  return {
    ...config,
    assets: {
      referral: `${ASSET_BASE}/i86@3x.png`,
      profit: `${ASSET_BASE}/i87@3x.png`
    },
    radarPeople: COMMON_RADAR_PEOPLE,
    openRules: COMMON_RULES
  }
}

Component({
  properties: {
    levelKey: {
      type: String,
      value: 'basic'
    }
  },

  data: buildDisplayData('basic'),

  observers: {
    levelKey(levelKey) {
      this.setData(buildDisplayData(levelKey))
    }
  },

  lifetimes: {
    attached() {
      this.setData(buildDisplayData(this.properties.levelKey))
    }
  },

  methods: {
    handleMatch() {
      wx.showToast({
        title: '适配功能待接入',
        icon: 'none'
      })
    },

    handleUnlockRole() {
      wx.navigateTo({
        url: '/pages/role/apply/index'
      })
    },

    handleOpenMember() {
      wx.showToast({
        title: '会员支付待接入',
        icon: 'none'
      })
    },

    handleAgreement() {
      wx.showToast({
        title: '服务协议待补充',
        icon: 'none'
      })
    }
  }
})

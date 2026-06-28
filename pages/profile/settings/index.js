const ASSET_BASE = '/pages/profile/settings/assets'

function asset(name) {
  return `${ASSET_BASE}/${name}`
}

const SETTING_SECTIONS = [
  {
    title: '账号安全',
    rows: [
      { id: 'payPassword', label: '支付密码', icon: asset('icon-pay-lock.png'), value: '未设置', arrow: true },
      { id: 'loginPassword', label: '登录密码', icon: asset('icon-login-lock.png'), arrow: true },
      { id: 'phone', label: '更换手机号', icon: asset('icon-mobile.png'), value: '138****8888', arrow: true },
      { id: 'facePay', label: '指纹/面容支付', icon: asset('icon-pay-shield.png'), switch: true, enabled: true }
    ]
  },
  {
    title: '通知设置',
    rows: [
      { id: 'gamePush', label: '局消息推送', icon: asset('icon-bell.png'), switch: true, enabled: true },
      { id: 'systemNotice', label: '系统通知', icon: asset('icon-calendar.png'), switch: true, enabled: true },
      { id: 'subscribeNotice', label: '订阅消息', icon: asset('icon-mail.png'), switch: true, enabled: true },
      { id: 'emailNotice', label: '邮件通知', icon: asset('icon-mail-off.png'), switch: true, enabled: false },
      { id: 'quietHours', label: '消息免打扰', icon: asset('icon-clock.png'), value: '22:00 - 08:00', arrow: true }
    ]
  },
  {
    title: '隐私设置',
    rows: [
      { id: 'showGames', label: '允许他人查看我的局', icon: asset('icon-eye.png'), switch: true, enabled: true },
      { id: 'showReviews', label: '允许他人查看我的评价', icon: asset('icon-comment.png'), switch: true, enabled: true },
      { id: 'findByPhone', label: '允许通过手机号找到我', icon: asset('icon-phone.png'), switch: true, enabled: false },
      { id: 'personalized', label: '个性化推送', icon: asset('icon-star.png'), switch: true, enabled: true },
      { id: 'privacySummary', label: '隐私政策摘要', icon: asset('icon-document.png'), arrow: true },
      { id: 'thirdPartyList', label: '第三方共享清单', icon: asset('icon-users.png'), arrow: true },
      { id: 'collectionList', label: '信息收集清单', icon: asset('icon-document-list.png'), arrow: true }
    ]
  },
  {
    title: '通用设置',
    rows: [
      { id: 'clearCache', label: '清除缓存', icon: asset('icon-trash.png'), value: '当前缓存 23.5MB', arrow: true },
      { id: 'about', label: '关于我们', icon: asset('icon-info.png'), value: '版本 v1.0.0', arrow: true }
    ]
  }
]

Page({
  data: {
    sections: SETTING_SECTIONS,
    chevronIcon: asset('icon-chevron-right.svg'),
    previewMode: 'static'
  },

  onLoad(options = {}) {
    this.setData({
      previewMode: options.mode || 'static'
    })
  },

  handleSwitch(event) {
    const { sectionIndex, rowIndex } = event.currentTarget.dataset
    const key = `sections[${sectionIndex}].rows[${rowIndex}].enabled`
    const current = this.data.sections[sectionIndex].rows[rowIndex].enabled

    this.setData({
      [key]: !current
    })
  },

  handleRowTap(event) {
    const { label } = event.currentTarget.dataset

    if (!label || typeof wx === 'undefined') {
      return
    }

    wx.showToast({
      title: '待接入',
      icon: 'none'
    })
  },

  handleLogout() {
    if (typeof wx === 'undefined') {
      return
    }

    wx.showToast({
      title: '退出登录待接入',
      icon: 'none'
    })
  }
})

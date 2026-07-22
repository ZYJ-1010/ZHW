Component({
  properties: {
    participant: {
      type: Object,
      value: {}
    },
    compact: {
      type: Boolean,
      value: false
    }
  },

  methods: {
    handleBlueBadgeTap() {
      const badge = (this.properties.participant || {}).blueBadge || {}
      wx.showModal({
        title: badge.label || '蓝标认证',
        content: [badge.ruleText, badge.contactText].filter(Boolean).join('\n'),
        showCancel: false,
        confirmText: '我知道了'
      })
    },

    handleTap() {
      this.triggerEvent('tapcard', {
        participant: this.properties.participant
      })
    }
  }
})

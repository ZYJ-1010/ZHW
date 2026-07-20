Component({
  properties: {
    item: {
      type: Object,
      value: {},
      observer(item) {
        this.syncTagTone(item)
      }
    },
    variant: {
      type: String,
      value: ''
    }
  },

  data: {
    defaultCover: 'https://static.haowan.net.cn/miniprogram/components/game-card/assets/game-cover-default.png',
    tagTone: '',
    joinedText: ''
  },

  methods: {
    syncTagTone(item) {
      this.setData({
        tagTone: this.resolveTagTone(item),
        joinedText: this.resolveJoinedText(item)
      })
    },

    resolveTagTone(item = {}) {
      const presetTone = item.tagTone || item.statusTone || item.typeTone

      if (presetTone) {
        return presetTone
      }

      const label = item.tag || item.statusText || item.typeName || ''

      if (label.indexOf('\u4efb\u52a1') > -1) {
        return 'task'
      }

      if (label.indexOf('\u63a2\u7d22') > -1) {
        return 'explore'
      }

      return ''
    },

    resolveJoinedText(item = {}) {
      const text = item.joinedShortText || item.joinedText || ''

      return text
        .replace(/\+(\d+)\u4f4d\u73a9\u5bb6\u5df2\u5165\u5c40/, '+$1\u5df2\u5165\u5c40')
        .replace(/(\d+)\u4f4d\u73a9\u5bb6\u5df2\u5165\u5c40/, '$1\u5df2\u5165\u5c40')
    },

    handleTap() {
      this.triggerEvent('cardtap', {
        item: this.properties.item
      })
    },

    handlePrimaryTap() {
      const item = this.properties.item || {}

      this.triggerEvent('cardaction', {
        item,
        action: 'primary',
        label: item.action || ''
      })
    },

    handleActionTap(event) {
      const action = event.currentTarget.dataset.action || ''

      this.triggerEvent('cardaction', {
        item: this.properties.item,
        action,
        label: action
      })
    }
  }
})

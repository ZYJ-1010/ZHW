Component({
  properties: {
    item: {
      type: Object,
      value: {}
    }
  },

  data: {
    defaultCover: '/components/game-card/assets/game-cover-default.png',
    defaultAvatars: ['🤒', '🤡', '😵']
  },

  methods: {
    handleTap() {
      this.triggerEvent('cardtap', {
        item: this.properties.item
      })
    }
  }
})

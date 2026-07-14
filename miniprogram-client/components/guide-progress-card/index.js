Component({
  properties: {
    item: {
      type: Object,
      value: {}
    }
  },

  methods: {
    onCardTap() {
      this.triggerEvent('cardtap', {
        item: this.properties.item
      })
    }
  }
})

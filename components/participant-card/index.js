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
    handleTap() {
      this.triggerEvent('tapcard', {
        participant: this.properties.participant
      })
    }
  }
})

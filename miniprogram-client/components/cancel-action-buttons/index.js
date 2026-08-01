Component({
  properties: {
    barStyle: {
      type: String,
      value: ''
    },
    barClass: {
      type: String,
      value: ''
    },
    secondaryText: {
      type: String,
      value: '再考虑一下'
    },
    primaryText: {
      type: String,
      value: '确认取消'
    },
    primaryDisabled: {
      type: Boolean,
      value: false
    },
    primaryLoading: {
      type: Boolean,
      value: false
    },
    primaryTone: {
      type: String,
      value: 'danger'
    },
    tipText: {
      type: String,
      value: ''
    }
  },

  methods: {
    onSecondaryTap() {
      this.triggerEvent('secondarytap')
    },

    onPrimaryTap() {
      if (this.properties.primaryDisabled || this.properties.primaryLoading) {
        return
      }

      this.triggerEvent('primarytap')
    }
  }
})

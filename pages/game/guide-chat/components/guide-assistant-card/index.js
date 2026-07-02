Component({
  properties: {
    card: {
      type: Object,
      value: {}
    }
  },

  methods: {
    handleAcceptTap() {
      this.triggerEvent('accept')
    },

    handleDeclineTap() {
      this.triggerEvent('decline')
    }
  }
})

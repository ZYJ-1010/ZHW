Component({
  options: {
    addGlobalClass: true,
    styleIsolation: 'shared'
  },

  properties: {
    tip: {
      type: String,
      value: ''
    },
    loading: {
      type: Boolean,
      value: false
    },
    loadingText: {
      type: String,
      value: '处理中...'
    },
    confirmText: {
      type: String,
      value: '确认参加'
    }
  },

  methods: {
    handleDecline() {
      if (this.data.loading) {
        return
      }

      this.triggerEvent('decline')
    },

    handleConfirm() {
      if (this.data.loading) {
        return
      }

      this.triggerEvent('confirm')
    }
  }
})

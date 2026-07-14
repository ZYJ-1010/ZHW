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
    declineText: {
      type: String,
      value: '婉拒'
    },
    confirmText: {
      type: String,
      value: '确认参加'
    },
    confirmDisabled: {
      type: Boolean,
      value: false
    },
    confirmDisabledReason: {
      type: String,
      value: ''
    },
    singleReadonly: {
      type: Boolean,
      value: false
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
      if (this.data.confirmDisabled) {
        this.triggerEvent('confirmdisabled', {
          reason: this.data.confirmDisabledReason
        })
        return
      }

      this.triggerEvent('confirm')
    }
  }
})

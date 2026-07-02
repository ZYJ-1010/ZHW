Component({
  properties: {
    title: {
      type: String,
      value: '取消原因'
    },
    requiredText: {
      type: String,
      value: '（必填）'
    },
    options: {
      type: Array,
      value: []
    },
    selectedKey: {
      type: String,
      value: ''
    },
    value: {
      type: String,
      value: ''
    },
    placeholder: {
      type: String,
      value: '详细说明取消原因...'
    },
    maxlength: {
      type: Number,
      value: 50
    },
    variant: {
      type: String,
      value: 'center'
    },
    cardClass: {
      type: String,
      value: ''
    }
  },

  data: {
    displayOptions: []
  },

  observers: {
    'options, selectedKey': function () {
      this.refreshDisplayOptions()
    }
  },

  lifetimes: {
    attached() {
      this.refreshDisplayOptions()
    }
  },

  methods: {
    getSelectedKeys() {
      if (this.properties.selectedKey) {
        return [this.properties.selectedKey]
      }

      return (this.properties.options || [])
        .filter((item) => item && item.active)
        .map((item) => item.key)
        .filter(Boolean)
    },

    refreshDisplayOptions() {
      const selectedKeySet = this.getSelectedKeys().reduce((result, key) => {
        result[key] = true
        return result
      }, {})

      this.setData({
        displayOptions: (this.properties.options || []).map((item) => ({
          ...item,
          active: Boolean(selectedKeySet[item.key] || item.active),
          showCheck: this.properties.variant === 'check' && Boolean(selectedKeySet[item.key] || item.active)
        }))
      })
    },

    onOptionTap(event) {
      const key = event.currentTarget.dataset.key

      this.triggerEvent('reasonchange', {
        key,
        selectedKey: key
      })
    },

    onInput(event) {
      this.triggerEvent('reasoninput', {
        value: event.detail.value
      })
    }
  }
})

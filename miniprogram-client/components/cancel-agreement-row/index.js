Component({
  properties: {
    checked: {
      type: Boolean,
      value: false
    },
    content: {
      type: String,
      value: ''
    },
    rowClass: {
      type: String,
      value: ''
    }
  },

  methods: {
    onToggle() {
      this.triggerEvent('toggle', {
        checked: !this.properties.checked
      })
    }
  }
})

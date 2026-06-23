Component({
  properties: {
    title: {
      type: String,
      value: ''
    },
    items: {
      type: Array,
      value: []
    },
    variant: {
      type: String,
      value: 'default'
    },
    customClass: {
      type: String,
      value: ''
    }
  },

  methods: {
    onActionTap(event) {
      const { index } = event.currentTarget.dataset

      this.triggerEvent('actiontap', {
        index,
        item: this.properties.items[index]
      })
    }
  }
})

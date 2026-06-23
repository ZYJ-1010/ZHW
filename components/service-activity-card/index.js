const DEFAULT_INFO = {
  title: '活动信息',
  avatarText: '',
  avatarClass: '',
  name: '',
  roleLabel: '',
  roleClass: '',
  serviceTitle: '',
  rows: []
}

Component({
  properties: {
    info: {
      type: Object,
      value: {}
    }
  },

  data: {
    resolvedInfo: DEFAULT_INFO
  },

  observers: {
    info(info) {
      this.setData({
        resolvedInfo: this.normalizeInfo(info)
      })
    }
  },

  lifetimes: {
    attached() {
      this.setData({
        resolvedInfo: this.normalizeInfo(this.properties.info)
      })
    }
  },

  methods: {
    normalizeInfo(info = {}) {
      return {
        ...DEFAULT_INFO,
        ...info,
        rows: Array.isArray(info.rows) ? info.rows : []
      }
    }
  }
})

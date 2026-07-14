const { getSurnameInitials } = require('../../utils/avatar')

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
      const name = info.name || info.expertName || info.playerName || info.guideName || ''

      return {
        ...DEFAULT_INFO,
        ...info,
        avatarText: getSurnameInitials(name, info.avatarText || DEFAULT_INFO.avatarText),
        rows: Array.isArray(info.rows) ? info.rows : []
      }
    }
  }
})

const { getSurnameInitials } = require('../../utils/avatar')

const EMPTY_INFO = {
  headerTitle: '',
  title: '',
  guideLabel: '',
  guideName: '',
  expertLabel: '',
  miniProgramText: '',
  confirmText: '',
  expert: {
    name: '',
    avatarText: '',
    desc: '',
    intro: '',
    tags: []
  },
  stats: []
}

Component({
  properties: {
    info: {
      type: Object,
      value: {}
    }
  },

  data: {
    resolvedInfo: EMPTY_INFO
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
      const expert = {
        ...EMPTY_INFO.expert,
        ...(info.expert || {})
      }
      const normalizedExpert = {
        ...expert,
        avatarText: getSurnameInitials(expert.name, expert.avatarText)
      }

      return {
        ...EMPTY_INFO,
        ...info,
        expert: normalizedExpert,
        stats: Array.isArray(info.stats) ? info.stats : EMPTY_INFO.stats
      }
    },

    handleConfirmTap() {
      this.triggerEvent('confirm', {
        info: this.data.resolvedInfo
      })
    }
  }
})

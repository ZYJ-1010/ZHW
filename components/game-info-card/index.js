const { getSurnameInitials } = require('../../utils/avatar')

const DEFAULT_INFO = {
  headerTitle: '组局信息',
  title: '',
  guideLabel: '领路人',
  guideName: '',
  expertLabel: '行家',
  miniProgramText: '小程序 · 真好玩',
  confirmText: '查看详情并确认',
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
      const expert = {
        ...DEFAULT_INFO.expert,
        ...(info.expert || {})
      }
      const normalizedExpert = {
        ...expert,
        avatarText: getSurnameInitials(expert.name, expert.avatarText)
      }

      return {
        ...DEFAULT_INFO,
        ...info,
        expert: normalizedExpert,
        stats: Array.isArray(info.stats) ? info.stats : DEFAULT_INFO.stats
      }
    },

    handleConfirmTap() {
      this.triggerEvent('confirm', {
        info: this.data.resolvedInfo
      })
    }
  }
})

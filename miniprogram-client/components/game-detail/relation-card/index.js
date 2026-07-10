Component({
  options: {
    addGlobalClass: true,
    styleIsolation: 'shared'
  },

  data: {
    resolvedRelation: {}
  },

  properties: {
    relation: {
      type: Object,
      value: {},
      observer() {
        this.updateResolvedRelation()
      }
    }
  },

  lifetimes: {
    attached() {
      this.updateResolvedRelation()
    }
  },

  methods: {
    updateResolvedRelation() {
      const relation = this.properties.relation || {}
      const expert = relation.expert || {}
      const guide = relation.guide || {}
      const player = relation.player || {}
      const hasConfirmedText = typeof relation.confirmedText === 'string' && relation.confirmedText
      const totalCount = Number(relation.totalCount || 0)
      const confirmedCount = Number(relation.confirmedCount || 0)
      const noticeVisible = typeof relation.noticeVisible === 'boolean'
        ? relation.noticeVisible
        : Boolean(player.confirmed)

      this.setData({
        resolvedRelation: {
          ...relation,
          title: relation.title || '组局关系图',
          confirmedText: hasConfirmedText ? relation.confirmedText : `${confirmedCount}/${totalCount}人已确认`,
          noticeVisible,
          noticeText: relation.noticeText || (player.confirmed ? '玩家已确认意向，等待你最终审核组局' : ''),
          expert,
          guide,
          player: {
            ...player,
            statusText: player.statusText || (player.confirmed ? '玩家已确认' : '待确认')
          }
        }
      })
    }
  }
})

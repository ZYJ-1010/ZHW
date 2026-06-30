const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

const ASSET_BASE = '/pages/profile/system-management/block-settings/assets'

Page({
  data: {
    icons: {
      lightbulb: `${ASSET_BASE}/icon-lightbulb.svg`,
      target: `${ASSET_BASE}/icon-target.svg`,
      warning: `${ASSET_BASE}/icon-warning.png`
    },
    keywordInput: '',
    keywords: [],
    keywordItems: [],
    suggestions: [],
    stats: [],
    rules: []
  },

  onLoad() {
    this.loadKeywords()
  },

  async loadKeywords() {
    try {
      const data = await profileService.getSystemBlockKeywords()
      const keywordItems = this.normalizeList(data.keywords || data.list || data.items)

      this.setData({
        keywordItems,
        keywords: keywordItems.map((item) => item.keyword || item.text || item.name || '').filter(Boolean),
        suggestions: this.normalizeList(data.suggestions || data.recommendations)
          .map((item) => item.keyword || item.text || item.name || item)
          .filter(Boolean),
        stats: this.normalizeList(data.stats),
        rules: this.normalizeList(data.rules || data.ruleLines)
      })
    } catch (error) {
      this.setData({
        keywords: [],
        keywordItems: [],
        suggestions: [],
        stats: [],
        rules: []
      })
      toast.info(error.message || '关键词屏蔽加载失败')
    }
  },

  handleKeywordInput(event) {
    this.setData({
      keywordInput: event.detail.value
    })
  },

  async addKeyword(keyword) {
    const value = (keyword || this.data.keywordInput || '').trim()

    if (!value) {
      toast.info('请输入关键词')
      return
    }

    if (this.data.keywords.includes(value)) {
      toast.info('关键词已存在')
      return
    }

    try {
      await profileService.addSystemBlockKeyword({
        keyword: value
      })
      this.setData({
        keywordInput: ''
      })
      this.loadKeywords()
    } catch (error) {
      toast.info(error.message || '添加关键词失败')
    }
  },

  handleAddTap() {
    this.addKeyword()
  },

  handleSuggestionTap(event) {
    this.addKeyword(event.currentTarget.dataset.keyword)
  },

  async handleRemoveTap(event) {
    const index = Number(event.currentTarget.dataset.index)
    const item = this.data.keywordItems[index]
    const keywordId = item && (item.keywordId || item.id)

    if (!keywordId) {
      toast.info('缺少关键词信息')
      return
    }

    try {
      await profileService.removeSystemBlockKeyword({
        keywordId
      })
      this.loadKeywords()
    } catch (error) {
      toast.info(error.message || '删除关键词失败')
    }
  },

  async handleResetTap() {
    const ids = this.data.keywordItems
      .map((item) => item.keywordId || item.id)
      .filter(Boolean)

    if (!ids.length) {
      toast.info('暂无关键词可重置')
      return
    }

    try {
      await Promise.all(ids.map((keywordId) => profileService.removeSystemBlockKeyword({ keywordId })))
      this.loadKeywords()
      toast.info('关键词已重置')
    } catch (error) {
      toast.info(error.message || '关键词重置失败')
    }
  },

  handleSaveTap() {
    toast.success('关键词屏蔽已同步')
  },

  normalizeList(list) {
    return Array.isArray(list) ? list : []
  }
})

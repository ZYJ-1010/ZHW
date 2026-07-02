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
    blockSettings: null,
    keywordInput: '',
    keywords: [],
    suggestions: [],
    stats: [],
    rules: []
  },

  onLoad() {
    this.loadKeywords()
  },

  async loadKeywords() {
    try {
      const data = await profileService.getSystemBlockSettings()
      this.setData({
        blockSettings: data || null,
        keywords: Array.isArray(data.keywords) ? data.keywords : [],
        suggestions: Array.isArray(data.suggestions) ? data.suggestions : [],
        stats: Array.isArray(data.keywordStats) ? data.keywordStats : [],
        rules: Array.isArray(data.keywordRules) ? data.keywordRules : []
      })
    } catch (error) {
      this.setData({ keywords: [], suggestions: [], stats: [], rules: [] })
      toast.info(error.message || '关键词配置暂时不可用')
    }
  },

  handleKeywordInput(event) {
    this.setData({
      keywordInput: event.detail.value
    })
  },

  addKeyword(keyword) {
    const value = (keyword || this.data.keywordInput || '').trim()

    if (!value) {
      toast.info('请输入关键词')
      return
    }

    if (this.data.keywords.includes(value)) {
      toast.info('关键词已存在')
      return
    }

    if (this.data.keywords.length >= 20) {
      toast.info('最多可设置20个关键词')
      return
    }

    this.setData({
      keywords: this.data.keywords.concat(value),
      keywordInput: ''
    })
  },

  handleAddTap() {
    this.addKeyword()
  },

  handleSuggestionTap(event) {
    this.addKeyword(event.currentTarget.dataset.keyword)
  },

  handleRemoveTap(event) {
    const index = Number(event.currentTarget.dataset.index)

    this.setData({
      keywords: this.data.keywords.filter((_, itemIndex) => itemIndex !== index)
    })
  },

  handleResetTap() {
    this.setData({
      keywords: []
    })
    toast.info('关键词已重置')
  },

  async handleSaveTap() {
    try {
      await profileService.saveSystemBlockSettings({
        ...(this.data.blockSettings || {}),
        keywords: this.data.keywords
      })
      toast.success('关键词屏蔽已保存')
    } catch (error) {
      toast.info(error.message || '保存失败')
    }
  }
})

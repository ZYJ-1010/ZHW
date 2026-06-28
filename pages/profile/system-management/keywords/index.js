const toast = require('../../../../utils/toast')

const ASSET_BASE = '/pages/profile/system-management/block-settings/assets'

Page({
  data: {
    icons: {
      lightbulb: `${ASSET_BASE}/icon-lightbulb.svg`,
      target: `${ASSET_BASE}/icon-target.svg`,
      warning: `${ASSET_BASE}/icon-warning.png`
    },
    keywordInput: '',
    keywords: ['培训', '课程', '收费教学', '加微信', '私下'],
    suggestions: ['微商', '直销', '刷单', '贷款', '兼职', '代理', '拉群', '推广'],
    stats: [
      { value: '5', label: '已设置' },
      { value: '20', label: '上限', color: 'gold' },
      { value: '模糊', label: '匹配模式', color: 'green' }
    ],
    rules: [
      { prefix: '支持', strong: '模糊匹配', suffix: '，如"培训"匹配"培训机构""培训课程"' },
      { prefix: '最多可设', strong: '20个', suffix: '关键词' },
      { prefix: '关键词屏蔽仅影响内容展示，', strong: '不影响用户间交互', suffix: '' },
      { prefix: '生效范围：局标题、描述、评论、私信内容', strong: '', suffix: '' }
    ]
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

  handleSaveTap() {
    toast.success('关键词屏蔽已保存')
  }
})

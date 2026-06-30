const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

const LOCAL_ASSET_BASE = '/pages/profile/system-management/credit-appeal/assets'
const BLOCK_ASSET_BASE = '/pages/profile/system-management/block-settings/assets'

Page({
  data: {
    activeReason: '',
    recordId: '',
    appealContent: '',
    appealContentLength: 0,
    submitClass: 'disabled',
    icons: {
      ban: `${BLOCK_ASSET_BASE}/icon-ban.svg`,
      chevron: `${BLOCK_ASSET_BASE}/icon-chevron-right.svg`,
      plus: `${BLOCK_ASSET_BASE}/icon-plus.svg`,
      clock: `${LOCAL_ASSET_BASE}/icon-clock.svg`
    },
    reasonOptions: [],
    relatedRecord: {},
    appealPlaceholder: '',
    uploadRequirement: '',
    uploadSlots: [
      { key: 'slot-1' },
      { key: 'slot-2' },
      { key: 'slot-3' },
      { key: 'slot-4' }
    ],
    reviewTitle: '',
    reviewRules: []
  },

  onLoad(options = {}) {
    this.setData({
      recordId: options.recordId || options.id || ''
    })
    this.loadAppealOptions()
  },

  async loadAppealOptions() {
    try {
      const data = await profileService.getCreditAppealOptions({
        recordId: this.data.recordId
      })
      const activeReason = data.defaultReason || data.activeReason || data.reasons && data.reasons[0] && data.reasons[0].key || ''

      this.setData({
        activeReason,
        reasonOptions: this.buildReasonOptions(data.reasons || data.reasonOptions, activeReason),
        relatedRecord: data.relatedRecord || data.record || {},
        appealPlaceholder: data.appealPlaceholder || data.placeholder || '',
        uploadRequirement: data.uploadRequirement || '',
        reviewTitle: data.reviewTitle || '',
        reviewRules: this.normalizeList(data.reviewRules || data.rules)
      })
    } catch (error) {
      this.setData({
        activeReason: '',
        reasonOptions: [],
        relatedRecord: {},
        appealPlaceholder: '',
        uploadRequirement: '',
        reviewTitle: '',
        reviewRules: []
      })
      toast.info(error.message || '信用申诉配置加载失败')
    }
  },

  handleReasonTap(event) {
    const { key } = event.currentTarget.dataset

    if (key && key !== this.data.activeReason) {
      this.setData({
        activeReason: key,
        reasonOptions: this.buildReasonOptions(this.data.reasonOptions, key)
      })
    }
  },

  handleRecordTap() {
    toast.developing('处罚记录详情待接入信用明细接口')
  },

  handleContentInput(event) {
    const value = event.detail.value || ''

    this.setData({
      appealContent: value,
      appealContentLength: value.length,
      submitClass: value.trim() ? 'ready' : 'disabled'
    })
  },

  handleUploadTap() {
    toast.developing('证明材料上传待接入文件接口')
  },

  async handleSubmitTap() {
    if (!this.data.appealContent.trim()) {
      toast.info('请先填写详细说明')
      return
    }

    try {
      await profileService.submitCreditAppeal({
        recordId: this.data.recordId,
        reason: this.data.activeReason,
        content: this.data.appealContent
      })
      toast.success('申诉已提交')
    } catch (error) {
      toast.info(error.message || '信用申诉提交失败')
    }
  },

  buildReasonOptions(list, activeReason) {
    return this.normalizeList(list).map((item) => ({
      ...item,
      key: item.key || item.reason || item.id,
      className: [
        'reason-pill',
        (item.key || item.reason || item.id) === activeReason ? 'active' : '',
        (item.key || item.reason || item.id) === 'other' ? 'compact' : ''
      ].filter(Boolean).join(' ')
    }))
  },

  normalizeList(list) {
    return Array.isArray(list) ? list : []
  }
})

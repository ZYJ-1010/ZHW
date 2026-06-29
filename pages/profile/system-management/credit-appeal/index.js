const toast = require('../../../../utils/toast')

const LOCAL_ASSET_BASE = '/pages/profile/system-management/credit-appeal/assets'
const BLOCK_ASSET_BASE = '/pages/profile/system-management/block-settings/assets'

Page({
  data: {
    activeReason: 'misjudge',
    appealContent: '',
    appealContentLength: 0,
    submitClass: 'disabled',
    icons: {
      ban: `${BLOCK_ASSET_BASE}/icon-ban.svg`,
      chevron: `${BLOCK_ASSET_BASE}/icon-chevron-right.svg`,
      plus: `${BLOCK_ASSET_BASE}/icon-plus.svg`,
      clock: `${LOCAL_ASSET_BASE}/icon-clock.svg`
    },
    reasonOptions: [
      // Static fallback only; formal options should come from backend config.
      { key: 'misjudge', label: '误判扣分', className: 'reason-pill active' },
      { key: 'system', label: '系统错误', className: 'reason-pill' },
      { key: 'special', label: '特殊情况', className: 'reason-pill' },
      { key: 'other', label: '其他', className: 'reason-pill compact' }
    ],
    relatedRecord: {
      title: '管理员处罚 -10分',
      desc: '违规行为 · 02-28'
    },
    appealPlaceholder: '请详细说明申诉原因，包括但不限于事件经过、时间、涉及人员等信息...',
    uploadRequirement: '支持 JPG、PNG 格式，单张不超过 5MB',
    uploadSlots: [
      { key: 'slot-1' },
      { key: 'slot-2' },
      { key: 'slot-3' },
      { key: 'slot-4' }
    ],
    reviewTitle: '处理时效',
    reviewRules: [
      { key: 'first', text: '提交后24小时内初审' },
      { key: 'review', text: '复杂情况48小时内复核' },
      { key: 'notify', text: '结果将通过站内消息通知' }
    ]
  },

  handleReasonTap(event) {
    const { key } = event.currentTarget.dataset

    if (key && key !== this.data.activeReason) {
      this.setData({
        activeReason: key,
        reasonOptions: this.data.reasonOptions.map((item) => ({
          ...item,
          className: [
            'reason-pill',
            item.key === key ? 'active' : '',
            item.key === 'other' ? 'compact' : ''
          ].filter(Boolean).join(' ')
        }))
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

  handleSubmitTap() {
    if (!this.data.appealContent.trim()) {
      toast.info('请先填写详细说明')
      return
    }

    toast.developing('信用申诉提交接口待接入')
  }
})

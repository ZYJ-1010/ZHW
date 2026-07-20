const toast = require('../../../utils/toast')

const CREATE_PREVIEW_STORAGE_KEY = 'game_create_preview_v1'
const CREATE_PREVIEW_ACTION_KEY = 'game_create_preview_action_v1'
const DEFAULT_GAME_COVER = 'https://static.haowan.net.cn/miniprogram/components/game-card/assets/game-cover-default.png'

function textOrPending(value, pendingText) {
  const text = String(value || '').trim()
  return text || pendingText
}

function splitLines(value) {
  return String(value || '').split(/\n+/).map((item) => item.trim()).filter(Boolean)
}

function buildPreview(draft = {}) {
  const form = draft.form || {}
  const timeDraft = draft.timeDraft || {}
  const signupTimeDraft = draft.signupTimeDraft || {}
  const location = draft.locationInfo || {}
  const media = Array.isArray(draft.descriptionMedia) ? draft.descriptionMedia : []
  const gameTime = draft.gameTimeConfirmed
    ? `${timeDraft.startDate || ''} ${timeDraft.startTime || ''} - ${timeDraft.endDate || ''} ${timeDraft.endTime || ''}`.trim()
    : ''
  const signupTime = draft.signupTimeConfirmed
    ? `${signupTimeDraft.startDate || ''} ${signupTimeDraft.startTime || ''} - ${signupTimeDraft.endDate || ''} ${signupTimeDraft.endTime || ''}`.trim()
    : ''

  return {
    coverImage: textOrPending(draft.coverImage, DEFAULT_GAME_COVER),
    title: textOrPending(form.theme, '未填写局主题'),
    time: textOrPending(gameTime, '局时间待填写'),
    signupTime: textOrPending(signupTime, '报名时间待填写'),
    location: textOrPending(location.name || location.address, '组局地址待填写'),
    type: textOrPending(draft.typeText || form.type, '局类型待选择'),
    fee: textOrPending(draft.feeTypeText || form.feeType, '收费类型待选择'),
    participation: textOrPending(draft.participationText || form.participation, '参与方式待选择'),
    capacity: Number(form.capacity) > 0 ? `${Number(form.capacity)} 人` : '人数待填写',
    tags: Array.isArray(draft.tagLabels) ? draft.tagLabels : [],
    introduction: textOrPending(form.intro, '局介绍待填写'),
    highlights: splitLines(form.highlights),
    description: textOrPending(form.description, '局详情待填写'),
    notice: textOrPending(form.notice, '局须知待填写'),
    audience: textOrPending(form.audience, '受众群体待填写'),
    completionRules: Array.isArray(draft.completionRuleLabels) ? draft.completionRuleLabels : [],
    images: media.filter((item) => item && item.type === 'image' && item.tempFilePath).map((item) => item.tempFilePath),
    videoCount: media.filter((item) => item && item.type === 'video').length
  }
}

Page({
  data: {
    preview: null
  },

  onLoad() {
    try {
      const draft = wx.getStorageSync(CREATE_PREVIEW_STORAGE_KEY)
      if (!draft || !draft.form) {
        toast.info('预览数据已失效，请返回重新打开')
        return
      }
      this.setData({ preview: buildPreview(draft) })
    } catch (error) {
      toast.info('预览数据读取失败，请返回重新打开')
    }
  },

  onBack() {
    wx.navigateBack({ delta: 1 })
  },

  publish() {
    try {
      wx.setStorageSync(CREATE_PREVIEW_ACTION_KEY, { type: 'publish', createdAt: Date.now() })
      wx.navigateBack({ delta: 1 })
    } catch (error) {
      toast.info('发布准备失败，请返回创建页直接发布')
    }
  }
})

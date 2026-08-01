const gameService = require('../../../../../services/game')
const { toUserMessage } = require('../../../../../utils/user-message')

const ENTRY_TYPE_MAP = {
  link: 'link',
  qrcode: 'qrcode',
  poster: 'poster',
  share_card: 'link',
  q: 'qrcode',
  p: 'poster',
  l: 'link'
}

const MODE_CONFIG = {
  link: {
    title: '分享邀请码',
    subtitle: '好友打开分享卡片后将通过你的邀请完成注册',
    actionText: '分享给微信好友'
  },
  qrcode: {
    title: '分享邀请二维码',
    subtitle: '好友扫码后将通过你的邀请完成注册',
    actionText: '分享给微信好友'
  },
  poster: {
    title: '生成邀请海报',
    subtitle: '把邀请码发给好友，好友可通过邀请入口完成注册',
    actionText: '分享海报给好友'
  }
}

function normalizeEntryType(value) {
  const key = String(value || '').trim().toLowerCase()
  return ENTRY_TYPE_MAP[key] || 'link'
}

function modeConfig(entryType) {
  return MODE_CONFIG[entryType] || MODE_CONFIG.link
}

Page({
  data: {
    loading: true,
    inviteCode: '',
    entryType: 'link',
    title: MODE_CONFIG.link.title,
    subtitle: MODE_CONFIG.link.subtitle,
    actionText: MODE_CONFIG.link.actionText,
    qrcodeUrl: '',
    urlLink: '',
    urlLinkReady: false,
    errorText: ''
  },

  onLoad(options = {}) {
    const entryType = normalizeEntryType(options.entryType || options.mode || '')
    const config = modeConfig(entryType)

    wx.setNavigationBarTitle({ title: config.title })
    this.setData({
      entryType,
      inviteCode: '',
      title: config.title,
      subtitle: config.subtitle,
      actionText: config.actionText
    })
    this.loadInviteMaterial(entryType)
  },

  async loadInviteMaterial(entryType = this.data.entryType) {
    try {
      const result = await gameService.createInviteEntry({ entryType })
      const inviteCode = result.inviteCode || ''
      const path = result.path || `/pages/login/invite/index?inviteCode=${encodeURIComponent(inviteCode)}&entryType=${encodeURIComponent(entryType)}`
      const qrcodeError = entryType === 'qrcode' && !result.wxaCodeDataUrl ? '小程序码生成失败，可先使用分享按钮发送邀请码' : ''
      this.setData({
        loading: false,
        inviteCode,
        qrcodeUrl: result.wxaCodeDataUrl || '',
        urlLink: result.urlLink || path,
        urlLinkReady: result.urlLinkReady === true,
        errorText: qrcodeError || (result.urlLinkReady === true ? '' : '邀请链接将在小程序正式发布后开放；当前可使用微信转发或二维码邀请好友。')
      })
    } catch (error) {
      this.setData({
        loading: false,
        inviteCode: '',
        urlLink: '',
        urlLinkReady: false,
        errorText: toUserMessage(error && error.message, '邀请物料生成失败')
      })
    }
  },

  saveQrcode() {
    if (!this.data.qrcodeUrl) return
    wx.getImageInfo({
      src: this.data.qrcodeUrl,
      success: ({ path }) => wx.saveImageToPhotosAlbum({
        filePath: path,
        success: () => wx.showToast({ title: '已保存到相册' }),
        fail: () => wx.showToast({ title: '保存失败，请授权相册权限', icon: 'none' })
      }),
      fail: () => wx.showToast({ title: '二维码读取失败', icon: 'none' })
    })
  },

  copyInvitePath() {
    if (!this.data.urlLinkReady) {
      wx.showToast({ title: '邀请链接将在小程序正式发布后开放', icon: 'none' })
      return
    }
    const text = this.data.urlLink || `/pages/login/invite/index?inviteCode=${encodeURIComponent(this.data.inviteCode)}&entryType=${encodeURIComponent(this.data.entryType || 'link')}`
    if (!this.data.inviteCode) {
      wx.showToast({ title: '邀请码生成中', icon: 'none' })
      return
    }
    wx.setClipboardData({
      data: text,
      success: () => wx.showToast({ title: '已复制邀请入口' })
    })
  },

  onShareAppMessage() {
    const entryType = this.data.entryType || 'link'
    return {
      title: '邀请你加入真好玩',
      path: `/pages/login/invite/index?inviteCode=${encodeURIComponent(this.data.inviteCode)}&entryType=${encodeURIComponent(entryType)}`
    }
  }
})

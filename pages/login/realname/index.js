const toast = require('../../../utils/toast')
const userService = require('../../../services/user')
const { navigateShellRoute } = require('../../../utils/shell-nav')
const { ROUTES } = require('../../../config/routes')

function isValidRealname(realname) {
  return /^[\u4e00-\u9fa5A-Za-z·\s]{2,20}$/.test(String(realname || '').trim())
}

function isValidIdNumber(idNumber) {
  return /(^\d{15}$)|(^\d{17}[\dX]$)/.test(String(idNumber || '').trim().toUpperCase())
}

Page({
  data: {
    realname: '',
    idNumber: '',
    isUiPreview: false,
    isCompleting: false
  },

  onLoad(options = {}) {
    this.setData({
      isUiPreview: options.ui === '1'
    })
  },

  onRealnameInput(event) {
    this.setData({
      realname: String(event.detail.value || '').trim()
    })
  },

  onIdNumberInput(event) {
    this.setData({
      idNumber: String(event.detail.value || '').trim().toUpperCase()
    })
  },

  showAuthFailModal() {
    wx.showModal({
      title: '认证失败',
      content: '认证失败，请重新核对后填写',
      showCancel: true,
      cancelText: '稍后再试',
      confirmText: '继续认证',
      success(result) {
        if (result.cancel) {
          navigateShellRoute('/pages/login/index')
        }
      }
    })
  },

  async completeRealname() {
    if (this.data.isCompleting) {
      return
    }

    if (this.data.isUiPreview) {
      wx.showModal({
        title: '注意事项',
        content: '请按真实身份信息完成认证',
        showCancel: false,
        confirmText: '我知道了',
        success() {
          navigateShellRoute('/pages/login/index?ui=1&mode=newbieTasks')
        }
      })
      return
    }

    if (!isValidRealname(this.data.realname) || !isValidIdNumber(this.data.idNumber)) {
      this.showAuthFailModal()
      return
    }

    this.setData({
      isCompleting: true
    })

    try {
      await userService.submitRealnameAuth({
        realName: this.data.realname,
        idCard: this.data.idNumber
      })

      wx.reLaunch({
        url: `/${ROUTES.playerHome}`
      })
    } catch (error) {
      this.showAuthFailModal()
    } finally {
      this.setData({
        isCompleting: false
      })
    }
  }
})

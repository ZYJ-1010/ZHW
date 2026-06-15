const toast = require('../../../utils/toast')
const userService = require('../../../services/user')

const TEST_REALNAME = '测试用户'
const TEST_ID_NUMBER = '110101199003070011'

function isValidRealname(realname) {
  return /^[\u4e00-\u9fa5A-Za-z·\s]{2,20}$/.test(String(realname || '').trim())
}

function isValidIdNumber(idNumber) {
  return /(^\d{15}$)|(^\d{17}[\dX]$)/.test(String(idNumber || '').trim().toUpperCase())
}

Page({
  data: {
    realname: TEST_REALNAME,
    idNumber: TEST_ID_NUMBER,
    isCompleting: false
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
          toast.developing()
        }
      }
    })
  },

  async completeRealname() {
    if (this.data.isCompleting) {
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
        realname: this.data.realname,
        idNumber: this.data.idNumber
      })

      wx.redirectTo({
        url: '/pages/login/index?ui=1&mode=newbieTasks'
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

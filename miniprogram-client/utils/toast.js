const { toUserMessage } = require('./user-message')

function success(title) {
  wx.showToast({
    title: toUserMessage(title, '操作成功'),
    icon: 'success'
  })
}

function info(title) {
  wx.showToast({
    title: toUserMessage(title),
    icon: 'none'
  })
}

function developing(content = '功能正在开发中') {
  wx.showModal({
    title: '提示',
    content,
    showCancel: false,
    confirmText: '知道了'
  })
}

module.exports = {
  success,
  info,
  developing
}

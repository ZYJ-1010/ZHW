function success(title) {
  wx.showToast({
    title,
    icon: 'success'
  })
}

function info(title) {
  wx.showToast({
    title,
    icon: 'none'
  })
}

module.exports = {
  success,
  info
}

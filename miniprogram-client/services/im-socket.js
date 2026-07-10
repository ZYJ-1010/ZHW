const env = require('../config/env')

function socketBaseUrl() {
  if (env.isMock) {
    return ''
  }
  if (env.socketBaseUrl) {
    return env.socketBaseUrl
  }
  if (!env.baseUrl) {
    return ''
  }
  return env.baseUrl.replace(/^http:\/\//, 'ws://').replace(/^https:\/\//, 'wss://')
}

function socketURL(gameId) {
  const baseUrl = socketBaseUrl()
  if (!baseUrl) {
    return ''
  }
  return `${baseUrl}/api/app/im/ws?gameId=${encodeURIComponent(gameId)}`
}

function connectGameSocket(gameId, handlers = {}) {
  const url = socketURL(gameId)
  const token = typeof wx !== 'undefined' && typeof wx.getStorageSync === 'function'
    ? wx.getStorageSync('enjoy_token')
    : ''

  if (!url || !token || typeof wx === 'undefined' || typeof wx.connectSocket !== 'function') {
    return null
  }

  let opened = false
  const task = wx.connectSocket({
    url,
    header: {
      Authorization: `Bearer ${token}`
    }
  })

  task.onOpen(() => {
    opened = true
    if (typeof handlers.onOpen === 'function') {
      handlers.onOpen()
    }
  })

  task.onMessage((event) => {
    let payload = event.data
    try {
      payload = typeof event.data === 'string' ? JSON.parse(event.data) : event.data
    } catch (error) {
      return
    }

    if (!payload || typeof payload !== 'object') {
      return
    }

    if (payload.type === 'message' && typeof handlers.onMessage === 'function') {
      handlers.onMessage(payload.data)
      return
    }

    if (payload.type === 'connected' && typeof handlers.onConnected === 'function') {
      handlers.onConnected(payload.data)
      return
    }

    if (payload.type === 'error' && typeof handlers.onError === 'function') {
      handlers.onError(payload)
    }
  })

  task.onClose((event) => {
    opened = false
    if (typeof handlers.onClose === 'function') {
      handlers.onClose(event)
    }
  })

  task.onError((event) => {
    opened = false
    if (typeof handlers.onError === 'function') {
      handlers.onError(event)
    }
  })

  return {
    isOpen() {
      return opened
    },
    sendMessage(message) {
      return new Promise((resolve, reject) => {
        if (!opened) {
          reject(new Error('socket not connected'))
          return
        }
        task.send({
          data: JSON.stringify({
            type: 'send_message',
            payload: message
          }),
          success: resolve,
          fail: reject
        })
      })
    },
    close() {
      if (opened) {
        task.close({})
      }
      opened = false
    }
  }
}

module.exports = {
  connectGameSocket
}

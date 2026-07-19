const env = require('../config/env')
const { getAuthToken } = require('../utils/auth-session')

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
  const token = getAuthToken()

  if (!url || !token || typeof wx === 'undefined' || typeof wx.connectSocket !== 'function') {
    return null
  }

  let opened = false
  let requestSequence = 0
  const pendingRequests = new Map()

  function rejectPending(error) {
    pendingRequests.forEach((pending) => {
      clearTimeout(pending.timer)
      pending.reject(error)
    })
    pendingRequests.clear()
  }

  function settleRequest(payload) {
    const requestId = payload && payload.requestId
    if (!requestId || !pendingRequests.has(requestId)) {
      return
    }
    const pending = pendingRequests.get(requestId)
    pendingRequests.delete(requestId)
    clearTimeout(pending.timer)
    if (payload.type === 'error') {
      const error = new Error(payload.message || 'IM 消息发送失败')
      error.code = payload.code
      pending.reject(error)
      return
    }
    pending.resolve(payload.data)
  }
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

    settleRequest(payload)

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
    rejectPending(new Error('IM 连接已断开'))
    if (typeof handlers.onClose === 'function') {
      handlers.onClose(event)
    }
  })

  task.onError((event) => {
    opened = false
    rejectPending(new Error('IM 连接失败'))
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
        requestSequence += 1
        const requestId = `im-${Date.now()}-${requestSequence}`
        const timer = setTimeout(() => {
          pendingRequests.delete(requestId)
          reject(new Error('IM 消息发送超时'))
        }, 10000)
        pendingRequests.set(requestId, { resolve, reject, timer })
        task.send({
          data: JSON.stringify({
            type: 'send_message',
            requestId,
            payload: message
          }),
          success: () => {},
          fail: (error) => {
            pendingRequests.delete(requestId)
            clearTimeout(timer)
            reject(error)
          }
        })
      })
    },
    close() {
      rejectPending(new Error('IM 连接已关闭'))
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

const env = require('../config/env')

const LEVEL_WEIGHT = {
  debug: 10,
  info: 20,
  warn: 30,
  error: 40,
  silent: 100
}

const MAX_DEPTH = 4
const MAX_KEYS = 20
const MAX_ARRAY_LENGTH = 20
const timers = new Map()

let currentLevel = normalizeLevel(env.currentLogLevel)

const SENSITIVE_KEYS = new Set([
  'token',
  'authorization',
  'phone',
  'phonenumber',
  'mobile',
  'mobilephone',
  'password',
  'passwd',
  'pwd',
  'pass',
  'verifycode',
  'smscode',
  'captcha',
  'idcard',
  'idnumber',
  'idno',
  'identityno',
  'invite',
  'invitecode',
  'invitecontext',
  'inviterelation',
  'invitername',
  'inviteruserid',
  'scene',
  'encrypteddata',
  'iv'
])

function normalizeLevel(level) {
  const normalized = String(level || '').toLowerCase()

  if (Object.prototype.hasOwnProperty.call(LEVEL_WEIGHT, normalized)) {
    return normalized
  }

  return env.currentEnv === env.ENV.PROD ? 'warn' : 'debug'
}

function setLevel(level) {
  currentLevel = normalizeLevel(level)
}

function getLevel() {
  return currentLevel
}

function shouldLog(level) {
  if (currentLevel === 'silent') {
    return false
  }

  return (LEVEL_WEIGHT[level] || LEVEL_WEIGHT.error) >= (LEVEL_WEIGHT[currentLevel] || LEVEL_WEIGHT.error)
}

function isPlainObject(value) {
  if (!value || Object.prototype.toString.call(value) !== '[object Object]') {
    return false
  }

  const prototype = Object.getPrototypeOf(value)
  return prototype === Object.prototype || prototype === null
}

function normalizeKeyName(key) {
  return String(key || '').replace(/[_-]/g, '').toLowerCase()
}

function isSensitiveKey(key) {
  return SENSITIVE_KEYS.has(normalizeKeyName(key))
}

function maskPhone(value) {
  const text = String(value == null ? '' : value)

  if (text.length < 7) {
    return '***'
  }

  return `${text.slice(0, 3)}****${text.slice(-4)}`
}

function maskIdCard(value) {
  const text = String(value == null ? '' : value)

  if (text.length < 10) {
    return '***'
  }

  return `${text.slice(0, 6)}********${text.slice(-4)}`
}

function maskToken(value) {
  const text = String(value == null ? '' : value)

  if (text.length <= 8) {
    return '***'
  }

  return `${text.slice(0, 4)}...${text.slice(-4)}`
}

function maskValueByKey(key, value) {
  const normalizedKey = normalizeKeyName(key)

  if (normalizedKey === 'phone' || normalizedKey === 'phonenumber' || normalizedKey === 'mobile' || normalizedKey === 'mobilephone') {
    return maskPhone(value)
  }

  if (normalizedKey === 'idcard' || normalizedKey === 'idnumber' || normalizedKey === 'idno' || normalizedKey === 'identityno') {
    return maskIdCard(value)
  }

  if (normalizedKey === 'token' || normalizedKey === 'authorization' || normalizedKey === 'encrypteddata') {
    return maskToken(value)
  }

  return '***'
}

function serializeError(error) {
  const detail = {
    name: error.name || 'Error',
    message: error.message || '',
    stack: error.stack || ''
  }

  if (error.code !== undefined) {
    detail.code = error.code
  }

  if (error.errMsg !== undefined) {
    detail.errMsg = error.errMsg
  }

  if (error.cause !== undefined) {
    detail.cause = sanitize(error.cause)
  }

  return detail
}

function sanitize(value, seen = new WeakSet(), depth = 0, path = []) {
  if (value === null || value === undefined) {
    return value
  }

  const currentKey = path[path.length - 1]
  const parentPath = path.slice(0, -1)
  const hasSensitiveAncestor = parentPath.some(isSensitiveKey)

  if (hasSensitiveAncestor) {
    return '***'
  }

  if (isSensitiveKey(currentKey)) {
    return maskValueByKey(currentKey, value)
  }

  const valueType = typeof value

  if (valueType === 'string' || valueType === 'number' || valueType === 'boolean' || valueType === 'bigint') {
    return value
  }

  if (valueType === 'symbol') {
    return value.toString()
  }

  if (valueType === 'function') {
    return `[Function ${value.name || 'anonymous'}]`
  }

  if (value instanceof Date) {
    return value.toISOString()
  }

  if (value instanceof Error) {
    return serializeError(value)
  }

  if (seen.has(value)) {
    return '[Circular]'
  }

  if (depth >= MAX_DEPTH) {
    return Array.isArray(value) ? `[Array(${value.length})]` : '[Object]'
  }

  seen.add(value)

  if (Array.isArray(value)) {
    const result = []
    const limit = Math.min(value.length, MAX_ARRAY_LENGTH)

    for (let index = 0; index < limit; index += 1) {
      result.push(sanitize(value[index], seen, depth + 1, path.concat(String(index))))
    }

    if (value.length > MAX_ARRAY_LENGTH) {
      result.push(`...${value.length - MAX_ARRAY_LENGTH} more`)
    }

    return result
  }

  if (value instanceof Map) {
    const result = {}
    let count = 0

    value.forEach((mapValue, mapKey) => {
      if (count >= MAX_KEYS) {
        return
      }

      result[String(mapKey)] = sanitize(mapValue, seen, depth + 1, path.concat(String(mapKey)))
      count += 1
    })

    if (value.size > MAX_KEYS) {
      result.__truncated__ = value.size - MAX_KEYS
    }

    return result
  }

  if (value instanceof Set) {
    return sanitize(Array.from(value), seen, depth, path)
  }

  const keys = Object.keys(value)
  const result = {}
  const limit = Math.min(keys.length, MAX_KEYS)

  for (let index = 0; index < limit; index += 1) {
    const key = keys[index]
    result[key] = sanitize(value[key], seen, depth + 1, path.concat(key))
  }

  if (keys.length > MAX_KEYS) {
    result.__truncated__ = keys.length - MAX_KEYS
  }

  return result
}

function formatTime(date = new Date()) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  const milliseconds = String(date.getMilliseconds()).padStart(3, '0')

  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}.${milliseconds}`
}

function buildPrefix(level, scope) {
  const parts = [`${formatTime()}`, level.toUpperCase()]

  if (scope) {
    parts.push(scope)
  }

  return `[${parts.join('][')}]`
}

function getConsoleMethod(level) {
  if (typeof console === 'undefined') {
    return null
  }

  const method = console[level] || console.log
  return typeof method === 'function' ? method : null
}

function normalizeMessage(message, detail) {
  if (detail === undefined && (message instanceof Error || Array.isArray(message) || isPlainObject(message))) {
    return {
      message: '',
      detail: message
    }
  }

  if (message instanceof Error) {
    return {
      message: message.message || 'Error',
      detail: message
    }
  }

  return {
    message: message === undefined || message === null ? '' : String(message),
    detail
  }
}

function emit(level, scope, message, detail) {
  if (!shouldLog(level)) {
    return
  }

  const method = getConsoleMethod(level)

  if (!method) {
    return
  }

  const normalized = normalizeMessage(message, detail)
  const prefix = buildPrefix(level, scope)

  if (normalized.detail === undefined) {
    method.call(console, prefix, normalized.message)
    return
  }

  if (normalized.message) {
    method.call(console, prefix, normalized.message, sanitize(normalized.detail))
    return
  }

  method.call(console, prefix, sanitize(normalized.detail))
}

function createTimerKey(scope, label) {
  return `${scope || 'root'}::${String(label)}`
}

function createLogger(scope = '') {
  return {
    debug(message, detail) {
      emit('debug', scope, message, detail)
    },

    info(message, detail) {
      emit('info', scope, message, detail)
    },

    warn(message, detail) {
      emit('warn', scope, message, detail)
    },

    error(message, detail) {
      emit('error', scope, message, detail)
    },

    time(label, detail) {
      timers.set(createTimerKey(scope, label), Date.now())

      if (detail !== undefined) {
        emit('debug', scope, `${label} start`, detail)
      }
    },

    timeEnd(label, message, detail) {
      const timerKey = createTimerKey(scope, label)
      const startedAt = timers.get(timerKey)

      if (!startedAt) {
        emit('warn', scope, `${label} timer missing`)
        return 0
      }

      timers.delete(timerKey)

      const duration = Date.now() - startedAt
      const payload = isPlainObject(detail)
        ? Object.assign({}, detail, { duration })
        : detail === undefined
          ? { duration }
          : { detail, duration }

      emit('info', scope, message || label, payload)
      return duration
    },

    measure(label, fn, detail) {
      const startedAt = Date.now()

      try {
        const result = fn()

        if (result && typeof result.then === 'function') {
          return result.then((value) => {
            emit('info', scope, label, Object.assign({}, isPlainObject(detail) ? detail : {}, { duration: Date.now() - startedAt }))
            return value
          }).catch((error) => {
            emit('error', scope, label, Object.assign({}, isPlainObject(detail) ? detail : {}, { duration: Date.now() - startedAt, error }))
            throw error
          })
        }

        emit('info', scope, label, Object.assign({}, isPlainObject(detail) ? detail : {}, { duration: Date.now() - startedAt }))
        return result
      } catch (error) {
        emit('error', scope, label, Object.assign({}, isPlainObject(detail) ? detail : {}, { duration: Date.now() - startedAt, error }))
        throw error
      }
    }
  }
}

const baseLogger = createLogger()

module.exports = Object.assign(baseLogger, {
  createLogger,
  setLevel,
  getLevel,
  sanitize,
  normalizeLevel
})

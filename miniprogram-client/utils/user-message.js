function hasChinese(text) {
  return /[\u4e00-\u9fff]/.test(String(text || ''))
}

function toUserMessage(message, fallback = '操作失败，请稍后重试') {
  const text = String(message || '').trim()

  if (!text || hasChinese(text)) {
    return text || fallback
  }

  const normalized = text.toLowerCase()
  const messageMap = [
    [/realname|strong identity|identity.*verified/, '请先完成实名认证后再操作'],
    [/invite.*(expired|invalid|already bound)|invalid.*invite/, '邀请码无效、已失效或已被使用，请重新获取邀请码'],
    [/invalid.*phone|phone.*invalid/, '手机号格式不正确'],
    [/invalid.*(sms|code)|(sms|code).*invalid/, '短信验证码错误或已失效'],
    [/invalid.*id.*card|id.*card.*invalid/, '身份证号格式不正确'],
    [/invalid request/, '提交信息有误，请检查后重试'],
    [/game.*full|already full|capacity.*full/, '本局已满员，暂不能继续申请'],
    [/permission denied|forbidden|not allowed/, '当前账号没有此操作权限'],
    [/unauthorized|token.*(expired|invalid)|login.*required|authentication/, '登录状态已失效，请重新登录'],
    [/sms.*(rate|daily|frequent|too many)|too many requests|\b429\b/, '操作过于频繁，请稍后再试'],
    [/network|timeout|request.*fail|connection.*(fail|closed)|socket.*(connect|closed|disconnect)/, '网络连接异常，请检查网络后重试'],
    [/not found|does not exist/, '相关内容不存在或已被删除'],
    [/file.*(invalid|fail)|upload.*fail/, '文件处理失败，请重新选择后重试'],
    [/wechat|wx\.login/, '微信授权暂时不可用，请稍后重试']
  ]
  const matched = messageMap.find(([pattern]) => pattern.test(normalized))

  return matched ? matched[1] : fallback
}

module.exports = {
  toUserMessage
}

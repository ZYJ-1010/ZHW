const fs = require('fs')
const path = require('path')

const root = path.resolve(__dirname, '..')
const checklistPath = path.join(root, '文档', '功能清单-汇总版.md')

const updates = new Map([
  ['局审核', '✔ 本机前后端深链路通过：创建局后后台审核通过，前端局状态可继续流转，操作日志接口可查。'],
  ['入局申请', '✔ 本机前后端深链路通过：玩家申请、重复申请拦截、组局者审核通过、成员表落库已验收。'],
  ['成团与开始', '✔ 本机前后端深链路通过：少于 5 人禁止手动开始，5 人可手动开始，8 人自动满员锁定并阻止继续入局。'],
  ['局内 IM', '✔ 本机前后端深链路通过：成员进入房间、发送文字、发送文件、非成员禁止进入、后台查看消息和文件数已验收。'],
  ['文件传输', '✔ 本机前后端深链路通过：chat_file 上传凭证、fileId 回传、文件消息发送、后台 IM 文件统计已验收；真实对象存储联调后置。'],
  ['评价体系', '✔ 本机前后端深链路通过：服务确认后生成评价待办，成员互评、低分评价、再玩意向沉淀已验收。'],
  ['信用与成长', '✔ 本机前后端深链路通过：低分评价扣信用、信用中心回显、信用申诉、后台处理申诉并恢复信用已验收。'],
  ['收益与分润', '✔ 本机前后端深链路通过：本地收益模板 seed、评价完成后生成收益记录、个人收益汇总回显已验收；真实支付/分账联调后置。'],
  ['举报与申诉', '✔ 本机前后端深链路通过：举报提交、聊天证据绑定、后台批量处理、通知回显已验收。'],
  ['个人中心', '✔ 本机前后端深链路通过：个人主页、资产、积分、兑换商品、订单列表、订单详情、取消订单和取消后列表状态回显已验收。']
])

const text = fs.readFileSync(checklistPath, 'utf8')
const lines = text.split(/\r?\n/)
let updated = 0

const next = lines.map((line) => {
  if (!line.startsWith('| ') || line.startsWith('| ---')) {
    return line
  }
  const cells = line.split('|')
  if (cells.length < 6) {
    return line
  }
  const moduleName = cells[1].trim()
  if (!updates.has(moduleName)) {
    return line
  }
  cells[5] = ` ${updates.get(moduleName)} `
  updated += 1
  return cells.join('|')
})

if (updated !== updates.size) {
  throw new Error(`expected ${updates.size} updates, wrote ${updated}`)
}

fs.writeFileSync(checklistPath, next.join('\n'), 'utf8')
console.log(`updated=${updated}`)

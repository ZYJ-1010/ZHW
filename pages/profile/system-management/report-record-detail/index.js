const toast = require('../../../../utils/toast')

Page({
  data: {
    icons: {
      check: '/pages/profile/system-management/report-center/assets/icon-check.svg'
    },
    basicInfo: [
      { label: '举报编号', value: 'RP20240612001' },
      { label: '举报类型', value: '诱导私下交易' },
      { label: '被举报人', value: '用户A (ID: 00527)' },
      { label: '举报时间', value: '2024-06-12 10:20' },
      { label: '完成时间', value: '2024-06-12 16:45' },
      { label: '处理人', value: '平台审核员_01' }
    ],
    reportReason: '该用户在组局过程中多次诱导我通过微信进行私下交易，绕过平台支付系统，并承诺给予额外优惠。我已保存相关聊天记录作为证据。',
    evidence: [
      { label: '聊天记录1', index: '图1', tone: 'teal' },
      { label: '聊天记录2', index: '图2', tone: 'red' },
      { label: '转账截图', index: '图3', tone: 'purple' }
    ],
    resultNote: '经平台审核，被举报人确实存在诱导用户进行私下交易的行为，违反了平台《交易行为规范》第3.2条。已对其采取相应处罚措施，感谢您的监督与举报。',
    timeline: [
      { time: '2024-06-12 10:20', title: '提交举报', desc: '您提交了举报申请，等待平台审核' },
      { time: '2024-06-12 11:30', title: '平台受理', desc: '平台已受理您的举报，开始核实调查' },
      { time: '2024-06-12 14:00', title: '证据核实', desc: '平台审核员已核实您提供的证据材料' },
      { time: '2024-06-12 16:45', title: '处理完成', desc: '举报属实，已对被举报人进行处罚，奖励已发放', state: 'current' }
    ]
  },

  handleBackList() {
    wx.redirectTo({
      url: '/pages/profile/system-management/report-records/index'
    })
  },

  handleRateTap() {
    toast.developing('处理评价待接入评价接口')
  }
})

const toast = require('../../../../utils/toast')

Page({
  data: {
    basicInfo: [
      { label: '申诉编号', value: 'AP20240608003' },
      { label: '被举报类型', value: '言语骚扰' },
      { label: '举报人', value: '玩家_06688' },
      { label: '申诉时间', value: '2024-06-08 09:15' },
      { label: '当前状态', value: '⟳ 处理中', className: 'appeal-status-value' }
    ],
    appealReason: '在服务过程中，我与玩家进行了正常的业务沟通，所有对话内容均可在平台聊天记录中查证，不存在任何骚扰或不当言论。对方因为服务结果不满而恶意举报，请平台核实。',
    evidence: [
      { label: '完整聊天记录', index: '图1', tone: 'teal' },
      { label: '服务完成凭证', index: '图2', tone: 'purple' }
    ],
    originalInfo: [
      { label: '举报编号', value: 'RP20240607015' },
      { label: '举报时间', value: '2024-06-07 20:30' }
    ],
    originalReason: '该行家在服务过程中多次发送不当言论，对我进行言语骚扰，严重影响服务体验。',
    timeline: [
      { time: '2024-06-07 20:30', title: '收到举报', desc: '平台收到玩家_06688的举报申请' },
      { time: '2024-06-07 22:00', title: '初步核实', desc: '平台审核员初步核实举报内容' },
      { time: '2024-06-08 09:15', title: '提交申诉', desc: '您提交了申诉申请及相关证据材料' },
      { time: '2024-06-08 10:00', title: '申诉审核中', desc: '平台正在审核您的申诉材料，预计1-3个工作日完成', state: 'current' },
      { time: '待定', title: '处理结果', desc: '等待最终审核结果', state: 'pending' }
    ]
  },

  handleBackList() {
    wx.redirectTo({
      url: '/pages/profile/system-management/report-appeals/index'
    })
  },

  handleWithdrawTap() {
    toast.developing('撤回申诉待接入申诉接口')
  }
})

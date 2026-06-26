Page({
  data: {
    sections: [
      {
        title: '账户安全',
        rows: [
          { label: '设置支付密码', value: '未设置' },
          { label: '更换手机号', value: '' }
        ]
      },
      {
        title: '通知设置',
        rows: [
          { label: '推送通知', switch: true, enabled: true },
          { label: '邮件通知', switch: true, enabled: false }
        ]
      },
      {
        title: '隐私设置',
        rows: [
          { label: '隐私政策摘要', value: '' },
          { label: '个性化推送', switch: true, enabled: true },
          { label: '第三方共享清单', value: '' },
          { label: '信息收集清单', value: '' }
        ]
      },
      {
        title: '其他',
        rows: [
          { label: '清除缓存', value: '12.6MB' },
          { label: '关于我们', value: '' }
        ]
      }
    ]
  }
})

# 墨刀原型 PT 标注总目录

- 原型：小程序开发（对外）副本
- 画布基准：375 × 812pt
- 已解析页面画布：105 个（与导出包记录的 106 个画布相比，另有一个无页面标识的元节点）
- 完整元素 PT、文字样式、原始填充/边框/阴影、图片和 SVG 引用：`all-screens-pt.json`
- 可复用 SVG 图标：`reusable-svg-icons/`（73 个去重文件）及 `reusable-svg-icons.json`
- 导出包内另有 21 个非演示上传 SVG：`reusable-upload-svg/`；67 个带姓名缩写的演示头像已排除，详见 `uploaded-svg-manifest.json`。
- 图片引用清单：`image-reference-manifest.json`；演示头像、封面和业务数据不复制。

## 使用规则

- 只复刻布局、样式和交互结构；用户、榜单、金额、头像、局内容等演示数据不进入业务前端。
- 页面坐标和尺寸均为 PT；转换至 750rpx 小程序宽度时，数值乘以 2。
- 系统状态栏和设备壳仅为原型环境，不直接实现。

## 页面清单

1. 顶层 / 首页（`sl9w6l9vTtCZ4SBvgRtNS0`；元素 2484）
2. 首页 / 首页调色说明页（`rbpVJ1KqdUofMwkXx`；元素 368）
3. 首页 / 玩家首页（`rbpVJ1UKNmIoWvrPM`；元素 362）
4. 首页 / 行家首页（`rbpVM1dv8dqBpYzrn`；元素 348）
5. 首页 / 领路人首页（`rbpVM1dz7XbQDGWVu`；元素 369）
6. 首页 / 无行家权限提示页（`rbpVM1e9sEPewjWyh`；元素 337）
7. 首页 / 无领路人权限提示页（`rbpVM1eGtDn23KFRI`；元素 336）
8. 首页 / 权益对比页（`rbpVM1eKsOdEa6e6f`；元素 103）
9. 首页 / 申请行家操作页（`rbpVM1eOF1Y9QUqFe`；元素 201）
10. 首页 / 申请领路人操作页（`rbpVM1eRJEVdajcJr`；元素 102）
11. 首页 / 审核进度页（`rbpVM1eTdsXVE9nEC`；元素 70）
12. 首页 / 审核通过页（`rbpVM1eWSSxJPXckk`；元素 144）
13. 首页 / 审核驳回页（`rbpVM1eZIw6iHGkf2`；元素 66）
14. 首页 / 申请进度查看页（`rbpVM7ov9TRG0r9qv`；元素 87）
15. 首页 / 申请状态提示页（`rbpVM85ja9uJ0cuh8`；元素 360）
16. 顶层 / 地图首页（`rbpVFDJKkOR1Pi9Cf`；元素 145）
17. 地图首页 / 组局盲盒（`rbpVFhRyVQ2NHmNBL`；元素 159）
18. 地图首页 / 城市图鉴页（`rbpVFhS48kWjXc2EJ`；元素 168）
19. 地图首页 / 足迹热力图页（`rbpVFhS7WqaJ2x2hW`；元素 164）
20. 地图首页 / 好友点亮城市页（`rbpVFhSBTgayJQGMn`；元素 177）
21. 地图首页 / 我的城市故事页（`rbpVFhSE0oS6AjVr6`；元素 369）
22. 地图首页 / 实景打卡（`rbpVFiON3TujC3TzB`；元素 138）
23. 顶层 / 元宇宙（预留）（`rbpVFceLYGDAsrJh7`；元素 83）
24. 元宇宙（预留） / 元宇宙内页及管理中心（`rbpVFUTHHS6oGHHvp`；元素 240）
25. 顶层 / 消息（`sl9w6sosTtCZ4VRcWeu1gL`；元素 750）
26. 顶层 / 关系网（`rbpVFcuW29xOmYntM`；元素 186）
27. 顶层 / 局前大厅（`rbpVFhgG2Zfb0ZCXF`；元素 708）
28. 顶层 / 组局主流程（`rbpVFdDimM6YDUiyN`；元素 0）
29. 组局主流程 / 玩家自申请入局（`rbpVFi5FJYvefTET2`；元素 93）
30. 组局主流程 / 玩家被邀约确认页（`rbpVFiDQ6F3BQSBzl`；元素 355）
31. 组局主流程 / 组局取消页（`rbpVFiES7ekWvx0Sv`；元素 360）
32. 组局主流程 / 组局成功页（`rbpVFiEbKOtoG9qSi`；元素 645）
33. 组局主流程 / 领路人接收页（`rbpVFiFCEYIV2h96w`；元素 296）
34. 组局主流程 / 行家审核列表页（`rbpVFiFGy5KflgnSU`；元素 109）
35. 组局主流程 / 行家审核详情页（`rbpVFiFEOMVVMkruU`；元素 269）
36. 组局主流程 / 组局支付页（`rbpVFiSE9f9FsEa6n`；元素 90）
37. 组局主流程 / 评价页面（`rbpVG9nC9JtYtDDkB`；元素 1077）
38. 组局主流程 / 评价页面 / 行家评价页面（`rbpVMQanua6yEuivN`；元素 175）
39. 组局主流程 / 评价页面 / 领路人评价页面（`rbpVMQasvm9FK03g3`；元素 178）
40. 组局主流程 / 评价页面 / 玩家评价页面（`rbpVMQaxhZpnPmvQP`；元素 178）
41. 组局主流程 / 评价页面 / 评价+再玩一局意向确认页（`rbpVMQb0gFFZ7cKYQ`；元素 109）
42. 组局主流程 / 评价页面 / 再玩一局推荐页（`rbpVMQbOj0Yrou7RF`；元素 100）
43. 组局主流程 / 评价页面 / 同局好友再玩一局确认页（`rbpVMQbS1o1vmcio5`；元素 121）
44. 组局主流程 / 评价页面 / 系统推荐适配组局页（`rbpVMQbVM5V3C9lJg`；元素 220）
45. 组局主流程 / 交付确认页（`rbpVG9nyvViQNuAgw`；元素 345）
46. 组局主流程 / 领路人发起邀请页（`rbpVGAwrA847B8W4N`；元素 477）
47. 组局主流程 / 交付操作页（`rbpVFiElZPe6aqwj0`；元素 45）
48. 组局主流程 / 局前大厅（`rbpVJ1xyHBI8P3ZK6`；元素 421）
49. 顶层 / 邀请注册（`rbpVKh0hIsrlUvh6E`；元素 0）
50. 邀请注册 / 新用户邀请注册登录（`rbpVLeQ5JmcyvOq5u`；元素 204）
51. 邀请注册 / 已注册用户登录（`rbpVLeQA1W2NFnftA`；元素 185）
52. 邀请注册 / 忘记密码找回（`rbpVLeQAbZ6l5DEXL`；元素 80）
53. 顶层 / 我的（`sl9w79qkTtCZ4Vtka6qDNq`；元素 2197）
54. 我的 / 母版（`rbpVFiQuOriYNNHMj`；元素 124）
55. 我的 / 会员中心（`rbpVG9tHRq5qVkIhh`；元素 270）
56. 我的 / 足迹中心（`rbpVMU5E3ygiFHXaY`；元素 4）
57. 我的 / 足迹中心 / 我的城市故事（`rbpVMU5IQCNhC984V`；元素 4）
58. 我的 / 足迹中心 / 我的成就墙（`rbpVFhSVzsIaBGKCJ`；元素 178）
59. 我的 / 系统管理（`rbpVMU5UyTrd75Zcf`；元素 0）
60. 我的 / 系统管理 / 我的资料（`rbpVMU6C8OyV7bhvu`；元素 132）
61. 我的 / 系统管理 / 信用中心（`rbpVMU6Fl6m63YHQb`；元素 117）
62. 我的 / 系统管理 / 信用中心 / 信用申诉页（`rbpVMXq8S43dKj4d6`；元素 79）
63. 我的 / 系统管理 / 举报中心（`rbpVMU6JYCtR2I8hX`；元素 83）
64. 我的 / 系统管理 / 举报中心 / 我的申诉（`rbpVMY2mnvk7IIIJW`；元素 87）
65. 我的 / 系统管理 / 举报中心 / 查看处理记录页（`rbpVMY2tpPfv7gAZQ`；元素 136）
66. 我的 / 系统管理 / 举报中心 / 举报详情页（`rbpVMY78YLKV8YWlr`；元素 127）
67. 我的 / 系统管理 / 举报中心 / 申诉详情页（`rbpVMY7AtsNY8fZeP`；元素 117）
68. 我的 / 系统管理 / 举报中心 / 处理记录详情页（`rbpVMY7I0CvO79mDQ`；元素 132）
69. 我的 / 系统管理 / 签署协议（`rbpVMU6NBr8VEcb4A`；元素 48）
70. 我的 / 系统管理 / 签署协议 / 协议详情弹窗（`rbpVMXv13u5MUlrr4`；元素 80）
71. 我的 / 系统管理 / 建议反馈（`rbpVMU6Ri5n9ABKcX`；元素 116）
72. 我的 / 系统管理 / 建议反馈 / 反馈记录（`rbpVMXjpUmOEsI6XL`；元素 78）
73. 我的 / 系统管理 / 建议反馈 / 反馈详情页（`rbpVMXjuO4nVzI3LA`；元素 70）
74. 我的 / 系统管理 / 建议反馈 / 快捷反馈弹窗页（`rbpVMXjxOF9BJQpdm`；元素 45）
75. 我的 / 系统管理 / 建议反馈 / 反馈成功页（`rbpVMXk0BnCTlA5i5`；元素 99）
76. 我的 / 系统管理 / 系统设置（`rbpVMU6bjYrmWMsnN`；元素 4）
77. 我的 / 系统管理 / 系统设置 / 系统设置主页（`rbpVMZhoHu6IdAPZA`；元素 115）
78. 我的 / 系统管理 / 技能配置（`rbpVMW4KaaCmhratz`；元素 426）
79. 我的 / 系统管理 / 屏蔽设置（`rbpVMW4ONBLPWSD4M`；元素 119）
80. 我的 / 系统管理 / 屏蔽设置 / 保护模式设置页（`rbpVMWlJiCV1hLtz8`；元素 112）
81. 我的 / 系统管理 / 屏蔽设置 / 分场景配置页（`rbpVMWnaJFWjYvN6O`；元素 154）
82. 我的 / 系统管理 / 屏蔽设置 / 白名单设置页（`rbpVMWpQg8mbB5F84`；元素 163）
83. 我的 / 系统管理 / 屏蔽设置 / 保护期续期页（`rbpVMWqDX0w9FpNiE`；元素 60）
84. 我的 / 系统管理 / 屏蔽设置 / 用户屏蔽（`rbpVMXQAveYUBKCjz`；元素 250）
85. 我的 / 系统管理 / 屏蔽设置 / 关键词屏蔽（`rbpVMXQGKkTEygrFj`；元素 141）
86. 我的 / 服务中心（`rbpVMU5Xr6EdHTb5R`；元素 0）
87. 我的 / 服务中心 / 组局管理（`rbpVMU4NjgyjM03AR`；元素 0）
88. 我的 / 服务中心 / 组局管理 / 评价管理（`rbpVMTtamztvbQ58T`；元素 258）
89. 我的 / 服务中心 / 我的邀请（`rbpVMU4T4W6tXN6Jg`；元素 114）
90. 我的 / 服务中心 / 我的组局（`rbpVMU4LV1EPJVPo0`；元素 106）
91. 我的 / 资产中心（`rbpVMU5kBrAlHZSOl`；元素 0）
92. 我的 / 资产中心 / 资产管理（`rbpVMU4jSNOe0FVRk`；元素 140）
93. 我的 / 资产中心 / 积分商城（`rbpVMU59i2Od5BKsZ`；元素 4）
94. 我的 / 资产中心 / 我的积分（`rbpVMU56G5BlMVxS1`；元素 185）
95. 我的 / 资产中心 / 押金中心（`rbpVMU63drrcdKrXt`；元素 4）
96. 我的 / 资产中心 / 开票中心（`rbpVMU6WBoYq25Lip`；元素 4）
97. 我的 / 个人中心（`rbpVMU8v45uUxpjIE`；元素 298）
98. 顶层 / 发起组局（`rbpVM7tSbhQoEVKSf`；元素 237）
99. 顶层 / 角色申请（`rbpVHu7ovAlnGQ1X9`；元素 1）
100. 顶层 / 企业认证（`rbpVJTnRsWbZ2uucy`；元素 133）
101. 顶层 / 身份认证（`rbpVJTozah2XU4TEG`；元素 1）
102. 顶层 / 画布 3（`rbpVJTlrgXNR21bi`；元素 0）
103. 顶层 / 画布 1（`rbpVJQ64o9qM3Zzsh`；元素 113）
104. 顶层 / 画布 1（`rbpVMQSgWuBG846KJ`；元素 4）
105. 顶层 / 画布 3（`rbpVJTlSPnDauFkXF`；元素 793）

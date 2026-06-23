# Enjoy UI Style Guide

本文记录当前小程序的基础控件规范。第一阶段只覆盖登录链路，首页和强业务模块后续单独统一。

## 适用范围

- 当前适用：登录、邀请注册、找回密码、微信授权、实名认证弹窗、新手任务弹窗中的基础控件。
- 暂不适用：首页、地图、排行榜、成就、组局卡片等强业务视觉模块。
- 组件库暂不引入，先使用项目自有 `enjoy-` 前缀样式。

## 登录链路版式

- 目标画布宽度按 `750rpx` 理解。
- 登录链路主体内容优先使用左右 `62rpx` 安全边距，对应内容宽 `626rpx`。
- 页面标题位置、表单纵向位置、弹窗位置尊重现有原型还原，本阶段不强行统一。
- 控件内部高度、圆角、字号、颜色走全局基础样式。

## 基础控件

- 样式文件：`styles/enjoy-ui.wxss`，由 `app.wxss` 全局引入。
- 输入框：
  - `.enjoy-field`：默认高度 `80rpx`，圆角 `20rpx`，背景 `#f4f4f6`。
  - `.enjoy-field--comfortable`：高度 `88rpx`。
  - `.enjoy-field--large`：高度 `104rpx`，圆角 `28rpx`。
  - `.enjoy-input`、`.enjoy-input--comfortable`、`.enjoy-input--large` 对应输入文字尺寸。
- 按钮：
  - `.enjoy-button`：默认高度 `82rpx`，圆角 `20rpx`。
  - `.enjoy-button--primary`：紫粉渐变 `#8a5cf6 -> #e44f9b`。
  - `.enjoy-button--secondary`：浅灰底、次级文字。
  - `.enjoy-button--success`：成功色 `#22c79a`。
  - `.enjoy-button--disabled` 或 disabled 态：弱灰底、弱文字。
- 验证码：
  - `.enjoy-code-box`：`78rpx x 88rpx`，圆角 `16rpx`。
  - `.enjoy-code-box--filled`：边框使用主紫色。
  - `.enjoy-code-action`：验证码文字入口统一为 `24rpx`、`#8a5cf6`、`700`。
  - `.enjoy-code-action--disabled`：倒计时或不可点状态使用 `#a8a8b2`。
- 勾选：
  - `.enjoy-check`：圆形基础勾选控件。
  - `.enjoy-check--active`：完成/选中态，使用成功色 `#22c79a`。

## 当前约定

- 页面原有布局类可以与 `enjoy-` 类叠加，例如 `class="field enjoy-field enjoy-field--comfortable"`。
- 页面级 WXSS 只保留定位、宽度、间距等版式规则；颜色、圆角、字号、高度优先使用 `enjoy-` 样式。
- 验证码入口在登录链路统一文案为“获取验证码”；倒计时可通过周边文案表达，例如“59s 后可 获取验证码”。
- 后续统一首页时，应先确认原型当前画布，再为首页业务模块单独定义视觉规则。

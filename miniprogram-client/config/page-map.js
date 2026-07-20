const PAGE_GROUPS = [
  {
    code: 'auth',
    title: '邀请注册与登录',
    priority: 'P0',
    batch: '第 1 批：入口闭环',
    route: '/pages/entry/index',
    routes: ['/pages/entry/index', '/pages/home/guest/index', '/pages/login/index', '/pages/login/forgot/index'],
    screens: ['进入页面', '登录前首页', '用户注册登录', '微信授权登录（登录页内部流程）', '忘记密码找回']
  },
  {
    code: 'home',
    title: '首页与角色入口',
    priority: 'P0',
    batch: '第 1 批：入口闭环',
    route: '/pages/home/player/index',
    routes: ['/pages/home/guest/index', '/pages/home/player/index', '/pages/home/expert/index', '/pages/home/guide/index'],
    screens: ['登录后玩家首页', '登录前首页', '行家首页', '领路人首页', '无行家权限提示页', '无领路人权限提示页', '权益对比页', '首页调色说明页']
  },
  {
    code: 'role',
    title: '角色申请与审核状态',
    priority: 'P0',
    batch: '第 2 批：角色与组局基础',
    route: '/pages/role/flow/index?mode=roleComparison&single=1',
    screens: ['申请行家操作页', '申请领路人操作页', '审核进度页', '审核通过页', '审核驳回页', '申请状态提示页', '申请进度查看页']
  },
  {
    code: 'game',
    title: '组局主流程',
    priority: 'P0',
    batch: '第 2-3 批：组局基础与交付闭环',
    route: '/pages/game/hall/index',
    routes: ['/pages/game/hall/index', '/pages/game/create/index', '/pages/game/detail/index', '/pages/game/audit/index', '/pages/game/audit-detail/index', '/pages/game/applications/index', '/pages/game/delivery/index', '/pages/game/review/index'],
    screens: ['局前大厅', '发起组局', '玩家自申请入局', '领路人发起引荐页', '玩家被邀约确认页', '行家审核列表页', '行家审核详情页', '领路人接收页', '组局支付页', '组局成功页', '组局取消页', '交付操作页', '交付确认页', '评价页面']
  },
  {
    code: 'map',
    title: '地图与城市探索',
    priority: 'P0/P1',
    batch: '第 4 批：地图、IM、成长扩展',
    route: '/pages/map/index',
    routes: ['/pages/map/index', '/pages/map/blind-route/index', '/pages/map/city-atlas/index', '/pages/map/real-checkin/index', '/pages/map/footprint-heatmap/index', '/pages/map/friend-city/index', '/pages/map/my-city/index'],
    screens: ['地图首页', '组局分布地球网', '组局盲盒', '城市图鉴页', '足迹热力图页', '实景打卡', '好友点亮城市页', '我的城市故事页']
  },
  {
    code: 'message',
    title: '消息',
    priority: 'P0',
    batch: '第 4 批：地图、IM、成长扩展',
    route: '/pages/message/index',
    routes: ['/pages/message/index', '/pages/message/my/index', '/pages/message/trade-warning/index', '/pages/message/system-detail/index'],
    screens: ['消息', '我的消息', '交易预警', '系统通知']
  },
  {
    code: 'relation',
    title: '关系',
    priority: 'P0',
    batch: '第 4 批：地图、IM、成长扩展',
    route: '/pages/relation/network/index',
    routes: ['/pages/relation/network/index'],
    screens: ['关系网首页']
  },
  {
    code: 'profile',
    title: '我的与成长',
    priority: 'P0/P1',
    batch: '第 1-4 批持续补齐',
    route: '/pages/profile/index',
    routes: ['/pages/profile/index', '/pages/profile/asset-center/manage/index', '/pages/profile/asset-center/mall/index', '/pages/profile/asset-center/orders/index', '/pages/profile/asset-center/points/index', '/pages/profile/system-management/profile-info/index', '/pages/profile/system-management/skill-config/index', '/pages/profile/system-management/service-case-detail/index', '/pages/profile/member/index', '/pages/profile/achievements/index', '/pages/profile/footprint/achievements/index'],
    screens: ['我的', '资产管理', '积分商城', '我的订单', '积分中心', '系统管理-我的资料', '系统管理-技能配置', '系统管理-技能配置-服务案例详情', '会员中心', '成就页', '我的足迹-我的成就墙']
  },
  {
    code: 'reserved',
    title: '预留能力',
    priority: 'P1/P2',
    batch: '预留',
    route: '/pages/placeholder/metaverse/index',
    screens: ['元宇宙（预留）', '元宇宙内页及管理中心', '母版']
  }
]

module.exports = {
  PAGE_GROUPS
}

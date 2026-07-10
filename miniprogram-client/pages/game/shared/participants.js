const DEFAULT_GAME_PARTICIPANTS = [
  {
    id: 'luyi',
    name: '陆毅',
    avatarSrc: 'https://static.haowan.net.cn/miniprogram/pages/home/player/assets/ranking-avatar-01.png',
    avatarText: '陆',
    role: '玩家',
    roleClass: 'player',
    position: '总经理 | 上海创世界科技有限公司',
    topic: 'AI赋能与市场运营助力企业IP打造',
    primaryTag: '第一标签：上海TMT投资领军者',
    tags: ['数字化内容服务'],
    location: '上海市浦东新区沙新镇黄赵路310号',
    distance: '2.1 km'
  },
  {
    id: 'linyi',
    name: '林一',
    avatarSrc: 'https://static.haowan.net.cn/miniprogram/pages/home/player/assets/ranking-avatar-02.png',
    avatarText: '林',
    role: '行家',
    roleClass: 'expert',
    position: '品牌创始人 | 杭州欣悦服装工作',
    topic: '企业家服务平台',
    primaryTag: '第一标签：女性高品质服装领先者',
    tags: ['品牌增长', '企业服务'],
    location: '杭州市上城区',
    distance: '2.1 km'
  }
]

function getDefaultGameParticipants() {
  return DEFAULT_GAME_PARTICIPANTS.map((participant) => ({
    ...participant,
    tags: Array.isArray(participant.tags) ? participant.tags.slice() : []
  }))
}

module.exports = {
  getDefaultGameParticipants
}

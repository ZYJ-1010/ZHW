// The backend is the source of truth for real-name initials. This small map is
// only a UI fallback for a few frames before authorized in-game data arrives.
const COMMON_INITIALS = {
  张: 'Z', 三: 'S', 王: 'W', 小: 'X', 明: 'M', 欧: 'O', 阳: 'Y', 娜: 'N',
  李: 'L', 赵: 'Z', 钱: 'Q', 孙: 'S', 周: 'Z', 吴: 'W', 郑: 'Z', 冯: 'F',
  陈: 'C', 褚: 'C', 卫: 'W', 蒋: 'J', 沈: 'S', 韩: 'H', 杨: 'Y', 朱: 'Z',
  秦: 'Q', 尤: 'Y', 许: 'X', 何: 'H', 吕: 'L', 施: 'S', 孔: 'K', 曹: 'C',
  严: 'Y', 华: 'H', 金: 'J', 魏: 'W', 陶: 'T', 姜: 'J', 戚: 'Q', 谢: 'X',
  邹: 'Z', 喻: 'Y', 柏: 'B', 水: 'S', 窦: 'D', 章: 'Z', 云: 'Y', 苏: 'S',
  潘: 'P', 葛: 'G', 奚: 'X', 范: 'F', 彭: 'P', 郎: 'L', 鲁: 'L', 韦: 'W',
  昌: 'C', 马: 'M', 苗: 'M', 凤: 'F', 花: 'H', 方: 'F', 俞: 'Y', 任: 'R',
  袁: 'Y', 柳: 'L', 鲍: 'B', 史: 'S', 唐: 'T', 费: 'F', 廉: 'L', 岑: 'C',
  薛: 'X', 雷: 'L', 贺: 'H', 倪: 'N', 汤: 'T', 滕: 'T', 殷: 'Y', 罗: 'L',
  毕: 'B', 郝: 'H', 邬: 'W', 安: 'A', 常: 'C', 乐: 'Y', 于: 'Y', 时: 'S',
  傅: 'F', 皮: 'P', 卞: 'B', 齐: 'Q', 康: 'K', 伍: 'W', 余: 'Y', 元: 'Y',
  卜: 'B', 顾: 'G', 孟: 'M', 平: 'P', 黄: 'H', 和: 'H', 穆: 'M', 萧: 'X',
  尹: 'Y', 姚: 'Y', 邵: 'S', 湛: 'Z', 汪: 'W', 祁: 'Q', 毛: 'M', 禹: 'Y',
  狄: 'D', 米: 'M', 贝: 'B', 臧: 'Z', 戴: 'D', 宋: 'S', 茅: 'M', 庞: 'P',
  熊: 'X', 纪: 'J', 舒: 'S', 屈: 'Q', 项: 'X', 祝: 'Z', 董: 'D', 梁: 'L',
  杜: 'D', 阮: 'R', 蓝: 'L', 闵: 'M', 席: 'X', 季: 'J', 麻: 'M', 强: 'Q',
  贾: 'J', 路: 'L', 娄: 'L', 危: 'W', 江: 'J', 童: 'T', 颜: 'Y', 郭: 'G',
  梅: 'M', 盛: 'S', 林: 'L', 刁: 'D', 钟: 'Z', 徐: 'X', 邱: 'Q', 高: 'G',
  夏: 'X', 蔡: 'C', 田: 'T', 樊: 'F', 胡: 'H', 凌: 'L', 霍: 'H', 虞: 'Y',
  万: 'W', 支: 'Z', 柯: 'K', 昝: 'Z', 管: 'G', 卢: 'L', 莫: 'M', 房: 'F',
  裘: 'Q', 缪: 'M', 干: 'G', 解: 'X', 应: 'Y', 宗: 'Z', 丁: 'D', 宣: 'X',
  邓: 'D', 单: 'S', 杭: 'H', 洪: 'H', 包: 'B', 左: 'Z', 石: 'S', 崔: 'C',
  吉: 'J', 龚: 'G', 程: 'C', 嵇: 'J', 邢: 'X', 裴: 'P', 陆: 'L', 荣: 'R',
  翁: 'W', 荀: 'X', 羊: 'Y', 惠: 'H', 甄: 'Z', 曲: 'Q', 家: 'J', 国: 'G',
  文: 'W', 子: 'Z', 伟: 'W', 强: 'Q', 磊: 'L', 洋: 'Y', 勇: 'Y', 艳: 'Y',
  杰: 'J', 娟: 'J', 涛: 'T', 超: 'C', 秀: 'X', 英: 'Y', 霞: 'X', 平: 'P',
  刚: 'G', 桂: 'G', 玲: 'L', 建: 'J', 军: 'J', 玉: 'Y', 兰: 'L', 红: 'H'
}

function normalizeServerInitials(value) {
  return String(value || '').trim().replace(/[^A-Za-z]/g, '').slice(0, 2).toUpperCase()
}

function getRealNameInitials(name, serverInitials = '') {
  const authoritative = normalizeServerInitials(serverInitials)
  if (authoritative) return authoritative

  const chars = Array.from(String(name || '').trim()).filter(char => /[A-Za-z\u4e00-\u9fff]/.test(char)).slice(0, 2)
  return chars.map(char => COMMON_INITIALS[char] || (/^[A-Za-z]$/.test(char) ? char.toUpperCase() : '')).join('')
}

module.exports = {
  getRealNameInitials,
  // Compatibility alias for existing pages. Semantics now use the first two
  // real-name characters, not the old surname-first-two-letters rule.
  getSurnameInitials: getRealNameInitials
}

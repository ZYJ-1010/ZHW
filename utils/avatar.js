const SURNAME_INITIALS = {
  赵: 'ZH',
  钱: 'QI',
  孙: 'SU',
  李: 'LI',
  刘: 'LI',
  周: 'ZH',
  吴: 'WU',
  郑: 'ZH',
  王: 'WA',
  冯: 'FE',
  陈: 'CH',
  褚: 'CH',
  卫: 'WE',
  蒋: 'JI',
  沈: 'SH',
  韩: 'HA',
  杨: 'YA',
  朱: 'ZH',
  秦: 'QI',
  尤: 'YO',
  许: 'XU',
  何: 'HE',
  吕: 'LY',
  施: 'SH',
  张: 'ZH',
  孔: 'KO',
  曹: 'CA',
  严: 'YA',
  华: 'HU',
  金: 'JI',
  魏: 'WE',
  陶: 'TA',
  姜: 'JI',
  戚: 'QI',
  谢: 'XI',
  邹: 'ZO',
  喻: 'YU',
  柏: 'BA',
  水: 'SH',
  窦: 'DO',
  章: 'ZH',
  云: 'YU',
  苏: 'SU',
  潘: 'PA',
  葛: 'GE',
  奚: 'XI',
  范: 'FA',
  彭: 'PE',
  郎: 'LA',
  鲁: 'LU',
  韦: 'WE',
  昌: 'CH',
  马: 'MA',
  苗: 'MI',
  凤: 'FE',
  花: 'HU',
  方: 'FA',
  俞: 'YU',
  任: 'RE',
  袁: 'YU',
  柳: 'LI',
  鲍: 'BA',
  史: 'SH',
  唐: 'TA',
  费: 'FE',
  廉: 'LI',
  岑: 'CE',
  薛: 'XU',
  雷: 'LE',
  贺: 'HE',
  倪: 'NI',
  汤: 'TA',
  滕: 'TE',
  殷: 'YI',
  罗: 'LU',
  郭: 'GU',
  毕: 'BI',
  郝: 'HA',
  邬: 'WU',
  安: 'AN',
  常: 'CH',
  乐: 'YU',
  于: 'YU',
  时: 'SH',
  傅: 'FU',
  皮: 'PI',
  卞: 'BI',
  齐: 'QI',
  康: 'KA',
  伍: 'WU',
  余: 'YU',
  元: 'YU',
  卜: 'BU',
  顾: 'GU',
  孟: 'ME',
  平: 'PI',
  黄: 'HU',
  和: 'HE',
  穆: 'MU',
  萧: 'XI',
  尹: 'YI',
  林: 'LI',
  徐: 'XU',
  高: 'GA',
  梁: 'LI',
  宋: 'SO',
  向: 'XI',
  程: 'CH',
  曾: 'ZE',
  叶: 'YE',
  胡: 'HU',
  夏: 'XI',
  钟: 'ZH',
  田: 'TI',
  杜: 'DU',
  丁: 'DI',
  邓: 'DE',
  江: 'JI',
  崔: 'CU',
  戴: 'DA',
  段: 'DU',
  侯: 'HO',
  赖: 'LA',
  廖: 'LI',
  龙: 'LO',
  陆: 'LU',
  毛: 'MA',
  莫: 'MO',
  石: 'SH',
  万: 'WA',
  熊: 'XI',
  姚: 'YA',
  易: 'YI',
  龚: 'GO',
  文: 'WE',
  翟: 'ZH',
  付: 'FU',
  兰: 'LA',
  闫: 'YA',
  欧: 'OU',
  司: 'SI',
  上: 'SH',
  诸: 'ZH',
  皇: 'HU',
  慕: 'MU',
  端: 'DU',
  令: 'LI',
  单: 'SH',
  仇: 'QI'
}

function normalizeFallback(value, defaultValue = '') {
  const text = String(value || '')
    .trim()
    .replace(/[^a-zA-Z]/g, '')
    .slice(0, 2)
    .toUpperCase()

  return text || defaultValue
}

function isChineseCharacter(value) {
  return /^[\u4e00-\u9fff]$/.test(value)
}

function getSurnameInitials(name, fallback = '') {
  const text = String(name || '').trim()
  const surname = text.charAt(0)
  const fallbackText = normalizeFallback(fallback)

  if (surname && SURNAME_INITIALS[surname]) {
    return SURNAME_INITIALS[surname]
  }

  if (isChineseCharacter(surname)) {
    return fallbackText
  }

  const latinParts = text.match(/[A-Za-z]+/g)

  if (latinParts && latinParts.length) {
    return latinParts[latinParts.length - 1].slice(0, 2).toUpperCase()
  }

  return fallbackText
}

module.exports = {
  getSurnameInitials
}

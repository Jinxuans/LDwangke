/**
 * AI 校正工具 - 清理用户输入的下单信息
 * 支持格式：学校 账号 密码 / 账号 密码 / 各种标签格式
 */

// ===== 全角 → 半角 =====
function toHalfWidth(str: string): string {
  return str.replace(/[\uff01-\uff5e]/g, (ch) =>
    String.fromCharCode(ch.charCodeAt(0) - 0xfee0),
  ).replace(/\u3000/g, ' '); // 全角空格
}

// ===== 中文标点 → 英文标点 =====
const punctuationMap: Record<string, string> = {
  '，': ',', '。': '.', '！': '!', '？': '?', '；': ';', '：': ':',
  '\u201c': '"', '\u201d': '"', '\u2018': "'", '\u2019': "'",
  '（': '(', '）': ')', '【': '[', '】': ']', '《': '<', '》': '>',
  '—': '-', '…': '...', '·': '.',
};

function convertPunctuation(text: string): string {
  return text.replace(/[，。！？；：""''（）【】《》—…·]/g, (ch) => punctuationMap[ch] ?? ch);
}

// ===== 学校名关键词（更全面的匹配） =====
const SCHOOL_KEYWORDS =
  /大学|学院|学校|职院|职大|高专|技师|技校|继续教育|电大|开放大学|党校|研究院|进修|师范|师专|工学|理工|科技|商学|医学|农学|林学|财经|政法|外语|外国语|体育|美术|音乐|艺术|传媒|建筑|航空|航天|海事|海洋|矿业|石油|地质|水利|电力|邮电|铁道|交通|信息|软件/;

function isSchoolName(text: string): boolean {
  return SCHOOL_KEYWORDS.test(text);
}

// ===== 去除常见标签前缀 =====
function stripLabels(text: string): string {
  return text.replace(
    /(?:学校|院校|机构|平台|账号|用户名?|手机号?|密码|口令|pass(?:word)?|user(?:name)?|account|school)\s*[:：=\s]\s*/gi,
    ' ',
  );
}

// ===== 单行校正 =====
function aiRevise(info: string): string {
  if (!info.trim()) return '';

  // 1. 全角 → 半角
  info = toHalfWidth(info);

  // 2. 中文标点 → 英文
  info = convertPunctuation(info);

  // 3. 去除标签前缀
  info = stripLabels(info);

  // 4. 常见分隔符统一为空格（tab / 逗号 / 竖线 / 斜杠）
  info = info.replace(/[,|/\\\t]+/g, ' ');

  // 5. 合并多余空格
  info = info.replace(/\s+/g, ' ').trim();

  // 6. 按空格拆分各段
  const parts = info.split(' ').filter(Boolean);
  if (parts.length === 0) return '';

  // 7. 智能提取：学校名 + 账号 + 密码
  let school = '';
  const nonSchoolParts: string[] = [];

  for (const part of parts) {
    // 纯中文段 → 判断是否学校名
    if (/[\u4e00-\u9fa5]/.test(part)) {
      if (isSchoolName(part) && !school) {
        school = part;
      }
      // 非学校名的中文段丢弃（如"请输入"等干扰文字）
    } else {
      nonSchoolParts.push(part);
    }
  }

  // 8. 重组：学校(可选) 账号 密码 [...其他字段]
  const result = school
    ? [school, ...nonSchoolParts].join(' ')
    : nonSchoolParts.join(' ');

  return result;
}

// ===== 多行校正（带去重） =====
export function aiReviseMultiline(info: string): string {
  const seen = new Set<string>();
  return info
    .split('\n')
    .map((line) => aiRevise(line))
    .filter((line) => {
      if (!line) return false;
      // 按账号去重（取第二段或第一段作为 key）
      const parts = line.split(' ');
      const key = parts.length >= 2 ? parts[1]! : parts[0]!;
      if (seen.has(key)) return false;
      seen.add(key);
      return true;
    })
    .join('\n');
}

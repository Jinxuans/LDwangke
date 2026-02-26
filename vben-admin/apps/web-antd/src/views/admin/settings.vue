<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Button, Input, InputNumber, Switch, Select, SelectOption,
  message, Spin, Tabs, TabPane, Row, Col, Divider, Tag, Checkbox,
} from 'ant-design-vue';
import {
  SaveOutlined, ReloadOutlined, DesktopOutlined, TeamOutlined,
  CreditCardOutlined, AppstoreOutlined, SearchOutlined, SettingOutlined,
  LayoutOutlined, SafetyCertificateOutlined, GiftOutlined,
  PlusOutlined, DeleteOutlined,
} from '@ant-design/icons-vue';
import { getConfigApi, saveConfigApi, getPayDataApi, savePayDataApi } from '#/api/admin';
import { updatePreferences } from '@vben/preferences';

const loading = ref(false);
const saving = ref(false);
const activeTab = ref('wzpz');
const config = ref<Record<string, string>>({});
const payData = ref<Record<string, string>>({});
const backup = ref<{ config: Record<string, string>; payData: Record<string, string> }>({ config: {}, payData: {} });

async function loadAll() {
  loading.value = true;
  try {
    const [cRes, pRes] = await Promise.all([getConfigApi(), getPayDataApi()]);
    config.value = cRes;
    payData.value = pRes;
    backup.value = {
      config: JSON.parse(JSON.stringify(config.value)),
      payData: JSON.parse(JSON.stringify(payData.value)),
    };
  } catch (e) {
    console.error('加载设置失败:', e);
  } finally {
    loading.value = false;
  }
}

async function handleSave() {
  saving.value = true;
  try {
    await Promise.all([saveConfigApi(config.value), savePayDataApi(payData.value)]);
    message.success('保存成功');
    backup.value = {
      config: JSON.parse(JSON.stringify(config.value)),
      payData: JSON.parse(JSON.stringify(payData.value)),
    };
    // 立即应用配置到前端
    applyConfig();
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    saving.value = false;
  }
}

function applyConfig() {
  const c = config.value;
  if (c.sitename) {
    updatePreferences({ app: { name: c.sitename } });
    document.title = c.sitename;
  }
  if (c.sykg === '1') {
    updatePreferences({ app: { watermark: true } });
  } else if (c.sykg === '0') {
    updatePreferences({ app: { watermark: false } });
  }
  if (c.hlogo) {
    updatePreferences({ logo: { source: c.hlogo } });
  } else if (c.logo) {
    updatePreferences({ logo: { source: c.logo } });
  }
}

function handleReset() {
  config.value = JSON.parse(JSON.stringify(backup.value.config));
  payData.value = JSON.parse(JSON.stringify(backup.value.payData));
  message.info('已重置为上次保存的值');
}

function getVal(key: string, def = '') {
  return config.value[key] ?? def;
}
function setVal(key: string, val: string) {
  config.value[key] = val;
}
function getPayVal(key: string, def = '') {
  return payData.value[key] ?? def;
}
function setPayVal(key: string, val: string) {
  payData.value[key] = val;
}

// ===== 充值赠送规则 =====
interface BonusRule { min: number; max: number; bonus_pct: number; }
interface BonusActivity { enabled: boolean; weekdays: number[]; rules: BonusRule[]; }
interface BonusConfig { enabled: boolean; rules: BonusRule[]; activity: BonusActivity; }

const emptyActivity = (): BonusActivity => ({ enabled: false, weekdays: [], rules: [] });

const bonusConfig = ref<BonusConfig>({ enabled: false, rules: [], activity: emptyActivity() });

function parseBonusConfig() {
  const raw = config.value['recharge_bonus_rules'];
  if (!raw) { bonusConfig.value = { enabled: false, rules: [], activity: emptyActivity() }; return; }
  try {
    bonusConfig.value = JSON.parse(raw);
    if (!bonusConfig.value.rules) bonusConfig.value.rules = [];
    if (!bonusConfig.value.activity) bonusConfig.value.activity = emptyActivity();
    if (!bonusConfig.value.activity.weekdays) bonusConfig.value.activity.weekdays = [];
    if (!bonusConfig.value.activity.rules) bonusConfig.value.activity.rules = [];
  } catch { bonusConfig.value = { enabled: false, rules: [], activity: emptyActivity() }; }
}

function syncBonusConfig() {
  config.value['recharge_bonus_rules'] = JSON.stringify(bonusConfig.value);
}

function addBonusRule() {
  const rules = bonusConfig.value.rules;
  const lastMax = rules.length > 0 ? rules[rules.length - 1]!.max : 0;
  rules.push({ min: lastMax, max: lastMax + 500, bonus_pct: 5 });
  syncBonusConfig();
}

function removeBonusRule(idx: number) {
  bonusConfig.value.rules.splice(idx, 1);
  syncBonusConfig();
}

function updateBonusRule(idx: number, field: keyof BonusRule, val: number) {
  (bonusConfig.value.rules[idx] as any)[field] = val;
  syncBonusConfig();
}

function addActivityRule() {
  const rules = bonusConfig.value.activity.rules;
  const lastMax = rules.length > 0 ? rules[rules.length - 1]!.max : 0;
  rules.push({ min: lastMax, max: lastMax + 500, bonus_pct: 10 });
  syncBonusConfig();
}

function removeActivityRule(idx: number) {
  bonusConfig.value.activity.rules.splice(idx, 1);
  syncBonusConfig();
}

function updateActivityRule(idx: number, field: keyof BonusRule, val: number) {
  (bonusConfig.value.activity.rules[idx] as any)[field] = val;
  syncBonusConfig();
}

function toggleBonusEnabled(v: boolean) {
  bonusConfig.value.enabled = v;
  syncBonusConfig();
}

function toggleActivityEnabled(v: boolean) {
  bonusConfig.value.activity.enabled = v;
  syncBonusConfig();
}

const weekdayLabels = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];

function toggleWeekday(day: number) {
  const wds = bonusConfig.value.activity.weekdays;
  const idx = wds.indexOf(day);
  if (idx >= 0) wds.splice(idx, 1);
  else wds.push(day);
  wds.sort((a, b) => a - b);
  syncBonusConfig();
}

onMounted(async () => { await loadAll(); parseBonusConfig(); });
</script>

<template>
  <Page title="系统设置" content-class="p-4">
    <Spin :spinning="loading">
      <Card :body-style="{ padding: '0' }">
        <Tabs v-model:activeKey="activeTab" class="settings-tabs" size="small">
          <!-- 网站配置 -->
          <TabPane key="wzpz">
            <template #tab><DesktopOutlined class="mr-1" />网站配置</template>
            <div class="tab-body">
              <Row :gutter="[24, 24]">
                <Col :span="24"><Divider orientation="left" class="!my-0">基本信息</Divider></Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">站点名称</label>
                  <Input :value="getVal('sitename')" @update:value="(v: string) => setVal('sitename', v)" placeholder="请输入站点名称" />
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">资源版本号</label>
                  <Input :value="getVal('version')" @update:value="(v: string) => setVal('version', v)" placeholder="如 1.0.1" />
                  <div class="text-xs text-gray-400 dark:text-gray-500 mt-1">显示在页面底部，用于标识当前系统版本</div>
                </Col>

                <Col :span="24"><Divider orientation="left" class="!my-0">SEO 设置</Divider></Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">SEO 关键词</label>
                  <Input :value="getVal('keywords')" @update:value="(v: string) => setVal('keywords', v)" placeholder="SEO关键词" />
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">SEO 介绍</label>
                  <Input :value="getVal('description')" @update:value="(v: string) => setVal('description', v)" placeholder="SEO描述" />
                </Col>

                <Col :span="24"><Divider orientation="left" class="!my-0">品牌与视觉</Divider></Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">登录页面 LOGO 地址</label>
                  <Input :value="getVal('logo')" @update:value="(v: string) => setVal('logo', v)" placeholder="https://..." />
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">主页顶部 LOGO 地址</label>
                  <Input :value="getVal('hlogo')" @update:value="(v: string) => setVal('hlogo', v)" placeholder="https://..." />
                </Col>

                <Col :span="24"><Divider orientation="left" class="!my-0">系统功能开关</Divider></Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <div>
                      <label class="field-label !mb-0">全站水印</label>
                      <div class="text-xs text-gray-400 dark:text-gray-500">开启后前台页面将显示防截图水印</div>
                    </div>
                    <Switch :checked="getVal('sykg', '0') === '1'" @change="(v: any) => setVal('sykg', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <div>
                      <label class="field-label !mb-0 text-red-500">维护模式</label>
                      <div class="text-xs text-red-400">开启后普通用户将无法访问前台，仅管理员可用</div>
                    </div>
                    <Switch :checked="getVal('bz', '0') === '1'" @change="(v: any) => setVal('bz', v ? '1' : '0')" />
                  </div>
                </Col>

                <Col :span="24"><Divider orientation="left" class="!my-0">站点公告</Divider></Col>
                <Col :span="24">
                  <label class="field-label">登录页弹窗公告</label>
                  <Input.TextArea :value="getVal('notice')" @update:value="(v: string) => setVal('notice', v)" :rows="6" placeholder="公告内容（支持HTML）" />
                </Col>
                <Col :span="24">
                  <label class="field-label">弹窗公告</label>
                  <Input.TextArea :value="getVal('tcgonggao')" @update:value="(v: string) => setVal('tcgonggao', v)" :rows="4" placeholder="弹窗公告内容" />
                </Col>
              </Row>
            </div>
          </TabPane>

          <!-- 代理配置 -->
          <TabPane key="dlpz">
            <template #tab><TeamOutlined class="mr-1" />代理配置</template>
            <div class="tab-body">
              <Row :gutter="[24, 24]">
                <Col :span="24"><Divider orientation="left" class="!my-0">功能开关</Divider></Col>
                <Col :xs="24" :lg="8">
                  <div class="switch-row">
                    <div>
                      <label class="field-label !mb-0">邀请注册</label>
                      <div class="text-xs text-gray-400 dark:text-gray-500">允许用户通过邀请码注册</div>
                    </div>
                    <Switch :checked="getVal('user_yqzc', '0') === '1'" @change="(v: any) => setVal('user_yqzc', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :xs="24" :lg="8">
                  <div class="switch-row">
                    <div>
                      <label class="field-label !mb-0">上级迁移</label>
                      <div class="text-xs text-gray-400 dark:text-gray-500">允许下级主动迁移至其他上级名下</div>
                    </div>
                    <Switch :checked="getVal('sjqykg', '0') === '1'" @change="(v: any) => setVal('sjqykg', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :xs="24" :lg="8">
                  <div class="switch-row">
                    <div>
                      <label class="field-label !mb-0">跨户开号</label>
                      <div class="text-xs text-gray-400 dark:text-gray-500">允许代理跨级给他人开通下级账号</div>
                    </div>
                    <Switch :checked="getVal('user_htkh', '0') === '1'" @change="(v: any) => setVal('user_htkh', v ? '1' : '0')" />
                  </div>
                </Col>

                <Col :span="24"><Divider orientation="left" class="!my-0">平开代理限制</Divider></Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">平开控制</label>
                  <Select :value="getVal('dl_pkkg', '0')" @change="(v: any) => setVal('dl_pkkg', String(v))" class="w-full">
                    <SelectOption value="0">无限制（正常开启）</SelectOption>
                    <SelectOption value="1">禁止所有等级平开</SelectOption>
                    <SelectOption value="2">顶级平开需双倍余额</SelectOption>
                    <SelectOption value="3">所有等级平开需双倍余额</SelectOption>
                  </Select>
                  <div class="text-xs text-gray-400 dark:text-gray-500 mt-1">控制代理开设同级下级时的限制规则</div>
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">顶级代理费率定义</label>
                  <Input :value="getVal('djfl')" @update:value="(v: string) => setVal('djfl', v)" placeholder="如 0.5" />
                  <div class="text-xs text-gray-400 dark:text-gray-500 mt-1">用于判断是否为顶级代理（用户addprice等于此值即为顶级）</div>
                </Col>

                <Col :span="24"><Divider orientation="left" class="!my-0">跨户充值设置</Divider></Col>
                <Col :span="24">
                  <label class="field-label">授权用户 UID 列表</label>
                  <Input :value="getVal('cross_recharge_uids')" @update:value="(v: string) => setVal('cross_recharge_uids', v)" placeholder="多个UID用逗号分隔，如 2,5,12" />
                  <div class="text-xs text-gray-400 dark:text-gray-500 mt-1">拥有跨户充值权限的用户UID，多个用英文逗号分隔。管理员(UID=1)默认拥有权限。被授权用户可在代理列表中向任意用户转账充值（从自己余额按费率换算扣除）。</div>
                </Col>
              </Row>
            </div>
          </TabPane>

          <!-- 支付配置 -->
          <TabPane key="zfpz">
            <template #tab><CreditCardOutlined class="mr-1" />支付配置</template>
            <div class="tab-body">
              <Row :gutter="[24, 16]">
                <Col :xs="24" :lg="12">
                  <label class="field-label">最低充值</label>
                  <Input :value="getVal('zdpay', '1')" @update:value="(v: string) => setVal('zdpay', v)" placeholder="1" prefix="¥" />
                </Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <label class="field-label !mb-0">非直系代理充值</label>
                    <Switch :checked="getVal('non_direct_recharge_enable') === '1'" @change="(v: any) => setVal('non_direct_recharge_enable', v ? '1' : '0')" />
                  </div>
                  <div class="text-xs text-gray-400 dark:text-gray-500 mt-1">关闭后非直系代理无法在线充值</div>
                </Col>
                <Col :span="24"><Divider class="!my-2">支付渠道开关</Divider></Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <label class="field-label !mb-0">QQ支付</label>
                    <Switch :checked="getPayVal('is_qqpay') === '1'" @change="(v: any) => setPayVal('is_qqpay', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <label class="field-label !mb-0">微信支付</label>
                    <Switch :checked="getPayVal('is_wxpay') === '1'" @change="(v: any) => setPayVal('is_wxpay', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <label class="field-label !mb-0">支付宝支付</label>
                    <Switch :checked="getPayVal('is_alipay') === '1'" @change="(v: any) => setPayVal('is_alipay', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <label class="field-label !mb-0">USDT支付</label>
                    <Switch :checked="getPayVal('is_usdt') === '1'" @change="(v: any) => setPayVal('is_usdt', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :span="24"><Divider class="!my-2">充值赠送规则</Divider></Col>
                <Col :span="24">
                  <div class="switch-row mb-3">
                    <div>
                      <label class="field-label !mb-0">启用充值赠送</label>
                      <div class="text-xs text-gray-400 dark:text-gray-500">开启后用户充值达到指定金额区间时自动赠送对应比例余额</div>
                    </div>
                    <Switch :checked="bonusConfig.enabled" @change="(v: any) => toggleBonusEnabled(!!v)" />
                  </div>
                </Col>

                <template v-if="bonusConfig.enabled">
                  <!-- 区间规则 -->
                  <Col :span="24">
                    <div class="flex items-center justify-between mb-2">
                      <label class="field-label !mb-0">赠送区间规则</label>
                      <Button size="small" type="dashed" @click="addBonusRule"><PlusOutlined /> 添加规则</Button>
                    </div>
                    <div v-if="bonusConfig.rules.length === 0" class="text-gray-400 text-sm py-2">暂无规则，请点击上方添加。</div>
                    <div v-for="(rule, idx) in bonusConfig.rules" :key="idx" class="bonus-rule-row">
                      <span class="text-sm text-gray-500 mr-1">充值</span>
                      <InputNumber :value="rule.min" :min="0" :step="100" size="small" style="width: 100px" @change="(v: any) => updateBonusRule(idx, 'min', v ?? 0)" />
                      <span class="text-sm text-gray-500 mx-1">~</span>
                      <InputNumber :value="rule.max" :min="rule.min + 1" :step="100" size="small" style="width: 100px" @change="(v: any) => updateBonusRule(idx, 'max', v ?? 0)" />
                      <span class="text-sm text-gray-500 mx-1">元，赠送</span>
                      <InputNumber :value="rule.bonus_pct" :min="0" :max="100" :step="1" size="small" style="width: 80px" @change="(v: any) => updateBonusRule(idx, 'bonus_pct', v ?? 0)" />
                      <span class="text-sm text-gray-500 ml-1">%</span>
                      <Button size="small" type="text" danger class="ml-2" @click="removeBonusRule(idx)"><DeleteOutlined /></Button>
                    </div>
                  </Col>

                  <!-- 活动日设置 -->
                  <Col :span="24" class="mt-2">
                    <div class="switch-row mb-3">
                      <div>
                        <label class="field-label !mb-0">活动日规则</label>
                        <div class="text-xs text-gray-400 dark:text-gray-500">开启后，在指定星期使用独立赠送规则替换普通规则，用户端只显示"今日爆率很高"</div>
                      </div>
                      <Switch :checked="bonusConfig.activity.enabled" @change="(v: any) => toggleActivityEnabled(!!v)" />
                    </div>
                  </Col>

                  <template v-if="bonusConfig.activity.enabled">
                    <Col :span="24">
                      <label class="field-label">选择活动日（星期几）</label>
                      <div class="flex flex-wrap gap-2">
                        <Tag
                          v-for="(label, idx) in weekdayLabels" :key="idx"
                          :color="bonusConfig.activity.weekdays.includes(idx) ? 'blue' : ''"
                          class="cursor-pointer select-none !text-sm !px-3 !py-1"
                          @click="toggleWeekday(idx)"
                        >{{ label }}</Tag>
                      </div>
                    </Col>
                    <Col :span="24" class="mt-2">
                      <label class="field-label">活动日提示文案</label>
                      <Input :value="bonusConfig.activity.hint || ''" @update:value="(v: string) => { bonusConfig.activity.hint = v; syncBonusConfig(); }" placeholder="今日爆率很高，充值更划算！" />
                      <div class="text-xs text-gray-400 dark:text-gray-500 mt-1">用户在充值页面看到的活动日提示，留空则显示默认文案</div>
                    </Col>
                    <Col :span="24" class="mt-3">
                      <div class="flex items-center justify-between mb-2">
                        <label class="field-label !mb-0">活动日赠送规则（替换普通规则）</label>
                        <Button size="small" type="dashed" @click="addActivityRule"><PlusOutlined /> 添加规则</Button>
                      </div>
                      <div v-if="bonusConfig.activity.rules.length === 0" class="text-gray-400 text-sm py-2">暂无活动日规则，请点击上方添加。</div>
                      <div v-for="(rule, idx) in bonusConfig.activity.rules" :key="idx" class="bonus-rule-row">
                        <span class="text-sm text-gray-500 mr-1">充值</span>
                        <InputNumber :value="rule.min" :min="0" :step="100" size="small" style="width: 100px" @change="(v: any) => updateActivityRule(idx, 'min', v ?? 0)" />
                        <span class="text-sm text-gray-500 mx-1">~</span>
                        <InputNumber :value="rule.max" :min="rule.min + 1" :step="100" size="small" style="width: 100px" @change="(v: any) => updateActivityRule(idx, 'max', v ?? 0)" />
                        <span class="text-sm text-gray-500 mx-1">元，赠送</span>
                        <InputNumber :value="rule.bonus_pct" :min="0" :max="100" :step="1" size="small" style="width: 80px" @change="(v: any) => updateActivityRule(idx, 'bonus_pct', v ?? 0)" />
                        <span class="text-sm text-gray-500 ml-1">%</span>
                        <Button size="small" type="text" danger class="ml-2" @click="removeActivityRule(idx)"><DeleteOutlined /></Button>
                      </div>
                    </Col>
                  </template>
                </template>

                <Col :span="24"><Divider class="!my-2">易支付配置</Divider></Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">易支付API</label>
                  <Input :value="getPayVal('epay_api')" @update:value="(v: string) => setPayVal('epay_api', v)" placeholder="http://www.example.com/" />
                </Col>
                <Col :xs="24" :lg="6">
                  <label class="field-label">商户ID</label>
                  <Input :value="getPayVal('epay_pid')" @update:value="(v: string) => setPayVal('epay_pid', v)" placeholder="商户ID" />
                </Col>
                <Col :xs="24" :lg="6">
                  <label class="field-label">商户KEY</label>
                  <Input.Password :value="getPayVal('epay_key')" @update:value="(v: string) => setPayVal('epay_key', v)" placeholder="商户密钥" />
                </Col>
              </Row>
            </div>
          </TabPane>

          <!-- 前台配置 -->
          <TabPane key="qtpz">
            <template #tab><LayoutOutlined class="mr-1" />前台配置</template>
            <div class="tab-body">
              <Row :gutter="[24, 24]">
                <Col :span="24"><Divider orientation="left" class="!my-0">前台界面与特效</Divider></Col>
                <Col :xs="24" :lg="8">
                  <div class="switch-row">
                    <div>
                      <label class="field-label !mb-0">前台背景特效</label>
                      <div class="text-xs text-gray-400 dark:text-gray-500">开启后前台页面将显示动态背景特效</div>
                    </div>
                    <Switch :checked="getVal('webVfx_open', '0') === '1'" @change="(v: any) => setVal('webVfx_open', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :xs="24" :lg="8">
                  <div class="switch-row">
                    <div>
                      <label class="field-label !mb-0">防调试模式</label>
                      <div class="text-xs text-gray-400 dark:text-gray-500">开启后禁用F12和右键等开发者工具</div>
                    </div>
                    <Switch :checked="getVal('anti_debug', '0') === '1'" @change="(v: any) => setVal('anti_debug', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :xs="24" :lg="8">
                  <div class="switch-row">
                    <div>
                      <label class="field-label !mb-0">前台登录提示</label>
                      <div class="text-xs text-gray-400 dark:text-gray-500">未登录时访问需要登录的页面弹出提示</div>
                    </div>
                    <Switch :checked="getVal('onlineStore_trdltz', '0') === '1'" @change="(v: any) => setVal('onlineStore_trdltz', v ? '1' : '0')" />
                  </div>
                </Col>

                <Col :span="24"><Divider orientation="left" class="!my-0">视觉与样式</Divider></Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">特效类型</label>
                  <Select :value="getVal('webVfx', '0')" @change="(v: any) => setVal('webVfx', String(v))" class="w-full">
                    <SelectOption value="0">默认特效</SelectOption>
                    <SelectOption value="1">樱花飘落</SelectOption>
                    <SelectOption value="2">雪花飘落</SelectOption>
                    <SelectOption value="3">星星闪烁</SelectOption>
                    <SelectOption value="4">彩色气泡</SelectOption>
                    <SelectOption value="5">黑客帝国代码雨</SelectOption>
                  </Select>
                </Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row h-full items-start">
                    <div>
                      <label class="field-label !mb-0">自定义字体</label>
                      <div class="text-xs text-gray-400 dark:text-gray-500">开启后前台将加载并使用自定义字体</div>
                    </div>
                    <Switch :checked="getVal('fontsZDY', '0') === '1'" @change="(v: any) => setVal('fontsZDY', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :span="24">
                  <label class="field-label">字体CSS代码 (URL)</label>
                  <Input :value="getVal('fontsFamily')" @update:value="(v: string) => setVal('fontsFamily', v)" placeholder="例如: https://fonts.googleapis.com/css2?family=Noto+Sans+SC&display=swap" />
                  <div class="text-xs text-gray-400 dark:text-gray-500 mt-1">需开启自定义字体功能才能生效。填入包含 @font-face 的 CSS 链接或直接写 CSS 代码。</div>
                </Col>

                <Col :span="24"><Divider orientation="left" class="!my-0">渠道与提示</Divider></Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <div>
                      <label class="field-label !mb-0">下单说明展示</label>
                      <div class="text-xs text-gray-400 dark:text-gray-500">在商品下单页面显示下单说明模块</div>
                    </div>
                    <Switch :checked="getVal('xdsmopen', '0') === '1'" @change="(v: any) => setVal('xdsmopen', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <div>
                      <label class="field-label !mb-0">渠道公告</label>
                      <div class="text-xs text-gray-400 dark:text-gray-500">开启后，主页将显示公告信息</div>
                    </div>
                    <Switch :checked="getVal('qd_notice_open', '0') === '1'" @change="(v: any) => setVal('qd_notice_open', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :span="24">
                  <label class="field-label">自定义代码(底部)</label>
                  <Input.TextArea :value="getVal('bottom_code')" @update:value="(v: string) => setVal('bottom_code', v)" :rows="4" placeholder="此处可以放网站统计代码、客服代码等，将插入在页面底部。" />
                </Col>
              </Row>
            </div>
          </TabPane>

          <!-- 分类配置 -->
          <TabPane key="flpz">
            <template #tab><AppstoreOutlined class="mr-1" />分类配置</template>
            <div class="tab-body">
              <Row :gutter="[24, 24]">
                <Col :span="24"><Divider orientation="left" class="!my-0">分类展示</Divider></Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <div>
                      <label class="field-label !mb-0">分类开关</label>
                      <div class="text-xs text-gray-400 dark:text-gray-500">开启后在前台显示商品分类</div>
                    </div>
                    <Switch :checked="getVal('flkg', '0') === '1'" @change="(v: any) => setVal('flkg', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">分类类型</label>
                  <Select :value="getVal('fllx', '0')" @change="(v: any) => setVal('fllx', String(v))" class="w-full">
                    <SelectOption value="0">侧边栏分类</SelectOption>
                    <SelectOption value="1">下单页面选择框分类</SelectOption>
                    <SelectOption value="2">下单页面单选框分类</SelectOption>
                  </Select>
                </Col>
              </Row>
            </div>
          </TabPane>

          <!-- 登录设置 -->
          <TabPane key="dlsz">
            <template #tab><SafetyCertificateOutlined class="mr-1" />登录设置</template>
            <div class="tab-body">
              <Row :gutter="[24, 24]">
                <Col :span="24"><Divider orientation="left" class="!my-0">安全验证</Divider></Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <div>
                      <label class="field-label !mb-0">滑块验证</label>
                      <div class="text-xs text-gray-400 dark:text-gray-500">开启后登录页面需要完成滑块验证才能登录</div>
                    </div>
                    <Switch :checked="getVal('login_slider_verify', '1') === '1'" @change="(v: any) => setVal('login_slider_verify', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <div>
                      <label class="field-label !mb-0">邮箱验证</label>
                      <div class="text-xs text-gray-400 dark:text-gray-500">开启后注册时需要填写邮箱并验证邮箱验证码</div>
                    </div>
                    <Switch :checked="getVal('login_email_verify', '0') === '1'" @change="(v: any) => setVal('login_email_verify', v ? '1' : '0')" />
                  </div>
                </Col>
              </Row>
            </div>
          </TabPane>

          <!-- 签到设置 -->
          <TabPane key="qdsz">
            <template #tab><GiftOutlined class="mr-1" />签到设置</template>
            <div class="tab-body">
              <Row :gutter="[24, 24]">
                <Col :span="24"><Divider orientation="left" class="!my-0">基础开关</Divider></Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <div>
                      <label class="field-label !mb-0">签到功能</label>
                      <div class="text-xs text-gray-400 dark:text-gray-500">开启后用户可每日签到领取随机奖励</div>
                    </div>
                    <Switch :checked="getVal('checkin_enabled', '0') === '1'" @change="(v: any) => setVal('checkin_enabled', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <div>
                      <label class="field-label !mb-0">需要有订单</label>
                      <div class="text-xs text-gray-400 dark:text-gray-500">开启后用户必须有历史订单才能签到</div>
                    </div>
                    <Switch :checked="getVal('checkin_order_required', '0') === '1'" @change="(v: any) => setVal('checkin_order_required', v ? '1' : '0')" />
                  </div>
                </Col>

                <Col :span="24"><Divider orientation="left" class="!my-0">条件与限制</Divider></Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">最低余额要求</label>
                  <Input :value="getVal('checkin_min_balance', '10')" @update:value="(v: string) => setVal('checkin_min_balance', v)" placeholder="10" prefix="¥" />
                  <div class="text-xs text-gray-400 dark:text-gray-500 mt-1">用户余额不低于此值才能签到</div>
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">每日签到名额</label>
                  <Input :value="getVal('checkin_max_users', '10')" @update:value="(v: string) => setVal('checkin_max_users', v)" placeholder="10" suffix="人" />
                  <div class="text-xs text-gray-400 dark:text-gray-500 mt-1">每天最多允许签到的人数</div>
                </Col>

                <Col :span="24"><Divider orientation="left" class="!my-0">奖励设置</Divider></Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">最小奖励金额</label>
                  <Input :value="getVal('checkin_min_reward', '0.1')" @update:value="(v: string) => setVal('checkin_min_reward', v)" placeholder="0.1" prefix="¥" />
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">最大奖励金额</label>
                  <Input :value="getVal('checkin_max_reward', '0.2')" @update:value="(v: string) => setVal('checkin_max_reward', v)" placeholder="0.2" prefix="¥" />
                </Col>
              </Row>
            </div>
          </TabPane>

          <!-- 推送与同步 -->
          <TabPane key="tspz">
            <template #tab><SettingOutlined class="mr-1" />推送与同步</template>
            <div class="tab-body">
              <Row :gutter="[24, 24]">
                <Col :span="24"><Divider orientation="left" class="!my-0">WxPusher 微信推送</Divider></Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">WxPusher AppToken</label>
                  <Input.Password :value="getVal('wxpusher_token')" @update:value="(v: string) => setVal('wxpusher_token', v)" placeholder="AT_xxxxxxx" />
                  <div class="text-xs text-gray-400 dark:text-gray-500 mt-1">在 wxpusher.zjiecode.com 创建应用后获取</div>
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">WxPusher 应用ID</label>
                  <Input :value="getVal('wxpusher_appid')" @update:value="(v: string) => setVal('wxpusher_appid', v)" placeholder="应用ID（用于生成关注二维码）" />
                </Col>

                <Col :span="24"><Divider orientation="left" class="!my-0">Pup 自动登录</Divider></Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">Pup 登录地址</label>
                  <Input :value="getVal('pup_base_url')" @update:value="(v: string) => setVal('pup_base_url', v)" placeholder="https://demo.yehuimei.xyz/autologin/index.php" />
                  <div class="text-xs text-gray-400 dark:text-gray-500 mt-1">Pup 浏览器插件自动登录的目标地址</div>
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">Pup Plan</label>
                  <Input :value="getVal('pup_plan')" @update:value="(v: string) => setVal('pup_plan', v)" placeholder="计划名称（留空为默认）" />
                </Col>

                <Col :span="24"><Divider orientation="left" class="!my-0">自动上下架同步</Divider></Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">自动同步货源 HID</label>
                  <Input :value="getVal('auto_sync_hids')" @update:value="(v: string) => setVal('auto_sync_hids', v)" placeholder="多个用逗号分隔，如 1,2,3" />
                  <div class="text-xs text-gray-400 dark:text-gray-500 mt-1">填写需要自动上下架同步的货源HID，系统每30分钟自动执行一次</div>
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">同步价格倍率</label>
                  <Input :value="getVal('auto_sync_rate', '5')" @update:value="(v: string) => setVal('auto_sync_rate', v)" placeholder="5" />
                  <div class="text-xs text-gray-400 dark:text-gray-500 mt-1">上游价格 × 此倍率 = 本站售价</div>
                </Col>
              </Row>
            </div>
          </TabPane>

          <!-- 其他设置 -->
          <TabPane key="qtpz2">
            <template #tab><SettingOutlined class="mr-1" />其他设置</template>
            <div class="tab-body">
              <Row :gutter="[24, 16]">
                <Col :xs="24" :lg="12">
                  <label class="field-label">反调试保护（F12/DevTools检测）</label>
                  <Select :value="getVal('anti_debug', '1')" @change="(v: any) => setVal('anti_debug', String(v))" class="w-full">
                    <SelectOption value="1">开启</SelectOption>
                    <SelectOption value="0">关闭</SelectOption>
                  </Select>
                  <div class="text-xs text-gray-400 dark:text-gray-500 mt-1">关闭后移动端将不再因误判而自动跳转，PC端也不再拦截开发者工具</div>
                </Col>
              </Row>
            </div>
          </TabPane>
        </Tabs>

        <!-- 底部操作栏 -->
        <div class="save-bar">
          <Button @click="handleReset">
            <template #icon><ReloadOutlined /></template>
            重置
          </Button>
          <Button type="primary" :loading="saving" @click="handleSave">
            <template #icon><SaveOutlined /></template>
            保存设置
          </Button>
        </div>
      </Card>
    </Spin>
  </Page>
</template>

<style scoped>
.settings-tabs :deep(.ant-tabs-nav) {
  padding: 0 24px;
  margin-bottom: 0;
}
.settings-tabs :deep(.ant-tabs-nav-more[aria-hidden="true"]) {
  visibility: hidden;
  pointer-events: none;
}
.tab-body {
  padding: 24px;
}
.field-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: #555;
  margin-bottom: 6px;
}
html.dark .field-label { color: #aaa; }
.switch-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 10px 16px;
  border-radius: 8px;
  background: #fafafa;
  border: 1px solid #f0f0f0;
}
.switch-row > div:first-child {
  min-width: 0;
  flex: 1;
  padding-right: 12px;
}
.switch-row .ant-switch {
  flex-shrink: 0;
}
html.dark .switch-row { background: #1f1f1f; border-color: #333; }
.save-bar {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #f0f0f0;
  background: #fafafa;
}
html.dark .save-bar { background: #1f1f1f; border-top-color: #333; }
.bonus-rule-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
  padding: 8px 12px;
  margin-bottom: 8px;
  border-radius: 8px;
  background: #fafafa;
  border: 1px solid #f0f0f0;
}
html.dark .bonus-rule-row { background: #1f1f1f; border-color: #333; }
</style>

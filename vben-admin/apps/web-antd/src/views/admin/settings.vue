<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Button, Input, Switch, Select, SelectOption,
  message, Spin, Tabs, TabPane, Row, Col, Divider,
} from 'ant-design-vue';
import {
  SaveOutlined, ReloadOutlined, DesktopOutlined, TeamOutlined,
  CreditCardOutlined, AppstoreOutlined, SearchOutlined, SettingOutlined,
  LayoutOutlined, SafetyCertificateOutlined, GiftOutlined,
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

onMounted(loadAll);
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
              <Row :gutter="[24, 16]">
                <Col :xs="24" :lg="12">
                  <label class="field-label">站点名字</label>
                  <Input :value="getVal('sitename')" @update:value="(v: string) => setVal('sitename', v)" placeholder="站点名字" />
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">SEO关键词</label>
                  <Input :value="getVal('keywords')" @update:value="(v: string) => setVal('keywords', v)" placeholder="SEO关键词" />
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">SEO介绍</label>
                  <Input :value="getVal('description')" @update:value="(v: string) => setVal('description', v)" placeholder="SEO描述" />
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">登录页面LOGO地址</label>
                  <Input :value="getVal('logo')" @update:value="(v: string) => setVal('logo', v)" placeholder="https://..." />
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">主页顶部LOGO地址</label>
                  <Input :value="getVal('hlogo')" @update:value="(v: string) => setVal('hlogo', v)" placeholder="https://..." />
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">是否开启水印</label>
                  <Select :value="getVal('sykg', '1')" @change="(v: any) => setVal('sykg', String(v))" class="w-full">
                    <SelectOption value="1">开启</SelectOption>
                    <SelectOption value="0">关闭</SelectOption>
                  </Select>
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">维护模式</label>
                  <Select :value="getVal('bz', '0')" @change="(v: any) => setVal('bz', String(v))" class="w-full">
                    <SelectOption value="1">开启（仅管理员可访问）</SelectOption>
                    <SelectOption value="0">关闭</SelectOption>
                  </Select>
                  <div class="text-xs text-gray-400 mt-1">开启后普通用户将无法访问前台，仅管理员可用</div>
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">资源版本号</label>
                  <Input :value="getVal('version')" @update:value="(v: string) => setVal('version', v)" placeholder="如 1.0.1" />
                  <div class="text-xs text-gray-400 mt-1">显示在页面底部，用于标识当前系统版本</div>
                </Col>
                <Col :span="24">
                  <label class="field-label">公告</label>
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
              <Row :gutter="[24, 16]">
                <Col :xs="24" :lg="12">
                  <label class="field-label">是否开启上级迁移功能</label>
                  <Select :value="getVal('sjqykg', '0')" @change="(v: any) => setVal('sjqykg', String(v))" class="w-full">
                    <SelectOption value="1">开启</SelectOption>
                    <SelectOption value="0">关闭</SelectOption>
                  </Select>
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">是否允许邀请码注册</label>
                  <Select :value="getVal('user_yqzc', '0')" @change="(v: any) => setVal('user_yqzc', String(v))" class="w-full">
                    <SelectOption value="1">允许</SelectOption>
                    <SelectOption value="0">拒绝</SelectOption>
                  </Select>
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">是否允许后台开户</label>
                  <Select :value="getVal('user_htkh', '0')" @change="(v: any) => setVal('user_htkh', String(v))" class="w-full">
                    <SelectOption value="1">允许</SelectOption>
                    <SelectOption value="0">拒绝</SelectOption>
                  </Select>
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">代理开通价格</label>
                  <Input :value="getVal('user_ktmoney')" @update:value="(v: string) => setVal('user_ktmoney', v)" placeholder="0" prefix="¥" />
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">商城开通价格</label>
                  <Input :value="getVal('mall_open_price')" @update:value="(v: string) => setVal('mall_open_price', v)" placeholder="99" prefix="¥" />
                  <div class="text-xs text-gray-400 mt-1">代理开通商城所需余额，默认 99 元</div>
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">代理平开控制</label>
                  <Select :value="getVal('dl_pkkg', '0')" @change="(v: any) => setVal('dl_pkkg', String(v))" class="w-full">
                    <SelectOption value="0">不限制</SelectOption>
                    <SelectOption value="1">顶级不允许平开</SelectOption>
                    <SelectOption value="2">顶级平开需双倍余额</SelectOption>
                    <SelectOption value="3">所有等级平开需双倍余额</SelectOption>
                  </Select>
                  <div class="text-xs text-gray-400 mt-1">控制代理开设同级下级时的限制规则</div>
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">顶级代理费率</label>
                  <Input :value="getVal('djfl')" @update:value="(v: string) => setVal('djfl', v)" placeholder="如 0.5" />
                  <div class="text-xs text-gray-400 mt-1">用于判断是否为顶级代理（addprice 等于此值即为顶级）</div>
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
                  <div class="text-xs text-gray-400 mt-1">关闭后非直系代理无法在线充值</div>
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

          <!-- 分类配置 -->
          <TabPane key="flpz">
            <template #tab><AppstoreOutlined class="mr-1" />分类配置</template>
            <div class="tab-body">
              <Row :gutter="[24, 16]">
                <Col :xs="24" :lg="12">
                  <label class="field-label">分类开关</label>
                  <Select :value="getVal('flkg', '1')" @change="(v: any) => setVal('flkg', String(v))" class="w-full">
                    <SelectOption value="1">开启</SelectOption>
                    <SelectOption value="0">关闭</SelectOption>
                  </Select>
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

          <!-- 查课配置 -->
          <TabPane key="ckpz">
            <template #tab><SearchOutlined class="mr-1" />查课配置</template>
            <div class="tab-body">
              <Row :gutter="[24, 16]">
                <Col :xs="24" :lg="12">
                  <label class="field-label">是否开启API调用功能</label>
                  <Select :value="getVal('settings', '1')" @change="(v: any) => setVal('settings', String(v))" class="w-full">
                    <SelectOption value="1">开启API调用</SelectOption>
                    <SelectOption value="0">关闭API调用</SelectOption>
                  </Select>
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">API调用扣费限制</label>
                  <Input :value="getVal('api_proportion')" @update:value="(v: string) => setVal('api_proportion', v)" placeholder="0" suffix="%" />
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">API查课余额限制</label>
                  <Input :value="getVal('api_ck')" @update:value="(v: string) => setVal('api_ck', v)" placeholder="0" prefix="¥" />
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">API下单余额限制</label>
                  <Input :value="getVal('api_xd')" @update:value="(v: string) => setVal('api_xd', v)" placeholder="0" prefix="¥" />
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">API同步随机时间最小</label>
                  <Input :value="getVal('api_tongb')" @update:value="(v: string) => setVal('api_tongb', v)" placeholder="0" suffix="分钟" />
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">API同步随机时间最大</label>
                  <Input :value="getVal('api_tongbc')" @update:value="(v: string) => setVal('api_tongbc', v)" placeholder="0" suffix="分钟" />
                </Col>
              </Row>
            </div>
          </TabPane>

          <!-- 前端显示 -->
          <TabPane key="qdpz">
            <template #tab><LayoutOutlined class="mr-1" />前端显示</template>
            <div class="tab-body">
              <Row :gutter="[24, 16]">
                <Col :xs="24" :lg="12">
                  <label class="field-label">自定义字体开关</label>
                  <Select :value="getVal('fontsZDY', '0')" @change="(v: any) => setVal('fontsZDY', String(v))" class="w-full">
                    <SelectOption value="1">开启</SelectOption>
                    <SelectOption value="0">关闭</SelectOption>
                  </Select>
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">字体族</label>
                  <Input :value="getVal('fontsFamily')" @update:value="(v: string) => setVal('fontsFamily', v)" placeholder="如: 'Noto Serif SC', serif" />
                </Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <label class="field-label !mb-0">渠道公告</label>
                    <Switch :checked="getVal('qd_notice_open') === '1'" @change="(v: any) => setVal('qd_notice_open', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <label class="field-label !mb-0">下单扫码</label>
                    <Switch :checked="getVal('xdsmopen') === '1'" @change="(v: any) => setVal('xdsmopen', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <label class="field-label !mb-0">代理登入跳转</label>
                    <Switch :checked="getVal('onlineStore_trdltz') === '1'" @change="(v: any) => setVal('onlineStore_trdltz', v ? '1' : '0')" />
                  </div>
                  <div class="text-xs text-gray-400 mt-1">开启后管理员可从前台跳转代理登入页</div>
                </Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <label class="field-label !mb-0">自定义特效</label>
                    <Switch :checked="getVal('webVfx_open') === '1'" @change="(v: any) => setVal('webVfx_open', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :span="24" v-if="getVal('webVfx_open') === '1'">
                  <label class="field-label">特效代码</label>
                  <Input.TextArea :value="getVal('webVfx')" @update:value="(v: string) => setVal('webVfx', v)" :rows="4" placeholder="自定义 CSS/JS 特效代码" />
                </Col>
              </Row>
            </div>
          </TabPane>

          <!-- 登录设置 -->
          <TabPane key="dlsz">
            <template #tab><SafetyCertificateOutlined class="mr-1" />登录设置</template>
            <div class="tab-body">
              <Row :gutter="[24, 16]">
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <div>
                      <label class="field-label !mb-0">滑块验证</label>
                      <div class="text-xs text-gray-400">开启后登录页面需要完成滑块验证才能登录</div>
                    </div>
                    <Switch :checked="getVal('login_slider_verify', '1') === '1'" @change="(v: any) => setVal('login_slider_verify', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <div>
                      <label class="field-label !mb-0">邮箱验证</label>
                      <div class="text-xs text-gray-400">开启后注册时需要填写邮箱并验证邮箱验证码</div>
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
              <Row :gutter="[24, 16]">
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <div>
                      <label class="field-label !mb-0">签到功能</label>
                      <div class="text-xs text-gray-400">开启后用户可每日签到领取随机奖励</div>
                    </div>
                    <Switch :checked="getVal('checkin_enabled', '0') === '1'" @change="(v: any) => setVal('checkin_enabled', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :xs="24" :lg="12">
                  <div class="switch-row">
                    <div>
                      <label class="field-label !mb-0">需要有订单</label>
                      <div class="text-xs text-gray-400">开启后用户必须有历史订单才能签到</div>
                    </div>
                    <Switch :checked="getVal('checkin_order_required', '0') === '1'" @change="(v: any) => setVal('checkin_order_required', v ? '1' : '0')" />
                  </div>
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">最低余额要求</label>
                  <Input :value="getVal('checkin_min_balance', '10')" @update:value="(v: string) => setVal('checkin_min_balance', v)" placeholder="10" prefix="¥" />
                  <div class="text-xs text-gray-400 mt-1">用户余额不低于此值才能签到</div>
                </Col>
                <Col :xs="24" :lg="12">
                  <label class="field-label">每日签到名额</label>
                  <Input :value="getVal('checkin_max_users', '10')" @update:value="(v: string) => setVal('checkin_max_users', v)" placeholder="10" suffix="人" />
                  <div class="text-xs text-gray-400 mt-1">每天最多允许签到的人数</div>
                </Col>
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

          <!-- 其他设置 -->
          <TabPane key="qtpz">
            <template #tab><SettingOutlined class="mr-1" />其他设置</template>
            <div class="tab-body">
              <Row :gutter="[24, 16]">
                <Col :xs="24" :lg="12">
                  <label class="field-label">反调试保护（F12/DevTools检测）</label>
                  <Select :value="getVal('anti_debug', '1')" @change="(v: any) => setVal('anti_debug', String(v))" class="w-full">
                    <SelectOption value="1">开启</SelectOption>
                    <SelectOption value="0">关闭</SelectOption>
                  </Select>
                  <div class="text-xs text-gray-400 mt-1">关闭后移动端将不再因误判而自动跳转，PC端也不再拦截开发者工具</div>
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
.switch-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  border-radius: 8px;
  background: #fafafa;
  border: 1px solid #f0f0f0;
}
.save-bar {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #f0f0f0;
  background: #fafafa;
}
</style>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import { Card, Button, Input, Textarea, Switch, InputNumber, Select, message, Alert, Modal, Spin } from 'ant-design-vue';
import { SaveOutlined, ShopOutlined, UnlockOutlined } from '@ant-design/icons-vue';
import { getTenantShopApi, saveTenantMallConfigApi, saveTenantShopApi, type TenantInfo } from '#/api/tenant';
import { requestClient } from '#/api/request';

const loading = ref(false);
const saving = ref(false);
const opened = ref(false);
const openPrice = ref(0);
const opening = ref(false);
const openShopName = ref('');
const mallDomainSuffix = ref('');
const subdomainPrefix = ref('');

const form = reactive<Partial<TenantInfo>>({
  tid: 0,
  shop_name: '',
  shop_logo: '',
  shop_desc: '',
  status: 1,
  domain: '',
});

const mallConfig = reactive({
  register_enabled: false,
  promotion_enabled: false,
  commission_rate: 0,
  show_categories: true,
  popup_notice_html: '',
  customer_service: {
    enabled: false,
    type: 'wechat',
    value: '',
    label: '联系客服',
  },
});

function normalizeDomain(raw?: string) {
  let value = String(raw || '').trim().toLowerCase();
  value = value.replace(/^https?:\/\//, '');
  value = value.replace(/\/.*$/, '');
  value = value.replace(/:\d+$/, '');
  return value;
}

function normalizeSubdomainPrefix(raw?: string) {
  return String(raw || '')
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9-.]/g, '')
    .replace(/\.+/g, '.')
    .replace(/^\.+|\.+$/g, '');
}

function extractPrefixFromDomain(domain: string, suffix: string) {
  const normalizedDomain = normalizeDomain(domain);
  const normalizedSuffix = normalizeDomain(suffix);
  if (!normalizedDomain || !normalizedSuffix) return '';
  if (normalizedDomain === normalizedSuffix) return '';
  const suffixToken = `.${normalizedSuffix}`;
  if (!normalizedDomain.endsWith(suffixToken)) return '';
  return normalizedDomain.slice(0, -suffixToken.length);
}

const mallDomainEnabled = computed(() => !!mallDomainSuffix.value);

const effectiveDomain = computed(() => {
  if (!mallDomainEnabled.value) return '';
  const prefix = normalizeSubdomainPrefix(subdomainPrefix.value);
  return prefix ? `${prefix}.${mallDomainSuffix.value}` : '';
});

const mallUrl = computed(() => {
  const domain = effectiveDomain.value;
  if (domain) {
    return `https://${domain}/mall/`;
  }
  if (!form.tid) return '';
  return `${window.location.origin}/mall/${form.tid}/`;
});

async function load() {
  loading.value = true;
  try {
    const res = await getTenantShopApi();
    const data = res;
    Object.assign(form, data);
    mallDomainSuffix.value = normalizeDomain(data?.mall_domain_suffix);
    form.domain = normalizeDomain(data?.domain);
    mallConfig.register_enabled = false;
    mallConfig.promotion_enabled = false;
    mallConfig.commission_rate = 0;
    mallConfig.show_categories = true;
    mallConfig.popup_notice_html = '';
    mallConfig.customer_service = {
      enabled: false,
      type: 'wechat',
      value: '',
      label: '联系客服',
    };
    if (data?.mall_config) {
      try {
        const cfg = JSON.parse(data.mall_config);
        mallConfig.register_enabled = !!cfg?.register_enabled;
        mallConfig.promotion_enabled = !!cfg?.promotion_enabled;
        mallConfig.commission_rate = Number(cfg?.commission_rate || 0);
        mallConfig.show_categories = cfg?.show_categories !== false;
        mallConfig.popup_notice_html = String(cfg?.popup_notice_html || '');
        mallConfig.customer_service = {
          enabled: !!cfg?.customer_service?.enabled,
          type: String(cfg?.customer_service?.type || 'wechat'),
          value: String(cfg?.customer_service?.value || ''),
          label: String(cfg?.customer_service?.label || '联系客服'),
        };
      } catch {}
    }
    subdomainPrefix.value = extractPrefixFromDomain(form.domain || '', mallDomainSuffix.value);
    opened.value = true;
  } catch {
    opened.value = false;
    try {
      const r = await requestClient.get('/tenant/mall-open-price');
      openPrice.value = (r as any)?.data?.price ?? (r as any)?.price ?? 99;
    } catch {}
  } finally {
    loading.value = false;
  }
}

async function handleOpen() {
  if (!openShopName.value.trim()) {
    message.warning('请填写店铺名称');
    return;
  }
  Modal.confirm({
    title: '确认开通商城',
    content: `将从余额扣除 ${openPrice.value} 元开通商城，确认继续？`,
    okText: '确认开通',
    cancelText: '取消',
    onOk: async () => {
      opening.value = true;
      try {
        await requestClient.post('/tenant/mall-open', { shop_name: openShopName.value });
        message.success('商城开通成功');
        await load();
      } catch (e: any) {
        message.error(e?.message || '开通失败');
      } finally {
        opening.value = false;
      }
    },
  });
}

async function handleSave() {
  if (!form.shop_name?.trim()) { message.warning('请填写店铺名称'); return; }
  if (mallDomainEnabled.value) {
    subdomainPrefix.value = normalizeSubdomainPrefix(subdomainPrefix.value);
    if (!subdomainPrefix.value) {
      message.warning('请填写商城子域名前缀');
      return;
    }
    form.domain = `${subdomainPrefix.value}.${mallDomainSuffix.value}`;
  } else {
    form.domain = '';
  }
  saving.value = true;
  try {
    await saveTenantShopApi({ ...form });
    await saveTenantMallConfigApi(JSON.stringify({
      register_enabled: mallConfig.register_enabled,
      promotion_enabled: mallConfig.promotion_enabled,
      commission_rate: Number(mallConfig.commission_rate || 0),
      show_categories: mallConfig.show_categories,
      popup_notice_html: mallConfig.popup_notice_html,
      customer_service: mallConfig.customer_service,
    }));
    message.success('保存成功');
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    saving.value = false;
  }
}

onMounted(load);
</script>

<template>
  <Page title="店铺设置" content-class="p-4">
    <Spin :spinning="loading">
      <!-- 未开通：付费开通引导 -->
      <div v-if="!loading && !opened" style="max-width: 480px">
        <Card>
          <div class="text-center py-6 space-y-4">
            <ShopOutlined style="font-size: 56px; color: #1677ff" />
            <div>
              <div class="text-xl font-semibold mb-1">开通专属商城</div>
              <div class="text-gray-500 text-sm">开通后即可拥有独立商城，销售课程，管理学员</div>
            </div>
            <div class="text-3xl font-bold text-blue-500">
              ¥{{ openPrice }}
              <span class="text-base font-normal text-gray-400 dark:text-gray-500">一次性开通</span>
            </div>
            <div style="max-width: 280px; margin: 0 auto">
              <Input v-model:value="openShopName" placeholder="请输入店铺名称" size="large" class="mb-3" />
              <Button type="primary" size="large" block :loading="opening" @click="handleOpen">
                <template #icon><UnlockOutlined /></template>
                立即开通
              </Button>
            </div>
          </div>
        </Card>
      </div>

      <!-- 已开通：店铺设置 -->
      <template v-if="opened">
        <div v-if="mallUrl" class="mb-4" style="max-width: 600px">
          <Alert :type="form.domain ? 'success' : 'info'" show-icon>
            <template #message>
              <span class="text-sm">我的商城地址：</span>
              <a :href="mallUrl" target="_blank" class="font-medium mx-2">{{ mallUrl }}</a>
              <Button size="small" type="link" @click="() => { navigator.clipboard.writeText(mallUrl); message.success('已复制'); }">复制</Button>
              <Button size="small" type="primary" ghost :href="mallUrl" target="_blank">
                <template #icon><ShopOutlined /></template>
                进入商城
              </Button>
            </template>
            <template #description>
              <span v-if="form.domain">当前已启用域名访问，用户可直接通过该域名打开商城，无需再带店铺ID。</span>
              <span v-else>当前未启用商城域名，系统仍使用默认地址 `/mall/{{ form.tid }}/` 访问。</span>
            </template>
          </Alert>
        </div>

        <Card style="max-width: 600px">
          <div class="space-y-5">
            <div>
              <label class="block text-sm font-medium mb-1">店铺名称</label>
              <Input v-model:value="form.shop_name" placeholder="请输入店铺名称" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Logo URL</label>
              <Input v-model:value="form.shop_logo" placeholder="https://..." />
              <img v-if="form.shop_logo" :src="form.shop_logo" class="mt-2 h-16 rounded" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">跑马灯公告</label>
              <Textarea v-model:value="form.shop_desc" :rows="3" placeholder="展示在商城首页滚动公告区域的文本内容" />
              <div class="mt-2 text-xs text-gray-500">这里更适合放活动提醒、优惠信息、客服说明，不再当普通店铺简介使用。</div>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">商城域名</label>
              <template v-if="mallDomainEnabled">
                <Input
                  v-model:value="subdomainPrefix"
                  :addon-after="`.${mallDomainSuffix}`"
                  placeholder="例如：zhangsan"
                  @blur="subdomainPrefix = normalizeSubdomainPrefix(subdomainPrefix)"
                />
                <div class="mt-2 text-xs leading-6 text-gray-500">
                  平台已配置商城主域名 <code>{{ mallDomainSuffix }}</code>，这里只需要填写你店铺的子域名前缀。
                </div>
                <div class="text-xs leading-6 text-gray-500">
                  例如前缀填 <code>zhangsan</code>，商城访问地址就是 <code>https://zhangsan.{{ mallDomainSuffix }}/mall/</code>。
                </div>
              </template>
              <template v-else>
                <Alert
                  type="warning"
                  show-icon
                  message="主站尚未配置商城主域名"
                  description="当前不开放商城域名配置。请联系平台管理员先在系统设置中配置“商城主域名”，配置完成后这里才可填写子域名前缀。"
                />
              </template>
            </div>
            <div class="flex items-center gap-3">
              <label class="text-sm font-medium">店铺状态</label>
              <Switch v-model:checked="form.status" :checked-value="1" :un-checked-value="0" checked-children="开启" un-checked-children="关闭" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-2">商城分类栏</label>
              <div class="flex items-center gap-3">
                <Switch v-model:checked="mallConfig.show_categories" checked-children="显示" un-checked-children="隐藏" />
                <span class="text-xs text-gray-500">关闭后，C 端首页不显示分类切换栏，商品仍会正常展示。</span>
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium mb-2">会员注册</label>
              <div class="flex items-center gap-3">
                <Switch v-model:checked="mallConfig.register_enabled" checked-children="开启" un-checked-children="关闭" />
                <span class="text-xs text-gray-500">开启后，访客可在商城 C 端自助注册会员账号。</span>
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium mb-2">推广返利</label>
              <div class="space-y-3">
                <div class="flex items-center gap-3">
                  <Switch v-model:checked="mallConfig.promotion_enabled" checked-children="开启" un-checked-children="关闭" />
                  <span class="text-xs text-gray-500">会员可生成推广商城链接，带客下单成功后获得返利。</span>
                </div>
                <div class="max-w-[220px]">
                  <InputNumber
                    v-model:value="mallConfig.commission_rate"
                    :min="0"
                    :max="100"
                    :precision="2"
                    style="width: 100%"
                    addon-after="%"
                    placeholder="返利比例"
                  />
                  <div class="mt-2 text-xs text-gray-500">按商城支付金额返利，建议先从 1%-10% 起步。</div>
                </div>
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium mb-2">弹窗公告（支持 HTML）</label>
              <Textarea
                v-model:value="mallConfig.popup_notice_html"
                :rows="6"
                placeholder="<p>支持基础 HTML，例如标题、段落、加粗、链接等</p>"
              />
              <div class="mt-2 text-xs text-gray-500">
                用户进入商城首页时会弹出一次。可用于活动公告、发货说明、免责提醒等。
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium mb-2">客服信息</label>
              <div class="space-y-3 rounded-xl border border-gray-200 p-4">
                <div class="flex items-center gap-3">
                  <Switch v-model:checked="mallConfig.customer_service.enabled" checked-children="开启" un-checked-children="关闭" />
                  <span class="text-xs text-gray-500">开启后会在商城首页、查进度页、我的页面展示客服入口。</span>
                </div>
                <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
                  <div>
                    <label class="mb-1 block text-sm font-medium">客服方式</label>
                    <Select
                      v-model:value="mallConfig.customer_service.type"
                      style="width: 100%"
                      :options="[
                        { label: '微信', value: 'wechat' },
                        { label: 'QQ', value: 'qq' },
                        { label: '电话', value: 'phone' },
                        { label: '链接', value: 'link' },
                      ]"
                    />
                  </div>
                  <div>
                    <label class="mb-1 block text-sm font-medium">按钮文案</label>
                    <Input v-model:value="mallConfig.customer_service.label" placeholder="例如：联系客服" />
                  </div>
                </div>
                <div>
                  <label class="mb-1 block text-sm font-medium">客服内容</label>
                  <Input v-model:value="mallConfig.customer_service.value" placeholder="填写微信号、QQ号、手机号或客服链接" />
                </div>
              </div>
            </div>
            <div>
              <Button type="primary" :loading="saving" @click="handleSave">
                <template #icon><SaveOutlined /></template>
                保存设置
              </Button>
            </div>
          </div>
        </Card>
      </template>
    </Spin>
  </Page>
</template>

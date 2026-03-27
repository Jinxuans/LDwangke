<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card,
  Button,
  Input,
  Switch,
  message,
  Divider,
  Alert,
  RadioGroup,
  RadioButton,
} from 'ant-design-vue';
import { SaveOutlined, KeyOutlined, AlipayCircleOutlined, WechatOutlined, QqOutlined } from '@ant-design/icons-vue';
import { getTenantShopApi, saveTenantPayConfigApi } from '#/api/tenant';

const loading = ref(false);
const saving = ref(false);

type PayMode = 'site_balance' | 'merchant_epay';

interface PayConfig {
  pay_mode: PayMode;
  epay_api: string;
  epay_pid: string;
  epay_key: string;
  is_alipay: string;
  is_wxpay: string;
  is_qqpay: string;
}

const form = reactive<PayConfig>({
  pay_mode: 'site_balance',
  epay_api: '',
  epay_pid: '',
  epay_key: '',
  is_alipay: '1',
  is_wxpay: '0',
  is_qqpay: '0',
});

const isSiteBalanceMode = computed(() => form.pay_mode === 'site_balance');

function resolvePayMode(cfg: Record<string, any>): PayMode {
  const mode = String(cfg.pay_mode || cfg.mode || '').trim();
  if (mode === 'merchant_epay' || mode === 'site_balance') {
    return mode;
  }
  if (cfg.epay_api || cfg.epay_pid || cfg.epay_key) {
    return 'merchant_epay';
  }
  return 'site_balance';
}

async function load() {
  loading.value = true;
  try {
    const res = await getTenantShopApi();
    const data = res;
    if (data?.pay_config) {
      try {
        const cfg = JSON.parse(data.pay_config);
        Object.assign(form, {
          pay_mode: resolvePayMode(cfg),
          epay_api: cfg.epay_api || '',
          epay_pid: cfg.epay_pid || '',
          epay_key: cfg.epay_key || '',
          is_alipay:
            cfg.is_alipay === '1' || cfg.is_alipay === true ? '1' : '0',
          is_wxpay: cfg.is_wxpay === '1' || cfg.is_wxpay === true ? '1' : '0',
          is_qqpay: cfg.is_qqpay === '1' || cfg.is_qqpay === true ? '1' : '0',
        });
      } catch {}
    }
  } catch (e: any) {
    message.error(e?.message || '加载失败');
  } finally {
    loading.value = false;
  }
}

async function handleSave() {
  if (form.pay_mode === 'merchant_epay') {
    if (!form.epay_api.trim()) {
      message.warning('请填写易支付 API 地址');
      return;
    }
    if (!form.epay_pid.trim()) {
      message.warning('请填写商户 PID');
      return;
    }
    if (!form.epay_key.trim()) {
      message.warning('请填写商户密钥');
      return;
    }
    if (
      form.is_alipay !== '1' &&
      form.is_wxpay !== '1' &&
      form.is_qqpay !== '1'
    ) {
      message.warning('请至少开启一种支付方式');
      return;
    }
  }
  saving.value = true;
  try {
    const payload = form.pay_mode === 'site_balance'
      ? { pay_mode: 'site_balance' }
      : { ...form };
    await saveTenantPayConfigApi(JSON.stringify(payload));
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
  <Page title="支付配置" content-class="p-4">
    <Alert
      type="info"
      show-icon
      class="mb-4"
      :message="isSiteBalanceMode ? '默认使用站长代收。C端支付成功后，货款先进入您的主站余额，再自动扣供货价下单，利润会留在余额中。' : '商家直连模式下，C端用户购买商品时将跳转到您配置的易支付页面完成付款。'"
      style="max-width: 600px"
    />

    <Card :loading="loading" style="max-width: 600px">
      <div class="space-y-5">
        <div>
          <div class="mb-3 flex items-center gap-2">
            <span class="text-base font-semibold">收款模式</span>
          </div>

          <RadioGroup v-model:value="form.pay_mode" button-style="solid">
            <RadioButton value="site_balance">站长代收</RadioButton>
            <RadioButton value="merchant_epay">商家直连</RadioButton>
          </RadioGroup>

          <div class="mt-3 text-xs text-gray-400 dark:text-gray-500">
            站长代收为默认模式；若切换为商家直连，则由您自己的易支付接口完成收款。
          </div>
        </div>

        <Divider />

        <div v-if="!isSiteBalanceMode">
          <div class="mb-3 flex items-center gap-2">
            <KeyOutlined />
            <span class="text-base font-semibold">易支付接口配置</span>
          </div>

          <div class="space-y-4">
            <div>
              <label class="mb-1 block text-sm font-medium">
                API 地址 <span class="text-red-500">*</span>
              </label>
              <Input
                v-model:value="form.epay_api"
                placeholder="https://your-epay.com/"
                allow-clear
              />
              <div class="mt-1 text-xs text-gray-400 dark:text-gray-500">
                易支付站点地址，末尾带斜杠
              </div>
            </div>

            <div>
              <label class="mb-1 block text-sm font-medium">
                商户 PID <span class="text-red-500">*</span>
              </label>
              <Input
                v-model:value="form.epay_pid"
                placeholder="请输入商户 PID"
                allow-clear
              />
            </div>

            <div>
              <label class="mb-1 block text-sm font-medium">
                商户密钥 <span class="text-red-500">*</span>
              </label>
              <Input.Password
                v-model:value="form.epay_key"
                placeholder="请输入商户密钥"
                allow-clear
              />
            </div>
          </div>
        </div>

        <div v-else>
          <Alert
            type="success"
            show-icon
            message="当前为站长代收模式"
            description="商城实付金额会先充入您的主站余额，再按供货价自动扣款下单。您无需单独维护易支付参数。"
          />
        </div>

        <Divider v-if="!isSiteBalanceMode" />

        <div v-if="!isSiteBalanceMode">
          <div class="mb-3 text-base font-semibold">对外展示支付方式</div>
          <div class="mb-3 text-xs text-gray-400 dark:text-gray-500">
            这里只控制商家直连模式下商城对外开放哪些支付渠道。
          </div>
          <div class="space-y-3">
            <div
              class="flex items-center justify-between rounded-lg bg-gray-50 dark:bg-gray-800 px-3 py-2"
            >
              <div class="flex items-center gap-2">
                <AlipayCircleOutlined style="font-size:24px;color:#1677ff" />
                <span class="text-sm font-medium">支付宝</span>
              </div>
              <Switch
                :checked="form.is_alipay === '1'"
                checked-children="开"
                un-checked-children="关"
                @change="(v: boolean) => (form.is_alipay = v ? '1' : '0')"
              />
            </div>

            <div
              class="flex items-center justify-between rounded-lg bg-gray-50 dark:bg-gray-800 px-3 py-2"
            >
              <div class="flex items-center gap-2">
                <WechatOutlined style="font-size:24px;color:#07c160" />
                <span class="text-sm font-medium">微信支付</span>
              </div>
              <Switch
                :checked="form.is_wxpay === '1'"
                checked-children="开"
                un-checked-children="关"
                @change="(v: boolean) => (form.is_wxpay = v ? '1' : '0')"
              />
            </div>

            <div
              class="flex items-center justify-between rounded-lg bg-gray-50 dark:bg-gray-800 px-3 py-2"
            >
              <div class="flex items-center gap-2">
                <QqOutlined style="font-size:24px;color:#12b7f5" />
                <span class="text-sm font-medium">QQ 钱包</span>
              </div>
              <Switch
                :checked="form.is_qqpay === '1'"
                checked-children="开"
                un-checked-children="关"
                @change="(v: boolean) => (form.is_qqpay = v ? '1' : '0')"
              />
            </div>
          </div>
        </div>

        <div v-else>
          <Alert
            type="info"
            show-icon
            message="平台代收模式下支付方式自动读取平台配置"
            description="当前商城展示的支付宝、微信支付、QQ支付渠道会直接读取平台管理员在系统设置中的支付开关，商家无需单独配置。"
          />
        </div>

        <Divider />

        <div>
          <Button type="primary" :loading="saving" @click="handleSave">
            <template #icon><SaveOutlined /></template>
            保存配置
          </Button>
        </div>
      </div>
    </Card>
  </Page>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card,
  Button,
  Input,
  Switch,
  message,
  Divider,
  Alert,
} from 'ant-design-vue';
import { SaveOutlined, KeyOutlined, AlipayCircleOutlined, WechatOutlined, QqOutlined } from '@ant-design/icons-vue';
import { getTenantShopApi, saveTenantPayConfigApi } from '#/api/tenant';

const loading = ref(false);
const saving = ref(false);

interface PayConfig {
  epay_api: string;
  epay_pid: string;
  epay_key: string;
  is_alipay: string;
  is_wxpay: string;
  is_qqpay: string;
}

const form = reactive<PayConfig>({
  epay_api: '',
  epay_pid: '',
  epay_key: '',
  is_alipay: '1',
  is_wxpay: '0',
  is_qqpay: '0',
});

async function load() {
  loading.value = true;
  try {
    const res = await getTenantShopApi();
    const data = res;
    if (data?.pay_config) {
      try {
        const cfg = JSON.parse(data.pay_config);
        Object.assign(form, {
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
  saving.value = true;
  try {
    await saveTenantPayConfigApi(JSON.stringify(form));
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
      message="配置易支付接口，C端用户购买商品时将跳转到对应支付页面完成付款。"
      style="max-width: 600px"
    />

    <Card :loading="loading" style="max-width: 600px">
      <div class="space-y-5">
        <div>
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
              <div class="mt-1 text-xs text-gray-400">
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

        <Divider />

        <div>
          <div class="mb-3 text-base font-semibold">开启支付方式</div>
          <div class="space-y-3">
            <div
              class="flex items-center justify-between rounded-lg bg-gray-50 px-3 py-2"
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
              class="flex items-center justify-between rounded-lg bg-gray-50 px-3 py-2"
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
              class="flex items-center justify-between rounded-lg bg-gray-50 px-3 py-2"
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

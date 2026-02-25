<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import { Card, Button, Input, Textarea, Switch, message, Alert, Modal, Spin } from 'ant-design-vue';
import { SaveOutlined, ShopOutlined, UnlockOutlined } from '@ant-design/icons-vue';
import { getTenantShopApi, saveTenantShopApi, type TenantInfo } from '#/api/tenant';
import { requestClient } from '#/api/request';

const loading = ref(false);
const saving = ref(false);
const opened = ref(false);
const openPrice = ref(0);
const opening = ref(false);
const openShopName = ref('');

const form = reactive<Partial<TenantInfo>>({
  tid: 0,
  shop_name: '',
  shop_logo: '',
  shop_desc: '',
  status: 1,
  domain: '',
});

const mallUrl = computed(() => {
  if (!form.tid) return '';
  const base = form.domain ? `https://${form.domain}` : window.location.origin;
  return `${base}/mall/${form.tid}`;
});

async function load() {
  loading.value = true;
  try {
    const res = await getTenantShopApi();
    const data = res;
    Object.assign(form, data);
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
  saving.value = true;
  try {
    await saveTenantShopApi({ ...form });
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
          <Alert type="info" show-icon>
            <template #message>
              <span class="text-sm">我的商城地址：</span>
              <a :href="mallUrl" target="_blank" class="font-medium mx-2">{{ mallUrl }}</a>
              <Button size="small" type="link" @click="() => { navigator.clipboard.writeText(mallUrl); message.success('已复制'); }">复制</Button>
              <Button size="small" type="primary" ghost :href="mallUrl" target="_blank">
                <template #icon><ShopOutlined /></template>
                进入商城
              </Button>
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
              <label class="block text-sm font-medium mb-1">店铺简介</label>
              <Textarea v-model:value="form.shop_desc" :rows="3" placeholder="展示在商城首页的简介内容" />
            </div>
            <div class="flex items-center gap-3">
              <label class="text-sm font-medium">店铺状态</label>
              <Switch v-model:checked="form.status" :checked-value="1" :un-checked-value="0" checked-children="开启" un-checked-children="关闭" />
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

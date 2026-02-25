<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Alert, Button, Spin } from 'ant-design-vue';
import { Page } from '@vben/common-ui';
import { getModuleFrameUrl, getModulesByTypeApi, type DynamicModule } from '#/api/module';
import { useAccessStore } from '@vben/stores';

const route = useRoute();
const router = useRouter();
const accessStore = useAccessStore();

const appId = ref(route.params.appId as string);
const moduleType = ref(route.path.split('/')[1] || 'sport');
const moduleName = ref('');
const frameUrl = ref('');
const loading = ref(true);
const error = ref('');

async function loadFrameUrl() {
  loading.value = true;
  error.value = '';
  try {
    // 获取模块名称
    const raw = await getModulesByTypeApi(moduleType.value);
    let list = raw;
    if (Array.isArray(list)) {
      const mod = list.find((m: DynamicModule) => m.app_id === appId.value);
      if (mod) moduleName.value = mod.name;
    }
    // 获取 frame URL
    const urlRaw = await getModuleFrameUrl(appId.value);
    const res = urlRaw;
    if (res?.frame_url) {
      frameUrl.value = res.frame_url;
    } else {
      error.value = '该模块未配置前端页面，请在管理后台设置 view_url';
    }
  } catch (e: any) {
    error.value = e?.message || '获取模块页面失败';
  } finally {
    loading.value = false;
  }
}

function retry() {
  frameUrl.value = '';
  loadFrameUrl();
}

function goBack() {
  router.push(`/${moduleType.value}/hub`);
}

function onIframeLoad() {
  loading.value = false;
}

// 监听 iframe postMessage（余额变更等）
function onMessage(event: MessageEvent) {
  if (!event.data || typeof event.data !== 'object') return;
  if (event.data.type === 'balance_changed') {
    // 刷新用户信息（余额）
    accessStore.setAccessToken(accessStore.accessToken);
  }
}

onMounted(() => {
  loadFrameUrl();
  window.addEventListener('message', onMessage);
});

onUnmounted(() => {
  window.removeEventListener('message', onMessage);
});
</script>

<template>
  <Page auto-content-height content-class="p-0">
    <div class="flex h-full flex-col">
      <!-- 顶部工具栏 -->
      <div class="flex items-center gap-3 px-4 py-2 border-b border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-[#1f1f1f]">
        <Button size="small" @click="goBack">返回大厅</Button>
        <span v-if="moduleName" class="font-medium">{{ moduleName }}</span>
        <Button v-if="error" size="small" type="link" @click="retry">重试</Button>
      </div>

      <!-- 加载中 -->
      <Spin v-if="loading && !frameUrl" class="flex flex-1 items-center justify-center" />

      <!-- 错误提示 -->
      <Alert
        v-if="error"
        :message="error"
        type="warning"
        show-icon
        class="m-4"
      >
        <template #action>
          <Button size="small" type="primary" @click="retry">重新加载</Button>
        </template>
      </Alert>

      <!-- iframe -->
      <iframe
        v-if="frameUrl"
        :src="frameUrl"
        class="h-full w-full flex-1 border-none"
        allow="clipboard-write"
        @load="onIframeLoad"
      />
    </div>
  </Page>
</template>

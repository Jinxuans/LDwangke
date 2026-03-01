<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { getExtMenusPublicApi } from '#/api/ext-menu';

const route = useRoute();
const iframeSrc = ref('');
const pageTitle = ref('扩展页面');
const loading = ref(true);

onMounted(async () => {
  const id = Number(route.params.id);
  if (!id) { loading.value = false; return; }
  try {
    const list = await getExtMenusPublicApi();
    const item = list?.find((m) => m.id === id);
    if (item) {
      iframeSrc.value = item.url;
      pageTitle.value = item.title;
      document.title = item.title;
    }
  } catch { /* ignore */ }
  loading.value = false;
});
</script>

<template>
  <div class="iframe-container">
    <div v-if="loading" class="flex h-full items-center justify-center text-gray-400">
      加载中...
    </div>
    <iframe
      v-else-if="iframeSrc"
      :src="iframeSrc"
      :title="pageTitle"
      class="iframe-content"
      frameborder="0"
      allowfullscreen
    />
    <div v-else class="flex h-full items-center justify-center text-gray-400">
      未配置页面地址
    </div>
  </div>
</template>

<style scoped>
.iframe-container {
  width: 100%;
  height: calc(100vh - 100px);
  overflow: hidden;
}
.iframe-content {
  width: 100%;
  height: 100%;
  border: none;
}
</style>

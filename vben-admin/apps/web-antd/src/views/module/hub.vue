<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Card, Empty, Spin, Input, Tag } from 'ant-design-vue';
import { Page } from '@vben/common-ui';
import { getModulesByTypeApi, type DynamicModule } from '#/api/module';
import { iconRegistry } from '#/components/IconSelect.vue';

const route = useRoute();
const router = useRouter();

const typeLabels: Record<string, string> = {
  sport: '运动大厅',
  intern: '实习大厅',
  paper: '论文大厅',
};

const moduleType = computed(() => route.path.split('/')[1] || 'sport');
const title = computed(() => typeLabels[moduleType.value] || '模块大厅');

const loading = ref(false);
const modules = ref<DynamicModule[]>([]);
const searchKeyword = ref('');

async function loadModules() {
  loading.value = true;
  try {
    const raw = await getModulesByTypeApi(moduleType.value);
    let list = raw;
    if (!Array.isArray(list)) list = [];
    modules.value = list;
  } catch {
    modules.value = [];
  } finally {
    loading.value = false;
  }
}

const availableModules = computed(() => {
  return modules.value.filter(m => m.view_url && m.view_url.trim() !== '');
});

const filteredModules = computed(() => {
  const kw = searchKeyword.value.trim().toLowerCase();
  if (!kw) return availableModules.value;
  return availableModules.value.filter(m =>
    m.name.toLowerCase().includes(kw) ||
    (m.description && m.description.toLowerCase().includes(kw)),
  );
});

// 旧 lucide:xxx 格式 → 新 key 的映射（兼容历史数据）
const legacyKeyMap: Record<string, string> = {
  'lucide:zap': 'thunderbolt', 'lucide:globe': 'global', 'lucide:cloud': 'cloud',
  'lucide:heart-pulse': 'heart', 'lucide:footprints': 'dashboard', 'lucide:activity': 'dashboard',
  'lucide:bike': 'car', 'lucide:map-pin': 'environment', 'lucide:briefcase': 'solution',
  'lucide:file-text': 'filetext', 'lucide:graduation-cap': 'read', 'lucide:cloud-sun': 'cloud',
};
function resolveIconKey(icon: string): string | undefined {
  if (!icon) return undefined;
  if (iconRegistry[icon]) return icon;
  if (legacyKeyMap[icon]) return legacyKeyMap[icon];
  return undefined;
}

function goDetail(mod: DynamicModule) {
  router.push(`/${moduleType.value}/${mod.app_id}`);
}

onMounted(loadModules);
</script>

<template>
  <Page :title="title" content-class="p-4">
    <Spin :spinning="loading">
      <div v-if="availableModules.length > 3" class="mb-4" style="max-width: 320px;">
        <Input
          v-model:value="searchKeyword"
          placeholder="搜索模块..."
          allow-clear
        />
      </div>

      <Empty v-if="!loading && filteredModules.length === 0" :description="searchKeyword ? '没有找到匹配的模块' : '暂无可用模块'" />

      <div v-else class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
        <Card
          v-for="mod in filteredModules"
          :key="mod.id"
          hoverable
          class="module-card cursor-pointer"
          @click="goDetail(mod)"
        >
          <div class="flex flex-col items-center gap-2 py-4 px-2">
            <component :is="iconRegistry[resolveIconKey(mod.icon) || 'appstore']?.comp" class="text-2xl text-gray-500" />
            <span class="text-base font-semibold">{{ mod.name }}</span>
            <span v-if="mod.description" class="text-xs text-gray-400 text-center" style="line-height: 1.3; max-height: 2.6em; overflow: hidden;">{{ mod.description }}</span>
            <Tag v-if="mod.price" color="orange" style="margin: 0;">{{ mod.price }}</Tag>
          </div>
        </Card>
      </div>
    </Spin>
  </Page>
</template>

<style scoped>
.module-card {
  transition: transform 0.15s, box-shadow 0.15s;
}
.module-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}
</style>

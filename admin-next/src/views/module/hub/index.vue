<template>
  <div class="module-hub-page art-full-height">
    <ElCard class="art-table-card">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b-d px-5 py-4">
        <div class="flex flex-wrap items-center gap-3">
          <h2 class="text-lg font-semibold text-g-900">{{ title }}</h2>
          <ElTag effect="plain">模块类型 {{ moduleType }}</ElTag>
          <ElTag type="success" effect="plain">可用模块 {{ filteredModules.length }}</ElTag>
        </div>

        <div class="flex flex-wrap gap-3">
          <ElButton plain :loading="loading" @click="loadModules">刷新</ElButton>
          <ElButton plain @click="searchKeyword = ''">清空搜索</ElButton>
        </div>
      </div>

      <div class="px-5 py-4">
        <div class="grid gap-4 md:grid-cols-[320px_auto]">
          <div>
            <label class="mb-2 block text-sm font-medium text-g-800">搜索模块</label>
            <ElInput v-model="searchKeyword" clearable placeholder="搜索模块名称或简介" />
          </div>
        </div>
      </div>
    </ElCard>

    <ElCard class="art-table-card mt-5">
      <div v-loading="loading">
        <ElEmpty v-if="!loading && !filteredModules.length" description="暂无可用模块" />

        <div v-else class="grid gap-5 sm:grid-cols-2 xl:grid-cols-4">
          <article
            v-for="item in filteredModules"
            :key="item.id"
            class="cursor-pointer rounded-[24px] border border-[var(--art-card-border)] bg-box p-5 transition duration-200 hover:-translate-y-0.5 hover:shadow-[0_18px_40px_rgba(15,23,42,0.08)]"
            @click="goDetail(item)"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-[var(--el-color-primary-light-9)] text-[var(--el-color-primary)]">
                <Icon :icon="item.icon || 'mdi:puzzle-outline'" class="text-2xl" />
              </div>
              <ElTag v-if="item.price" type="warning" effect="plain">{{ item.price }}</ElTag>
            </div>

            <h3 class="mt-5 text-lg font-semibold text-g-900">{{ item.name }}</h3>
            <p class="mt-2 line-clamp-3 min-h-[72px] text-sm leading-6 text-g-500">
              {{ item.description || '未填写模块说明' }}
            </p>

            <div class="mt-4 flex items-center justify-between text-xs text-g-400">
              <span>App ID {{ item.app_id }}</span>
              <span>排序 {{ item.sort }}</span>
            </div>
          </article>
        </div>
      </div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, ref } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { Icon } from '@iconify/vue'
  import { fetchLegacyModulesByType, type LegacyDynamicModule } from '@/api/legacy/module'

  defineOptions({ name: 'DynamicModuleHubPage' })

  const route = useRoute()
  const router = useRouter()

  const loading = ref(false)
  const modules = ref<LegacyDynamicModule[]>([])
  const searchKeyword = ref('')

  const typeLabels: Record<string, string> = {
    sport: '运动大厅',
    intern: '实习大厅',
    paper: '论文大厅'
  }

  const moduleType = computed(() => String(route.path.split('/')[1] || 'sport'))
  const title = computed(() => typeLabels[moduleType.value] || '模块大厅')
  const filteredModules = computed(() => {
    const keyword = searchKeyword.value.trim().toLowerCase()
    const available = modules.value.filter((item) => item.view_url?.trim())
    if (!keyword) {
      return available
    }
    return available.filter((item) => {
      const name = item.name?.toLowerCase() || ''
      const desc = item.description?.toLowerCase() || ''
      return name.includes(keyword) || desc.includes(keyword)
    })
  })

  const loadModules = async () => {
    loading.value = true
    try {
      const result = await fetchLegacyModulesByType(moduleType.value)
      modules.value = Array.isArray(result) ? result : []
    } catch {
      modules.value = []
    } finally {
      loading.value = false
    }
  }

  const goDetail = (item: LegacyDynamicModule) => {
    router.push(`/${moduleType.value}/${item.app_id}`)
  }

  onMounted(() => {
    loadModules()
  })
</script>

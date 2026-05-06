<template>
  <div class="module-frame-page art-full-height">
    <div class="flex h-full flex-col rounded-[24px] border border-[var(--art-card-border)] bg-box">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b border-[var(--art-card-border)] px-5 py-4">
        <div>
          <p class="text-base font-semibold text-g-900">{{ moduleName || '模块详情' }}</p>
          <p class="mt-1 text-sm text-g-500">当前模块类型：{{ moduleType }} / App ID：{{ appId }}</p>
        </div>
        <div class="flex flex-wrap gap-3">
          <ElButton plain @click="goBack">返回大厅</ElButton>
          <ElButton v-if="error" plain @click="loadFrameUrl">重新加载</ElButton>
        </div>
      </div>

      <div v-if="loading && !frameUrl" class="flex flex-1 items-center justify-center">
        <ElEmpty description="模块页面加载中" />
      </div>

      <div v-else-if="error" class="p-5">
        <ElAlert :title="error" type="warning" show-icon :closable="false" />
      </div>

      <iframe
        v-if="frameUrl"
        :src="frameUrl"
        class="min-h-0 flex-1 border-0"
        allow="clipboard-write"
        @load="loading = false"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, ref, watch } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import { fetchLegacyModuleFrameUrl, fetchLegacyModulesByType } from '@/api/legacy/module'

  defineOptions({ name: 'DynamicModuleFramePage' })

  const route = useRoute()
  const router = useRouter()

  const appId = ref(String(route.params.appId || ''))
  const moduleType = ref(String(route.path.split('/')[1] || 'sport'))
  const moduleName = ref('')
  const frameUrl = ref('')
  const loading = ref(true)
  const error = ref('')

  const loadFrameUrl = async () => {
    loading.value = true
    error.value = ''
    frameUrl.value = ''

    try {
      const list = await fetchLegacyModulesByType(moduleType.value)
      const current = Array.isArray(list) ? list.find((item) => item.app_id === appId.value) : undefined
      moduleName.value = current?.name || ''

      const result = await fetchLegacyModuleFrameUrl(appId.value)
      if (!result?.frame_url) {
        error.value = '该模块未配置前端页面地址，请先到后台模块管理中完善 view_url。'
        return
      }
      frameUrl.value = result.frame_url
    } catch (err: any) {
      error.value = err?.message || '获取模块页面失败'
    } finally {
      loading.value = false
    }
  }

  const goBack = () => {
    router.push(`/${moduleType.value}/hub`)
  }

  watch(
    () => route.fullPath,
    () => {
      appId.value = String(route.params.appId || '')
      moduleType.value = String(route.path.split('/')[1] || 'sport')
      loadFrameUrl()
    }
  )

  onMounted(() => {
    loadFrameUrl()
  })
</script>

<template>
  <div class="flex h-full min-h-[50vh] items-center justify-center p-6">
    <div class="art-card-sm w-full max-w-2xl p-8">
      <p class="mb-2 text-sm font-medium text-[var(--art-gray-500)]">页面提示</p>
      <h1 class="mb-3 text-3xl font-semibold text-[var(--art-text-gray-900)]">
        {{ title }}
      </h1>
      <p class="mb-8 leading-7 text-[var(--art-gray-600)]">
        {{ description }}
      </p>

      <div class="flex flex-wrap gap-3">
        <ElButton type="primary" @click="router.push({ name: 'Login' })">返回登录</ElButton>
        <ElButton @click="router.back()">返回上一页</ElButton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  defineOptions({ name: 'LegacyRoutePlaceholder' })

  const route = useRoute()
  const router = useRouter()

  const title = computed(() => (route.meta.title as string) || '页面暂未开放')
  const description = computed(
    () =>
      (route.meta.description as string) ||
      '当前页面暂未开放，请先返回上一页或联系管理员确认入口状态。'
  )
</script>

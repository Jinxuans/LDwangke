<template>
  <div class="flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <section class="art-card-sm overflow-hidden">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b-d px-5 py-4">
        <div class="flex flex-wrap items-center gap-3">
          <h2 class="text-lg font-semibold text-g-900">实习助手</h2>
          <ElTag effect="plain">嵌入模式</ElTag>
        </div>

        <div class="flex flex-wrap gap-3">
          <ElButton plain @click="reloadFrame">刷新</ElButton>
          <ElButton plain @click="openInNewTab">新标签打开</ElButton>
        </div>
      </div>

      <iframe :key="frameKey" class="min-h-[calc(100vh-260px)] w-full border-0 bg-white" :src="frameSrc" title="实习助手" />
    </section>
  </div>
</template>

<script setup lang="ts">
  defineOptions({ name: 'PluginSXZSPage' })

  const frameSrc = '/php-api/sxzsredis/sxzs.php'
  const frameKey = ref(0)

  const reloadFrame = () => {
    frameKey.value += 1
  }

  const openInNewTab = () => {
    window.open(frameSrc, '_blank', 'noopener,noreferrer')
  }
</script>

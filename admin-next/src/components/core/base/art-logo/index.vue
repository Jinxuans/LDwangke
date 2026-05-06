<!-- 系统logo -->
<template>
  <div class="flex-cc">
    <img :style="logoStyle" :src="currentLogo" alt="logo" class="w-full h-full object-contain" />
  </div>
</template>

<script setup lang="ts">
  import defaultLogo from '@imgs/common/logo.webp'
  import { useSiteStore } from '@/store/modules/site'

  defineOptions({ name: 'ArtLogo' })

  interface Props {
    /** logo 大小 */
    size?: number | string
  }

  const props = withDefaults(defineProps<Props>(), {
    size: 36
  })

  const siteStore = useSiteStore()
  const logoStyle = computed(() => ({ width: `${props.size}px` }))
  const currentLogo = computed(() => siteStore.logoUrl || defaultLogo)
</script>

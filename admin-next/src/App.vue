<template>
  <ElConfigProvider
    size="default"
    :locale="locales[language]"
    :z-index="3000"
    :card="{
      shadow: 'never'
    }"
  >
    <RouterView></RouterView>
    <ArtWatermark v-if="route.name !== 'Login'" :content="watermarkContent" />
  </ElConfigProvider>
</template>

<script setup lang="ts">
  import { useSettingStore } from './store/modules/setting'
  import { useUserStore } from './store/modules/user'
  import { useSiteStore } from './store/modules/site'
  import zh from 'element-plus/es/locale/lang/zh-cn'
  import en from 'element-plus/es/locale/lang/en'
  import { systemUpgrade } from './utils/sys'
  import { toggleTransition } from './utils/ui/animation'
  import { checkStorageCompatibility } from './utils/storage'
  import { initializeTheme } from './hooks/core/useTheme'

  const settingStore = useSettingStore()
  const userStore = useUserStore()
  const siteStore = useSiteStore()
  const route = useRoute()
  const { language } = storeToRefs(userStore)
  const { systemName } = storeToRefs(siteStore)

  const locales = {
    zh: zh,
    en: en
  }

  const watermarkContent = computed(() => {
    const userInfo = userStore.getUserInfo
    const username = userInfo.username || userInfo.userName || ''
    const userId = userInfo.userId ? String(userInfo.userId) : ''

    return [systemName.value, username, userId].filter(Boolean)
  })

  onBeforeMount(() => {
    toggleTransition(true)
    settingStore.applyVisualBaseline()
    initializeTheme()
    siteStore.initPublicConfig(true)
  })

  onMounted(() => {
    checkStorageCompatibility()
    toggleTransition(false)
    systemUpgrade()
  })
</script>

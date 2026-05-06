import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import AppConfig from '@/config'
import { fetchLegacySiteConfig } from '@/api/legacy/site'
import type { LegacySiteConfig } from '@/types/legacy-contract'
import { useSettingStore } from './setting'

function getConfiguredHost() {
  const { VITE_API_PROXY_URL, VITE_API_URL } = import.meta.env
  const candidates = [VITE_API_PROXY_URL, VITE_API_URL]

  for (const candidate of candidates) {
    if (!candidate) {
      continue
    }

    if (/^https?:\/\//i.test(candidate)) {
      try {
        return new URL(candidate).origin
      } catch {
        continue
      }
    }
  }

  return window.location.origin
}

function resolveAssetUrl(url = '') {
  const trimmed = url.trim()
  if (!trimmed) {
    return ''
  }

  if (/^https?:\/\//i.test(trimmed)) {
    return trimmed
  }

  if (trimmed.startsWith('//')) {
    return `${window.location.protocol}${trimmed}`
  }

  const baseHost = getConfiguredHost()
  const normalizedPath = trimmed.startsWith('/') ? trimmed : `/${trimmed}`
  return `${baseHost}${normalizedPath}`
}

function updateFavicon(url: string) {
  if (!url || typeof document === 'undefined') {
    return
  }

  const linkId = 'dynamic-site-favicon'
  let link =
    (document.querySelector(`link#${linkId}`) as HTMLLinkElement | null) ||
    (document.querySelector('link[rel="icon"]') as HTMLLinkElement | null)

  if (!link) {
    link = document.createElement('link')
    link.rel = 'icon'
    document.head.appendChild(link)
  }

  link.id = linkId
  link.href = url
}

function upsertMetaTag(name: string, content: string) {
  if (typeof document === 'undefined' || !content) {
    return
  }

  let element = document.querySelector(`meta[name="${name}"]`) as HTMLMetaElement | null
  if (!element) {
    element = document.createElement('meta')
    element.name = name
    document.head.appendChild(element)
  }

  element.content = content
}

export const useSiteStore = defineStore(
  'siteStore',
  () => {
    const settingStore = useSettingStore()
    const config = ref<LegacySiteConfig>({})
    const loaded = ref(false)
    const loading = ref(false)

    const systemName = computed(() => config.value.sitename || AppConfig.systemInfo.name)
    const systemDescription = computed(
      () => config.value.description || AppConfig.systemInfo.description || ''
    )
    const logoUrl = computed(() => resolveAssetUrl(config.value.hlogo || config.value.logo))
    const faviconUrl = computed(() => resolveAssetUrl(config.value.logo || config.value.hlogo))

    const applyDocumentBranding = () => {
      if (typeof document === 'undefined') {
        return
      }

      if (!document.title) {
        document.title = systemName.value
      } else if (document.title.includes(' - ')) {
        const [pageTitle] = document.title.split(' - ')
        document.title = `${pageTitle} - ${systemName.value}`
      } else if (document.title === AppConfig.systemInfo.name) {
        document.title = systemName.value
      }

      updateFavicon(faviconUrl.value)
      upsertMetaTag('keywords', config.value.keywords || '')
      upsertMetaTag('description', systemDescription.value)
    }

    const applyWatermarkSetting = () => {
      settingStore.setWatermarkVisible(config.value.sykg === '1')
    }

    const setConfig = (nextConfig: LegacySiteConfig) => {
      config.value = nextConfig
      applyDocumentBranding()
      applyWatermarkSetting()
    }

    const initPublicConfig = async (force = false) => {
      if (loading.value || (loaded.value && !force)) {
        return
      }

      loading.value = true

      try {
        const result = await fetchLegacySiteConfig()
        setConfig(result || {})
        loaded.value = true
      } catch (error) {
        console.warn('[SiteStore] Failed to fetch public site config:', error)
      } finally {
        loading.value = false
      }
    }

    return {
      config,
      loaded,
      loading,
      systemName,
      systemDescription,
      logoUrl,
      faviconUrl,
      setConfig,
      initPublicConfig
    }
  },
  {
    persist: {
      key: 'site',
      storage: localStorage
    }
  }
)


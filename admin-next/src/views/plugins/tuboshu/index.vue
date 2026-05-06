<template>
  <div class="flex min-h-[calc(100vh-180px)] flex-col gap-5">
    <section class="art-card-sm overflow-hidden">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b-d px-5 py-4">
        <div class="flex flex-wrap items-center gap-3">
          <h2 class="text-lg font-semibold text-g-900">土拨鼠论文</h2>
          <ElTag type="success" effect="plain">代理运行中</ElTag>
        </div>

        <ElButton plain :loading="loading || reloading" @click="reloadIframe">刷新</ElButton>
      </div>

      <div v-if="loading" class="flex min-h-[480px] items-center justify-center">
        <ElSkeleton animated :rows="6" class="max-w-[520px] px-6" />
      </div>
      <iframe
        ref="iframeRef"
        class="min-h-[calc(100vh-260px)] w-full border-0 transition-opacity duration-300"
        :class="loading ? 'opacity-0' : 'opacity-100'"
        frameborder="0"
        title="土拨鼠论文"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'PluginTuboshuPage' })

  const { VITE_API_URL } = import.meta.env

  const userStore = useUserStore()
  const loading = ref(true)
  const reloading = ref(false)
  const iframeRef = ref<HTMLIFrameElement>()
  let currentBlobUrl = ''

  const apiBase = String(VITE_API_URL || '').replace(/\/$/, '')

  const getAuthToken = () => {
    const token = userStore.accessToken || ''
    return token ? (token.startsWith('Bearer ') ? token : `Bearer ${token}`) : ''
  }

  const revokeCurrentBlob = () => {
    if (currentBlobUrl) {
      URL.revokeObjectURL(currentBlobUrl)
      currentBlobUrl = ''
    }
  }

  const handleIframeMessage = (event: MessageEvent) => {
    const data = event.data
    if (!data?.type) return

    if (data.type === 'tuboshu_loaded') {
      loading.value = false
      reloading.value = false
      return
    }

    if (data.type !== 'tuboshu_api_request') return

    const { payload, requestId } = data
    const sendResponse = (result: any) => {
      iframeRef.value?.contentWindow?.postMessage(
        {
          payload: result,
          requestId,
          type: 'tuboshu_api_response'
        },
        '*'
      )
    }

    const isBlobResponse = payload?.specialConfig?.responseType === 'blob'

    fetch(`${apiBase}/tuboshu/route`, {
      body: JSON.stringify({
        isBlob: isBlobResponse,
        method: payload.method || 'GET',
        params: payload.data,
        path: payload.url
      }),
      headers: {
        Authorization: getAuthToken(),
        'Content-Type': 'application/json'
      },
      method: 'POST'
    })
      .then((response) => (isBlobResponse ? response.blob() : response.json()))
      .then((result) => sendResponse(result))
      .catch((error) => sendResponse({ message: error?.message || '请求失败', success: false }))
  }

  const loadConfig = async () => {
    try {
      const response = await fetch(`${apiBase}/tuboshu/config`, {
        headers: { Authorization: getAuthToken() }
      })
      const json = await response.json()
      const config = json?.data || json || {}
      return {
        priceConfig: {
          ...(config?.price_config || {}),
          PAGE_VISIBILITY: config?.page_visibility || {}
        },
        priceRatio: config?.user_price_ratio || config?.price_ratio || 5
      }
    } catch {
      return {
        priceConfig: {},
        priceRatio: 5
      }
    }
  }

  const buildIframeHtml = (priceConfig: Record<string, any>, priceRatio: number) => {
    const authToken = getAuthToken()
    const routeFormDataUrl = `${apiBase}/tuboshu/route-formdata`

    return `<!DOCTYPE html>
<html lang="zh">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width,initial-scale=1.0" />
    <link href="https://unpkg.com/quasar@2.10.1/dist/quasar.prod.css" rel="stylesheet" />
    <link rel="stylesheet" href="https://unpkg.com/@afyercu/tuboshu-components@0.0.28/dist/lib/tuboshu-components.css" />
    <script type="importmap">
      {"imports":{"vue":"https://unpkg.com/vue@3.5.13/dist/vue.esm-browser.prod.js","tuboshu-components":"https://unpkg.com/@afyercu/tuboshu-components@0.0.28/dist/lib/index.es.js"}}
    <\/script>
    <style>
      body { margin: 0; background: transparent; }
      .container { max-width: 1200px; margin: 0 auto; padding: 18px; }
      p { margin: 0; padding: 0; }
      @media (max-width: 768px) {
        .container { padding: 0; }
        .tabs-container { overflow-x: auto; padding: 10px; width: 100%; }
      }
    </style>
  </head>
  <body>
    <div id="app" style="display:none">
      <subsite-provider :config="priceConfig" :price-ratio="priceRatio" :is-admin="false">
        <div class="container">
          <tabs v-model="currentTab">
            <tabs-list class="w-full tabs-container">
              <tabs-trigger
                v-for="([key,page],index) in Object.entries(pages).filter(([key,page]) => !hidePage.includes(key)).filter(([,page]) => page.__name !== 'SubsitePage')"
                :key="index"
                :value="page.__name"
              >
                {{ pageNameMap[page.__name] || page.__name }}
              </tabs-trigger>
            </tabs-list>
            <tabs-content v-for="(page,index) in pages" :key="index" :value="page.__name" style="min-height:calc(100vh - 60px)">
              <component :is="page"></component>
            </tabs-content>
          </tabs>
        </div>
      </subsite-provider>
    </div>
    <script>window.process={env:{NODE_ENV:'production'}}<\/script>
    <script src="https://code.iconify.design/1/1.0.6/iconify.min.js" defer><\/script>
    <script type="module">
      import { createApp, shallowRef } from 'vue'
      import { setupVueApp } from 'tuboshu-components'
      import * as TuBoShu from 'tuboshu-components'

      const priceConfig = ${JSON.stringify(priceConfig)}
      const pages = shallowRef(TuBoShu.pages)
      const pageVisibility = priceConfig.PAGE_VISIBILITY || {}
      const hidePage = Object.keys(TuBoShu.pages).filter((key) => pageVisibility.hasOwnProperty(key) && pageVisibility[key] === false)
      const pageNameMap = {
        AccountTable: '生成记录',
        ChartPage: '图表生成',
        ChatPage: '简单订单',
        ComponentStagePage: '订单提交',
        KnowledgeBasePage: '知识库',
        PointsExchangePage: '点数兑换',
        ReductionPage: '降AIGC/降重',
        SubsitePage: '价格设置',
        TemplatePage: 'Word模板生成',
        ThesisGeneratePage: '智码方舟',
        TicketPage: '工单反馈'
      }
      const authToken = ${JSON.stringify(authToken)}
      const routeFormDataUrl = ${JSON.stringify(routeFormDataUrl)}
      const pendingMap = new Map()

      window.addEventListener('message', (event) => {
        if (event.data?.type !== 'tuboshu_api_response') return
        const callback = pendingMap.get(event.data.requestId)
        if (callback) {
          pendingMap.delete(event.data.requestId)
          callback(event.data.payload)
        }
      })

      const app = createApp({
        data() {
          return {
            currentTab: 'ComponentStagePage',
            pageNameMap: { ...TuBoShu.pageNameMap, ...pageNameMap },
            priceConfig,
            priceRatio: Number(${JSON.stringify(priceRatio)})
          }
        },
        methods: {
          handleApiRequest(event) {
            const { payload } = event.detail
            const respond = (requestId, result) => {
              window.dispatchEvent(new CustomEvent('api_response', { detail: { payload: result, requestId } }))
            }

            if (payload.specialConfig?.type === 'formData') {
              payload.data.append('path', payload.url)
              payload.data.append('method', payload.method)
              fetch(routeFormDataUrl, {
                method: 'POST',
                body: payload.data,
                headers: { Authorization: authToken }
              })
                .then((response) => response.json())
                .then((result) => respond(payload.id, result))
                .catch((error) => respond(payload.id, { message: error?.message || '请求失败', success: false }))
              return
            }

            const requestId = payload.id || ('r' + Date.now() + '_' + Math.random())
            const timer = setTimeout(() => {
              pendingMap.delete(requestId)
              respond(payload.id, { message: 'timeout', success: false })
            }, 120000)

            pendingMap.set(requestId, (result) => {
              clearTimeout(timer)
              respond(payload.id, result)
            })

            window.parent.postMessage({
              payload: {
                data: payload.data,
                method: payload.method || 'GET',
                specialConfig: payload.specialConfig,
                url: payload.url
              },
              requestId,
              type: 'tuboshu_api_request'
            }, '*')
          }
        },
        beforeMount() {
          window.addEventListener('api_request', this.handleApiRequest)
        },
        mounted() {
          document.getElementById('app').style.display = 'block'
          window.parent.postMessage({ type: 'tuboshu_loaded' }, '*')
        },
        unmounted() {
          window.removeEventListener('api_request', this.handleApiRequest)
        },
        setup() {
          return { hidePage, pages }
        }
      })

      setupVueApp(app)
    <\/script>
  </body>
</html>`
  }

  const mountIframe = async () => {
    loading.value = true
    const { priceConfig, priceRatio } = await loadConfig()
    const html = buildIframeHtml(priceConfig, priceRatio)
    revokeCurrentBlob()
    currentBlobUrl = URL.createObjectURL(new Blob([html], { type: 'text/html' }))
    if (iframeRef.value) {
      iframeRef.value.src = currentBlobUrl
    }
    setTimeout(() => {
      if (loading.value) {
        loading.value = false
        reloading.value = false
      }
    }, 15000)
  }

  const reloadIframe = async () => {
    reloading.value = true
    await mountIframe()
  }

  onMounted(() => {
    window.addEventListener('message', handleIframeMessage)
    mountIframe()
  })

  onBeforeUnmount(() => {
    window.removeEventListener('message', handleIframeMessage)
    revokeCurrentBlob()
  })
</script>

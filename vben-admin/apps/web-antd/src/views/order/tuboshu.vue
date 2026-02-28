<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
import { Page } from '@vben/common-ui';
import { Spin } from 'ant-design-vue';
import { useAccessStore } from '@vben/stores';
import { useAppConfig } from '@vben/hooks';

const accessStore = useAccessStore();
const isAdmin = computed(() => {
  const codes = accessStore.accessCodes;
  return codes.includes('super') || codes.includes('admin');
});

const loading = ref(true);
const iframeRef = ref<HTMLIFrameElement>();
const { apiURL } = useAppConfig(import.meta.env, import.meta.env.PROD);

function getAuthToken() {
  return accessStore.accessToken ? `Bearer ${accessStore.accessToken}` : '';
}

// 父页面：代理 iframe 的 API 请求 + 监听加载完成
function handleIframeMessage(event: MessageEvent) {
  const { data } = event;
  if (!data?.type) return;

  if (data.type === 'tuboshu_loaded') {
    loading.value = false;
    return;
  }

  if (data.type !== 'tuboshu_api_request') return;

  const { requestId, payload } = data;
  const routeUrl = `${apiURL}/tuboshu/route`;

  const sendResponse = (result: any) => {
    iframeRef.value?.contentWindow?.postMessage(
      { type: 'tuboshu_api_response', requestId, payload: result },
      '*',
    );
  };

  const isBlobResponse = payload.specialConfig?.responseType === 'blob';

  fetch(routeUrl, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': getAuthToken(),
    },
    body: JSON.stringify({
      method: payload.method || 'GET',
      path: payload.url,
      params: payload.data,
      isBlob: isBlobResponse,
    }),
  })
    .then((res) => (isBlobResponse ? res.blob() : res.json()))
    .then((result) => sendResponse(result))
    .catch((error) => sendResponse({ success: false, message: error.message }));
}

async function loadConfig() {
  try {
    const resp = await fetch(`${apiURL}/tuboshu/config`, {
      headers: { 'Authorization': getAuthToken() },
    });
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
    const json = await resp.json();
    const cfg = json?.data ?? json;
    return {
      priceConfig: {
        ...(cfg?.price_config ?? {}),
        PAGE_VISIBILITY: cfg?.page_visibility ?? {},
      },
      priceRatio: cfg?.user_price_ratio ?? cfg?.price_ratio ?? 5,
    };
  } catch {
    return { priceConfig: {}, priceRatio: 5 };
  }
}

function buildIframeHTML(priceConfig: any, priceRatio: number) {
  const admin = isAdmin.value;
  const formDataUrl = `${window.location.origin}${apiURL}/tuboshu/route-formdata`;
  const authToken = getAuthToken();

  return `<!DOCTYPE html>
<html lang="zh"><head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<link href="https://unpkg.com/quasar@2.10.1/dist/quasar.prod.css" rel="stylesheet">
<link rel="stylesheet" href="https://unpkg.com/@afyercu/tuboshu-components@0.0.28/dist/lib/tuboshu-components.css">
<script type="importmap">{"imports":{"vue":"https://unpkg.com/vue@3/dist/vue.esm-browser.js","tuboshu-components":"https://unpkg.com/@afyercu/tuboshu-components@0.0.28/dist/lib/index.es.js"}}<\/script>
<style>body{margin:0;background:transparent}.container{max-width:1200px;margin:0 auto;padding:20px}p{padding:0;margin:0}@media(max-width:768px){.container{padding:0}.tabs-container{padding:10px;width:100%;overflow-x:auto}}</style>
</head><body>
<div id="app" style="display:none">
  <subsite-provider :config="priceConfig" :price-ratio="priceRatio" :is-admin="isAdmin">
    <div class="container">
      <tabs v-model="currentTab">
        <tabs-list class="w-full tabs-container">
          <tabs-trigger v-for="([,page],index) in Object.entries(pages).filter(([k,p])=>!hidePage.includes(k)).filter(([,p])=>p.__name!=='SubsitePage'||showPriceConfig)" :key="index" :value="page.__name">
            {{pageNameMap[page.__name]||page.__name}}
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
import{createApp,shallowRef}from'vue'
import{setupVueApp}from'tuboshu-components'
import*as TuBoShu from'tuboshu-components'

const priceConfig=${JSON.stringify(priceConfig)};
const pages=shallowRef(TuBoShu.pages);
const pv=priceConfig.PAGE_VISIBILITY||{};
const hidePage=Object.keys(TuBoShu.pages).filter(p=>pv.hasOwnProperty(p)&&pv[p]===false);
const pageNameMap={ComponentStagePage:'订单提交',ChatPage:'简单订单',ChartPage:'图表生成',TemplatePage:'Word模板生成',ReductionPage:'降AIGC/降重',AccountTable:'生成记录',SubsitePage:'价格设置',TicketPage:'工单反馈',PointsExchangePage:'点数兑换',ThesisGeneratePage:'智码方舟',KnowledgeBasePage:'知识库'};

const authToken='${authToken}';
const formDataUrl='${formDataUrl}';
const pending=new Map();

window.addEventListener('message',e=>{
  if(e.data?.type==='tuboshu_api_response'){
    const fn=pending.get(e.data.requestId);
    if(fn){pending.delete(e.data.requestId);fn(e.data.payload)}
  }
});

const app=createApp({
  data(){return{currentTab:'ComponentStagePage',showPriceConfig:${admin},isAdmin:${admin},priceConfig,pageNameMap:{...TuBoShu.pageNameMap,...pageNameMap},priceRatio:Number(${priceRatio})}},
  methods:{
    handleApiRequest(event){
      const{payload}=event.detail;
      const respond=(id,result)=>{window.dispatchEvent(new CustomEvent('api_response',{detail:{requestId:id,payload:result}}))};
      const isFormData=payload.specialConfig?.type==='formData';
      if(isFormData){
        payload.data.append('path',payload.url);
        payload.data.append('method',payload.method);
        fetch(formDataUrl,{method:'POST',body:payload.data,headers:{'Authorization':authToken}})
          .then(r=>r.json()).then(r=>respond(payload.id,r))
          .catch(e=>respond(payload.id,{success:false,message:e.message}));
        return;
      }
      const rid=payload.id||('r'+Date.now()+'_'+Math.random());
      const tm=setTimeout(()=>{pending.delete(rid);respond(payload.id,{success:false,message:'timeout'})},120000);
      pending.set(rid,r=>{clearTimeout(tm);respond(payload.id,r)});
      window.parent.postMessage({type:'tuboshu_api_request',requestId:rid,payload:{method:payload.method||'GET',url:payload.url,data:payload.data,specialConfig:payload.specialConfig}},'*');
    }
  },
  beforeMount(){window.addEventListener('api_request',this.handleApiRequest)},
  mounted(){document.getElementById('app').style.display='block';window.parent.postMessage({type:'tuboshu_loaded'},'*')},
  unmounted(){window.removeEventListener('api_request',this.handleApiRequest)},
  setup(){return{pages,hidePage}}
});
setupVueApp(app);
<\/script></body></html>`;
}

onMounted(async () => {
  window.addEventListener('message', handleIframeMessage);
  const { priceConfig, priceRatio } = await loadConfig();
  const html = buildIframeHTML(priceConfig, priceRatio);
  const blob = new Blob([html], { type: 'text/html' });
  const blobUrl = URL.createObjectURL(blob);
  if (iframeRef.value) {
    iframeRef.value.src = blobUrl;
    setTimeout(() => { loading.value = false; }, 15000);
  }
});

onBeforeUnmount(() => {
  window.removeEventListener('message', handleIframeMessage);
  if (iframeRef.value?.src?.startsWith('blob:')) {
    URL.revokeObjectURL(iframeRef.value.src);
  }
});
</script>

<template>
  <Page title="土拨鼠论文" content-class="p-0">
    <div v-if="loading" class="flex items-center justify-center" style="min-height: 400px;">
      <Spin size="large" tip="正在加载土拨鼠论文组件..." />
    </div>
    <iframe
      ref="iframeRef"
      class="tuboshu-iframe"
      :style="{ opacity: loading ? 0 : 1 }"
      frameborder="0"
    ></iframe>
  </Page>
</template>

<style scoped>
.tuboshu-iframe {
  width: 100%;
  min-height: calc(100vh - 120px);
  border: none;
  transition: opacity 0.3s;
}
</style>

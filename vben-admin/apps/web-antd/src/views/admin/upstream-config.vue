<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Button, Input, InputNumber, Row, Col, Spin, Tag,
  message, Form, FormItem, Alert,
  Table, Modal, Select, SelectOption, Space, Popconfirm, Tabs, TabPane
} from 'ant-design-vue';
import {
  ReloadOutlined, PlusOutlined,
} from '@ant-design/icons-vue';
import {
  yfdkConfigGetApi, yfdkConfigSaveApi,
  sxdkConfigGetApi, sxdkConfigSaveApi,
  tutuqgUpstreamConfigGetApi, tutuqgUpstreamConfigSaveApi,
  hzwSocketConfigGetApi, hzwSocketConfigSaveApi,
  type YFDKConfig, type SXDKConfig, type TutuQGUpstreamConfig, type HZWSocketConfig,
} from '#/api/upstream-config';
import {
  tuboshuConfigGetApi, tuboshuConfigSaveApi,
  type TuboshuUpstreamConfig,
} from '#/api/tuboshu';
import {
  appuiConfigGetApi, appuiConfigSaveApi,
  type AppuiConfig,
} from '#/api/appui';
import {
  sdxyConfigGetApi, sdxyConfigSaveApi,
  type SDXYConfig as SDXYModuleConfig,
} from '#/api/sdxy';
import {
  ydsjConfigGetApi, ydsjConfigSaveApi,
  type YDSJConfig,
} from '#/api/ydsj';
import {
  xmProjectListApi, xmProjectSaveApi, xmProjectDeleteApi,
  type XMProjectItem,
} from '#/api/xm-project';
import {
  wAppListApi, wAppSaveApi, wAppDeleteApi,
  type WAppItem,
} from '#/api/w-app';
import {
  getYongyeConfig, saveYongyeConfig,
  type YongyeConfig,
} from '#/api/yongye';
import {
  paperConfigGetApi, paperConfigSaveApi,
  type PaperConfig,
} from '#/api/paper';

// ========== 状态 ==========
const loading = ref(false);
const activeKey = ref('api'); // 'api' or 'project'

// 图图强国
const tutuqgConfig = reactive<TutuQGUpstreamConfig>({
  base_url: '',
  key: '',
  price_increment: 0,
});
const tutuqgSaving = ref(false);

// YF打卡
const yfdkConfig = reactive<YFDKConfig>({
  base_url: '',
  token: '',
});
const yfdkSaving = ref(false);

// 泰山打卡
const sxdkConfig = reactive<SXDKConfig>({
  base_url: '',
  token: '',
  admin: '',
});
const sxdkSaving = ref(false);

// HZW Socket
const hzwSocketConfig = reactive<HZWSocketConfig>({
  socket_url: '',
});
const hzwSocketSaving = ref(false);

// 土拨鼠论文
const tuboshuConfig = reactive<TuboshuUpstreamConfig>({
  price_ratio: 5,
  price_config: {},
  page_visibility: {},
});
const tuboshuSaving = ref(false);

// Appui打卡
const appuiConfig = reactive<AppuiConfig>({
  base_url: '',
  uid: '',
  key: '',
  price_increment: 0,
  courses: [],
});
const appuiSaving = ref(false);

// 闪电运动
const sdxyModuleConfig = reactive<SDXYModuleConfig>({
  base_url: '',
  uid: '',
  key: '',
  price_per_km: 10,
});
const sdxySaving = ref(false);

// 运动世界
const ydsjConfig = reactive<YDSJConfig>({
  base_url: '',
  token: '',
  uid: '',
  key: '',
  price_multiple: 5,
  xbd_morning_price: 6,
  xbd_exercise_price: 6.5,
  real_cost_multiple: 1,
});
const ydsjSaving = ref(false);

// 永夜运动
const yongyeConfig = reactive<YongyeConfig>({
  api_url: '',
  token: '',
  dj: 0,
  zs: 1.25,
  beis: 1.3,
  xzdj: 0,
  xzmo: 100,
  tk: 0.01,
  content: '',
  tcgg: '',
});
const yongyeSaving = ref(false);

// 智文论文
const paperConfig = reactive<Partial<PaperConfig>>({
  lunwen_api_username: '',
  lunwen_api_password: '',
});
const paperSaving = ref(false);

// 小米运动项目
const xmProjects = ref<XMProjectItem[]>([]);
const xmLoading = ref(false);
const xmModalVisible = ref(false);
const xmSaving = ref(false);
const xmForm = reactive<Partial<XMProjectItem>>({
  id: 0, name: '', description: '', price: 0, query: 1, password: 1,
  url: '', uid: '', key: '', token: '', type: 0, p_id: '', status: 0,
});

const xmColumns = [
  { title: 'ID', dataIndex: 'id', width: 60 },
  { title: '项目名称', dataIndex: 'name', ellipsis: true },
  { title: '价格', dataIndex: 'price', width: 80 },
  { title: '上游项目ID', dataIndex: 'p_id', width: 100 },
  { title: '认证类型', key: 'type', width: 90 },
  { title: '状态', key: 'status', width: 80 },
  { title: '操作', key: 'action', width: 120 },
];

async function loadXmProjects() {
  xmLoading.value = true;
  try {
    const res = await xmProjectListApi();
    xmProjects.value = res || [];
  } catch (e) { console.error(e); }
  finally { xmLoading.value = false; }
}

function openXmAdd() {
  Object.assign(xmForm, { id: 0, name: '', description: '', price: 0, query: 1, password: 1, url: '', uid: '', key: '', token: '', type: 0, p_id: '', status: 0 });
  xmModalVisible.value = true;
}

function openXmEdit(row: XMProjectItem) {
  Object.assign(xmForm, { ...row });
  xmModalVisible.value = true;
}

async function submitXmForm() {
  if (!xmForm.name?.trim()) { message.warning('项目名称不能为空'); return; }
  xmSaving.value = true;
  try {
    await xmProjectSaveApi({ ...xmForm });
    message.success(xmForm.id ? '保存成功' : '添加成功');
    xmModalVisible.value = false;
    loadXmProjects();
  } catch (e: any) { message.error(e?.message || '保存失败'); }
  finally { xmSaving.value = false; }
}

async function deleteXmProject(id: number) {
  try {
    await xmProjectDeleteApi(id);
    message.success('删除成功');
    loadXmProjects();
  } catch (e: any) { message.error(e?.message || '删除失败'); }
}

// 鲸鱼运动项目
const wApps = ref<WAppItem[]>([]);
const wLoading = ref(false);
const wModalVisible = ref(false);
const wSaving = ref(false);
const wForm = reactive<Partial<WAppItem>>({
  id: 0, name: '', code: '', org_app_id: '', status: 0, description: '', price: 1,
  cac_type: 'KM', url: '', key: '', uid: '', token: '', type: '1',
});

const wColumns = [
  { title: 'ID', dataIndex: 'id', width: 50 },
  { title: '名称', dataIndex: 'name', ellipsis: true },
  { title: '代码', dataIndex: 'code', width: 80 },
  { title: '源台ID', dataIndex: 'org_app_id', width: 80 },
  { title: '价格', dataIndex: 'price', width: 70 },
  { title: '计费', key: 'cac_type', width: 70 },
  { title: '类型', key: 'w_type', width: 90 },
  { title: '状态', key: 'w_status', width: 70 },
  { title: '操作', key: 'w_action', width: 120 },
];

async function loadWApps() {
  wLoading.value = true;
  try {
    const res = await wAppListApi();
    wApps.value = res || [];
  } catch (e) { console.error(e); }
  finally { wLoading.value = false; }
}

function openWAdd() {
  Object.assign(wForm, { id: 0, name: '', code: '', org_app_id: '', status: 0, description: '', price: 1, cac_type: 'KM', url: '', key: '', uid: '', token: '', type: '1' });
  wModalVisible.value = true;
}

function openWEdit(row: WAppItem) {
  Object.assign(wForm, { ...row });
  wModalVisible.value = true;
}

async function submitWForm() {
  if (!wForm.name?.trim() || !wForm.code?.trim()) { message.warning('名称和代码不能为空'); return; }
  wSaving.value = true;
  try {
    await wAppSaveApi({ ...wForm });
    message.success(wForm.id ? '保存成功' : '添加成功');
    wModalVisible.value = false;
    loadWApps();
  } catch (e: any) { message.error(e?.message || '保存失败'); }
  finally { wSaving.value = false; }
}

async function deleteWApp(id: number) {
  try {
    await wAppDeleteApi(id);
    message.success('删除成功');
    loadWApps();
  } catch (e: any) { message.error(e?.message || '删除失败'); }
}

// ========== 加载 ==========
async function loadAll() {
  loading.value = true;
  try {
    const [tutuqgRes, yfdkRes, sxdkRes, hzwRes, tuboshuRes, appuiRes, sdxyRes, ydsjRes, yongyeRes, paperRes] = await Promise.allSettled([
      tutuqgUpstreamConfigGetApi(),
      yfdkConfigGetApi(),
      sxdkConfigGetApi(),
      hzwSocketConfigGetApi(),
      tuboshuConfigGetApi(),
      appuiConfigGetApi(),
      sdxyConfigGetApi(),
      ydsjConfigGetApi(),
      getYongyeConfig(),
      paperConfigGetApi(),
    ]);

    if (tutuqgRes.status === 'fulfilled' && tutuqgRes.value) {
      Object.assign(tutuqgConfig, tutuqgRes.value);
    }
    if (yfdkRes.status === 'fulfilled' && yfdkRes.value) {
      Object.assign(yfdkConfig, yfdkRes.value);
    }
    if (sxdkRes.status === 'fulfilled' && sxdkRes.value) {
      Object.assign(sxdkConfig, sxdkRes.value);
    }
    if (hzwRes.status === 'fulfilled' && hzwRes.value) {
      Object.assign(hzwSocketConfig, hzwRes.value);
    }
    if (tuboshuRes.status === 'fulfilled' && tuboshuRes.value) {
      Object.assign(tuboshuConfig, tuboshuRes.value);
    }
    if (appuiRes.status === 'fulfilled' && appuiRes.value) {
      Object.assign(appuiConfig, appuiRes.value);
    }
    if (sdxyRes.status === 'fulfilled' && sdxyRes.value) {
      Object.assign(sdxyModuleConfig, sdxyRes.value);
    }
    if (ydsjRes.status === 'fulfilled' && ydsjRes.value) {
      Object.assign(ydsjConfig, ydsjRes.value);
    }
    if (yongyeRes.status === 'fulfilled' && yongyeRes.value) {
      Object.assign(yongyeConfig, yongyeRes.value);
    }
    if (paperRes.status === 'fulfilled' && paperRes.value) {
      Object.assign(paperConfig, paperRes.value);
    }
  } catch (e) {
    console.error('加载对接配置失败:', e);
  } finally {
    loading.value = false;
  }
}

// ========== 保存 ==========
async function saveTutuqg() {
  tutuqgSaving.value = true;
  try {
    await tutuqgUpstreamConfigSaveApi({ ...tutuqgConfig });
    message.success('图图强国配置保存成功');
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    tutuqgSaving.value = false;
  }
}

async function saveYfdk() {
  yfdkSaving.value = true;
  try {
    await yfdkConfigSaveApi({ ...yfdkConfig });
    message.success('YF打卡配置保存成功');
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    yfdkSaving.value = false;
  }
}

async function saveSxdk() {
  sxdkSaving.value = true;
  try {
    await sxdkConfigSaveApi({ ...sxdkConfig });
    message.success('泰山打卡配置保存成功');
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    sxdkSaving.value = false;
  }
}

async function saveHzwSocket() {
  hzwSocketSaving.value = true;
  try {
    await hzwSocketConfigSaveApi({ ...hzwSocketConfig });
    message.success('HZW Socket 配置保存成功，客户端已重启');
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    hzwSocketSaving.value = false;
  }
}

async function saveTuboshu() {
  tuboshuSaving.value = true;
  try {
    await tuboshuConfigSaveApi({ ...tuboshuConfig });
    message.success('土拨鼠论文配置保存成功');
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    tuboshuSaving.value = false;
  }
}

async function saveAppui() {
  appuiSaving.value = true;
  try {
    await appuiConfigSaveApi({ ...appuiConfig });
    message.success('Appui打卡配置保存成功');
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    appuiSaving.value = false;
  }
}

async function saveSdxy() {
  sdxySaving.value = true;
  try {
    await sdxyConfigSaveApi({ ...sdxyModuleConfig });
    message.success('闪电运动配置保存成功');
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    sdxySaving.value = false;
  }
}

async function saveYdsj() {
  ydsjSaving.value = true;
  try {
    await ydsjConfigSaveApi({ ...ydsjConfig });
    message.success('运动世界配置保存成功');
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    ydsjSaving.value = false;
  }
}

async function saveYongye() {
  yongyeSaving.value = true;
  try {
    await saveYongyeConfig({ ...yongyeConfig });
    message.success('永夜运动配置保存成功');
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    yongyeSaving.value = false;
  }
}

async function savePaper() {
  paperSaving.value = true;
  try {
    await paperConfigSaveApi({ ...paperConfig } as PaperConfig);
    message.success('智文论文配置保存成功');
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    paperSaving.value = false;
  }
}

function isConfigured(url: string) {
  return url !== undefined && url !== null && url.trim() !== '';
}

onMounted(() => {
  loadAll();
  loadXmProjects();
  loadWApps();
});
</script>

<template>
  <Page title="对接中心" description="管理上游API对接配置，修改后即时生效。">
    <template #extra>
      <Button type="primary" @click="loadAll" :loading="loading" ghost>
        <template #icon><ReloadOutlined /></template>
        刷新
      </Button>
    </template>

    <div class="upstream-tabs">
      <Tabs v-model:activeKey="activeKey">
        <TabPane key="api" tab="接口配置">
          <Spin :spinning="loading">
            <Row :gutter="[16, 16]">
              <!-- 图图强国 -->
              <Col :xs="24" :md="12" :xl="8">
                <Card :bordered="false" class="cfg-card">
                  <template #title>
                    <span class="card-title">图图强国</span>
                    <Tag v-if="isConfigured(tutuqgConfig.base_url)" color="success" class="ml-2">已对接</Tag>
                    <Tag v-else color="default" class="ml-2">未配置</Tag>
                  </template>
                  <template #extra>
                    <Button type="primary" size="small" :loading="tutuqgSaving" @click="saveTutuqg">保存</Button>
                  </template>
                  <Form layout="vertical" :colon="false">
                    <FormItem label="API 地址">
                      <Input v-model:value="tutuqgConfig.base_url" placeholder="https://api.example.com" />
                    </FormItem>
                    <FormItem label="密钥">
                      <Input.Password v-model:value="tutuqgConfig.key" placeholder="请输入密钥" />
                    </FormItem>
                    <FormItem label="加价倍率">
                      <InputNumber v-model:value="tutuqgConfig.price_increment" :min="0" :step="0.1" :precision="2" class="w-full" />
                    </FormItem>
                  </Form>
                </Card>
              </Col>

              <!-- YF打卡 -->
              <Col :xs="24" :md="12" :xl="8">
                <Card :bordered="false" class="cfg-card">
                  <template #title>
                    <span class="card-title">YF打卡</span>
                    <Tag v-if="isConfigured(yfdkConfig.base_url)" color="success" class="ml-2">已对接</Tag>
                    <Tag v-else color="default" class="ml-2">未配置</Tag>
                  </template>
                  <template #extra>
                    <Button type="primary" size="small" :loading="yfdkSaving" @click="saveYfdk">保存</Button>
                  </template>
                  <Form layout="vertical" :colon="false">
                    <FormItem label="API 地址">
                      <Input v-model:value="yfdkConfig.base_url" placeholder="https://dk.blwl.fun/api/" />
                    </FormItem>
                    <FormItem label="Token">
                      <Input.Password v-model:value="yfdkConfig.token" placeholder="请输入 Token" />
                    </FormItem>
                  </Form>
                </Card>
              </Col>

              <!-- 泰山打卡 -->
              <Col :xs="24" :md="12" :xl="8">
                <Card :bordered="false" class="cfg-card">
                  <template #title>
                    <span class="card-title">泰山打卡</span>
                    <Tag v-if="isConfigured(sxdkConfig.base_url)" color="success" class="ml-2">已对接</Tag>
                    <Tag v-else color="default" class="ml-2">未配置</Tag>
                  </template>
                  <template #extra>
                    <Button type="primary" size="small" :loading="sxdkSaving" @click="saveSxdk">保存</Button>
                  </template>
                  <Form layout="vertical" :colon="false">
                    <FormItem label="API 地址">
                      <Input v-model:value="sxdkConfig.base_url" placeholder="http://..." />
                    </FormItem>
                    <FormItem label="Token">
                      <Input.Password v-model:value="sxdkConfig.token" placeholder="请输入 Token" />
                    </FormItem>
                    <FormItem label="管理账号">
                      <Input v-model:value="sxdkConfig.admin" placeholder="请输入管理账号" />
                    </FormItem>
                  </Form>
                </Card>
              </Col>

              <!-- HZW实时进度 -->
              <Col :xs="24" :md="12" :xl="8">
                <Card :bordered="false" class="cfg-card">
                  <template #title>
                    <span class="card-title">HZW 实时进度</span>
                    <Tag v-if="isConfigured(hzwSocketConfig.socket_url)" color="success" class="ml-2">已启用</Tag>
                    <Tag v-else color="default" class="ml-2">未配置</Tag>
                  </template>
                  <template #extra>
                    <Button type="primary" size="small" :loading="hzwSocketSaving" @click="saveHzwSocket">保存</Button>
                  </template>
                  <Form layout="vertical" :colon="false">
                    <FormItem label="Socket 地址">
                      <Input v-model:value="hzwSocketConfig.socket_url" placeholder="http://socket.biedawo.org" />
                    </FormItem>
                  </Form>
                  <Alert message="保存后自动启用实时推送，订单状态将自动同步。" type="info" show-icon class="mt-2" />
                </Card>
              </Col>

              <!-- 土拨鼠论文 -->
              <Col :xs="24" :md="12" :xl="8">
                <Card :bordered="false" class="cfg-card">
                  <template #title>
                    <span class="card-title">土拨鼠论文</span>
                    <Tag v-if="tuboshuConfig.price_ratio > 0" color="success" class="ml-2">已配置</Tag>
                    <Tag v-else color="default" class="ml-2">未配置</Tag>
                  </template>
                  <template #extra>
                    <Button type="primary" size="small" :loading="tuboshuSaving" @click="saveTuboshu">保存</Button>
                  </template>
                  <Form layout="vertical" :colon="false">
                    <FormItem label="价格倍率">
                      <InputNumber v-model:value="tuboshuConfig.price_ratio" :min="0.1" :step="0.5" :precision="1" class="w-full" />
                    </FormItem>
                  </Form>
                  <Alert message="Token 在货源中心配置，价格可在论文页面修改。" type="info" show-icon class="mt-2" />
                </Card>
              </Col>

              <!-- 智文论文 -->
              <Col :xs="24" :md="12" :xl="8">
                <Card :bordered="false" class="cfg-card">
                  <template #title>
                    <span class="card-title">智文论文</span>
                    <Tag v-if="isConfigured(paperConfig.lunwen_api_username || '')" color="success" class="ml-2">已配置</Tag>
                    <Tag v-else color="default" class="ml-2">未配置</Tag>
                  </template>
                  <template #extra>
                    <Button type="primary" size="small" :loading="paperSaving" @click="savePaper">保存</Button>
                  </template>
                  <Form layout="vertical" :colon="false">
                    <FormItem label="API 账号">
                      <Input v-model:value="paperConfig.lunwen_api_username" placeholder="请输入登录账号" />
                    </FormItem>
                    <FormItem label="API 密码">
                      <Input.Password v-model:value="paperConfig.lunwen_api_password" placeholder="请输入登录密码" />
                    </FormItem>
                  </Form>
                  <Alert message="价格配置请前往 后台管理 → 上游对接 → 智文论文配置 页面修改。" type="info" show-icon class="mt-2" />
                </Card>
              </Col>

              <!-- Appui打卡 -->
              <Col :xs="24" :md="12" :xl="8">
                <Card :bordered="false" class="cfg-card">
                  <template #title>
                    <span class="card-title">Appui打卡</span>
                    <Tag v-if="isConfigured(appuiConfig.base_url)" color="success" class="ml-2">已对接</Tag>
                    <Tag v-else color="default" class="ml-2">未配置</Tag>
                  </template>
                  <template #extra>
                    <Button type="primary" size="small" :loading="appuiSaving" @click="saveAppui">保存</Button>
                  </template>
                  <Form layout="vertical" :colon="false">
                    <FormItem label="API 地址">
                      <Input v-model:value="appuiConfig.base_url" placeholder="https://..." />
                    </FormItem>
                    <FormItem label="上游 UID">
                      <Input v-model:value="appuiConfig.uid" placeholder="请输入 UID" />
                    </FormItem>
                    <FormItem label="密钥">
                      <Input.Password v-model:value="appuiConfig.key" placeholder="请输入密钥" />
                    </FormItem>
                    <FormItem label="加价金额">
                      <InputNumber v-model:value="appuiConfig.price_increment" :min="0" :step="0.5" :precision="2" class="w-full" />
                    </FormItem>
                  </Form>
                </Card>
              </Col>

              <!-- 闪电运动 -->
              <Col :xs="24" :md="12" :xl="8">
                <Card :bordered="false" class="cfg-card">
                  <template #title>
                    <span class="card-title">闪电运动</span>
                    <Tag v-if="isConfigured(sdxyModuleConfig.base_url)" color="success" class="ml-2">已对接</Tag>
                    <Tag v-else color="default" class="ml-2">未配置</Tag>
                  </template>
                  <template #extra>
                    <Button type="primary" size="small" :loading="sdxySaving" @click="saveSdxy">保存</Button>
                  </template>
                  <Form layout="vertical" :colon="false">
                    <FormItem label="API 地址">
                      <Input v-model:value="sdxyModuleConfig.base_url" placeholder="https://..." />
                    </FormItem>
                    <FormItem label="上游 UID">
                      <Input v-model:value="sdxyModuleConfig.uid" placeholder="请输入 UID" />
                    </FormItem>
                    <FormItem label="密钥">
                      <Input.Password v-model:value="sdxyModuleConfig.key" placeholder="请输入密钥" />
                    </FormItem>
                    <FormItem label="每公里价格">
                      <InputNumber v-model:value="sdxyModuleConfig.price_per_km" :min="0" :step="0.5" :precision="2" class="w-full" />
                    </FormItem>
                  </Form>
                </Card>
              </Col>

              <!-- 永夜运动 -->
              <Col :xs="24" :md="12" :xl="8">
                <Card :bordered="false" class="cfg-card">
                  <template #title>
                    <span class="card-title">永夜运动</span>
                    <Tag v-if="isConfigured(yongyeConfig.api_url)" color="success" class="ml-2">已对接</Tag>
                    <Tag v-else color="default" class="ml-2">未配置</Tag>
                  </template>
                  <template #extra>
                    <Button type="primary" size="small" :loading="yongyeSaving" @click="saveYongye">保存</Button>
                  </template>
                  <Form layout="vertical" :colon="false">
                    <FormItem label="API 地址">
                      <Input v-model:value="yongyeConfig.api_url" placeholder="https://yy.rgrg.cc/api" />
                    </FormItem>
                    <FormItem label="Token">
                      <Input.Password v-model:value="yongyeConfig.token" placeholder="请输入上游 Token" />
                    </FormItem>
                    <Row :gutter="12">
                      <Col :span="8">
                        <FormItem label="赠送倍率">
                          <InputNumber v-model:value="yongyeConfig.zs" :min="0" :step="0.05" :precision="2" class="w-full" />
                        </FormItem>
                      </Col>
                      <Col :span="8">
                        <FormItem label="价格倍数">
                          <InputNumber v-model:value="yongyeConfig.beis" :min="0" :step="0.1" :precision="2" class="w-full" />
                        </FormItem>
                      </Col>
                      <Col :span="8">
                        <FormItem label="退款费率">
                          <InputNumber v-model:value="yongyeConfig.tk" :min="0" :max="1" :step="0.01" :precision="2" class="w-full" />
                        </FormItem>
                      </Col>
                    </Row>
                    <Row :gutter="12">
                      <Col :span="8">
                        <FormItem label="等级(dj)">
                          <InputNumber v-model:value="yongyeConfig.dj" :min="0" :step="0.1" :precision="2" class="w-full" />
                        </FormItem>
                      </Col>
                      <Col :span="8">
                        <FormItem label="限制等级">
                          <InputNumber v-model:value="yongyeConfig.xzdj" :min="0" :step="0.1" :precision="2" class="w-full" />
                        </FormItem>
                      </Col>
                      <Col :span="8">
                        <FormItem label="限制余额">
                          <InputNumber v-model:value="yongyeConfig.xzmo" :min="0" :step="10" class="w-full" />
                        </FormItem>
                      </Col>
                    </Row>
                  </Form>
                </Card>
              </Col>

              <!-- 运动世界 -->
              <Col :xs="24" :md="12" :xl="8">
                <Card :bordered="false" class="cfg-card">
                  <template #title>
                    <span class="card-title">运动世界</span>
                    <Tag v-if="isConfigured(ydsjConfig.base_url)" color="success" class="ml-2">已对接</Tag>
                    <Tag v-else color="default" class="ml-2">未配置</Tag>
                  </template>
                  <template #extra>
                    <Button type="primary" size="small" :loading="ydsjSaving" @click="saveYdsj">保存</Button>
                  </template>
                  <Form layout="vertical" :colon="false">
                    <FormItem label="API 地址">
                      <Input v-model:value="ydsjConfig.base_url" placeholder="http://103.149.27.248:5000" />
                    </FormItem>
                    <FormItem label="Bearer Token">
                      <Input.Password v-model:value="ydsjConfig.token" placeholder="请输入上游 Bearer Token" />
                    </FormItem>
                    <Row :gutter="12">
                      <Col :span="12">
                        <FormItem label="价格倍率">
                          <InputNumber v-model:value="ydsjConfig.price_multiple" :min="0" :step="0.1" :precision="1" class="w-full" />
                        </FormItem>
                      </Col>
                      <Col :span="12">
                        <FormItem label="扣费倍率">
                          <InputNumber v-model:value="ydsjConfig.real_cost_multiple" :min="0" :step="0.1" :precision="1" class="w-full" />
                        </FormItem>
                      </Col>
                    </Row>
                    <Row :gutter="12">
                      <Col :span="12">
                        <FormItem label="晨跑价格">
                          <InputNumber v-model:value="ydsjConfig.xbd_morning_price" :min="0" :step="0.1" :precision="1" class="w-full" />
                        </FormItem>
                      </Col>
                      <Col :span="12">
                        <FormItem label="课外跑价格">
                          <InputNumber v-model:value="ydsjConfig.xbd_exercise_price" :min="0" :step="0.1" :precision="1" class="w-full" />
                        </FormItem>
                      </Col>
                    </Row>
                  </Form>
                </Card>
              </Col>
            </Row>
          </Spin>
        </TabPane>

        <TabPane key="project" tab="项目管理">
          <div class="mt-3 space-y-4">
            <!-- 小米运动 -->
            <Card :bordered="false" class="cfg-card">
              <template #title>
                <span class="card-title">小米运动</span>
                <Tag color="processing" class="ml-2">{{ xmProjects.length }} 个项目</Tag>
              </template>
              <template #extra>
                <Space>
                  <Button size="small" @click="loadXmProjects" :loading="xmLoading">刷新</Button>
                  <Button type="primary" size="small" @click="openXmAdd">
                    <template #icon><PlusOutlined /></template>
                    添加
                  </Button>
                </Space>
              </template>
              <Table :columns="xmColumns" :data-source="xmProjects" :loading="xmLoading" :pagination="false" row-key="id" size="small">
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'type'">
                    <Tag :color="record.type === 0 ? 'blue' : 'green'">{{ record.type === 0 ? 'Key' : 'Token' }}</Tag>
                  </template>
                  <template v-else-if="column.key === 'status'">
                    <Tag :color="record.status === 0 ? 'green' : 'red'">{{ record.status === 0 ? '正常' : '下架' }}</Tag>
                  </template>
                  <template v-else-if="column.key === 'action'">
                    <Space :size="0">
                      <Button type="link" size="small" @click="openXmEdit(record)">编辑</Button>
                      <Popconfirm title="确定删除？" @confirm="deleteXmProject(record.id)">
                        <Button type="link" danger size="small">删除</Button>
                      </Popconfirm>
                    </Space>
                  </template>
                </template>
              </Table>
            </Card>

            <!-- 鲸鱼运动 -->
            <Card :bordered="false" class="cfg-card">
              <template #title>
                <span class="card-title">鲸鱼运动</span>
                <Tag color="processing" class="ml-2">{{ wApps.length }} 个项目</Tag>
              </template>
              <template #extra>
                <Space>
                  <Button size="small" @click="loadWApps" :loading="wLoading">刷新</Button>
                  <Button type="primary" size="small" @click="openWAdd">
                    <template #icon><PlusOutlined /></template>
                    添加
                  </Button>
                </Space>
              </template>
              <Table :columns="wColumns" :data-source="wApps" :loading="wLoading" :pagination="false" row-key="id" size="small">
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'cac_type'">
                    <Tag :color="record.cac_type === 'TS' ? 'orange' : 'cyan'">{{ record.cac_type === 'TS' ? '按次' : '按KM' }}</Tag>
                  </template>
                  <template v-else-if="column.key === 'w_type'">
                    <Tag :color="record.type === '2' ? 'purple' : record.type === '1' ? 'blue' : 'default'">{{ record.type === '2' ? '鲸鱼' : record.type === '1' ? 'Token' : 'Key' }}</Tag>
                  </template>
                  <template v-else-if="column.key === 'w_status'">
                    <Tag :color="record.status === 0 ? 'green' : 'red'">{{ record.status === 0 ? '上架' : '下架' }}</Tag>
                  </template>
                  <template v-else-if="column.key === 'w_action'">
                    <Space :size="0">
                      <Button type="link" size="small" @click="openWEdit(record)">编辑</Button>
                      <Popconfirm title="确定删除？" @confirm="deleteWApp(record.id)">
                        <Button type="link" danger size="small">删除</Button>
                      </Popconfirm>
                    </Space>
                  </template>
                </template>
              </Table>
            </Card>
          </div>
        </TabPane>
      </Tabs>
    </div>

    <!-- 小米运动编辑弹窗 -->
    <Modal v-model:open="xmModalVisible" :title="xmForm.id ? '编辑项目' : '添加项目'" @ok="submitXmForm" :confirm-loading="xmSaving" ok-text="保存" width="640px">
      <Form layout="vertical" :colon="false" class="mt-3">
        <Row :gutter="16">
          <Col :span="12">
            <FormItem label="项目名称">
              <Input v-model:value="xmForm.name" placeholder="例如：某某跑步" />
            </FormItem>
          </Col>
          <Col :span="12">
            <FormItem label="价格（元/公里）">
              <InputNumber v-model:value="xmForm.price" :min="0" :step="0.1" :precision="2" class="w-full" />
            </FormItem>
          </Col>
        </Row>
        <FormItem label="说明">
          <Input.TextArea v-model:value="xmForm.description" :rows="2" placeholder="项目描述" />
        </FormItem>
        <Row :gutter="16">
          <Col :span="12">
            <FormItem label="认证类型">
              <Select v-model:value="xmForm.type" class="w-full">
                <SelectOption :value="0">Key 认证</SelectOption>
                <SelectOption :value="1">Token 认证</SelectOption>
              </Select>
            </FormItem>
          </Col>
          <Col :span="12">
            <FormItem label="上游项目ID">
              <Input v-model:value="xmForm.p_id" placeholder="上游 p_id" />
            </FormItem>
          </Col>
        </Row>
        <FormItem label="API 地址">
          <Input v-model:value="xmForm.url" placeholder="上游接口地址" />
        </FormItem>
        <Row :gutter="16">
          <Col :span="12">
            <FormItem label="UID">
              <Input v-model:value="xmForm.uid" placeholder="上游用户ID" />
            </FormItem>
          </Col>
          <Col :span="12">
            <FormItem v-if="xmForm.type === 0" label="Key">
              <Input.Password v-model:value="xmForm.key" placeholder="API密钥" />
            </FormItem>
            <FormItem v-else label="Token">
              <Input.Password v-model:value="xmForm.token" placeholder="Token" />
            </FormItem>
          </Col>
        </Row>
        <Row :gutter="16">
          <Col :span="8">
            <FormItem label="状态">
              <Select v-model:value="xmForm.status" class="w-full">
                <SelectOption :value="0">正常</SelectOption>
                <SelectOption :value="1">下架</SelectOption>
              </Select>
            </FormItem>
          </Col>
          <Col :span="8">
            <FormItem label="支持查课">
              <Select v-model:value="xmForm.query" class="w-full">
                <SelectOption :value="1">是</SelectOption>
                <SelectOption :value="0">否</SelectOption>
              </Select>
            </FormItem>
          </Col>
          <Col :span="8">
            <FormItem label="需要密码">
              <Select v-model:value="xmForm.password" class="w-full">
                <SelectOption :value="1">是</SelectOption>
                <SelectOption :value="0">否</SelectOption>
              </Select>
            </FormItem>
          </Col>
        </Row>
      </Form>
    </Modal>

    <!-- 鲸鱼运动编辑弹窗 -->
    <Modal v-model:open="wModalVisible" :title="wForm.id ? '编辑项目' : '添加项目'" @ok="submitWForm" :confirm-loading="wSaving" ok-text="保存" width="640px">
      <Form layout="vertical" :colon="false" class="mt-3">
        <Row :gutter="16">
          <Col :span="12">
            <FormItem label="项目名称">
              <Input v-model:value="wForm.name" placeholder="例如：步道乐跑" />
            </FormItem>
          </Col>
          <Col :span="12">
            <FormItem label="项目代码">
              <Input v-model:value="wForm.code" placeholder="bdlp / keep / ymty" />
            </FormItem>
          </Col>
        </Row>
        <Row :gutter="16">
          <Col :span="8">
            <FormItem label="源台项目ID">
              <Input v-model:value="wForm.org_app_id" placeholder="org_app_id" />
            </FormItem>
          </Col>
          <Col :span="8">
            <FormItem label="底价">
              <InputNumber v-model:value="wForm.price" :min="0" :step="0.1" :precision="2" class="w-full" />
            </FormItem>
          </Col>
          <Col :span="8">
            <FormItem label="计费方式">
              <Select v-model:value="wForm.cac_type" class="w-full">
                <SelectOption value="TS">按次</SelectOption>
                <SelectOption value="KM">按公里</SelectOption>
              </Select>
            </FormItem>
          </Col>
        </Row>
        <FormItem label="项目说明">
          <Input.TextArea v-model:value="wForm.description" :rows="2" placeholder="展示到下单页面" />
        </FormItem>
        <FormItem label="对接URL">
          <Input v-model:value="wForm.url" placeholder="上游接口地址" />
        </FormItem>
        <Row :gutter="16">
          <Col :span="12">
            <FormItem label="认证类型">
              <Select v-model:value="wForm.type" class="w-full">
                <SelectOption value="0">Key+UID</SelectOption>
                <SelectOption value="1">Token (X-WTK)</SelectOption>
                <SelectOption value="2">鲸鱼(Jingyu)格式</SelectOption>
              </Select>
            </FormItem>
          </Col>
          <Col :span="12">
            <FormItem label="状态">
              <Select v-model:value="wForm.status" class="w-full">
                <SelectOption :value="0">上架</SelectOption>
                <SelectOption :value="1">下架</SelectOption>
              </Select>
            </FormItem>
          </Col>
        </Row>
        <Row :gutter="16">
          <Col :span="8">
            <FormItem label="UID">
              <Input v-model:value="wForm.uid" placeholder="对接UID" />
            </FormItem>
          </Col>
          <Col :span="8">
            <FormItem label="Key">
              <Input.Password v-model:value="wForm.key" placeholder="对接密钥" />
            </FormItem>
          </Col>
          <Col :span="8">
            <FormItem label="Token">
              <Input.Password v-model:value="wForm.token" placeholder="源台Token" />
            </FormItem>
          </Col>
        </Row>
      </Form>
    </Modal>
  </Page>
</template>

<style scoped>
.upstream-tabs :deep(.ant-tabs-nav) {
  margin-bottom: 0;
}
.cfg-card {
  height: 100%;
  border-radius: 8px;
  transition: box-shadow 0.2s ease;
}
.cfg-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}
.card-title {
  font-weight: 600;
  font-size: 15px;
}
</style>

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
  xmProviderListApi, xmProviderSaveApi, xmProviderDeleteApi,
  xmProviderTestApi, xmProviderFetchProjectsApi, xmProviderImportProjectsApi, xmProviderSyncProjectsApi,
  type XMProviderItem, type XMUpstreamProjectItem,
} from '#/api/xm-provider';
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
import {
  getYFDKProjectsApi, syncYFDKProjectsApi, updateYFDKProjectApi, deleteYFDKProjectApi,
  type YFDKAdminProject,
} from '#/api/admin';
import {
  tuzhiConfigGetApi, tuzhiConfigSaveApi,
  tuzhiAdminGetGoodsApi, tuzhiGoodsOverridesGetApi, tuzhiGoodsOverridesSaveApi,
  type TuZhiConfig, type TuZhiGoodsOverride,
} from '#/api/tuzhi';

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
  endpoint: '/flash/api.php',
  uid: '',
  key: '',
  timeout: 30,
  price: 10,
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
  lunwen_api_6000_price: '30',
  lunwen_api_8000_price: '40',
  lunwen_api_10000_price: '50',
  lunwen_api_12000_price: '60',
  lunwen_api_15000_price: '75',
  lunwen_api_rws_price: '10',
  lunwen_api_ktbg_price: '10',
  lunwen_api_jdaigchj_price: '10',
  lunwen_api_xgdl_price: '3',
  lunwen_api_jcl_price: '3',
  lunwen_api_jdaigcl_price: '3',
});
const paperSaving = ref(false);

// 土拨鼠页面显示选项
const tuboshuPageOptions = [
  { key: 'ComponentStagePage', label: '分步对话' },
  { key: 'ChatPage', label: 'AI对话' },
  { key: 'ChartPage', label: '图表生成' },
  { key: 'TemplatePage', label: '模板中心' },
  { key: 'ReductionPage', label: '论文降重' },
  { key: 'AccountTable', label: '账户管理' },
  { key: 'TicketPage', label: '工单系统' },
];

// 凸知打卡
const tuzhiConfig = reactive<TuZhiConfig>({
  daka_api_username: '',
  daka_api_password: '',
});
const tuzhiSaving = ref(false);
const tuzhiGoods = ref<any[]>([]);
const tuzhiOverrides = ref<TuZhiGoodsOverride[]>([]);
const tuzhiGoodsLoading = ref(false);
const tuzhiOverridesSaving = ref(false);

const tuzhiGoodsColumns = [
  { title: 'ID', dataIndex: 'id', width: 60 },
  { title: '商品名称', dataIndex: 'name', ellipsis: true },
  { title: '上游价格', key: 'upstream_price', width: 90 },
  { title: '计费方式', key: 'billing', width: 90 },
  { title: '覆盖售价', key: 'override_price', width: 110 },
  { title: '上架', key: 'enabled', width: 80 },
];

// 小米运动：连接 + 项目
const xmProviders = ref<XMProviderItem[]>([]);
const xmProvidersLoading = ref(false);
const xmProviderModalVisible = ref(false);
const xmProviderSaving = ref(false);
const xmProviderTesting = ref(false);
const xmProviderForm = reactive<Partial<XMProviderItem>>({
  id: 0,
  name: '',
  base_url: '',
  auth_type: 0,
  uid: '',
  key: '',
  token: '',
  status: 0,
  remark: '',
});

const xmProjects = ref<XMProjectItem[]>([]);
const xmLoading = ref(false);
const xmModalVisible = ref(false);
const xmSaving = ref(false);
const xmForm = reactive<Partial<XMProjectItem>>({
  id: 0,
  provider_id: 0,
  provider_name: '',
  name: '',
  description: '',
  price: 0,
  upstream_price: 0,
  query: 1,
  password: 1,
  p_id: '',
  status: 0,
  sort_order: 0,
  sync_mode: 1,
});

const xmImportVisible = ref(false);
const xmImportLoading = ref(false);
const xmImportSubmitting = ref(false);
const xmImportProvider = ref<XMProviderItem | null>(null);
const xmImportProjects = ref<XMUpstreamProjectItem[]>([]);
const xmSelectedImportProjectIds = ref<string[]>([]);
const xmImportForm = reactive({
  price_multiplier: 1,
  price_addition: 0,
  overwrite_local_price: true,
});

function handleXmImportSelectionChange(keys: (number | string)[]) {
  xmSelectedImportProjectIds.value = keys.map((key) => String(key));
}

const xmSyncVisible = ref(false);
const xmSyncLoading = ref(false);
const xmSyncProvider = ref<XMProviderItem | null>(null);
const xmSyncForm = reactive({
  provider_id: 0,
  sync_name: true,
  sync_description: true,
  sync_upstream_price: true,
  sync_query: true,
  sync_password: true,
  overwrite_local_price: false,
  price_multiplier: 1,
  price_addition: 0,
});

const xmProviderColumns = [
  { title: 'ID', dataIndex: 'id', width: 60 },
  { title: '连接名称', dataIndex: 'name', width: 180, ellipsis: true },
  { title: '认证', key: 'auth_type', width: 90 },
  { title: 'API 地址', dataIndex: 'base_url', ellipsis: true },
  { title: '项目数', dataIndex: 'project_count', width: 80 },
  { title: '最近同步', dataIndex: 'last_sync_at', width: 160 },
  { title: '状态', key: 'status', width: 80 },
  { title: '操作', key: 'action', width: 280 },
];

const xmColumns = [
  { title: 'ID', dataIndex: 'id', width: 60 },
  { title: '项目名称', dataIndex: 'name', ellipsis: true },
  { title: '来源连接', dataIndex: 'provider_name', width: 140, ellipsis: true },
  { title: '本地售价', dataIndex: 'price', width: 90 },
  { title: '上游价格', dataIndex: 'upstream_price', width: 90 },
  { title: '上游项目ID', dataIndex: 'p_id', width: 100 },
  { title: '状态', key: 'status', width: 80 },
  { title: '操作', key: 'action', width: 120 },
];

const xmImportColumns = [
  { title: '上游ID', dataIndex: 'id', width: 80 },
  { title: '项目名称', dataIndex: 'name', ellipsis: true },
  { title: '上游价格', dataIndex: 'price', width: 90 },
  { title: '支持查询', key: 'query', width: 90 },
  { title: '需要密码', key: 'password', width: 90 },
];

async function loadXmProviders() {
  xmProvidersLoading.value = true;
  try {
    const res = await xmProviderListApi();
    xmProviders.value = res || [];
  } catch (e) {
    console.error(e);
  } finally {
    xmProvidersLoading.value = false;
  }
}

async function loadXmProjects() {
  xmLoading.value = true;
  try {
    const res = await xmProjectListApi();
    xmProjects.value = res || [];
  } catch (e) {
    console.error(e);
  } finally {
    xmLoading.value = false;
  }
}

function openXmProviderAdd() {
  Object.assign(xmProviderForm, {
    id: 0,
    name: '',
    base_url: '',
    auth_type: 0,
    uid: '',
    key: '',
    token: '',
    status: 0,
    remark: '',
  });
  xmProviderModalVisible.value = true;
}

function openXmProviderEdit(row: XMProviderItem) {
  Object.assign(xmProviderForm, { ...row });
  xmProviderModalVisible.value = true;
}

async function submitXmProviderForm() {
  if (!xmProviderForm.name?.trim()) { message.warning('连接名称不能为空'); return; }
  if (!xmProviderForm.base_url?.trim()) { message.warning('API 地址不能为空'); return; }
  xmProviderSaving.value = true;
  try {
    await xmProviderSaveApi({ ...xmProviderForm });
    message.success(xmProviderForm.id ? '连接保存成功' : '连接添加成功');
    xmProviderModalVisible.value = false;
    loadXmProviders();
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    xmProviderSaving.value = false;
  }
}

async function deleteXmProvider(id: number) {
  try {
    await xmProviderDeleteApi(id);
    message.success('删除成功');
    loadXmProviders();
  } catch (e: any) {
    message.error(e?.message || '删除失败');
  }
}

async function testXmProvider(record: XMProviderItem) {
  xmProviderTesting.value = true;
  try {
    const res = await xmProviderTestApi(record.id);
    message.success(`${res?.message || '连接成功'}，拉到 ${res?.project_count || 0} 个项目`);
  } catch (e: any) {
    message.error(e?.message || '连接失败');
  } finally {
    xmProviderTesting.value = false;
  }
}

async function fetchXmProviderProjects(record: XMProviderItem) {
  xmImportLoading.value = true;
  xmImportProvider.value = record;
  try {
    const res = await xmProviderFetchProjectsApi(record.id);
    xmImportProjects.value = Array.isArray(res) ? res : [];
    xmSelectedImportProjectIds.value = xmImportProjects.value.map((item) => item.id);
    Object.assign(xmImportForm, {
      price_multiplier: 1,
      price_addition: 0,
      overwrite_local_price: true,
    });
    xmImportVisible.value = true;
  } catch (e: any) {
    message.error(e?.message || '拉取项目失败');
  } finally {
    xmImportLoading.value = false;
  }
}

async function submitXmImportProjects() {
  if (!xmImportProvider.value) return;
  if (xmSelectedImportProjectIds.value.length === 0) {
    message.warning('请至少选择一个项目');
    return;
  }
  xmImportSubmitting.value = true;
  try {
    const res = await xmProviderImportProjectsApi({
      provider_id: xmImportProvider.value.id,
      project_ids: xmSelectedImportProjectIds.value,
      price_multiplier: xmImportForm.price_multiplier,
      price_addition: xmImportForm.price_addition,
      overwrite_local_price: xmImportForm.overwrite_local_price,
    });
    const summary = res?.summary || {};
    message.success(`导入完成：新增 ${summary.created || 0}，更新 ${summary.updated || 0}`);
    xmImportVisible.value = false;
    loadXmProjects();
    loadXmProviders();
  } catch (e: any) {
    message.error(e?.message || '导入失败');
  } finally {
    xmImportSubmitting.value = false;
  }
}

function openXmSync(record: XMProviderItem) {
  xmSyncProvider.value = record;
  Object.assign(xmSyncForm, {
    provider_id: record.id,
    sync_name: true,
    sync_description: true,
    sync_upstream_price: true,
    sync_query: true,
    sync_password: true,
    overwrite_local_price: false,
    price_multiplier: 1,
    price_addition: 0,
  });
  xmSyncVisible.value = true;
}

async function submitXmSyncProjects() {
  xmSyncLoading.value = true;
  try {
    const res = await xmProviderSyncProjectsApi({ ...xmSyncForm });
    const summary = res?.summary || {};
    message.success(`同步完成：更新 ${summary.updated || 0}，跳过 ${summary.skipped || 0}`);
    xmSyncVisible.value = false;
    loadXmProjects();
    loadXmProviders();
  } catch (e: any) {
    message.error(e?.message || '同步失败');
  } finally {
    xmSyncLoading.value = false;
  }
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
    message.success('保存成功');
    xmModalVisible.value = false;
    loadXmProjects();
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    xmSaving.value = false;
  }
}

async function deleteXmProject(id: number) {
  try {
    await xmProjectDeleteApi(id);
    message.success('删除成功');
    loadXmProjects();
  } catch (e: any) {
    message.error(e?.message || '删除失败');
  }
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

// YF打卡项目
const yfdkProjects = ref<YFDKAdminProject[]>([]);
const yfdkLoading = ref(false);
const yfdkEditVisible = ref(false);
const yfdkEditForm = reactive({
  id: 0, sell_price: 0, enabled: 1, sort: 10, content: '',
});

const yfdkColumns = [
  { title: 'ID', dataIndex: 'id', width: 60 },
  { title: 'CID', dataIndex: 'cid', width: 80 },
  { title: '项目名称', dataIndex: 'name', ellipsis: true },
  { title: '说明', dataIndex: 'content', width: 200, ellipsis: true },
  { title: '成本价', key: 'cost_price', width: 90 },
  { title: '售价', key: 'sell_price', width: 90 },
  { title: '排序', dataIndex: 'sort', width: 70 },
  { title: '状态', key: 'enabled', width: 90 },
  { title: '操作', key: 'action', width: 150 },
];

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

// YF打卡项目操作
async function loadYFDKProjects() {
  yfdkLoading.value = true;
  try {
    const res = await getYFDKProjectsApi();
    yfdkProjects.value = res || [];
  } catch (e: any) {
    message.error(e?.message || '加载项目列表失败');
    console.error(e);
  } finally {
    yfdkLoading.value = false;
  }
}

async function syncYFDKProjects() {
  try {
    const res = await syncYFDKProjectsApi();
    message.success(res?.msg || '同步成功');
    loadYFDKProjects();
  } catch (e: any) {
    message.error(e?.message || '同步失败');
  }
}

function openYFDKEdit(record: YFDKAdminProject) {
  Object.assign(yfdkEditForm, {
    id: record.id, sell_price: record.sell_price, enabled: record.enabled, sort: record.sort, content: record.content });
  yfdkEditVisible.value = true;
}

async function submitYFDKEdit() {
  try {
    await updateYFDKProjectApi({ ...yfdkEditForm });
    message.success('保存成功');
    yfdkEditVisible.value = false;
    loadYFDKProjects();
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  }
}

async function deleteYFDKProject(id: number) {
  try {
    await deleteYFDKProjectApi(id);
    message.success('删除成功');
    loadYFDKProjects();
  } catch (e: any) {
    message.error(e?.message || '删除失败');
  }
}

// ========== 加载 ==========
async function loadAll() {
  loading.value = true;
  try {
    const [tutuqgRes, yfdkRes, sxdkRes, hzwRes, tuboshuRes, appuiRes, sdxyRes, ydsjRes, yongyeRes, paperRes, tuzhiRes] = await Promise.allSettled([
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
      tuzhiConfigGetApi(),
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
    if (tuzhiRes.status === 'fulfilled' && tuzhiRes.value) {
      Object.assign(tuzhiConfig, tuzhiRes.value);
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

async function saveTuzhi() {
  tuzhiSaving.value = true;
  try {
    await tuzhiConfigSaveApi({ ...tuzhiConfig });
    message.success('凸知打卡配置保存成功');
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    tuzhiSaving.value = false;
  }
}

async function loadTuzhiGoods() {
  tuzhiGoodsLoading.value = true;
  try {
    const [goodsRes, ovRes] = await Promise.all([
      tuzhiAdminGetGoodsApi(),
      tuzhiGoodsOverridesGetApi(),
    ]);
    tuzhiGoods.value = goodsRes || [];
    tuzhiOverrides.value = ovRes || [];
  } catch (e: any) {
    message.error(e?.message || '加载商品失败');
  } finally {
    tuzhiGoodsLoading.value = false;
  }
}

function getTuzhiOverride(goodsId: number): TuZhiGoodsOverride {
  const found = tuzhiOverrides.value.find(o => o.goods_id === goodsId);
  return found || { goods_id: goodsId, price: 0, enabled: 1 };
}

function setTuzhiOverrideField(goodsId: number, field: 'price' | 'enabled', val: any) {
  const idx = tuzhiOverrides.value.findIndex(o => o.goods_id === goodsId);
  if (idx >= 0) {
    (tuzhiOverrides.value[idx] as any)[field] = val;
  } else {
    const ov: TuZhiGoodsOverride = { goods_id: goodsId, price: 0, enabled: 1 };
    (ov as any)[field] = val;
    tuzhiOverrides.value.push(ov);
  }
}

async function saveTuzhiOverrides() {
  tuzhiOverridesSaving.value = true;
  try {
    await tuzhiGoodsOverridesSaveApi(tuzhiOverrides.value);
    message.success('商品配置保存成功');
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    tuzhiOverridesSaving.value = false;
  }
}

function isConfigured(url: string) {
  return url !== undefined && url !== null && url.trim() !== '';
}

onMounted(() => {
  loadAll();
  loadXmProviders();
  loadXmProjects();
  loadWApps();
  loadYFDKProjects();
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
              <Col :xs="24">
                <Card :bordered="false" class="cfg-card">
                  <template #title>
                    <span class="card-title">YF打卡</span>
                    <Tag v-if="isConfigured(yfdkConfig.base_url)" color="success" class="ml-2">已对接</Tag>
                    <Tag v-else color="default" class="ml-2">未配置</Tag>
                  </template>
                  <template #extra>
                    <Button type="primary" size="small" :loading="yfdkSaving" @click="saveYfdk">保存配置</Button>
                  </template>
                  <Row :gutter="[16, 16]">
                    <Col :xs="24" :md="12">
                      <Form layout="vertical" :colon="false">
                        <FormItem label="API 地址">
                          <Input v-model:value="yfdkConfig.base_url" placeholder="https://dk.blwl.fun/api/" />
                        </FormItem>
                        <FormItem label="Token">
                          <Input.Password v-model:value="yfdkConfig.token" placeholder="请输入 Token" />
                        </FormItem>
                      </Form>
                    </Col>
                    <Col :xs="24" :md="12">
                      <div class="mb-3">
                        <Space class="mb-2">
                          <Button type="primary" size="small" @click="syncYFDKProjects">
                            <template #icon><ReloadOutlined /></template>
                            从上游同步
                          </Button>
                          <Button size="small" @click="loadYFDKProjects" :loading="yfdkLoading">刷新</Button>
                        </Space>
                        <Tag color="processing">{{ yfdkProjects.length }} 个项目</Tag>
                      </div>
                      <Table :columns="yfdkColumns" :data-source="yfdkProjects" :loading="yfdkLoading" :pagination="false" row-key="id" size="small" bordered>
                        <template #bodyCell="{ column, record }">
                          <template v-if="column.key === 'cost_price'">
                            ¥{{ Number(record.cost_price).toFixed(2) }}
                          </template>
                          <template v-else-if="column.key === 'sell_price'">
                            <span class="font-semibold">¥{{ Number(record.sell_price).toFixed(2) }}</span>
                          </template>
                          <template v-else-if="column.key === 'enabled'">
                            <Tag :color="record.enabled === 1 ? 'green' : 'red'">{{ record.enabled === 1 ? '启用' : '禁用' }}</Tag>
                          </template>
                          <template v-else-if="column.key === 'action'">
                            <Space :size="0">
                              <Button type="link" size="small" @click="openYFDKEdit(record)">编辑</Button>
                              <Popconfirm title="确定删除？" @confirm="deleteYFDKProject(record.id)">
                                <Button type="link" danger size="small">删除</Button>
                              </Popconfirm>
                            </Space>
                          </template>
                        </template>
                      </Table>
                    </Col>
                  </Row>
                </Card>
              </Col>

              <!-- 凸知打卡 -->
              <Col :xs="24">
                <Card :bordered="false" class="cfg-card">
                  <template #title>
                    <span class="card-title">凸知打卡</span>
                    <Tag v-if="isConfigured(tuzhiConfig.daka_api_username)" color="success" class="ml-2">已对接</Tag>
                    <Tag v-else color="default" class="ml-2">未配置</Tag>
                  </template>
                  <template #extra>
                    <Button type="primary" size="small" :loading="tuzhiSaving" @click="saveTuzhi">保存配置</Button>
                  </template>
                  <Row :gutter="[16, 16]">
                    <Col :xs="24" :md="12">
                      <Form layout="vertical" :colon="false">
                        <FormItem label="上游账号">
                          <Input v-model:value="tuzhiConfig.daka_api_username" placeholder="请输入上游账号" />
                        </FormItem>
                        <FormItem label="上游密码">
                          <Input.Password v-model:value="tuzhiConfig.daka_api_password" placeholder="请输入上游密码" />
                        </FormItem>
                      </Form>
                    </Col>
                    <Col :xs="24" :md="12">
                      <div class="mb-3">
                        <Space class="mb-2">
                          <Button type="primary" size="small" @click="loadTuzhiGoods" :loading="tuzhiGoodsLoading">
                            <template #icon><ReloadOutlined /></template>
                            拉取商品
                          </Button>
                          <Button size="small" :loading="tuzhiOverridesSaving" @click="saveTuzhiOverrides" type="primary" ghost>保存商品配置</Button>
                        </Space>
                        <Tag color="processing">{{ tuzhiGoods.length }} 个商品</Tag>
                      </div>
                      <Table :columns="tuzhiGoodsColumns" :data-source="tuzhiGoods" :loading="tuzhiGoodsLoading"
                             :pagination="false" row-key="id" size="small" bordered>
                        <template #bodyCell="{ column, record }">
                          <template v-if="column.key === 'upstream_price'">
                            ¥{{ Number(record.price).toFixed(2) }}
                          </template>
                          <template v-else-if="column.key === 'billing'">
                            <Tag :color="record.billing_method === 2 ? 'orange' : 'blue'">
                              {{ record.billing_method === 2 ? '按月' : '按日' }}
                            </Tag>
                          </template>
                          <template v-else-if="column.key === 'override_price'">
                            <InputNumber :value="getTuzhiOverride(record.id).price" :min="0" :step="0.01" :precision="2" size="small" style="width:90px"
                              @change="(v: any) => setTuzhiOverrideField(record.id, 'price', v || 0)" />
                          </template>
                          <template v-else-if="column.key === 'enabled'">
                            <Select :value="getTuzhiOverride(record.id).enabled" size="small" style="width:70px"
                              @change="(v: any) => setTuzhiOverrideField(record.id, 'enabled', v)">
                              <SelectOption :value="1">上架</SelectOption>
                              <SelectOption :value="0">下架</SelectOption>
                            </Select>
                          </template>
                        </template>
                      </Table>
                    </Col>
                  </Row>
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
                    <FormItem label="页面显示控制">
                      <div class="grid grid-cols-2 gap-2">
                        <label v-for="opt in tuboshuPageOptions" :key="opt.key" class="flex items-center gap-1 text-sm">
                          <input type="checkbox" :checked="tuboshuConfig.page_visibility?.[opt.key] !== false"
                            @change="(e: Event) => { if (!tuboshuConfig.page_visibility) tuboshuConfig.page_visibility = {}; tuboshuConfig.page_visibility[opt.key] = (e.target as HTMLInputElement).checked; }" />
                          {{ opt.label }}
                        </label>
                      </div>
                    </FormItem>
                  </Form>
                  <Alert message="Token 在货源中心配置，详细价格可在论文页面管理。" type="info" show-icon class="mt-2" />
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
                    <div class="text-sm font-medium mb-2 mt-1">论文价格配置（元）</div>
                    <Row :gutter="12">
                      <Col :span="8"><FormItem label="6000字"><Input v-model:value="paperConfig.lunwen_api_6000_price" /></FormItem></Col>
                      <Col :span="8"><FormItem label="8000字"><Input v-model:value="paperConfig.lunwen_api_8000_price" /></FormItem></Col>
                      <Col :span="8"><FormItem label="10000字"><Input v-model:value="paperConfig.lunwen_api_10000_price" /></FormItem></Col>
                    </Row>
                    <Row :gutter="12">
                      <Col :span="8"><FormItem label="12000字"><Input v-model:value="paperConfig.lunwen_api_12000_price" /></FormItem></Col>
                      <Col :span="8"><FormItem label="15000字"><Input v-model:value="paperConfig.lunwen_api_15000_price" /></FormItem></Col>
                      <Col :span="8"><FormItem label="任务书"><Input v-model:value="paperConfig.lunwen_api_rws_price" /></FormItem></Col>
                    </Row>
                    <Row :gutter="12">
                      <Col :span="8"><FormItem label="开题报告"><Input v-model:value="paperConfig.lunwen_api_ktbg_price" /></FormItem></Col>
                      <Col :span="8"><FormItem label="降AIGC+查重"><Input v-model:value="paperConfig.lunwen_api_jdaigchj_price" /></FormItem></Col>
                      <Col :span="8"><FormItem label="段落修改"><Input v-model:value="paperConfig.lunwen_api_xgdl_price" /></FormItem></Col>
                    </Row>
                    <Row :gutter="12">
                      <Col :span="8"><FormItem label="文本降重"><Input v-model:value="paperConfig.lunwen_api_jcl_price" /></FormItem></Col>
                      <Col :span="8"><FormItem label="降AIGC率"><Input v-model:value="paperConfig.lunwen_api_jdaigcl_price" /></FormItem></Col>
                    </Row>
                  </Form>
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
                    <FormItem label="API 路径">
                      <Input v-model:value="sdxyModuleConfig.endpoint" placeholder="/flash/api.php" />
                    </FormItem>
                    <FormItem label="上游 UID">
                      <Input v-model:value="sdxyModuleConfig.uid" placeholder="请输入 UID" />
                    </FormItem>
                    <FormItem label="密钥">
                      <Input.Password v-model:value="sdxyModuleConfig.key" placeholder="请输入密钥" />
                    </FormItem>
                    <Row :gutter="12">
                      <Col :span="12">
                        <FormItem label="每次价格">
                          <InputNumber v-model:value="sdxyModuleConfig.price" :min="0" :step="0.5" :precision="2" class="w-full" />
                        </FormItem>
                      </Col>
                      <Col :span="12">
                        <FormItem label="超时(秒)">
                          <InputNumber v-model:value="sdxyModuleConfig.timeout" :min="5" :max="120" :step="5" class="w-full" />
                        </FormItem>
                      </Col>
                    </Row>
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
                      <Input v-model:value="ydsjConfig.base_url" placeholder="http://xxx.xxx.xxx:7799/LearnExp" />
                    </FormItem>
                    <FormItem label="上游 UID">
                      <Input v-model:value="ydsjConfig.uid" placeholder="对接站用户UID" />
                    </FormItem>
                    <FormItem label="密钥">
                      <Input.Password v-model:value="ydsjConfig.key" placeholder="对接站密钥" />
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
            <Card :bordered="false" class="cfg-card">
              <template #title>
                <span class="card-title">小米运动连接</span>
                <Tag color="processing" class="ml-2">{{ xmProviders.length }} 条连接</Tag>
              </template>
              <template #extra>
                <Space>
                  <Button size="small" @click="loadXmProviders" :loading="xmProvidersLoading">刷新</Button>
                  <Button type="primary" size="small" @click="openXmProviderAdd">
                    <template #icon><PlusOutlined /></template>
                    添加连接
                  </Button>
                </Space>
              </template>
              <Alert
                class="mb-3"
                type="info"
                show-icon
                message="先配置上游连接，再拉取 getProjects 批量导入项目；倍率和附加价会直接算出本地基础售价。"
              />
              <Table :columns="xmProviderColumns" :data-source="xmProviders" :loading="xmProvidersLoading" :pagination="false" row-key="id" size="small">
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'auth_type'">
                    <Tag :color="record.auth_type === 0 ? 'blue' : 'green'">{{ record.auth_type === 0 ? 'Key' : 'Token' }}</Tag>
                  </template>
                  <template v-else-if="column.key === 'status'">
                    <Tag :color="record.status === 0 ? 'green' : 'red'">{{ record.status === 0 ? '正常' : '停用' }}</Tag>
                  </template>
                  <template v-else-if="column.key === 'action'">
                    <Space :size="0" wrap>
                      <Button type="link" size="small" @click="openXmProviderEdit(record)">编辑</Button>
                      <Button type="link" size="small" :loading="xmProviderTesting" @click="testXmProvider(record)">测试</Button>
                      <Button type="link" size="small" @click="fetchXmProviderProjects(record)">导入项目</Button>
                      <Button type="link" size="small" @click="openXmSync(record)">一键同步</Button>
                      <Popconfirm title="确定删除连接？" @confirm="deleteXmProvider(record.id)">
                        <Button type="link" danger size="small">删除</Button>
                      </Popconfirm>
                    </Space>
                  </template>
                </template>
              </Table>
            </Card>

            <Card :bordered="false" class="cfg-card">
              <template #title>
                <span class="card-title">小米运动项目</span>
                <Tag color="processing" class="ml-2">{{ xmProjects.length }} 个项目</Tag>
              </template>
              <template #extra>
                <Button size="small" @click="loadXmProjects" :loading="xmLoading">刷新</Button>
              </template>
              <Table :columns="xmColumns" :data-source="xmProjects" :loading="xmLoading" :pagination="false" row-key="id" size="small">
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'status'">
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

    <Modal v-model:open="xmProviderModalVisible" :title="xmProviderForm.id ? '编辑连接' : '添加连接'" @ok="submitXmProviderForm" :confirm-loading="xmProviderSaving" ok-text="保存" width="640px">
      <Form layout="vertical" :colon="false" class="mt-3">
        <Row :gutter="16">
          <Col :span="12">
            <FormItem label="连接名称">
              <Input v-model:value="xmProviderForm.name" placeholder="例如：Spiderman 主号" />
            </FormItem>
          </Col>
          <Col :span="12">
            <FormItem label="认证类型">
              <Select v-model:value="xmProviderForm.auth_type" class="w-full">
                <SelectOption :value="0">Key 认证</SelectOption>
                <SelectOption :value="1">Token 认证</SelectOption>
              </Select>
            </FormItem>
          </Col>
        </Row>
        <FormItem label="API 地址">
          <Input v-model:value="xmProviderForm.base_url" placeholder="例如：https://spiderman.sbs/api/xm_apis.php" />
        </FormItem>
        <Row :gutter="16">
          <Col :span="12">
            <FormItem label="UID" v-if="xmProviderForm.auth_type === 0">
              <Input v-model:value="xmProviderForm.uid" placeholder="上游用户 UID" />
            </FormItem>
            <FormItem label="Token" v-else>
              <Input.Password v-model:value="xmProviderForm.token" placeholder="上游 Token" />
            </FormItem>
          </Col>
          <Col :span="12">
            <FormItem label="Key" v-if="xmProviderForm.auth_type === 0">
              <Input.Password v-model:value="xmProviderForm.key" placeholder="上游 API Key" />
            </FormItem>
            <FormItem label="状态" v-else>
              <Select v-model:value="xmProviderForm.status" class="w-full">
                <SelectOption :value="0">正常</SelectOption>
                <SelectOption :value="1">停用</SelectOption>
              </Select>
            </FormItem>
          </Col>
        </Row>
        <FormItem label="状态" v-if="xmProviderForm.auth_type === 0">
          <Select v-model:value="xmProviderForm.status" class="w-full">
            <SelectOption :value="0">正常</SelectOption>
            <SelectOption :value="1">停用</SelectOption>
          </Select>
        </FormItem>
        <FormItem label="备注">
          <Input.TextArea v-model:value="xmProviderForm.remark" :rows="2" placeholder="记录这个连接的来源、限制或同步说明" />
        </FormItem>
      </Form>
    </Modal>

    <Modal v-model:open="xmImportVisible" title="导入上游项目" @ok="submitXmImportProjects" :confirm-loading="xmImportSubmitting" ok-text="确认导入" width="920px">
      <Alert
        class="mb-3"
        type="info"
        show-icon
        :message="xmImportProvider ? `当前连接：${xmImportProvider.name}` : '请选择连接'"
      />
      <Row :gutter="16" class="mb-3">
        <Col :span="8">
          <FormItem label="本地价格倍率">
            <InputNumber v-model:value="xmImportForm.price_multiplier" :min="0" :step="0.1" :precision="2" class="w-full" />
          </FormItem>
        </Col>
        <Col :span="8">
          <FormItem label="附加价">
            <InputNumber v-model:value="xmImportForm.price_addition" :step="0.1" :precision="2" class="w-full" />
          </FormItem>
        </Col>
        <Col :span="8">
          <FormItem label="覆盖已有本地价格">
            <Select v-model:value="xmImportForm.overwrite_local_price" class="w-full">
              <SelectOption :value="true">是</SelectOption>
              <SelectOption :value="false">否</SelectOption>
            </Select>
          </FormItem>
        </Col>
      </Row>
      <Table
        :columns="xmImportColumns"
        :data-source="xmImportProjects"
        :loading="xmImportLoading"
        :pagination="false"
        row-key="id"
        size="small"
        :row-selection="{
          selectedRowKeys: xmSelectedImportProjectIds,
          onChange: handleXmImportSelectionChange,
        }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'query'">
            <Tag :color="record.query === 1 ? 'green' : 'default'">{{ record.query === 1 ? '支持' : '不支持' }}</Tag>
          </template>
          <template v-else-if="column.key === 'password'">
            <Tag :color="record.password === 1 ? 'orange' : 'default'">{{ record.password === 1 ? '需要' : '不需要' }}</Tag>
          </template>
        </template>
      </Table>
    </Modal>

    <Modal v-model:open="xmSyncVisible" title="同步上游项目" @ok="submitXmSyncProjects" :confirm-loading="xmSyncLoading" ok-text="开始同步" width="700px">
      <Alert
        class="mb-3"
        type="info"
        show-icon
        :message="xmSyncProvider ? `将从连接 ${xmSyncProvider.name} 重新拉取项目并更新本地映射。` : '请选择连接'"
      />
      <Row :gutter="16">
        <Col :span="8">
          <FormItem label="同步名称">
            <Select v-model:value="xmSyncForm.sync_name" class="w-full">
              <SelectOption :value="true">是</SelectOption>
              <SelectOption :value="false">否</SelectOption>
            </Select>
          </FormItem>
        </Col>
        <Col :span="8">
          <FormItem label="同步说明">
            <Select v-model:value="xmSyncForm.sync_description" class="w-full">
              <SelectOption :value="true">是</SelectOption>
              <SelectOption :value="false">否</SelectOption>
            </Select>
          </FormItem>
        </Col>
        <Col :span="8">
          <FormItem label="同步上游价格">
            <Select v-model:value="xmSyncForm.sync_upstream_price" class="w-full">
              <SelectOption :value="true">是</SelectOption>
              <SelectOption :value="false">否</SelectOption>
            </Select>
          </FormItem>
        </Col>
        <Col :span="8">
          <FormItem label="同步查询能力">
            <Select v-model:value="xmSyncForm.sync_query" class="w-full">
              <SelectOption :value="true">是</SelectOption>
              <SelectOption :value="false">否</SelectOption>
            </Select>
          </FormItem>
        </Col>
        <Col :span="8">
          <FormItem label="同步密码要求">
            <Select v-model:value="xmSyncForm.sync_password" class="w-full">
              <SelectOption :value="true">是</SelectOption>
              <SelectOption :value="false">否</SelectOption>
            </Select>
          </FormItem>
        </Col>
        <Col :span="8">
          <FormItem label="重算本地价格">
            <Select v-model:value="xmSyncForm.overwrite_local_price" class="w-full">
              <SelectOption :value="true">是</SelectOption>
              <SelectOption :value="false">否</SelectOption>
            </Select>
          </FormItem>
        </Col>
        <Col :span="12">
          <FormItem label="本地价格倍率">
            <InputNumber v-model:value="xmSyncForm.price_multiplier" :min="0" :step="0.1" :precision="2" class="w-full" />
          </FormItem>
        </Col>
        <Col :span="12">
          <FormItem label="附加价">
            <InputNumber v-model:value="xmSyncForm.price_addition" :step="0.1" :precision="2" class="w-full" />
          </FormItem>
        </Col>
      </Row>
    </Modal>

    <Modal v-model:open="xmModalVisible" title="编辑项目" @ok="submitXmForm" :confirm-loading="xmSaving" ok-text="保存" width="640px">
      <Form layout="vertical" :colon="false" class="mt-3">
        <Row :gutter="16">
          <Col :span="12">
            <FormItem label="项目名称">
              <Input v-model:value="xmForm.name" placeholder="本地展示名称" />
            </FormItem>
          </Col>
          <Col :span="12">
            <FormItem label="来源连接">
              <Input :value="xmForm.provider_name" disabled />
            </FormItem>
          </Col>
        </Row>
        <Row :gutter="16">
          <Col :span="12">
            <FormItem label="上游项目ID">
              <Input :value="xmForm.p_id" disabled />
            </FormItem>
          </Col>
          <Col :span="12">
            <FormItem label="排序">
              <InputNumber v-model:value="xmForm.sort_order" :min="0" :step="1" class="w-full" />
            </FormItem>
          </Col>
        </Row>
        <FormItem label="项目说明">
          <Input.TextArea v-model:value="xmForm.description" :rows="2" placeholder="本地展示说明" />
        </FormItem>
        <Row :gutter="16">
          <Col :span="12">
            <FormItem label="本地基础售价">
              <InputNumber v-model:value="xmForm.price" :min="0" :step="0.1" :precision="4" class="w-full" />
            </FormItem>
          </Col>
          <Col :span="12">
            <FormItem label="上游价格">
              <InputNumber v-model:value="xmForm.upstream_price" :min="0" :step="0.1" :precision="4" class="w-full" disabled />
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
            <FormItem label="支持查询">
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

    <!-- YF打卡编辑弹窗 -->
    <Modal v-model:open="yfdkEditVisible" title="编辑项目" @ok="submitYFDKEdit" ok-text="保存" width="480px">
      <Form layout="vertical" :colon="false" class="mt-3">
        <FormItem label="说明">
          <Input.TextArea v-model:value="yfdkEditForm.content" :rows="2" placeholder="项目描述" />
        </FormItem>
        <Row :gutter="16">
          <Col :span="12">
            <FormItem label="售价">
              <InputNumber v-model:value="yfdkEditForm.sell_price" :min="0" :step="0.01" :precision="2" class="w-full" />
            </FormItem>
          </Col>
          <Col :span="12">
            <FormItem label="排序">
              <InputNumber v-model:value="yfdkEditForm.sort" :min="0" :step="1" class="w-full" />
            </FormItem>
          </Col>
        </Row>
        <FormItem label="状态">
          <Select v-model:value="yfdkEditForm.enabled" class="w-full">
            <SelectOption :value="1">启用</SelectOption>
            <SelectOption :value="0">禁用</SelectOption>
          </Select>
        </FormItem>
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

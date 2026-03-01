<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Tag, Space, Tabs, TabPane, Switch, InputNumber,
  Modal, message, Select, SelectOption, Spin, Statistic, Badge, Alert,
  Popconfirm, Empty, Tooltip, Collapse, CollapsePanel,
} from 'ant-design-vue';
import {
  SyncOutlined, SettingOutlined, FileTextOutlined, PlayCircleOutlined,
  EyeOutlined, CheckCircleOutlined, WarningOutlined, QuestionCircleOutlined,
  ToolOutlined, ThunderboltOutlined, SaveOutlined,
} from '@ant-design/icons-vue';
import {
  getSyncConfigApi, saveSyncConfigApi, syncPreviewApi, syncExecuteApi,
  getSyncLogsApi, getMonitoredSuppliersApi,
  getLonglongToolConfigApi, saveLonglongToolConfigApi,
  longlongToolSyncApi, getLonglongToolStatusApi,
  type SyncConfig, type SyncPreviewResult, type SyncLogItem, type MonitoredSupplier,
  type LonglongToolConfig, type LonglongToolStatus,
} from '#/api/sync-monitor';
import { getSupplierListApi, type SupplierItem } from '#/api/admin';

const activeTab = ref('dashboard');

// ===== 配置 =====
const config = ref<Partial<SyncConfig>>({
  supplier_ids: '', price_rates: {}, category_rates: {},
  sync_price: true, sync_status: true, sync_content: true, sync_name: false,
  clone_enabled: false, force_price_up: false,
  clone_category: false, skip_categories: [], name_replace: {},
  secret_price_rate: 0, auto_sync_enabled: false, auto_sync_interval: 30,
});
const configLoading = ref(false);
const allSuppliers = ref<SupplierItem[]>([]);
const selectedHIDs = ref<number[]>([]);

async function loadConfig() {
  configLoading.value = true;
  try {
    const [cfgRes, supRes] = await Promise.all([getSyncConfigApi(), getSupplierListApi()]);
    const cfg = cfgRes;
    Object.assign(config.value, cfg);
    allSuppliers.value = (supRes) || [];
    selectedHIDs.value = (cfg.supplier_ids || '').split(',').filter((s: string) => s).map(Number);
  } catch (e) { console.error(e); }
  finally { configLoading.value = false; }
}

function onSupplierChange(val: number[]) {
  selectedHIDs.value = val;
  config.value.supplier_ids = val.join(',');
  // 为新增的货源设置默认倍率
  val.forEach(hid => {
    const key = String(hid);
    if (!config.value.price_rates![key]) {
      config.value.price_rates![key] = 1;
    }
  });
}

async function saveConfig() {
  try {
    await saveSyncConfigApi(config.value);
    message.success('配置已保存');
    loadDashboard();
  } catch (e: any) { message.error(e?.message || '保存失败'); }
}

// ===== 名称替换管理 =====
const newReplaceOld = ref('');
const newReplaceNew = ref('');
function addNameReplace() {
  if (!newReplaceOld.value) return;
  if (!config.value.name_replace) config.value.name_replace = {};
  config.value.name_replace[newReplaceOld.value] = newReplaceNew.value;
  newReplaceOld.value = '';
  newReplaceNew.value = '';
}
function removeNameReplace(key: string) {
  if (config.value.name_replace) {
    delete config.value.name_replace[key];
    config.value.name_replace = { ...config.value.name_replace };
  }
}

// ===== 跳过分类管理 =====
const newSkipCat = ref('');
function addSkipCategory() {
  if (!newSkipCat.value) return;
  if (!config.value.skip_categories) config.value.skip_categories = [];
  if (!config.value.skip_categories.includes(newSkipCat.value)) {
    config.value.skip_categories.push(newSkipCat.value);
  }
  newSkipCat.value = '';
}
function removeSkipCategory(id: string) {
  if (config.value.skip_categories) {
    config.value.skip_categories = config.value.skip_categories.filter(s => s !== id);
  }
}

// ===== 仪表盘 =====
const monitored = ref<MonitoredSupplier[]>([]);
const dashLoading = ref(false);

async function loadDashboard() {
  dashLoading.value = true;
  try {
    const res = await getMonitoredSuppliersApi();
    monitored.value = res ?? [];
  } catch (e) { console.error(e); }
  finally { dashLoading.value = false; }
}

// ===== 预览 & 执行 =====
const previewVisible = ref(false);
const previewLoading = ref(false);
const previewResult = ref<SyncPreviewResult | null>(null);
const executeLoading = ref(false);

async function openPreview(hid: number) {
  previewVisible.value = true;
  previewLoading.value = true;
  previewResult.value = null;
  try {
    const res = await syncPreviewApi(hid);
    previewResult.value = res;
  } catch (e: any) {
    message.error('预览失败: ' + (e?.message || ''));
  } finally { previewLoading.value = false; }
}

async function executeSync() {
  if (!previewResult.value) return;
  executeLoading.value = true;
  try {
    const res = await syncExecuteApi(previewResult.value.supplier_id);
    const data = res;
    message.success(`同步完成：应用 ${data.applied} 项，失败 ${data.failed} 项`);
    previewVisible.value = false;
    loadDashboard();
    loadLogs();
  } catch (e: any) {
    message.error('同步失败: ' + (e?.message || ''));
  } finally { executeLoading.value = false; }
}

const diffColumns = [
  { title: '操作', dataIndex: 'action', width: 100 },
  { title: 'CID', dataIndex: 'cid', width: 70 },
  { title: '商品名称', dataIndex: 'name', ellipsis: true },
  { title: '分类', dataIndex: 'category', width: 100 },
  { title: '变更前', dataIndex: 'old_value', width: 150, ellipsis: true },
  { title: '变更后', dataIndex: 'new_value', width: 150, ellipsis: true },
];

const actionColors: Record<string, string> = {
  '更新价格': 'blue', '更新说明': 'cyan', '更新名称': 'purple',
  '上架': 'green', '下架': 'red', '克隆上架': 'orange', '新增分类': 'geekblue',
};

// ===== 日志 =====
const logs = ref<SyncLogItem[]>([]);
const logTotal = ref(0);
const logPage = ref(1);
const logLoading = ref(false);
const logFilter = ref({ supplier_id: 0, action: '' });

async function loadLogs() {
  logLoading.value = true;
  try {
    const res = await getSyncLogsApi({
      page: logPage.value, page_size: 50,
      supplier_id: logFilter.value.supplier_id || undefined,
      action: logFilter.value.action || undefined,
    });
    const data = res;
    logs.value = data.list || [];
    logTotal.value = data.total || 0;
  } catch (e) { console.error(e); }
  finally { logLoading.value = false; }
}

const logColumns = [
  { title: 'ID', dataIndex: 'id', width: 60 },
  { title: '时间', dataIndex: 'sync_time', width: 160 },
  { title: '货源', dataIndex: 'supplier_name', width: 120 },
  { title: '操作', dataIndex: 'action', width: 100 },
  { title: '商品', dataIndex: 'product_name', ellipsis: true },
  { title: '变更前', dataIndex: 'data_before', width: 150, ellipsis: true },
  { title: '变更后', dataIndex: 'data_after', width: 150, ellipsis: true },
];

const totalDiffs = computed(() => {
  if (!previewResult.value?.summary) return 0;
  return Object.values(previewResult.value.summary).reduce((a, b) => a + b, 0);
});

// ===== 龙龙一键对接工具 =====
const llConfig = ref<Partial<LonglongToolConfig>>({
  long_host: '', access_key: '', docking: '',
  mysql_host: '127.0.0.1', mysql_port: '3306',
  mysql_user: '', mysql_password: '', mysql_database: '',
  class_table: '', order_table: '',
  rate: '1.5', name_prefix: '', category: '', sort: '0',
  cover_price: true, cover_desc: true, cover_name: false,
  cron_value: '30', cron_unit: 'minute',
});
const llLoading = ref(false);
const llSaving = ref(false);
const llSyncing = ref(false);
const llStatus = ref<Partial<LonglongToolStatus>>({});

async function loadLonglongConfig() {
  llLoading.value = true;
  try {
    const [cfgRes, statusRes] = await Promise.all([
      getLonglongToolConfigApi(),
      getLonglongToolStatusApi(),
    ]);
    Object.assign(llConfig.value, cfgRes);
    Object.assign(llStatus.value, statusRes);
  } catch (e) { console.error(e); }
  finally { llLoading.value = false; }
}

async function saveLonglongConfig() {
  llSaving.value = true;
  try {
    await saveLonglongToolConfigApi(llConfig.value);
    message.success('龙龙对接配置已保存');
  } catch (e: any) { message.error(e?.message || '保存失败'); }
  finally { llSaving.value = false; }
}

async function triggerLonglongSync() {
  llSyncing.value = true;
  try {
    const res = await longlongToolSyncApi();
    message.success(res?.msg || '同步完成');
    // 刷新状态
    const st = await getLonglongToolStatusApi();
    Object.assign(llStatus.value, st);
  } catch (e: any) { message.error(e?.message || '同步失败'); }
  finally { llSyncing.value = false; }
}

async function refreshLLStatus() {
  try {
    const st = await getLonglongToolStatusApi();
    Object.assign(llStatus.value, st);
  } catch (e) { console.error(e); }
}

onMounted(() => {
  loadConfig();
  loadDashboard();
  loadLogs();
  loadLonglongConfig();
});
</script>

<template>
  <Page auto-content-height>
    <Card :bordered="false">
      <Tabs v-model:activeKey="activeTab">
        <!-- ========== 仪表盘 ========== -->
        <TabPane key="dashboard">
          <template #tab><SyncOutlined /> 同步概况</template>
          <Spin :spinning="dashLoading">
            <Alert v-if="!monitored.length" message="未配置监听货源，请在「同步设置」中选择要监听的货源" type="info" show-icon class="mb-4" />
            <div v-else class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
              <Card v-for="sup in monitored" :key="sup.hid" size="small" :bordered="true" hoverable>
                <div class="flex items-center justify-between mb-2">
                  <div>
                    <span class="font-semibold text-base">{{ sup.name }}</span>
                    <Tag color="blue" class="ml-2">{{ sup.pt_name }}</Tag>
                  </div>
                  <Tag :color="sup.status === '1' ? 'green' : 'default'">{{ sup.status === '1' ? '启用' : '禁用' }}</Tag>
                </div>
                <div class="grid grid-cols-3 gap-2 mb-3">
                  <Statistic title="本地商品" :value="sup.local_count" :value-style="{ fontSize: '18px' }" />
                  <Statistic title="在架" :value="sup.active_count" :value-style="{ fontSize: '18px', color: '#52c41a' }" />
                  <div>
                    <div class="text-xs text-gray-400 dark:text-gray-500">余额</div>
                    <div class="font-semibold" style="font-size: 18px; color: #1890ff">{{ sup.money || '0' }}</div>
                  </div>
                </div>
                <div class="flex gap-2">
                  <Button type="primary" size="small" @click="openPreview(sup.hid)" :loading="previewLoading && previewResult?.supplier_id === sup.hid">
                    <template #icon><EyeOutlined /></template>
                    预览差异
                  </Button>
                </div>
              </Card>
            </div>
          </Spin>
        </TabPane>

        <!-- ========== 同步设置 ========== -->
        <TabPane key="settings">
          <template #tab><SettingOutlined /> 同步设置</template>
          <Spin :spinning="configLoading">
            <div class="max-w-3xl">
              <Card title="监听货源" size="small" class="mb-4">
                <Select
                  mode="multiple"
                  :value="selectedHIDs"
                  @change="onSupplierChange"
                  placeholder="选择要监听的货源"
                  style="width: 100%"
                  :filter-option="(input: string, option: any) => option.label?.toLowerCase().includes(input.toLowerCase())"
                >
                  <SelectOption v-for="sup in allSuppliers" :key="sup.hid" :value="sup.hid" :label="`${sup.name} (HID:${sup.hid})`">
                    {{ sup.name }} <Tag>HID:{{ sup.hid }}</Tag>
                  </SelectOption>
                </Select>
              </Card>

              <Card size="small" class="mb-4" v-if="selectedHIDs.length">
                <template #title>
                  价格倍率
                  <Tooltip title="本地售价 = 上游原价 × 倍率。例如上游 2 元、倍率 5 → 本地 10 元">
                    <QuestionCircleOutlined class="ml-1 text-gray-400" />
                  </Tooltip>
                </template>
                <div class="space-y-2">
                  <div v-for="hid in selectedHIDs" :key="hid" class="flex items-center gap-3">
                    <span class="w-40 truncate font-medium">{{ allSuppliers.find(s => s.hid === hid)?.name || `HID:${hid}` }}</span>
                    <span class="text-gray-500 text-xs">上游价 ×</span>
                    <InputNumber
                      :value="config.price_rates?.[String(hid)] ?? 1"
                      @update:value="(v: number | null) => { if (config.price_rates) config.price_rates[String(hid)] = v ?? 1; }"
                      :min="0.01" :max="100" :step="0.1" :precision="2"
                      style="width: 100px"
                    />
                    <span class="text-gray-400 dark:text-gray-500 text-xs">= 本地售价</span>
                  </div>
                </div>
              </Card>

              <Card title="同步开关" size="small" class="mb-4">
                <div class="grid grid-cols-2 gap-y-3 gap-x-6">
                  <div class="flex items-center justify-between">
                    <span>同步价格 <Tooltip title="上游价格变动时，按倍率更新本地售价"><QuestionCircleOutlined class="text-gray-400" /></Tooltip></span>
                    <Switch v-model:checked="config.sync_price" />
                  </div>
                  <div class="flex items-center justify-between">
                    <span>同步上下架 <Tooltip title="上游下架→本地也下架（status=0）；上游恢复→本地也恢复"><QuestionCircleOutlined class="text-gray-400" /></Tooltip></span>
                    <Switch v-model:checked="config.sync_status" />
                  </div>
                  <div class="flex items-center justify-between">
                    <span>同步说明 <Tooltip title="上游商品描述变动时，同步更新到本地"><QuestionCircleOutlined class="text-gray-400" /></Tooltip></span>
                    <Switch v-model:checked="config.sync_content" />
                  </div>
                  <div class="flex items-center justify-between">
                    <span>同步名称 <Tooltip title="上游商品名变动时更新本地。默认关闭，避免覆盖你自己改过的名称"><QuestionCircleOutlined class="text-gray-400" /></Tooltip></span>
                    <Switch v-model:checked="config.sync_name" />
                  </div>
                  <div class="flex items-center justify-between">
                    <span>克隆上架 <Tooltip title="上游有新商品但本地没有时，自动创建并上架到本地"><QuestionCircleOutlined class="text-gray-400" /></Tooltip></span>
                    <Switch v-model:checked="config.clone_enabled" />
                  </div>
                  <div class="flex items-center justify-between">
                    <span>克隆分类 <Tooltip title="克隆新商品时，同步创建上游的分类到本地"><QuestionCircleOutlined class="text-gray-400" /></Tooltip></span>
                    <Switch v-model:checked="config.clone_category" />
                  </div>
                  <div class="flex items-center justify-between">
                    <span>只涨不降 <Tooltip title="价格只会上调不会下调。上游涨价你跟涨，上游降价你不降"><QuestionCircleOutlined class="text-gray-400" /></Tooltip></span>
                    <Switch v-model:checked="config.force_price_up" />
                  </div>
                </div>
              </Card>

              <Card size="small" class="mb-4">
                <template #title>
                  密价倍率
                  <Tooltip title="密价 = 本地售价 × 密价倍率。设为 0 表示不设密价"><QuestionCircleOutlined class="ml-1 text-gray-400" /></Tooltip>
                </template>
                <div class="flex items-center gap-3">
                  <span class="text-gray-500 text-xs">本地售价 ×</span>
                  <InputNumber
                    v-model:value="config.secret_price_rate"
                    :min="0" :max="10" :step="0.1" :precision="2"
                    placeholder="0 = 不设密价"
                    style="width: 140px"
                  />
                  <span class="text-gray-400 dark:text-gray-500 text-xs">= 密价（0 为不设置）</span>
                </div>
              </Card>

              <Card size="small" class="mb-4" v-if="config.skip_categories || true">
                <template #title>
                  跳过分类
                  <Tooltip title="填入上游分类ID，这些分类的商品不参与同步"><QuestionCircleOutlined class="ml-1 text-gray-400" /></Tooltip>
                </template>
                <div class="flex flex-wrap gap-2 mb-2" v-if="config.skip_categories?.length">
                  <Tag v-for="id in config.skip_categories" :key="id" closable @close="removeSkipCategory(id)" color="red">
                    分类ID: {{ id }}
                  </Tag>
                </div>
                <div class="flex items-center gap-2">
                  <input v-model="newSkipCat" placeholder="上游分类ID" class="ant-input" style="width: 160px" @keyup.enter="addSkipCategory" />
                  <Button size="small" @click="addSkipCategory">添加</Button>
                </div>
              </Card>

              <Card size="small" class="mb-4">
                <template #title>
                  名称替换
                  <Tooltip title="克隆/同步名称时，自动将商品名中的指定文字替换为新文字"><QuestionCircleOutlined class="ml-1 text-gray-400" /></Tooltip>
                </template>
                <div class="space-y-2 mb-2" v-if="config.name_replace && Object.keys(config.name_replace).length">
                  <div v-for="(newVal, oldVal) in config.name_replace" :key="oldVal" class="flex items-center gap-2">
                    <Tag color="red">{{ oldVal }}</Tag>
                    <span class="text-gray-400">→</span>
                    <Tag color="green">{{ newVal || '(删除)' }}</Tag>
                    <Button size="small" danger type="text" @click="removeNameReplace(String(oldVal))">删除</Button>
                  </div>
                </div>
                <div class="flex items-center gap-2">
                  <input v-model="newReplaceOld" placeholder="原文字" class="ant-input" style="width: 140px" />
                  <span class="text-gray-400">→</span>
                  <input v-model="newReplaceNew" placeholder="替换为（留空=删除）" class="ant-input" style="width: 160px" @keyup.enter="addNameReplace" />
                  <Button size="small" @click="addNameReplace">添加</Button>
                </div>
              </Card>

              <Card size="small" class="mb-4">
                <template #title>
                  自动定时同步
                  <Tooltip title="开启后，系统将按设定的间隔自动执行同步（无需手动预览确认）"><QuestionCircleOutlined class="ml-1 text-gray-400" /></Tooltip>
                </template>
                <div class="flex items-center gap-4">
                  <div class="flex items-center gap-2">
                    <span>启用</span>
                    <Switch v-model:checked="config.auto_sync_enabled" />
                  </div>
                  <div class="flex items-center gap-2" v-if="config.auto_sync_enabled">
                    <span class="text-sm">间隔</span>
                    <InputNumber v-model:value="config.auto_sync_interval" :min="5" :max="1440" style="width: 90px" />
                    <span class="text-xs text-gray-500">分钟</span>
                  </div>
                </div>
                <div v-if="config.auto_sync_enabled" class="text-xs text-orange-500 mt-2 bg-orange-50 p-2 rounded">
                  自动同步将直接应用所有差异（等同于手动执行同步），请确保配置正确后再开启
                </div>
              </Card>

              <Button type="primary" @click="saveConfig">保存配置</Button>
            </div>
          </Spin>
        </TabPane>

        <!-- ========== 更多工具 ========== -->
        <TabPane key="tools">
          <template #tab><ToolOutlined /> 更多工具</template>
          <Spin :spinning="llLoading">
            <div class="max-w-4xl space-y-4">
              <Collapse default-active-key="1" class="bg-white dark:bg-gray-800">
                <CollapsePanel key="1">
                  <template #header>
                    <span class="font-semibold text-base">龙龙一键对接工具</span>
                  </template>
                  <template #extra>
                    <Button type="primary" size="small" :loading="llSaving" @click.stop="saveLonglongConfig">
                      <template #icon><SaveOutlined /></template>
                      保存配置
                    </Button>
                  </template>

                  <Alert message="龙龙平台已内置到系统，无需在服务器上安装命令行工具。保存配置后，系统后台自动执行产品同步和订单监听。" type="info" show-icon class="mb-4" />

                  <!-- 运行状态 -->
                  <Card title="运行状态" size="small" class="mb-4" type="inner">
                    <div class="grid grid-cols-2 gap-4 md:grid-cols-4 mb-3">
                      <div class="text-center">
                        <Badge :status="llStatus.sync_running ? 'processing' : 'default'" />
                        <span class="ml-1 text-sm">产品同步</span>
                        <div class="text-xs text-gray-400 mt-1">{{ llStatus.sync_running ? '运行中' : '未启动' }}</div>
                      </div>
                      <div class="text-center">
                        <Badge :status="llStatus.listen_running ? 'processing' : 'default'" />
                        <span class="ml-1 text-sm">订单监听</span>
                        <div class="text-xs text-gray-400 mt-1">{{ llStatus.listen_running ? '运行中' : '未启动' }}</div>
                      </div>
                      <div class="text-center">
                        <div class="text-sm font-semibold" style="color: #1890ff">{{ llStatus.sync_count || 0 }}</div>
                        <div class="text-xs text-gray-400">累计同步次数</div>
                      </div>
                      <div class="text-center">
                        <div class="text-sm font-semibold" style="color: #52c41a">{{ llStatus.listen_count || 0 }}</div>
                        <div class="text-xs text-gray-400">累计更新订单数</div>
                      </div>
                    </div>
                    <div v-if="llStatus.last_sync_time" class="text-xs text-gray-400 mb-1">
                      上次同步：{{ llStatus.last_sync_time }} — {{ llStatus.last_sync_msg }}
                    </div>
                    <div v-if="llStatus.last_listen_at" class="text-xs text-gray-400 mb-1">
                      上次监听：{{ llStatus.last_listen_at }} — {{ llStatus.last_listen_msg }}
                    </div>
                    <div class="flex gap-2 mt-3">
                      <Button type="primary" :loading="llSyncing" @click="triggerLonglongSync">
                        <template #icon><ThunderboltOutlined /></template>
                        立即同步产品
                      </Button>
                      <Button @click="refreshLLStatus">
                        <template #icon><SyncOutlined /></template>
                        刷新状态
                      </Button>
                    </div>
                  </Card>

                  <!-- 基础对接参数 -->
                  <Card title="1. 基础对接参数" size="small" class="mb-4" type="inner">
                    <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
                      <div>
                        <label class="text-xs font-medium text-gray-600 dark:text-gray-400 mb-1 block">源台 IP 或域名</label>
                        <input v-model="llConfig.long_host" class="ant-input" placeholder="例如: 106.52.65.78" />
                      </div>
                      <div>
                        <label class="text-xs font-medium text-gray-600 dark:text-gray-400 mb-1 block">源台 Access Key</label>
                        <input v-model="llConfig.access_key" class="ant-input" placeholder="请从源台获取" />
                      </div>
                      <div>
                        <label class="text-xs font-medium text-gray-600 dark:text-gray-400 mb-1 block">
                          对接的本地货源ID (docking)
                          <Tooltip title="网课接口配置列表中对应的HID，用于绑定订单"><QuestionCircleOutlined class="ml-1 text-gray-400" /></Tooltip>
                        </label>
                        <input v-model="llConfig.docking" class="ant-input" placeholder="例如: 11" />
                      </div>
                    </div>
                  </Card>

                  <!-- 商品同步规则 -->
                  <Card title="2. 商品同步规则" size="small" class="mb-4" type="inner">
                    <div class="grid grid-cols-1 gap-4 md:grid-cols-4 mb-4">
                      <div>
                        <label class="text-xs font-medium text-gray-600 dark:text-gray-400 mb-1 block">
                          价格倍率
                          <Tooltip title="本地上架价格 = 源台成本价 x 倍率"><QuestionCircleOutlined class="ml-1 text-gray-400" /></Tooltip>
                        </label>
                        <input v-model="llConfig.rate" class="ant-input" placeholder="例如: 8" />
                      </div>
                      <div>
                        <label class="text-xs font-medium text-gray-600 dark:text-gray-400 mb-1 block">放入本地分类ID</label>
                        <input v-model="llConfig.category" class="ant-input" placeholder="例如: 22" />
                      </div>
                      <div>
                        <label class="text-xs font-medium text-gray-600 dark:text-gray-400 mb-1 block">商品名称前缀</label>
                        <input v-model="llConfig.name_prefix" class="ant-input" placeholder="例如: 龙龙-" />
                      </div>
                      <div>
                        <label class="text-xs font-medium text-gray-600 dark:text-gray-400 mb-1 block">商品排序数字</label>
                        <input v-model="llConfig.sort" class="ant-input" placeholder="默认 0" />
                      </div>
                    </div>

                    <div class="bg-gray-50 dark:bg-gray-800 p-3 rounded grid grid-cols-3 gap-y-3 gap-x-6">
                      <div class="flex items-center gap-3">
                        <Switch v-model:checked="llConfig.cover_price" />
                        <span class="text-sm">强制覆盖本地价格</span>
                      </div>
                      <div class="flex items-center gap-3">
                        <Switch v-model:checked="llConfig.cover_desc" />
                        <span class="text-sm">强制覆盖本地介绍</span>
                      </div>
                      <div class="flex items-center gap-3">
                        <Switch v-model:checked="llConfig.cover_name" />
                        <span class="text-sm">强制覆盖本地名称</span>
                      </div>
                    </div>
                    
                    <Collapse ghost class="mt-2">
                      <CollapsePanel key="extra" header="高级表名设置（普通用户无需修改）">
                        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                          <div>
                            <label class="text-xs font-medium text-gray-600 dark:text-gray-400 mb-1 block">商品分类表名 (class-table)</label>
                            <input v-model="llConfig.class_table" class="ant-input" placeholder="留空自动识别" />
                          </div>
                          <div>
                            <label class="text-xs font-medium text-gray-600 dark:text-gray-400 mb-1 block">订单记录表名 (order-table)</label>
                            <input v-model="llConfig.order_table" class="ant-input" placeholder="留空自动识别" />
                          </div>
                        </div>
                      </CollapsePanel>
                    </Collapse>
                  </Card>

                  <!-- 同步频率 -->
                  <Card title="3. 自动同步频率" size="small" class="mb-4" type="inner">
                    <div class="flex items-center gap-2">
                      <span class="text-sm font-medium">每隔</span>
                      <InputNumber v-model:value="llConfig.cron_value" :min="1" :max="99" style="width: 80px" />
                      <Select v-model:value="llConfig.cron_unit" style="width: 100px">
                        <SelectOption value="minute">分钟</SelectOption>
                        <SelectOption value="hour">小时</SelectOption>
                      </Select>
                      <span class="text-sm">自动执行一次产品同步（订单监听间隔更短，自动计算）</span>
                    </div>
                  </Card>
                </CollapsePanel>
              </Collapse>
            </div>
          </Spin>
        </TabPane>

        <!-- ========== 变更日志 ========== -->
        <TabPane key="logs">
          <template #tab><FileTextOutlined /> 变更日志 <Badge :count="logTotal" :overflow-count="9999" :number-style="{ backgroundColor: '#1890ff', fontSize: '10px' }" /></template>
          <div class="mb-3 flex items-center gap-3">
            <Select v-model:value="logFilter.action" placeholder="操作类型" allow-clear style="width: 140px" @change="loadLogs">
              <SelectOption value="">全部</SelectOption>
              <SelectOption value="更新价格">更新价格</SelectOption>
              <SelectOption value="更新说明">更新说明</SelectOption>
              <SelectOption value="更新名称">更新名称</SelectOption>
              <SelectOption value="上架">上架</SelectOption>
              <SelectOption value="下架">下架</SelectOption>
              <SelectOption value="克隆上架">克隆上架</SelectOption>
              <SelectOption value="新增分类">新增分类</SelectOption>
            </Select>
            <Button @click="loadLogs"><template #icon><SyncOutlined /></template></Button>
          </div>
          <Table
            :data-source="logs" :columns="logColumns" :loading="logLoading"
            :pagination="{ current: logPage, total: logTotal, pageSize: 50, onChange: (p: number) => { logPage = p; loadLogs(); }, showTotal: (t: number) => `共 ${t} 条`, size: 'small' }"
            row-key="id" size="small" bordered
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.dataIndex === 'action'">
                <Tag :color="actionColors[record.action] || 'default'">{{ record.action }}</Tag>
              </template>
            </template>
          </Table>
        </TabPane>
      </Tabs>
    </Card>

    <!-- ========== 预览弹窗 ========== -->
    <Modal
      v-model:open="previewVisible"
      :title="`同步预览 - ${previewResult?.supplier_name || ''}`"
      width="900px"
      :footer="null"
    >
      <Spin :spinning="previewLoading">
        <template v-if="previewResult">
          <div class="mb-4 grid grid-cols-4 gap-3">
            <Card size="small"><Statistic title="上游商品" :value="previewResult.upstream_count" /></Card>
            <Card size="small"><Statistic title="本地商品" :value="previewResult.local_count" /></Card>
            <Card size="small"><Statistic title="差异项" :value="totalDiffs" :value-style="{ color: totalDiffs > 0 ? '#fa541c' : '#52c41a' }" /></Card>
            <Card size="small">
              <div class="text-xs text-gray-500 mb-1">差异概况</div>
              <div class="flex flex-wrap gap-1">
                <Tag v-for="(count, act) in previewResult.summary" :key="act" :color="actionColors[act as string] || 'default'">
                  {{ act }}: {{ count }}
                </Tag>
              </div>
            </Card>
          </div>

          <Alert v-if="totalDiffs === 0" message="没有差异，本地商品与上游完全一致" type="success" show-icon class="mb-3" />

          <div v-if="totalDiffs > 0">
            <div class="mb-3 flex items-center justify-between">
              <span class="text-sm text-gray-500">共 {{ totalDiffs }} 项变更</span>
              <Popconfirm title="确认执行同步？所有差异将被应用到本地数据库" @confirm="executeSync">
                <Button type="primary" :loading="executeLoading" danger>
                  <template #icon><PlayCircleOutlined /></template>
                  执行同步
                </Button>
              </Popconfirm>
            </div>
            <Table
              :data-source="previewResult.diffs"
              :columns="diffColumns"
              :pagination="{ pageSize: 20, size: 'small', showTotal: (t: number) => `共 ${t} 条` }"
              :row-key="(r: any) => `${r.action}_${r.cid}_${r.name}`"
              size="small" bordered
              :scroll="{ y: 400 }"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.dataIndex === 'action'">
                  <Tag :color="actionColors[record.action] || 'default'">{{ record.action }}</Tag>
                </template>
                <template v-if="column.dataIndex === 'cid'">
                  {{ record.cid || '-' }}
                </template>
              </template>
            </Table>
          </div>
        </template>
        <Empty v-else-if="!previewLoading" description="加载中..." />
      </Spin>
    </Modal>
  </Page>
</template>

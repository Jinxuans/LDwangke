<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Tag, Space, Tabs, TabPane, Switch, InputNumber,
  Modal, message, Select, SelectOption, Spin, Statistic, Badge, Alert,
  Popconfirm, Empty, Tooltip,
} from 'ant-design-vue';
import {
  SyncOutlined, SettingOutlined, FileTextOutlined, PlayCircleOutlined,
  EyeOutlined, CheckCircleOutlined, WarningOutlined, QuestionCircleOutlined,
} from '@ant-design/icons-vue';
import {
  getSyncConfigApi, saveSyncConfigApi, syncPreviewApi, syncExecuteApi,
  getSyncLogsApi, getMonitoredSuppliersApi,
  type SyncConfig, type SyncPreviewResult, type SyncLogItem, type MonitoredSupplier,
} from '#/api/sync-monitor';
import { getSupplierListApi, type SupplierItem } from '#/api/admin';

const activeTab = ref('dashboard');

// ===== 配置 =====
const config = ref<Partial<SyncConfig>>({
  supplier_ids: '', price_rates: {}, category_rates: {},
  sync_price: true, sync_status: true, sync_content: true, sync_name: false,
  clone_enabled: false, force_price_up: false,
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
  '上架': 'green', '下架': 'red', '克隆上架': 'orange',
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

onMounted(() => {
  loadConfig();
  loadDashboard();
  loadLogs();
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
                    <span>同步上下架 <Tooltip title="上游下架→本地也下架；上游恢复→本地也恢复"><QuestionCircleOutlined class="text-gray-400" /></Tooltip></span>
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
                    <span>只涨不降 <Tooltip title="价格只会上调不会下调。上游涨价你跟涨，上游降价你不降"><QuestionCircleOutlined class="text-gray-400" /></Tooltip></span>
                    <Switch v-model:checked="config.force_price_up" />
                  </div>
                </div>
              </Card>

              <Button type="primary" @click="saveConfig">保存配置</Button>
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

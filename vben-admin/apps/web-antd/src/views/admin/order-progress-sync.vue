<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Button, Row, Col, Statistic, Tag, Spin, InputNumber, Switch, Select, SelectOption, Table, message,
} from 'ant-design-vue';
import {
  SyncOutlined, ClockCircleOutlined, CheckCircleOutlined, CloseCircleOutlined, DashboardOutlined,
} from '@ant-design/icons-vue';
import {
  getOrderProgressSyncStatsApi,
  getOrderProgressSyncLogsApi,
  updateOrderProgressSyncConfigApi,
  runOrderProgressSyncApi,
  getSupplierListApi,
  type OrderProgressSyncStats,
  type OrderProgressSyncLog,
  type SupplierItem,
} from '#/api/admin';

const loading = ref(false);
const saving = ref(false);
const running = ref(false);
const stats = ref<OrderProgressSyncStats | null>(null);
const logs = ref<OrderProgressSyncLog[]>([]);
const suppliers = ref<SupplierItem[]>([]);

const enabled = ref(true);
const intervalSec = ref(120);
const batchEnabled = ref(true);
const batchIntervalSec = ref(120);
const supplierIDs = ref<number[]>([]);
const excludedStatuses = ref<string[]>(['已完成', '已退款', '已取消', '失败']);
const rules = ref<OrderProgressSyncStats['rules']>([]);

const statusOptions = ['已完成', '已退款', '已取消', '失败', '异常'];

function formatLogLines(log: OrderProgressSyncLog): string[] {
  const t = log.time;
  const tag = '[AutoSync]';
  const modeLabel = log.mode === 'batch' ? '主订单批量进度同步' : '主订单自动同步';
  const lines: string[] = [];
  lines.push(`${t} ${tag} 开始执行${modeLabel}`);
  if (log.updated === 0 && log.failed === 0 && !log.error) {
    lines.push(`${t} ${tag} 当前没有可同步的已对接订单`);
  } else {
    if (log.failed > 0 && log.sample_errors?.length) {
      for (const err of log.sample_errors) {
        lines.push(`${t} ${tag} 失败样例: ${err}`);
      }
    }
    if (log.error) {
      lines.push(`${t} ${tag} 错误: ${log.error}`);
    }
  }
  const completeLabel = log.mode === 'batch' ? '批量进度同步完成' : '同步完成';
  lines.push(`${t} ${tag} ${completeLabel}，更新 ${log.updated} 个订单，失败 ${log.failed} 个（耗时 ${log.duration_ms}ms）`);
  return lines;
}

const logText = computed(() => {
  return [...logs.value].reverse().flatMap(formatLogLines).join('\n');
});
const ruleColumns = [
  { title: '时间区间', dataIndex: 'label', key: 'label', width: 120 },
  { title: '启用', dataIndex: 'enabled', key: 'enabled', width: 90 },
  { title: '同步间隔(分钟)', dataIndex: 'interval_minutes', key: 'interval_minutes', width: 160 },
  { title: '规则说明', key: 'desc' },
];

async function loadData() {
  loading.value = true;
  try {
    const [statsRes, logsRes, supplierRes] = await Promise.all([
      getOrderProgressSyncStatsApi(),
      getOrderProgressSyncLogsApi(20),
      getSupplierListApi(),
    ]);
    stats.value = statsRes;
    logs.value = logsRes || [];
    suppliers.value = supplierRes || [];
    enabled.value = statsRes.enabled;
    intervalSec.value = statsRes.interval_sec || 120;
    batchEnabled.value = statsRes.batch_enabled ?? true;
    batchIntervalSec.value = statsRes.batch_interval_sec || 120;
    supplierIDs.value = statsRes.supplier_ids || [];
    excludedStatuses.value = statsRes.excluded_statuses?.length ? statsRes.excluded_statuses : ['已完成', '已退款', '已取消', '失败'];
    rules.value = statsRes.rules || [];
  } catch (e) {
    console.error('加载主订单同步配置失败', e);
    message.error('加载主订单同步配置失败');
  } finally {
    loading.value = false;
  }
}

async function saveConfig() {
  saving.value = true;
  try {
    stats.value = await updateOrderProgressSyncConfigApi({
      enabled: enabled.value,
      interval_sec: Number(intervalSec.value || 0),
      batch_enabled: batchEnabled.value,
      batch_interval_sec: Number(batchIntervalSec.value || 0),
      supplier_ids: supplierIDs.value,
      excluded_statuses: excludedStatuses.value,
      rules: rules.value,
    });
    logs.value = await getOrderProgressSyncLogsApi(20);
    message.success('主订单同步配置已保存');
  } catch (e) {
    console.error('保存主订单同步配置失败', e);
    message.error('保存主订单同步配置失败');
  } finally {
    saving.value = false;
  }
}

async function runNow() {
  running.value = true;
  try {
    stats.value = await runOrderProgressSyncApi();
    logs.value = await getOrderProgressSyncLogsApi(20);
    message.success('已触发主订单同步');
  } catch (e) {
    console.error('执行主订单同步失败', e);
    message.error('执行主订单同步失败');
  } finally {
    running.value = false;
  }
}

onMounted(loadData);
</script>

<template>
  <Page title="主订单同步" content-class="p-4">
    <Spin :spinning="loading">
      <Row :gutter="[16, 16]" class="mb-4">
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="调度频率" :value="stats?.interval_sec || 0" suffix="秒" :value-style="{ color: '#1677ff' }">
              <template #prefix><ClockCircleOutlined /></template>
            </Statistic>
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="批量频率" :value="stats?.batch_interval_sec || 0" suffix="秒" :value-style="{ color: '#13c2c2' }">
              <template #prefix><ClockCircleOutlined /></template>
            </Statistic>
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="启用规则" :value="rules.filter(rule => rule.enabled).length" :value-style="{ color: '#13c2c2' }" />
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="最近更新" :value="stats?.last_updated || 0" :value-style="{ color: '#52c41a' }">
              <template #prefix><CheckCircleOutlined /></template>
            </Statistic>
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="批量更新" :value="stats?.batch_last_updated || 0" :value-style="{ color: '#08979c' }">
              <template #prefix><CheckCircleOutlined /></template>
            </Statistic>
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="最近失败" :value="stats?.last_failed || 0" :value-style="{ color: '#ff4d4f' }">
              <template #prefix><CloseCircleOutlined /></template>
            </Statistic>
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="批量失败" :value="stats?.batch_last_failed || 0" :value-style="{ color: '#d4380d' }">
              <template #prefix><CloseCircleOutlined /></template>
            </Statistic>
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="累计轮次" :value="stats?.total_runs || 0" :value-style="{ color: '#722ed1' }" />
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="批量轮次" :value="stats?.batch_total_runs || 0" :value-style="{ color: '#531dab' }" />
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <div class="text-xs text-gray-500 mb-1">逐单状态</div>
            <Tag :color="stats?.enabled ? (stats?.running ? 'blue' : 'green') : 'default'">
              {{ stats?.enabled ? (stats?.running ? '同步中' : '已启用') : '已停用' }}
            </Tag>
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <div class="text-xs text-gray-500 mb-1">批量状态</div>
            <Tag :color="stats?.batch_enabled ? (stats?.batch_running ? 'blue' : 'cyan') : 'default'">
              {{ stats?.batch_enabled ? (stats?.batch_running ? '同步中' : '已启用') : '已停用' }}
            </Tag>
          </Card>
        </Col>
      </Row>

      <Row :gutter="[16, 16]">
        <Col :xs="24" :lg="14">
          <Card title="同步配置" size="small">
            <div class="space-y-4">
              <div class="flex items-center gap-3">
                <span class="text-gray-500">启用自动同步</span>
                <Switch v-model:checked="enabled" />
              </div>
              <div class="flex items-center gap-3 flex-wrap">
                <span class="text-gray-500">全局调度器每隔</span>
                <InputNumber v-model:value="intervalSec" :min="10" :max="86400" />
                <span class="text-gray-500">秒运行一次</span>
              </div>
              <div class="rounded bg-cyan-50 p-3 text-sm text-cyan-700" style="line-height: 1.8;">
                <strong>批量进度专用调度</strong><br />
                配置了批量进度接口的上游将不再参与下方“时间区间规则”的逐单定时同步，而是改由独立的批量调度器按固定间隔拉取增量变更。
              </div>
              <div class="flex items-center gap-3">
                <span class="text-gray-500">启用批量进度调度</span>
                <Switch v-model:checked="batchEnabled" />
              </div>
              <div class="flex items-center gap-3 flex-wrap">
                <span class="text-gray-500">批量调度器每隔</span>
                <InputNumber v-model:value="batchIntervalSec" :min="10" :max="86400" />
                <span class="text-gray-500">秒运行一次</span>
              </div>
              <div>
                <div class="mb-2 text-gray-500">限制自动同步货源</div>
                <Select
                  mode="multiple"
                  v-model:value="supplierIDs"
                  placeholder="留空表示所有已对接货源"
                  style="width: 100%"
                >
                  <SelectOption v-for="sup in suppliers" :key="sup.hid" :value="sup.hid">
                    {{ sup.name }} (HID:{{ sup.hid }})
                  </SelectOption>
                </Select>
              </div>
              <div>
                <div class="mb-2 text-gray-500">全局终态排除</div>
                <Select
                  mode="multiple"
                  v-model:value="excludedStatuses"
                  placeholder="这些状态的订单不会再自动同步"
                  style="width: 100%"
                >
                  <SelectOption v-for="status in statusOptions" :key="status" :value="status">
                    {{ status }}
                  </SelectOption>
                </Select>
              </div>
              <div>
                <div class="mb-2 text-gray-500">时间区间规则</div>
                <Table :columns="ruleColumns" :data-source="rules" :pagination="false" size="small" row-key="key">
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'enabled'">
                      <Switch v-model:checked="record.enabled" />
                    </template>
                    <template v-else-if="column.key === 'interval_minutes'">
                      <InputNumber v-model:value="record.interval_minutes" :min="1" :max="10080" />
                    </template>
                    <template v-else-if="column.key === 'desc'">
                      <span v-if="record.max_age_hours > 0">
                        订单创建于 {{ record.min_age_hours }}h - {{ record.max_age_hours }}h 前
                      </span>
                      <span v-else>
                        订单创建于 {{ record.min_age_hours }}h 前
                      </span>
                    </template>
                  </template>
                </Table>
              </div>
              <div class="flex gap-2">
                <Button type="primary" :loading="saving" @click="saveConfig">保存配置</Button>
                <Button :loading="running" @click="runNow">
                  <template #icon><SyncOutlined /></template>
                  立即同步一轮
                </Button>
              </div>
            </div>
          </Card>
        </Col>

        <Col :xs="24" :lg="10">
          <Card title="详细信息" size="small">
            <div class="space-y-4">
              <div class="flex items-center justify-between">
                <span class="text-gray-500">上次执行</span>
                <span class="font-medium">{{ stats?.last_run_time || '暂无' }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-gray-500">下次执行</span>
                <span class="font-medium">{{ stats?.next_run_time || '暂无' }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-gray-500">批量下次执行</span>
                <span class="font-medium">{{ stats?.batch_next_run_time || '暂无' }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-gray-500">货源范围</span>
                <span class="font-medium">{{ stats?.supplier_ids?.length ? `${stats.supplier_ids.length} 个指定货源` : '全部货源' }}</span>
              </div>
              <div>
                <div class="text-gray-500 mb-2">当前排除状态</div>
                <div class="flex flex-wrap gap-2">
                  <Tag v-for="status in (stats?.excluded_statuses || [])" :key="status" color="red">
                    {{ status }}
                  </Tag>
                </div>
              </div>
              <div class="rounded bg-blue-50 p-3 text-sm text-blue-600" style="line-height: 1.8;">
                <DashboardOutlined class="mr-1" />
                <strong>说明：</strong><br />
                逐单调度按固定秒数运行一次，运行时再根据订单创建时间命中对应规则，并用 `updatetime` 判断是否到期需要同步。<br />
                批量调度不走时间区间规则，只按批量间隔执行，并且只处理已配置批量进度接口的上游。<br />
                货源范围留空表示全部货源；终态排除用于避免已完成、已退款等订单继续被轮询。
              </div>
              <div class="rounded bg-red-50 p-3 text-sm text-red-600" v-if="stats?.last_error">
                最近错误：{{ stats.last_error }}
              </div>
              <div class="rounded bg-red-50 p-3 text-sm text-red-600" v-if="stats?.batch_last_error">
                批量最近错误：{{ stats.batch_last_error }}
              </div>
            </div>
          </Card>
        </Col>
      </Row>

      <Card title="最近日志" size="small" class="mt-4">
        <pre class="m-0 overflow-auto rounded bg-gray-950 p-3 text-xs leading-6 text-green-400" style="max-height:480px">{{ logText || '暂无日志' }}</pre>
      </Card>
    </Spin>
  </Page>
</template>

<style scoped>
.space-y-4 > * + * {
  margin-top: 16px;
}
</style>

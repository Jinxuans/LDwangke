<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Button, Row, Col, Statistic, Tag, Spin, Progress, InputNumber, Space, Table, message,
} from 'ant-design-vue';
import {
  ThunderboltOutlined, SyncOutlined, PauseCircleOutlined,
  CheckCircleOutlined, CloseCircleOutlined, ClockCircleOutlined,
  DashboardOutlined,
} from '@ant-design/icons-vue';
import {
  getDockSchedulerStatsApi,
  getDockSchedulerLogsApi,
  updateDockSchedulerConfigApi,
  runDockSchedulerApi,
  type DockSchedulerStats,
  type DockSchedulerLog,
} from '#/api/admin';

const loading = ref(false);
const running = ref(false);
const saving = ref(false);
const stats = ref<DockSchedulerStats | null>(null);
const logs = ref<DockSchedulerLog[]>([]);
const intervalSec = ref(30);
const batchLimit = ref(100);
const configTouched = ref(false);
let timer: ReturnType<typeof setInterval> | null = null;

async function loadStats() {
  try {
    const [statsRes, logsRes] = await Promise.all([
      getDockSchedulerStatsApi(),
      getDockSchedulerLogsApi(20),
    ]);
    stats.value = statsRes;
    logs.value = logsRes;
    if (!configTouched.value) {
      intervalSec.value = statsRes.interval_sec || 30;
      batchLimit.value = statsRes.batch_limit || 100;
    }
  } catch (e) {
    console.error('加载待对接调度状态失败', e);
  }
}

async function runNow() {
  running.value = true;
  try {
    stats.value = await runDockSchedulerApi();
    logs.value = await getDockSchedulerLogsApi(20);
  } catch (e) {
    console.error('执行待对接调度失败', e);
    message.error('执行失败');
  } finally {
    running.value = false;
  }
}

async function saveConfig() {
  saving.value = true;
  try {
    stats.value = await updateDockSchedulerConfigApi({
      interval_sec: Number(intervalSec.value || 0),
      batch_limit: Number(batchLimit.value || 0),
    });
    logs.value = await getDockSchedulerLogsApi(20);
    configTouched.value = false;
    message.success('调度配置已保存');
  } catch (e) {
    console.error('保存待对接调度配置失败', e);
    message.error('保存配置失败');
  } finally {
    saving.value = false;
  }
}

function startAutoRefresh() {
  timer = setInterval(loadStats, 3000);
}

const logColumns = [
  { title: '时间', dataIndex: 'time', key: 'time', width: 160 },
  { title: '来源', dataIndex: 'trigger', key: 'trigger', width: 90 },
  { title: '级别', dataIndex: 'level', key: 'level', width: 80 },
  { title: '消息', dataIndex: 'message', key: 'message', ellipsis: true },
  { title: '抓取', dataIndex: 'fetched', key: 'fetched', width: 70 },
  { title: '成功', dataIndex: 'success', key: 'success', width: 70 },
  { title: '失败', dataIndex: 'fail', key: 'fail', width: 70 },
  { title: '耗时', dataIndex: 'duration_ms', key: 'duration_ms', width: 90 },
];

onMounted(async () => {
  loading.value = true;
  await loadStats();
  loading.value = false;
  startAutoRefresh();
});

onUnmounted(() => {
  if (timer) clearInterval(timer);
});
</script>

<template>
  <Page title="待对接订单调度" content-class="p-4">
    <Spin :spinning="loading">
      <Row :gutter="[16, 16]" class="mb-4">
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="调度间隔" :value="stats?.interval_sec || 0" suffix="秒" :value-style="{ color: '#1677ff' }">
              <template #prefix><ClockCircleOutlined /></template>
            </Statistic>
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="每轮批量" :value="stats?.batch_limit || 0" :value-style="{ color: '#1677ff' }">
              <template #prefix><DashboardOutlined /></template>
            </Statistic>
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="运行中" :value="stats?.active || 0" :value-style="{ color: '#13c2c2' }">
              <template #prefix><ThunderboltOutlined /></template>
            </Statistic>
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="待对接" :value="stats?.pending || 0" :value-style="{ color: '#fa8c16' }">
              <template #prefix><PauseCircleOutlined /></template>
            </Statistic>
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="累计成功" :value="stats?.total_success || 0" :value-style="{ color: '#52c41a' }">
              <template #prefix><CheckCircleOutlined /></template>
            </Statistic>
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="累计失败" :value="stats?.total_fail || 0" :value-style="{ color: '#ff4d4f' }">
              <template #prefix><CloseCircleOutlined /></template>
            </Statistic>
          </Card>
        </Col>
      </Row>

      <Row :gutter="[16, 16]">
        <Col :xs="24" :lg="12">
          <Card title="调度状态" size="small">
            <template #extra>
              <Button type="link" size="small" @click="loadStats"><SyncOutlined /> 刷新</Button>
            </template>
            <div class="space-y-4">
              <div class="flex items-center justify-between">
                <span class="text-gray-500">运行状态</span>
                <Tag :color="stats?.running ? 'green' : 'red'">
                  {{ stats?.running ? '运行中' : '已停止' }}
                </Tag>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-gray-500">上次执行</span>
                <span class="font-medium">{{ stats?.last_run_time || '暂无' }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-gray-500">执行来源</span>
                <span class="font-medium">{{ stats?.last_trigger || '暂无' }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-gray-500">本轮抓取</span>
                <span class="font-medium">{{ stats?.last_fetched || 0 }} / {{ stats?.batch_limit || 0 }}</span>
              </div>
              <Progress
                :percent="stats?.batch_limit ? Math.round(((stats?.last_fetched || 0) / stats.batch_limit) * 100) : 0"
                :stroke-color="'#1677ff'"
                :status="'active'"
              />
              <div class="flex items-center justify-between">
                <span class="text-gray-500">本轮成功</span>
                <span class="font-medium text-green-600">{{ stats?.last_success || 0 }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-gray-500">本轮失败</span>
                <span class="font-medium text-red-500">{{ stats?.last_fail || 0 }}</span>
              </div>
              <div class="rounded bg-red-50 p-3 text-sm text-red-600" v-if="stats?.last_error">
                最近错误：{{ stats.last_error }}
              </div>
            </div>
          </Card>
        </Col>

        <Col :xs="24" :lg="12">
          <Card title="调度配置" size="small">
            <div class="space-y-4">
              <div>
                <div class="mb-2 text-gray-500">当前配置</div>
                <div class="text-3xl font-bold" style="color: #1677ff;">
                  {{ stats?.interval_sec || 0 }}s / {{ stats?.batch_limit || 0 }}单
                </div>
              </div>
              <div>
                <div class="mb-2 text-gray-500">简单配置</div>
                <Space wrap>
                  <span class="text-gray-500">间隔</span>
                  <InputNumber v-model:value="intervalSec" :min="5" :max="3600" @change="configTouched = true" />
                  <span class="text-gray-500">秒</span>
                  <span class="text-gray-500">批量</span>
                  <InputNumber v-model:value="batchLimit" :min="1" :max="1000" @change="configTouched = true" />
                  <span class="text-gray-500">单</span>
                  <Button type="primary" :loading="saving" @click="saveConfig">
                    保存配置
                  </Button>
                </Space>
              </div>
              <div>
                <div class="mb-2 text-gray-500">手动执行一轮</div>
                <Space>
                  <Button type="primary" :loading="running" @click="runNow">
                    立即执行
                  </Button>
                  <Button @click="loadStats">
                    刷新日志
                  </Button>
                </Space>
              </div>
              <div class="mt-4 rounded bg-blue-50 p-3 text-sm text-blue-600" style="line-height: 1.8;">
                <DashboardOutlined class="mr-1" />
                <strong>说明：</strong><br />
                系统会按固定间隔扫描 `dockstatus IN (0, 2)` 的主订单。<br />
                其中 `0` 是待对接，`2` 是上次对接失败后的自动重试。<br />
                每轮最多处理设定批量的订单，然后直接调用上游对接。<br />
                “立即执行”会额外触发一轮，不会关闭定时调度。
              </div>
            </div>
          </Card>
        </Col>
      </Row>

      <Card title="运行摘要" size="small" class="mt-4" v-if="stats">
        <Row :gutter="16">
          <Col :span="6">
            <div class="text-center">
              <div class="text-3xl font-bold" style="color: #52c41a;">
                {{ stats.total_success + stats.total_fail > 0 ? Math.round((stats.total_success / (stats.total_success + stats.total_fail)) * 100) : 0 }}%
              </div>
              <div class="text-xs text-gray-500 mt-1">累计成功率</div>
            </div>
          </Col>
          <Col :span="6">
            <div class="text-center">
              <div class="text-3xl font-bold" style="color: #1677ff;">
                {{ stats.total_runs || 0 }}
              </div>
              <div class="text-xs text-gray-500 mt-1">累计轮次</div>
            </div>
          </Col>
          <Col :span="6">
            <div class="text-center">
              <div class="text-3xl font-bold" style="color: #fa8c16;">
                {{ stats.pending || 0 }}
              </div>
              <div class="text-xs text-gray-500 mt-1">当前待对接</div>
            </div>
          </Col>
          <Col :span="6">
            <div class="text-center">
              <div class="text-3xl font-bold" style="color: #13c2c2;">
                {{ stats.last_fetched || 0 }}
              </div>
              <div class="text-xs text-gray-500 mt-1">最近抓取数</div>
            </div>
          </Col>
        </Row>
      </Card>

      <Card title="最近日志" size="small" class="mt-4">
        <Table
          :columns="logColumns"
          :data-source="logs"
          :pagination="false"
          size="small"
          row-key="id"
          :scroll="{ x: 980 }"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'level'">
              <Tag :color="record.level === 'error' ? 'red' : 'blue'">
                {{ record.level }}
              </Tag>
            </template>
            <template v-else-if="column.key === 'trigger'">
              <Tag :color="record.trigger === 'manual' ? 'gold' : record.trigger === 'config' ? 'cyan' : 'blue'">
                {{ record.trigger }}
              </Tag>
            </template>
            <template v-else-if="column.key === 'duration_ms'">
              {{ record.duration_ms }} ms
            </template>
          </template>
        </Table>
      </Card>
    </Spin>
  </Page>
</template>

<style scoped>
.space-y-4 > * + * {
  margin-top: 16px;
}
</style>

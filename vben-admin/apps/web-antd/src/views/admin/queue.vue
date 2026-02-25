<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Button, InputNumber, Row, Col, Statistic, Tag, Space, message, Spin, Progress,
} from 'ant-design-vue';
import {
  ThunderboltOutlined, SyncOutlined, PauseCircleOutlined,
  CheckCircleOutlined, CloseCircleOutlined, ClockCircleOutlined,
  DashboardOutlined, SettingOutlined,
} from '@ant-design/icons-vue';
import {
  getQueueStatsApi, setQueueConcurrencyApi, type QueueStats,
} from '#/api/admin';

const loading = ref(false);
const stats = ref<QueueStats | null>(null);
const newWorkers = ref(5);
const editing = ref(false);
const saving = ref(false);
let timer: ReturnType<typeof setInterval> | null = null;

async function loadStats() {
  try {
    const raw = await getQueueStatsApi();
    stats.value = raw;
    if (!editing.value) {
      newWorkers.value = stats.value?.max_workers || 5;
    }
  } catch (e) {
    console.error('加载队列状态失败', e);
  }
}

async function handleSave() {
  if (newWorkers.value < 1 || newWorkers.value > 100) {
    message.warning('并发数范围: 1 ~ 100');
    return;
  }
  saving.value = true;
  try {
    const raw = await setQueueConcurrencyApi(newWorkers.value);
    stats.value = raw;
    editing.value = false;
    message.success(`并发数已调整为 ${stats.value?.max_workers}`);
  } catch {
    message.error('调整失败');
  } finally {
    saving.value = false;
  }
}

function startAutoRefresh() {
  timer = setInterval(loadStats, 3000);
}

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
  <Page title="对接队列" content-class="p-4">
    <Spin :spinning="loading">
      <!-- 顶部状态概览 -->
      <Row :gutter="[16, 16]" class="mb-4">
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="最大并发" :value="stats?.max_workers || 0" :value-style="{ color: '#1677ff' }">
              <template #prefix><SettingOutlined /></template>
            </Statistic>
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="活跃线程" :value="stats?.active || 0" :value-style="{ color: '#1677ff' }">
              <template #prefix><ThunderboltOutlined /></template>
            </Statistic>
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="排队中" :value="stats?.pending || 0" :value-style="{ color: '#fa8c16' }">
              <template #prefix><ClockCircleOutlined /></template>
            </Statistic>
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="处理中" :value="stats?.processing || 0" :value-style="{ color: '#13c2c2' }">
              <template #prefix><SyncOutlined spin /></template>
            </Statistic>
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="已完成" :value="stats?.completed || 0" :value-style="{ color: '#52c41a' }">
              <template #prefix><CheckCircleOutlined /></template>
            </Statistic>
          </Card>
        </Col>
        <Col :xs="12" :sm="8" :lg="4">
          <Card size="small" class="text-center">
            <Statistic title="失败" :value="stats?.failed || 0" :value-style="{ color: '#ff4d4f' }">
              <template #prefix><CloseCircleOutlined /></template>
            </Statistic>
          </Card>
        </Col>
      </Row>

      <!-- 队列详情 -->
      <Row :gutter="[16, 16]">
        <Col :xs="24" :lg="12">
          <Card title="队列状态" size="small">
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
                <span class="text-gray-500">队列使用率</span>
                <span class="font-medium">{{ stats?.queue_size || 0 }} / {{ stats?.queue_cap || 0 }}</span>
              </div>
              <Progress
                :percent="stats?.queue_cap ? Math.round(((stats?.queue_size || 0) / stats.queue_cap) * 100) : 0"
                :stroke-color="(stats?.queue_size || 0) / (stats?.queue_cap || 1) > 0.8 ? '#ff4d4f' : '#1677ff'"
                :status="'active'"
              />
              <div class="flex items-center justify-between">
                <span class="text-gray-500">Worker 使用率</span>
                <span class="font-medium">{{ stats?.active || 0 }} / {{ stats?.max_workers || 0 }}</span>
              </div>
              <Progress
                :percent="stats?.max_workers ? Math.round(((stats?.active || 0) / stats.max_workers) * 100) : 0"
                :stroke-color="'#13c2c2'"
                :status="'active'"
              />
            </div>
          </Card>
        </Col>

        <Col :xs="24" :lg="12">
          <Card title="并发设置" size="small">
            <div class="space-y-4">
              <div>
                <div class="mb-2 text-gray-500">当前最大并发数</div>
                <div class="text-3xl font-bold" style="color: #1677ff;">{{ stats?.max_workers || 0 }}</div>
              </div>
              <div>
                <div class="mb-2 text-gray-500">调整并发数（1 ~ 100）</div>
                <Space>
                  <InputNumber
                    v-model:value="newWorkers"
                    :min="1"
                    :max="100"
                    @focus="editing = true"
                    style="width: 120px;"
                  />
                  <Button type="primary" :loading="saving" @click="handleSave">
                    应用
                  </Button>
                </Space>
              </div>
              <div class="mt-4 rounded bg-blue-50 p-3 text-sm text-blue-600" style="line-height: 1.8;">
                <DashboardOutlined class="mr-1" />
                <strong>说明：</strong><br />
                并发数决定同时向上游供应商发起对接请求的最大数量。<br />
                设置过高可能导致上游限流，建议根据供应商承载能力调整。<br />
                调整后立即生效，无需重启服务。
              </div>
            </div>
          </Card>
        </Col>
      </Row>

      <!-- 统计摘要 -->
      <Card title="运行摘要" size="small" class="mt-4" v-if="stats">
        <Row :gutter="16">
          <Col :span="6">
            <div class="text-center">
              <div class="text-3xl font-bold" style="color: #52c41a;">
                {{ stats.completed + stats.failed > 0 ? Math.round((stats.completed / (stats.completed + stats.failed)) * 100) : 0 }}%
              </div>
              <div class="text-xs text-gray-500 mt-1">成功率</div>
            </div>
          </Col>
          <Col :span="6">
            <div class="text-center">
              <div class="text-3xl font-bold" style="color: #1677ff;">
                {{ (stats.completed || 0) + (stats.failed || 0) }}
              </div>
              <div class="text-xs text-gray-500 mt-1">总处理量</div>
            </div>
          </Col>
          <Col :span="6">
            <div class="text-center">
              <div class="text-3xl font-bold" style="color: #fa8c16;">
                {{ (stats.pending || 0) + (stats.processing || 0) }}
              </div>
              <div class="text-xs text-gray-500 mt-1">待处理</div>
            </div>
          </Col>
          <Col :span="6">
            <div class="text-center">
              <div class="text-3xl font-bold" style="color: #13c2c2;">
                {{ stats.active || 0 }}
              </div>
              <div class="text-xs text-gray-500 mt-1">当前工作中</div>
            </div>
          </Col>
        </Row>
      </Card>
    </Spin>
  </Page>
</template>

<style scoped>
.space-y-4 > * + * {
  margin-top: 16px;
}
</style>

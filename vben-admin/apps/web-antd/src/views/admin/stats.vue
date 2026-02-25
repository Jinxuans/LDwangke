<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Row, Col, Table, Tag, Spin, Select, SelectOption, Space,
} from 'ant-design-vue';
import { getStatsReportApi, type StatsReport } from '#/api/admin';

const loading = ref(false);
const days = ref(30);
const data = ref<StatsReport>({
  daily: [],
  by_class: [],
  by_status: [],
  top_users: [],
});

async function loadStats() {
  loading.value = true;
  try {
    const raw = await getStatsReportApi(days.value);
    data.value = raw;
  } catch (e) {
    console.error('加载统计失败:', e);
  } finally {
    loading.value = false;
  }
}

function handleDaysChange() {
  loadStats();
}

const maxDailyOrders = () => Math.max(...(data.value.daily || []).map((d) => d.orders), 1);
const maxDailyIncome = () => Math.max(...(data.value.daily || []).map((d) => d.income), 1);

const statusColors: Record<string, string> = {
  '待处理': 'default', '进行中': 'processing', '已完成': 'success',
  '异常': 'error', '已退款': 'warning', '已取消': 'default',
};

const classColumns = [
  { title: '课程', dataIndex: 'name', key: 'name', ellipsis: true },
  { title: '订单数', dataIndex: 'count', key: 'count', width: 100, sorter: (a: any, b: any) => a.count - b.count },
  { title: '收入', dataIndex: 'income', key: 'income', width: 120, sorter: (a: any, b: any) => a.income - b.income },
];

const userColumns = [
  { title: '排名', key: 'rank', width: 60, align: 'center' as const },
  { title: '用户', dataIndex: 'username', key: 'username' },
  { title: '订单数', dataIndex: 'orders', key: 'orders', width: 100 },
  { title: '消费金额', dataIndex: 'total', key: 'total', width: 120 },
];

onMounted(loadStats);
</script>

<template>
  <Page title="数据统计" content-class="p-4">
    <div class="mb-4 flex justify-end">
      <Space>
        <span class="text-sm text-gray-500">统计周期：</span>
        <Select v-model:value="days" style="width: 120px" @change="handleDaysChange">
          <SelectOption :value="7">近7天</SelectOption>
          <SelectOption :value="15">近15天</SelectOption>
          <SelectOption :value="30">近30天</SelectOption>
          <SelectOption :value="90">近90天</SelectOption>
        </Select>
      </Space>
    </div>

    <Spin :spinning="loading">
      <!-- 每日趋势 -->
      <Card title="每日订单趋势" class="mb-4">
        <div v-if="data.daily && data.daily.length > 0">
          <div class="flex items-end gap-1" style="height: 200px; overflow-x: auto">
            <div
              v-for="item in data.daily"
              :key="item.date"
              class="flex flex-col items-center justify-end"
              :style="{ minWidth: days <= 15 ? '40px' : '20px', flex: '1' }"
            >
              <span class="text-xs text-gray-500 mb-1" v-if="days <= 15">{{ item.orders }}</span>
              <div
                class="w-full rounded-t bg-blue-500 transition-all min-h-[2px]"
                :style="{ height: `${(item.orders / maxDailyOrders()) * 160}px` }"
                :title="`${item.date}: ${item.orders}单 / ¥${item.income.toFixed(2)}`"
              />
              <span class="text-xs text-gray-400 mt-1" v-if="days <= 15">{{ item.date?.slice(5) }}</span>
            </div>
          </div>
        </div>
        <div v-else class="text-center text-gray-400 py-8">暂无数据</div>
      </Card>

      <Row :gutter="[16, 16]">
        <!-- 状态分布 -->
        <Col :xs="24" :lg="8">
          <Card title="订单状态分布">
            <div v-if="data.by_status && data.by_status.length > 0" class="space-y-3">
              <div v-for="item in data.by_status" :key="item.status" class="flex items-center gap-3">
                <Tag :color="statusColors[item.status] || 'default'" class="min-w-[60px] text-center">
                  {{ item.status }}
                </Tag>
                <div class="flex-1 bg-gray-100 rounded h-5 overflow-hidden">
                  <div
                    class="h-full bg-blue-400 rounded transition-all"
                    :style="{ width: `${(item.count / Math.max(...data.by_status.map(s => s.count), 1)) * 100}%` }"
                  />
                </div>
                <span class="text-sm font-medium min-w-[40px] text-right">{{ item.count }}</span>
              </div>
            </div>
            <div v-else class="text-center text-gray-400 py-8">暂无数据</div>
          </Card>
        </Col>

        <!-- 课程排行 -->
        <Col :xs="24" :lg="16">
          <Card title="课程排行">
            <Table
              :data-source="data.by_class || []"
              :columns="classColumns"
              :pagination="false"
              row-key="name"
              size="small"
              :scroll="{ y: 300 }"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'income'">
                  <span class="text-orange-500 font-medium">¥{{ record.income?.toFixed(2) }}</span>
                </template>
              </template>
            </Table>
          </Card>
        </Col>
      </Row>

      <!-- 用户排行 -->
      <Card title="用户消费排行" class="mt-4">
        <Table
          :data-source="data.top_users || []"
          :columns="userColumns"
          :pagination="false"
          row-key="uid"
          size="small"
        >
          <template #bodyCell="{ column, record, index }">
            <template v-if="column.key === 'rank'">
              <Tag :color="index < 3 ? ['gold', 'silver', '#cd7f32'][index] : 'default'">
                {{ index + 1 }}
              </Tag>
            </template>
            <template v-if="column.key === 'total'">
              <span class="text-orange-500 font-medium">¥{{ record.total?.toFixed(2) }}</span>
            </template>
          </template>
        </Table>
      </Card>
    </Spin>
  </Page>
</template>

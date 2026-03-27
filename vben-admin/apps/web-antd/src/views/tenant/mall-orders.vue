<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { Page } from '@vben/common-ui';
import { Button, Card, Descriptions, DescriptionsItem, Drawer, Space, Table, Tag } from 'ant-design-vue';
import { getOrderLogsApi, type OrderLogEntry } from '#/api/order';
import { getTenantMallLinkedOrdersApi, getTenantMallOrdersApi, type MallPayOrder, type TenantLinkedOrder } from '#/api/tenant';

const loading = ref(false);
const list = ref<MallPayOrder[]>([]);
const total = ref(0);
const page = ref(1);
const router = useRouter();
const detailVisible = ref(false);
const detailLoading = ref(false);
const current = ref<MallPayOrder | null>(null);
const linkedOrders = ref<TenantLinkedOrder[]>([]);
const orderLogs = ref<Record<number, OrderLogEntry[]>>({});
const logLoading = ref<Record<number, boolean>>({});

async function load() {
  loading.value = true;
  try {
    const res = await getTenantMallOrdersApi({ page: page.value, limit: 20 });
    const d = res;
    list.value = d?.list ?? [];
    total.value = d?.total ?? 0;
  } finally {
    loading.value = false;
  }
}

function statusTag(status: number) {
  const map: Record<number, { color: string; text: string }> = {
    0: { color: 'default', text: '待支付' },
    1: { color: 'blue', text: '已支付' },
    2: { color: 'green', text: '已下单' },
    [-1]: { color: 'red', text: '失败' },
  };
  return map[status] ?? { color: 'default', text: '未知' };
}

function payTypeLabel(t: string) {
  const map: Record<string, string> = { alipay: '支付宝', wxpay: '微信', qqpay: 'QQ支付' };
  return map[t] ?? t;
}

function processText(v?: string) {
  if (!v) return '-';
  const n = Number(v);
  return Number.isFinite(n) && `${n}` === `${v}` ? `${n}%` : v;
}

async function openDetail(row: MallPayOrder) {
  current.value = row;
  detailVisible.value = true;
  detailLoading.value = true;
  linkedOrders.value = [];
  try {
    linkedOrders.value = await getTenantMallLinkedOrdersApi(row.id);
  } finally {
    detailLoading.value = false;
  }
}

async function loadLogs(oid: number) {
  if (orderLogs.value[oid] || logLoading.value[oid]) return;
  logLoading.value = { ...logLoading.value, [oid]: true };
  try {
    orderLogs.value = {
      ...orderLogs.value,
      [oid]: await getOrderLogsApi(oid),
    };
  } finally {
    logLoading.value = { ...logLoading.value, [oid]: false };
  }
}

onMounted(load);
</script>

<template>
  <Page title="商城支付订单" content-class="p-4">
    <Card>
      <div class="mb-3 flex items-center justify-between gap-3">
        <div class="text-sm text-gray-500">共 {{ total }} 笔订单</div>
        <Button type="primary" ghost @click="router.push('/tenant/withdraw')">商城提现</Button>
      </div>
      <Table
        :data-source="list"
        :loading="loading"
        :pagination="{
          current: page,
          pageSize: 20,
          total,
          onChange: (p: number) => { page = p; load(); },
        }"
        row-key="id"
        size="small"
        bordered
        :scroll="{ x: 900 }"
      >
        <Table.Column title="订单号" data-index="out_trade_no" :width="200" />
        <Table.Column title="账号" data-index="account" :width="140" />
        <Table.Column title="商品" :width="180">
          <template #default="{ record }">
            <div class="text-xs leading-5">
              <div class="font-medium">{{ record.product_name || `商品#${record.cid}` }}</div>
              <div class="text-gray-400">{{ record.course_name || '-' }}</div>
            </div>
          </template>
        </Table.Column>
        <Table.Column title="支付方式" :width="90">
          <template #default="{ record }">{{ payTypeLabel(record.pay_type) }}</template>
        </Table.Column>
        <Table.Column title="金额" data-index="money" :width="90">
          <template #default="{ record }">¥{{ record.money }}</template>
        </Table.Column>
        <Table.Column title="状态" :width="90">
          <template #default="{ record }">
            <Tag :color="statusTag(record.status).color">{{ statusTag(record.status).text }}</Tag>
          </template>
        </Table.Column>
        <Table.Column title="业务订单ID" data-index="order_id" :width="100">
          <template #default="{ record }">
            <span v-if="record.order_id">{{ record.order_id }}</span>
            <span v-else class="text-gray-400">—</span>
          </template>
        </Table.Column>
        <Table.Column title="进度" :width="180">
          <template #default="{ record }">
            <div class="text-xs leading-5">
              <div>状态：{{ record.order_status || '—' }}</div>
              <div>进度：{{ processText(record.order_process) }}</div>
              <div>关联订单：{{ record.order_count || (record.order_id ? 1 : 0) }}</div>
            </div>
          </template>
        </Table.Column>
        <Table.Column title="备注" :width="220">
          <template #default="{ record }">{{ record.order_remarks || record.remark || '-' }}</template>
        </Table.Column>
        <Table.Column title="下单时间" data-index="addtime" :width="160" />
        <Table.Column title="操作" :width="100" fixed="right">
          <template #default="{ record }">
            <Button size="small" type="link" @click="openDetail(record)">详情</Button>
          </template>
        </Table.Column>
      </Table>
    </Card>

    <Drawer
      v-model:open="detailVisible"
      width="720"
      title="商城订单详情"
      destroy-on-close
    >
      <template v-if="current">
        <Descriptions bordered :column="1" size="small" class="mb-4">
          <DescriptionsItem label="支付订单号">{{ current.out_trade_no }}</DescriptionsItem>
          <DescriptionsItem label="商品">{{ current.product_name || `商品#${current.cid}` }}</DescriptionsItem>
          <DescriptionsItem label="选购课程">{{ current.course_name || '-' }}</DescriptionsItem>
          <DescriptionsItem label="支付金额">¥{{ current.money }}</DescriptionsItem>
          <DescriptionsItem label="支付方式">{{ payTypeLabel(current.pay_type) }}</DescriptionsItem>
          <DescriptionsItem label="支付状态">
            <Tag :color="statusTag(current.status).color">{{ statusTag(current.status).text }}</Tag>
          </DescriptionsItem>
        </Descriptions>

        <Card title="真实订单" :loading="detailLoading">
          <template v-if="linkedOrders.length">
            <div
              v-for="order in linkedOrders"
              :key="order.oid"
              class="mb-4 rounded border border-gray-200 p-3"
            >
              <div class="mb-2 flex items-center justify-between gap-3">
                <div class="font-medium">订单 #{{ order.oid }} / {{ order.kcname || order.ptname }}</div>
                <Space>
                  <Tag color="blue">{{ order.status || '待处理' }}</Tag>
                  <span class="text-xs text-gray-400">{{ processText(order.process) }}</span>
                </Space>
              </div>
              <div class="space-y-1 text-xs text-gray-600">
                <div>账号：{{ order.user }}</div>
                <div>课程ID：{{ order.kcid || '-' }}</div>
                <div>供货价：¥{{ order.fees }}</div>
                <div>备注：{{ order.remarks || '-' }}</div>
                <div>提交时间：{{ order.addtime || '-' }}</div>
              </div>
              <div class="mt-3">
                <Button size="small" @click="loadLogs(order.oid)">查看日志</Button>
              </div>
              <div v-if="logLoading[order.oid]" class="mt-2 text-xs text-gray-400">日志加载中...</div>
              <div v-else-if="orderLogs[order.oid]?.length" class="mt-2 space-y-2">
                <div
                  v-for="(log, idx) in orderLogs[order.oid]"
                  :key="`${order.oid}-${idx}`"
                  class="rounded bg-gray-50 px-3 py-2 text-xs leading-5"
                >
                  <div class="font-medium">{{ log.time || '-' }} / {{ log.status || '-' }} / {{ log.process || '-' }}</div>
                  <div>{{ log.remarks || '-' }}</div>
                </div>
              </div>
            </div>
          </template>
          <template v-else>
            <div class="text-sm text-gray-400">暂未生成真实订单</div>
          </template>
        </Card>
      </template>
    </Drawer>
  </Page>
</template>

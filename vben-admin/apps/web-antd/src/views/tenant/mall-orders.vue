<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import { Card, Table, Tag } from 'ant-design-vue';
import { getTenantMallOrdersApi, type MallPayOrder } from '#/api/tenant';

const loading = ref(false);
const list = ref<MallPayOrder[]>([]);
const total = ref(0);
const page = ref(1);

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

onMounted(load);
</script>

<template>
  <Page title="商城支付订单" content-class="p-4">
    <Card>
      <div class="mb-3 text-sm text-gray-500">共 {{ total }} 笔订单</div>
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
        <Table.Column title="商品ID" data-index="cid" :width="80" />
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
        <Table.Column title="备注" data-index="remark" />
        <Table.Column title="下单时间" data-index="addtime" :width="160" />
      </Table>
    </Card>
  </Page>
</template>

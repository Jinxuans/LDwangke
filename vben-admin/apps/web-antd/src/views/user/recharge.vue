<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Button, InputNumber, Table, Tag, Pagination, message, Alert, RadioGroup, RadioButton,
} from 'ant-design-vue';
import { DollarOutlined, ReloadOutlined, AlipayCircleOutlined, WechatOutlined, QqOutlined } from '@ant-design/icons-vue';
import {
  getPayChannelsApi, createPayOrderApi, getPayOrdersApi, checkPayStatusApi,
  type PayOrder, type PayChannel,
} from '#/api/user-center';

const quickAmounts = [50, 100, 200, 500, 1000];
const amount = ref(50);
const creating = ref(false);
const channels = ref<PayChannel[]>([]);
const selectedChannel = ref('');

// 充值记录
const loading = ref(false);
const orders = ref<PayOrder[]>([]);
const pagination = reactive({ page: 1, limit: 10, total: 0 });

async function loadChannels() {
  try {
    const raw = await getPayChannelsApi();
    channels.value = raw;
    if (!Array.isArray(channels.value)) channels.value = [];
    if (channels.value.length > 0 && !selectedChannel.value) {
      selectedChannel.value = channels.value[0]!.key;
    }
  } catch (e) {
    console.error(e);
  }
}

async function handlePay() {
  if (!amount.value || amount.value < 1) {
    message.warning('请输入有效金额');
    return;
  }
  if (!selectedChannel.value) {
    message.warning('请选择支付方式');
    return;
  }
  creating.value = true;
  try {
    const rawPay = await createPayOrderApi(amount.value, selectedChannel.value);
    const res = rawPay;
    if (res.pay_url) {
      message.success('订单已创建，正在跳转支付...');
      window.open(res.pay_url, '_blank');
    } else {
      message.success(`充值订单已创建，单号：${res.out_trade_no}`);
    }
    loadOrders(1);
  } catch (e: any) {
    message.error(e?.message || '创建订单失败');
  } finally {
    creating.value = false;
  }
}

async function loadOrders(page = 1) {
  loading.value = true;
  pagination.page = page;
  try {
    const rawOrders = await getPayOrdersApi(pagination.page, pagination.limit);
    const res = rawOrders;
    orders.value = res.list || [];
    pagination.total = res.pagination?.total || 0;
  } catch (e) {
    console.error(e);
  } finally {
    loading.value = false;
  }
}

function statusText(s: number) {
  if (s === 1) return '已支付';
  if (s === 2) return '已到账';
  return '待支付';
}
function statusColor(s: number) {
  if (s >= 1) return 'green';
  return 'orange';
}

async function handleCheckPay(outTradeNo: string) {
  try {
    const res = await checkPayStatusApi(outTradeNo);
    if (res.status === 1) {
      message.success(res.msg);
      loadOrders(pagination.page);
    } else {
      message.info(res.msg);
    }
  } catch (e: any) {
    message.error(e?.message || '查询失败');
  }
}

const channelIcons: Record<string, string> = {
  alipay: 'alipay',
  wxpay: 'wechat',
  qqpay: 'qq',
};

const columns = [
  { title: '订单号', dataIndex: 'out_trade_no', key: 'out_trade_no', ellipsis: true },
  { title: '金额', key: 'money', width: 100 },
  { title: '状态', key: 'status', width: 80 },
  { title: '时间', dataIndex: 'addtime', key: 'addtime', width: 160 },
  { title: '操作', key: 'action', width: 100 },
];

onMounted(() => {
  loadChannels();
  loadOrders(1);
});
</script>

<template>
  <Page title="充值" content-class="p-4">
    <Card title="在线充值" class="mb-4">
      <Alert message="选择金额和支付方式，点击立即充值后将跳转到支付页面。" type="info" show-icon class="mb-4" />

      <div class="mb-4">
        <div class="mb-2 font-medium">快捷金额</div>
        <div class="flex flex-wrap gap-3">
          <Button
            v-for="a in quickAmounts"
            :key="a"
            :type="amount === a ? 'primary' : 'default'"
            size="large"
            @click="amount = a"
          >
            ¥{{ a }}
          </Button>
        </div>
      </div>

      <div class="mb-4">
        <div class="mb-2 font-medium">自定义金额</div>
        <InputNumber
          v-model:value="amount"
          :min="1"
          :max="10000"
          :step="10"
          :precision="2"
          size="large"
          prefix="¥"
          style="max-width: 200px; min-width: 120px; width: 100%"
        />
      </div>

      <div class="mb-4" v-if="channels.length > 0">
        <div class="mb-2 font-medium">支付方式</div>
        <RadioGroup v-model:value="selectedChannel" button-style="solid" size="large">
          <RadioButton v-for="ch in channels" :key="ch.key" :value="ch.key">
            <AlipayCircleOutlined v-if="ch.key === 'alipay'" class="mr-1" />
            <WechatOutlined v-if="ch.key === 'wxpay'" class="mr-1" />
            <QqOutlined v-if="ch.key === 'qqpay'" class="mr-1" />
            {{ ch.label }}
          </RadioButton>
        </RadioGroup>
      </div>
      <Alert v-else type="warning" message="暂无可用支付渠道，请联系管理员配置。" show-icon class="mb-4" />

      <Button
        type="primary"
        size="large"
        :loading="creating"
        :disabled="!selectedChannel || !amount"
        @click="handlePay"
      >
        <template #icon><DollarOutlined /></template>
        立即充值 ¥{{ amount || 0 }}
      </Button>
    </Card>

    <Card title="充值记录">
      <div class="flex justify-end mb-3">
        <Button @click="loadOrders(pagination.page)">
          <template #icon><ReloadOutlined /></template>
        </Button>
      </div>
      <Table :columns="columns" :data-source="orders" :loading="loading" :pagination="false" row-key="oid" size="small" bordered :scroll="{ x: 600 }">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'money'">
            <span class="font-medium">¥{{ Number(record.money).toFixed(2) }}</span>
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="statusColor(record.status)">{{ statusText(record.status) }}</Tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <Button v-if="record.status === 0" type="link" size="small" @click="handleCheckPay(record.out_trade_no)">检测到账</Button>
          </template>
        </template>
      </Table>
      <div class="flex justify-center mt-4" v-if="pagination.total > pagination.limit">
        <Pagination v-model:current="pagination.page" :total="pagination.total" :page-size="pagination.limit" :show-total="(total: number) => `共 ${total} 条`" @change="(p: number) => loadOrders(p)" />
      </div>
    </Card>
  </Page>
</template>

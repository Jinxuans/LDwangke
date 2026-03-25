<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import { Page } from '@vben/common-ui';
import {
  Card,
  Button,
  InputNumber,
  Input,
  Table,
  Tag,
  Pagination,
  message,
  Alert,
  RadioGroup,
  RadioButton,
  Divider,
} from 'ant-design-vue';
import {
  DollarOutlined,
  ReloadOutlined,
  AlipayCircleOutlined,
  WechatOutlined,
  QqOutlined,
  GiftOutlined,
  FireOutlined,
} from '@ant-design/icons-vue';
import {
  getPayChannelsApi,
  createPayOrderApi,
  getPayOrdersApi,
  checkPayStatusApi,
  type PayOrder,
  type PayChannel,
} from '#/api/user-center';
import { useCardKeyApi } from '#/api/auxiliary';
import { getSiteConfigApi } from '#/api/admin';

const router = useRouter();

// ===== 充值赠送规则 =====
interface BonusRule {
  min: number;
  max: number;
  bonus_pct: number;
}
interface BonusActivity {
  enabled: boolean;
  weekdays: number[];
  rules: BonusRule[];
  hint?: string;
}
interface BonusConfig {
  enabled: boolean;
  rules: BonusRule[];
  activity: BonusActivity;
}

const bonusConfig = ref<BonusConfig | null>(null);

const isActivityDay = computed(() => {
  if (!bonusConfig.value?.activity?.enabled) return false;
  const weekday = new Date().getDay(); // 0=Sunday ... 6=Saturday
  return (bonusConfig.value.activity.weekdays ?? []).includes(weekday);
});

// 活动日使用独立规则，非活动日使用普通规则
const activeRules = computed(() => {
  if (!bonusConfig.value?.enabled) return [];
  if (isActivityDay.value && bonusConfig.value.activity.rules?.length) {
    return bonusConfig.value.activity.rules;
  }
  return bonusConfig.value.rules;
});

const bonusPreview = computed(() => {
  if (!activeRules.value.length || !amount.value) return null;
  const money = amount.value;
  let pct = 0;
  for (const r of activeRules.value) {
    if (money >= r.min && money < r.max) {
      pct = r.bonus_pct;
      break;
    }
  }
  if (pct <= 0) return null;
  const bonus = Math.round(money * pct) / 100;
  return { pct, bonus, total: money + bonus };
});

async function loadBonusConfig() {
  try {
    const cfg = await getSiteConfigApi();
    if (cfg?.recharge_bonus_rules) {
      bonusConfig.value = JSON.parse(cfg.recharge_bonus_rules);
    }
  } catch {
    /* ignore */
  }
}

// 卡密充值
const cardKeyCode = ref('');
const cardKeyLoading = ref(false);

async function handleCardKey() {
  const code = cardKeyCode.value.trim();
  if (!code) {
    message.warning('请输入卡密');
    return;
  }
  cardKeyLoading.value = true;
  try {
    const res = await useCardKeyApi(code);
    message.success(`充值成功，到账 ¥${res.money}`);
    cardKeyCode.value = '';
    loadOrders(1);
  } catch (e: any) {
    message.error(e?.message || '卡密使用失败');
  } finally {
    cardKeyLoading.value = false;
  }
}

const quickAmounts = [50, 100, 200, 500, 1000];
const amount = ref(50);
const creating = ref(false);
const channels = ref<PayChannel[]>([]);
const selectedChannel = ref('');
const hasOnlineRechargePermission = computed(() => channels.value.length > 0);

// 充值记录
const loading = ref(false);
const orders = ref<PayOrder[]>([]);
const pagination = reactive({ page: 1, limit: 10, total: 0 });

async function loadChannels() {
  try {
    const raw = await getPayChannelsApi();
    const list = Array.isArray(raw) ? raw : [];
    channels.value = list.map((ch: any) => ({
      key: ch.key || ch.type || '',
      label: ch.label || ch.name || '',
    }));
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

/**
 * 读取支付平台回跳参数。
 * 说明：
 * 1. 当前站点使用 hash 路由，支付平台会把参数拼在 # 号后；
 * 2. 少数情况下参数也可能出现在 search 中，这里一并兼容。
 */
function getPayReturnParams() {
  const fromSearch = new URLSearchParams(window.location.search);
  const hash = window.location.hash || '';
  const hashQueryIndex = hash.indexOf('?');
  const fromHash =
    hashQueryIndex >= 0
      ? new URLSearchParams(hash.slice(hashQueryIndex + 1))
      : new URLSearchParams();

  const read = (key: string) => fromHash.get(key) || fromSearch.get(key) || '';
  return {
    outTradeNo: read('out_trade_no'),
    tradeStatus: read('trade_status'),
    money: read('money'),
  };
}

/**
 * 清理支付平台回跳参数。
 * 处理完成后只保留正式充值页地址，避免地址栏长期挂着一串支付参数。
 */
async function clearPayReturnParams() {
  await router.replace('/user/recharge');
  await nextTick();
  window.history.replaceState(
    {},
    '',
    `${window.location.origin}${window.location.pathname}${window.location.hash}`,
  );
}

/**
 * 页面加载时自动检测支付结果。
 * 只有当 URL 中携带 out_trade_no 时才触发，避免影响普通进入充值页的用户。
 */
async function handlePayReturn() {
  const params = getPayReturnParams();
  if (!params.outTradeNo) {
    return;
  }

  try {
    const res = await checkPayStatusApi(params.outTradeNo);
    if (res.status === 1) {
      message.success(res.msg || '支付成功');
      loadOrders(1);
    } else {
      message.info(res.msg || '订单未支付，请稍后重试');
    }
  } catch (e: any) {
    message.error(e?.message || '查询支付状态失败');
  } finally {
    clearPayReturnParams();
  }
}

const channelIcons: Record<string, string> = {
  alipay: 'alipay',
  wxpay: 'wechat',
  qqpay: 'qq',
};

const columns = [
  {
    title: '订单号',
    dataIndex: 'out_trade_no',
    key: 'out_trade_no',
    ellipsis: true,
  },
  { title: '金额', key: 'money', width: 100 },
  { title: '状态', key: 'status', width: 80 },
  { title: '时间', dataIndex: 'addtime', key: 'addtime', width: 160 },
  { title: '操作', key: 'action', width: 100 },
];

onMounted(() => {
  loadChannels();
  loadOrders(1);
  loadBonusConfig();
  handlePayReturn();
});
</script>

<template>
  <Page title="充值" content-class="p-4">
    <!-- 活动日提示（不展示具体规则） -->
    <Alert
      v-if="hasOnlineRechargePermission && isActivityDay"
      type="success"
      class="mb-4"
      show-icon
    >
      <template #icon><FireOutlined class="text-red-500" /></template>
      <template #message
        ><span class="text-base font-bold text-red-500">{{
          bonusConfig?.activity?.hint || '今日爆率很高，充值更划算！'
        }}</span></template
      >
    </Alert>

    <!-- 充值赠送规则展示（非活动日显示普通规则） -->
    <Card
      v-if="
        hasOnlineRechargePermission &&
        bonusConfig?.enabled &&
        bonusConfig.rules.length &&
        !isActivityDay
      "
      class="mb-4"
    >
      <template #title>
        <span><GiftOutlined class="mr-1 text-orange-500" />充值赠送活动</span>
      </template>
      <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
        <div
          v-for="(rule, idx) in bonusConfig.rules"
          :key="idx"
          class="bonus-card"
        >
          <div class="text-xs text-gray-400">
            充值 ¥{{ rule.min }} ~ ¥{{ rule.max }}
          </div>
          <div class="my-1 text-lg font-bold text-orange-500">
            赠送 {{ rule.bonus_pct }}%
          </div>
        </div>
      </div>
    </Card>

    <Card v-if="hasOnlineRechargePermission" title="在线充值" class="mb-4">
      <Alert
        message="选择金额和支付方式，点击立即充值后将跳转到支付页面。"
        type="info"
        show-icon
        class="mb-4"
      />

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

      <!-- 赠送预览 -->
      <div
        v-if="bonusPreview"
        class="mb-4 rounded-lg border border-orange-200 bg-orange-50 p-3 dark:border-orange-800 dark:bg-orange-900/20"
      >
        <div class="flex flex-wrap items-center gap-2">
          <GiftOutlined class="text-orange-500" />
          <span class="text-sm"
            >充值 <b>¥{{ amount }}</b></span
          >
          <span class="text-sm"
            >→ 赠送
            <b class="text-orange-500">¥{{ bonusPreview.bonus.toFixed(2) }}</b
            >（{{ bonusPreview.pct }}%）</span
          >
          <span class="text-sm"
            >→ 实际到账
            <b class="text-base text-green-600"
              >¥{{ bonusPreview.total.toFixed(2) }}</b
            ></span
          >
        </div>
      </div>

      <div class="mb-4">
        <div class="mb-2 font-medium">支付方式</div>
        <div class="flex flex-wrap gap-3">
          <div
            v-for="ch in channels"
            :key="ch.key"
            class="pay-channel-btn"
            :class="{ 'pay-channel-active': selectedChannel === ch.key }"
            @click="selectedChannel = ch.key"
          >
            <AlipayCircleOutlined
              v-if="ch.key === 'alipay'"
              class="mr-2 text-lg text-blue-500"
            />
            <WechatOutlined
              v-if="ch.key === 'wxpay'"
              class="mr-2 text-lg text-green-500"
            />
            <QqOutlined
              v-if="ch.key === 'qqpay'"
              class="mr-2 text-lg text-blue-400"
            />
            <span class="text-sm font-medium">{{ ch.label }}</span>
          </div>
        </div>
      </div>

      <Button
        type="primary"
        size="large"
        :loading="creating"
        :disabled="!selectedChannel || !amount"
        @click="handlePay"
      >
        <template #icon><DollarOutlined /></template>
        立即充值 ¥{{ amount || 0 }}
        <span v-if="bonusPreview" class="ml-1 text-xs opacity-80"
          >（到账 ¥{{ bonusPreview.total.toFixed(2) }}）</span
        >
      </Button>
    </Card>

    <Alert
      v-else
      type="warning"
      message="当前账号没有在线充值权限，请联系管理员或上级处理。"
      show-icon
      class="mb-4"
    />

    <Card title="卡密充值" class="mb-4">
      <div class="flex flex-col gap-3 sm:flex-row">
        <Input
          v-model:value="cardKeyCode"
          placeholder="请输入卡密"
          allow-clear
          size="large"
          class="flex-1"
          @press-enter="handleCardKey"
        />
        <Button
          type="primary"
          size="large"
          :loading="cardKeyLoading"
          @click="handleCardKey"
          class="w-full sm:w-auto"
        >
          兑换充值
        </Button>
      </div>
    </Card>

    <Card title="充值记录">
      <div class="mb-3 flex justify-end">
        <Button @click="loadOrders(pagination.page)">
          <template #icon><ReloadOutlined /></template>
        </Button>
      </div>
      <Table
        :columns="columns"
        :data-source="orders"
        :loading="loading"
        :pagination="false"
        row-key="oid"
        size="small"
        bordered
        :scroll="{ x: 600 }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'money'">
            <span class="font-medium"
              >¥{{ Number(record.money).toFixed(2) }}</span
            >
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="statusColor(record.status)">{{
              statusText(record.status)
            }}</Tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <Button
              v-if="record.status === 0"
              type="link"
              size="small"
              @click="handleCheckPay(record.out_trade_no)"
              >检测到账</Button
            >
          </template>
        </template>
      </Table>
      <div
        class="mt-4 flex justify-center"
        v-if="pagination.total > pagination.limit"
      >
        <Pagination
          v-model:current="pagination.page"
          :total="pagination.total"
          :page-size="pagination.limit"
          :show-total="(total: number) => `共 ${total} 条`"
          @change="(p: number) => loadOrders(p)"
        />
      </div>
    </Card>
  </Page>
</template>

<style scoped>
.bonus-card {
  padding: 12px 16px;
  border-radius: 10px;
  background: linear-gradient(135deg, #fff7ed, #ffedd5);
  border: 1px solid #fed7aa;
  text-align: center;
  transition: transform 0.15s;
}
.bonus-card:hover {
  transform: translateY(-2px);
}
html.dark .bonus-card {
  background: linear-gradient(135deg, #431407, #7c2d12);
  border-color: #9a3412;
}
.pay-channel-btn {
  display: inline-flex;
  align-items: center;
  padding: 10px 24px;
  border: 2px solid #d9d9d9;
  border-radius: 8px;
  cursor: pointer;
  font-size: 15px;
  font-weight: 500;
  transition: all 0.2s;
  user-select: none;
}
.pay-channel-btn:hover {
  border-color: #1677ff;
  color: #1677ff;
}
.pay-channel-active {
  border-color: #1677ff;
  background: #e6f4ff;
  color: #1677ff;
}
html.dark .pay-channel-btn {
  border-color: #424242;
}
html.dark .pay-channel-btn:hover {
  border-color: #1677ff;
}
html.dark .pay-channel-active {
  border-color: #1677ff;
  background: rgba(22, 119, 255, 0.15);
}
</style>

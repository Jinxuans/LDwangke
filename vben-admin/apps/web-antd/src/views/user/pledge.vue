<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Button, Tag, Empty, Popconfirm, message, Progress, Spin, Row, Col, Alert,
} from 'ant-design-vue';
import {
  ClockCircleOutlined, DollarOutlined, TagOutlined, CalendarOutlined,
  CheckCircleOutlined, WarningOutlined,
} from '@ant-design/icons-vue';
import {
  getUserPledgeConfigsApi, createPledgeApi, cancelPledgeApi, getMyPledgesApi,
  type PledgeConfig, type PledgeRecord,
} from '#/api/auxiliary';

const configLoading = ref(false);
const configs = ref<PledgeConfig[]>([]);
const pledgeLoading = ref(false);
const myPledges = ref<PledgeRecord[]>([]);

function calcExpiry(addtime: string, days: number) {
  const start = new Date(addtime);
  const end = new Date(start.getTime() + days * 86400000);
  return end;
}

function formatDate(d: Date) {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  const h = String(d.getHours()).padStart(2, '0');
  const min = String(d.getMinutes()).padStart(2, '0');
  return `${y}-${m}-${day} ${h}:${min}`;
}

function remainDays(addtime: string, days: number) {
  const end = calcExpiry(addtime, days);
  const diff = end.getTime() - Date.now();
  return Math.max(0, Math.ceil(diff / 86400000));
}

function progressPercent(addtime: string, days: number) {
  const start = new Date(addtime).getTime();
  const total = days * 86400000;
  const elapsed = Date.now() - start;
  return Math.min(100, Math.max(0, Math.round((elapsed / total) * 100)));
}

function progressStatus(addtime: string, days: number): 'success' | 'active' | 'exception' {
  const remain = remainDays(addtime, days);
  if (remain <= 0) return 'success';
  if (remain <= 3) return 'exception';
  return 'active';
}

async function loadConfigs() {
  configLoading.value = true;
  try {
    const res = await getUserPledgeConfigsApi();
    configs.value = Array.isArray(res) ? res : [];
  } catch (e) { console.error(e); }
  finally { configLoading.value = false; }
}

async function loadMyPledges() {
  pledgeLoading.value = true;
  try {
    const res = await getMyPledgesApi();
    myPledges.value = Array.isArray(res) ? res : [];
  } catch (e) { console.error(e); }
  finally { pledgeLoading.value = false; }
}

async function handlePledge(configId: number) {
  try {
    await createPledgeApi(configId);
    message.success('质押成功');
    loadConfigs();
    loadMyPledges();
  } catch (e: any) { message.error(e?.message || '质押失败'); }
}

async function handleCancel(id: number) {
  try {
    await cancelPledgeApi(id);
    message.success('取消质押成功，余额已退还');
    loadMyPledges();
  } catch (e: any) { message.error(e?.message || '取消失败'); }
}

onMounted(() => { loadConfigs(); loadMyPledges(); });
</script>

<template>
  <Page title="质押折扣" content-class="p-4">
    <Spin :spinning="pledgeLoading">
      <!-- 我的质押 -->
      <Card title="我的质押" class="mb-4">
        <template v-if="myPledges.length === 0 && !pledgeLoading">
          <Empty description="暂无生效中的质押，请在下方选择方案开始质押" />
        </template>

        <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4" v-if="myPledges.length > 0">
          <Card
            v-for="p in myPledges" :key="p.id" size="small"
            class="pledge-card"
            :class="remainDays(p.addtime, p.days) <= 3 ? 'pledge-expiring' : 'pledge-active'"
          >
            <!-- 头部：分类名 + 状态 -->
            <div class="flex items-center justify-between mb-3">
              <div class="text-base font-bold" :class="remainDays(p.addtime, p.days) <= 3 ? 'text-orange-600' : 'text-blue-600'">
                {{ p.category_name }}
              </div>
              <Tag v-if="remainDays(p.addtime, p.days) <= 0" color="red">已到期</Tag>
              <Tag v-else-if="remainDays(p.addtime, p.days) <= 3" color="orange">即将到期</Tag>
              <Tag v-else color="green">生效中</Tag>
            </div>

            <!-- 进度条 -->
            <Progress
              :percent="progressPercent(p.addtime, p.days)"
              :status="progressStatus(p.addtime, p.days)"
              :show-info="false"
              :stroke-width="6"
              class="mb-3"
            />

            <!-- 详细信息 -->
            <div class="space-y-2 text-sm">
              <div class="flex items-center justify-between">
                <span class="text-gray-500"><DollarOutlined class="mr-1" />质押金额</span>
                <span class="font-semibold">¥{{ p.amount }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-gray-500"><TagOutlined class="mr-1" />折扣率</span>
                <Tag color="blue">{{ (p.discount_rate * 100).toFixed(0) }}% 折</Tag>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-gray-500"><CalendarOutlined class="mr-1" />质押时间</span>
                <span>{{ p.addtime }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-gray-500"><ClockCircleOutlined class="mr-1" />到期时间</span>
                <span class="font-medium" :class="remainDays(p.addtime, p.days) <= 3 ? 'text-orange-600' : ''">
                  {{ formatDate(calcExpiry(p.addtime, p.days)) }}
                </span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-gray-500"><ClockCircleOutlined class="mr-1" />剩余天数</span>
                <span class="font-bold text-lg" :class="remainDays(p.addtime, p.days) <= 3 ? 'text-red-500' : 'text-green-600'">
                  {{ remainDays(p.addtime, p.days) }} 天
                </span>
              </div>
            </div>

            <!-- 操作 -->
            <div class="mt-4">
              <Popconfirm title="提前取消会扣除部分质押金，确定取消？" @confirm="handleCancel(p.id)">
                <Button danger block size="small">取消质押</Button>
              </Popconfirm>
            </div>
          </Card>
        </div>
      </Card>
    </Spin>

    <!-- 可用质押方案 -->
    <Spin :spinning="configLoading">
      <Card title="可用质押方案">
        <template v-if="configs.length === 0 && !configLoading">
          <Empty description="暂无可用的质押方案" />
        </template>

        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4" v-if="configs.length > 0">
          <Card v-for="cfg in configs" :key="cfg.id" size="small" hoverable class="config-card">
            <div class="text-center mb-3">
              <div class="text-lg font-bold text-blue-600">{{ cfg.category_name }}</div>
              <div class="text-xs text-gray-400 mt-1">质押 {{ cfg.days }} 天享折扣</div>
            </div>
            <div class="space-y-2 text-sm">
              <div class="flex justify-between">
                <span class="text-gray-500">质押金额</span>
                <span class="font-semibold text-blue-600">¥{{ cfg.amount }}</span>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-gray-500">折扣率</span>
                <Tag color="blue">{{ (cfg.discount_rate * 100).toFixed(0) }}%</Tag>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500">质押天数</span>
                <span>{{ cfg.days }} 天</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500">提前取消扣费</span>
                <span class="text-orange-500">{{ (cfg.cancel_fee * 100).toFixed(0) }}%</span>
              </div>
            </div>
            <Popconfirm :title="`确定质押 ¥${cfg.amount}？将从余额中扣除，${cfg.days}天后自动退还`" @confirm="handlePledge(cfg.id)">
              <Button type="primary" block class="mt-4">立即质押</Button>
            </Popconfirm>
          </Card>
        </div>

        <Alert v-if="configs.length > 0" class="mt-4" type="info" show-icon
          message="质押说明：质押后对应分类的课程将享受折扣价格，到期后质押金自动退还。提前取消将扣除一定比例手续费。"
        />
      </Card>
    </Spin>
  </Page>
</template>

<style scoped>
.pledge-card {
  border-radius: 8px;
  transition: box-shadow 0.2s;
}
.pledge-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}
.pledge-active {
  border-left: 3px solid #52c41a;
}
.pledge-expiring {
  border-left: 3px solid #fa8c16;
}
.config-card {
  border-radius: 8px;
  border: 1px solid #e8e8e8;
  transition: all 0.2s;
}
.config-card:hover {
  border-color: #1677ff;
  box-shadow: 0 4px 12px rgba(22, 119, 255, 0.1);
}
</style>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal, message,
  Select, SelectOption, Form, FormItem, DatePicker, Alert,
} from 'ant-design-vue';
import type { SDXYOrder } from '#/api/sdxy';
import {
  sdxyOrderListApi,
  sdxyGetPriceApi,
  sdxyAddOrderApi,
  sdxyDeleteOrderApi,
} from '#/api/sdxy';

// ---------- 状态 ----------
const loading = ref(false);
const orders = ref<SDXYOrder[]>([]);
const total = ref(0);
const pagination = reactive({ page: 1, limit: 20 });
const search = reactive({ type: '1', keyword: '', status: '' });
const pricePerKM = ref(0);

// 添加弹窗
const addVisible = ref(false);
const addLoading = ref(false);
const addForm = reactive({
  user: '',
  pass: '',
  school: '',
  distance: 2,
  day: 7,
  start_date: '',
  start_hour: 6,
  start_minute: 0,
  end_hour: 8,
  end_minute: 0,
  run_week: ['1', '2', '3', '4', '5'] as string[],
});

// 选项
const weekOptions = [
  { label: '周一', value: '1' },
  { label: '周二', value: '2' },
  { label: '周三', value: '3' },
  { label: '周四', value: '4' },
  { label: '周五', value: '5' },
  { label: '周六', value: '6' },
  { label: '周日', value: '7' },
];
const statusOptions = [
  { label: '所有订单', value: '' },
  { label: '等待处理', value: '1' },
  { label: '处理成功', value: '2' },
  { label: '退款成功', value: '3' },
];
const searchTypeOptions = [
  { label: '订单ID', value: '1' },
  { label: '下单账号', value: '2' },
  { label: '下单密码', value: '3' },
  { label: '用户UID', value: '4' },
];

// 价格计算
const addTotal = ref(0);
function calcPrice() {
  addTotal.value = Math.round(pricePerKM.value * addForm.distance * addForm.day * 100) / 100;
}

// ---------- 列表 ----------
async function fetchOrders() {
  loading.value = true;
  try {
    const res: any = await sdxyOrderListApi({
      page: pagination.page,
      limit: pagination.limit,
      searchType: search.keyword ? search.type : undefined,
      keyword: search.keyword || undefined,
      status: search.status || undefined,
    });
    orders.value = res.list || [];
    total.value = res.total || 0;
  } finally {
    loading.value = false;
  }
}

function getStatusText(s: number) {
  const map: Record<number, string> = { 1: '等待处理', 2: '处理成功', 3: '退款成功' };
  return map[s] || `${s}`;
}
function getStatusColor(s: number) {
  const map: Record<number, string> = { 1: 'warning', 2: 'success', 3: 'default' };
  return map[s] || 'error';
}
function getWeekText(week: string) {
  if (!week) return '';
  const map: Record<string, string> = { '1': '一', '2': '二', '3': '三', '4': '四', '5': '五', '6': '六', '7': '日' };
  return week.split(',').map((w) => `周${map[w] || w}`).join(' ');
}

// ---------- 下单 ----------
async function handleAdd() {
  if (!addForm.user || !addForm.pass || addForm.day < 1) {
    message.warning('请填写完整信息');
    return;
  }
  addLoading.value = true;
  try {
    await sdxyAddOrderApi({
      user: addForm.user,
      pass: addForm.pass,
      school: addForm.school,
      distance: String(addForm.distance),
      day: String(addForm.day),
      start_date: addForm.start_date,
      start_hour: String(addForm.start_hour),
      start_minute: String(addForm.start_minute),
      end_hour: String(addForm.end_hour),
      end_minute: String(addForm.end_minute),
      run_week: addForm.run_week.join(','),
    });
    message.success('下单成功');
    addVisible.value = false;
    fetchOrders();
  } catch (e: any) {
    message.error(e?.message || '下单失败');
  } finally {
    addLoading.value = false;
  }
}

// ---------- 退款 ----------
function handleRefund(record: SDXYOrder) {
  Modal.confirm({
    title: '确认退款',
    content: `确定要退款订单 #${record.id}（账号：${record.user}）吗？`,
    onOk: async () => {
      try {
        await sdxyDeleteOrderApi(record.id);
        message.success('退款成功');
        fetchOrders();
      } catch (e: any) {
        message.error(e?.message || '退款失败');
      }
    },
  });
}

// ---------- 表格列 ----------
const columns = [
  { title: 'ID', dataIndex: 'id', width: 70 },
  { title: 'UID', dataIndex: 'uid', width: 60 },
  { title: '账号', key: 'account', width: 130 },
  { title: '学校', dataIndex: 'school', width: 100 },
  { title: '公里数', dataIndex: 'distance', width: 80 },
  { title: '天数', dataIndex: 'day', width: 60 },
  { title: '开始日期', dataIndex: 'start_date', width: 110 },
  { title: '时间', key: 'time', width: 100 },
  { title: '跑步周期', key: 'run_week', width: 140 },
  { title: '状态', key: 'status', width: 90 },
  { title: '备注', dataIndex: 'remarks', ellipsis: true },
  { title: '下单时间', dataIndex: 'addtime', width: 150 },
  { title: '操作', key: 'action', width: 80, fixed: 'right' as const },
];

onMounted(async () => {
  try {
    const res: any = await sdxyGetPriceApi();
    pricePerKM.value = res.price || 0;
  } catch {}
  calcPrice();
  fetchOrders();
});
</script>

<template>
  <Page title="闪电运动" description="管理闪电运动跑步订单">
    <Card class="mb-4" :bordered="false">
      <Alert message="下单前请确认账号密码正确，跑步期间切勿登录账号！" type="warning" show-icon class="mb-4" />
      <div class="flex flex-wrap items-center gap-3">
        <Space>
          <Button type="primary" @click="addVisible = true">添加订单</Button>
          <Tag color="orange">{{ pricePerKM }} 元/公里</Tag>
        </Space>
        <div class="flex-1" />
        <Space wrap>
          <Select v-model:value="search.status" style="width: 120px" @change="fetchOrders">
            <SelectOption v-for="o in statusOptions" :key="o.value" :value="o.value">{{ o.label }}</SelectOption>
          </Select>
          <Input.Group compact>
            <Select v-model:value="search.type" style="width: 100px">
              <SelectOption v-for="o in searchTypeOptions" :key="o.value" :value="o.value">{{ o.label }}</SelectOption>
            </Select>
            <Input.Search v-model:value="search.keyword" placeholder="关键词" style="width: 180px" @search="fetchOrders" allow-clear />
          </Input.Group>
        </Space>
      </div>
    </Card>

    <Card :bordered="false">
      <Table :columns="columns" :data-source="orders" :loading="loading" :pagination="{
        current: pagination.page, pageSize: pagination.limit, total,
        showSizeChanger: true, pageSizeOptions: ['20', '50', '100'],
        showTotal: (t: number) => `共 ${t} 条`,
        onChange: (p: number, s: number) => { pagination.page = p; pagination.limit = s; fetchOrders(); },
        onShowSizeChange: (_: number, s: number) => { pagination.limit = s; pagination.page = 1; fetchOrders(); },
      }" row-key="id" :scroll="{ x: 1200 }" size="small">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'account'">
            <div class="leading-tight">{{ record.user }}</div>
            <div class="text-gray-400 text-xs">{{ record.pass }}</div>
          </template>
          <template v-if="column.key === 'time'">
            {{ record.start_hour }}:{{ record.start_minute }} - {{ record.end_hour }}:{{ record.end_minute }}
          </template>
          <template v-if="column.key === 'run_week'">{{ getWeekText(record.run_week) }}</template>
          <template v-if="column.key === 'status'">
            <Tag :color="getStatusColor(record.status)">{{ getStatusText(record.status) }}</Tag>
          </template>
          <template v-if="column.key === 'action'">
            <Button size="small" danger @click="handleRefund(record)" :disabled="record.status === 3">退款</Button>
          </template>
        </template>
      </Table>
    </Card>

    <!-- 添加弹窗 -->
    <Modal v-model:open="addVisible" title="添加订单" width="550px" :confirm-loading="addLoading" @ok="handleAdd">
      <Form layout="horizontal" :label-col="{ span: 5 }">
        <FormItem label="手机号"><Input v-model:value="addForm.user" placeholder="请输入账号" /></FormItem>
        <FormItem label="用户密码"><Input v-model:value="addForm.pass" placeholder="请输入密码" /></FormItem>
        <FormItem label="跑区简称"><Input v-model:value="addForm.school" placeholder="例如：东校区" /></FormItem>
        <FormItem label="日公里数">
          <InputNumber v-model:value="addForm.distance" :min="1" :max="100" :step="0.1" @change="calcPrice" />
        </FormItem>
        <FormItem label="跑步天数">
          <InputNumber v-model:value="addForm.day" :min="1" :max="365" @change="calcPrice" />
        </FormItem>
        <FormItem label="开始日期">
          <DatePicker v-model:value="addForm.start_date" value-format="YYYY-MM-DD" />
        </FormItem>
        <FormItem label="开始小时"><InputNumber v-model:value="addForm.start_hour" :min="0" :max="23" /></FormItem>
        <FormItem label="开始分钟"><InputNumber v-model:value="addForm.start_minute" :min="0" :max="59" /></FormItem>
        <FormItem label="结束小时"><InputNumber v-model:value="addForm.end_hour" :min="0" :max="23" /></FormItem>
        <FormItem label="结束分钟"><InputNumber v-model:value="addForm.end_minute" :min="0" :max="59" /></FormItem>
        <FormItem label="跑步周期">
          <Select v-model:value="addForm.run_week" mode="multiple" placeholder="请选择跑步周期">
            <SelectOption v-for="w in weekOptions" :key="w.value" :value="w.value">{{ w.label }}</SelectOption>
          </Select>
        </FormItem>
        <FormItem label="价格总计">
          <span style="color: red; font-weight: 800; font-size: 18px">￥{{ addTotal.toFixed(2) }}</span>
        </FormItem>
      </Form>
    </Modal>

  </Page>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal, message,
  Select, SelectOption, Form, FormItem, Tabs, TabPane, Checkbox,
} from 'ant-design-vue';
import type { YongyeOrder, YongyeStudent } from '#/api/plugins/yongye';
import {
  getYongyeOrders,
  getYongyeStudents,
  getYongyeSchools,
  addYongyeOrder,
  refundYongyeOrder,
  refundYongyeStudent,
  toggleYongyePolling,
} from '#/api/plugins/yongye';

// ---------- 状态 ----------
const activeTab = ref('orders');
const loading = ref(false);
const orders = ref<YongyeOrder[]>([]);
const students = ref<YongyeStudent[]>([]);
const total = ref(0);
const pagination = reactive({ page: 1, limit: 20 });
const search = reactive({ keyword: '', status: '' });

// 学校列表
const schoolList = ref<any[]>([]);
const schoolLoading = ref(false);

// 添加弹窗
const addVisible = ref(false);
const addLoading = ref(false);
const addForm = reactive({
  school: '自动识别',
  user: '',
  pass: '',
  zkm: 2,
  type: 0,
  ks_h: 9,
  ks_m: 0,
  js_h: 21,
  js_m: 0,
  weeks: '12345',
  isPolling: 0,
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
const runTypeOptions = [
  { label: '正常跑(课外)', value: 0 },
  { label: '晨跑', value: 1 },
];
const statusOptions = [
  { label: '全部', value: '' },
  { label: '未提交', value: '0' },
  { label: '已提交', value: '1' },
  { label: '请求失败', value: '2' },
  { label: '已关闭', value: '3' },
  { label: '轮询中', value: '5' },
];
const stuStatusOptions: Record<number, string> = {
  0: '正常',
  1: '暂停',
  2: '完成',
  3: '退单',
};

// ---------- 列表 ----------
async function fetchOrders() {
  loading.value = true;
  try {
    const res: any = await getYongyeOrders({
      page: pagination.page,
      limit: pagination.limit,
      keyword: search.keyword || undefined,
      status: search.status || undefined,
    });
    orders.value = res.list || [];
    total.value = res.total || 0;
  } finally {
    loading.value = false;
  }
}

async function fetchStudents() {
  loading.value = true;
  try {
    const res: any = await getYongyeStudents({ keyword: search.keyword || undefined });
    students.value = res || [];
  } finally {
    loading.value = false;
  }
}

async function fetchSchools() {
  schoolLoading.value = true;
  try {
    const res: any = await getYongyeSchools();
    if (res?.data) {
      schoolList.value = res.data;
    }
  } catch {
    // ignore
  } finally {
    schoolLoading.value = false;
  }
}

function getStatusText(s: number) {
  const map: Record<number, string> = { 0: '未提交', 1: '已提交', 2: '请求失败', 3: '已关闭', 5: '轮询中' };
  return map[s] || `${s}`;
}
function getStatusColor(s: number) {
  const map: Record<number, string> = { 0: 'default', 1: 'processing', 2: 'error', 3: 'warning', 5: 'success' };
  return map[s] || 'default';
}
function getRunTypeText(t: number) {
  return t === 1 ? '晨跑' : '课外跑';
}
function getWeeksText(w: string) {
  if (!w) return '';
  const map: Record<string, string> = { '1': '一', '2': '二', '3': '三', '4': '四', '5': '五', '6': '六', '7': '日' };
  return w.split('').map((c) => `周${map[c] || c}`).join(' ');
}

// ---------- 下单 ----------
function openAddModal() {
  addVisible.value = true;
  if (schoolList.value.length === 0) {
    fetchSchools();
  }
}

// weeks 多选辅助
const selectedWeeks = ref<string[]>(['1', '2', '3', '4', '5']);
function syncWeeks() {
  selectedWeeks.value.sort();
  addForm.weeks = selectedWeeks.value.join('');
}

async function handleAdd() {
  if (!addForm.user || !addForm.pass || addForm.zkm <= 0) {
    message.warning('请填写完整信息');
    return;
  }
  syncWeeks();
  addLoading.value = true;
  try {
    await addYongyeOrder({ ...addForm });
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
function handleRefund(record: YongyeOrder) {
  Modal.confirm({
    title: '确认退款',
    content: `确定要退款订单 #${record.id}（账号：${record.user}，预扣：${record.yfees}元）吗？`,
    onOk: async () => {
      try {
        await refundYongyeOrder(record.id);
        message.success('退款成功');
        fetchOrders();
      } catch (e: any) {
        message.error(e?.message || '退款失败');
      }
    },
  });
}

// ---------- 退单（学生） ----------
function handleRefundStudent(record: YongyeStudent) {
  Modal.confirm({
    title: '确认退单',
    content: `确定要退单学生 ${record.user} 吗？此操作将通知上游取消。`,
    onOk: async () => {
      try {
        await refundYongyeStudent(record.user, record.type);
        message.success('退单请求已发送');
        fetchStudents();
      } catch (e: any) {
        message.error(e?.message || '退单失败');
      }
    },
  });
}

// ---------- 轮询开关 ----------
async function handleTogglePolling(record: YongyeOrder) {
  try {
    await toggleYongyePolling(record.id);
    message.success(record.pol === 0 ? '已开启轮询' : '已关闭轮询');
    fetchOrders();
  } catch (e: any) {
    message.error(e?.message || '操作失败');
  }
}

// ---------- 表格列 ----------
const orderColumns = [
  { title: 'ID', dataIndex: 'id', width: 70 },
  { title: 'UID', dataIndex: 'uid', width: 60 },
  { title: '学校', dataIndex: 'school', width: 100, ellipsis: true },
  { title: '账号', key: 'account', width: 130 },
  { title: '公里', dataIndex: 'zkm', width: 60 },
  { title: '类型', key: 'run_type', width: 80 },
  { title: '时间', key: 'time', width: 110 },
  { title: '周天', key: 'weeks', width: 140 },
  { title: '预扣', dataIndex: 'yfees', width: 70 },
  { title: '状态', key: 'status', width: 90 },
  { title: '轮询', key: 'pol', width: 70 },
  { title: '日志', dataIndex: 'tktext', ellipsis: true },
  { title: '时间', dataIndex: 'addtime', width: 150 },
  { title: '操作', key: 'action', width: 160, fixed: 'right' as const },
];

const studentColumns = [
  { title: 'ID', dataIndex: 'id', width: 60 },
  { title: 'UID', dataIndex: 'uid', width: 60 },
  { title: '学号', dataIndex: 'user', width: 130 },
  { title: '密码', dataIndex: 'pass', width: 100 },
  { title: '类型', key: 'type', width: 80 },
  { title: '公里', dataIndex: 'zkm', width: 60 },
  { title: '周天', key: 'weeks', width: 140 },
  { title: '状态', key: 'status', width: 80 },
  { title: '退单公里', dataIndex: 'tdkm', width: 80 },
  { title: '退单金额', dataIndex: 'tdmoney', width: 80 },
  { title: '最后更新', dataIndex: 'last_time', width: 150 },
  { title: '操作', key: 'action', width: 100, fixed: 'right' as const },
];

function onTabChange(key: string) {
  activeTab.value = key;
  if (key === 'orders') fetchOrders();
  else fetchStudents();
}

onMounted(() => fetchOrders());
</script>

<template>
  <Page title="永夜运动" description="永夜运动世界订单管理">
    <Card class="mb-4" :bordered="false">
      <div class="flex flex-wrap items-center gap-3">
        <Space>
          <Button type="primary" @click="openAddModal">添加订单</Button>
        </Space>
        <div class="flex-1" />
        <Space wrap>
          <Select v-model:value="search.status" style="width: 110px" @change="fetchOrders" v-if="activeTab === 'orders'">
            <SelectOption v-for="o in statusOptions" :key="o.value" :value="o.value">{{ o.label }}</SelectOption>
          </Select>
          <Input.Search v-model:value="search.keyword" placeholder="搜索账号/ID" style="width: 200px"
            @search="activeTab === 'orders' ? fetchOrders() : fetchStudents()" allow-clear />
        </Space>
      </div>
    </Card>

    <Card :bordered="false">
      <Tabs :active-key="activeTab" @change="onTabChange">
        <TabPane key="orders" tab="订单列表">
          <Table :columns="orderColumns" :data-source="orders" :loading="loading" :pagination="{
            current: pagination.page, pageSize: pagination.limit, total,
            showSizeChanger: true, pageSizeOptions: ['20', '50', '100'],
            showTotal: (t: number) => `共 ${t} 条`,
            onChange: (p: number, s: number) => { pagination.page = p; pagination.limit = s; fetchOrders(); },
            onShowSizeChange: (_: number, s: number) => { pagination.limit = s; pagination.page = 1; fetchOrders(); },
          }" row-key="id" :scroll="{ x: 1500 }" size="small">
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'account'">
                <div class="leading-tight">{{ record.user }}</div>
                <div class="text-gray-400 text-xs">{{ record.pass }}</div>
              </template>
              <template v-if="column.key === 'run_type'">{{ getRunTypeText(record.type) }}</template>
              <template v-if="column.key === 'time'">
                {{ String(record.ks_h).padStart(2, '0') }}:{{ String(record.ks_m).padStart(2, '0') }} - {{ String(record.js_h).padStart(2, '0') }}:{{ String(record.js_m).padStart(2, '0') }}
              </template>
              <template v-if="column.key === 'weeks'">{{ getWeeksText(record.weeks) }}</template>
              <template v-if="column.key === 'status'">
                <Tag :color="getStatusColor(record.dockstatus)">{{ getStatusText(record.dockstatus) }}</Tag>
              </template>
              <template v-if="column.key === 'pol'">
                <Tag :color="record.pol === 1 ? 'green' : 'default'">{{ record.pol === 1 ? '开' : '关' }}</Tag>
              </template>
              <template v-if="column.key === 'action'">
                <Space>
                  <Button size="small" @click="handleTogglePolling(record)">{{ record.pol === 0 ? '开轮询' : '关轮询' }}</Button>
                  <Button size="small" danger @click="handleRefund(record)" :disabled="record.dockstatus === 3">退款</Button>
                </Space>
              </template>
            </template>
          </Table>
        </TabPane>

        <TabPane key="students" tab="学生列表">
          <Table :columns="studentColumns" :data-source="students" :loading="loading"
            :pagination="false" row-key="id" :scroll="{ x: 1200 }" size="small">
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'type'">{{ getRunTypeText(record.type) }}</template>
              <template v-if="column.key === 'weeks'">{{ getWeeksText(record.weeks) }}</template>
              <template v-if="column.key === 'status'">
                <Tag :color="record.status === 3 ? 'error' : record.status === 2 ? 'success' : 'processing'">
                  {{ stuStatusOptions[record.status] || record.status }}
                </Tag>
              </template>
              <template v-if="column.key === 'action'">
                <Button size="small" danger @click="handleRefundStudent(record)" :disabled="record.status === 3">退单</Button>
              </template>
            </template>
          </Table>
        </TabPane>
      </Tabs>
    </Card>

    <!-- 添加弹窗 -->
    <Modal v-model:open="addVisible" title="添加订单" width="550px" :confirm-loading="addLoading" @ok="handleAdd">
      <Form layout="horizontal" :label-col="{ span: 5 }">
        <FormItem label="跑步类型">
          <Select v-model:value="addForm.type">
            <SelectOption v-for="o in runTypeOptions" :key="o.value" :value="o.value">{{ o.label }}</SelectOption>
          </Select>
        </FormItem>
        <FormItem label="学校">
          <Select v-model:value="addForm.school" :loading="schoolLoading" show-search
            :filter-option="(input: string, option: any) => (option?.label || option?.children || '').toLowerCase().includes(input.toLowerCase())">
            <SelectOption value="自动识别">自动识别</SelectOption>
            <SelectOption v-for="s in schoolList" :key="s.name" :value="s.name">{{ s.name }}</SelectOption>
          </Select>
        </FormItem>
        <FormItem label="学号"><Input v-model:value="addForm.user" placeholder="请输入学号" /></FormItem>
        <FormItem label="密码"><Input v-model:value="addForm.pass" placeholder="请输入密码" /></FormItem>
        <FormItem label="公里数"><InputNumber v-model:value="addForm.zkm" :min="0.1" :max="50" :step="0.5" style="width: 100%" /></FormItem>
        <FormItem label="开始小时"><InputNumber v-model:value="addForm.ks_h" :min="6" :max="22" style="width: 100%" /></FormItem>
        <FormItem label="开始分钟"><InputNumber v-model:value="addForm.ks_m" :min="0" :max="59" style="width: 100%" /></FormItem>
        <FormItem label="结束小时"><InputNumber v-model:value="addForm.js_h" :min="6" :max="22" style="width: 100%" /></FormItem>
        <FormItem label="结束分钟"><InputNumber v-model:value="addForm.js_m" :min="0" :max="59" style="width: 100%" /></FormItem>
        <FormItem label="跑步周天">
          <Checkbox.Group v-model:value="selectedWeeks" :options="weekOptions" @change="syncWeeks" />
        </FormItem>
        <FormItem label="轮询模式">
          <Select v-model:value="addForm.isPolling">
            <SelectOption :value="0">关闭</SelectOption>
            <SelectOption :value="1">开启</SelectOption>
          </Select>
        </FormItem>
      </Form>
    </Modal>
  </Page>
</template>

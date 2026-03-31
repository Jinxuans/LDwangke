<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal, message,
  Select, SelectOption, Form, FormItem, Row, Col, TimePicker, Textarea,
} from 'ant-design-vue';
import type { AppuiOrder, AppuiCourse } from '#/api/plugins/appui';
import {
  appuiOrderListApi,
  appuiGetCoursesApi,
  appuiGetPriceApi,
  appuiAddOrderApi,
  appuiEditOrderApi,
  appuiRenewOrderApi,
  appuiDeleteOrderApi,
} from '#/api/plugins/appui';

// ---------- 状态 ----------
const loading = ref(false);
const orders = ref<AppuiOrder[]>([]);
const total = ref(0);
const pagination = reactive({ page: 1, limit: 20 });
const search = reactive({ type: '1', keyword: '' });

// 平台列表
const courseList = ref<AppuiCourse[]>([]);

// 添加弹窗
const addVisible = ref(false);
const addLoading = ref(false);
const addForm = reactive({
  pid: '',
  week: [] as string[],
  report: [] as string[],
  shangban_time: '',
  xiaban_time: '',
  school: '',
  user: '',
  pass: '',
  userName: '',
  address: '',
  days: 30,
});

// 编辑弹窗
const editVisible = ref(false);
const editLoading = ref(false);
const editForm = reactive({
  id: 0,
  week: [] as string[],
  report: [] as string[],
  shangban_time: '',
  xiaban_time: '',
  pass: '',
  address: '',
});

// 续费弹窗
const renewVisible = ref(false);
const renewLoading = ref(false);
const renewForm = reactive({ id: 0, pid: '', days: 30 });

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
const reportOptions = [
  { label: '日报', value: 'daily' },
  { label: '周报', value: 'weekly' },
  { label: '月报', value: 'monthly' },
];

const searchTypeOptions = [
  { label: '订单ID', value: '1' },
  { label: '下单账号', value: '2' },
  { label: '下单密码', value: '3' },
  { label: '用户UID', value: '4' },
];

// 当前选中的平台信息
const selectedCourse = computed(() => courseList.value.find((c) => c.pid === addForm.pid));
const courseTip = computed(() => selectedCourse.value?.content || '');
const showSchoolInput = computed(() => selectedCourse.value?.yes_school === 1);

// 价格
const addPrice = ref(0);
const renewPrice = ref(0);

// ---------- 列表 ----------
async function fetchOrders() {
  loading.value = true;
  try {
    const res: any = await appuiOrderListApi({
      page: pagination.page,
      limit: pagination.limit,
      searchType: search.keyword ? search.type : undefined,
      keyword: search.keyword || undefined,
    });
    orders.value = res.list || [];
    total.value = res.total || 0;
  } finally {
    loading.value = false;
  }
}

function getPlatformName(pid: string) {
  const c = courseList.value.find((item) => item.pid === pid);
  return c ? c.name : pid;
}

// ---------- 下单 ----------
async function calcAddPrice() {
  if (!addForm.pid || addForm.days < 1) return;
  try {
    const res: any = await appuiGetPriceApi(addForm.pid, addForm.days);
    addPrice.value = res.price || 0;
  } catch { addPrice.value = 0; }
}

async function handleAdd() {
  if (!addForm.pid || !addForm.user || !addForm.pass || addForm.days < 1) {
    message.warning('请填写完整信息');
    return;
  }
  addLoading.value = true;
  try {
    await appuiAddOrderApi({
      pid: addForm.pid,
      user: addForm.user,
      pass: addForm.pass,
      userName: addForm.userName,
      address: addForm.address,
      days: addForm.days,
      week: addForm.week.join(','),
      report: addForm.report.join(','),
      shangban_time: addForm.shangban_time,
      xiaban_time: addForm.xiaban_time,
      school: addForm.school,
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

// ---------- 编辑 ----------
function openEdit(record: AppuiOrder) {
  editForm.id = record.id;
  editForm.pass = record.pass;
  editForm.address = record.address;
  editForm.week = record.week ? record.week.split(',') : [];
  editForm.report = record.report ? record.report.split(',') : [];
  editForm.shangban_time = record.shangban_time;
  editForm.xiaban_time = record.xiaban_time;
  editVisible.value = true;
}

async function handleEdit() {
  editLoading.value = true;
  try {
    await appuiEditOrderApi({
      id: editForm.id,
      pass: editForm.pass,
      address: editForm.address,
      week: editForm.week.join(','),
      report: editForm.report.join(','),
      shangban_time: editForm.shangban_time,
      xiaban_time: editForm.xiaban_time,
    });
    message.success('修改成功');
    editVisible.value = false;
    fetchOrders();
  } catch (e: any) {
    message.error(e?.message || '修改失败');
  } finally {
    editLoading.value = false;
  }
}

// ---------- 续费 ----------
function openRenew(record: AppuiOrder) {
  renewForm.id = record.id;
  renewForm.pid = record.pid;
  renewForm.days = 30;
  renewPrice.value = 0;
  renewVisible.value = true;
  calcRenewPrice();
}

async function calcRenewPrice() {
  if (!renewForm.pid || renewForm.days < 1) return;
  try {
    const res: any = await appuiGetPriceApi(renewForm.pid, renewForm.days);
    renewPrice.value = res.price || 0;
  } catch { renewPrice.value = 0; }
}

async function handleRenew() {
  renewLoading.value = true;
  try {
    await appuiRenewOrderApi(renewForm.id, renewForm.days);
    message.success('续费成功');
    renewVisible.value = false;
    fetchOrders();
  } catch (e: any) {
    message.error(e?.message || '续费失败');
  } finally {
    renewLoading.value = false;
  }
}

// ---------- 退款 ----------
function handleRefund(record: AppuiOrder) {
  Modal.confirm({
    title: '确认退款',
    content: `确定要退款订单 #${record.id}（账号：${record.user}）吗？`,
    onOk: async () => {
      try {
        await appuiDeleteOrderApi(record.id);
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
  { title: '平台', dataIndex: 'pid', width: 100 },
  { title: '账号', dataIndex: 'user', width: 120 },
  { title: '姓名', dataIndex: 'name', width: 80 },
  { title: '天数', key: 'days', width: 90 },
  { title: '状态', dataIndex: 'status', width: 80 },
  { title: '打卡地址', dataIndex: 'address', ellipsis: true },
  { title: '下单时间', dataIndex: 'addtime', width: 150 },
  { title: '操作', key: 'action', width: 160, fixed: 'right' as const },
];

onMounted(async () => {
  try {
    const res: any = await appuiGetCoursesApi();
    courseList.value = res || [];
  } catch {}
  fetchOrders();
});
</script>

<template>
  <Page title="Appui打卡" description="管理Appui打卡订单">
    <Card class="mb-4" :bordered="false">
      <div class="flex flex-wrap items-center gap-3">
        <Button type="primary" @click="addVisible = true">添加订单</Button>
        <div class="flex-1" />
        <Space wrap>
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
        current: pagination.page,
        pageSize: pagination.limit,
        total,
        showSizeChanger: true,
        showTotal: (t: number) => `共 ${t} 条`,
        pageSizeOptions: ['20', '50', '100'],
        onChange: (p: number, s: number) => { pagination.page = p; pagination.limit = s; fetchOrders(); },
        onShowSizeChange: (_: number, s: number) => { pagination.limit = s; pagination.page = 1; fetchOrders(); },
      }" row-key="id" :scroll="{ x: 1000 }" size="small">
        <template #bodyCell="{ column, record }">
          <template v-if="column.dataIndex === 'pid'">{{ getPlatformName(record.pid) }}</template>
          <template v-if="column.key === 'days'">{{ record.residue_day }} / {{ record.total_day }}</template>
          <template v-if="column.dataIndex === 'status'">
            <Tag v-if="record.status === '进行中'" color="processing">进行中</Tag>
            <Tag v-else-if="record.status === '已完成'" color="success">已完成</Tag>
            <Tag v-else-if="record.status === '待处理'" color="warning">待处理</Tag>
            <Tag v-else-if="record.status === '已退款'" color="default">已退款</Tag>
            <Tag v-else color="error">{{ record.status }}</Tag>
          </template>
          <template v-if="column.key === 'action'">
            <Space>
              <Button size="small" @click="openEdit(record)">编辑</Button>
              <Button size="small" @click="openRenew(record)">续费</Button>
              <Button size="small" danger @click="handleRefund(record)">退款</Button>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <!-- 添加弹窗 -->
    <Modal v-model:open="addVisible" title="添加订单" width="600px" :confirm-loading="addLoading" @ok="handleAdd">
      <Form layout="horizontal" :label-col="{ span: 5 }">
        <FormItem label="选择平台">
          <Select v-model:value="addForm.pid" placeholder="请选择平台" @change="calcAddPrice">
            <SelectOption v-for="c in courseList" :key="c.pid" :value="c.pid">{{ c.name }}（{{ c.price }}/天）</SelectOption>
          </Select>
        </FormItem>
        <FormItem v-if="courseTip" label="课程说明"><span style="color: red">{{ courseTip }}</span></FormItem>
        <FormItem label="周期选择">
          <Select v-model:value="addForm.week" mode="multiple" placeholder="请选择周期">
            <SelectOption v-for="w in weekOptions" :key="w.value" :value="w.value">{{ w.label }}</SelectOption>
          </Select>
        </FormItem>
        <FormItem label="报表选择">
          <Select v-model:value="addForm.report" mode="multiple" placeholder="请选择报表">
            <SelectOption v-for="r in reportOptions" :key="r.value" :value="r.value">{{ r.label }}</SelectOption>
          </Select>
        </FormItem>
        <FormItem label="上班时间"><TimePicker v-model:value="addForm.shangban_time" format="HH:mm" value-format="HH:mm" /></FormItem>
        <FormItem label="下班时间"><TimePicker v-model:value="addForm.xiaban_time" format="HH:mm" value-format="HH:mm" /></FormItem>
        <FormItem v-if="showSchoolInput" label="用户学校"><Input v-model:value="addForm.school" placeholder="请输入学校" /></FormItem>
        <FormItem label="用户账号"><Input v-model:value="addForm.user" placeholder="请输入账号" /></FormItem>
        <FormItem label="用户密码"><Input v-model:value="addForm.pass" placeholder="请输入密码" /></FormItem>
        <FormItem label="用户姓名"><Input v-model:value="addForm.userName" placeholder="查询后自动填充" /></FormItem>
        <FormItem label="签到地址"><Textarea v-model:value="addForm.address" placeholder="请输入签到地址" :rows="2" /></FormItem>
        <FormItem label="下单天数">
          <InputNumber v-model:value="addForm.days" :min="1" :max="365" @change="calcAddPrice" />
          <span style="color: red; margin-left: 12px; font-weight: 800; font-size: 18px">￥{{ addPrice.toFixed(2) }}</span>
        </FormItem>
      </Form>
    </Modal>

    <!-- 编辑弹窗 -->
    <Modal v-model:open="editVisible" title="编辑订单" :confirm-loading="editLoading" @ok="handleEdit">
      <Form layout="horizontal" :label-col="{ span: 5 }">
        <FormItem label="周期选择">
          <Select v-model:value="editForm.week" mode="multiple" placeholder="请选择周期">
            <SelectOption v-for="w in weekOptions" :key="w.value" :value="w.value">{{ w.label }}</SelectOption>
          </Select>
        </FormItem>
        <FormItem label="报表选择">
          <Select v-model:value="editForm.report" mode="multiple" placeholder="请选择报表">
            <SelectOption v-for="r in reportOptions" :key="r.value" :value="r.value">{{ r.label }}</SelectOption>
          </Select>
        </FormItem>
        <FormItem label="上班时间"><TimePicker v-model:value="editForm.shangban_time" format="HH:mm" value-format="HH:mm" /></FormItem>
        <FormItem label="下班时间"><TimePicker v-model:value="editForm.xiaban_time" format="HH:mm" value-format="HH:mm" /></FormItem>
        <FormItem label="用户密码"><Input v-model:value="editForm.pass" placeholder="请输入用户密码" /></FormItem>
        <FormItem label="签到地址"><Textarea v-model:value="editForm.address" placeholder="请输入签到地址" :rows="2" /></FormItem>
      </Form>
    </Modal>

    <!-- 续费弹窗 -->
    <Modal v-model:open="renewVisible" title="续费" :confirm-loading="renewLoading" @ok="handleRenew">
      <Form layout="horizontal" :label-col="{ span: 5 }">
        <FormItem label="续费天数">
          <InputNumber v-model:value="renewForm.days" :min="1" :max="365" @change="calcRenewPrice" />
          <span style="color: red; margin-left: 12px; font-weight: 800; font-size: 18px">￥{{ renewPrice.toFixed(2) }}</span>
        </FormItem>
      </Form>
    </Modal>

  </Page>
</template>

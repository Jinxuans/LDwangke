<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal, message,
  Select, SelectOption, Form, FormItem, Row, Col, Alert, Tabs, TabPane,
} from 'ant-design-vue';
import type { YDSJOrder } from '#/api/plugins/ydsj';
import {
  ydsjOrderListApi, ydsjAddOrderApi, ydsjRefundOrderApi,
  ydsjEditRemarksApi, ydsjSyncOrderApi, ydsjToggleRunApi,
  ydsjGetSchoolsApi, ydsjGetPriceApi,
} from '#/api/plugins/ydsj';

// ---------- 状态 ----------
const loading = ref(false);
const orders = ref<YDSJOrder[]>([]);
const total = ref(0);
const pagination = reactive({ page: 1, limit: 20 });
const search = reactive({ type: '1', keyword: '', status: '' });

// 学校列表
const schools = ref<any[]>([]);

// 添加弹窗
const addVisible = ref(false);
const addLoading = ref(false);
const addForm = reactive({
  school: '',
  user: '',
  pass: '',
  distance: '2',
  run_type: 1,
  start_hour: '12',
  start_minute: '0',
  end_hour: '21',
  end_minute: '0',
  run_week: ['1', '2', '3', '4', '5', '6', '7'] as string[],
  remarks: '',
});
const estimatedPrice = ref<number | null>(null);
const priceLoading = ref(false);

// 订单详情弹窗
const detailVisible = ref(false);
const detailRecord = ref<YDSJOrder | null>(null);
const detailTab = ref('info');

// 修改备注弹窗
const remarksVisible = ref(false);
const remarksLoading = ref(false);
const remarksForm = reactive({ id: 0, user: '', remarks: '' });

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
  { label: '运动世界晨跑', value: 0 },
  { label: '运动世界课外跑', value: 1 },
  { label: '小步点课外跑', value: 2 },
  { label: '小步点晨跑', value: 3 },
];
const statusOptions = [
  { label: '所有订单', value: '' },
  { label: '等待处理', value: '1' },
  { label: '处理成功', value: '2' },
  { label: '处理失败', value: '3' },
  { label: '退款成功', value: '4' },
  { label: '申请退款', value: '5' },
];
const searchTypeOptions = [
  { label: '订单ID', value: '1' },
  { label: '下单账号', value: '2' },
  { label: '下单密码', value: '3' },
  { label: '用户UID', value: '4' },
];

// ---------- 学校 ----------
async function loadSchools() {
  try {
    const res = await ydsjGetSchoolsApi();
    schools.value = res || [];
  } catch (e) { console.error(e); }
}

// ---------- 价格预览 ----------
async function refreshPrice() {
  const dist = parseFloat(addForm.distance);
  if (!dist || dist <= 0 || !addForm.school) {
    estimatedPrice.value = null;
    return;
  }
  priceLoading.value = true;
  try {
    const res: any = await ydsjGetPriceApi(addForm.run_type, dist);
    estimatedPrice.value = res?.price ?? null;
  } catch { estimatedPrice.value = null; }
  finally { priceLoading.value = false; }
}

// 监听影响价格的字段
watch(() => [addForm.run_type, addForm.distance, addForm.school], () => {
  refreshPrice();
});

// ---------- 列表 ----------
async function fetchOrders() {
  loading.value = true;
  try {
    const res: any = await ydsjOrderListApi({
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
  const map: Record<number, string> = { 1: '等待处理', 2: '处理成功', 3: '处理失败', 4: '退款成功', 5: '申请退款' };
  return map[s] || `${s}`;
}
function getStatusColor(s: number) {
  const map: Record<number, string> = { 1: 'warning', 2: 'success', 3: 'error', 4: 'default', 5: 'processing' };
  return map[s] || 'default';
}
function getRunTypeText(t: number) {
  return runTypeOptions.find((o) => o.value === t)?.label || `${t}`;
}
function getWeekText(week: string) {
  if (!week) return '';
  const map: Record<string, string> = { '1': '一', '2': '二', '3': '三', '4': '四', '5': '五', '6': '六', '7': '日' };
  return week.split(',').map((w) => `周${map[w] || w}`).join(' ');
}

// ---------- 下单 ----------
function openAdd() {
  Object.assign(addForm, {
    school: '', user: '', pass: '', distance: '2', run_type: 1,
    start_hour: '12', start_minute: '0', end_hour: '21', end_minute: '0',
    run_week: ['1', '2', '3', '4', '5', '6', '7'], remarks: '',
  });
  estimatedPrice.value = null;
  addVisible.value = true;
}

async function handleAdd() {
  if (!addForm.user || !addForm.pass || !addForm.distance) {
    message.warning('请填写完整信息');
    return;
  }
  addLoading.value = true;
  try {
    await ydsjAddOrderApi({
      school: addForm.school || '自动识别',
      user: addForm.user,
      pass: addForm.pass,
      distance: addForm.distance,
      run_type: addForm.run_type,
      start_hour: addForm.start_hour,
      start_minute: addForm.start_minute,
      end_hour: addForm.end_hour,
      end_minute: addForm.end_minute,
      run_week: addForm.run_week.join(','),
      remarks: addForm.remarks,
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
function handleRefund(record: YDSJOrder) {
  Modal.confirm({
    title: '确认退款',
    content: `确定要退款订单 #${record.id}（账号：${record.user}）吗？`,
    onOk: async () => {
      try {
        await ydsjRefundOrderApi(record.id);
        message.success('退款成功');
        fetchOrders();
      } catch (e: any) {
        message.error(e?.message || '退款失败');
      }
    },
  });
}

// ---------- 修改备注 ----------
function openEditRemarks(record: YDSJOrder) {
  remarksForm.id = record.id;
  remarksForm.user = record.user;
  remarksForm.remarks = record.remarks || '';
  remarksVisible.value = true;
}

async function handleSaveRemarks() {
  remarksLoading.value = true;
  try {
    await ydsjEditRemarksApi(remarksForm.id, remarksForm.remarks);
    message.success('备注修改成功');
    remarksVisible.value = false;
    fetchOrders();
  } catch (e: any) {
    message.error(e?.message || '修改失败');
  } finally {
    remarksLoading.value = false;
  }
}

// ---------- 订单详情 ----------
function openDetail(record: YDSJOrder) {
  detailRecord.value = record;
  detailTab.value = 'info';
  detailVisible.value = true;
}

// ---------- 手动同步 ----------
async function handleSync(record: YDSJOrder) {
  try {
    await ydsjSyncOrderApi(record.id);
    message.success('同步成功');
    fetchOrders();
  } catch (e: any) {
    message.error(e?.message || '同步失败');
  }
}

// ---------- 切换跑步状态 ----------
async function handleToggleRun(record: YDSJOrder) {
  try {
    await ydsjToggleRunApi(record.id);
    message.success(record.is_run === 1 ? '已暂停' : '已开启');
    fetchOrders();
  } catch (e: any) {
    message.error(e?.message || '操作失败');
  }
}

// ---------- 表格列 ----------
const columns = [
  { title: 'ID', dataIndex: 'id', width: 70 },
  { title: 'UID', dataIndex: 'uid', width: 60 },
  { title: '学校', dataIndex: 'school', width: 100, ellipsis: true },
  { title: '账号', key: 'account', width: 130 },
  { title: '里程', key: 'distance', width: 80 },
  { title: '类型', key: 'run_type', width: 110 },
  { title: '时间', key: 'time', width: 100 },
  { title: '跑步周期', key: 'run_week', width: 140 },
  { title: '跑步状态', key: 'is_run', width: 90 },
  { title: '订单状态', key: 'status', width: 90 },
  { title: '费用', key: 'fees', width: 110 },
  { title: '备注', dataIndex: 'remarks', ellipsis: true },
  { title: '下单时间', dataIndex: 'addtime', width: 150 },
  { title: '操作', key: 'action', width: 240, fixed: 'right' as const },
];

onMounted(async () => {
  await loadSchools();
  fetchOrders();
});
</script>

<template>
  <Page title="运动世界" description="管理运动世界/小步点跑步订单">
    <Card class="mb-4" :bordered="false">
      <Alert message="下单后请提醒学生退出登录（不是关闭APP，是在设置里退出账号登录）！录单跑步时间最好大于4个小时！" type="warning" show-icon class="mb-4" />
      <div class="flex flex-wrap items-center gap-3">
        <Space>
          <Button type="primary" @click="openAdd">添加订单</Button>
        </Space>
        <div class="flex-1" />
        <Space wrap>
          <Select v-model:value="search.status" style="width: 120px" @change="() => { pagination.page = 1; fetchOrders(); }">
            <SelectOption v-for="o in statusOptions" :key="o.value" :value="o.value">{{ o.label }}</SelectOption>
          </Select>
          <Input.Group compact>
            <Select v-model:value="search.type" style="width: 100px">
              <SelectOption v-for="o in searchTypeOptions" :key="o.value" :value="o.value">{{ o.label }}</SelectOption>
            </Select>
            <Input.Search v-model:value="search.keyword" placeholder="关键词" style="width: 180px" @search="() => { pagination.page = 1; fetchOrders(); }" allow-clear />
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
      }" row-key="id" :scroll="{ x: 1500 }" size="small">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'account'">
            <div class="leading-tight">{{ record.user }}</div>
            <div class="text-gray-400 text-xs">{{ record.pass }}</div>
          </template>
          <template v-else-if="column.key === 'distance'">
            {{ record.distance }} KM
          </template>
          <template v-else-if="column.key === 'run_type'">{{ getRunTypeText(record.run_type) }}</template>
          <template v-else-if="column.key === 'time'">
            {{ record.start_hour }}:{{ String(record.start_minute).padStart(2, '0') }} - {{ record.end_hour }}:{{ String(record.end_minute).padStart(2, '0') }}
          </template>
          <template v-else-if="column.key === 'run_week'">{{ getWeekText(record.run_week) }}</template>
          <template v-else-if="column.key === 'is_run'">
            <Tag :color="record.is_run === 1 ? 'success' : 'default'">{{ record.is_run === 1 ? '运行中' : '已暂停' }}</Tag>
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="getStatusColor(record.status)">{{ getStatusText(record.status) }}</Tag>
          </template>
          <template v-else-if="column.key === 'fees'">
            <div class="leading-tight">预扣: ¥{{ record.fees || '0' }}</div>
            <div class="text-gray-400 text-xs" v-if="record.real_fees">实际: ¥{{ record.real_fees }}</div>
            <div class="text-green-500 text-xs" v-if="record.refund_money && record.refund_money !== '0' && record.refund_money !== ''">退款: ¥{{ record.refund_money }}</div>
          </template>
          <template v-else-if="column.key === 'action'">
            <Space :size="4" wrap>
              <Button size="small" @click="openDetail(record)">详情</Button>
              <Button size="small" @click="openEditRemarks(record)">备注</Button>
              <Button size="small" @click="handleToggleRun(record)">{{ record.is_run === 1 ? '暂停' : '开启' }}</Button>
              <Button size="small" @click="handleSync(record)">同步</Button>
              <Button size="small" danger @click="handleRefund(record)" :disabled="record.status === 4 || record.status === 5">退款</Button>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <!-- 添加弹窗 -->
    <Modal v-model:open="addVisible" title="添加订单" width="600px" :confirm-loading="addLoading" @ok="handleAdd">
      <Form layout="horizontal" :label-col="{ span: 5 }" class="mt-4">
        <FormItem label="跑步类型">
          <Select v-model:value="addForm.run_type">
            <SelectOption v-for="o in runTypeOptions" :key="o.value" :value="o.value">{{ o.label }}</SelectOption>
          </Select>
        </FormItem>
        <FormItem label="跑步学校">
          <Select v-model:value="addForm.school" show-search placeholder="请选择学校（可搜索）" allow-clear
            :filter-option="(input: string, option: any) => (option?.label ?? '').toLowerCase().includes(input.toLowerCase())"
            style="width: 100%">
            <SelectOption value="">自动识别</SelectOption>
            <SelectOption v-for="s in schools" :key="s.school_name" :value="s.school_name" :label="s.school_name">
              {{ s.school_name }}
            </SelectOption>
          </Select>
        </FormItem>
        <FormItem label="用户账号"><Input v-model:value="addForm.user" placeholder="请输入账号" /></FormItem>
        <FormItem label="用户密码"><Input v-model:value="addForm.pass" placeholder="请输入密码" /></FormItem>
        <FormItem label="总公里数">
          <InputNumber v-model:value="addForm.distance" :min="0.1" :max="500" :step="0.1" :precision="1" style="width: 100%" />
        </FormItem>
        <FormItem label="开始时间">
          <Row :gutter="8">
            <Col :span="11">
              <InputNumber v-model:value="addForm.start_hour" :min="0" :max="23" :step="1" style="width: 100%" addon-after="时" />
            </Col>
            <Col :span="2" class="text-center leading-8">:</Col>
            <Col :span="11">
              <InputNumber v-model:value="addForm.start_minute" :min="0" :max="59" :step="1" style="width: 100%" addon-after="分" />
            </Col>
          </Row>
        </FormItem>
        <FormItem label="结束时间">
          <Row :gutter="8">
            <Col :span="11">
              <InputNumber v-model:value="addForm.end_hour" :min="0" :max="23" :step="1" style="width: 100%" addon-after="时" />
            </Col>
            <Col :span="2" class="text-center leading-8">:</Col>
            <Col :span="11">
              <InputNumber v-model:value="addForm.end_minute" :min="0" :max="59" :step="1" style="width: 100%" addon-after="分" />
            </Col>
          </Row>
        </FormItem>
        <FormItem label="跑步周期">
          <Select v-model:value="addForm.run_week" mode="multiple" placeholder="请选择跑步周期">
            <SelectOption v-for="w in weekOptions" :key="w.value" :value="w.value">{{ w.label }}</SelectOption>
          </Select>
        </FormItem>
        <FormItem label="备注">
          <Input.TextArea v-model:value="addForm.remarks" placeholder="请输入备注（可选）" :rows="2" />
        </FormItem>
        <FormItem label="预估费用">
          <span v-if="estimatedPrice !== null" style="color: red; font-weight: 800; font-size: 18px">¥{{ estimatedPrice.toFixed(2) }}</span>
          <span v-else-if="priceLoading" class="text-gray-400">计算中...</span>
          <span v-else class="text-gray-400">请选择学校和公里数</span>
        </FormItem>
      </Form>
    </Modal>

    <!-- 修改备注弹窗 -->
    <Modal v-model:open="remarksVisible" title="修改备注" width="440px" :confirm-loading="remarksLoading" @ok="handleSaveRemarks">
      <Form layout="horizontal" :label-col="{ span: 5 }" class="mt-4">
        <FormItem label="订单">
          <span>#{{ remarksForm.id }}（{{ remarksForm.user }}）</span>
        </FormItem>
        <FormItem label="备注">
          <Input.TextArea v-model:value="remarksForm.remarks" placeholder="请输入备注" :rows="3" />
        </FormItem>
      </Form>
    </Modal>

    <!-- 订单详情弹窗 -->
    <Modal v-model:open="detailVisible" title="订单详情" width="650px" :footer="null">
      <template v-if="detailRecord">
        <Tabs v-model:activeKey="detailTab">
          <TabPane key="info" tab="订单信息">
            <div v-if="detailRecord.info" v-html="detailRecord.info" class="text-sm" />
            <div v-else class="text-gray-400">暂无订单信息</div>
          </TabPane>
          <TabPane key="tmp_info" tab="退款/操作信息">
            <div v-if="detailRecord.tmp_info" v-html="detailRecord.tmp_info" class="text-sm" />
            <div v-else class="text-gray-400">暂无操作信息</div>
          </TabPane>
        </Tabs>
      </template>
    </Modal>

  </Page>
</template>

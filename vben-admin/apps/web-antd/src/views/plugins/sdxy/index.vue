<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal, message,
  Select, SelectOption, Form, FormItem, DatePicker, Alert, Steps, Step,
  Spin, Descriptions, DescriptionsItem, Checkbox, CheckboxGroup, Divider,
  Radio, RadioGroup,
} from 'ant-design-vue';
import type { SDXYOrder } from '#/api/plugins/sdxy';
import {
  sdxyOrderListApi,
  sdxyGetPriceApi,
  sdxyAddOrderApi,
  sdxyRefundOrderApi,
  sdxyPauseOrderApi,
  sdxyGetUserInfoApi,
  sdxySendCodeApi,
  sdxyGetUserInfoByCodeApi,
  sdxyGetRunTaskApi,
  sdxyDelayTaskApi,
} from '#/api/plugins/sdxy';

// ---------- 状态 ----------
const loading = ref(false);
const orders = ref<SDXYOrder[]>([]);
const total = ref(0);
const pagination = reactive({ page: 1, limit: 20 });
const search = reactive({ type: '1', keyword: '', status: '' });
const pricePerTask = ref(0);

// 添加弹窗 & 步骤
const addVisible = ref(false);
const addLoading = ref(false);
const addStep = ref(0); // 0=填写账户, 1=选择参数, 2=确认下单
const queryLoading = ref(false);
const queryMsg = ref('');

// 登录方式: password / code
const loginMode = ref<'password' | 'code'>('password');
const codeCountdown = ref(0);

// 用户信息（从上游查询返回）
interface ZoneItem { id: string; name: string; [key: string]: any; }
interface RunRuleItem { id: string; label: string; [key: string]: any; }
const studentInfo = ref<Record<string, any> | null>(null);
const zones = ref<ZoneItem[]>([]);
const runRules = ref<RunRuleItem[]>([]);

const addForm = reactive({
  phone: '',
  password: '',
  code: '', // 验证码
  student_id: '',
  zone_id: '',
  zone_name: '',
  run_type: '1', // 1=有效跑 2=自由跑
  run_rule_id: '',
  dis: 2,
  start_date: '',
  start_hour: 6,
  start_minute: 0,
  end_hour: 8,
  end_minute: 0,
  day: 7,
  run_week: [1, 2, 3, 4, 5] as number[],
});

const weekOptions = [
  { label: '周一', value: 1 },
  { label: '周二', value: 2 },
  { label: '周三', value: 3 },
  { label: '周四', value: 4 },
  { label: '周五', value: 5 },
  { label: '周六', value: 6 },
  { label: '周日', value: 7 },
];
const statusOptions = [
  { label: '所有订单', value: '' },
  { label: '等待处理', value: '1' },
  { label: '进行中', value: '2' },
  { label: '已完成', value: '3' },
  { label: '失败', value: '4' },
  { label: '已退款', value: '5' },
];
const searchTypeOptions = [
  { label: '订单ID', value: '1' },
  { label: '下单账号', value: '2' },
  { label: '下单密码', value: '3' },
  { label: '用户UID', value: '4' },
];
const statusMap: Record<string, { text: string; color: string }> = {
  '1': { text: '等待处理', color: 'warning' },
  '2': { text: '进行中', color: 'processing' },
  '3': { text: '已完成', color: 'success' },
  '4': { text: '失败', color: 'error' },
  '5': { text: '已退款', color: 'default' },
};

// 生成 task_list
function generateTaskList() {
  if (!addForm.start_date || addForm.day <= 0) return [];
  const tasks: Record<string, any>[] = [];
  const startDate = new Date(addForm.start_date);
  const startTime = `${String(addForm.start_hour).padStart(2, '0')}:${String(addForm.start_minute).padStart(2, '0')}`;
  const endTime = `${String(addForm.end_hour).padStart(2, '0')}:${String(addForm.end_minute).padStart(2, '0')}`;

  for (let i = 0; i < addForm.day * 2; i++) {
    const d = new Date(startDate);
    d.setDate(d.getDate() + i);
    const dayOfWeek = d.getDay() === 0 ? 7 : d.getDay(); // 1=Mon...7=Sun
    if (!addForm.run_week.includes(dayOfWeek)) continue;

    const dateStr = d.toISOString().split('T')[0]!;
    tasks.push({
      date: dateStr,
      start_time: startTime,
      end_time: endTime,
    });
    if (tasks.length >= addForm.day) break;
  }
  return tasks;
}

// 预估任务数
const taskCount = computed(() => generateTaskList().length);

// 预估价格
const estimatedPrice = computed(() => {
  return Math.round(pricePerTask.value * taskCount.value * 100) / 100;
});

// ---------- 加载 ----------
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

// ---------- 查询用户信息 ----------
async function queryUserInfo() {
  if (!addForm.phone) { message.warning('请输入手机号'); return; }

  queryLoading.value = true;
  queryMsg.value = '';
  studentInfo.value = null;
  zones.value = [];
  runRules.value = [];

  try {
    let res: any;
    if (loginMode.value === 'password') {
      res = await sdxyGetUserInfoApi({ phone: addForm.phone, password: addForm.password });
    } else {
      if (!addForm.code) { message.warning('请输入验证码'); queryLoading.value = false; return; }
      res = await sdxyGetUserInfoByCodeApi({ phone: addForm.phone, code: addForm.code });
    }

    // 解析上游响应 — 兼容多种格式
    const data = res;
    if (!data) {
      queryMsg.value = '上游无响应';
      addStep.value = 1;
      return;
    }

    // 提取学生信息
    const student = data.student ?? data;
    studentInfo.value = student;

    // 提取 student_id
    addForm.student_id = String(student.id ?? student.student_id ?? '');

    // 提取跑区列表
    const zoneList = student.zones ?? student.zone_list ?? data.zones ?? [];
    if (Array.isArray(zoneList)) {
      zones.value = zoneList.map((z: any) => ({
        id: String(z.id ?? z.zone_id ?? ''),
        name: String(z.name ?? z.zone_name ?? ''),
        ...z,
      }));
    }

    // 提取运行规则列表
    const ruleList = student.run_rules ?? student.run_rule_list ?? data.run_rules ?? [];
    if (Array.isArray(ruleList)) {
      runRules.value = ruleList.map((r: any) => ({
        id: String(r.id ?? r.run_rule_id ?? ''),
        label: String(r.label ?? r.name ?? ''),
        ...r,
      }));
    }
    // 如果有单个 run_rule 对象
    if (runRules.value.length === 0 && student.run_rule) {
      const rr = student.run_rule;
      runRules.value = [{
        id: String(rr.id ?? rr.run_rule_id ?? ''),
        label: String(rr.label ?? rr.name ?? ''),
        ...rr,
      }];
      addForm.run_rule_id = runRules.value[0]!.id;
    }

    // 提取学校名
    if (student.school) {
      const schoolName = typeof student.school === 'string' ? student.school : (student.school.name ?? '');
      addForm.zone_name = schoolName;
    }

    queryMsg.value = '查询成功';
    addStep.value = 1;
  } catch (e: any) {
    console.error('查询用户信息失败', e);
    queryMsg.value = e?.message || '查询失败，请检查账号密码';
    addStep.value = 1;
  } finally {
    queryLoading.value = false;
  }
}

// 发送验证码
async function handleSendCode() {
  if (!addForm.phone) { message.warning('请输入手机号'); return; }
  try {
    await sdxySendCodeApi({ phone: addForm.phone });
    message.success('验证码已发送');
    codeCountdown.value = 60;
    const timer = setInterval(() => {
      codeCountdown.value--;
      if (codeCountdown.value <= 0) clearInterval(timer);
    }, 1000);
  } catch (e: any) {
    message.error(e?.message || '发送失败');
  }
}

// 选择跑区
function onZoneSelect(zoneId: string) {
  const z = zones.value.find(x => x.id === zoneId);
  if (z) {
    addForm.zone_id = z.id;
    addForm.zone_name = z.name;
  }
}

// ---------- 下单 ----------
async function handleAdd() {
  if (!addForm.phone) { message.warning('请输入手机号'); return; }
  if (!addForm.zone_id && !addForm.zone_name) { message.warning('请选择跑区'); return; }

  const taskList = generateTaskList();
  if (taskList.length === 0) { message.warning('请配置跑步任务（检查日期和周期设置）'); return; }

  const form: Record<string, any> = {
    phone: addForm.phone,
    password: addForm.password,
    dis: String(addForm.dis),
    zone_id: addForm.zone_id,
    zone_name: addForm.zone_name,
    run_type: addForm.run_type,
    student_id: addForm.student_id,
    run_rule_id: addForm.run_rule_id,
    task_list: taskList,
  };

  addLoading.value = true;
  try {
    await sdxyAddOrderApi(form);
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
  if (!record.agg_order_id) { message.warning('该订单没有源台ID，无法退款'); return; }
  Modal.confirm({
    title: '确认退款',
    content: `确定要退款订单 #${record.id}（账号：${record.user}）吗？`,
    onOk: async () => {
      try {
        await sdxyRefundOrderApi(record.agg_order_id);
        message.success('退款成功');
        fetchOrders();
      } catch (e: any) {
        message.error(e?.message || '退款失败');
      }
    },
  });
}

// ---------- 暂停/恢复 ----------
async function handlePause(record: SDXYOrder) {
  if (!record.agg_order_id) { message.warning('该订单没有源台ID'); return; }
  const newPause = record.pause === 1 ? 0 : 1;
  try {
    await sdxyPauseOrderApi(record.agg_order_id, newPause);
    message.success(newPause === 1 ? '已暂停' : '已恢复');
    fetchOrders();
  } catch (e: any) {
    message.error(e?.message || '操作失败');
  }
}

// ---------- 任务日志 ----------
const logVisible = ref(false);
const logLoading = ref(false);
const logData = ref<any[]>([]);
const logOrderId = ref('');

async function viewLog(record: SDXYOrder) {
  if (!record.sdxy_order_id) { message.warning('该订单没有子订单ID'); return; }
  logOrderId.value = record.sdxy_order_id;
  logVisible.value = true;
  logLoading.value = true;
  try {
    const res: any = await sdxyGetRunTaskApi(record.sdxy_order_id, 1, 50);
    logData.value = res?.list ?? res?.data ?? (Array.isArray(res) ? res : []);
  } catch (e: any) {
    message.error(e?.message || '获取日志失败');
    logData.value = [];
  } finally {
    logLoading.value = false;
  }
}

// ---------- 延时任务 ----------
async function handleDelay(record: SDXYOrder) {
  if (!record.agg_order_id) { message.warning('该订单没有源台ID'); return; }
  Modal.confirm({
    title: '延迟任务',
    content: `确定要延迟订单 #${record.id} 的下一个任务吗？`,
    onOk: async () => {
      try {
        await sdxyDelayTaskApi(record.agg_order_id);
        message.success('延迟成功');
        fetchOrders();
      } catch (e: any) {
        message.error(e?.message || '延迟失败');
      }
    },
  });
}

// ---------- 工具 ----------
function getWeekText(weeks: number[]) {
  const map: Record<number, string> = { 1: '一', 2: '二', 3: '三', 4: '四', 5: '五', 6: '六', 7: '日' };
  return weeks.map(w => `周${map[w] || w}`).join(' ');
}

function openAdd() {
  Object.assign(addForm, {
    phone: '', password: '', code: '', student_id: '',
    zone_id: '', zone_name: '', run_type: '1', run_rule_id: '',
    dis: 2, start_date: '', start_hour: 6, start_minute: 0,
    end_hour: 8, end_minute: 0, day: 7,
    run_week: [1, 2, 3, 4, 5],
  });
  studentInfo.value = null;
  zones.value = [];
  runRules.value = [];
  queryMsg.value = '';
  loginMode.value = 'password';
  addStep.value = 0;
  addVisible.value = true;
}

// ---------- 表格列 ----------
const columns = [
  { title: 'ID', dataIndex: 'id', width: 70 },
  { title: '源台ID', dataIndex: 'agg_order_id', width: 85, ellipsis: true },
  { title: '账号', key: 'account', width: 130 },
  { title: '学校', dataIndex: 'school', width: 100, ellipsis: true },
  { title: '任务数', dataIndex: 'num', width: 70 },
  { title: '公里', dataIndex: 'distance', width: 70 },
  { title: '跑步规则', dataIndex: 'run_rule', width: 100, ellipsis: true },
  { title: '金额', key: 'fees', width: 80 },
  { title: '状态', key: 'status', width: 90 },
  { title: '暂停', key: 'pause', width: 60 },
  { title: '下单时间', dataIndex: 'created_at', width: 150 },
  { title: '操作', key: 'action', width: 220, fixed: 'right' as const },
];

const logColumns = [
  { title: '任务ID', dataIndex: 'id', width: 80 },
  { title: '日期', dataIndex: 'date', width: 110 },
  { title: '开始时间', dataIndex: 'start_time', width: 90 },
  { title: '状态', dataIndex: 'status', width: 80 },
  { title: '结果', dataIndex: 'result', ellipsis: true },
];

onMounted(async () => {
  try {
    const res: any = await sdxyGetPriceApi();
    pricePerTask.value = res.price || 0;
  } catch {}
  fetchOrders();
});
</script>

<template>
  <Page title="闪电运动" description="管理闪电闪动校园跑步订单">
    <Card class="mb-4" :bordered="false">
      <Alert message="下单前请确认账号密码正确，跑步期间切勿登录账号！下单后请联系客服进行短信验证码授权。" type="warning" show-icon class="mb-4" />
      <div class="flex flex-wrap items-center gap-3">
        <Space>
          <Button type="primary" @click="openAdd">添加订单</Button>
          <Tag color="orange">{{ pricePerTask }} 元/次</Tag>
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
      }" row-key="id" :scroll="{ x: 1200 }" size="small">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'account'">
            <div class="leading-tight">{{ record.user }}</div>
            <div class="text-gray-400 text-xs">{{ record.pass }}</div>
          </template>
          <template v-else-if="column.key === 'fees'">
            <span class="font-semibold">¥{{ Number(record.fees).toFixed(2) }}</span>
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="(statusMap[record.status] || { color: 'default' }).color">
              {{ (statusMap[record.status] || { text: record.status }).text }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'pause'">
            <Tag v-if="record.pause === 1" color="orange">暂停</Tag>
            <span v-else class="text-gray-400">-</span>
          </template>
          <template v-else-if="column.key === 'action'">
            <Space>
              <Button size="small" @click="viewLog(record)" :disabled="!record.sdxy_order_id">日志</Button>
              <Button size="small" @click="handlePause(record)" :disabled="!record.agg_order_id || record.status === '5'">
                {{ record.pause === 1 ? '恢复' : '暂停' }}
              </Button>
              <Button size="small" @click="handleDelay(record)" :disabled="!record.agg_order_id || record.status === '5'">延时</Button>
              <Button size="small" danger @click="handleRefund(record)"
                :disabled="!record.agg_order_id || record.status === '5'">退款</Button>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <!-- 添加订单弹窗 —— 多步流程 -->
    <Modal
      v-model:open="addVisible"
      title="添加闪电运动订单"
      width="660px"
      :footer="null"
      :maskClosable="false"
    >
      <Steps :current="addStep" size="small" class="mt-2 mb-6">
        <Step title="填写账户" />
        <Step title="选择参数" />
        <Step title="确认下单" />
      </Steps>

      <!-- ========== Step 0: 填写账户 ========== -->
      <div v-show="addStep === 0">
        <Form layout="horizontal" :label-col="{ span: 5 }">
          <FormItem label="登录方式">
            <RadioGroup v-model:value="loginMode">
              <Radio value="password">密码登录</Radio>
              <Radio value="code">验证码登录</Radio>
            </RadioGroup>
          </FormItem>
          <FormItem label="手机号">
            <Input v-model:value="addForm.phone" placeholder="请输入手机号" />
          </FormItem>
          <FormItem v-if="loginMode === 'password'" label="密码">
            <Input.Password v-model:value="addForm.password" placeholder="选填，部分学校不需要" />
          </FormItem>
          <template v-if="loginMode === 'code'">
            <FormItem label="验证码">
              <Space>
                <Input v-model:value="addForm.code" placeholder="请输入验证码" style="width: 150px" />
                <Button :disabled="codeCountdown > 0" @click="handleSendCode">
                  {{ codeCountdown > 0 ? `${codeCountdown}s` : '发送验证码' }}
                </Button>
              </Space>
            </FormItem>
          </template>
        </Form>
        <div class="flex justify-end gap-2 mt-4">
          <Button @click="addVisible = false">取消</Button>
          <Button type="primary" :loading="queryLoading" @click="queryUserInfo">
            查询用户信息
          </Button>
        </div>
      </div>

      <!-- ========== Step 1: 选择参数 ========== -->
      <div v-show="addStep === 1">
        <Spin :spinning="queryLoading">
          <Form layout="horizontal" :label-col="{ span: 5 }">
            <!-- 用户信息展示 -->
            <FormItem v-if="studentInfo" label="用户信息">
              <div class="text-sm">
                <span v-if="studentInfo.name">{{ studentInfo.name }}</span>
                <span v-if="studentInfo.school" class="ml-2 text-gray-500">
                  {{ typeof studentInfo.school === 'string' ? studentInfo.school : studentInfo.school?.name }}
                </span>
                <span v-if="addForm.student_id" class="ml-2 text-gray-400 text-xs">ID: {{ addForm.student_id }}</span>
              </div>
            </FormItem>

            <!-- 跑区选择 -->
            <FormItem label="跑区">
              <Select
                v-if="zones.length > 0"
                v-model:value="addForm.zone_id"
                placeholder="请选择跑区"
                style="width: 100%"
                show-search
                :filter-option="(input: string, option: any) => (option?.label ?? '').toLowerCase().includes(input.toLowerCase())"
                @change="onZoneSelect"
              >
                <SelectOption v-for="z in zones" :key="z.id" :value="z.id" :label="z.name">
                  {{ z.name }}
                </SelectOption>
              </Select>
              <Input v-else v-model:value="addForm.zone_name" placeholder="手动输入跑区名称" />
              <div v-if="queryMsg" class="text-xs mt-1" :class="zones.length > 0 ? 'text-green-500' : 'text-orange-500'">
                {{ queryMsg }}
                <span v-if="zones.length > 0">（共 {{ zones.length }} 个跑区）</span>
              </div>
            </FormItem>

            <!-- 跑步类型 -->
            <FormItem label="跑步类型">
              <RadioGroup v-model:value="addForm.run_type">
                <Radio value="1">有效跑</Radio>
                <Radio value="2">自由跑</Radio>
              </RadioGroup>
            </FormItem>

            <!-- 运行规则 -->
            <FormItem v-if="runRules.length > 0" label="运行规则">
              <Select v-model:value="addForm.run_rule_id" placeholder="请选择运行规则" style="width: 100%">
                <SelectOption v-for="r in runRules" :key="r.id" :value="r.id">
                  {{ r.label }}
                </SelectOption>
              </Select>
            </FormItem>

            <Divider class="my-3" />

            <!-- 跑步配置 -->
            <FormItem label="每次公里数">
              <InputNumber v-model:value="addForm.dis" :min="0.5" :max="100" :step="0.5" :precision="1" style="width: 100%" />
            </FormItem>
            <FormItem label="开始日期">
              <DatePicker v-model:value="addForm.start_date" value-format="YYYY-MM-DD" style="width: 100%" />
            </FormItem>
            <FormItem label="跑步天数">
              <InputNumber v-model:value="addForm.day" :min="1" :max="365" :step="1" style="width: 100%" />
            </FormItem>
            <FormItem label="开始时间">
              <Space>
                <InputNumber v-model:value="addForm.start_hour" :min="0" :max="23" addon-after="时" style="width: 100px" />
                <InputNumber v-model:value="addForm.start_minute" :min="0" :max="59" addon-after="分" style="width: 100px" />
              </Space>
            </FormItem>
            <FormItem label="结束时间">
              <Space>
                <InputNumber v-model:value="addForm.end_hour" :min="0" :max="23" addon-after="时" style="width: 100px" />
                <InputNumber v-model:value="addForm.end_minute" :min="0" :max="59" addon-after="分" style="width: 100px" />
              </Space>
            </FormItem>
            <FormItem label="跑步周期">
              <CheckboxGroup v-model:value="addForm.run_week" :options="weekOptions" />
            </FormItem>

            <Divider class="my-3" />

            <FormItem label="预计任务">
              <span class="font-semibold">{{ taskCount }} 次</span>
              <span class="text-gray-400 ml-2">({{ addForm.day }} 天 × 筛选周期)</span>
            </FormItem>
            <FormItem label="预估费用">
              <span style="color: red; font-weight: 800; font-size: 18px">¥{{ estimatedPrice.toFixed(2) }}</span>
              <span class="text-gray-400 ml-2">({{ pricePerTask }}元/次 × {{ taskCount }}次)</span>
            </FormItem>
          </Form>
        </Spin>
        <div class="flex justify-end gap-2 mt-4">
          <Button @click="addStep = 0">上一步</Button>
          <Button type="primary" @click="addStep = 2"
            :disabled="(!addForm.zone_id && !addForm.zone_name) || taskCount === 0">
            下一步
          </Button>
        </div>
      </div>

      <!-- ========== Step 2: 确认下单 ========== -->
      <div v-show="addStep === 2">
        <Descriptions bordered :column="1" size="small" class="mb-4">
          <DescriptionsItem label="手机号">{{ addForm.phone }}</DescriptionsItem>
          <DescriptionsItem label="跑区">{{ addForm.zone_name || addForm.zone_id }}</DescriptionsItem>
          <DescriptionsItem label="跑步类型">{{ addForm.run_type === '1' ? '有效跑' : '自由跑' }}</DescriptionsItem>
          <DescriptionsItem v-if="addForm.run_rule_id" label="运行规则">
            {{ runRules.find(r => r.id === addForm.run_rule_id)?.label || addForm.run_rule_id }}
          </DescriptionsItem>
          <DescriptionsItem label="每次公里数">{{ addForm.dis }} km</DescriptionsItem>
          <DescriptionsItem label="开始日期">{{ addForm.start_date }}</DescriptionsItem>
          <DescriptionsItem label="跑步时间">
            {{ String(addForm.start_hour).padStart(2, '0') }}:{{ String(addForm.start_minute).padStart(2, '0') }}
            -
            {{ String(addForm.end_hour).padStart(2, '0') }}:{{ String(addForm.end_minute).padStart(2, '0') }}
          </DescriptionsItem>
          <DescriptionsItem label="跑步周期">{{ getWeekText(addForm.run_week) }}</DescriptionsItem>
          <DescriptionsItem label="任务数">{{ taskCount }} 次</DescriptionsItem>
          <DescriptionsItem label="预估费用">
            <span style="color: red; font-weight: 800; font-size: 16px">¥{{ estimatedPrice.toFixed(2) }}</span>
          </DescriptionsItem>
        </Descriptions>
        <Alert message="请仔细核对以上信息，下单后将自动扣费。跑步期间切勿登录账号！" type="warning" show-icon />
        <div class="flex justify-end gap-2 mt-4">
          <Button @click="addStep = 1">上一步</Button>
          <Button type="primary" danger :loading="addLoading" @click="handleAdd">
            确认下单 ¥{{ estimatedPrice.toFixed(2) }}
          </Button>
        </div>
      </div>
    </Modal>

    <!-- 任务日志弹窗 -->
    <Modal v-model:open="logVisible" :title="`任务日志 - ${logOrderId}`" width="700px" :footer="null">
      <Table :columns="logColumns" :data-source="logData" :loading="logLoading"
        :pagination="false" row-key="id" size="small" :scroll="{ y: 400 }" />
    </Modal>
  </Page>
</template>

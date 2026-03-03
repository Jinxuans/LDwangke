<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal, message,
  Select, SelectOption, Form, FormItem, Alert, Steps, Step, Radio, RadioGroup,
  Spin, Descriptions, DescriptionsItem, Divider,
} from 'ant-design-vue';
import type { WAppUser, WOrder } from '#/api/w';
import {
  wGetAppsApi, wGetOrdersApi, wAddOrderApi,
  wRefundOrderApi, wSyncOrderApi, wResumeOrderApi,
  wProxyActionApi,
} from '#/api/w';

// ---------- 状态 ----------
const loading = ref(false);
const orders = ref<WOrder[]>([]);
const total = ref(0);
const pagination = reactive({ page: 1, page_size: 20 });
const search = reactive({ account: '', status: '', app_id: '' });
const apps = ref<WAppUser[]>([]);

// 添加弹窗 & 步骤
const addVisible = ref(false);
const addLoading = ref(false);
const addStep = ref(0); // 0=填写账户, 1=选择跑区参数, 2=确认下单
const queryLoading = ref(false);

// 跑区数据
interface ZoneItem { id: string; name: string; [key: string]: any; }
const zones = ref<ZoneItem[]>([]);
const queryMsg = ref('');

const addForm = reactive({
  app_id: undefined as number | undefined,
  account: '',
  password: '',
  zone_id: '',
  zone_name: '',
  run_type: 1, // 1=有效跑 2=自由跑
  dis: 2,
  task_count: 7,
});

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '正常', value: 'NORMAL' },
  { label: '下单中', value: 'ADDING' },
  { label: '待下单', value: 'WAITADD' },
  { label: '已退款', value: 'REFUND' },
  { label: '待退款', value: 'WAITREFUND' },
  { label: '退款失败', value: 'REFUNDFAIL' },
];

const statusMap: Record<string, { text: string; color: string }> = {
  NORMAL: { text: '正常', color: 'success' },
  ADDING: { text: '下单中', color: 'processing' },
  WAITADD: { text: '待下单', color: 'warning' },
  REFUND: { text: '已退款', color: 'default' },
  WAITREFUND: { text: '待退款', color: 'orange' },
  REFUNDFAIL: { text: '退款失败', color: 'error' },
};

// 选中项目
const selectedApp = computed(() => {
  if (!addForm.app_id) return null;
  return apps.value.find(a => a.app_id === addForm.app_id) || null;
});

// 根据 code 判断字段配置
const codeFieldConfig = computed(() => {
  const code = selectedApp.value?.code || '';
  switch (code) {
    case 'bdlp':
      return { accountLabel: '学号/UID', needPassword: false, zoneLabel: '跑区', accountField: 'uid', zoneField: 'school_name' };
    case 'yyd':
      return { accountLabel: '学号', needPassword: true, zoneLabel: '学校', accountField: 'number', zoneField: 'school_name' };
    case 'keep':
      return { accountLabel: '手机号', needPassword: true, zoneLabel: '跑区', accountField: 'phone', zoneField: 'zone_name' };
    case 'ymty':
      return { accountLabel: '手机号', needPassword: true, zoneLabel: '跑区', accountField: 'phone', zoneField: 'zone_name' };
    default:
      return { accountLabel: '账号', needPassword: true, zoneLabel: '跑区', accountField: 'phone', zoneField: 'zone_name' };
  }
});

// 预估价格
const estimatedPrice = computed(() => {
  const app = selectedApp.value;
  if (!app) return 0;
  if (app.cac_type === 'TS') {
    return Math.round(app.price * addForm.task_count * 100) / 100;
  }
  return Math.round(app.price * addForm.task_count * addForm.dis * 100) / 100;
});

// 切换项目时重置跑区
watch(() => addForm.app_id, () => {
  zones.value = [];
  addForm.zone_id = '';
  addForm.zone_name = '';
  queryMsg.value = '';
  if (addStep.value > 0) addStep.value = 0;
});

// ---------- 加载 ----------
async function loadApps() {
  try {
    const res: any = await wGetAppsApi();
    apps.value = res || [];
  } catch (e) { console.error(e); }
}

async function fetchOrders() {
  loading.value = true;
  try {
    const res: any = await wGetOrdersApi({
      page: pagination.page,
      page_size: pagination.page_size,
      account: search.account || undefined,
      status: search.status || undefined,
      app_id: search.app_id || undefined,
    });
    orders.value = res.list || [];
    total.value = res.total || 0;
  } catch (e) {
    console.error('加载订单失败', e);
  } finally {
    loading.value = false;
  }
}

// ---------- 查询跑区 ----------
async function queryZones() {
  const app = selectedApp.value;
  if (!app) { message.warning('请先选择项目'); return; }
  if (!addForm.account) { message.warning('请输入账号'); return; }
  if (codeFieldConfig.value.needPassword && !addForm.password) { message.warning('请输入密码'); return; }

  queryLoading.value = true;
  queryMsg.value = '';
  zones.value = [];

  try {
    const code = app.code;
    // 构造查询参数
    const formData: Record<string, any> = {};
    formData[codeFieldConfig.value.accountField] = addForm.account;
    if (addForm.password) formData.password = addForm.password;

    const res: any = await wProxyActionApi(app.app_id, `get_${code}_zone_data`, { form: formData });

    // res 就是上游原始响应（已被 requestClient 拦截器解包为 data 部分）
    // 上游格式通常为 {code: 1, data: [...], msg: "..."}
    const upstream = res;
    if (!upstream) {
      queryMsg.value = '上游无响应';
      addStep.value = 1;
      return;
    }

    const upCode = upstream.code ?? upstream.Code;
    const upMsg = upstream.msg ?? upstream.message ?? '';
    const upData = upstream.data ?? upstream.Data ?? [];

    if (upCode === 1 || upCode === '1' || upCode === 0 || upCode === '0') {
      // 解析跑区列表
      if (Array.isArray(upData)) {
        zones.value = upData.map((z: any) => ({
          id: String(z.id ?? z.zone_id ?? z.school_id ?? ''),
          name: String(z.name ?? z.zone_name ?? z.school_name ?? ''),
          ...z,
        }));
      }
      queryMsg.value = upMsg || '查询成功';
    } else {
      queryMsg.value = upMsg || '查询失败';
    }

    addStep.value = 1;
  } catch (e: any) {
    console.error('查询跑区失败', e);
    queryMsg.value = e?.message || '查询跑区失败，可手动输入跑区';
    addStep.value = 1;
  } finally {
    queryLoading.value = false;
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
  const app = selectedApp.value;
  if (!app) { message.warning('请选择项目'); return; }
  if (!addForm.account) { message.warning('请输入账号'); return; }
  if (addForm.task_count < 1) { message.warning('请输入次数'); return; }
  if (app.cac_type === 'KM' && addForm.dis <= 0) { message.warning('请输入公里数'); return; }
  if (!addForm.zone_name && !addForm.zone_id) { message.warning('请选择或输入跑区'); return; }

  const code = app.code;
  const cfg = codeFieldConfig.value;

  // 构造 form 数据（Jingyu 格式）
  const form: Record<string, any> = {
    dis: addForm.dis,
    task_list: addForm.task_count,
    run_type: addForm.run_type,
  };
  form[cfg.accountField] = addForm.account;
  if (addForm.password) form.password = addForm.password;
  form[cfg.zoneField] = addForm.zone_name;
  if (addForm.zone_id) form.zone_id = addForm.zone_id;

  // 构造 task_list 数组（用于 AddOrder 校验）
  const taskList = Array.from({ length: addForm.task_count }, () => ({}));

  addLoading.value = true;
  try {
    await wAddOrderApi({
      app_id: app.app_id,
      a_school: addForm.zone_name,
      a_account: addForm.account,
      a_password: addForm.password,
      dis: addForm.dis,
      task_list: taskList,
      form,
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

// ---------- 操作 ----------
function handleRefund(record: WOrder) {
  Modal.confirm({
    title: '确认退款',
    content: `确定要退款订单 #${record.id}（账号：${record.account}）吗？`,
    onOk: async () => {
      try {
        await wRefundOrderApi(record.id);
        message.success('退款成功');
        fetchOrders();
      } catch (e: any) {
        message.error(e?.message || '退款失败');
      }
    },
  });
}

async function handleSync(record: WOrder) {
  try {
    await wSyncOrderApi(record.id);
    message.success('同步成功');
    fetchOrders();
  } catch (e: any) {
    message.error(e?.message || '同步失败');
  }
}

async function handleResume(record: WOrder) {
  try {
    await wResumeOrderApi(record.id);
    message.success('重新提交成功');
    fetchOrders();
  } catch (e: any) {
    message.error(e?.message || '重新提交失败');
  }
}

function getAppName(appId: number) {
  const a = apps.value.find(x => x.app_id === appId);
  return a ? a.name : `#${appId}`;
}

function openAdd() {
  Object.assign(addForm, {
    app_id: apps.value.length > 0 ? apps.value[0]!.app_id : undefined,
    account: '', password: '', zone_id: '', zone_name: '',
    run_type: 1, dis: 2, task_count: 7,
  });
  zones.value = [];
  queryMsg.value = '';
  addStep.value = 0;
  addVisible.value = true;
}

// ---------- 表格列 ----------
const columns = [
  { title: 'ID', dataIndex: 'id', width: 70 },
  { title: '源台ID', key: 'agg_order_id', width: 85 },
  { title: '项目', key: 'app', width: 100 },
  { title: '账号', key: 'account', width: 140 },
  { title: '学校', dataIndex: 'school', width: 100, ellipsis: true },
  { title: '次数', dataIndex: 'num', width: 60 },
  { title: '金额', key: 'cost', width: 80 },
  { title: '状态', key: 'status', width: 90 },
  { title: '暂停', key: 'pause', width: 60 },
  { title: '更新时间', dataIndex: 'updated', width: 160 },
  { title: '操作', key: 'action', width: 200, fixed: 'right' as const },
];

onMounted(async () => {
  await loadApps();
  fetchOrders();
});
</script>

<template>
  <Page title="鲸鱼运动" description="管理鲸鱼运动跑步订单">
    <Card class="mb-4" :bordered="false">
      <Alert message="下单前请确认账号密码正确，跑步期间切勿登录账号！" type="warning" show-icon class="mb-4" />
      <div class="flex flex-wrap items-center gap-3">
        <Space>
          <Button type="primary" @click="openAdd" :disabled="apps.length === 0">添加订单</Button>
          <Tag v-if="apps.length === 0" color="red">暂无可用项目</Tag>
          <Tag v-else color="blue">{{ apps.length }} 个项目可用</Tag>
        </Space>
        <div class="flex-1" />
        <Space wrap>
          <Select v-model:value="search.status" style="width: 120px" @change="() => { pagination.page = 1; fetchOrders(); }">
            <SelectOption v-for="o in statusOptions" :key="o.value" :value="o.value">{{ o.label }}</SelectOption>
          </Select>
          <Select v-model:value="search.app_id" placeholder="项目筛选" allow-clear style="width: 140px" @change="() => { pagination.page = 1; fetchOrders(); }">
            <SelectOption value="">全部项目</SelectOption>
            <SelectOption v-for="a in apps" :key="a.app_id" :value="String(a.app_id)">{{ a.name }}</SelectOption>
          </Select>
          <Input.Search v-model:value="search.account" placeholder="搜索账号" style="width: 180px" @search="() => { pagination.page = 1; fetchOrders(); }" allow-clear />
        </Space>
      </div>
    </Card>

    <Card :bordered="false">
      <Table :columns="columns" :data-source="orders" :loading="loading" :pagination="{
        current: pagination.page, pageSize: pagination.page_size, total,
        showSizeChanger: true, pageSizeOptions: ['20', '50', '100'],
        showTotal: (t: number) => `共 ${t} 条`,
        onChange: (p: number, s: number) => { pagination.page = p; pagination.page_size = s; fetchOrders(); },
        onShowSizeChange: (_: number, s: number) => { pagination.page_size = s; pagination.page = 1; fetchOrders(); },
      }" row-key="id" :scroll="{ x: 1200 }" size="small">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'agg_order_id'">
            <span v-if="record.agg_order_id" class="text-xs">{{ record.agg_order_id }}</span>
            <span v-else class="text-gray-400">-</span>
          </template>
          <template v-else-if="column.key === 'app'">
            {{ getAppName(record.app_id) }}
          </template>
          <template v-else-if="column.key === 'account'">
            <div>{{ record.account }}</div>
            <div class="text-gray-400 text-xs">{{ record.password }}</div>
          </template>
          <template v-else-if="column.key === 'cost'">
            <span class="font-semibold">¥{{ Number(record.cost).toFixed(2) }}</span>
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="(statusMap[record.status] || { color: 'default' }).color">
              {{ (statusMap[record.status] || { text: record.status }).text }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'pause'">
            <Tag v-if="record.pause" color="orange">暂停</Tag>
            <span v-else class="text-gray-400">-</span>
          </template>
          <template v-else-if="column.key === 'action'">
            <Space>
              <Button size="small" @click="handleSync(record)" :disabled="!record.agg_order_id">同步</Button>
              <Button size="small" @click="handleResume(record)" v-if="record.status === 'WAITADD'">重新提交</Button>
              <Button size="small" danger @click="handleRefund(record)"
                :disabled="record.status === 'REFUND' || record.deleted">退款</Button>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <!-- 添加订单弹窗 —— 多步流程 -->
    <Modal
      v-model:open="addVisible"
      title="添加跑步订单"
      width="640px"
      :footer="null"
      :maskClosable="false"
    >
      <Steps :current="addStep" size="small" class="mt-2 mb-6">
        <Step title="填写账户" />
        <Step title="选择跑区" />
        <Step title="确认下单" />
      </Steps>

      <!-- ========== Step 0: 填写账户信息 ========== -->
      <div v-show="addStep === 0">
        <Form layout="horizontal" :label-col="{ span: 5 }">
          <FormItem label="选择项目">
            <Select v-model:value="addForm.app_id" placeholder="请选择项目" style="width: 100%">
              <SelectOption v-for="a in apps" :key="a.app_id" :value="a.app_id">
                {{ a.name }} — ¥{{ a.price }}/{{ a.cac_type === 'TS' ? '次' : 'km' }}
              </SelectOption>
            </Select>
            <div v-if="selectedApp?.description" class="text-gray-400 text-xs mt-1">{{ selectedApp.description }}</div>
          </FormItem>
          <FormItem :label="codeFieldConfig.accountLabel">
            <Input v-model:value="addForm.account" :placeholder="`请输入${codeFieldConfig.accountLabel}`" />
          </FormItem>
          <FormItem v-if="codeFieldConfig.needPassword" label="密码">
            <Input.Password v-model:value="addForm.password" placeholder="账号密码" />
          </FormItem>
        </Form>
        <div class="flex justify-end gap-2 mt-4">
          <Button @click="addVisible = false">取消</Button>
          <Button type="primary" :loading="queryLoading" @click="queryZones">
            查询跑区
          </Button>
        </div>
      </div>

      <!-- ========== Step 1: 选择跑区 & 参数 ========== -->
      <div v-show="addStep === 1">
        <Spin :spinning="queryLoading">
          <Form layout="horizontal" :label-col="{ span: 5 }">
            <FormItem :label="codeFieldConfig.zoneLabel">
              <Select
                v-if="zones.length > 0"
                v-model:value="addForm.zone_id"
                :placeholder="`请选择${codeFieldConfig.zoneLabel}`"
                style="width: 100%"
                show-search
                :filter-option="(input: string, option: any) => (option?.label ?? '').toLowerCase().includes(input.toLowerCase())"
                @change="onZoneSelect"
              >
                <SelectOption v-for="z in zones" :key="z.id" :value="z.id" :label="z.name">
                  {{ z.name }}
                </SelectOption>
              </Select>
              <Input
                v-else
                v-model:value="addForm.zone_name"
                :placeholder="`手动输入${codeFieldConfig.zoneLabel}名称`"
              />
              <div v-if="queryMsg" class="text-xs mt-1" :class="zones.length > 0 ? 'text-green-500' : 'text-orange-500'">
                {{ queryMsg }}
                <span v-if="zones.length > 0">（共 {{ zones.length }} 个{{ codeFieldConfig.zoneLabel }}）</span>
              </div>
            </FormItem>

            <FormItem label="跑步类型">
              <RadioGroup v-model:value="addForm.run_type">
                <Radio :value="1">有效跑</Radio>
                <Radio :value="2">自由跑</Radio>
              </RadioGroup>
            </FormItem>

            <FormItem label="次数">
              <InputNumber v-model:value="addForm.task_count" :min="1" :max="365" :step="1" style="width: 100%" />
            </FormItem>

            <FormItem v-if="selectedApp && selectedApp.cac_type === 'KM'" label="每次公里数">
              <InputNumber v-model:value="addForm.dis" :min="0.5" :max="100" :step="0.5" :precision="1" style="width: 100%" />
            </FormItem>

            <FormItem label="预估费用">
              <span style="color: red; font-weight: 800; font-size: 18px">¥{{ estimatedPrice.toFixed(2) }}</span>
              <span v-if="selectedApp" class="text-gray-400 ml-2">
                ({{ selectedApp.price }}元/{{ selectedApp.cac_type === 'TS' ? '次' : 'km' }}
                × {{ addForm.task_count }}次
                <template v-if="selectedApp.cac_type === 'KM'">× {{ addForm.dis }}km</template>)
              </span>
            </FormItem>
          </Form>
        </Spin>
        <div class="flex justify-end gap-2 mt-4">
          <Button @click="addStep = 0">上一步</Button>
          <Button type="primary" @click="addStep = 2" :disabled="!addForm.zone_id && !addForm.zone_name">
            下一步
          </Button>
        </div>
      </div>

      <!-- ========== Step 2: 确认下单 ========== -->
      <div v-show="addStep === 2">
        <Descriptions bordered :column="1" size="small" class="mb-4">
          <DescriptionsItem label="项目">{{ selectedApp?.name }}</DescriptionsItem>
          <DescriptionsItem :label="codeFieldConfig.accountLabel">{{ addForm.account }}</DescriptionsItem>
          <DescriptionsItem v-if="codeFieldConfig.needPassword" label="密码">{{ addForm.password ? '******' : '-' }}</DescriptionsItem>
          <DescriptionsItem :label="codeFieldConfig.zoneLabel">{{ addForm.zone_name || addForm.zone_id }}</DescriptionsItem>
          <DescriptionsItem label="跑步类型">{{ addForm.run_type === 1 ? '有效跑' : '自由跑' }}</DescriptionsItem>
          <DescriptionsItem label="次数">{{ addForm.task_count }} 次</DescriptionsItem>
          <DescriptionsItem v-if="selectedApp?.cac_type === 'KM'" label="每次公里数">{{ addForm.dis }} km</DescriptionsItem>
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
  </Page>
</template>

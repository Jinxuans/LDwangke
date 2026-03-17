<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import {
  Card, Table, Button, Tag, Space, Input, Select, SelectOption,
  Modal, message, Popconfirm, Form, FormItem, Switch, Tabs, TabPane,
  Tooltip, Typography, Collapse, CollapsePanel, Badge, Alert,
} from 'ant-design-vue';
import {
  PlusOutlined, ReloadOutlined, EditOutlined, DeleteOutlined,
  CodeOutlined, ImportOutlined, ThunderboltOutlined, CheckCircleOutlined,
  WarningOutlined, SearchOutlined, RadarChartOutlined,
} from '@ant-design/icons-vue';
import {
  getPlatformConfigsApi, savePlatformConfigApi,
  deletePlatformConfigApi, parsePHPCodeApi, detectPlatformApi, autoDetectSaveApi,
  type PlatformConfig, type ParsedPHPResult, type DetectResult,
} from '#/api/platform-config';

const loading = ref(false);
const tableData = ref<PlatformConfig[]>([]);
const searchText = ref('');

const editVisible = ref(false);
const editLoading = ref(false);
const editForm = ref<Partial<PlatformConfig>>({});
const isEdit = ref(false);

const phpVisible = ref(false);
const phpLoading = ref(false);
const phpCode = ref('');
const phpResult = ref<ParsedPHPResult | null>(null);

const detectVisible = ref(false);
const detectLoading = ref(false);
const detectSaveLoading = ref(false);
const detectForm = ref({ url: '', uid: '', key: '', token: '', cookie: '', pt: '', name: '' });
const detectResult = ref<DetectResult | null>(null);
const customQueryDrivers = new Set(['local_time', 'local_script', 'xxt_query', 'KUN_custom', 'simple_custom', 'yyy_custom', 'tuboshu_custom', 'nx_custom', 'lgwk_custom']);

function formatJsonTemplate(raw: string) {
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}

const actionParamHelp = '接口配置现在只认显式命名空间模板变量。通用变量：{{supplier.uid}} {{supplier.key}} {{supplier.token}} {{order.yid}}。其余字段按动作区块使用 {{order.*}} 或 {{action.*}}，不再支持旧的裸字段和隐式映射。';
const queryParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","school":"{{action.school}}","user":"{{action.user}}","pass":"{{action.password}}","platform":"{{action.platform}}"}');
const queryParamHelp = `查课动作变量：{{action.school}} {{action.user}} {{action.password}} {{action.platform}}。默认模板：${queryParamMapExample}`;
const orderParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","platform":"{{order.noun}}","school":"{{order.school}}","user":"{{order.user}}","pass":"{{order.pass}}","kcid":"{{order.kcid}}","kcname":"{{order.kcname}}"}');
const orderParamHelp = `下单对接发生在本地订单入库之后，所以上游下单映射应使用订单字段：{{order.noun}} {{order.school}} {{order.user}} {{order.pass}} {{order.kcid}} {{order.kcname}}。默认模板：${orderParamMapExample}`;
const progressParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","username":"{{order.user}}","kcname":"{{order.kcname}}","cid":"{{order.noun}}","yid":"{{order.yid}}"}');
const batchProgressParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}"}');
const batchProgressParamHelp = '批量进度接口是供应商级增量拉取，只用于自动同步，不会逐单携带订单参数主动请求。常用模板只需要 {{supplier.uid}} {{supplier.key}}；如果上游要求批次上下文，也可以使用 {{batch.count}} {{batch.yids}} {{batch.users}}。';
const categoryParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}"}');
const classListParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}"}');
const pauseParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}"}');
const resumeParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}"}');
const changePassParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}","newPwd":"{{action.new_password}}"}');
const changePassParamHelp = `可用变量：{{supplier.uid}} {{supplier.key}} {{order.yid}} {{action.new_password}}。默认模板：${changePassParamMapExample}`;
const resubmitParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}"}');
const logParamMapExample = formatJsonTemplate('{"dtoken":"{{supplier.token}}","account":"{{order.user}}","password":"{{order.pass}}","course":"{{order.kcname}}","courseId":"{{order.kcid}}"}');
const balanceParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}"}');
const reportParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}","question":"{{action.content}}"}');
const reportParamHelp = `提工单动作变量：{{action.content}} {{action.ticket_type}}，订单标识仍用 {{order.yid}}。标准模板：${reportParamMapExample}`;
const getReportParamMapExample = formatJsonTemplate('{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","reportId":"{{action.report_id}}"}');
const getReportParamHelp = `查询工单结果变量：{{action.report_id}}。默认模板：${getReportParamMapExample}`;

const filteredData = computed(() => {
  if (!searchText.value) return tableData.value;
  const s = searchText.value.toLowerCase();
  return tableData.value.filter(
    (c) => c.pt.toLowerCase().includes(s) || c.name.toLowerCase().includes(s),
  );
});

const columns = [
  { title: '平台标识', dataIndex: 'pt', width: 100, fixed: 'left' as const },
  { title: '名称', dataIndex: 'name', width: 100 },
  { title: '认证', dataIndex: 'auth_type', width: 80 },
  { title: '成功码', dataIndex: 'success_codes', width: 80 },
  { title: '查课接口', key: 'query_endpoint', width: 180 },
  { title: '下单接口', key: 'order_endpoint', width: 180 },
  { title: '进度接口', key: 'progress_endpoint', width: 180 },
  { title: '课程列表', key: 'class_list_endpoint', width: 180 },
  { title: '特性', key: 'features', width: 180 },
  { title: '操作', key: 'action', width: 120, fixed: 'right' as const },
];

function firstPath(path?: string, fallbackPath = '') {
  const normalizedPath = path?.trim();
  if (normalizedPath) return normalizedPath;
  if (fallbackPath.trim()) {
    return fallbackPath.trim();
  }
  return '';
}

function formatEndpoint(path?: string, queryDriver?: string) {
  const normalizedPath = firstPath(path);
  if (normalizedPath) return normalizedPath;
  if (queryDriver && customQueryDrivers.has(queryDriver.trim())) {
    return `专用驱动: ${queryDriver.trim()}`;
  }
  return '-';
}

async function loadData() {
  loading.value = true;
  try {
    const res = await getPlatformConfigsApi();
    tableData.value = res ?? [];
  } catch (e: any) {
    message.error('加载失败: ' + (e?.message || '未知错误'));
  } finally {
    loading.value = false;
  }
}

function openAdd() {
  isEdit.value = false;
  editForm.value = {
    auth_type: 'uid_key', success_codes: '0',
    query_path: '', query_method: '', query_body_type: '', query_param_map: '',
    order_path: '', order_method: '', order_body_type: '', order_param_map: '',
    progress_path: '', progress_method: '', progress_body_type: '', progress_param_map: '',
    batch_progress_path: '', batch_progress_method: '', batch_progress_body_type: '', batch_progress_param_map: '',
    category_path: '', category_method: '', category_body_type: '', category_param_map: '',
    class_list_path: '', class_list_method: '', class_list_body_type: '', class_list_param_map: '',
    pause_path: '', pause_method: '', pause_body_type: '', pause_param_map: '',
    pause_id_param: '', resume_path: '', resume_method: '', resume_body_type: '', resume_param_map: '',
    change_pass_path: '', change_pass_method: '', change_pass_body_type: '', change_pass_param_map: '', change_pass_param: '', change_pass_id_param: '',
    resubmit_method: '', resubmit_body_type: '', resubmit_param_map: '',
    resubmit_path: '', resubmit_id_param: '',
    log_path: '', log_method: '', log_body_type: '', log_param_map: '', log_id_param: '',
    balance_path: '', balance_money_field: '', balance_method: '', balance_body_type: '', balance_param_map: '', balance_auth_type: '',
    report_path: '', report_method: '', report_body_type: '', report_param_map: '',
    get_report_path: '', get_report_method: '', get_report_body_type: '', get_report_param_map: '',
  };
  editVisible.value = true;
}

function openEdit(row: PlatformConfig) {
  isEdit.value = true;
  editForm.value = {
    ...row,
    progress_param_map: row.progress_param_map || '',
    batch_progress_param_map: row.batch_progress_param_map || '',
  };
  editVisible.value = true;
}

async function handleSave() {
  if (!editForm.value.pt) { message.warning('请填写平台标识'); return; }
  editLoading.value = true;
  try {
    await savePlatformConfigApi(editForm.value);
    message.success('保存成功');
    editVisible.value = false;
    loadData();
  } catch (e: any) { message.error('保存失败: ' + (e?.message || '未知错误')); }
  finally { editLoading.value = false; }
}

async function handleDelete(pt: string) {
  try {
    await deletePlatformConfigApi(pt);
    message.success('删除成功');
    loadData();
  } catch (e: any) { message.error('删除失败: ' + (e?.message || '未知错误')); }
}

function openPHPImport() {
  phpCode.value = ''; phpResult.value = null; phpVisible.value = true;
}

async function parsePHP() {
  if (!phpCode.value.trim()) { message.warning('请粘贴 PHP 代码'); return; }
  phpLoading.value = true;
  try {
    const res = await parsePHPCodeApi(phpCode.value);
    phpResult.value = res;
    if (phpResult.value && phpResult.value.confidence >= 30) {
      message.success(`解析成功，置信度 ${phpResult.value.confidence}%`);
    } else { message.warning('解析置信度较低，建议手动核对'); }
  } catch (e: any) { message.error('解析失败: ' + (e?.message || '未知错误')); }
  finally { phpLoading.value = false; }
}

function applyPHPResult() {
  if (!phpResult.value) return;
  const r = phpResult.value;
  editForm.value = {
    pt: r.pt, name: r.name, auth_type: r.auth_type || 'uid_key',
    success_codes: r.success_codes || '0',
    use_json: r.use_json, query_path: firstPath(r.query_path, '/api.php?act=get'), query_method: 'POST', query_body_type: '', query_param_map: '',
    order_path: firstPath(r.order_path, '/api.php?act=add'), order_method: 'POST', order_body_type: '', order_param_map: '',
    progress_path: firstPath(r.progress_path, '/api.php?act=chadan2'),
    progress_method: r.progress_method || 'POST',
    progress_body_type: '',
    progress_param_map: '',
    batch_progress_path: '',
    batch_progress_method: '',
    batch_progress_body_type: '',
    batch_progress_param_map: '',
    category_path: '/api.php?act=getcate', category_method: 'POST', category_body_type: '', category_param_map: '',
    class_list_path: '/api.php?act=getclass', class_list_method: 'POST', class_list_body_type: '', class_list_param_map: '',
    pause_path: firstPath(r.pause_path, '/api.php?act=zt'), pause_method: 'POST', pause_body_type: '', pause_param_map: '',
    pause_id_param: r.pause_id_param || 'id',
    resume_method: 'POST', resume_body_type: '', resume_param_map: '',
    change_pass_path: firstPath(r.change_pass_path, '/api.php?act=gaimi'), change_pass_method: 'POST', change_pass_body_type: '', change_pass_param_map: '', change_pass_param: r.change_pass_param || 'newPwd',
    change_pass_id_param: r.change_pass_id_param || 'id',
    resubmit_path: '/api.php?act=budan', resubmit_method: 'POST', resubmit_body_type: '', resubmit_param_map: '', resubmit_id_param: 'id',
    log_path: firstPath(r.log_path, '/api.php?act=xq'), log_body_type: '', log_param_map: '', log_id_param: r.log_id_param || 'id',
    returns_yid: r.returns_yid, log_method: 'POST',
    balance_path: firstPath(r.balance_path, '/api.php?act=getmoney'),
    balance_money_field: r.balance_money_field || 'money', balance_method: 'POST', balance_body_type: '', balance_param_map: '',
    balance_auth_type: '', report_method: 'POST', report_body_type: '', report_param_map: '',
    get_report_method: 'POST', get_report_body_type: '', get_report_param_map: '', source_code: phpCode.value,
  };
  isEdit.value = false; phpVisible.value = false; editVisible.value = true;
  message.info('已填充解析结果，请核对后保存');
}

function openDetect() {
  detectForm.value = { url: '', uid: '', key: '', token: '', cookie: '', pt: '', name: '' };
  detectResult.value = null; detectVisible.value = true;
}

async function runDetect() {
  if (!detectForm.value.url) { message.warning('请填写平台 URL'); return; }
  detectLoading.value = true;
  try {
    const res = await detectPlatformApi(detectForm.value);
    detectResult.value = res;
    if (detectResult.value?.success) {
      message.success(`检测成功！余额: ${detectResult.value.balance_money || '未知'}`);
    } else { message.warning('未检测到可用接口，请检查 URL 和凭证'); }
  } catch (e: any) { message.error('检测失败: ' + (e?.message || '未知错误')); }
  finally { detectLoading.value = false; }
}

async function autoDetectSave() {
  if (!detectForm.value.url) { message.warning('请填写平台 URL'); return; }
  if (!detectForm.value.pt) { message.warning('请填写平台标识'); return; }
  detectSaveLoading.value = true;
  try {
    const res = await autoDetectSaveApi({
      url: detectForm.value.url, uid: detectForm.value.uid, key: detectForm.value.key,
      token: detectForm.value.token, cookie: detectForm.value.cookie,
      pt: detectForm.value.pt, name: detectForm.value.name,
    });
    detectResult.value = res.detect;
    if (res.success) {
      message.success(res.msg);
      loadData();
    } else {
      message.warning(res.msg);
    }
  } catch (e: any) { message.error('操作失败: ' + (e?.message || '未知错误')); }
  finally { detectSaveLoading.value = false; }
}

function applyDetectResult() {
  if (!detectResult.value) return;
  const r = detectResult.value;
  const cfg = r.config || {};
  editForm.value = {
    pt: detectForm.value.pt || '', name: detectForm.value.name || r.suggested_name || '', auth_type: cfg.auth_type || 'uid_key',
    success_codes: cfg.success_codes || '0',
    use_json: cfg.use_json === 'true', query_path: firstPath(cfg.query_path, '/api.php?act=get'), query_method: cfg.query_method || 'POST', query_body_type: cfg.query_body_type || '', query_param_map: cfg.query_param_map || '',
    order_path: firstPath(cfg.order_path, '/api.php?act=add'), order_method: cfg.order_method || 'POST', order_body_type: cfg.order_body_type || '', order_param_map: cfg.order_param_map || '',
    progress_path: firstPath(cfg.progress_path, '/api.php?act=chadan2'),
    progress_method: cfg.progress_method || 'POST',
    progress_body_type: cfg.progress_body_type || '',
    progress_param_map: cfg.progress_param_map || '',
    batch_progress_path: firstPath(cfg.batch_progress_path),
    batch_progress_method: cfg.batch_progress_method || '',
    batch_progress_body_type: cfg.batch_progress_body_type || '',
    batch_progress_param_map: cfg.batch_progress_param_map || '',
    category_path: firstPath(cfg.category_path, '/api.php?act=getcate'), category_method: cfg.category_method || 'POST', category_body_type: cfg.category_body_type || '', category_param_map: cfg.category_param_map || '',
    class_list_path: firstPath(cfg.class_list_path, '/api.php?act=getclass'), class_list_method: cfg.class_list_method || 'POST', class_list_body_type: cfg.class_list_body_type || '', class_list_param_map: cfg.class_list_param_map || '',
    pause_path: firstPath(cfg.pause_path, '/api.php?act=zt'), pause_method: cfg.pause_method || 'POST', pause_body_type: cfg.pause_body_type || '', pause_param_map: cfg.pause_param_map || '',
    pause_id_param: cfg.pause_id_param || 'id', resume_method: cfg.resume_method || 'POST', resume_body_type: cfg.resume_body_type || '', resume_param_map: cfg.resume_param_map || '', change_pass_path: firstPath(cfg.change_pass_path, '/api.php?act=gaimi'),
    change_pass_method: cfg.change_pass_method || 'POST', change_pass_body_type: cfg.change_pass_body_type || '', change_pass_param_map: cfg.change_pass_param_map || '', change_pass_param: 'newPwd', change_pass_id_param: 'id',
    resubmit_path: firstPath(cfg.resubmit_path, '/api.php?act=budan'), resubmit_method: cfg.resubmit_method || 'POST', resubmit_body_type: cfg.resubmit_body_type || '', resubmit_param_map: cfg.resubmit_param_map || '', resubmit_id_param: cfg.resubmit_id_param || 'id', refresh_path: cfg.refresh_path || '',
    log_path: firstPath(cfg.log_path, '/api.php?act=xq'),
    log_method: cfg.log_method || 'POST', log_body_type: cfg.log_body_type || '', log_param_map: cfg.log_param_map || '', log_id_param: 'id',
    balance_path: firstPath(cfg.balance_path, '/api.php?act=getmoney'),
    balance_money_field: cfg.balance_money_field || 'money', balance_method: 'POST', balance_body_type: cfg.balance_body_type || '', balance_param_map: cfg.balance_param_map || '',
    balance_auth_type: cfg.auth_type === 'bearer_token' ? 'bearer_token' : '',
    report_path: cfg.report_path || '', report_method: cfg.report_method || 'POST', report_body_type: cfg.report_body_type || '', report_param_map: cfg.report_param_map || '', get_report_path: cfg.get_report_path || '',
    get_report_method: cfg.get_report_method || 'POST', get_report_body_type: cfg.get_report_body_type || '', get_report_param_map: cfg.get_report_param_map || '', report_param_style: cfg.report_param_style || '', report_auth_type: cfg.report_auth_type || '',
  };
  isEdit.value = false; detectVisible.value = false; editVisible.value = true;
  message.info('已填充检测结果，请核对后保存');
}

const statusColor: Record<string, string> = {
  ok: 'green', fail: 'red', error: 'orange', timeout: 'orange',
};

onMounted(loadData);
</script>

<template>
  <div>
    <div class="mb-4 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <Badge :count="tableData.length" :number-style="{ backgroundColor: '#1890ff' }" />
        <span class="text-sm text-gray-500">个平台配置</span>
      </div>
      <Space>
        <Input.Search v-model:value="searchText" placeholder="搜索平台" style="width: 200px" allow-clear />
        <Button type="primary" @click="openDetect">
          <template #icon><RadarChartOutlined /></template>
          自动检测
        </Button>
        <Button type="primary" @click="openPHPImport">
          <template #icon><CodeOutlined /></template>
          PHP 导入
        </Button>
        <Button type="primary" @click="openAdd">
          <template #icon><PlusOutlined /></template>
          手动添加
        </Button>
        <Button @click="loadData">
          <template #icon><ReloadOutlined /></template>
        </Button>
      </Space>
    </div>

    <Table
      :columns="columns" :data-source="filteredData" :loading="loading"
      :pagination="{ pageSize: 50 }" :scroll="{ x: 1100 }" row-key="pt" size="small"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.dataIndex === 'pt'">
          <Tag color="blue">{{ record.pt }}</Tag>
        </template>
        <template v-else-if="column.dataIndex === 'auth_type'">
          <Tag :color="record.auth_type === 'uid_key' ? 'default' : 'orange'">{{ record.auth_type }}</Tag>
        </template>
        <template v-else-if="column.key === 'query_endpoint'">
          <span>{{ formatEndpoint(record.query_path, record.query_act) }}</span>
        </template>
        <template v-else-if="column.key === 'order_endpoint'">
          <span>{{ formatEndpoint(record.order_path) }}</span>
        </template>
        <template v-else-if="column.key === 'progress_endpoint'">
          <span>{{ formatEndpoint(record.progress_path) }}</span>
        </template>
        <template v-else-if="column.key === 'class_list_endpoint'">
          <span>{{ formatEndpoint(record.class_list_path) }}</span>
        </template>
        <template v-else-if="column.key === 'features'">
          <Space :size="2" wrap>
            <Tag v-if="record.use_json" color="purple" :bordered="false">JSON</Tag>
            <Tag v-if="record.need_proxy" color="red" :bordered="false">代理</Tag>
            <Tag v-if="record.returns_yid" color="green" :bordered="false">返YID</Tag>
            <Tag v-if="record.progress_path" color="cyan" :bordered="false">自定义进度路径</Tag>
            <Tag v-if="record.batch_progress_path" color="geekblue" :bordered="false">批量进度</Tag>
            <Tag v-if="record.pause_path" color="cyan" :bordered="false">自定义暂停路径</Tag>
            <Tag v-if="record.change_pass_path" color="cyan" :bordered="false">自定义改密路径</Tag>
            <Tag v-if="record.extra_params" color="gold" :bordered="false">额外参数</Tag>
          </Space>
        </template>
        <template v-else-if="column.key === 'action'">
          <Space>
            <Tooltip title="编辑">
              <Button type="link" size="small" @click="openEdit(record as PlatformConfig)"><EditOutlined /></Button>
            </Tooltip>
            <Popconfirm :title="`确定删除 ${record.pt}？`" @confirm="handleDelete(record.pt)">
              <Button type="link" size="small" danger><DeleteOutlined /></Button>
            </Popconfirm>
          </Space>
        </template>
      </template>
    </Table>

    <!-- 编辑弹窗 -->
    <Modal v-model:open="editVisible" :title="isEdit ? `编辑平台：${editForm.pt}` : '添加平台配置'" :confirm-loading="editLoading" width="700px" @ok="handleSave">
      <Form :label-col="{ span: 6 }" :wrapper-col="{ span: 18 }" size="small" class="mt-4">
        <Typography.Paragraph type="secondary" class="mb-3 text-xs">
          {{ actionParamHelp }}
        </Typography.Paragraph>
        <Tabs size="small">
          <TabPane key="basic" tab="基本">
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="平台标识" required><Input v-model:value="editForm.pt" :disabled="isEdit" placeholder="如 newplat" /></FormItem>
              <FormItem label="名称"><Input v-model:value="editForm.name" placeholder="如 新平台" /></FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="认证方式">
                <Select v-model:value="editForm.auth_type">
                  <SelectOption value="uid_key">uid + key</SelectOption>
                  <SelectOption value="token_only">token（pass字段）</SelectOption>
                  <SelectOption value="token_field">token（token字段）</SelectOption>
                  <SelectOption value="none">无认证</SelectOption>
                </Select>
              </FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="成功码"><Input v-model:value="editForm.success_codes" placeholder="0,1,200" /></FormItem>
              <FormItem label="JSON请求"><Switch v-model:checked="editForm.use_json" /></FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="需要代理"><Switch v-model:checked="editForm.need_proxy" /></FormItem>
              <FormItem label="返回YID"><Switch v-model:checked="editForm.returns_yid" /></FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="额外参数"><Switch v-model:checked="editForm.extra_params" /></FormItem>
              <FormItem label="YID在数组"><Switch v-model:checked="editForm.yid_in_data_array" /></FormItem>
            </div>
          </TabPane>
          <TabPane key="query" tab="查课">
            <FormItem label="查课路径"><Input v-model:value="editForm.query_path" placeholder="如 /api/query-course 或 api.php?act=get" /></FormItem>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="请求方式">
                <Select v-model:value="editForm.query_method" allow-clear placeholder="留空=未配置">
                  <SelectOption value="">未配置</SelectOption>
                  <SelectOption value="POST">POST</SelectOption>
                  <SelectOption value="GET">GET</SelectOption>
                  <SelectOption value="PUT">PUT</SelectOption>
                </Select>
              </FormItem>
              <FormItem label="Body类型">
                <Select v-model:value="editForm.query_body_type" allow-clear placeholder="留空=自动">
                  <SelectOption value="">自动</SelectOption>
                  <SelectOption value="form">form</SelectOption>
                  <SelectOption value="json">json</SelectOption>
                  <SelectOption value="query">query</SelectOption>
                </Select>
              </FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="兼容参数风格"><Input v-model:value="editForm.query_param_style" placeholder="standard" /></FormItem>
              <FormItem label="需要轮询"><Switch v-model:checked="editForm.query_polling" /></FormItem>
            </div>
            <Alert class="mb-3" type="info" show-icon :message="queryParamHelp" />
            <FormItem label="参数映射JSON">
              <Input.TextArea v-model:value="editForm.query_param_map" :rows="4" :placeholder="queryParamMapExample" />
            </FormItem>
          </TabPane>
          <TabPane key="order" tab="下单">
            <FormItem label="下单路径"><Input v-model:value="editForm.order_path" placeholder="如 /api/order/create 或 api.php?act=add" /></FormItem>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="请求方式">
                <Select v-model:value="editForm.order_method" allow-clear placeholder="留空=未配置">
                  <SelectOption value="">未配置</SelectOption>
                  <SelectOption value="POST">POST</SelectOption>
                  <SelectOption value="GET">GET</SelectOption>
                  <SelectOption value="PUT">PUT</SelectOption>
                </Select>
              </FormItem>
              <FormItem label="Body类型">
                <Select v-model:value="editForm.order_body_type" allow-clear placeholder="留空=自动">
                  <SelectOption value="">自动</SelectOption>
                  <SelectOption value="form">form</SelectOption>
                  <SelectOption value="json">json</SelectOption>
                  <SelectOption value="query">query</SelectOption>
                </Select>
              </FormItem>
            </div>
            <Alert class="mb-3" type="info" show-icon :message="orderParamHelp" />
            <FormItem label="参数映射JSON">
              <Input.TextArea v-model:value="editForm.order_param_map" :rows="4" :placeholder="orderParamMapExample" />
            </FormItem>
          </TabPane>
          <TabPane key="progress" tab="进度">
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="进度路径"><Input v-model:value="editForm.progress_path" placeholder="如 /api/search 或 api.php?act=chadan2" /></FormItem>
              <FormItem label="请求方式">
                <Select v-model:value="editForm.progress_method" allow-clear placeholder="留空=未配置">
                  <SelectOption value="">未配置</SelectOption>
                  <SelectOption value="POST">POST</SelectOption>
                  <SelectOption value="GET">GET</SelectOption>
                </Select>
              </FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="Body类型">
                <Select v-model:value="editForm.progress_body_type" allow-clear placeholder="留空=自动">
                  <SelectOption value="">自动</SelectOption>
                  <SelectOption value="form">form</SelectOption>
                  <SelectOption value="json">json</SelectOption>
                  <SelectOption value="query">query</SelectOption>
                </Select>
              </FormItem>
            </div>
            <FormItem label="参数映射JSON">
              <Input.TextArea v-model:value="editForm.progress_param_map" :rows="4" :placeholder="progressParamMapExample" />
            </FormItem>
            <Typography.Text type="secondary" class="mb-2 mt-3 block">批量进度（自动同步增量拉取）</Typography.Text>
            <Alert class="mb-3" type="info" show-icon :message="batchProgressParamHelp" />
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="批量进度路径"><Input v-model:value="editForm.batch_progress_path" placeholder="如 /api.php?act=pljd" /></FormItem>
              <FormItem label="请求方式">
                <Select v-model:value="editForm.batch_progress_method" allow-clear placeholder="留空=未配置">
                  <SelectOption value="">未配置</SelectOption>
                  <SelectOption value="POST">POST</SelectOption>
                  <SelectOption value="GET">GET</SelectOption>
                </Select>
              </FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="Body类型">
                <Select v-model:value="editForm.batch_progress_body_type" allow-clear placeholder="留空=自动">
                  <SelectOption value="">自动</SelectOption>
                  <SelectOption value="form">form</SelectOption>
                  <SelectOption value="json">json</SelectOption>
                  <SelectOption value="query">query</SelectOption>
                </Select>
              </FormItem>
            </div>
            <FormItem label="参数映射JSON">
              <Input.TextArea v-model:value="editForm.batch_progress_param_map" :rows="3" :placeholder="batchProgressParamMapExample" />
            </FormItem>
          </TabPane>
          <TabPane key="catalog" tab="分类/课程列表">
            <Typography.Text type="secondary" class="mb-2 block">分类接口</Typography.Text>
            <FormItem label="分类路径"><Input v-model:value="editForm.category_path" placeholder="如 /api/categories 或 api.php?act=getcate" /></FormItem>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="请求方式">
                <Select v-model:value="editForm.category_method" allow-clear placeholder="留空=未配置">
                  <SelectOption value="">未配置</SelectOption>
                  <SelectOption value="POST">POST</SelectOption>
                  <SelectOption value="GET">GET</SelectOption>
                  <SelectOption value="PUT">PUT</SelectOption>
                </Select>
              </FormItem>
              <FormItem label="Body类型">
                <Select v-model:value="editForm.category_body_type" allow-clear placeholder="留空=自动">
                  <SelectOption value="">自动</SelectOption>
                  <SelectOption value="form">form</SelectOption>
                  <SelectOption value="json">json</SelectOption>
                  <SelectOption value="query">query</SelectOption>
                </Select>
              </FormItem>
            </div>
            <FormItem label="参数映射JSON">
              <Input.TextArea v-model:value="editForm.category_param_map" :rows="3" placeholder='{"uid":"{{supplier.uid}}","key":"{{supplier.key}}"}' />
            </FormItem>
            <Typography.Text type="secondary" class="mb-2 mt-3 block">课程列表接口</Typography.Text>
            <FormItem label="课程列表路径"><Input v-model:value="editForm.class_list_path" placeholder="如 /api/getclass 或 api.php?act=getclass" /></FormItem>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="请求方式">
                <Select v-model:value="editForm.class_list_method" allow-clear placeholder="留空=未配置">
                  <SelectOption value="">未配置</SelectOption>
                  <SelectOption value="POST">POST</SelectOption>
                  <SelectOption value="GET">GET</SelectOption>
                  <SelectOption value="PUT">PUT</SelectOption>
                </Select>
              </FormItem>
              <FormItem label="Body类型">
                <Select v-model:value="editForm.class_list_body_type" allow-clear placeholder="留空=自动">
                  <SelectOption value="">自动</SelectOption>
                  <SelectOption value="form">form</SelectOption>
                  <SelectOption value="json">json</SelectOption>
                  <SelectOption value="query">query</SelectOption>
                </Select>
              </FormItem>
            </div>
            <FormItem label="参数映射JSON">
              <Input.TextArea v-model:value="editForm.class_list_param_map" :rows="3" placeholder='{"uid":"{{supplier.uid}}","key":"{{supplier.key}}"}' />
            </FormItem>
          </TabPane>
          <TabPane key="other" tab="暂停/改密/日志">
            <Typography.Text type="secondary" class="mb-2 block">暂停/恢复</Typography.Text>
            <FormItem label="暂停路径"><Input v-model:value="editForm.pause_path" placeholder="如 /api/stop 或 api.php?act=zt" /></FormItem>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="暂停方式">
                <Select v-model:value="editForm.pause_method" allow-clear placeholder="留空=未配置">
                  <SelectOption value="">未配置</SelectOption>
                  <SelectOption value="POST">POST</SelectOption>
                  <SelectOption value="GET">GET</SelectOption>
                  <SelectOption value="PUT">PUT</SelectOption>
                </Select>
              </FormItem>
              <FormItem label="暂停Body">
                <Select v-model:value="editForm.pause_body_type" allow-clear placeholder="留空=自动">
                  <SelectOption value="">自动</SelectOption>
                  <SelectOption value="form">form</SelectOption>
                  <SelectOption value="json">json</SelectOption>
                  <SelectOption value="query">query</SelectOption>
                </Select>
              </FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="暂停ID参数"><Input v-model:value="editForm.pause_id_param" placeholder="id" /></FormItem>
              <FormItem label="恢复路径"><Input v-model:value="editForm.resume_path" /></FormItem>
            </div>
            <FormItem label="暂停参数映射">
              <Input.TextArea v-model:value="editForm.pause_param_map" :rows="3" placeholder='{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}"}' />
            </FormItem>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="恢复方式">
                <Select v-model:value="editForm.resume_method" allow-clear placeholder="留空=未配置">
                  <SelectOption value="">未配置</SelectOption>
                  <SelectOption value="POST">POST</SelectOption>
                  <SelectOption value="GET">GET</SelectOption>
                  <SelectOption value="PUT">PUT</SelectOption>
                </Select>
              </FormItem>
              <FormItem label="恢复Body">
                <Select v-model:value="editForm.resume_body_type" allow-clear placeholder="留空=自动">
                  <SelectOption value="">自动</SelectOption>
                  <SelectOption value="form">form</SelectOption>
                  <SelectOption value="json">json</SelectOption>
                  <SelectOption value="query">query</SelectOption>
                </Select>
              </FormItem>
            </div>
            <FormItem label="恢复参数映射">
              <Input.TextArea v-model:value="editForm.resume_param_map" :rows="3" placeholder='{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}"}' />
            </FormItem>
            <Typography.Text type="secondary" class="mb-2 mt-3 block">改密码</Typography.Text>
            <FormItem label="改密路径"><Input v-model:value="editForm.change_pass_path" placeholder="如 /api/update 或 api.php?act=gaimi" /></FormItem>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="改密方式">
                <Select v-model:value="editForm.change_pass_method" allow-clear placeholder="留空=未配置">
                  <SelectOption value="">未配置</SelectOption>
                  <SelectOption value="POST">POST</SelectOption>
                  <SelectOption value="GET">GET</SelectOption>
                  <SelectOption value="PUT">PUT</SelectOption>
                </Select>
              </FormItem>
              <FormItem label="改密Body">
                <Select v-model:value="editForm.change_pass_body_type" allow-clear placeholder="留空=自动">
                  <SelectOption value="">自动</SelectOption>
                  <SelectOption value="form">form</SelectOption>
                  <SelectOption value="json">json</SelectOption>
                  <SelectOption value="query">query</SelectOption>
                </Select>
              </FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="密码参数名"><Input v-model:value="editForm.change_pass_param" placeholder="newPwd" /></FormItem>
              <FormItem label="ID参数名"><Input v-model:value="editForm.change_pass_id_param" placeholder="id" /></FormItem>
            </div>
            <Alert class="mb-3" type="info" show-icon :message="changePassParamHelp" />
            <FormItem label="改密参数映射">
              <Input.TextArea v-model:value="editForm.change_pass_param_map" :rows="3" :placeholder="changePassParamMapExample" />
            </FormItem>
            <Typography.Text type="secondary" class="mb-2 mt-3 block">补单</Typography.Text>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="补单路径"><Input v-model:value="editForm.resubmit_path" placeholder="如 /api/reset" /></FormItem>
              <FormItem label="补单ID参数"><Input v-model:value="editForm.resubmit_id_param" placeholder="id" /></FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="补单方式">
                <Select v-model:value="editForm.resubmit_method" allow-clear placeholder="留空=未配置">
                  <SelectOption value="">未配置</SelectOption>
                  <SelectOption value="POST">POST</SelectOption>
                  <SelectOption value="GET">GET</SelectOption>
                  <SelectOption value="PUT">PUT</SelectOption>
                </Select>
              </FormItem>
              <FormItem label="补单Body">
                <Select v-model:value="editForm.resubmit_body_type" allow-clear placeholder="留空=自动">
                  <SelectOption value="">自动</SelectOption>
                  <SelectOption value="form">form</SelectOption>
                  <SelectOption value="json">json</SelectOption>
                  <SelectOption value="query">query</SelectOption>
                </Select>
              </FormItem>
            </div>
            <FormItem label="补单参数映射">
              <Input.TextArea v-model:value="editForm.resubmit_param_map" :rows="3" placeholder='{"uid":"{{supplier.uid}}","key":"{{supplier.key}}","id":"{{order.yid}}"}' />
            </FormItem>
            <Typography.Text type="secondary" class="mb-2 mt-3 block">日志</Typography.Text>
            <FormItem label="日志路径"><Input v-model:value="editForm.log_path" placeholder="如 /log/ 或 api.php?act=xq" /></FormItem>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="日志方式">
                <Select v-model:value="editForm.log_method" allow-clear placeholder="留空=未配置">
                  <SelectOption value="">未配置</SelectOption>
                  <SelectOption value="POST">POST</SelectOption>
                  <SelectOption value="GET">GET</SelectOption>
                </Select>
              </FormItem>
              <FormItem label="日志Body">
                <Select v-model:value="editForm.log_body_type" allow-clear placeholder="留空=自动">
                  <SelectOption value="">自动</SelectOption>
                  <SelectOption value="form">form</SelectOption>
                  <SelectOption value="json">json</SelectOption>
                  <SelectOption value="query">query</SelectOption>
                </Select>
              </FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="日志ID参数"><Input v-model:value="editForm.log_id_param" placeholder="id" /></FormItem>
            </div>
            <FormItem label="日志参数映射">
              <Input.TextArea v-model:value="editForm.log_param_map" :rows="3" placeholder='{"dtoken":"{{supplier.token}}","account":"{{order.user}}","password":"{{order.pass}}","course":"{{order.kcname}}","courseId":"{{order.kcid}}"}' />
            </FormItem>
          </TabPane>
          <TabPane key="balance" tab="余额">
            <FormItem label="余额路径"><Input v-model:value="editForm.balance_path" placeholder="如 /api/getinfo 或 api.php?act=getmoney" /></FormItem>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="金额字段路径">
                <Select v-model:value="editForm.balance_money_field">
                  <SelectOption value="money">money（根级）</SelectOption>
                  <SelectOption value="data.money">data.money</SelectOption>
                  <SelectOption value="data">data（整个data就是金额）</SelectOption>
                  <SelectOption value="data.remainscore">data.remainscore</SelectOption>
                </Select>
              </FormItem>
              <FormItem label="请求方式">
                <Select v-model:value="editForm.balance_method" allow-clear placeholder="留空=未配置">
                  <SelectOption value="">未配置</SelectOption>
                  <SelectOption value="POST">POST</SelectOption>
                  <SelectOption value="GET">GET</SelectOption>
                </Select>
              </FormItem>
              <FormItem label="Body类型">
                <Select v-model:value="editForm.balance_body_type" allow-clear placeholder="留空=自动">
                  <SelectOption value="">自动</SelectOption>
                  <SelectOption value="form">form</SelectOption>
                  <SelectOption value="json">json</SelectOption>
                  <SelectOption value="query">query</SelectOption>
                </Select>
              </FormItem>
            </div>
            <FormItem label="认证覆盖">
              <Select v-model:value="editForm.balance_auth_type" allow-clear placeholder="留空跟随全局认证">
                <SelectOption value="">跟随全局</SelectOption>
                <SelectOption value="uid_key">uid + key</SelectOption>
                <SelectOption value="token_only">token（pass字段）</SelectOption>
                <SelectOption value="bearer_token">Bearer token</SelectOption>
              </Select>
            </FormItem>
            <FormItem label="参数映射JSON">
              <Input.TextArea v-model:value="editForm.balance_param_map" :rows="3" placeholder='{"uid":"{{supplier.uid}}","key":"{{supplier.key}}"}' />
            </FormItem>
          </TabPane>
          <TabPane key="report" tab="工单/刷新">
            <Typography.Text type="secondary" class="mb-2 block">工单（提交/查询）</Typography.Text>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="提交工单路径"><Input v-model:value="editForm.report_path" placeholder="如 /api/submitWork" /></FormItem>
              <FormItem label="查询工单路径"><Input v-model:value="editForm.get_report_path" placeholder="如 /api/queryWork" /></FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="提交方式">
                <Select v-model:value="editForm.report_method" allow-clear placeholder="留空=未配置">
                  <SelectOption value="">未配置</SelectOption>
                  <SelectOption value="POST">POST</SelectOption>
                  <SelectOption value="GET">GET</SelectOption>
                  <SelectOption value="PUT">PUT</SelectOption>
                </Select>
              </FormItem>
              <FormItem label="提交Body">
                <Select v-model:value="editForm.report_body_type" allow-clear placeholder="留空=自动">
                  <SelectOption value="">自动</SelectOption>
                  <SelectOption value="form">form</SelectOption>
                  <SelectOption value="json">json</SelectOption>
                  <SelectOption value="query">query</SelectOption>
                </Select>
              </FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="参数风格">
                <Select v-model:value="editForm.report_param_style" allow-clear placeholder="留空=standard">
                  <SelectOption value="">standard（uid+key）</SelectOption>
                  <SelectOption value="token">token（JSON body）</SelectOption>
                </Select>
              </FormItem>
              <FormItem label="认证覆盖">
                <Select v-model:value="editForm.report_auth_type" allow-clear placeholder="留空跟随全局">
                  <SelectOption value="">跟随全局</SelectOption>
                  <SelectOption value="token_only">token_only</SelectOption>
                </Select>
              </FormItem>
            </div>
            <FormItem label="提交参数映射">
              <Alert class="mb-3" type="info" show-icon :message="reportParamHelp" />
              <Input.TextArea v-model:value="editForm.report_param_map" :rows="3" :placeholder="reportParamMapExample" />
            </FormItem>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="查询方式">
                <Select v-model:value="editForm.get_report_method" allow-clear placeholder="留空=未配置">
                  <SelectOption value="">未配置</SelectOption>
                  <SelectOption value="POST">POST</SelectOption>
                  <SelectOption value="GET">GET</SelectOption>
                  <SelectOption value="PUT">PUT</SelectOption>
                </Select>
              </FormItem>
              <FormItem label="查询Body">
                <Select v-model:value="editForm.get_report_body_type" allow-clear placeholder="留空=自动">
                  <SelectOption value="">自动</SelectOption>
                  <SelectOption value="form">form</SelectOption>
                  <SelectOption value="json">json</SelectOption>
                  <SelectOption value="query">query</SelectOption>
                </Select>
              </FormItem>
            </div>
            <FormItem label="查询参数映射">
              <Alert class="mb-3" type="info" show-icon :message="getReportParamHelp" />
              <Input.TextArea v-model:value="editForm.get_report_param_map" :rows="3" :placeholder="getReportParamMapExample" />
            </FormItem>
            <Typography.Text type="secondary" class="mb-2 mt-3 block">刷新进度</Typography.Text>
            <FormItem label="刷新路径"><Input v-model:value="editForm.refresh_path" placeholder="如 /api/refresh" /></FormItem>
          </TabPane>
        </Tabs>
      </Form>
    </Modal>

    <!-- PHP 导入弹窗 -->
    <Modal v-model:open="phpVisible" title="从 PHP 代码导入平台配置" width="750px" :footer="null">
      <Alert message="粘贴 PHP 的 if 代码块，系统会自动解析出配置" type="info" show-icon class="mb-3" />
      <Input.TextArea v-model:value="phpCode" :rows="12" placeholder='if ($type == "newplat") { ... }' style="font-family: monospace; font-size: 12px" />
      <div class="mt-3 flex items-center justify-between">
        <Button type="primary" :loading="phpLoading" @click="parsePHP">
          <template #icon><ThunderboltOutlined /></template>
          解析代码
        </Button>
      </div>
      <template v-if="phpResult">
        <div class="mt-4 rounded border p-3 bg-gray-50 dark:bg-gray-800">
          <div class="mb-2 flex items-center gap-2">
            <CheckCircleOutlined v-if="phpResult.confidence >= 50" style="color: #52c41a" />
            <WarningOutlined v-else style="color: #faad14" />
            <Typography.Text strong>解析结果（置信度 {{ phpResult.confidence }}%）</Typography.Text>
          </div>
          <div class="grid grid-cols-3 gap-2 text-xs">
            <div><b>平台：</b>{{ phpResult.pt || '未识别' }}</div>
            <div><b>认证：</b>{{ phpResult.auth_type }}</div>
            <div><b>成功码：</b>{{ phpResult.success_codes }}</div>
            <div><b>查课：</b>{{ formatEndpoint(phpResult.query_path, phpResult.query_act) }}</div>
            <div><b>下单：</b>{{ formatEndpoint(phpResult.order_path) }}</div>
            <div><b>进度：</b>{{ formatEndpoint(phpResult.progress_path) }}</div>
            <div><b>改密：</b>{{ formatEndpoint(phpResult.change_pass_path) }} ({{ phpResult.change_pass_param }})</div>
            <div><b>暂停：</b>{{ formatEndpoint(phpResult.pause_path) }}</div>
            <div><b>JSON：</b>{{ phpResult.use_json ? '是' : '否' }}</div>
          </div>
          <div v-if="phpResult.warnings?.length" class="mt-2">
            <Tag v-for="w in phpResult.warnings" :key="w" color="orange" class="mb-1">{{ w }}</Tag>
          </div>
          <div class="mt-3 text-right">
            <Button type="primary" @click="applyPHPResult">
              <template #icon><ImportOutlined /></template>
              应用并编辑
            </Button>
          </div>
        </div>
      </template>
    </Modal>

    <!-- 自动检测弹窗 -->
    <Modal v-model:open="detectVisible" title="自动检测平台接口" width="800px" :footer="null">
      <Alert message="输入上游平台的 URL 和凭证，系统会自动探测支持的接口和参数格式" type="info" show-icon class="mb-3" />
      <Form :label-col="{ span: 4 }" :wrapper-col="{ span: 20 }" size="small">
        <FormItem label="平台 URL" required><Input v-model:value="detectForm.url" placeholder="http://xxx.com" /></FormItem>
        <div class="grid grid-cols-2 gap-x-4">
          <FormItem label="UID"><Input v-model:value="detectForm.uid" placeholder="上游账号" /></FormItem>
          <FormItem label="Key"><Input v-model:value="detectForm.key" placeholder="上游密钥" /></FormItem>
        </div>
        <div class="grid grid-cols-2 gap-x-4">
          <FormItem label="Token"><Input v-model:value="detectForm.token" placeholder="Bearer Token（可选）" /></FormItem>
          <FormItem label="Cookie"><Input v-model:value="detectForm.cookie" placeholder="Cookie（可选）" /></FormItem>
        </div>
        <div class="grid grid-cols-2 gap-x-4">
          <FormItem label="平台标识"><Input v-model:value="detectForm.pt" placeholder="如 newplat（一键保存时必填）" /></FormItem>
          <FormItem label="平台名称"><Input v-model:value="detectForm.name" placeholder="如 新平台（留空自动生成）" /></FormItem>
        </div>
        <FormItem :wrapper-col="{ offset: 4 }">
          <Space>
            <Button type="primary" :loading="detectLoading" @click="runDetect">
              <template #icon><SearchOutlined /></template>
              仅检测
            </Button>
            <Button type="primary" :loading="detectSaveLoading" @click="autoDetectSave" style="background: #52c41a; border-color: #52c41a">
              <template #icon><ThunderboltOutlined /></template>
              一键检测并保存
            </Button>
          </Space>
        </FormItem>
      </Form>
      <template v-if="detectResult">
        <div class="mt-2 rounded border p-4" :class="detectResult.success ? 'bg-green-50 dark:bg-green-900/20' : 'bg-red-50 dark:bg-red-900/20'">
          <div class="mb-3 flex items-center justify-between">
            <div class="flex items-center gap-2">
              <CheckCircleOutlined v-if="detectResult.success" style="color: #52c41a; font-size: 18px" />
              <WarningOutlined v-else style="color: #ff4d4f; font-size: 18px" />
              <Typography.Text strong style="font-size: 15px">{{ detectResult.success ? '检测成功' : '未检测到可用接口' }}</Typography.Text>
            </div>
            <Button v-if="detectResult.success" type="primary" @click="applyDetectResult">
              <template #icon><ImportOutlined /></template>
              应用配置
            </Button>
          </div>
          <template v-if="detectResult.success">
            <div class="mb-3 grid grid-cols-2 gap-3">
              <Card size="small" :bordered="true"><div class="text-xs text-gray-500">认证方式</div><div class="font-medium">{{ detectResult.auth_type }}</div></Card>
              <Card size="small" :bordered="true"><div class="text-xs text-gray-500">成功码</div><div class="font-medium">{{ detectResult.success_code }}</div></Card>
            </div>
            <div class="mb-3 grid grid-cols-4 gap-3">
              <Card size="small" :bordered="true" :style="{ borderColor: detectResult.balance_ok ? '#52c41a' : '#d9d9d9' }"><div class="text-xs text-gray-500">余额</div><div class="font-medium">{{ detectResult.balance_ok ? detectResult.balance_money : '不可用' }}</div></Card>
              <Card size="small" :bordered="true" :style="{ borderColor: detectResult.query_ok ? '#52c41a' : '#d9d9d9' }"><div class="text-xs text-gray-500">查课</div><div class="font-medium">{{ detectResult.query_ok ? '✓' : '✗' }}</div></Card>
              <Card size="small" :bordered="true" :style="{ borderColor: detectResult.class_list_ok ? '#52c41a' : '#d9d9d9' }"><div class="text-xs text-gray-500">课程列表</div><div class="font-medium">{{ detectResult.class_list_ok ? '✓' : '✗' }}</div></Card>
              <Card size="small" :bordered="true" :style="{ borderColor: detectResult.category_ok ? '#52c41a' : '#d9d9d9' }"><div class="text-xs text-gray-500">分类</div><div class="font-medium">{{ detectResult.category_ok ? '✓' : '✗' }}</div></Card>
            </div>
          </template>
          <Collapse size="small" class="mt-2">
            <CollapsePanel key="probes" header="探测详情">
              <div class="space-y-1">
                <div v-for="(p, idx) in detectResult.probes" :key="idx" class="flex items-center gap-2 rounded px-2 py-1 text-xs" :class="p.status === 'ok' ? 'bg-green-50 dark:bg-green-900/20' : 'bg-gray-50 dark:bg-gray-800'">
                  <Tag :color="statusColor[p.status] || 'default'" style="min-width: 40px; text-align: center">{{ p.status }}</Tag>
                  <Tag>{{ p.method }}</Tag>
                  <span class="font-medium" style="min-width: 180px">{{ p.endpoint }}</span>
                  <span v-if="p.code" class="text-gray-500">code={{ p.code }}</span>
                  <span class="truncate text-gray-400">{{ p.msg }}</span>
                </div>
              </div>
            </CollapsePanel>
          </Collapse>
        </div>
      </template>
    </Modal>
  </div>
</template>

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
  deletePlatformConfigApi, parsePHPCodeApi, detectPlatformApi,
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
const detectForm = ref({ url: '', uid: '', key: '', token: '', cookie: '' });
const detectResult = ref<DetectResult | null>(null);

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
  { title: '查课', dataIndex: 'query_act', width: 100 },
  { title: '下单', dataIndex: 'order_act', width: 80 },
  { title: '进度', dataIndex: 'progress_act', width: 100 },
  { title: '改密', dataIndex: 'change_pass_act', width: 80 },
  { title: '特性', key: 'features', width: 180 },
  { title: '操作', key: 'action', width: 120, fixed: 'right' as const },
];

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
    auth_type: 'uid_key', api_path_style: 'standard', success_codes: '0',
    query_act: 'get', order_act: 'add', progress_act: 'chadan2',
    progress_no_yid: 'chadan', progress_method: 'POST', pause_act: 'zt',
    pause_id_param: 'id', change_pass_act: 'gaimi', change_pass_param: 'newPwd', change_pass_id_param: 'id',
    log_act: 'xq', log_method: 'POST', log_id_param: 'id',
    balance_act: 'getmoney', balance_money_field: 'money', balance_method: 'POST', balance_auth_type: '',
  };
  editVisible.value = true;
}

function openEdit(row: PlatformConfig) {
  isEdit.value = true;
  editForm.value = { ...row };
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
    api_path_style: r.api_path_style || 'standard', success_codes: r.success_codes || '0',
    use_json: r.use_json, query_act: r.query_act || 'get', query_path: r.query_path || '',
    order_act: r.order_act || 'add', order_path: r.order_path || '',
    progress_act: r.progress_act || 'chadan2', progress_path: r.progress_path || '',
    progress_method: r.progress_method || 'POST', pause_act: r.pause_act || 'zt',
    pause_path: r.pause_path || '', pause_id_param: r.pause_id_param || 'id', change_pass_act: r.change_pass_act || 'gaimi',
    change_pass_path: r.change_pass_path || '', change_pass_param: r.change_pass_param || 'newPwd',
    change_pass_id_param: r.change_pass_id_param || 'id', log_act: r.log_act || 'xq',
    log_path: r.log_path || '', log_id_param: r.log_id_param || 'id',
    returns_yid: r.returns_yid, progress_no_yid: 'chadan', log_method: 'POST',
    balance_act: r.balance_act || 'getmoney', balance_path: r.balance_path || '',
    balance_money_field: r.balance_money_field || 'money', balance_method: 'POST',
    balance_auth_type: '', source_code: phpCode.value,
  };
  isEdit.value = false; phpVisible.value = false; editVisible.value = true;
  message.info('已填充解析结果，请核对后保存');
}

function openDetect() {
  detectForm.value = { url: '', uid: '', key: '', token: '', cookie: '' };
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

function applyDetectResult() {
  if (!detectResult.value) return;
  const r = detectResult.value;
  const cfg = r.config || {};
  editForm.value = {
    pt: '', name: r.suggested_name || '', auth_type: cfg.auth_type || 'uid_key',
    api_path_style: cfg.api_path_style || 'standard', success_codes: cfg.success_codes || '0',
    use_json: cfg.use_json === 'true', query_act: cfg.query_act || '', query_path: cfg.query_path || '',
    order_act: cfg.order_act || '', order_path: cfg.order_path || '',
    progress_act: cfg.progress_act || '', progress_no_yid: cfg.progress_no_yid || '',
    progress_path: cfg.progress_path || '', progress_method: cfg.progress_method || 'POST',
    pause_act: cfg.pause_act || '', pause_path: cfg.pause_path || '',
    pause_id_param: cfg.pause_id_param || 'id', change_pass_act: cfg.change_pass_act || '', change_pass_path: cfg.change_pass_path || '',
    change_pass_param: 'newPwd', change_pass_id_param: 'id',
    resubmit_path: cfg.resubmit_path || '', resubmit_id_param: cfg.resubmit_id_param || 'id', refresh_path: cfg.refresh_path || '',
    log_act: cfg.log_act || '', log_path: cfg.log_path || '',
    log_method: cfg.log_method || 'POST', log_id_param: 'id',
    balance_act: cfg.balance_act || '', balance_path: cfg.balance_path || '',
    balance_money_field: cfg.balance_money_field || 'money', balance_method: 'POST',
    balance_auth_type: cfg.auth_type === 'bearer_token' ? 'bearer_token' : '',
    report_path: cfg.report_path || '', get_report_path: cfg.get_report_path || '',
    report_param_style: cfg.report_param_style || '', report_auth_type: cfg.report_auth_type || '',
  };
  isEdit.value = false; detectVisible.value = false; editVisible.value = true;
  message.info('已填充检测结果，请设置平台标识后保存');
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
        <template v-else-if="column.dataIndex === 'query_act'">
          <span>{{ record.query_act }}</span>
          <span v-if="record.query_path" class="ml-1 text-xs text-gray-400">{{ record.query_path }}</span>
        </template>
        <template v-else-if="column.key === 'features'">
          <Space :size="2" wrap>
            <Tag v-if="record.use_json" color="purple" :bordered="false">JSON</Tag>
            <Tag v-if="record.need_proxy" color="red" :bordered="false">代理</Tag>
            <Tag v-if="record.returns_yid" color="green" :bordered="false">返YID</Tag>
            <Tag v-if="record.progress_path" color="cyan" :bordered="false">自定义进度路径</Tag>
            <Tag v-if="record.pause_path" color="cyan" :bordered="false">自定义暂停路径</Tag>
            <Tag v-if="record.change_pass_path" color="cyan" :bordered="false">自定义改密路径</Tag>
            <Tag v-if="record.extra_params" color="gold" :bordered="false">额外参数</Tag>
          </Space>
        </template>
        <template v-else-if="column.key === 'action'">
          <Space>
            <Tooltip title="编辑">
              <Button type="link" size="small" @click="openEdit(record)"><EditOutlined /></Button>
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
              <FormItem label="API风格">
                <Select v-model:value="editForm.api_path_style">
                  <SelectOption value="standard">/api.php?act=</SelectOption>
                  <SelectOption value="rest">REST 自定义路径</SelectOption>
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
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="查课 act"><Input v-model:value="editForm.query_act" placeholder="get" /></FormItem>
              <FormItem label="查课路径"><Input v-model:value="editForm.query_path" placeholder="REST时填" /></FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="参数风格"><Input v-model:value="editForm.query_param_style" placeholder="standard" /></FormItem>
              <FormItem label="需要轮询"><Switch v-model:checked="editForm.query_polling" /></FormItem>
            </div>
          </TabPane>
          <TabPane key="order" tab="下单">
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="下单 act"><Input v-model:value="editForm.order_act" placeholder="add" /></FormItem>
              <FormItem label="下单路径"><Input v-model:value="editForm.order_path" placeholder="REST时填" /></FormItem>
            </div>
          </TabPane>
          <TabPane key="progress" tab="进度">
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="有YID act"><Input v-model:value="editForm.progress_act" placeholder="chadan2" /></FormItem>
              <FormItem label="无YID act"><Input v-model:value="editForm.progress_no_yid" placeholder="chadan" /></FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="进度路径"><Input v-model:value="editForm.progress_path" placeholder="如 /api/search" /></FormItem>
              <FormItem label="请求方式">
                <Select v-model:value="editForm.progress_method">
                  <SelectOption value="POST">POST</SelectOption>
                  <SelectOption value="GET">GET</SelectOption>
                </Select>
              </FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="用id参数"><Switch v-model:checked="editForm.use_id_param" /></FormItem>
              <FormItem label="用uuid参数"><Switch v-model:checked="editForm.use_uuid_param" /></FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="总传username"><Switch v-model:checked="editForm.always_username" /></FormItem>
              <FormItem label="需要认证"><Switch v-model:checked="editForm.progress_needs_auth" /></FormItem>
            </div>
          </TabPane>
          <TabPane key="other" tab="暂停/改密/日志">
            <Typography.Text type="secondary" class="mb-2 block">暂停/恢复</Typography.Text>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="暂停 act"><Input v-model:value="editForm.pause_act" placeholder="zt" /></FormItem>
              <FormItem label="暂停路径"><Input v-model:value="editForm.pause_path" placeholder="如 /api/stop" /></FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="暂停ID参数"><Input v-model:value="editForm.pause_id_param" placeholder="id" /></FormItem>
              <FormItem label="恢复 act"><Input v-model:value="editForm.resume_act" /></FormItem>
              <FormItem label="恢复路径"><Input v-model:value="editForm.resume_path" /></FormItem>
            </div>
            <Typography.Text type="secondary" class="mb-2 mt-3 block">改密码</Typography.Text>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="改密 act"><Input v-model:value="editForm.change_pass_act" placeholder="gaimi" /></FormItem>
              <FormItem label="改密路径"><Input v-model:value="editForm.change_pass_path" placeholder="如 /api/update" /></FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="密码参数名"><Input v-model:value="editForm.change_pass_param" placeholder="newPwd" /></FormItem>
              <FormItem label="ID参数名"><Input v-model:value="editForm.change_pass_id_param" placeholder="id" /></FormItem>
            </div>
            <Typography.Text type="secondary" class="mb-2 mt-3 block">补单</Typography.Text>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="补单路径"><Input v-model:value="editForm.resubmit_path" placeholder="如 /api/reset" /></FormItem>
              <FormItem label="补单ID参数"><Input v-model:value="editForm.resubmit_id_param" placeholder="id" /></FormItem>
            </div>
            <Typography.Text type="secondary" class="mb-2 mt-3 block">日志</Typography.Text>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="日志 act"><Input v-model:value="editForm.log_act" placeholder="xq" /></FormItem>
              <FormItem label="日志路径"><Input v-model:value="editForm.log_path" placeholder="如 /log/" /></FormItem>
            </div>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="日志方式">
                <Select v-model:value="editForm.log_method">
                  <SelectOption value="POST">POST</SelectOption>
                  <SelectOption value="GET">GET</SelectOption>
                </Select>
              </FormItem>
              <FormItem label="日志ID参数"><Input v-model:value="editForm.log_id_param" placeholder="id" /></FormItem>
            </div>
          </TabPane>
          <TabPane key="balance" tab="余额">
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="余额 act"><Input v-model:value="editForm.balance_act" placeholder="getmoney" /></FormItem>
              <FormItem label="余额路径"><Input v-model:value="editForm.balance_path" placeholder="如 /api/getinfo" /></FormItem>
            </div>
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
                <Select v-model:value="editForm.balance_method">
                  <SelectOption value="POST">POST</SelectOption>
                  <SelectOption value="GET">GET</SelectOption>
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
          </TabPane>
          <TabPane key="report" tab="工单/刷新">
            <Typography.Text type="secondary" class="mb-2 block">工单（提交/查询）</Typography.Text>
            <div class="grid grid-cols-2 gap-x-4">
              <FormItem label="提交工单路径"><Input v-model:value="editForm.report_path" placeholder="如 /api/submitWork" /></FormItem>
              <FormItem label="查询工单路径"><Input v-model:value="editForm.get_report_path" placeholder="如 /api/queryWork" /></FormItem>
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
            <div><b>查课：</b>{{ phpResult.query_act }}{{ phpResult.query_path ? ` (${phpResult.query_path})` : '' }}</div>
            <div><b>下单：</b>{{ phpResult.order_act }}</div>
            <div><b>进度：</b>{{ phpResult.progress_act }}</div>
            <div><b>改密：</b>{{ phpResult.change_pass_act }} ({{ phpResult.change_pass_param }})</div>
            <div><b>暂停：</b>{{ phpResult.pause_act }}</div>
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
        <FormItem :wrapper-col="{ offset: 4 }">
          <Button type="primary" :loading="detectLoading" @click="runDetect">
            <template #icon><SearchOutlined /></template>
            开始检测
          </Button>
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
            <div class="mb-3 grid grid-cols-3 gap-3">
              <Card size="small" :bordered="true"><div class="text-xs text-gray-500">认证方式</div><div class="font-medium">{{ detectResult.auth_type }}</div></Card>
              <Card size="small" :bordered="true"><div class="text-xs text-gray-500">成功码</div><div class="font-medium">{{ detectResult.success_code }}</div></Card>
              <Card size="small" :bordered="true"><div class="text-xs text-gray-500">API 风格</div><div class="font-medium">{{ detectResult.api_style }}</div></Card>
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

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue';
import {
  Alert,
  Button,
  Card,
  Form,
  FormItem,
  Input,
  InputNumber,
  InputPassword,
  Switch,
  Spin,
  Table,
  Tag,
  Tabs,
  TabPane,
  Collapse,
  CollapsePanel,
  Descriptions,
  DescriptionsItem,
  Badge,
  Modal,
  message,
} from 'ant-design-vue';
import {
  DatabaseOutlined,
  SyncOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  ThunderboltOutlined,
  ReloadOutlined,
  ApiOutlined,
} from '@ant-design/icons-vue';
import type {
  SyncResult,
  SyncTestResult,
  DBCompatCheckResult,
  DBCompatFixResult,
} from '#/api/admin';
import {
  dbSyncTestApi,
  dbSyncExecuteApi,
  dbCompatCheckApi,
  dbCompatFixApi,
} from '#/api/admin';

const activeTab = ref('sync');

// ===== 数据同步 =====
const syncForm = ref({
  host: 'localhost',
  port: 3306,
  db_name: '',
  user: 'root',
  password: '',
  update_existing: false,
});

const testing = ref(false);
const syncing = ref(false);
const testResult = ref<SyncTestResult | null>(null);
const syncResult = ref<SyncResult | null>(null);
const confirmationToken = ref('');
const syncResultRef = ref<HTMLElement | null>(null);

const tableLabels: Record<string, string> = {
  qingka_wangke_dengji: '等级',
  qingka_wangke_huoyuan: '货源',
  qingka_wangke_user: '用户',
  qingka_wangke_fenlei: '分类',
  qingka_wangke_class: '商品',
  qingka_wangke_config: '配置',
  qingka_wangke_gonggao: '公告',
  qingka_wangke_mijia: '密价',
  qingka_wangke_km: '卡密',
  qingka_wangke_order: '订单',
  qingka_wangke_pay: '支付',
};

const canSync = computed(() => {
  return syncForm.value.host && syncForm.value.db_name && syncForm.value.user;
});
const precheckPassed = computed(() => {
  return !testing.value && !!testResult.value?.connected && !!testResult.value?.ready && !!confirmationToken.value;
});

function resetSyncState() {
  confirmationToken.value = '';
  testResult.value = null;
  syncResult.value = null;
}

async function scrollToSyncResult() {
  await nextTick();
  syncResultRef.value?.scrollIntoView({
    behavior: 'smooth',
    block: 'start',
  });
}

watch(syncForm, () => {
  resetSyncState();
}, { deep: true });

async function doTest() {
  testing.value = true;
  confirmationToken.value = '';
  syncResult.value = null;
  try {
    testResult.value = await dbSyncTestApi(syncForm.value);
    confirmationToken.value = testResult.value.confirmation_token || '';
    if (!testResult.value.connected) {
      message.error('连接失败: ' + (testResult.value.error || '未知错误'));
    } else if (testResult.value.ready) {
      message.success('预检查通过');
    } else {
      message.warning('预检查未通过，请先处理结构差异');
    }
  } catch (e: any) {
    confirmationToken.value = '';
    message.error('测试失败: ' + (e?.message || e));
  } finally {
    testing.value = false;
  }
}

async function doSync() {
  if (!canSync.value) {
    message.warning('请填写完整的数据库连接信息');
    return;
  }
  if (!precheckPassed.value || !testResult.value) {
    message.warning('请先完成预检查并确保通过');
    return;
  }
  const warningText = testResult.value.warnings?.length ? `\n\n风险提示：${testResult.value.warnings.join('；')}` : '';
  Modal.confirm({
    title: '确认开始导入？',
    content: `请先确认已完成当前数据库备份。导入将写入当前系统的核心数据表，且预检查令牌仅在 10 分钟内有效。${warningText}`,
    okText: '确认导入',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      syncing.value = true;
      syncResult.value = null;
      try {
        syncResult.value = await dbSyncExecuteApi({
          ...syncForm.value,
          confirmation_token: confirmationToken.value,
        });
        confirmationToken.value = '';
        if (syncResult.value.success) {
          message.success('导入完成');
        } else {
          message.warning('导入完成，但有部分错误');
        }
      } catch (e: any) {
        message.error('同步失败: ' + (e?.message || e));
      } finally {
        syncing.value = false;
        if (syncResult.value) {
          await scrollToSyncResult();
        }
      }
    },
  });
}

const syncColumns = [
  { title: '数据类型', dataIndex: 'label', key: 'label', width: 100 },
  { title: '总数', dataIndex: 'total', key: 'total', width: 80 },
  { title: '新增', dataIndex: 'inserted', key: 'inserted', width: 80 },
  { title: '更新', dataIndex: 'updated', key: 'updated', width: 80 },
  { title: '跳过', dataIndex: 'skipped', key: 'skipped', width: 80 },
  { title: '失败', dataIndex: 'failed', key: 'failed', width: 80 },
];

const precheckColumns = [
  { title: '数据类型', dataIndex: 'label', key: 'label', width: 110 },
  { title: '命中源表', dataIndex: 'source_table', key: 'source_table', width: 180 },
  { title: '源库条数', dataIndex: 'source_count', key: 'source_count', width: 100 },
  { title: '本地条数', dataIndex: 'local_count', key: 'local_count', width: 100 },
  { title: '状态', dataIndex: 'ready', key: 'ready', width: 100 },
  { title: '缺失字段', dataIndex: 'missing_local_columns', key: 'missing_local_columns', width: 180 },
  { title: '说明', dataIndex: 'message', key: 'message' },
];

// ===== 结构检测 =====
const checking = ref(false);
const fixing = ref(false);
const checkResult = ref<DBCompatCheckResult | null>(null);
const fixResult = ref<DBCompatFixResult | null>(null);

async function doCheck() {
  checking.value = true;
  checkResult.value = null;
  fixResult.value = null;
  try {
    checkResult.value = await dbCompatCheckApi();
  } catch (e: any) {
    message.error('检查失败: ' + (e?.message || e));
  } finally {
    checking.value = false;
  }
}

async function doFix() {
  Modal.confirm({
    title: '确认执行修复？',
    content: '将自动创建缺失的表和列，此操作不可逆。建议先备份数据库。',
    okText: '执行修复',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      fixing.value = true;
      fixResult.value = null;
      try {
        fixResult.value = await dbCompatFixApi();
        message.success('修复完成');
        await doCheck();
      } catch (e: any) {
        message.error('修复失败: ' + (e?.message || e));
      } finally {
        fixing.value = false;
      }
    },
  });
}

const missingColColumns = [
  { title: '表名', dataIndex: 'table', key: 'table' },
  { title: '列名', dataIndex: 'column', key: 'column' },
  { title: '类型', dataIndex: 'type', key: 'type' },
];
</script>

<template>
  <div class="p-4">
    <Card>
      <template #title>
        <div class="flex items-center gap-2">
          <DatabaseOutlined class="text-lg" />
          <span>数据库工具</span>
        </div>
      </template>

      <Tabs v-model:activeKey="activeTab">
        <!-- ===== Tab 1: 同步数据 ===== -->
        <TabPane key="sync" tab="同步数据">
          <Alert
            class="mb-5"
            type="warning"
            show-icon
            message="用于从旧 29 系统一次性导入核心数据，导入前请先备份当前数据库"
          />

          <Form layout="horizontal" :label-col="{ span: 5 }" :wrapper-col="{ span: 14 }">
            <FormItem label="数据库地址" required>
              <Input v-model:value="syncForm.host" placeholder="localhost" />
            </FormItem>
            <FormItem label="数据库端口" required>
              <InputNumber
                v-model:value="syncForm.port"
                :min="1"
                :max="65535"
                style="width: 200px"
              />
            </FormItem>
            <FormItem label="数据库名" required>
              <Input v-model:value="syncForm.db_name" placeholder="请输入数据库名" />
            </FormItem>
            <FormItem label="数据库用户名" required>
              <Input v-model:value="syncForm.user" placeholder="root" />
            </FormItem>
            <FormItem label="数据库密码">
              <InputPassword v-model:value="syncForm.password" placeholder="请输入数据库密码" />
            </FormItem>
            <FormItem label="更新已存在数据">
              <div class="flex items-center gap-3">
                <Switch v-model:checked="syncForm.update_existing" />
                <span class="text-gray-400 text-xs dark:text-gray-500">
                  开启后会更新已存在的同主键记录（包含配置、公告、支付等核心数据）
                </span>
              </div>
            </FormItem>
          </Form>

          <Alert
            class="mb-5"
            type="info"
            show-icon
            message="将导入以下全部数据：等级、货源、用户、分类、商品、配置、公告、密价、卡密、订单、支付"
          />

          <div class="flex justify-center gap-3 mb-4">
            <Button @click="doTest" :loading="testing" :disabled="!canSync">
              <template #icon><ApiOutlined /></template>
              预检查
            </Button>
            <Button
              type="primary"
              danger
              @click="doSync"
              :loading="syncing"
              :disabled="!precheckPassed"
            >
              <template #icon><SyncOutlined /></template>
              开始导入
            </Button>
          </div>

          <Alert
            v-if="syncing"
            class="mb-4"
            type="info"
            show-icon
            message="正在导入数据"
            description="导入可能持续较久，完成后页面会自动定位到结果区域。"
          />
          <Alert
            v-else-if="syncResult"
            class="mb-4"
            :type="syncResult.success ? 'success' : 'warning'"
            show-icon
            :message="syncResult.success ? '导入已完成' : '导入已完成（有警告）'"
            :description="syncResult.summary"
          />

          <!-- 测试连接结果 -->
          <template v-if="testResult">
            <Card size="small" class="mb-4 shadow-sm" :bordered="true">
              <template #title>
                <span :class="testResult.connected && testResult.ready ? 'text-green-600' : 'text-orange-500'">
                  <CheckCircleOutlined v-if="testResult.connected && testResult.ready" class="mr-1" />
                  <ExclamationCircleOutlined v-else class="mr-1" />
                  {{ testResult.connected ? (testResult.ready ? '预检查通过' : '预检查未通过') : '连接失败' }}
                </span>
              </template>
              <template #extra>
                <span class="text-xs text-gray-400">{{ testResult.tested_at }}</span>
              </template>
              <template v-if="testResult.connected && testResult.tables">
                <div class="flex flex-wrap gap-4 mb-4">
                  <div v-for="(count, tbl) in testResult.tables" :key="tbl" class="text-sm">
                    <Tag :color="count >= 0 ? 'blue' : 'red'">
                      {{ tableLabels[tbl] || tbl }}：{{ count >= 0 ? count + ' 条' : '表不存在' }}
                    </Tag>
                  </div>
                </div>
                <div class="mb-3 rounded-md border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-500 dark:border-gray-700 dark:bg-gray-800/60 dark:text-gray-400">
                  {{ testResult.summary }}
                </div>
                <template v-if="testResult.warnings?.length">
                  <Alert
                    v-for="(warning, i) in testResult.warnings"
                    :key="i"
                    type="warning"
                    :message="warning"
                    class="mb-2"
                    show-icon
                  />
                </template>
                <Table
                  :dataSource="testResult.table_checks"
                  :columns="precheckColumns"
                  :pagination="false"
                  size="small"
                  rowKey="table"
                  class="mt-3"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'source_table'">
                      <span :class="record.source_table && record.source_table !== record.table ? 'text-blue-500' : 'text-gray-500'">
                        {{ record.source_table || '-' }}
                      </span>
                    </template>
                    <template v-else-if="column.key === 'ready'">
                      <Badge
                        :status="record.skip ? 'warning' : (record.ready ? 'success' : 'error')"
                        :text="record.skip ? '将跳过' : (record.ready ? '可导入' : '需处理')"
                      />
                    </template>
                    <template v-else-if="column.key === 'missing_local_columns'">
                      <span v-if="record.missing_local_columns?.length" class="text-red-500 text-xs">
                        {{ record.missing_local_columns.join(', ') }}
                      </span>
                      <span v-else class="text-gray-400">无</span>
                    </template>
                    <template v-else-if="column.key === 'source_count' || column.key === 'local_count'">
                      <span>{{ record[column.key] < 0 ? '不可用' : record[column.key] }}</span>
                    </template>
                  </template>
                </Table>
              </template>
              <template v-if="testResult.error">
                <div class="text-red-500 text-sm">{{ testResult.error }}</div>
              </template>
            </Card>
          </template>

          <!-- 同步结果 -->
          <Spin :spinning="syncing" tip="正在同步，请稍候...">
            <template v-if="syncResult">
              <div ref="syncResultRef">
                <Card size="small" :bordered="true" class="shadow-sm">
                  <template #title>
                    <span :class="syncResult.success ? 'text-green-600' : 'text-orange-500'">
                      <CheckCircleOutlined v-if="syncResult.success" class="mr-1" />
                      <ExclamationCircleOutlined v-else class="mr-1" />
                      {{ syncResult.success ? '同步完成' : '同步完成（有错误）' }}
                    </span>
                  </template>
                  <div class="mb-3 rounded-md border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-500 dark:border-gray-700 dark:bg-gray-800/60 dark:text-gray-400">
                    {{ syncResult.summary }} · {{ syncResult.sync_time }}
                  </div>
                  <Table
                    :dataSource="syncResult.details"
                    :columns="syncColumns"
                    :pagination="false"
                    size="small"
                    rowKey="table"
                  >
                    <template #bodyCell="{ column, record }">
                      <template v-if="column.key === 'label'">
                        <div class="flex flex-col">
                          <div class="flex items-center gap-2">
                            <span class="font-medium">{{ record.label }}</span>
                            <Tag v-if="record.skipped_empty" color="orange">已跳过（源表为空）</Tag>
                          </div>
                          <span v-if="record.source_table" class="text-xs text-gray-400">
                            源表：{{ record.source_table }}
                          </span>
                          <span v-if="record.message && !record.skipped_empty" class="text-xs text-gray-400">
                            {{ record.message }}
                          </span>
                        </div>
                      </template>
                      <template v-else-if="column.key === 'inserted' && record.inserted > 0">
                        <span class="text-green-600 font-medium">+{{ record.inserted }}</span>
                      </template>
                      <template v-else-if="column.key === 'updated' && record.updated > 0">
                        <span class="text-blue-500 font-medium">{{ record.updated }}</span>
                      </template>
                      <template v-else-if="column.key === 'failed' && record.failed > 0">
                        <span class="text-red-500 font-medium">{{ record.failed }}</span>
                      </template>
                    </template>
                  </Table>
                  <template v-if="syncResult.errors.length > 0">
                    <div class="mt-3">
                      <Alert
                        v-for="(err, i) in syncResult.errors"
                        :key="i"
                        type="error"
                        :message="err"
                        class="mb-1"
                        show-icon
                      />
                    </div>
                  </template>
                </Card>
              </div>
            </template>
          </Spin>
        </TabPane>

        <!-- ===== Tab 2: 结构检测 ===== -->
        <TabPane key="compat" tab="结构检测" @click="!checkResult && doCheck()">
          <div class="flex justify-end gap-2 mb-4">
            <Button @click="doCheck" :loading="checking">
              <template #icon><ReloadOutlined /></template>
              重新检测
            </Button>
            <Button type="primary" danger @click="doFix" :loading="fixing" :disabled="checking">
              <template #icon><ThunderboltOutlined /></template>
              一键修复
            </Button>
          </div>

          <Alert
            class="mb-4"
            type="info"
            show-icon
            message="自动检测当前数据库与系统所需结构的差异，可一键创建缺失的表和列。"
          />

          <Spin :spinning="checking" tip="正在检测...">
            <template v-if="checkResult">
              <Descriptions bordered :column="2" size="small" class="mb-4">
                <DescriptionsItem label="检测时间">{{ checkResult.check_time }}</DescriptionsItem>
                <DescriptionsItem label="总结">
                  <span class="font-medium">{{ checkResult.summary }}</span>
                </DescriptionsItem>
                <DescriptionsItem label="期望表数">
                  <Badge :count="checkResult.total_tables" :overflow-count="999" show-zero :number-style="{ backgroundColor: '#1677ff' }" />
                </DescriptionsItem>
                <DescriptionsItem label="状态">
                  <Tag v-if="checkResult.missing_tables.length === 0 && checkResult.missing_columns.length === 0" color="success">
                    <CheckCircleOutlined /> 完全兼容
                  </Tag>
                  <Tag v-else color="warning">
                    <ExclamationCircleOutlined /> 需要修复
                  </Tag>
                </DescriptionsItem>
              </Descriptions>

              <Collapse v-if="checkResult.missing_tables.length > 0 || checkResult.missing_columns.length > 0 || checkResult.existing_tables.length > 0 || checkResult.extra_tables.length > 0">
                <CollapsePanel v-if="checkResult.missing_tables.length > 0" key="missing-tables">
                  <template #header>
                    <span class="text-red-500 font-medium">缺失表 ({{ checkResult.missing_tables.length }})</span>
                  </template>
                  <div class="flex flex-wrap gap-2">
                    <Tag v-for="t in checkResult.missing_tables" :key="t" color="error">{{ t }}</Tag>
                  </div>
                </CollapsePanel>
                <CollapsePanel v-if="checkResult.missing_columns.length > 0" key="missing-cols">
                  <template #header>
                    <span class="text-orange-500 font-medium">缺失列 ({{ checkResult.missing_columns.length }})</span>
                  </template>
                  <Table :dataSource="checkResult.missing_columns" :columns="missingColColumns" :pagination="false" size="small" />
                </CollapsePanel>
                <CollapsePanel key="existing">
                  <template #header>
                    <span class="text-green-600 font-medium">已存在核心表 ({{ checkResult.existing_tables.length }})</span>
                  </template>
                  <div class="flex flex-wrap gap-2">
                    <Tag v-for="t in checkResult.existing_tables" :key="t" color="success">{{ t }}</Tag>
                  </div>
                </CollapsePanel>
                <CollapsePanel v-if="checkResult.extra_tables && checkResult.extra_tables.length > 0" key="extra">
                  <template #header>
                    <span class="text-blue-500 font-medium">数据库其他表 ({{ checkResult.extra_tables.length }})</span>
                  </template>
                  <div class="flex flex-wrap gap-2">
                    <Tag v-for="t in checkResult.extra_tables" :key="t" color="default">{{ t }}</Tag>
                  </div>
                </CollapsePanel>
              </Collapse>

              <Alert
                v-if="checkResult.missing_tables.length === 0 && checkResult.missing_columns.length === 0"
                type="success" show-icon message="数据库结构完全兼容，无需修复。" class="mt-4"
              />
            </template>
          </Spin>

          <!-- 修复结果 -->
          <template v-if="fixResult">
            <Card class="mt-4" size="small" title="修复结果">
              <Descriptions bordered :column="1" size="small">
                <DescriptionsItem label="修复时间">{{ fixResult.fix_time }}</DescriptionsItem>
                <DescriptionsItem label="总结">
                  <span class="font-medium">{{ fixResult.summary }}</span>
                </DescriptionsItem>
              </Descriptions>
              <div v-if="fixResult.tables_created.length > 0" class="mt-3">
                <div class="mb-1 font-medium text-green-600">创建的表：</div>
                <div class="flex flex-wrap gap-2">
                  <Tag v-for="t in fixResult.tables_created" :key="t" color="success">{{ t }}</Tag>
                </div>
              </div>
              <div v-if="fixResult.columns_added.length > 0" class="mt-3">
                <div class="mb-1 font-medium text-blue-600">添加的列：</div>
                <div class="flex flex-wrap gap-2">
                  <Tag v-for="c in fixResult.columns_added" :key="c" color="processing">{{ c }}</Tag>
                </div>
              </div>
              <div v-if="fixResult.errors.length > 0" class="mt-3">
                <div class="mb-1 font-medium text-red-500">错误：</div>
                <Alert v-for="(err, i) in fixResult.errors" :key="i" type="error" :message="err" class="mb-1" show-icon />
              </div>
              <Alert v-if="fixResult.admin_created" type="warning" show-icon message="已自动创建管理员账号：admin / admin123，请尽快修改密码！" class="mt-3" />
            </Card>
          </template>
        </TabPane>
      </Tabs>
    </Card>
  </div>
</template>

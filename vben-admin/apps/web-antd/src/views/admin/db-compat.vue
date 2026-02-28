<script setup lang="ts">
import { ref, computed } from 'vue';
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

const tableLabels: Record<string, string> = {
  qingka_wangke_huoyuan: '货源',
  qingka_wangke_user: '用户',
  qingka_wangke_fenlei: '分类',
  qingka_wangke_class: '商品',
  qingka_wangke_order: '订单',
};

const canSync = computed(() => {
  return syncForm.value.host && syncForm.value.db_name && syncForm.value.user;
});

async function doTest() {
  testing.value = true;
  testResult.value = null;
  try {
    testResult.value = await dbSyncTestApi(syncForm.value);
    if (testResult.value.connected) {
      message.success('连接成功');
    } else {
      message.error('连接失败: ' + (testResult.value.error || '未知错误'));
    }
  } catch (e: any) {
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
  syncing.value = true;
  syncResult.value = null;
  try {
    syncResult.value = await dbSyncExecuteApi(syncForm.value);
    if (syncResult.value.success) {
      message.success('同步完成');
    } else {
      message.warning('同步完成，但有部分错误');
    }
  } catch (e: any) {
    message.error('同步失败: ' + (e?.message || e));
  } finally {
    syncing.value = false;
  }
}

const syncColumns = [
  { title: '数据类型', dataIndex: 'label', key: 'label', width: 100 },
  { title: '总数', dataIndex: 'total', key: 'total', width: 80 },
  { title: '新增', dataIndex: 'inserted', key: 'inserted', width: 80 },
  { title: '更新', dataIndex: 'updated', key: 'updated', width: 80 },
  { title: '跳过', dataIndex: 'skipped', key: 'skipped', width: 80 },
  { title: '失败', dataIndex: 'failed', key: 'failed', width: 80 },
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
            message="支持从其他29系统同步用户数据到当前系统"
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
                  开启后会更新已存在的数据（用户、货源、分类、商品、订单）
                </span>
              </div>
            </FormItem>
          </Form>

          <Alert
            class="mb-5"
            type="info"
            show-icon
            message="将同步以下全部数据：货源、用户、分类、商品、订单"
          />

          <div class="flex justify-center gap-3 mb-4">
            <Button @click="doTest" :loading="testing" :disabled="!canSync">
              <template #icon><ApiOutlined /></template>
              测试连接
            </Button>
            <Button
              type="primary"
              @click="doSync"
              :loading="syncing"
              :disabled="!canSync"
            >
              <template #icon><SyncOutlined /></template>
              开始同步
            </Button>
          </div>

          <!-- 测试连接结果 -->
          <template v-if="testResult">
            <Card size="small" class="mb-4" :bordered="true">
              <template #title>
                <span :class="testResult.connected ? 'text-green-600' : 'text-red-500'">
                  {{ testResult.connected ? '✅ 连接成功' : '❌ 连接失败' }}
                </span>
              </template>
              <template v-if="testResult.connected && testResult.tables">
                <div class="flex flex-wrap gap-4">
                  <div v-for="(count, tbl) in testResult.tables" :key="tbl" class="text-sm">
                    <Tag :color="count >= 0 ? 'blue' : 'red'">
                      {{ tableLabels[tbl] || tbl }}：{{ count >= 0 ? count + ' 条' : '表不存在' }}
                    </Tag>
                  </div>
                </div>
              </template>
              <template v-if="testResult.error">
                <div class="text-red-500 text-sm">{{ testResult.error }}</div>
              </template>
            </Card>
          </template>

          <!-- 同步结果 -->
          <Spin :spinning="syncing" tip="正在同步，请稍候...">
            <template v-if="syncResult">
              <Card size="small" :bordered="true">
                <template #title>
                  <span :class="syncResult.success ? 'text-green-600' : 'text-orange-500'">
                    {{ syncResult.success ? '✅ 同步完成' : '⚠️ 同步完成（有错误）' }}
                  </span>
                </template>
                <div class="mb-3 text-sm text-gray-500 dark:text-gray-400">
                  {{ syncResult.summary }} · {{ syncResult.sync_time }}
                </div>
                <Table
                  :dataSource="syncResult.details"
                  :columns="syncColumns"
                  :pagination="false"
                  size="small"
                  rowKey="table"
                />
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

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal, Tabs, TabPane,
  Popconfirm, message, Pagination, Switch,
} from 'ant-design-vue';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons-vue';
import {
  getPledgeConfigListApi, savePledgeConfigApi, deletePledgeConfigApi,
  togglePledgeConfigApi, getPledgeRecordListApi,
  type PledgeConfig, type PledgeRecord,
} from '#/api/auxiliary';

const activeTab = ref('configs');

// ===== 配置列表 =====
const configLoading = ref(false);
const configs = ref<PledgeConfig[]>([]);

const configModalVisible = ref(false);
const configSaving = ref(false);
const configForm = reactive({
  id: 0, category_id: 0, amount: 0, discount_rate: 0, days: 30, cancel_fee: 0,
});

const configColumns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 60, align: 'center' as const },
  { title: '分类', dataIndex: 'category_name', key: 'category_name', width: 140 },
  { title: '质押金额', dataIndex: 'amount', key: 'amount', width: 110, align: 'center' as const },
  { title: '折扣率', dataIndex: 'discount_rate', key: 'discount_rate', width: 100, align: 'center' as const },
  { title: '天数', dataIndex: 'days', key: 'days', width: 80, align: 'center' as const },
  { title: '取消扣费%', key: 'cancel_fee', width: 110, align: 'center' as const },
  { title: '状态', key: 'status', width: 100, align: 'center' as const },
  { title: '操作', key: 'action', width: 160, align: 'center' as const, fixed: 'right' as const },
];

async function loadConfigs() {
  configLoading.value = true;
  try {
    const res = await getPledgeConfigListApi();
    configs.value = Array.isArray(res) ? res : [];
  } catch (e) { console.error(e); }
  finally { configLoading.value = false; }
}

function openAddConfig() {
  Object.assign(configForm, { id: 0, category_id: 0, amount: 0, discount_rate: 0, days: 30, cancel_fee: 0 });
  configModalVisible.value = true;
}

function openEditConfig(r: PledgeConfig) {
  Object.assign(configForm, {
    id: r.id, category_id: r.category_id, amount: r.amount,
    discount_rate: r.discount_rate, days: r.days, cancel_fee: r.cancel_fee,
  });
  configModalVisible.value = true;
}

async function handleSaveConfig() {
  if (!configForm.category_id) { message.warning('请填写分类ID'); return; }
  if (configForm.amount <= 0) { message.warning('质押金额必须大于0'); return; }
  configSaving.value = true;
  try {
    await savePledgeConfigApi({ ...configForm });
    message.success('保存成功');
    configModalVisible.value = false;
    loadConfigs();
  } catch (e: any) { message.error(e?.message || '保存失败'); }
  finally { configSaving.value = false; }
}

async function handleDeleteConfig(id: number) {
  try {
    await deletePledgeConfigApi(id);
    message.success('删除成功');
    loadConfigs();
  } catch (e: any) { message.error(e?.message || '删除失败'); }
}

async function handleToggle(id: number, checked: boolean) {
  try {
    await togglePledgeConfigApi(id, checked ? 1 : 0);
    message.success('更新成功');
    loadConfigs();
  } catch (e: any) { message.error(e?.message || '操作失败'); }
}

// ===== 质押记录 =====
const recordLoading = ref(false);
const records = ref<PledgeRecord[]>([]);
const recordPagination = reactive({ page: 1, limit: 20, total: 0 });

const recordColumns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 60, align: 'center' as const },
  { title: '用户', dataIndex: 'username', key: 'username', width: 120 },
  { title: '分类', dataIndex: 'category_name', key: 'category_name', width: 140 },
  { title: '质押金额', dataIndex: 'amount', key: 'amount', width: 110, align: 'center' as const },
  { title: '折扣率', dataIndex: 'discount_rate', key: 'discount_rate', width: 100, align: 'center' as const },
  { title: '天数', dataIndex: 'days', key: 'days', width: 80, align: 'center' as const },
  { title: '状态', key: 'status', width: 100, align: 'center' as const },
  { title: '质押时间', dataIndex: 'addtime', key: 'addtime', width: 170 },
  { title: '退还时间', dataIndex: 'endtime', key: 'endtime', width: 170 },
];

async function loadRecords(page = 1) {
  recordLoading.value = true;
  recordPagination.page = page;
  try {
    const res = await getPledgeRecordListApi({ page, limit: recordPagination.limit });
    records.value = res.list || [];
    recordPagination.total = res.pagination?.total || 0;
  } catch (e) { console.error(e); }
  finally { recordLoading.value = false; }
}

function onTabChange(key: string) {
  if (key === 'records' && records.value.length === 0) loadRecords();
}

onMounted(() => loadConfigs());
</script>

<template>
  <Page title="质押管理" content-class="p-4">
    <Card>
      <Tabs v-model:activeKey="activeTab" @change="onTabChange">
        <TabPane key="configs" tab="质押配置">
          <div class="mb-3 flex justify-end">
            <Button type="primary" size="small" @click="openAddConfig">
              <template #icon><PlusOutlined /></template>
              新增配置
            </Button>
          </div>

          <Table
            :columns="configColumns" :data-source="configs" :loading="configLoading"
            :pagination="false" row-key="id" size="small" bordered
            :scroll="{ x: 900 }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'cancel_fee'">
                {{ (record.cancel_fee * 100).toFixed(0) }}%
              </template>
              <template v-else-if="column.key === 'status'">
                <Switch
                  :checked="record.status === 1" size="small"
                  checked-children="启用" un-checked-children="禁用"
                  @change="(c: boolean) => handleToggle(record.id, c)"
                />
              </template>
              <template v-else-if="column.key === 'action'">
                <Space size="small">
                  <Button type="primary" size="small" @click="openEditConfig(record)">
                    <template #icon><EditOutlined /></template>
                  </Button>
                  <Popconfirm title="确定删除？" @confirm="handleDeleteConfig(record.id)">
                    <Button danger size="small">
                      <template #icon><DeleteOutlined /></template>
                    </Button>
                  </Popconfirm>
                </Space>
              </template>
            </template>
          </Table>
        </TabPane>

        <TabPane key="records" tab="质押记录">
          <Table
            :columns="recordColumns" :data-source="records" :loading="recordLoading"
            :pagination="false" row-key="id" size="small" bordered
            :scroll="{ x: 1100 }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'status'">
                <Tag :color="record.status === 1 ? 'processing' : 'default'">
                  {{ record.status === 1 ? '生效中' : '已退还' }}
                </Tag>
              </template>
            </template>
          </Table>

          <div class="flex justify-end mt-4" v-if="recordPagination.total > recordPagination.limit">
            <Pagination
              :current="recordPagination.page" :page-size="recordPagination.limit"
              :total="recordPagination.total"
              @change="(p: number) => loadRecords(p)"
            />
          </div>
        </TabPane>
      </Tabs>
    </Card>

    <!-- 配置弹窗 -->
    <Modal v-model:open="configModalVisible" :title="configForm.id ? '编辑质押配置' : '新增质押配置'"
           @ok="handleSaveConfig" :confirm-loading="configSaving" ok-text="保存" cancel-text="取消"
           :width="480" style="max-width: 95vw">
      <div class="space-y-4 py-2">
        <div>
          <label class="block text-sm font-medium mb-1">分类 ID</label>
          <InputNumber v-model:value="configForm.category_id" :min="1" style="width: 100%" placeholder="对应分类表的ID" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">质押金额（元）</label>
          <InputNumber v-model:value="configForm.amount" :min="0" :precision="2" style="width: 100%" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">折扣率</label>
          <InputNumber v-model:value="configForm.discount_rate" :min="0" :max="1" :step="0.01" :precision="2" style="width: 100%" placeholder="例如 0.85 表示85折" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">质押天数</label>
          <InputNumber v-model:value="configForm.days" :min="1" style="width: 100%" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">提前取消扣费比例</label>
          <InputNumber v-model:value="configForm.cancel_fee" :min="0" :max="1" :step="0.05" :precision="2" style="width: 100%" placeholder="例如 0.1 表示扣10%" />
        </div>
      </div>
    </Modal>
  </Page>
</template>

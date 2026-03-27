<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Button, Card, Input, Modal, Pagination, Select, SelectOption, Space, Spin, Table, Tag, message,
} from 'ant-design-vue';
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons-vue';
import { requestClient } from '#/api/request';

interface WithdrawItem {
  id: number;
  uid: number;
  username: string;
  amount: number;
  method: string;
  account_name: string;
  account_no: string;
  bank_name: string;
  note: string;
  status: number;
  audit_remark: string;
  audit_uid: number;
  audit_user: string;
  addtime: string;
  audit_time: string;
}

const loading = ref(false);
const reviewing = ref(false);
const list = ref<WithdrawItem[]>([]);
const pagination = reactive({ page: 1, limit: 20, total: 0 });
const filterUid = ref('');
const filterStatus = ref('');
const reviewVisible = ref(false);
const reviewStatus = ref<1 | -1>(1);
const reviewRemark = ref('');
const currentRow = ref<WithdrawItem | null>(null);

const columns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 70 },
  { title: 'UID', dataIndex: 'uid', key: 'uid', width: 70 },
  { title: '用户', dataIndex: 'username', key: 'username', width: 120 },
  { title: '金额', dataIndex: 'amount', key: 'amount', width: 110 },
  { title: '收款信息', key: 'account', ellipsis: true },
  { title: '状态', dataIndex: 'status', key: 'status', width: 100 },
  { title: '备注', dataIndex: 'note', key: 'note', ellipsis: true },
  { title: '申请时间', dataIndex: 'addtime', key: 'addtime', width: 170 },
  { title: '审核信息', key: 'audit', ellipsis: true },
  { title: '操作', key: 'action', width: 160, fixed: 'right' },
];

function statusMeta(status: number) {
  switch (status) {
    case 1:
      return { text: '已通过', color: 'green' };
    case -1:
      return { text: '已驳回', color: 'red' };
    default:
      return { text: '待审核', color: 'orange' };
  }
}

async function loadData(page = 1) {
  loading.value = true;
  pagination.page = page;
  try {
    const res = await requestClient.get<any>('/admin/withdraw/requests', {
      params: {
        page: pagination.page,
        limit: pagination.limit,
        uid: filterUid.value || undefined,
        status: filterStatus.value || undefined,
      },
    });
    list.value = res.list || [];
    pagination.total = res.pagination?.total || 0;
  } catch (e: any) {
    message.error(e?.message || '加载提现申请失败');
  } finally {
    loading.value = false;
  }
}

function handleReset() {
  filterUid.value = '';
  filterStatus.value = '';
  void loadData(1);
}

function openReview(row: WithdrawItem, status: 1 | -1) {
  currentRow.value = row;
  reviewStatus.value = status;
  reviewRemark.value = '';
  reviewVisible.value = true;
}

async function submitReview() {
  if (!currentRow.value) return;
  reviewing.value = true;
  try {
    await requestClient.post(`/admin/withdraw/${currentRow.value.id}/review`, {
      status: reviewStatus.value,
      remark: reviewRemark.value,
    });
    message.success('审核完成');
    reviewVisible.value = false;
    await loadData(pagination.page);
  } catch (e: any) {
    message.error(e?.message || '审核失败');
  } finally {
    reviewing.value = false;
  }
}

onMounted(() => {
  void loadData(1);
});
</script>

<template>
  <Page title="商家商城提现审核" content-class="p-4">
    <Card>
      <div class="mb-4 flex flex-wrap items-center gap-3">
        <Space wrap>
          <Input v-model:value="filterUid" placeholder="用户UID" allow-clear style="max-width: 120px; min-width: 80px" @pressEnter="() => loadData(1)" />
          <Select v-model:value="filterStatus" placeholder="状态" allow-clear style="max-width: 120px; min-width: 90px">
            <SelectOption value="">全部</SelectOption>
            <SelectOption value="0">待审核</SelectOption>
            <SelectOption value="1">已通过</SelectOption>
            <SelectOption value="-1">已驳回</SelectOption>
          </Select>
          <Button type="primary" @click="() => loadData(1)">
            <template #icon><SearchOutlined /></template>搜索
          </Button>
          <Button @click="handleReset">
            <template #icon><ReloadOutlined /></template>重置
          </Button>
        </Space>
      </div>

      <Spin :spinning="loading">
        <Table
          :data-source="list"
          :columns="columns"
          :pagination="false"
          row-key="id"
          size="small"
          bordered
          :scroll="{ x: 1200 }"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'amount'">
              <span class="font-medium">¥{{ Number(record.amount).toFixed(2) }}</span>
            </template>
            <template v-else-if="column.key === 'account'">
              <div class="text-xs leading-5">
                <div>{{ record.account_name }} / {{ record.account_no }}</div>
                <div class="text-gray-400">{{ record.bank_name || record.method || '-' }}</div>
              </div>
            </template>
            <template v-else-if="column.key === 'status'">
              <Tag :color="statusMeta(record.status).color">{{ statusMeta(record.status).text }}</Tag>
            </template>
            <template v-else-if="column.key === 'audit'">
              <div class="text-xs leading-5">
                <div>{{ record.audit_user || '-' }}</div>
                <div class="text-gray-400">{{ record.audit_time || '-' }}</div>
                <div class="text-gray-400">{{ record.audit_remark || '-' }}</div>
              </div>
            </template>
            <template v-else-if="column.key === 'action'">
              <Space v-if="record.status === 0">
                <Button type="link" size="small" @click="openReview(record, 1)">通过</Button>
                <Button type="link" size="small" danger @click="openReview(record, -1)">驳回</Button>
              </Space>
              <span v-else class="text-gray-400">已处理</span>
            </template>
          </template>
        </Table>

        <div class="mt-4 flex justify-end" v-if="pagination.total > pagination.limit">
          <Pagination
            :current="pagination.page"
            :total="pagination.total"
            :page-size="pagination.limit"
            show-size-changer
            @change="(p: number) => loadData(p)"
          />
        </div>
      </Spin>
    </Card>

    <Modal
      v-model:open="reviewVisible"
      :title="reviewStatus === 1 ? '通过提现申请' : '驳回提现申请'"
      :confirm-loading="reviewing"
      ok-text="提交"
      @ok="submitReview"
    >
      <div v-if="currentRow" class="space-y-4">
        <div class="rounded border border-gray-200 p-3 text-sm leading-6">
          <div>用户：{{ currentRow.username }}（UID {{ currentRow.uid }}）</div>
          <div>金额：¥{{ Number(currentRow.amount).toFixed(2) }}</div>
          <div>收款：{{ currentRow.account_name }} / {{ currentRow.account_no }}</div>
          <div>渠道：{{ currentRow.bank_name || currentRow.method || '-' }}</div>
        </div>
        <Input.TextArea
          v-model:value="reviewRemark"
          :rows="4"
          :placeholder="reviewStatus === 1 ? '可填写打款备注' : '请填写驳回原因'"
        />
      </div>
    </Modal>
  </Page>
</template>

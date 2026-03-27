<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Alert, Button, Card, Descriptions, DescriptionsItem, Form, FormItem, Input, InputNumber,
  Modal, Pagination, Select, SelectOption, Space, Spin, Table, Tag, message,
} from 'ant-design-vue';
import {
  createWithdrawRequestApi,
  getUserProfileApi,
  getWithdrawRequestsApi,
  type UserProfile,
  type WithdrawRequestItem,
} from '#/api/user-center';

const loading = ref(false);
const submitting = ref(false);
const profile = ref<UserProfile | null>(null);
const list = ref<WithdrawRequestItem[]>([]);
const pagination = reactive({ page: 1, limit: 20, total: 0 });
const statusFilter = ref<number | undefined>(undefined);
const modalVisible = ref(false);
const form = reactive({
  amount: undefined as number | undefined,
  method: 'manual',
  account_name: '',
  account_no: '',
  bank_name: '',
  note: '',
});

const availableBalance = computed(() => Number(profile.value?.mall_money || 0));
const frozenBalance = computed(() => Number(profile.value?.mall_cdmoney || 0));

const columns = [
  { title: '申请时间', dataIndex: 'addtime', key: 'addtime', width: 170 },
  { title: '金额', dataIndex: 'amount', key: 'amount', width: 110 },
  { title: '收款信息', key: 'account', ellipsis: true },
  { title: '状态', dataIndex: 'status', key: 'status', width: 100 },
  { title: '审核备注', dataIndex: 'audit_remark', key: 'audit_remark', ellipsis: true },
  { title: '审核时间', dataIndex: 'audit_time', key: 'audit_time', width: 170 },
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

function resetForm() {
  form.amount = undefined;
  form.method = 'manual';
  form.account_name = '';
  form.account_no = '';
  form.bank_name = '';
  form.note = '';
}

async function loadProfile() {
  profile.value = await getUserProfileApi();
}

async function loadList(page = 1) {
  loading.value = true;
  pagination.page = page;
  try {
    const res = await getWithdrawRequestsApi({
      page: pagination.page,
      limit: pagination.limit,
      status: statusFilter.value,
    });
    list.value = res.list || [];
    pagination.total = res.pagination?.total || 0;
  } finally {
    loading.value = false;
  }
}

async function loadAll(page = pagination.page) {
  await Promise.all([loadProfile(), loadList(page)]);
}

async function handleSubmit() {
  if (!form.amount || form.amount <= 0) {
    message.warning('请输入有效提现金额');
    return;
  }
  if (!form.account_name.trim() || !form.account_no.trim()) {
    message.warning('请填写完整收款信息');
    return;
  }
  submitting.value = true;
  try {
    await createWithdrawRequestApi({
      amount: form.amount,
      method: form.method,
      account_name: form.account_name,
      account_no: form.account_no,
      bank_name: form.bank_name,
      note: form.note,
    });
    message.success('提现申请已提交');
    modalVisible.value = false;
    resetForm();
    await loadAll(1);
  } catch (e: any) {
    message.error(e?.message || '提现申请提交失败');
  } finally {
    submitting.value = false;
  }
}

onMounted(() => {
  void loadAll(1);
});
</script>

<template>
  <Page title="提现中心" content-class="p-4">
    <div class="space-y-4">
      <Alert
        type="info"
        show-icon
        message="商城代收款会进入独立商城钱包。提现申请提交后，金额会先冻结到商城钱包冻结余额，等待后台审核。"
      />

      <Card>
        <Descriptions bordered :column="{ xs: 1, sm: 2 }" size="small">
          <DescriptionsItem label="可提现余额">
            <span class="text-green-600 font-semibold">¥{{ availableBalance.toFixed(2) }}</span>
          </DescriptionsItem>
          <DescriptionsItem label="冻结中余额">
            <span class="text-orange-500 font-semibold">¥{{ frozenBalance.toFixed(2) }}</span>
          </DescriptionsItem>
        </Descriptions>
        <div class="mt-4">
          <Button type="primary" @click="modalVisible = true">申请提现</Button>
        </div>
      </Card>

      <Card>
        <div class="mb-4 flex flex-wrap items-center gap-3">
          <Space wrap>
            <span class="text-sm text-gray-500">状态：</span>
            <Select
              v-model:value="statusFilter"
              placeholder="全部"
              allow-clear
              style="max-width: 140px; min-width: 100px"
              @change="() => loadList(1)"
            >
              <SelectOption :value="undefined">全部</SelectOption>
              <SelectOption :value="0">待审核</SelectOption>
              <SelectOption :value="1">已通过</SelectOption>
              <SelectOption :value="-1">已驳回</SelectOption>
            </Select>
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
            :scroll="{ x: 900 }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'amount'">
                <span class="font-medium">¥{{ Number(record.amount).toFixed(2) }}</span>
              </template>
              <template v-else-if="column.key === 'account'">
                <div class="text-xs leading-5">
                  <div>{{ record.account_name }} / {{ record.account_no }}</div>
                  <div class="text-gray-400">{{ record.bank_name || '未填写渠道' }}</div>
                </div>
              </template>
              <template v-else-if="column.key === 'status'">
                <Tag :color="statusMeta(record.status).color">{{ statusMeta(record.status).text }}</Tag>
              </template>
              <template v-else-if="column.key === 'audit_remark'">
                {{ record.audit_remark || '-' }}
              </template>
              <template v-else-if="column.key === 'audit_time'">
                {{ record.audit_time || '-' }}
              </template>
            </template>
          </Table>

          <div class="mt-4 flex justify-end" v-if="pagination.total > pagination.limit">
            <Pagination
              :current="pagination.page"
              :total="pagination.total"
              :page-size="pagination.limit"
              show-size-changer
              @change="(p: number) => loadList(p)"
            />
          </div>
        </Spin>
      </Card>
    </div>

    <Modal
      v-model:open="modalVisible"
      title="申请提现"
      ok-text="提交申请"
      :confirm-loading="submitting"
      @ok="handleSubmit"
      @cancel="resetForm"
    >
      <Form layout="vertical">
        <FormItem label="提现金额">
          <InputNumber v-model:value="form.amount" :min="0.01" :precision="2" style="width: 100%" />
        </FormItem>
        <FormItem label="提现方式">
          <Select v-model:value="form.method">
            <SelectOption value="manual">人工转账</SelectOption>
            <SelectOption value="bank">银行卡</SelectOption>
            <SelectOption value="alipay">支付宝</SelectOption>
            <SelectOption value="wechat">微信</SelectOption>
          </Select>
        </FormItem>
        <FormItem label="收款人">
          <Input v-model:value="form.account_name" placeholder="输入收款人姓名" />
        </FormItem>
        <FormItem label="收款账号">
          <Input v-model:value="form.account_no" placeholder="输入银行卡号/支付宝账号/微信号" />
        </FormItem>
        <FormItem label="开户行/渠道">
          <Input v-model:value="form.bank_name" placeholder="如招商银行 / 支付宝 / 微信" />
        </FormItem>
        <FormItem label="备注">
          <Input v-model:value="form.note" placeholder="选填" />
        </FormItem>
      </Form>
    </Modal>
  </Page>
</template>

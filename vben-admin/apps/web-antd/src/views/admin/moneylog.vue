<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Tag, Select, SelectOption, Input, Button, Space, Spin, Pagination,
} from 'ant-design-vue';
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons-vue';
import { requestClient } from '#/api/request';

interface AdminMoneyLog {
  id: number;
  uid: number;
  username: string;
  type: string;
  money: number;
  balance: number;
  remark: string;
  addtime: string;
}

const loading = ref(false);
const list = ref<AdminMoneyLog[]>([]);
const pagination = reactive({ page: 1, limit: 20, total: 0 });
const filterUid = ref('');
const filterType = ref('');

const columns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
  { title: 'UID', dataIndex: 'uid', key: 'uid', width: 60 },
  { title: '用户', dataIndex: 'username', key: 'username', width: 120 },
  { title: '类型', dataIndex: 'type', key: 'type', width: 80 },
  { title: '金额', dataIndex: 'money', key: 'money', width: 100 },
  { title: '变动后余额', dataIndex: 'balance', key: 'balance', width: 110 },
  { title: '备注', dataIndex: 'remark', key: 'remark', ellipsis: true },
  { title: '时间', dataIndex: 'addtime', key: 'addtime', width: 170 },
];

async function loadData(page = 1) {
  loading.value = true;
  pagination.page = page;
  try {
    const raw = await requestClient.get<any>('/admin/moneylog', {
      params: {
        page: pagination.page,
        limit: pagination.limit,
        uid: filterUid.value || undefined,
        type: filterType.value || undefined,
      },
    });
    const res = raw;
    list.value = res.list || [];
    pagination.total = res.pagination?.total || 0;
  } catch (e) {
    console.error('加载流水失败:', e);
  } finally {
    loading.value = false;
  }
}

function handleSearch() { loadData(1); }
function handleReset() {
  filterUid.value = '';
  filterType.value = '';
  loadData(1);
}

onMounted(() => loadData(1));
</script>

<template>
  <Page title="余额流水" content-class="p-4">
    <Card>
      <div class="mb-4 flex flex-wrap items-center gap-3">
        <Space wrap>
          <Input v-model:value="filterUid" placeholder="用户UID" allow-clear style="max-width: 120px; min-width: 80px" @pressEnter="handleSearch" />
          <Select v-model:value="filterType" placeholder="类型" allow-clear style="max-width: 120px; min-width: 80px">
            <SelectOption value="">全部</SelectOption>
            <SelectOption value="扣费">扣费</SelectOption>
            <SelectOption value="充值">充值</SelectOption>
            <SelectOption value="退款">退款</SelectOption>
            <SelectOption value="调整">调整</SelectOption>
            <SelectOption value="商城代收">商城代收</SelectOption>
            <SelectOption value="提现申请">提现申请</SelectOption>
            <SelectOption value="提现通过">提现通过</SelectOption>
            <SelectOption value="提现驳回">提现驳回</SelectOption>
          </Select>
          <Button type="primary" @click="handleSearch">
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
          :scroll="{ x: 800 }"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'type'">
              <Tag :color="record.type === '充值' || record.type === '商城代收' || record.type === '提现驳回'
                ? 'green'
                : record.type === '退款'
                  ? 'blue'
                  : record.type === '扣费' || record.type === '提现申请'
                    ? 'red'
                    : record.type === '调整' || record.type === '提现通过'
                      ? 'orange'
                      : 'default'">
                {{ record.type }}
              </Tag>
            </template>
            <template v-if="column.key === 'money'">
              <span :class="record.money >= 0 ? 'text-green-600' : 'text-red-500'" class="font-medium">
                {{ record.money >= 0 ? '+' : '' }}{{ Number(record.money).toFixed(2) }}
              </span>
            </template>
            <template v-if="column.key === 'balance'">
              <span class="font-medium">¥{{ Number(record.balance).toFixed(2) }}</span>
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
  </Page>
</template>

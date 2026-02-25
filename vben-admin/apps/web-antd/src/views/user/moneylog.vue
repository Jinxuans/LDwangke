<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Tag, Select, SelectOption, Space, Spin, Pagination,
} from 'ant-design-vue';
import { getMoneyLogApi, type MoneyLog } from '#/api/user-center';

const loading = ref(false);
const list = ref<MoneyLog[]>([]);
const total = ref(0);
const page = ref(1);
const limit = ref(20);
const typeFilter = ref('');

const columns = [
  { title: '时间', dataIndex: 'addtime', key: 'addtime', width: 170 },
  { title: '类型', dataIndex: 'type', key: 'type', width: 100 },
  { title: '金额', dataIndex: 'money', key: 'money', width: 120 },
  { title: '变动后余额', dataIndex: 'balance', key: 'balance', width: 120 },
  { title: '备注', dataIndex: 'remark', key: 'remark', ellipsis: true },
];

async function loadData() {
  loading.value = true;
  try {
    const res = await getMoneyLogApi({
      page: page.value,
      limit: limit.value,
      type: typeFilter.value || undefined,
    });
    list.value = res.list || [];
    total.value = res.pagination?.total || 0;
  } catch (e) {
    console.error('加载流水失败:', e);
  } finally {
    loading.value = false;
  }
}

function handlePageChange(p: number) {
  page.value = p;
  loadData();
}

function handleFilterChange() {
  page.value = 1;
  loadData();
}

onMounted(loadData);
</script>

<template>
  <Page title="余额流水" content-class="p-4">
    <Card>
      <div class="mb-4 flex flex-wrap items-center gap-3">
        <Space wrap>
          <span class="text-sm text-gray-500">类型筛选：</span>
          <Select
            v-model:value="typeFilter"
            placeholder="全部"
            allow-clear
            style="max-width: 140px; min-width: 100px"
            @change="handleFilterChange"
          >
            <SelectOption value="">全部</SelectOption>
            <SelectOption value="扣费">扣费</SelectOption>
            <SelectOption value="充值">充值</SelectOption>
            <SelectOption value="退款">退款</SelectOption>
            <SelectOption value="调整">调整</SelectOption>
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
          :scroll="{ x: 700 }"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'type'">
              <Tag
                :color="record.type === '充值' ? 'green' : record.type === '退款' ? 'blue' : record.type === '扣费' ? 'red' : 'default'"
              >
                {{ record.type }}
              </Tag>
            </template>
            <template v-if="column.key === 'money'">
              <span :class="record.money >= 0 ? 'text-green-600' : 'text-red-500'" class="font-medium">
                {{ record.money >= 0 ? '+' : '' }}{{ record.money.toFixed(2) }}
              </span>
            </template>
            <template v-if="column.key === 'balance'">
              <span class="font-medium">¥{{ record.balance.toFixed(2) }}</span>
            </template>
          </template>
        </Table>

        <div class="mt-4 flex justify-end" v-if="total > limit">
          <Pagination
            :current="page"
            :total="total"
            :page-size="limit"
            show-size-changer
            @change="handlePageChange"
          />
        </div>
      </Spin>
    </Card>
  </Page>
</template>

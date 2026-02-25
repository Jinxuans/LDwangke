<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Tag, Select, SelectOption, Input, Button, Space, Spin, Pagination,
} from 'ant-design-vue';
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons-vue';
import { getLogListApi, type LogItem } from '#/api/user-center';

const loading = ref(false);
const list = ref<LogItem[]>([]);
const pagination = reactive({ page: 1, limit: 20, total: 0 });
const filterType = ref('');
const filterKeywords = ref('');

const columns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
  { title: 'UID', dataIndex: 'uid', key: 'uid', width: 60 },
  { title: '类型', dataIndex: 'type', key: 'type', width: 100 },
  { title: '详情', dataIndex: 'text', key: 'text', ellipsis: true },
  { title: '金额', dataIndex: 'money', key: 'money', width: 90 },
  { title: '余额', dataIndex: 'smoney', key: 'smoney', width: 90 },
  { title: 'IP', dataIndex: 'ip', key: 'ip', width: 130 },
  { title: '时间', dataIndex: 'addtime', key: 'addtime', width: 170 },
];

async function loadData(page = 1) {
  loading.value = true;
  pagination.page = page;
  try {
    const res = await getLogListApi({
      page: pagination.page,
      limit: pagination.limit,
      type: filterType.value || undefined,
      keywords: filterKeywords.value || undefined,
    });
    list.value = res.list || [];
    pagination.total = res.pagination?.total || 0;
  } catch (e) {
    console.error('加载日志失败:', e);
  } finally {
    loading.value = false;
  }
}

function handleSearch() { loadData(1); }
function handleReset() {
  filterType.value = '';
  filterKeywords.value = '';
  loadData(1);
}

onMounted(() => loadData(1));
</script>

<template>
  <Page title="操作日志" content-class="p-4">
    <Card>
      <div class="mb-4 flex items-center gap-2 flex-wrap">
        <Select v-model:value="filterType" placeholder="搜索类型" allow-clear style="max-width: 120px; min-width: 80px">
          <SelectOption value="">全部</SelectOption>
          <SelectOption value="uid">UID</SelectOption>
          <SelectOption value="type">类型</SelectOption>
          <SelectOption value="text">内容</SelectOption>
          <SelectOption value="money">金额</SelectOption>
          <SelectOption value="ip">IP</SelectOption>
        </Select>
        <Input v-model:value="filterKeywords" placeholder="关键词" allow-clear style="max-width: 200px; min-width: 100px" @pressEnter="handleSearch" />
        <Button type="primary" @click="handleSearch"><template #icon><SearchOutlined /></template>搜索</Button>
        <Button @click="handleReset"><template #icon><ReloadOutlined /></template>重置</Button>
      </div>

      <Spin :spinning="loading">
        <Table :data-source="list" :columns="columns" :pagination="false" row-key="id" size="small" bordered :scroll="{ x: 850 }">
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'type'">
              <Tag>{{ record.type }}</Tag>
            </template>
            <template v-if="column.key === 'money'">
              <span :class="record.money >= 0 ? 'text-green-600' : 'text-red-500'" class="font-medium">
                {{ record.money >= 0 ? '+' : '' }}{{ Number(record.money).toFixed(2) }}
              </span>
            </template>
            <template v-if="column.key === 'smoney'">
              <span class="font-medium">¥{{ Number(record.smoney).toFixed(2) }}</span>
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

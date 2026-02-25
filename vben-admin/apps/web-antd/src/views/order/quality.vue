<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Table, Input, Select, SelectOption, Button, Tag, Progress,
  Pagination, Space, Card, Row, Col,
} from 'ant-design-vue';
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons-vue';
import {
  getOrderListApi,
  type OrderItem,
  type OrderListParams,
} from '#/api/order';

const loading = ref(false);
const tableData = ref<OrderItem[]>([]);
const pagination = reactive({ page: 1, limit: 20, total: 0 });
const search = reactive<OrderListParams>({ kcname: '', cid: '', status_text: '' });

function statusColor(s: string) {
  if (s === '已完成' || s === '已上号') return 'green';
  if (s === '进行中' || s === '刷课中') return 'blue';
  if (s === '异常' || s === '补刷中') return 'red';
  if (s === '待处理') return 'orange';
  return 'default';
}

function progressPercent(val: string): number {
  if (!val) return 0;
  const n = parseFloat(val.replace('%', ''));
  return isNaN(n) ? 0 : Math.min(100, Math.max(0, n));
}

async function loadData(page = 1) {
  loading.value = true;
  pagination.page = page;
  try {
    const raw = await getOrderListApi({
      ...search,
      page: pagination.page,
      limit: pagination.limit,
    });
    const res = raw;
    tableData.value = res.list || [];
    pagination.total = res.pagination?.total || 0;
  } catch (e) {
    console.error('加载失败:', e);
  } finally {
    loading.value = false;
  }
}

function handleSearch() { loadData(1); }
function handleReset() {
  Object.assign(search, { kcname: '', cid: '', status_text: '' });
  loadData(1);
}

const columns = [
  { title: '平台名称', dataIndex: 'ptname', key: 'ptname', width: 180, ellipsis: true },
  { title: '课程名称', dataIndex: 'kcname', key: 'kcname', width: 220, ellipsis: true },
  { title: '任务状态', dataIndex: 'status', key: 'status', width: 100 },
  { title: '进度', dataIndex: 'process', key: 'process', width: 150 },
  { title: '详情/考试状态', dataIndex: 'remarks', key: 'remarks', width: 200, ellipsis: true },
  { title: '提交时间', dataIndex: 'addtime', key: 'addtime', width: 160 },
];

onMounted(() => loadData(1));
</script>

<template>
  <Page title="质量查询" content-class="p-4">
    <Card class="mb-4">
      <Row :gutter="[16, 12]">
        <Col :xs="24" :sm="12" :md="6">
          <Input v-model:value="search.kcname" placeholder="课程名称" allow-clear @pressEnter="handleSearch" />
        </Col>
        <Col :xs="24" :sm="12" :md="6">
          <Select v-model:value="search.status_text" placeholder="任务状态" allow-clear style="width: 100%">
            <SelectOption v-for="s in ['待处理','进行中','已完成','异常','已取消','补刷中','出错啦']" :key="s" :value="s">{{ s }}</SelectOption>
          </Select>
        </Col>
        <Col :xs="24" :sm="12" :md="6">
          <Space>
            <Button type="primary" @click="handleSearch"><template #icon><SearchOutlined /></template>搜索</Button>
            <Button @click="handleReset"><template #icon><ReloadOutlined /></template>重置</Button>
          </Space>
        </Col>
      </Row>
    </Card>

    <Card>
      <Table
        :columns="columns"
        :data-source="tableData"
        :loading="loading"
        :pagination="false"
        row-key="oid"
        :scroll="{ x: 1000 }"
        size="small"
        bordered
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <Tag :color="statusColor(record.status)">{{ record.status || '待处理' }}</Tag>
          </template>
          <template v-else-if="column.key === 'process'">
            <Progress :percent="progressPercent(record.process)" size="small" />
          </template>
        </template>
      </Table>

      <div class="flex justify-center mt-4">
        <Pagination
          v-model:current="pagination.page"
          :total="pagination.total"
          :page-size="pagination.limit"
          :page-size-options="['20', '50', '100']"
          show-size-changer
          :show-total="(total: number) => `共 ${total} 条`"
          @change="(p: number) => loadData(p)"
          @showSizeChange="(_: number, s: number) => { pagination.limit = s; loadData(1); }"
        />
      </div>
    </Card>
  </Page>
</template>

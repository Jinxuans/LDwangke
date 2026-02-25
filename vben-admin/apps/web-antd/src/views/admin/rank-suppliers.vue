<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import { Card, Table, Spin, Button } from 'ant-design-vue';
import { ReloadOutlined } from '@ant-design/icons-vue';
import { getSupplierRankingApi, type SupplierRankItem } from '#/api/admin';

const loading = ref(false);
const list = ref<SupplierRankItem[]>([]);

const columns = [
  { title: '货源名称', dataIndex: 'name', key: 'name' },
  { title: '今日销量', dataIndex: 'today_count', key: 'today_count', align: 'center' as const, sorter: (a: SupplierRankItem, b: SupplierRankItem) => a.today_count - b.today_count },
  { title: '昨日销量', dataIndex: 'yesterday_count', key: 'yesterday_count', align: 'center' as const, sorter: (a: SupplierRankItem, b: SupplierRankItem) => a.yesterday_count - b.yesterday_count },
  { title: '总销量', dataIndex: 'total_count', key: 'total_count', align: 'center' as const, defaultSortOrder: 'descend' as const, sorter: (a: SupplierRankItem, b: SupplierRankItem) => a.total_count - b.total_count },
];

async function fetchData() {
  loading.value = true;
  try {
    const raw = await getSupplierRankingApi();
    let data = raw;
    if (!Array.isArray(data)) data = [];
    list.value = data;
  } catch (e) {
    console.error(e);
    list.value = [];
  } finally {
    loading.value = false;
  }
}

onMounted(fetchData);
</script>

<template>
  <Page title="货源排行" content-class="p-4">
    <Card>
      <template #title>
        <span>货源排行榜</span>
      </template>
      <template #extra>
        <Button @click="fetchData" :loading="loading">
          <template #icon><ReloadOutlined /></template>
          刷新
        </Button>
      </template>
      <Spin :spinning="loading">
        <Table
          :columns="columns"
          :data-source="list"
          :pagination="false"
          row-key="hid"
          size="middle"
        />
      </Spin>
    </Card>
  </Page>
</template>

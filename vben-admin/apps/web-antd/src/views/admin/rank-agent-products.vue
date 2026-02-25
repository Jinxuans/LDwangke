<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import { Card, Table, Spin, Button, InputNumber, Radio, Tag } from 'ant-design-vue';
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons-vue';
import { getAgentProductRankingApi, type AgentProductRankItem } from '#/api/admin';

const loading = ref(false);
const uid = ref<number | null>(null);
const timeType = ref('today');
const list = ref<AgentProductRankItem[]>([]);
const searched = ref(false);

const columns = [
  { title: '排名', dataIndex: 'rank', key: 'rank', width: 80, align: 'center' as const },
  { title: '商品名称', dataIndex: 'ptname', key: 'ptname', ellipsis: true },
  { title: '订单量', dataIndex: 'count', key: 'count', width: 120, align: 'center' as const },
  { title: '最后下单', dataIndex: 'latest', key: 'latest', width: 180, align: 'center' as const },
];

const timeOptions = [
  { label: '今日', value: 'today' },
  { label: '昨日', value: 'yesterday' },
  { label: '本周', value: 'week' },
  { label: '本月', value: 'month' },
];

async function fetchData() {
  if (!uid.value || uid.value <= 0) return;
  loading.value = true;
  searched.value = true;
  try {
    const raw = await getAgentProductRankingApi(uid.value, timeType.value);
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

function onTimeChange() {
  if (searched.value && uid.value && uid.value > 0) {
    fetchData();
  }
}
</script>

<template>
  <Page title="代理商品排行" content-class="p-4">
    <Card>
      <template #title>
        <span>代理商品排行</span>
      </template>
      <template #extra>
        <Button @click="fetchData" :loading="loading" :disabled="!uid || uid <= 0">
          <template #icon><ReloadOutlined /></template>
          刷新
        </Button>
      </template>

      <div class="mb-4 flex items-center gap-4 flex-wrap">
        <div class="flex items-center gap-2">
          <span class="font-semibold whitespace-nowrap">代理ID:</span>
          <InputNumber
            v-model:value="uid"
            placeholder="输入代理ID"
            :min="1"
            style="width: 160px"
            @press-enter="fetchData"
          />
          <Button type="primary" @click="fetchData" :loading="loading" :disabled="!uid || uid <= 0">
            <template #icon><SearchOutlined /></template>
            查询
          </Button>
        </div>

        <Radio.Group
          v-if="searched"
          v-model:value="timeType"
          button-style="solid"
          @change="onTimeChange"
        >
          <Radio.Button v-for="opt in timeOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </Radio.Button>
        </Radio.Group>
      </div>

      <Spin :spinning="loading">
        <template v-if="searched">
          <Table
            :columns="columns"
            :data-source="list"
            :pagination="false"
            row-key="rank"
            size="middle"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'rank'">
                <Tag :color="record.rank <= 3 ? 'gold' : 'default'">{{ record.rank }}</Tag>
              </template>
              <template v-if="column.key === 'count'">
                {{ record.count }} 单
              </template>
            </template>
            <template #emptyText>
              <div class="py-4 text-center text-gray-400">暂无订单数据</div>
            </template>
          </Table>
        </template>
        <template v-else>
          <div class="py-8 text-center text-gray-400">请输入代理ID进行查询</div>
        </template>
      </Spin>
    </Card>
  </Page>
</template>

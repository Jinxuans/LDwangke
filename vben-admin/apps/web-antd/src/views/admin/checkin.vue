<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import { Card, Table, DatePicker, Row, Col, Statistic } from 'ant-design-vue';
import { adminCheckinStatsApi, type CheckinStatsResult } from '#/api/plugins/checkin';
import dayjs, { type Dayjs } from 'dayjs';

const loading = ref(false);
const date = ref<Dayjs>(dayjs());
const stats = ref<CheckinStatsResult>({ total_users: 0, total_reward: 0, list: [], total: 0 });
const page = ref(1);
const limit = ref(20);

const columns = [
  { title: 'UID', dataIndex: 'uid', width: 80 },
  { title: '用户名', dataIndex: 'username' },
  { title: '奖励金额', dataIndex: 'reward_money', customRender: ({ text }: any) => `¥${text}` },
  { title: '签到时间', dataIndex: 'addtime' },
];

async function loadData() {
  loading.value = true;
  try {
    stats.value = await adminCheckinStatsApi({
      date: date.value.format('YYYY-MM-DD'),
      page: page.value,
      limit: limit.value,
    });
  } catch {} finally {
    loading.value = false;
  }
}

function onDateChange(val: Dayjs) {
  date.value = val;
  page.value = 1;
  loadData();
}

function onPageChange(p: number, size: number) {
  page.value = p;
  limit.value = size;
  loadData();
}

onMounted(loadData);
</script>

<template>
  <Page title="签到管理">
    <Card>
      <div style="margin-bottom: 16px">
        <DatePicker :value="date" @change="onDateChange" />
      </div>
      <Row :gutter="16" style="margin-bottom: 24px">
        <Col :span="12">
          <Statistic title="今日签到人数" :value="stats.total_users" suffix="人" />
        </Col>
        <Col :span="12">
          <Statistic title="今日发放奖励" :value="stats.total_reward" prefix="¥" :precision="2" />
        </Col>
      </Row>
      <Table
        :columns="columns"
        :data-source="stats.list"
        :loading="loading"
        row-key="uid"
        :pagination="{
          current: page,
          pageSize: limit,
          total: stats.total,
          showSizeChanger: true,
          onChange: onPageChange,
        }"
      />
    </Card>
  </Page>
</template>

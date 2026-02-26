<script setup lang="ts">
import { h, ref, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import { Card, Row, Col, Statistic, message } from 'ant-design-vue';
import { ShoppingOutlined, DollarOutlined, ClockCircleOutlined, CheckCircleOutlined } from '@ant-design/icons-vue';
import { getTenantOrderStatsApi, type TenantOrderStats } from '#/api/tenant';

const loading = ref(false);
const stats = ref<TenantOrderStats>({
  total: 0,
  today: 0,
  total_retail: 0,
  today_retail: 0,
  pending: 0,
  done: 0,
});

async function load() {
  loading.value = true;
  try {
    const res = await getTenantOrderStatsApi();
    stats.value = res;
  } catch (e: any) {
    message.error(e?.message || '加载失败');
  } finally {
    loading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <Page title="订单统计" content-class="p-4">
    <Row :gutter="[16, 16]">
      <Col :xs="24" :sm="12" :lg="6">
        <Card :loading="loading">
          <Statistic title="今日订单" :value="stats.today" :prefix="h(ShoppingOutlined)" :value-style="{ color: '#1677ff' }" />
        </Card>
      </Col>
      <Col :xs="24" :sm="12" :lg="6">
        <Card :loading="loading">
          <Statistic title="今日收入（元）" :value="stats.today_retail" :precision="2" :prefix="h(DollarOutlined)" :value-style="{ color: '#52c41a' }" />
        </Card>
      </Col>
      <Col :xs="24" :sm="12" :lg="6">
        <Card :loading="loading">
          <Statistic title="待处理订单" :value="stats.pending" :prefix="h(ClockCircleOutlined)" :value-style="{ color: '#faad14' }" />
        </Card>
      </Col>
      <Col :xs="24" :sm="12" :lg="6">
        <Card :loading="loading">
          <Statistic title="已完成订单" :value="stats.done" :prefix="h(CheckCircleOutlined)" :value-style="{ color: '#52c41a' }" />
        </Card>
      </Col>
      <Col :xs="24" :sm="12" :lg="6">
        <Card :loading="loading">
          <Statistic title="累计订单" :value="stats.total" :prefix="h(ShoppingOutlined)" />
        </Card>
      </Col>
      <Col :xs="24" :sm="12" :lg="6">
        <Card :loading="loading">
          <Statistic title="累计收入（元）" :value="stats.total_retail" :precision="2" :prefix="h(DollarOutlined)" />
        </Card>
      </Col>
    </Row>
  </Page>
</template>



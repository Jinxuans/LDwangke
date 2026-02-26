<script setup lang="ts">
import { ref } from 'vue';
import {
  Card, Input, Button, Table, Tag, Space, Empty, message,
} from 'ant-design-vue';
import { SearchOutlined } from '@ant-design/icons-vue';
import { checkOrderApi, type CheckOrderResult } from '#/api/auxiliary';

const searchUser = ref('');
const searchOid = ref('');
const loading = ref(false);
const list = ref<CheckOrderResult[]>([]);
const searched = ref(false);

const statusMap: Record<string, { text: string; color: string }> = {
  '等待中': { text: '等待中', color: 'default' },
  '进行中': { text: '进行中', color: 'processing' },
  '已完成': { text: '已完成', color: 'success' },
  '已退款': { text: '已退款', color: 'warning' },
  '异常': { text: '异常', color: 'error' },
};

const columns = [
  { title: '订单号', dataIndex: 'oid', key: 'oid', width: 90, align: 'center' as const },
  { title: '平台', dataIndex: 'ptname', key: 'ptname', width: 120, ellipsis: true },
  { title: '课程名称', dataIndex: 'kcname', key: 'kcname', ellipsis: true },
  { title: '状态', key: 'status', width: 100, align: 'center' as const },
  { title: '进度', dataIndex: 'process', key: 'process', width: 100, align: 'center' as const },
  { title: '备注', dataIndex: 'remarks', key: 'remarks', ellipsis: true, width: 160 },
  { title: '下单时间', dataIndex: 'addtime', key: 'addtime', width: 170 },
];

async function handleSearch() {
  const user = searchUser.value.trim();
  const oid = searchOid.value.trim();
  if (!user && !oid) {
    message.warning('请输入账号或订单号');
    return;
  }
  loading.value = true;
  searched.value = true;
  try {
    const res = await checkOrderApi({ user: user || undefined, oid: oid || undefined });
    list.value = res.list || [];
  } catch (e: any) {
    message.error(e?.message || '查询失败');
    list.value = [];
  } finally {
    loading.value = false;
  }
}

function getStatusInfo(status: string) {
  return statusMap[status] || { text: status, color: 'default' };
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 flex items-start justify-center pt-8 sm:pt-16 px-4">
    <div class="w-full max-w-3xl">
      <div class="text-center mb-6">
        <h1 class="text-2xl sm:text-3xl font-bold text-gray-800 mb-2">订单查询</h1>
        <p class="text-gray-500 text-sm">输入您的账号或订单号查询课程进度</p>
      </div>

      <Card class="mb-4 shadow-sm">
        <div class="flex flex-col sm:flex-row gap-3">
          <Input
            v-model:value="searchUser" placeholder="输入账号"
            allow-clear size="large" class="flex-1"
            @press-enter="handleSearch"
          />
          <Input
            v-model:value="searchOid" placeholder="或输入订单号"
            allow-clear size="large" class="flex-1"
            @press-enter="handleSearch"
          />
          <Button type="primary" size="large" :loading="loading" @click="handleSearch"
                  class="sm:w-auto w-full">
            <template #icon><SearchOutlined /></template>
            查询
          </Button>
        </div>
      </Card>

      <Card v-if="searched" class="shadow-sm">
        <template v-if="list.length === 0 && !loading">
          <Empty description="未找到相关订单" />
        </template>

        <!-- 移动端卡片布局 -->
        <div class="block sm:hidden space-y-3">
          <Card v-for="item in list" :key="item.oid" size="small" class="border">
            <div class="space-y-2 text-sm">
              <div class="flex justify-between">
                <span class="text-gray-500">订单号</span>
                <span class="font-mono font-medium">{{ item.oid }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500">平台</span>
                <span>{{ item.ptname }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500">课程</span>
                <span class="text-right max-w-[60%] truncate">{{ item.kcname }}</span>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-gray-500">状态</span>
                <Tag :color="getStatusInfo(item.status).color">{{ getStatusInfo(item.status).text }}</Tag>
              </div>
              <div class="flex justify-between" v-if="item.process">
                <span class="text-gray-500">进度</span>
                <span>{{ item.process }}</span>
              </div>
              <div class="flex justify-between" v-if="item.remarks">
                <span class="text-gray-500">备注</span>
                <span class="text-right max-w-[60%] truncate">{{ item.remarks }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500">时间</span>
                <span class="text-xs">{{ item.addtime }}</span>
              </div>
            </div>
          </Card>
        </div>

        <!-- 桌面端表格布局 -->
        <div class="hidden sm:block">
          <Table
            :columns="columns" :data-source="list" :loading="loading"
            :pagination="false" row-key="oid" size="small" bordered
            :scroll="{ x: 800 }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'status'">
                <Tag :color="getStatusInfo(record.status).color">
                  {{ getStatusInfo(record.status).text }}
                </Tag>
              </template>
            </template>
          </Table>
        </div>
      </Card>
    </div>
  </div>
</template>

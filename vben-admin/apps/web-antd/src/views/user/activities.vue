<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import { Card, Tag, Empty, Progress } from 'ant-design-vue';
import { GiftOutlined, TeamOutlined, ShoppingCartOutlined } from '@ant-design/icons-vue';
import { getPublicActivityListApi, type Activity } from '#/api/auxiliary';

const loading = ref(false);
const activities = ref<Activity[]>([]);

async function loadData() {
  loading.value = true;
  try {
    const res = await getPublicActivityListApi();
    activities.value = Array.isArray(res) ? res : [];
  } catch (e) { console.error(e); }
  finally { loading.value = false; }
}

function typeLabel(type: string) {
  return type === '1' ? '邀人活动' : '订单活动';
}

function typeColor(type: string) {
  return type === '1' ? 'blue' : 'green';
}

onMounted(loadData);
</script>

<template>
  <Page title="活动中心" content-class="p-4">
    <template v-if="activities.length === 0 && !loading">
      <Card>
        <Empty description="暂无进行中的活动" />
      </Card>
    </template>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <Card v-for="act in activities" :key="act.hid" hoverable size="small">
        <div class="flex items-center gap-2 mb-3">
          <component :is="act.type === '1' ? TeamOutlined : ShoppingCartOutlined"
                     class="text-xl text-blue-500" />
          <span class="text-base font-bold">{{ act.name }}</span>
          <Tag :color="typeColor(act.type)" class="ml-auto">{{ typeLabel(act.type) }}</Tag>
        </div>

        <div class="text-sm text-gray-600 mb-3">{{ act.yaoqiu }}</div>

        <div class="space-y-2 text-sm">
          <div class="flex justify-between">
            <span class="text-gray-500">要求数量</span>
            <span class="font-medium">{{ act.num }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500">奖励金额</span>
            <span class="font-medium text-red-500">¥{{ act.money }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500">活动时间</span>
            <span class="text-xs">{{ act.addtime }} ~ {{ act.endtime }}</span>
          </div>
        </div>

        <div class="mt-3 pt-3 border-t text-center">
          <GiftOutlined class="text-orange-500 mr-1" />
          <span class="text-sm text-gray-500">完成任务即可获得奖励</span>
        </div>
      </Card>
    </div>
  </Page>
</template>

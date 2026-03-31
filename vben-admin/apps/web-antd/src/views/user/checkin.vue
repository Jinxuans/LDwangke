<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import { Card, Button, message, Result } from 'ant-design-vue';
import { GiftOutlined, CheckCircleOutlined } from '@ant-design/icons-vue';
import { userCheckinApi, userCheckinStatusApi } from '#/api/plugins/checkin';

const loading = ref(false);
const checkedIn = ref(false);
const reward = ref(0);

async function loadStatus() {
  try {
    const res = await userCheckinStatusApi();
    checkedIn.value = res.checked_in;
    if (res.reward_money) reward.value = res.reward_money;
  } catch {}
}

async function doCheckin() {
  loading.value = true;
  try {
    const res = await userCheckinApi();
    reward.value = res.reward_money;
    checkedIn.value = true;
    message.success(`签到成功！奖励 ${res.reward_money} 元`);
  } catch (e: any) {
    message.error(e?.message || '签到失败');
  } finally {
    loading.value = false;
  }
}

onMounted(loadStatus);
</script>

<template>
  <Page title="每日签到">
    <Card>
      <Result
        v-if="checkedIn"
        status="success"
        title="今日已签到"
        :sub-title="`奖励 ${reward} 元已到账`"
      >
        <template #icon>
          <CheckCircleOutlined style="color: #52c41a" />
        </template>
      </Result>
      <div v-else style="text-align: center; padding: 40px 0">
        <GiftOutlined style="font-size: 64px; color: #faad14; margin-bottom: 24px" />
        <div class="mb-6 text-base text-gray-600 dark:text-gray-400">
          每日签到可获得随机奖励
        </div>
        <Button type="primary" size="large" :loading="loading" @click="doCheckin">
          立即签到
        </Button>
      </div>
    </Card>
  </Page>
</template>

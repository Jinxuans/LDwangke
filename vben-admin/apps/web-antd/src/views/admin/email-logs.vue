<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, Space, Tag, Select, SelectOption, Pagination,
} from 'ant-design-vue';
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons-vue';
import { getEmailSendLogsApi, type EmailSendLog } from '#/api/email-pool';

const loading = ref(false);
const list = ref<EmailSendLog[]>([]);
const total = ref(0);
const query = reactive({ page: 1, limit: 20, mail_type: '', status: -1, to_email: '' });

const columns = [
  { title: 'ID', dataIndex: 'id', width: 70 },
  { title: '类型', key: 'type', width: 90 },
  { title: '发件邮箱', dataIndex: 'from_email', width: 200, ellipsis: true },
  { title: '收件邮箱', dataIndex: 'to_email', width: 200, ellipsis: true },
  { title: '主题', dataIndex: 'subject', width: 250, ellipsis: true },
  { title: '状态', key: 'status', width: 80 },
  { title: '错误', dataIndex: 'error', width: 200, ellipsis: true },
  { title: '时间', dataIndex: 'addtime', width: 160 },
];

const typeMap: Record<string, string> = {
  register: '注册',
  reset: '重置密码',
  notify: '通知',
  mass: '群发',
  login_alert: '登录提醒',
  change_email: '邮箱变更',
};

async function loadData() {
  loading.value = true;
  try {
    const raw = await getEmailSendLogsApi({ ...query });
    const res = raw;
    list.value = res?.list ?? [];
    total.value = res?.total ?? 0;
    if (!Array.isArray(list.value)) list.value = [];
  } catch (e) { console.error(e); }
  finally { loading.value = false; }
}

function handleSearch() {
  query.page = 1;
  loadData();
}

function handlePageChange(page: number, pageSize: number) {
  query.page = page;
  query.limit = pageSize;
  loadData();
}

onMounted(loadData);
</script>

<template>
  <Page title="邮件发送日志" description="每封邮件的发送明细记录">
    <Card>
      <template #title>
        <Space wrap>
          <Select v-model:value="query.mail_type" style="width: 130px" placeholder="邮件类型" allow-clear>
            <SelectOption value="">全部类型</SelectOption>
            <SelectOption value="register">注册</SelectOption>
            <SelectOption value="reset">重置密码</SelectOption>
            <SelectOption value="notify">通知</SelectOption>
            <SelectOption value="mass">群发</SelectOption>
            <SelectOption value="login_alert">登录提醒</SelectOption>
            <SelectOption value="change_email">邮箱变更</SelectOption>
          </Select>
          <Select v-model:value="query.status" style="width: 110px">
            <SelectOption :value="-1">全部状态</SelectOption>
            <SelectOption :value="1">成功</SelectOption>
            <SelectOption :value="0">失败</SelectOption>
          </Select>
          <Input v-model:value="query.to_email" placeholder="收件邮箱" style="width: 200px" allow-clear />
          <Button type="primary" @click="handleSearch"><SearchOutlined /> 搜索</Button>
          <Button @click="loadData"><ReloadOutlined /> 刷新</Button>
        </Space>
      </template>

      <Table :columns="columns" :data-source="list" :loading="loading" row-key="id" :pagination="false" size="small" :scroll="{ x: 1200 }">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'type'">
            <Tag>{{ typeMap[record.mail_type] || record.mail_type }}</Tag>
          </template>
          <template v-if="column.key === 'status'">
            <Tag v-if="record.status === 1" color="green">成功</Tag>
            <Tag v-else color="red">失败</Tag>
          </template>
        </template>
      </Table>

      <div class="mt-4 flex justify-end">
        <Pagination
          :current="query.page"
          :total="total"
          :page-size="query.limit"
          show-size-changer
          show-quick-jumper
          :show-total="(t: number) => `共 ${t} 条`"
          @change="handlePageChange"
        />
      </div>
    </Card>
  </Page>
</template>

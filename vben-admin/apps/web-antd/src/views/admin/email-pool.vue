<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal, Switch, message, Select, SelectOption, Statistic, Row, Col,
} from 'ant-design-vue';
import {
  PlusOutlined, EditOutlined, DeleteOutlined, SendOutlined, ReloadOutlined,
  CheckCircleOutlined, CloseCircleOutlined, ExclamationCircleOutlined,
} from '@ant-design/icons-vue';
import {
  getEmailPoolListApi, saveEmailPoolApi, deleteEmailPoolApi,
  toggleEmailPoolApi, testEmailPoolApi, getEmailPoolStatsApi,
  resetEmailPoolCountersApi,
  type EmailPoolAccount, type EmailPoolSaveRequest, type EmailPoolStats,
} from '#/api/email-pool';

const loading = ref(false);
const list = ref<EmailPoolAccount[]>([]);
const stats = ref<EmailPoolStats>({ total_accounts: 0, active_accounts: 0, error_accounts: 0, today_sent: 0, today_fail: 0 });
const editVisible = ref(false);
const testVisible = ref(false);
const testTo = ref('');
const testId = ref(0);
const testLoading = ref(false);

const form = reactive<EmailPoolSaveRequest>({
  id: 0, name: '', host: '', port: 465, encryption: 'ssl',
  user: '', password: '', from_email: '', weight: 1,
  day_limit: 500, hour_limit: 50, status: 1,
});

const columns = [
  { title: 'ID', dataIndex: 'id', width: 60 },
  { title: '名称', dataIndex: 'name', width: 120 },
  { title: 'SMTP', key: 'smtp', width: 200 },
  { title: '权重', dataIndex: 'weight', width: 60 },
  { title: '限额', key: 'limits', width: 140 },
  { title: '今日/累计', key: 'counts', width: 130 },
  { title: '状态', key: 'status', width: 90 },
  { title: '最后使用', dataIndex: 'last_used', width: 160 },
  { title: '操作', key: 'action', width: 200 },
];

async function loadData() {
  loading.value = true;
  try {
    const raw = await getEmailPoolListApi();
    list.value = raw;
    if (!Array.isArray(list.value)) list.value = [];
    const rawS = await getEmailPoolStatsApi();
    stats.value = rawS;
  } catch (e) { console.error(e); }
  finally { loading.value = false; }
}

function openEdit(item?: EmailPoolAccount) {
  if (item) {
    Object.assign(form, {
      id: item.id, name: item.name, host: item.host, port: item.port,
      encryption: item.encryption, user: item.user, password: '',
      from_email: item.from_email, weight: item.weight,
      day_limit: item.day_limit, hour_limit: item.hour_limit, status: item.status,
    });
  } else {
    Object.assign(form, {
      id: 0, name: '', host: 'smtp.qq.com', port: 465, encryption: 'ssl',
      user: '', password: '', from_email: '', weight: 1,
      day_limit: 500, hour_limit: 50, status: 1,
    });
  }
  editVisible.value = true;
}

async function handleSave() {
  if (!form.host || !form.user) { message.warning('请填写SMTP信息'); return; }
  if (form.id === 0 && !form.password) { message.warning('新建邮箱必须填写授权码'); return; }
  try {
    await saveEmailPoolApi({ ...form });
    message.success('保存成功');
    editVisible.value = false;
    loadData();
  } catch (e: any) { message.error(e?.response?.data?.message || '保存失败'); }
}

async function handleDelete(id: number) {
  Modal.confirm({
    title: '确认删除该邮箱？',
    onOk: async () => {
      await deleteEmailPoolApi(id);
      message.success('已删除');
      loadData();
    },
  });
}

async function handleToggle(id: number, status: number) {
  await toggleEmailPoolApi(id, status);
  message.success('操作成功');
  loadData();
}

function openTest(id: number) {
  testId.value = id;
  testTo.value = '';
  testVisible.value = true;
}

async function handleTest() {
  if (!testTo.value) { message.warning('请输入测试收件邮箱'); return; }
  testLoading.value = true;
  try {
    await testEmailPoolApi(testId.value, testTo.value);
    message.success('测试邮件已发送');
    testVisible.value = false;
  } catch (e: any) { message.error(e?.response?.data?.message || '发送失败'); }
  finally { testLoading.value = false; }
}

async function handleResetCounters() {
  await resetEmailPoolCountersApi();
  message.success('计数器已重置');
  loadData();
}

onMounted(loadData);
</script>

<template>
  <Page title="邮箱轮询池" description="管理多个SMTP发件邮箱，系统自动轮询调度发送">
    <!-- 统计卡片 -->
    <Row :gutter="16" class="mb-4">
      <Col :span="4"><Card><Statistic title="邮箱总数" :value="stats.total_accounts" /></Card></Col>
      <Col :span="4"><Card><Statistic title="启用中" :value="stats.active_accounts" :value-style="{ color: '#52c41a' }" /></Card></Col>
      <Col :span="4"><Card><Statistic title="异常" :value="stats.error_accounts" :value-style="{ color: '#ff4d4f' }" /></Card></Col>
      <Col :span="6"><Card><Statistic title="今日成功" :value="stats.today_sent" :value-style="{ color: '#1890ff' }" /></Card></Col>
      <Col :span="6"><Card><Statistic title="今日失败" :value="stats.today_fail" :value-style="{ color: '#faad14' }" /></Card></Col>
    </Row>

    <Card>
      <template #title>
        <Space>
          <Button type="primary" @click="openEdit()"><PlusOutlined /> 新增邮箱</Button>
          <Button @click="handleResetCounters"><ReloadOutlined /> 重置计数</Button>
          <Button @click="loadData"><ReloadOutlined /> 刷新</Button>
        </Space>
      </template>

      <Table :columns="columns" :data-source="list" :loading="loading" row-key="id" :pagination="false" size="small" :scroll="{ x: 1100 }">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'smtp'">
            <div>{{ record.user }}</div>
            <div class="text-xs text-gray-400">{{ record.host }}:{{ record.port }} ({{ record.encryption }})</div>
          </template>
          <template v-if="column.key === 'limits'">
            <div>日限: {{ record.day_limit || '不限' }}</div>
            <div>时限: {{ record.hour_limit || '不限' }}</div>
          </template>
          <template v-if="column.key === 'counts'">
            <div>今日: {{ record.today_sent }} / 本时: {{ record.hour_sent }}</div>
            <div class="text-xs text-gray-400">累计: {{ record.total_sent }} 成功 / {{ record.total_fail }} 失败</div>
          </template>
          <template v-if="column.key === 'status'">
            <Tag v-if="record.status === 1" color="green">启用</Tag>
            <Tag v-else-if="record.status === 2" color="red">
              <ExclamationCircleOutlined /> 异常
            </Tag>
            <Tag v-else color="default">禁用</Tag>
            <div v-if="record.last_error && record.status === 2" class="text-xs text-red-400 mt-1 max-w-[120px] truncate" :title="record.last_error">{{ record.last_error }}</div>
          </template>
          <template v-if="column.key === 'action'">
            <Space size="small">
              <Button size="small" @click="openEdit(record)"><EditOutlined /></Button>
              <Button size="small" @click="openTest(record.id)"><SendOutlined /></Button>
              <Button v-if="record.status !== 1" size="small" type="primary" ghost @click="handleToggle(record.id, 1)">启用</Button>
              <Button v-if="record.status === 1" size="small" danger ghost @click="handleToggle(record.id, 0)">禁用</Button>
              <Button v-if="record.status === 2" size="small" type="primary" ghost @click="handleToggle(record.id, 1)">恢复</Button>
              <Button size="small" danger @click="handleDelete(record.id)"><DeleteOutlined /></Button>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <!-- 编辑弹窗 -->
    <Modal v-model:open="editVisible" :title="form.id ? '编辑邮箱' : '新增邮箱'" width="520px" @ok="handleSave">
      <div class="space-y-3 py-2">
        <div class="flex items-center gap-2">
          <span class="w-20 shrink-0 text-right">名称</span>
          <Input v-model:value="form.name" placeholder="发件人名称" />
        </div>
        <div class="flex items-center gap-2">
          <span class="w-20 shrink-0 text-right">SMTP</span>
          <Input v-model:value="form.host" placeholder="smtp.qq.com" class="flex-1" />
          <InputNumber v-model:value="form.port" :min="1" :max="65535" style="width: 90px" />
        </div>
        <div class="flex items-center gap-2">
          <span class="w-20 shrink-0 text-right">加密</span>
          <Select v-model:value="form.encryption" style="width: 120px">
            <SelectOption value="ssl">SSL</SelectOption>
            <SelectOption value="starttls">STARTTLS</SelectOption>
            <SelectOption value="none">无加密</SelectOption>
          </Select>
        </div>
        <div class="flex items-center gap-2">
          <span class="w-20 shrink-0 text-right">账号</span>
          <Input v-model:value="form.user" placeholder="SMTP账号/邮箱" />
        </div>
        <div class="flex items-center gap-2">
          <span class="w-20 shrink-0 text-right">授权码</span>
          <Input v-model:value="form.password" type="password" :placeholder="form.id ? '留空不修改' : '必填'" />
        </div>
        <div class="flex items-center gap-2">
          <span class="w-20 shrink-0 text-right">发件邮箱</span>
          <Input v-model:value="form.from_email" placeholder="留空=同账号" />
        </div>
        <div class="flex items-center gap-2">
          <span class="w-20 shrink-0 text-right">权重</span>
          <InputNumber v-model:value="form.weight" :min="1" :max="100" style="width: 90px" />
          <span class="w-20 shrink-0 text-right">日限额</span>
          <InputNumber v-model:value="form.day_limit" :min="0" style="width: 100px" />
        </div>
        <div class="flex items-center gap-2">
          <span class="w-20 shrink-0 text-right">时限额</span>
          <InputNumber v-model:value="form.hour_limit" :min="0" style="width: 100px" />
          <span class="w-20 shrink-0 text-right">状态</span>
          <Select v-model:value="form.status" style="width: 100px">
            <SelectOption :value="1">启用</SelectOption>
            <SelectOption :value="0">禁用</SelectOption>
          </Select>
        </div>
      </div>
    </Modal>

    <!-- 测试弹窗 -->
    <Modal v-model:open="testVisible" title="测试发送" @ok="handleTest" :confirm-loading="testLoading">
      <div class="py-2">
        <Input v-model:value="testTo" placeholder="输入测试收件邮箱地址" />
      </div>
    </Modal>
  </Page>
</template>

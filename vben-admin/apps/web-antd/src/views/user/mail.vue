<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Tag, Input, Modal, InputNumber,
  Pagination, Select, message, Collapse, CollapsePanel,
} from 'ant-design-vue';
import { SendOutlined, ReloadOutlined, SettingOutlined } from '@ant-design/icons-vue';
import {
  sendMassEmailApi,
  getEmailLogsApi,
  previewEmailRecipientsApi,
  getSMTPConfigApi,
  saveSMTPConfigApi,
  testSMTPApi,
  type EmailLog,
  type SMTPConfig,
} from '#/api/mail';

// 群发表单
const targetType = ref<'all' | 'direct' | 'indirect' | 'grade' | 'uids'>('all');
const targetGrade = ref('1');
const targetUids = ref('');
const subject = ref('');
const content = ref('');
const sending = ref(false);
const previewCount = ref(0);
const previewing = ref(false);

// SMTP 配置
const smtpForm = reactive<SMTPConfig>({
  host: 'smtp.qq.com',
  port: 465,
  user: '',
  password: '',
  from_name: '',
  encryption: 'ssl',
});
const smtpSaving = ref(false);
const smtpTesting = ref(false);
const testEmail = ref('');

async function loadSMTPConfig() {
  try {
    const raw = await getSMTPConfigApi();
    const res = raw;
    Object.assign(smtpForm, res);
  } catch (e) {}
}

async function saveSMTP() {
  if (!smtpForm.host || !smtpForm.user) {
    message.warning('请填写 SMTP 服务器和发件邮箱');
    return;
  }
  smtpSaving.value = true;
  try {
    await saveSMTPConfigApi({ ...smtpForm });
    message.success('SMTP 配置已保存');
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    smtpSaving.value = false;
  }
}

async function testSMTP() {
  if (!testEmail.value.trim()) {
    message.warning('请输入测试收件邮箱');
    return;
  }
  smtpTesting.value = true;
  try {
    const raw = await testSMTPApi(testEmail.value);
    const res = raw;
    message.success(res?.message || '测试邮件已发送');
  } catch (e: any) {
    message.error(e?.message || '测试失败');
  } finally {
    smtpTesting.value = false;
  }
}

// 记录列表
const loading = ref(false);
const logs = ref<EmailLog[]>([]);
const pagination = reactive({ page: 1, limit: 20, total: 0 });

// 计算 target 参数
const targetValue = computed(() => {
  if (targetType.value === 'all') return 'all';
  if (targetType.value === 'direct') return 'direct';
  if (targetType.value === 'indirect') return 'indirect';
  if (targetType.value === 'grade') return `grade:${targetGrade.value}`;
  return `uids:${targetUids.value}`;
});

// 加载记录
async function loadLogs(page = 1) {
  loading.value = true;
  pagination.page = page;
  try {
    const raw = await getEmailLogsApi({ page, limit: pagination.limit });
    const res = raw;
    logs.value = res.list || [];
    pagination.total = res.pagination?.total || 0;
  } catch (e) {
    console.error('加载发送记录失败:', e);
  } finally {
    loading.value = false;
  }
}

// 预览收件人数量
async function loadPreview() {
  if (!targetValue.value) return;
  previewing.value = true;
  try {
    const raw = await previewEmailRecipientsApi(targetValue.value);
    const res = raw;
    previewCount.value = res.count || 0;
  } catch (e: any) {
    previewCount.value = 0;
  } finally {
    previewing.value = false;
  }
}

// 目标变化时自动预览
watch([targetType, targetGrade], () => { loadPreview(); });

// 发送
async function handleSend() {
  if (!subject.value.trim()) {
    message.warning('请填写邮件标题');
    return;
  }
  if (!content.value.trim()) {
    message.warning('请填写邮件内容');
    return;
  }
  if (targetType.value === 'uids' && !targetUids.value.trim()) {
    message.warning('请填写目标 UID');
    return;
  }

  Modal.confirm({
    title: '确认发送',
    content: `将向 ${previewCount.value} 个邮箱发送邮件，确定？`,
    onOk: async () => {
      sending.value = true;
      try {
        const raw = await sendMassEmailApi({
          target: targetValue.value,
          subject: subject.value,
          content: content.value,
        });
        const res = raw;
        message.success(res.message || '发送任务已创建');
        subject.value = '';
        content.value = '';
        loadLogs(1);
      } catch (e: any) {
        message.error(e?.message || '发送失败');
      } finally {
        sending.value = false;
      }
    },
  });
}

// 状态标签
function statusTag(status: string) {
  const map: Record<string, { color: string; text: string }> = {
    sending: { color: 'processing', text: '发送中' },
    done: { color: 'success', text: '已完成' },
    partial: { color: 'warning', text: '部分失败' },
    failed: { color: 'error', text: '发送失败' },
  };
  return map[status] || { color: 'default', text: status };
}

// 目标显示
function targetDisplay(target: string) {
  if (target === 'all') return '全部用户';
  if (target === 'direct') return '直属用户';
  if (target === 'indirect') return '非直属用户';
  if (target.startsWith('grade:')) return `等级 ${target.slice(6)}`;
  if (target.startsWith('uids:')) return `UID: ${target.slice(5)}`;
  return target;
}

const columns = [
  { title: '状态', dataIndex: 'status', key: 'status', width: 90 },
  { title: '标题', dataIndex: 'subject', key: 'subject', ellipsis: true },
  { title: '收件范围', dataIndex: 'target', key: 'target', width: 120 },
  { title: '总数', dataIndex: 'total', key: 'total', width: 70 },
  { title: '成功', dataIndex: 'success_count', key: 'success', width: 70 },
  { title: '失败', dataIndex: 'fail_count', key: 'fail', width: 70 },
  { title: '时间', dataIndex: 'addtime', key: 'addtime', width: 160 },
];

onMounted(() => {
  loadSMTPConfig();
  loadLogs(1);
  loadPreview();
});
</script>

<template>
  <Page title="邮件群发" content-class="p-4">
    <!-- SMTP 配置 -->
    <Collapse class="mb-4">
      <CollapsePanel key="smtp" header="SMTP 邮箱配置">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-3xl">
          <div>
            <label class="block text-sm font-medium mb-1">SMTP 服务器</label>
            <Input v-model:value="smtpForm.host" placeholder="如 smtp.qq.com">
              <template #addonBefore>
                <Select v-model:value="smtpForm.encryption" style="width: 110px">
                  <Select.Option value="ssl">SSL</Select.Option>
                  <Select.Option value="starttls">STARTTLS</Select.Option>
                  <Select.Option value="none">无加密</Select.Option>
                </Select>
              </template>
            </Input>
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">端口</label>
            <InputNumber v-model:value="smtpForm.port" :min="1" :max="65535" style="width: 100%" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">发件邮箱</label>
            <Input v-model:value="smtpForm.user" placeholder="your@qq.com" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">授权码</label>
            <Input.Password v-model:value="smtpForm.password" placeholder="邮箱授权码（非登录密码）" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">发件人名称</label>
            <Input v-model:value="smtpForm.from_name" placeholder="系统通知" />
          </div>
        </div>
        <div class="flex flex-wrap items-center gap-3 mt-4">
          <Button type="primary" :loading="smtpSaving" @click="saveSMTP">
            <template #icon><SettingOutlined /></template>
            保存配置
          </Button>
          <div class="flex items-center gap-2">
            <Input v-model:value="testEmail" placeholder="测试收件邮箱" style="width: 220px" />
            <Button :loading="smtpTesting" @click="testSMTP">发送测试</Button>
          </div>
        </div>
        <div class="mt-3 text-gray-400 text-xs">
          常用配置：QQ邮箱 smtp.qq.com:465(SSL) | 163邮箱 smtp.163.com:465(SSL) | Gmail smtp.gmail.com:587(STARTTLS) | Outlook smtp-mail.outlook.com:587(STARTTLS)
        </div>
      </CollapsePanel>
    </Collapse>

    <!-- 写邮件 -->
    <Card title="发送邮件" class="mb-4">
      <div class="space-y-4 max-w-3xl">
        <div>
          <label class="block text-sm font-medium mb-1">收件人</label>
          <div class="flex flex-wrap items-center gap-2">
            <Select v-model:value="targetType" style="width: 140px">
              <Select.Option value="all">全部用户</Select.Option>
              <Select.Option value="direct">直属用户</Select.Option>
              <Select.Option value="indirect">非直属用户</Select.Option>
              <Select.Option value="grade">按等级</Select.Option>
              <Select.Option value="uids">指定UID</Select.Option>
            </Select>
            <Select
              v-if="targetType === 'grade'"
              v-model:value="targetGrade"
              style="width: 120px"
            >
              <Select.Option value="0">普通用户</Select.Option>
              <Select.Option value="1">VIP</Select.Option>
              <Select.Option value="2">管理员</Select.Option>
              <Select.Option value="3">超级管理员</Select.Option>
            </Select>
            <Input
              v-if="targetType === 'uids'"
              v-model:value="targetUids"
              placeholder="输入UID，用逗号分隔，如 1,2,3"
              style="width: 300px"
              @blur="loadPreview"
            />
            <Button size="small" @click="loadPreview" :loading="previewing">
              预览人数
            </Button>
            <Tag v-if="previewCount > 0" color="blue">{{ previewCount }} 人</Tag>
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">邮件标题</label>
          <Input v-model:value="subject" placeholder="请输入邮件标题" :maxlength="200" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">邮件内容（支持 HTML）</label>
          <Input.TextArea
            v-model:value="content"
            :rows="8"
            placeholder="请输入邮件内容，支持 HTML 格式"
          />
        </div>
        <div class="flex items-center gap-3">
          <Button
            type="primary"
            :loading="sending"
            :disabled="!subject.trim() || !content.trim() || previewCount === 0"
            @click="handleSend"
          >
            <template #icon><SendOutlined /></template>
            发送邮件
          </Button>
          <span class="text-gray-400 text-sm">邮件将异步发送，可在下方查看发送状态</span>
        </div>
      </div>
    </Card>

    <!-- 发送记录 -->
    <Card title="发送记录">
      <div class="flex justify-end mb-3">
        <Button @click="loadLogs(pagination.page)">
          <template #icon><ReloadOutlined /></template>
          刷新
        </Button>
      </div>

      <Table
        :columns="columns"
        :data-source="logs"
        :loading="loading"
        :pagination="false"
        row-key="id"
        size="small"
        :scroll="{ x: 700 }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <Tag :color="statusTag(record.status).color">
              {{ statusTag(record.status).text }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'target'">
            {{ targetDisplay(record.target) }}
          </template>
          <template v-else-if="column.key === 'success'">
            <span class="text-green-600">{{ record.success_count }}</span>
          </template>
          <template v-else-if="column.key === 'fail'">
            <span :class="record.fail_count > 0 ? 'text-red-500' : ''">{{ record.fail_count }}</span>
          </template>
        </template>
      </Table>

      <div class="flex justify-center mt-4">
        <Pagination
          v-model:current="pagination.page"
          :total="pagination.total"
          :page-size="pagination.limit"
          :show-total="(total: number) => `共 ${total} 条`"
          @change="(p: number) => loadLogs(p)"
        />
      </div>
    </Card>
  </Page>
</template>

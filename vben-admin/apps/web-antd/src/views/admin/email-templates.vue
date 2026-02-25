<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Button, Input, Textarea, Space, Tag, Modal, Switch, message, Tabs, TabPane,
} from 'ant-design-vue';
import { EditOutlined, EyeOutlined, SendOutlined } from '@ant-design/icons-vue';
import {
  getEmailTemplatesApi, saveEmailTemplateApi, previewEmailTemplateApi, testEmailTemplateApi,
  type EmailTemplate,
} from '#/api/email-template';

const loading = ref(false);
const templates = ref<EmailTemplate[]>([]);
const editVisible = ref(false);
const previewVisible = ref(false);
const testVisible = ref(false);
const previewHtml = ref('');
const previewSubject = ref('');
const testTo = ref('');
const testCode = ref('');
const testLoading = ref(false);
const saving = ref(false);

const editForm = ref({ id: 0, subject: '', content: '', status: 1 });

const codeNameMap: Record<string, { label: string; color: string; desc: string }> = {
  register: { label: '注册验证码', color: 'blue', desc: '用户注册时发送验证码邮件' },
  reset_password: { label: '重置密码', color: 'orange', desc: '用户找回密码时发送验证码邮件' },
  system_notify: { label: '系统通知', color: 'green', desc: '后台群发或系统通知邮件' },
};

async function loadData() {
  loading.value = true;
  try {
    const raw = await getEmailTemplatesApi();
    templates.value = raw;
    if (!Array.isArray(templates.value)) templates.value = [];
  } catch (e) { console.error(e); }
  finally { loading.value = false; }
}

function openEdit(tpl: EmailTemplate) {
  editForm.value = { id: tpl.id, subject: tpl.subject, content: tpl.content, status: tpl.status };
  editVisible.value = true;
}

async function handleSave() {
  if (!editForm.value.subject || !editForm.value.content) {
    message.warning('标题和内容不能为空');
    return;
  }
  saving.value = true;
  try {
    await saveEmailTemplateApi(editForm.value);
    message.success('保存成功');
    editVisible.value = false;
    loadData();
  } catch (e: any) { message.error('保存失败'); }
  finally { saving.value = false; }
}

async function handlePreview(code: string) {
  try {
    const raw = await previewEmailTemplateApi(code);
    const res = raw;
    previewSubject.value = res?.subject || '';
    previewHtml.value = res?.html || '';
    previewVisible.value = true;
  } catch (e: any) { message.error('预览失败'); }
}

function openTest(code: string) {
  testCode.value = code;
  testTo.value = '';
  testVisible.value = true;
}

async function handleTest() {
  if (!testTo.value) { message.warning('请输入测试收件邮箱'); return; }
  testLoading.value = true;
  try {
    await testEmailTemplateApi(testCode.value, testTo.value);
    message.success('测试邮件已发送');
    testVisible.value = false;
  } catch (e: any) { message.error('发送失败'); }
  finally { testLoading.value = false; }
}

onMounted(loadData);
</script>

<template>
  <Page title="邮件模板管理" description="管理注册验证码、重置密码、系统通知三类邮件模板，支持变量替换">
    <div class="space-y-4">
      <Card v-for="tpl in templates" :key="tpl.id">
        <template #title>
          <Space>
            <Tag :color="codeNameMap[tpl.code]?.color || 'default'">
              {{ codeNameMap[tpl.code]?.label || tpl.code }}
            </Tag>
            <span>{{ tpl.name }}</span>
            <Tag v-if="tpl.status === 1" color="green">启用</Tag>
            <Tag v-else color="default">禁用</Tag>
          </Space>
        </template>
        <template #extra>
          <Space>
            <Button size="small" @click="handlePreview(tpl.code)"><EyeOutlined /> 预览</Button>
            <Button size="small" @click="openTest(tpl.code)"><SendOutlined /> 测试</Button>
            <Button size="small" type="primary" @click="openEdit(tpl)"><EditOutlined /> 编辑</Button>
          </Space>
        </template>

        <div class="space-y-2">
          <div><span class="text-gray-500">描述：</span>{{ codeNameMap[tpl.code]?.desc || '' }}</div>
          <div><span class="text-gray-500">邮件标题：</span><code>{{ tpl.subject }}</code></div>
          <div>
            <span class="text-gray-500">可用变量：</span>
            <Tag v-for="v in (tpl.variables || '').split(',')" :key="v" class="mr-1">{{ '{' + v.trim() + '}' }}</Tag>
          </div>
          <div v-if="tpl.updated_at" class="text-xs text-gray-400">最后修改：{{ tpl.updated_at }}</div>
        </div>
      </Card>

      <Card v-if="templates.length === 0 && !loading">
        <div class="py-8 text-center text-gray-400">暂无模板，请先执行数据库迁移创建默认模板</div>
      </Card>
    </div>

    <!-- 编辑弹窗 -->
    <Modal v-model:open="editVisible" title="编辑邮件模板" width="700px" @ok="handleSave" :confirm-loading="saving">
      <div class="space-y-3 py-2">
        <div>
          <div class="mb-1 text-gray-600">邮件标题（支持变量如 {site_name}）</div>
          <Input v-model:value="editForm.subject" placeholder="{site_name} - 注册验证码" />
        </div>
        <div>
          <div class="mb-1 text-gray-600">邮件内容（HTML，支持变量）</div>
          <Textarea v-model:value="editForm.content" :rows="12" placeholder="HTML内容..." />
        </div>
        <div class="flex items-center gap-2">
          <span class="text-gray-600">状态：</span>
          <Switch v-model:checked="editForm.status" :checked-value="1" :un-checked-value="0" checked-children="启用" un-checked-children="禁用" />
        </div>
      </div>
    </Modal>

    <!-- 预览弹窗 -->
    <Modal v-model:open="previewVisible" title="模板预览" width="640px" :footer="null">
      <div class="mb-2"><strong>标题：</strong>{{ previewSubject }}</div>
      <div class="border rounded p-4" v-html="previewHtml"></div>
    </Modal>

    <!-- 测试发送弹窗 -->
    <Modal v-model:open="testVisible" title="测试发送" @ok="handleTest" :confirm-loading="testLoading">
      <div class="py-2">
        <Input v-model:value="testTo" placeholder="输入测试收件邮箱地址" />
      </div>
    </Modal>
  </Page>
</template>

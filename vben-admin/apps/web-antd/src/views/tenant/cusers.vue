<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import { Card, Table, Button, Input, Space, Modal, message } from 'ant-design-vue';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons-vue';
import { getTenantCUsersApi, saveTenantCUserApi, deleteTenantCUserApi, type CUser } from '#/api/tenant';

const loading = ref(false);
const list = ref<CUser[]>([]);
const total = ref(0);
const page = ref(1);
const editVisible = ref(false);
const form = reactive({ id: 0, account: '', password: '', nickname: '' });

async function load() {
  loading.value = true;
  try {
    const res = await getTenantCUsersApi({ page: page.value, limit: 20 });
    const d = res;
    list.value = d?.list ?? [];
    total.value = d?.total ?? 0;
  } catch (e: any) {
    message.error(e?.message || '加载失败');
  } finally {
    loading.value = false;
  }
}

function openAdd() {
  Object.assign(form, { id: 0, account: '', password: '', nickname: '' });
  editVisible.value = true;
}

function openEdit(u: CUser) {
  Object.assign(form, { id: u.id, account: u.account, password: '', nickname: u.nickname });
  editVisible.value = true;
}

async function handleSave() {
  if (!form.account) { message.warning('请填写账号'); return; }
  if (!form.id && !form.password) { message.warning('请填写密码'); return; }
  try {
    await saveTenantCUserApi({ ...form });
    message.success('保存成功');
    editVisible.value = false;
    load();
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  }
}

function handleDelete(u: CUser) {
  Modal.confirm({
    title: '确认删除',
    content: `删除会员「${u.account}」后不可恢复，确定继续？`,
    async onOk() {
      try {
        await deleteTenantCUserApi(u.id);
        message.success('已删除');
        load();
      } catch (e: any) {
        message.error(e?.message || '操作失败');
      }
    },
  });
}

onMounted(load);
</script>

<template>
  <Page title="会员管理" content-class="p-4">
    <Card>
      <div class="flex justify-between items-center mb-4">
        <span class="text-sm text-gray-500">共 {{ total }} 位会员</span>
        <Button type="primary" @click="openAdd">
          <template #icon><PlusOutlined /></template>
          添加会员
        </Button>
      </div>

      <Table
        :data-source="list"
        :loading="loading"
        :pagination="{ current: page, pageSize: 20, total, onChange: (p: number) => { page = p; load(); } }"
        row-key="id"
        size="small"
        bordered
      >
        <Table.Column title="账号" data-index="account" />
        <Table.Column title="昵称" data-index="nickname" />
        <Table.Column title="注册时间" data-index="addtime" :width="160" />
        <Table.Column title="操作" :width="140">
          <template #default="{ record }">
            <Space>
              <Button type="link" size="small" @click="openEdit(record)"><EditOutlined /> 编辑</Button>
              <Button type="link" size="small" danger @click="handleDelete(record)"><DeleteOutlined /> 删除</Button>
            </Space>
          </template>
        </Table.Column>
      </Table>
    </Card>

    <Modal v-model:open="editVisible" :title="form.id ? '编辑会员' : '添加会员'" @ok="handleSave" ok-text="保存">
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium mb-1">账号</label>
          <Input v-model:value="form.account" placeholder="登录账号" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">{{ form.id ? '新密码（留空不修改）' : '密码' }}</label>
          <Input.Password v-model:value="form.password" placeholder="登录密码" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">昵称（选填）</label>
          <Input v-model:value="form.nickname" placeholder="显示名称，默认同账号" />
        </div>
      </div>
    </Modal>
  </Page>
</template>

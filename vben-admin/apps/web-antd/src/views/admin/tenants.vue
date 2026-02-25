<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, Space, Tag, Modal, Form, FormItem,
  Pagination, Popconfirm, message, AutoComplete,
} from 'ant-design-vue';
import { ReloadOutlined, PlusOutlined } from '@ant-design/icons-vue';
import {
  adminGetTenantsApi, adminCreateTenantApi, adminSetTenantStatusApi,
  type AdminTenantItem,
} from '#/api/tenant';
import { getUserListApi, type UserItem } from '#/api/admin';

const loading = ref(false);
const list = ref<AdminTenantItem[]>([]);
const pagination = reactive({ page: 1, limit: 20, total: 0 });

// 新建弹窗
const createVisible = ref(false);
const createForm = reactive({ uid: undefined as number | undefined, shop_name: '' });
const createLoading = ref(false);

// 用户搜索
const userSearchText = ref('');
const userOptions = ref<{ value: string; uid: number }[]>([]);
let searchTimer: ReturnType<typeof setTimeout> | null = null;

function onUserSearch(val: string) {
  if (searchTimer) clearTimeout(searchTimer);
  if (!val.trim()) { userOptions.value = []; return; }
  searchTimer = setTimeout(async () => {
    try {
      const res = await getUserListApi({ keywords: val.trim(), limit: 10 });
      const data = res;
      userOptions.value = (data?.list || []).map((u: UserItem) => ({
        value: `${u.name}（${u.user}）UID:${u.uid}`,
        uid: u.uid,
      }));
    } catch {}
  }, 300);
}

function onUserSelect(_val: string, option: any) {
  createForm.uid = option.uid;
}

function onCreateOpen() {
  createForm.uid = undefined;
  createForm.shop_name = '';
  userSearchText.value = '';
  userOptions.value = [];
}

async function loadList(page = 1) {
  loading.value = true;
  pagination.page = page;
  try {
    const raw = await adminGetTenantsApi({ page: pagination.page, limit: pagination.limit });
    const res = raw;
    list.value = res.list || [];
    pagination.total = res.total || 0;
  } catch (e) {
    console.error(e);
  } finally {
    loading.value = false;
  }
}

async function handleCreate() {
  if (!createForm.uid || !createForm.shop_name) {
    message.warning('请先搜索并选择用户，并填写店铺名称');
    return;
  }
  createLoading.value = true;
  try {
    await adminCreateTenantApi({ uid: createForm.uid, shop_name: createForm.shop_name });
    message.success('创建成功');
    createVisible.value = false;
    createForm.uid = undefined;
    createForm.shop_name = '';
    loadList(1);
  } catch (e: any) {
    message.error(e?.message || '创建失败');
  } finally {
    createLoading.value = false;
  }
}

async function handleToggleStatus(record: AdminTenantItem) {
  const newStatus = record.status === 1 ? 0 : 1;
  try {
    await adminSetTenantStatusApi(record.tid, newStatus);
    message.success(newStatus === 1 ? '已启用' : '已禁用');
    await loadList(pagination.page);
  } catch (e: any) {
    message.error(e?.message || '操作失败');
  }
}

const columns = [
  { title: 'TID', dataIndex: 'tid', key: 'tid', width: 70 },
  { title: 'UID', dataIndex: 'uid', key: 'uid', width: 70 },
  { title: '店铺名称', dataIndex: 'shop_name', key: 'shop_name', ellipsis: true },
  { title: '状态', key: 'status', width: 90 },
  { title: '创建时间', dataIndex: 'addtime', key: 'addtime', width: 160 },
  { title: '操作', key: 'action', width: 120 },
];

onMounted(() => loadList(1));
</script>

<template>
  <Page title="租户管理" content-class="p-4">
    <Card>
      <div class="flex flex-wrap justify-between items-center gap-3 mb-4">
        <Space wrap>
          <Button type="primary" @click="createVisible = true; onCreateOpen()">
            <template #icon><PlusOutlined /></template>
            新建店铺
          </Button>
          <Button @click="loadList(pagination.page)">
            <template #icon><ReloadOutlined /></template>
          </Button>
        </Space>
        <Tag>共 {{ pagination.total }} 个店铺</Tag>
      </div>

      <Table :columns="columns" :data-source="list" :loading="loading" :pagination="false" row-key="tid" size="small" bordered :scroll="{ x: 700 }">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <Tag :color="record.status === 1 ? 'green' : 'red'">
              {{ record.status === 1 ? '正常' : '禁用' }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <Popconfirm
              :title="record.status === 1 ? '确定禁用该店铺？' : '确定启用该店铺？'"
              @confirm="handleToggleStatus(record)"
            >
              <Button type="link" size="small" :danger="record.status === 1">
                {{ record.status === 1 ? '禁用' : '启用' }}
              </Button>
            </Popconfirm>
          </template>
        </template>
      </Table>

      <div class="flex justify-center mt-4">
        <Pagination
          v-model:current="pagination.page"
          :total="pagination.total"
          :page-size="pagination.limit"
          :show-total="(total: number) => `共 ${total} 条`"
          @change="(p: number) => loadList(p)"
        />
      </div>
    </Card>

    <Modal v-model:open="createVisible" title="新建店铺" :confirm-loading="createLoading" ok-text="创建" @ok="handleCreate">
      <Form layout="vertical" class="mt-2">
        <FormItem label="选择用户" required>
          <AutoComplete
            v-model:value="userSearchText"
            :options="userOptions"
            style="width: 100%"
            placeholder="输入用户名/账号搜索"
            @search="onUserSearch"
            @select="onUserSelect"
          />
          <div v-if="createForm.uid" class="mt-1 text-xs text-gray-400">已选 UID: {{ createForm.uid }}</div>
        </FormItem>
        <FormItem label="店铺名称" required>
          <Input v-model:value="createForm.shop_name" placeholder="输入店铺名称" />
        </FormItem>
      </Form>
    </Modal>
  </Page>
</template>

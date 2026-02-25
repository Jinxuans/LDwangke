<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal,
  Pagination, Popconfirm, message, Select, SelectOption,
} from 'ant-design-vue';
import {
  SearchOutlined, ReloadOutlined, EditOutlined, LoginOutlined,
} from '@ant-design/icons-vue';
import { useAccessStore } from '@vben/stores';
import {
  getUserListApi, resetUserPassApi, setUserBalanceApi, setUserGradeApi,
  getGradeListApi, getSiteConfigApi, adminImpersonateApi,
  type UserItem, type GradeItem,
} from '#/api/admin';

const accessStore = useAccessStore();
const loading = ref(false);
const loginAsEnabled = ref(false);
const userList = ref<UserItem[]>([]);
const pagination = reactive({ page: 1, limit: 20, total: 0 });
const keywords = ref('');

// 等级列表（从 dengji 表加载）
const gradeOptions = ref<GradeItem[]>([]);

// 编辑弹窗
const editVisible = ref(false);
const editUser = ref<UserItem | null>(null);
const editBalance = ref(0);
const editAddPrice = ref(1);

async function loadGrades() {
  try {
    const raw = await getGradeListApi();
    gradeOptions.value = raw;
    if (!Array.isArray(gradeOptions.value)) gradeOptions.value = [];
  } catch (e) {}
}

async function loadUsers(page = 1) {
  loading.value = true;
  pagination.page = page;
  try {
    const raw = await getUserListApi({
      page: pagination.page,
      limit: pagination.limit,
      keywords: keywords.value,
    });
    const res = raw;
    userList.value = res.list || [];
    pagination.total = res.pagination?.total || 0;
  } catch (e) {
    console.error(e);
  } finally {
    loading.value = false;
  }
}

function openEdit(user: UserItem) {
  editUser.value = { ...user };
  editBalance.value = user.balance;
  editAddPrice.value = user.addprice;
  editVisible.value = true;
}

async function handleResetPass(uid: number) {
  let newPass = '';
  Modal.confirm({
    title: '重置密码',
    content: () => {
      const input = document.createElement('input');
      input.type = 'text';
      input.placeholder = '请输入新密码';
      input.style.cssText = 'width:100%;padding:4px 8px;border:1px solid #d9d9d9;border-radius:4px;margin-top:8px';
      input.oninput = (e) => { newPass = (e.target as HTMLInputElement).value; };
      return input;
    },
    async onOk() {
      if (!newPass || newPass.length < 6) {
        message.warning('密码不能少于6位');
        return Promise.reject();
      }
      await resetUserPassApi(uid, newPass);
      message.success('密码已重置');
    },
  });
}

async function handleSaveEdit() {
  if (!editUser.value) return;
  try {
    await setUserBalanceApi(editUser.value.uid, editBalance.value);
    await setUserGradeApi(editUser.value.uid, editAddPrice.value);
    message.success('保存成功');
    editVisible.value = false;
    loadUsers(pagination.page);
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  }
}

const columns = [
  { title: 'UID', dataIndex: 'uid', key: 'uid', width: 70 },
  { title: '账号', dataIndex: 'user', key: 'user', width: 140, ellipsis: true },
  { title: '昵称', dataIndex: 'name', key: 'name', width: 120, ellipsis: true },
  { title: '等级', key: 'grade', width: 120 },
  { title: '费率', key: 'addprice', width: 80 },
  { title: '余额', key: 'balance', width: 100 },
  { title: '注册时间', dataIndex: 'addtime', key: 'addtime', width: 160 },
  { title: '操作', key: 'action', width: 200 },
];

async function handleLoginAs(uid: number) {
  try {
    // 保存当前管理员token
    const currentToken = accessStore.accessToken;
    if (currentToken) {
      localStorage.setItem('admin_backup_token', currentToken);
    }
    const raw = await adminImpersonateApi(uid);
    const res = raw;
    if (res?.accessToken) {
      accessStore.setAccessToken(res.accessToken);
      window.location.href = '/';
    }
  } catch (e: any) {
    message.error(e?.message || '登入失败');
  }
}

onMounted(async () => {
  loadGrades();
  loadUsers(1);
  try {
    const cfg = await getSiteConfigApi();
    loginAsEnabled.value = cfg?.onlineStore_trdltz === '1';
  } catch {}
});
</script>

<template>
  <Page title="用户管理" content-class="p-4">
    <Card>
      <div class="flex flex-wrap justify-between items-center gap-3 mb-4">
        <Space wrap>
          <Input v-model:value="keywords" placeholder="搜索用户名/UID" allow-clear style="max-width: 200px; min-width: 120px" @pressEnter="loadUsers(1)" />
          <Button type="primary" @click="loadUsers(1)">
            <template #icon><SearchOutlined /></template>
          </Button>
          <Button @click="loadUsers(pagination.page)">
            <template #icon><ReloadOutlined /></template>
          </Button>
        </Space>
        <Tag>共 {{ pagination.total }} 个用户</Tag>
      </div>

      <Table :columns="columns" :data-source="userList" :loading="loading" :pagination="false" row-key="uid" size="small" bordered :scroll="{ x: 900 }">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'grade'">
            <Tag color="blue">{{ record.grade_name || '未知' }}</Tag>
          </template>
          <template v-else-if="column.key === 'addprice'">
            {{ record.addprice }}
          </template>
          <template v-else-if="column.key === 'balance'">
            ¥{{ Number(record.balance).toFixed(2) }}
          </template>
          <template v-else-if="column.key === 'action'">
            <Space size="small">
              <Button type="link" size="small" @click="openEdit(record)">
                <template #icon><EditOutlined /></template>
                编辑
              </Button>
              <Popconfirm title="确定重置密码为 123456？" @confirm="handleResetPass(record.uid)">
                <Button type="link" size="small" danger>重置密码</Button>
              </Popconfirm>
              <Button v-if="loginAsEnabled" type="link" size="small" @click="handleLoginAs(record.uid)">
                <template #icon><LoginOutlined /></template>
                登入
              </Button>
            </Space>
          </template>
        </template>
      </Table>

      <div class="flex justify-center mt-4">
        <Pagination
          v-model:current="pagination.page"
          v-model:pageSize="pagination.limit"
          :total="pagination.total"
          :page-size-options="['20', '50', '100']"
          show-size-changer
          :show-total="(total: number) => `共 ${total} 条`"
          @change="(p: number) => loadUsers(p)"
          @showSizeChange="(_: number, s: number) => { pagination.limit = s; loadUsers(1); }"
        />
      </div>
    </Card>

    <Modal v-model:open="editVisible" title="编辑用户" @ok="handleSaveEdit" ok-text="保存">
      <div class="space-y-4" v-if="editUser">
        <div>
          <label class="block text-sm font-medium mb-1">UID</label>
          <Input :value="String(editUser.uid)" disabled />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">账号</label>
          <Input :value="editUser.user" disabled />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">余额</label>
          <InputNumber v-model:value="editBalance" :min="0" :step="10" :precision="2" style="width: 100%" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">等级（费率）</label>
          <Select v-model:value="editAddPrice" style="width: 100%" show-search :filter-option="(input: string, option: any) => option.label?.toLowerCase().includes(input.toLowerCase())">
            <SelectOption
              v-for="g in gradeOptions"
              :key="g.id"
              :value="Number(g.rate)"
              :label="`${g.name}（${g.rate}）`"
            >
              {{ g.name }}（费率 {{ g.rate }}）
            </SelectOption>
          </Select>
        </div>
      </div>
    </Modal>
  </Page>
</template>

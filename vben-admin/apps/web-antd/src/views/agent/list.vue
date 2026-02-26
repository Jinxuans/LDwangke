<script setup lang="ts">
import { ref, h, onMounted, computed } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, Select, SelectOption, Tag, Space,
  Modal, InputNumber, Spin, Pagination, message,
} from 'ant-design-vue';
import {
  SearchOutlined, PlusOutlined, ReloadOutlined,
} from '@ant-design/icons-vue';
import { useUserStore } from '@vben/stores';
import type { AgentListItem } from '#/api/admin';
import {
  getAgentListApi, agentCreateApi, agentRechargeApi, agentDeductApi,
  agentChangeGradeApi, agentChangeStatusApi, agentResetPasswordApi,
  agentOpenKeyApi, agentSetInviteCodeApi, adminImpersonateApi,
  agentCrossRechargeCheckApi, agentCrossRechargeApi,
} from '#/api/admin';
import { useAccessStore } from '@vben/stores';
import { getAccessCodesApi, getUserInfoApi } from '#/api';
import type { GradeOption } from '#/api/user-center';
import { getUserGradeListApi } from '#/api/user-center';

const userStore = useUserStore();
const accessStore = useAccessStore();
const isAdmin = computed(() => (userStore.userInfo as any)?.roles?.includes('admin'));

// ===== 管理员免登录进入代理界面 =====
async function handleImpersonate(uid: number) {
  Modal.confirm({
    title: '提示',
    content: `确定要以 UID ${uid} 的身份进入系统吗？`,
    async onOk() {
      try {
        // 保存管理员 token 以便切回
        const currentToken = accessStore.accessToken;
        if (currentToken) {
          localStorage.setItem('admin_backup_token', currentToken);
        }
        const res = await adminImpersonateApi(uid);
        const data = res;
        if (data?.accessToken) {
          accessStore.setAccessToken(data.accessToken);
          // 刷新用户信息和权限码
          const [userRes, codesRes] = await Promise.all([
            getUserInfoApi(),
            getAccessCodesApi(),
          ]);
          const userInfo = userRes;
          const codes = codesRes;
          userStore.setUserInfo(userInfo);
          accessStore.setAccessCodes(codes);
          message.success(`已切换到 ${data.username || uid} 的身份`);
          // 刷新页面以重新加载菜单
          window.location.href = '/';
        }
      } catch (e: any) {
        message.error(e?.message || '切换失败');
      }
    },
  });
}

const loading = ref(false);
const tableData = ref<AgentListItem[]>([]);
const total = ref(0);
const currentPage = ref(1);
const pageSize = ref(20);
const searchType = ref('2');
const searchKeywords = ref('');
const gradeList = ref<GradeOption[]>([]);

async function loadData() {
  loading.value = true;
  try {
    const res = await getAgentListApi({
      page: currentPage.value,
      limit: pageSize.value,
      type: searchType.value,
      keywords: searchKeywords.value,
    });
    const d = res;
    tableData.value = d?.list ?? [];
    total.value = d?.pagination?.total ?? 0;
  } catch (e: any) {
    console.error('加载代理列表失败', e);
  } finally {
    loading.value = false;
  }
}

async function loadGrades() {
  try {
    const res = await getUserGradeListApi();
    gradeList.value = (res) || [];
  } catch { /* ignore */ }
}

function handleSearch() {
  currentPage.value = 1;
  loadData();
}

function handlePageChange(page: number) {
  currentPage.value = page;
  loadData();
}

function handleSizeChange(_: number, size: number) {
  pageSize.value = size;
  currentPage.value = 1;
  loadData();
}

// ===== 添加代理 =====
const addVisible = ref(false);
const addForm = ref({ nickname: '', user: '', pass: '', gradeId: undefined as number | undefined });
const addLoading = ref(false);

async function handleCreateAgent() {
  if (!addForm.value.nickname || !addForm.value.user || !addForm.value.pass || !addForm.value.gradeId) {
    message.error('请填写完整信息');
    return;
  }
  addLoading.value = true;
  try {
    // 第一次：获取费用预览
    const res1 = await agentCreateApi({
      nickname: addForm.value.nickname,
      user: addForm.value.user,
      pass: addForm.value.pass,
      gradeId: addForm.value.gradeId,
      type: 0,
    });
    const msg1 = (res1 as any)?.message || (res1 as any)?.msg || '确认创建？';
    Modal.confirm({
      title: '确认添加代理',
      content: msg1,
      async onOk() {
        const res2 = await agentCreateApi({
          nickname: addForm.value.nickname,
          user: addForm.value.user,
          pass: addForm.value.pass,
          gradeId: addForm.value.gradeId,
          type: 1,
        });
        message.success((res2 as any)?.message || '添加成功');
        addVisible.value = false;
        addForm.value = { nickname: '', user: '', pass: '', gradeId: undefined };
        loadData();
      },
    });
  } catch (e: any) {
    message.error(e?.message || '操作失败');
  } finally {
    addLoading.value = false;
  }
}

// ===== 修改等级 =====
const gradeVisible = ref(false);
const gradeUID = ref(0);
const gradeSelected = ref<number | undefined>(undefined);
const gradeLoading = ref(false);

function openChangeGrade(uid: number) {
  gradeUID.value = uid;
  gradeSelected.value = undefined;
  gradeVisible.value = true;
}

async function handleChangeGrade() {
  if (!gradeSelected.value) {
    message.error('请选择等级');
    return;
  }
  gradeLoading.value = true;
  try {
    const res1 = await agentChangeGradeApi({ uid: gradeUID.value, gradeId: gradeSelected.value, type: 0 });
    const msg1 = (res1 as any)?.message || '确认修改？';
    Modal.confirm({
      title: '确认修改等级',
      content: msg1,
      async onOk() {
        const res2 = await agentChangeGradeApi({ uid: gradeUID.value, gradeId: gradeSelected.value!, type: 1 });
        message.success((res2 as any)?.message || '修改成功');
        gradeVisible.value = false;
        loadData();
      },
    });
  } catch (e: any) {
    message.error(e?.message || '操作失败');
  } finally {
    gradeLoading.value = false;
  }
}

// ===== 充值 =====
function handleRecharge(uid: number) {
  const money = ref<number>(0);
  Modal.confirm({
    title: '充值余额',
    content: () => h(InputNumber, {
      value: money.value,
      'onUpdate:value': (v: number) => { money.value = v; },
      min: 0, precision: 2, placeholder: '请输入充值金额', style: 'width: 100%',
    }),
    async onOk() {
      if (!money.value || money.value <= 0) { message.error('请输入有效金额'); return; }
      try {
        await agentRechargeApi({ uid, money: money.value });
        message.success('充值成功');
        loadData();
      } catch (e: any) {
        message.error(e?.message || '充值失败');
      }
    },
  });
}

// ===== 扣款 =====
function handleDeduct(uid: number) {
  const money = ref<number>(0);
  Modal.confirm({
    title: '扣除余额',
    content: () => h(InputNumber, {
      value: money.value,
      'onUpdate:value': (v: number) => { money.value = v; },
      min: 0, precision: 2, placeholder: '请输入扣除金额', style: 'width: 100%',
    }),
    async onOk() {
      if (!money.value || money.value <= 0) { message.error('请输入有效金额'); return; }
      try {
        await agentDeductApi({ uid, money: money.value });
        message.success('扣除成功');
        loadData();
      } catch (e: any) {
        message.error(e?.message || '扣除失败');
      }
    },
  });
}

// ===== 封禁/解封 =====
function handleChangeStatus(uid: number, active: number) {
  const actionText = active === 1 ? '封禁' : '解封';
  Modal.confirm({
    title: '提示',
    content: `确定要${actionText}该用户吗？`,
    async onOk() {
      try {
        await agentChangeStatusApi({ uid, active });
        message.success('操作成功');
        loadData();
      } catch (e: any) {
        message.error(e?.message || '操作失败');
      }
    },
  });
}

// ===== 重置密码 =====
function handleResetPassword(uid: number) {
  Modal.confirm({
    title: '提示',
    content: '确定要重置该用户的密码吗？',
    async onOk() {
      try {
        const res = await agentResetPasswordApi({ uid });
        message.success((res as any)?.message || '重置成功');
        loadData();
      } catch (e: any) {
        message.error(e?.message || '重置失败');
      }
    },
  });
}

// ===== 开通密钥 =====
function handleOpenKey(uid: number) {
  Modal.confirm({
    title: '提示',
    content: '确定要为该代理开通密钥吗？(扣费5元)',
    async onOk() {
      try {
        await agentOpenKeyApi({ uid });
        message.success('开通成功');
        loadData();
      } catch (e: any) {
        message.error(e?.message || '开通失败');
      }
    },
  });
}

// ===== 设置邀请码 =====
function handleSetInviteCode(uid: number) {
  const yqm = ref('');
  Modal.confirm({
    title: '设置邀请码',
    content: () => h(Input, {
      value: yqm.value,
      'onUpdate:value': (v: string) => { yqm.value = v; },
      placeholder: '邀请码最低4位',
    }),
    async onOk() {
      if (!yqm.value || yqm.value.length < 4) { message.error('邀请码最少4位'); return; }
      try {
        await agentSetInviteCodeApi({ uid, yqm: yqm.value });
        message.success('设置成功');
        loadData();
      } catch (e: any) {
        message.error(e?.message || '设置失败');
      }
    },
  });
}

// ===== 跨户充值 =====
const crossRechargeAllowed = ref(false);
const crossVisible = ref(false);
const crossForm = ref({ uid: 0, targetName: '', money: 0 as number });
const crossLoading = ref(false);

async function loadCrossRechargePermission() {
  try {
    const res = await agentCrossRechargeCheckApi();
    crossRechargeAllowed.value = res?.allowed ?? false;
  } catch { /* ignore */ }
}

function openCrossRecharge(uid: number, name: string) {
  crossForm.value = { uid, targetName: name || String(uid), money: 0 };
  crossVisible.value = true;
}

async function handleCrossRecharge() {
  if (!crossForm.value.money || crossForm.value.money <= 0) {
    message.error('请输入有效金额');
    return;
  }
  crossLoading.value = true;
  try {
    await agentCrossRechargeApi({ uid: crossForm.value.uid, money: crossForm.value.money });
    message.success('跨户充值成功');
    crossVisible.value = false;
    loadData();
  } catch (e: any) {
    message.error(e?.message || '跨户充值失败');
  } finally {
    crossLoading.value = false;
  }
}

// ===== 表格列 =====
const columns = computed(() => {
  const cols: any[] = [];
  if (isAdmin.value) {
    cols.push({ title: '上级', dataIndex: 'uuid', key: 'uuid', width: 80, align: 'center' });
  }
  cols.push(
    { title: 'UID', dataIndex: 'uid', key: 'uid', width: 80, align: 'center' },
    { title: '头像', key: 'avatar', width: 60, align: 'center' },
    { title: '昵称', dataIndex: 'name', key: 'name', width: 100, align: 'center' },
    { title: '账号', dataIndex: 'user', key: 'user', width: 120, align: 'center' },
    { title: '余额', dataIndex: 'money', key: 'money', width: 100, align: 'center' },
    { title: '总充值', dataIndex: 'zcz', key: 'zcz', width: 100, align: 'center' },
    { title: '等级', dataIndex: 'addprice', key: 'addprice', width: 80, align: 'center' },
    { title: '订单量', dataIndex: 'dd', key: 'dd', width: 80, align: 'center' },
  );
  if (isAdmin.value) {
    cols.push({ title: '状态', key: 'status', width: 80, align: 'center' });
  }
  cols.push(
    { title: '密钥', key: 'keyStatus', width: 80, align: 'center' },
    { title: '邀请码', key: 'yqmCol', width: 100, align: 'center' },
    { title: '在线时间', dataIndex: 'endtime', key: 'endtime', width: 140, align: 'center' },
    { title: '添加时间', dataIndex: 'addtime', key: 'addtime', width: 140, align: 'center' },
    { title: '操作', key: 'action', width: 150, align: 'center', fixed: 'right' },
  );
  return cols;
});

onMounted(() => {
  loadData();
  loadGrades();
  loadCrossRechargePermission();
});
</script>

<template>
  <Page title="代理管理" content-class="p-4">
    <Spin :spinning="loading">
      <!-- 搜索栏 -->
      <Card class="mb-4">
        <div class="flex flex-wrap items-center gap-3">
          <Select v-model:value="searchType" style="width: 140px" placeholder="搜索类型">
            <SelectOption value="1">UID</SelectOption>
            <SelectOption value="2">用户名</SelectOption>
            <SelectOption value="3">邀请码</SelectOption>
            <SelectOption value="4">昵称</SelectOption>
            <SelectOption value="5">等级</SelectOption>
            <SelectOption value="6">余额</SelectOption>
            <SelectOption value="7">最后在线时间</SelectOption>
          </Select>
          <Input
            v-model:value="searchKeywords"
            placeholder="请输入查询内容"
            style="width: 240px"
            @press-enter="handleSearch"
          />
          <Button type="primary" @click="handleSearch">
            <template #icon><SearchOutlined /></template>
            搜索
          </Button>
          <Button @click="() => { searchKeywords = ''; handleSearch(); }">
            <template #icon><ReloadOutlined /></template>
            重置
          </Button>
          <Button type="primary" @click="addVisible = true">
            <template #icon><PlusOutlined /></template>
            添加代理
          </Button>
          <Button v-if="isAdmin || crossRechargeAllowed" type="primary" @click="crossVisible = true; crossForm = { uid: 0, targetName: '', money: 0 }">
            跨户充值
          </Button>
        </div>
      </Card>

      <!-- 表格 -->
      <Card>
        <Table
          :columns="columns"
          :data-source="tableData"
          :pagination="false"
          row-key="uid"
          size="small"
          :scroll="{ x: 1400 }"
          bordered
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'avatar'">
              <img
                :src="`//q2.qlogo.cn/headimg_dl?dst_uin=${record.user}&spec=640`"
                class="h-8 w-8 rounded-full"
                alt=""
              />
            </template>
            <template v-else-if="column.key === 'status'">
              <Tag
                :color="record.active === 1 ? 'green' : 'red'"
                class="cursor-pointer"
                @click="handleChangeStatus(record.uid, record.active)"
              >
                {{ record.active === 1 ? '正常' : '封禁' }}
              </Tag>
            </template>
            <template v-else-if="column.key === 'keyStatus'">
              <Tag v-if="record.key === 1" color="green">已开通</Tag>
              <Tag v-else color="red" class="cursor-pointer" @click="handleOpenKey(record.uid)">未开通</Tag>
            </template>
            <template v-else-if="column.key === 'yqmCol'">
              <Tag v-if="record.yqm" color="green">{{ record.yqm }}</Tag>
              <Tag v-else color="red" class="cursor-pointer" @click="handleSetInviteCode(record.uid)">无</Tag>
            </template>
            <template v-else-if="column.key === 'action'">
              <Space :size="2" wrap>
                <Button size="small" class="action-btn" @click="openChangeGrade(record.uid)">改价</Button>
                <Button size="small" class="action-btn" @click="handleRecharge(record.uid)">充值</Button>
                <Button v-if="isAdmin" size="small" class="action-btn" danger @click="handleDeduct(record.uid)">扣款</Button>
                <Button v-if="isAdmin" size="small" class="action-btn" @click="handleResetPassword(record.uid)">重置</Button>
                <Button v-if="isAdmin" size="small" class="action-btn" type="primary" @click="handleImpersonate(record.uid)">进入</Button>
              </Space>
            </template>
          </template>
        </Table>

        <div class="mt-4 flex justify-center">
          <Pagination
            v-model:current="currentPage"
            :total="total"
            :page-size="pageSize"
            :page-size-options="['20', '50', '100']"
            show-size-changer
            :show-total="(t: number) => `共 ${t} 条`"
            @change="handlePageChange"
            @show-size-change="handleSizeChange"
          />
        </div>
      </Card>
    </Spin>

    <!-- 添加代理弹窗 -->
    <Modal v-model:open="addVisible" title="添加代理" @ok="handleCreateAgent" :confirm-loading="addLoading">
      <div class="space-y-3 py-2">
        <div>
          <label class="mb-1 block text-sm font-medium">用户昵称</label>
          <Input v-model:value="addForm.nickname" placeholder="请输入用户昵称" />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium">用户账号</label>
          <Input v-model:value="addForm.user" placeholder="请输入用户账号（QQ号）" />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium">用户密码</label>
          <Input v-model:value="addForm.pass" placeholder="请输入用户密码" />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium">用户等级</label>
          <Select v-model:value="addForm.gradeId" style="width: 100%" placeholder="请选择等级">
            <SelectOption v-for="g in gradeList" :key="g.id" :value="g.id">
              {{ g.name }} - {{ g.rate }} - {{ g.money }}元
            </SelectOption>
          </Select>
        </div>
      </div>
    </Modal>

    <!-- 跨户充值弹窗 -->
    <Modal v-model:open="crossVisible" title="跨户充值" @ok="handleCrossRecharge" :confirm-loading="crossLoading" ok-text="确认充值">
      <div class="space-y-3 py-2">
        <div>
          <label class="mb-1 block text-sm font-medium">目标用户 UID</label>
          <InputNumber v-model:value="crossForm.uid" :min="1" placeholder="请输入目标用户UID" style="width: 100%" />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium">充值金额</label>
          <InputNumber v-model:value="crossForm.money" :min="1" :precision="2" placeholder="请输入充值金额" style="width: 100%" />
        </div>
        <div class="rounded bg-orange-50 dark:bg-orange-900/20 p-3 text-xs text-orange-600 dark:text-orange-400">
          提示：实际扣费 = 充值金额 × (您的费率 / 目标用户费率)，充值金额直接充入目标账户。
        </div>
      </div>
    </Modal>

    <!-- 修改等级弹窗 -->
    <Modal v-model:open="gradeVisible" title="修改等级" @ok="handleChangeGrade" :confirm-loading="gradeLoading">
      <div class="py-2">
        <Select v-model:value="gradeSelected" style="width: 100%" placeholder="请选择等级">
          <SelectOption v-for="g in gradeList" :key="g.id" :value="g.id">
            {{ g.name }} - {{ g.rate }} - {{ g.money }}元
          </SelectOption>
        </Select>
      </div>
    </Modal>
  </Page>
</template>

<style scoped>
.action-btn {
  padding: 0 6px !important;
  font-size: 12px !important;
  height: 22px !important;
  line-height: 20px !important;
}
</style>

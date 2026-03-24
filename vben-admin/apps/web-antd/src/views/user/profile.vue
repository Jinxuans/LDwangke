<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Row, Col, Statistic, Descriptions, DescriptionsItem, Button,
  Input, InputNumber, Modal, message, Spin, Tag, Popconfirm, Alert, Space,
  Select, SelectOption,
} from 'ant-design-vue';
import {
  WalletOutlined, ShoppingCartOutlined,
  KeyOutlined, CopyOutlined, ApiOutlined, TeamOutlined,
  LinkOutlined, BellOutlined,
} from '@ant-design/icons-vue';
import {
  getUserProfileApi, changePasswordApi, changePass2Api, setInviteRateApi, changeSecretKeyApi,
  setPushTokenApi, getUserGradeListApi, setMyGradeApi, setInviteCodeApi,
  migrateSuperiorApi,
  type UserProfile, type GradeOption,
} from '#/api/user-center';
import { getSiteConfigApi } from '#/api/admin';

const loading = ref(false);
const profile = ref<UserProfile | null>(null);

// 等级设置
const gradeVisible = ref(false);
const gradeLoading = ref(false);
const gradeOptions = ref<GradeOption[]>([]);
const selectedGrade = ref<number | undefined>(undefined);

// 邀请链接
const inviteUrl = computed(() => {
  if (!profile.value?.yqm) return '';
  const base = window.location.href.includes('#') ? `${window.location.origin}/#` : window.location.origin;
  return `${base}/auth/register?invite=${profile.value.yqm}`;
});

// 改密弹窗
const passVisible = ref(false);
const passLoading = ref(false);
const oldPass = ref('');
const newPass = ref('');
const newPass2 = ref('');

// 二级密码弹窗
const pass2Visible = ref(false);
const pass2Loading = ref(false);
const oldPass2 = ref('');
const newPass2Input = ref('');
const newPass2Confirm = ref('');

// 邀请费率
const rateVisible = ref(false);
const rateLoading = ref(false);
const inviteGradeId = ref<number | undefined>(undefined);

// 邀请码设置
const yqmVisible = ref(false);
const yqmLoading = ref(false);
const yqmInput = ref('');

// 推送Token
const pushTokenVisible = ref(false);
const pushTokenLoading = ref(false);
const pushTokenInput = ref('');

// 上级迁移
const migrateEnabled = ref(false);
const migrateVisible = ref(false);
const migrateLoading = ref(false);
const migrateUid = ref<number | null>(null);
const migrateYqm = ref('');

function formatGradeDisplay(name?: string, rate?: number | null, emptyText: string = '-') {
  if (!name) return emptyText;
  if (typeof rate === 'number' && Number.isFinite(rate)) {
    return `${name}（费率 ${rate.toFixed(2)}）`;
  }
  return name;
}

async function loadProfile() {
  loading.value = true;
  try {
    const raw = await getUserProfileApi();
    profile.value = raw;
    inviteGradeId.value = profile.value.invite_grade_id || profile.value.grade_id || undefined;
    // 检查上级迁移开关
    try {
      const cfg = await getSiteConfigApi();
      migrateEnabled.value = cfg?.sjqykg === '1';
    } catch { /* ignore */ }
  } catch (e) {
    console.error(e);
  } finally {
    loading.value = false;
  }
}

async function handleChangePass() {
  if (!oldPass.value || !newPass.value) { message.warning('请填写完整'); return; }
  if (newPass.value !== newPass2.value) { message.warning('两次输入的新密码不一致'); return; }
  passLoading.value = true;
  try {
    await changePasswordApi(oldPass.value, newPass.value);
    message.success('密码修改成功');
    passVisible.value = false;
    oldPass.value = newPass.value = newPass2.value = '';
  } catch (e: any) { message.error(e?.message || '修改失败'); }
  finally { passLoading.value = false; }
}

async function handleChangePass2() {
  if (!newPass2Input.value) { message.warning('请填写新二级密码'); return; }
  if (newPass2Input.value.length < 6) { message.warning('二级密码至少6位'); return; }
  if (newPass2Input.value !== newPass2Confirm.value) { message.warning('两次输入的密码不一致'); return; }
  pass2Loading.value = true;
  try {
    await changePass2Api(oldPass2.value, newPass2Input.value);
    message.success('二级密码修改成功');
    pass2Visible.value = false;
    oldPass2.value = newPass2Input.value = newPass2Confirm.value = '';
  } catch (e: any) { message.error(e?.message || '修改失败'); }
  finally { pass2Loading.value = false; }
}

async function handleSetRate() {
  if (!inviteGradeId.value) { message.warning('请选择邀请等级'); return; }
  rateLoading.value = true;
  try {
    await setInviteRateApi(inviteGradeId.value);
    message.success('设置成功');
    rateVisible.value = false;
    loadProfile();
  } catch (e: any) { message.error(e?.message || '设置失败'); }
  finally { rateLoading.value = false; }
}

async function handleKeyAction(type: number) {
  try {
    const raw = await changeSecretKeyApi(type);
    const res = raw;
    message.success(`操作成功，新密钥: ${res.key}`);
    loadProfile();
  } catch (e: any) { message.error(e?.message || '操作失败'); }
}

async function handleSetPushToken() {
  pushTokenLoading.value = true;
  try {
    await setPushTokenApi(pushTokenInput.value);
    message.success('设置成功');
    pushTokenVisible.value = false;
    loadProfile();
  } catch (e: any) { message.error(e?.message || '设置失败'); }
  finally { pushTokenLoading.value = false; }
}

async function handleSetYqm() {
  if (!yqmInput.value || yqmInput.value.length < 4) { message.warning('邀请码最少4位'); return; }
  yqmLoading.value = true;
  try {
    await setInviteCodeApi(yqmInput.value);
    message.success('邀请码设置成功');
    yqmVisible.value = false;
    loadProfile();
  } catch (e: any) { message.error(e?.message || '设置失败'); }
  finally { yqmLoading.value = false; }
}

function openPushTokenModal() {
  pushTokenInput.value = profile.value?.push_token || '';
  pushTokenVisible.value = true;
}

async function openGradeModal() {
  try {
    const raw = await getUserGradeListApi();
    const data = raw;
    gradeOptions.value = Array.isArray(data) ? data : [];
  } catch (e) {}
  selectedGrade.value = profile.value?.grade_id || undefined;
  gradeVisible.value = true;
}

async function openInviteGradeModal() {
  try {
    const raw = await getUserGradeListApi();
    const data = raw;
    gradeOptions.value = Array.isArray(data) ? data : [];
  } catch (e) {}
  inviteGradeId.value = profile.value?.invite_grade_id || profile.value?.grade_id || undefined;
  rateVisible.value = true;
}

async function handleSetGrade() {
  if (!selectedGrade.value) { message.warning('请选择等级'); return; }
  gradeLoading.value = true;
  try {
    await setMyGradeApi(selectedGrade.value);
    message.success('等级已更新');
    gradeVisible.value = false;
    loadProfile();
  } catch (e: any) { message.error(e?.message || '设置失败'); }
  finally { gradeLoading.value = false; }
}

async function handleMigrate() {
  if (!migrateUid.value || !migrateYqm.value) { message.warning('请填写完整'); return; }
  migrateLoading.value = true;
  try {
    const raw = await migrateSuperiorApi(migrateUid.value, migrateYqm.value);
    const res = raw;
    message.success(res?.msg || '迁移成功');
    migrateVisible.value = false;
    migrateUid.value = null;
    migrateYqm.value = '';
    loadProfile();
  } catch (e: any) { message.error(e?.message || '迁移失败'); }
  finally { migrateLoading.value = false; }
}

function copyText(text: string) {
  navigator.clipboard.writeText(text);
  message.success('已复制');
}

onMounted(loadProfile);
</script>

<template>
  <Page title="我的资料" content-class="p-4">
    <Spin :spinning="loading">
      <template v-if="profile">
        <!-- 统计卡片 -->
        <Row :gutter="[16, 16]" class="mb-4">
          <Col :xs="12" :sm="8" :lg="4">
            <Card><Statistic title="账户余额" :value="profile.money" :precision="2" prefix="¥" /></Card>
          </Col>
          <Col :xs="12" :sm="8" :lg="4" v-if="profile.cdmoney > 0">
            <Card><Statistic title="储值金额" :value="profile.cdmoney" :precision="2" prefix="¥" /></Card>
          </Col>
          <Col :xs="12" :sm="8" :lg="4">
            <Card><Statistic title="总充值" :value="profile.zcz" :precision="2" prefix="¥" /></Card>
          </Col>
          <Col :xs="12" :sm="8" :lg="4">
            <Card><Statistic title="总订单" :value="profile.order_total" /></Card>
          </Col>
          <Col :xs="12" :sm="8" :lg="4">
            <Card><Statistic title="今日订单" :value="profile.today_orders" /></Card>
          </Col>
          <Col :xs="12" :sm="8" :lg="4">
            <Card><Statistic title="今日消费" :value="profile.today_spend" :precision="2" prefix="¥" /></Card>
          </Col>
          <Col :xs="12" :sm="8" :lg="4" v-if="profile.dailitongji">
            <Card><Statistic title="代理总数" :value="profile.dailitongji.dlzs" /></Card>
          </Col>
        </Row>

        <!-- 系统公告 -->
        <Alert v-if="profile.notice" :message="profile.notice" type="info" show-icon class="mb-4" />
        <Alert v-if="profile.sjnotice" :message="'上级公告: ' + profile.sjnotice" type="warning" show-icon class="mb-4" />

        <!-- 基本信息 -->
        <Card title="基本信息" class="mb-4">
          <Descriptions bordered :column="{ xs: 1, sm: 2 }" size="small">
            <DescriptionsItem label="UID">{{ profile.uid }}</DescriptionsItem>
            <DescriptionsItem label="账号">{{ profile.user }}</DescriptionsItem>
            <DescriptionsItem label="昵称">{{ profile.name || '-' }}</DescriptionsItem>
            <DescriptionsItem label="上级">{{ profile.sjuser || '无' }}</DescriptionsItem>
            <DescriptionsItem label="邮箱">{{ profile.email || '-' }}</DescriptionsItem>
            <DescriptionsItem label="我的等级">
              <Tag color="blue">{{ formatGradeDisplay(profile.grade_name, profile.addprice, `费率 ${Number(profile.addprice || 0).toFixed(2)}`) }}</Tag>
              <Button v-if="profile.grade === '2' || profile.grade === '3'" type="link" size="small" @click="openGradeModal">设置</Button>
            </DescriptionsItem>
            <DescriptionsItem label="邀请码">
              <div class="flex items-center gap-1 flex-wrap">
                <code v-if="profile.yqm" class="bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded text-sm">{{ profile.yqm }}</code>
                <span v-else class="text-gray-400">未设置</span>
                <Button v-if="profile.yqm" type="link" size="small" @click="copyText(profile.yqm)"><CopyOutlined /></Button>
                <Button type="link" size="small" @click="yqmInput = profile.yqm || ''; yqmVisible = true">{{ profile.yqm ? '修改' : '设置' }}</Button>
              </div>
            </DescriptionsItem>
            <DescriptionsItem label="邀请等级">
              {{ formatGradeDisplay(profile.invite_grade_name, profile.invite_addprice, '未设置') }}
              <Button type="link" size="small" @click="openInviteGradeModal">设置</Button>
            </DescriptionsItem>
            <DescriptionsItem label="邀请链接">
              <div v-if="inviteUrl" class="flex items-center gap-1 flex-wrap">
                <code class="bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded text-xs break-all" style="max-width: 90vw">{{ inviteUrl }}</code>
                <Button type="link" size="small" @click="copyText(inviteUrl)"><CopyOutlined /> 复制链接</Button>
              </div>
              <span v-else class="text-gray-400">请先设置邀请码</span>
            </DescriptionsItem>
            <DescriptionsItem label="跨户充值权限">
              <Tag :color="profile.khcz ? 'green' : 'red'">{{ profile.khcz ? '已开通' : '关闭' }}</Tag>
            </DescriptionsItem>
          </Descriptions>
        </Card>

        <!-- API对接 -->
        <Card title="API对接" class="mb-4">
          <Descriptions bordered :column="1" size="small">
            <DescriptionsItem label="对接账号">
              <div class="flex items-center gap-2">
                <code class="bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded text-sm">{{ profile.uid }}</code>
                <Button type="link" size="small" @click="copyText(String(profile.uid))"><CopyOutlined /></Button>
              </div>
            </DescriptionsItem>
            <DescriptionsItem label="对接密钥">
              <div class="flex items-center gap-2">
                <code v-if="profile.key && profile.key !== '0'" class="bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded text-sm">{{ profile.key }}</code>
                <span v-else class="text-gray-400">未开通</span>
                <Button v-if="profile.key && profile.key !== '0'" type="link" size="small" @click="copyText(profile.key)"><CopyOutlined /></Button>
              </div>
            </DescriptionsItem>
            <DescriptionsItem label="推送Token">
              <div class="flex items-center gap-2">
                <code v-if="profile.push_token" class="bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded text-sm">{{ profile.push_token }}</code>
                <span v-else class="text-gray-400">未设置</span>
              </div>
            </DescriptionsItem>
          </Descriptions>
          <div class="mt-3 flex gap-2 flex-wrap">
            <Popconfirm v-if="!profile.key || profile.key === '0'" title="开通需要5元（余额≥100免费），确定？" @confirm="handleKeyAction(1)">
              <Button type="primary" size="small"><ApiOutlined /> 开通密钥</Button>
            </Popconfirm>
            <Popconfirm v-else title="确定更换密钥？旧密钥将失效" @confirm="handleKeyAction(3)">
              <Button size="small"><KeyOutlined /> 更换密钥</Button>
            </Popconfirm>
            <Button size="small" @click="openPushTokenModal"><BellOutlined /> 设置推送Token</Button>
          </div>
        </Card>

        <!-- 代理统计 -->
        <Card v-if="profile.dailitongji" title="代理统计" class="mb-4">
          <Row :gutter="[16, 16]">
            <Col :xs="12" :sm="6"><Statistic title="代理总数" :value="profile.dailitongji.dlzs"><template #prefix><TeamOutlined /></template></Statistic></Col>
            <Col :xs="12" :sm="6"><Statistic title="今日活跃" :value="profile.dailitongji.dldl" /></Col>
            <Col :xs="12" :sm="6"><Statistic title="今日新增" :value="profile.dailitongji.dlzc" /></Col>
            <Col :xs="12" :sm="6"><Statistic title="今日交单" :value="profile.dailitongji.jrjd" /></Col>
          </Row>
        </Card>

        <!-- 操作 -->
        <Card title="账户操作">
          <Space wrap>
            <Button type="primary" @click="passVisible = true"><KeyOutlined /> 修改密码</Button>
            <Button v-if="profile?.grade === '3'" type="primary" ghost @click="pass2Visible = true"><KeyOutlined /> 二级密码</Button>
            <Button @click="openInviteGradeModal"><TeamOutlined /> 设置邀请等级</Button>
            <Button v-if="profile?.grade === '2' || profile?.grade === '3'" @click="openGradeModal">设置我的等级</Button>
            <Button v-if="migrateEnabled" @click="migrateVisible = true"><LinkOutlined /> 上级迁移</Button>
          </Space>
        </Card>
      </template>
    </Spin>

    <!-- 修改密码弹窗 -->
    <Modal v-model:open="passVisible" title="修改密码" :confirm-loading="passLoading" @ok="handleChangePass" ok-text="确认修改">
      <div class="space-y-4">
        <div><label class="block text-sm font-medium mb-1">旧密码</label><Input.Password v-model:value="oldPass" placeholder="请输入旧密码" /></div>
        <div><label class="block text-sm font-medium mb-1">新密码</label><Input.Password v-model:value="newPass" placeholder="请输入新密码（至少6位）" /></div>
        <div><label class="block text-sm font-medium mb-1">确认新密码</label><Input.Password v-model:value="newPass2" placeholder="请再次输入新密码" /></div>
      </div>
    </Modal>

    <!-- 二级密码弹窗 -->
    <Modal v-model:open="pass2Visible" title="设置/修改二级密码" :confirm-loading="pass2Loading" @ok="handleChangePass2" ok-text="确认修改">
      <Alert message="二级密码用于管理员登录二次验证，首次设置无需填写旧密码" type="info" show-icon class="mb-4" />
      <div class="space-y-4">
        <div><label class="block text-sm font-medium mb-1">旧二级密码</label><Input.Password v-model:value="oldPass2" placeholder="首次设置可留空" /></div>
        <div><label class="block text-sm font-medium mb-1">新二级密码</label><Input.Password v-model:value="newPass2Input" placeholder="请输入新二级密码（至少6位）" /></div>
        <div><label class="block text-sm font-medium mb-1">确认新密码</label><Input.Password v-model:value="newPass2Confirm" placeholder="请再次输入新二级密码" /></div>
      </div>
    </Modal>

    <!-- 邀请等级弹窗 -->
    <Modal v-model:open="rateVisible" title="设置邀请等级" :confirm-loading="rateLoading" @ok="handleSetRate" ok-text="保存">
      <Alert message="受邀用户注册后会使用这里选定的等级费率" type="info" show-icon class="mb-4" />
      <Select v-model:value="inviteGradeId" style="width: 100%" show-search :filter-option="(input: string, option: any) => option.label?.toLowerCase().includes(input.toLowerCase())">
        <SelectOption
          v-for="g in gradeOptions"
          :key="g.id"
          :value="g.id"
          :label="`${g.name}（${g.rate}）`"
        >
          {{ g.name }}（费率 {{ g.rate }}）
        </SelectOption>
      </Select>
    </Modal>

    <!-- 邀请码弹窗 -->
    <Modal v-model:open="yqmVisible" title="设置邀请码" :confirm-loading="yqmLoading" @ok="handleSetYqm" ok-text="保存">
      <Alert message="邀请码最少4位，设置后其他人可通过邀请链接注册为你的下级" type="info" show-icon class="mb-4" />
      <Input v-model:value="yqmInput" placeholder="请输入邀请码（最少4位）" allow-clear />
    </Modal>

    <!-- 推送Token弹窗 -->
    <Modal v-model:open="pushTokenVisible" title="设置推送Token" :confirm-loading="pushTokenLoading" @ok="handleSetPushToken" ok-text="保存">
      <Alert message="设置后订单状态变更会通过推送通知" type="info" show-icon class="mb-4" />
      <Input v-model:value="pushTokenInput" placeholder="请输入推送Token" allow-clear />
    </Modal>

    <!-- 上级迁移弹窗 -->
    <Modal v-model:open="migrateVisible" title="上级迁移" :confirm-loading="migrateLoading" @ok="handleMigrate" ok-text="确认迁移">
      <Alert message="输入新上级的UID和邀请码，原上级需7天内无登录记录才可迁移" type="warning" show-icon class="mb-4" />
      <div class="space-y-4">
        <div><label class="block text-sm font-medium mb-1">新上级UID</label><InputNumber v-model:value="migrateUid" placeholder="输入新上级的UID" :min="1" style="width: 100%" /></div>
        <div><label class="block text-sm font-medium mb-1">新上级邀请码</label><Input v-model:value="migrateYqm" placeholder="输入新上级的邀请码" /></div>
      </div>
    </Modal>

    <!-- 等级设置弹窗 -->
    <Modal v-model:open="gradeVisible" title="设置我的等级" :confirm-loading="gradeLoading" @ok="handleSetGrade" ok-text="保存">
      <Alert message="选择等级后将更新你的费率，影响课程定价" type="warning" show-icon class="mb-4" />
      <Select v-model:value="selectedGrade" style="width: 100%" show-search :filter-option="(input: string, option: any) => option.label?.toLowerCase().includes(input.toLowerCase())">
        <SelectOption
          v-for="g in gradeOptions"
          :key="g.id"
          :value="g.id"
          :label="`${g.name}（${g.rate}）`"
        >
          {{ g.name }}（费率 {{ g.rate }}）
        </SelectOption>
      </Select>
    </Modal>
  </Page>
</template>

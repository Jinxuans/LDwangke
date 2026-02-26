<script setup lang="ts">
import { ref, reactive, onUnmounted } from 'vue';
import {
  Card, Input, Button, Table, Tag, Space, Empty, message, Modal, Form, Tooltip, QRCode
} from 'ant-design-vue';
import { SearchOutlined, QrcodeOutlined, MailOutlined, LinkOutlined, LoginOutlined } from '@ant-design/icons-vue';
import { checkOrderApi, type CheckOrderResult, pushBindWxApi, pushUnbindWxApi, pushBindEmailApi, pushUnbindEmailApi, pushBindShowDocApi, pushUnbindShowDocApi, pushWxQrcodeApi, pushWxScanUidApi, pushPupLoginApi } from '#/api/auxiliary';

const searchUser = ref('');
const searchOid = ref('');
const loading = ref(false);
const list = ref<CheckOrderResult[]>([]);
const searched = ref(false);

const statusMap: Record<string, { text: string; color: string }> = {
  '等待中': { text: '等待中', color: 'default' },
  '进行中': { text: '进行中', color: 'processing' },
  '已完成': { text: '已完成', color: 'success' },
  '已退款': { text: '已退款', color: 'warning' },
  '异常': { text: '异常', color: 'error' },
};

const columns = [
  { title: '订单号', dataIndex: 'oid', key: 'oid', width: 90, align: 'center' as const },
  { title: '平台', dataIndex: 'ptname', key: 'ptname', width: 120, ellipsis: true },
  { title: '账号信息', dataIndex: 'account', key: 'account', width: 160, ellipsis: true },
  { title: '课程名称', dataIndex: 'kcname', key: 'kcname', ellipsis: true },
  { title: '状态', key: 'status', width: 100, align: 'center' as const },
  { title: '进度', dataIndex: 'process', key: 'process', width: 100, align: 'center' as const },
  { title: '备注', dataIndex: 'remarks', key: 'remarks', ellipsis: true, width: 160 },
  { title: '操作', key: 'action', width: 220, align: 'center' as const },
];

// 微信扫码绑定相关
const wxModalVisible = ref(false);
const wxQrUrl = ref('');
const wxQrCode = ref('');
const wxScanTimer = ref<any>(null);
const currentAccount = ref('');

async function openWxBind(account: string) {
  if (!account) {
    message.warning('无法获取账号信息');
    return;
  }
  currentAccount.value = account;
  try {
    const res = await pushWxQrcodeApi({ account });
    wxQrUrl.value = res.url;
    wxQrCode.value = res.code;
    wxModalVisible.value = true;
    startWxScanPoll();
  } catch (e: any) {
    message.error(e?.message || '获取二维码失败');
  }
}

function startWxScanPoll() {
  stopWxScanPoll();
  wxScanTimer.value = setInterval(async () => {
    if (!wxModalVisible.value || !wxQrCode.value) return;
    try {
      const res = await pushWxScanUidApi({ code: wxQrCode.value });
      if (res.uid) {
        stopWxScanPoll();
        // 执行绑定
        const oids = list.value.filter(item => item.account === currentAccount.value).map(item => item.oid).join(',');
        await pushBindWxApi({ account: currentAccount.value, pushUid: res.uid, oids });
        message.success('微信推送绑定成功');
        wxModalVisible.value = false;
        handleSearch(); // 刷新列表
      }
    } catch (e) {
      // 忽略未扫码错误
    }
  }, 3000);
}

function stopWxScanPoll() {
  if (wxScanTimer.value) {
    clearInterval(wxScanTimer.value);
    wxScanTimer.value = null;
  }
}

onUnmounted(() => {
  stopWxScanPoll();
});

async function handleUnbindWx(account: string) {
  try {
    await pushUnbindWxApi({ account });
    message.success('已解绑微信推送');
    handleSearch();
  } catch (e: any) {
    message.error(e?.message || '解绑失败');
  }
}

// 邮箱绑定相关
const emailModalVisible = ref(false);
const emailForm = reactive({ account: '', email: '' });

function openEmailBind(account: string) {
  emailForm.account = account;
  emailForm.email = '';
  emailModalVisible.value = true;
}

async function submitEmailBind() {
  if (!emailForm.email) {
    message.warning('请输入邮箱');
    return;
  }
  try {
    await pushBindEmailApi({ account: emailForm.account, pushEmail: emailForm.email });
    message.success('邮箱推送绑定成功');
    emailModalVisible.value = false;
    handleSearch();
  } catch (e: any) {
    message.error(e?.message || '绑定失败');
  }
}

async function handleUnbindEmail(account: string) {
  try {
    await pushUnbindEmailApi({ account });
    message.success('已解绑邮箱推送');
    handleSearch();
  } catch (e: any) {
    message.error(e?.message || '解绑失败');
  }
}

// Pup 登录
async function handlePupLogin(oid: number) {
  try {
    const res = await pushPupLoginApi(oid);
    if (res.url) {
      window.open(res.url, '_blank');
    }
  } catch (e: any) {
    message.error(e?.message || '获取登录地址失败');
  }
}

async function handleSearch() {
  const user = searchUser.value.trim();
  const oid = searchOid.value.trim();
  if (!user && !oid) {
    message.warning('请输入账号或订单号');
    return;
  }
  loading.value = true;
  searched.value = true;
  try {
    const res = await checkOrderApi({ user: user || undefined, oid: oid || undefined });
    list.value = res.list || [];
  } catch (e: any) {
    message.error(e?.message || '查询失败');
    list.value = [];
  } finally {
    loading.value = false;
  }
}

function getStatusInfo(status: string) {
  return statusMap[status] || { text: status, color: 'default' };
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 flex items-start justify-center pt-8 sm:pt-16 px-4">
    <div class="w-full max-w-3xl">
      <div class="text-center mb-6">
        <h1 class="text-2xl sm:text-3xl font-bold text-gray-800 mb-2">订单查询</h1>
        <p class="text-gray-500 text-sm">输入您的账号或订单号查询课程进度</p>
      </div>

      <Card class="mb-4 shadow-sm">
        <div class="flex flex-col sm:flex-row gap-3">
          <Input
            v-model:value="searchUser" placeholder="输入账号"
            allow-clear size="large" class="flex-1"
            @press-enter="handleSearch"
          />
          <Input
            v-model:value="searchOid" placeholder="或输入订单号"
            allow-clear size="large" class="flex-1"
            @press-enter="handleSearch"
          />
          <Button type="primary" size="large" :loading="loading" @click="handleSearch"
                  class="sm:w-auto w-full">
            <template #icon><SearchOutlined /></template>
            查询
          </Button>
        </div>
      </Card>

      <Card v-if="searched" class="shadow-sm">
        <template v-if="list.length === 0 && !loading">
          <Empty description="未找到相关订单" />
        </template>

        <!-- 移动端卡片布局 -->
        <div class="block sm:hidden space-y-3">
          <Card v-for="item in list" :key="item.oid" size="small" class="border">
            <div class="space-y-2 text-sm">
              <div class="flex justify-between">
                <span class="text-gray-500">订单号</span>
                <span class="font-mono font-medium">{{ item.oid }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500">平台</span>
                <span>{{ item.ptname }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500">账号信息</span>
                <span class="text-right max-w-[60%] truncate">{{ item.account }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500">课程</span>
                <span class="text-right max-w-[60%] truncate">{{ item.kcname }}</span>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-gray-500">状态</span>
                <Tag :color="getStatusInfo(item.status).color">{{ getStatusInfo(item.status).text }}</Tag>
              </div>
              <div class="flex justify-between" v-if="item.process">
                <span class="text-gray-500">进度</span>
                <span>{{ item.process }}</span>
              </div>
              <div class="flex justify-between" v-if="item.remarks">
                <span class="text-gray-500">备注</span>
                <span class="text-right max-w-[60%] truncate">{{ item.remarks }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500">时间</span>
                <span class="text-xs">{{ item.addtime }}</span>
              </div>
              <div class="pt-2 border-t flex flex-wrap gap-2 justify-end">
                <template v-if="item.pushUid">
                  <Tag color="green">微信推送已绑定</Tag>
                  <Button size="small" type="link" danger @click="handleUnbindWx(item.account || '')">解绑微信</Button>
                </template>
                <template v-else>
                  <Button size="small" @click="openWxBind(item.account || '')">绑定微信</Button>
                </template>

                <template v-if="item.pushEmail">
                  <Tag color="blue">邮箱推送已绑定</Tag>
                  <Button size="small" type="link" danger @click="handleUnbindEmail(item.account || '')">解绑邮箱</Button>
                </template>
                <template v-else>
                  <Button size="small" @click="openEmailBind(item.account || '')">绑定邮箱</Button>
                </template>

                <Button size="small" type="primary" ghost @click="handlePupLogin(item.oid)" v-if="item.status !== '等待中'">
                  <template #icon><LoginOutlined /></template>
                  一键登录
                </Button>
              </div>
            </div>
          </Card>
        </div>

        <!-- 桌面端表格布局 -->
        <div class="hidden sm:block">
          <Table
            :columns="columns" :data-source="list" :loading="loading"
            :pagination="false" row-key="oid" size="small" bordered
            :scroll="{ x: 800 }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'status'">
                <Tag :color="getStatusInfo(record.status).color">
                  {{ getStatusInfo(record.status).text }}
                </Tag>
              </template>
              <template v-else-if="column.key === 'action'">
                <Space direction="vertical" size="small" class="w-full text-left">
                  <Space>
                    <template v-if="record.pushUid">
                      <Tag color="green">已绑微信</Tag>
                      <Button size="small" type="link" danger @click="handleUnbindWx(record.account || '')">解绑</Button>
                    </template>
                    <template v-else>
                      <Button size="small" @click="openWxBind(record.account || '')"><QrcodeOutlined /> 绑定微信</Button>
                    </template>
                  </Space>
                  <Space>
                    <template v-if="record.pushEmail">
                      <Tag color="blue">已绑邮箱</Tag>
                      <Button size="small" type="link" danger @click="handleUnbindEmail(record.account || '')">解绑</Button>
                    </template>
                    <template v-else>
                      <Button size="small" @click="openEmailBind(record.account || '')"><MailOutlined /> 绑定邮箱</Button>
                    </template>
                  </Space>
                  <Button v-if="record.status !== '等待中'" size="small" type="primary" ghost block @click="handlePupLogin(record.oid)">
                    <LoginOutlined /> 一键登录
                  </Button>
                </Space>
              </template>
            </template>
          </Table>
        </div>
      </Card>
    </div>

    <!-- 微信扫码绑定弹窗 -->
    <Modal v-model:open="wxModalVisible" title="绑定微信推送" :footer="null" :width="400" @cancel="stopWxScanPoll">
      <div class="text-center py-4">
        <p class="mb-4 text-gray-500">使用微信扫码，实时接收订单进度通知</p>
        <div v-if="wxQrUrl" class="inline-block p-2 bg-white border rounded shadow-sm">
          <QRCode :value="wxQrUrl" :size="200" />
        </div>
        <p class="mt-4 text-sm text-gray-400">扫码后请在手机上确认，窗口将自动关闭</p>
      </div>
    </Modal>

    <!-- 邮箱绑定弹窗 -->
    <Modal v-model:open="emailModalVisible" title="绑定邮箱推送" @ok="submitEmailBind">
      <div class="py-4">
        <p class="mb-4 text-gray-500">绑定邮箱后，订单进度更新将发送邮件通知您。</p>
        <Form layout="vertical">
          <Form.Item label="接收邮箱">
            <Input v-model:value="emailForm.email" placeholder="请输入您的电子邮箱" />
          </Form.Item>
        </Form>
      </div>
    </Modal>
  </div>
</template>

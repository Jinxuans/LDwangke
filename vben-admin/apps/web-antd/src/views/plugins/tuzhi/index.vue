<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, Space, Tag, Modal, message,
  Select, SelectOption, Form, FormItem, Pagination, Row, Col, Statistic,
  Spin, DatePicker, Switch, CheckboxGroup,
  Popconfirm,
} from 'ant-design-vue';
import {
  PlusOutlined, SyncOutlined, DeleteOutlined,
  SearchOutlined, ClockCircleOutlined,
  PlayCircleOutlined, EditOutlined,
} from '@ant-design/icons-vue';
import {
  tuzhiGetGoodsApi, tuzhiOrderListApi, tuzhiAddOrderApi,
  tuzhiEditOrderApi, tuzhiDeleteOrderApi,
  tuzhiCheckInWorkApi, tuzhiCheckOutWorkApi, tuzhiSyncOrdersApi,
  tuzhiGetSignInfoApi,
} from '#/api/plugins/tuzhi';
import { useAccessStore } from '@vben/stores';
import dayjs from 'dayjs';

const accessStore = useAccessStore();
const isAdmin = computed(() => {
  const codes = accessStore.accessCodes;
  return codes.includes('super') || codes.includes('admin');
});

const loading = ref(false);
const orders = ref<any[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const keyword = ref('');

// 商品列表
const goodsList = ref<any[]>([]);
const goodsLoading = ref(false);

// 下单表单
const addVisible = ref(false);
const addLoading = ref(false);
const addForm = ref<Record<string, any>>({
  goods_id: null,
  username: '',
  password: '',
  nickname: '',
  school: '',
  postname: '',
  address: '',
  address_lat: '',
  address_lng: '',
  work_time: '08:30',
  off_time: '17:30',
  work_days: ['1','2','3','4','5'],
  work_deadline: '',
  holiday_status: 0,
  daily_report: 0,
  weekly_report: 0,
  monthly_report: 0,
  weekly_report_time: 1,
  monthly_report_time: 0,
  is_off_time: 1,
  xz_push_url: '',
  images: '',
  token: '',
  uuid: '',
  user_school_id: '',
  random_phone: '',
});

// 编辑
const editVisible = ref(false);
const editLoading = ref(false);
const editForm = ref<Record<string, any>>({});

// 查看签到信息
const signInfoVisible = ref(false);
const signInfoLoading = ref(false);
const signInfoData = ref<any>(null);

// 周期选项
const weekOptions = [
  { label: '周一', value: '1' },
  { label: '周二', value: '2' },
  { label: '周三', value: '3' },
  { label: '周四', value: '4' },
  { label: '周五', value: '5' },
  { label: '周六', value: '6' },
  { label: '周日', value: '7' },
];

const statusMap: Record<number, { text: string; color: string }> = {
  0: { text: '正常', color: 'blue' },
  1: { text: '打卡中', color: 'processing' },
  2: { text: '关闭', color: 'error' },
  3: { text: '已完成', color: 'success' },
};

const isStatusMap: Record<number, { text: string; color: string }> = {
  0: { text: '失败', color: 'error' },
  1: { text: '正常', color: 'success' },
};

const columns = [
  { title: 'ID', dataIndex: 'id', width: 60 },
  { title: '账号', dataIndex: 'username', width: 130, ellipsis: true },
  { title: '姓名', dataIndex: 'nickname', width: 80 },
  { title: '学校', dataIndex: 'school', width: 120, ellipsis: true },
  { title: '岗位', dataIndex: 'postname', width: 100, ellipsis: true },
  { title: '打卡天数', key: 'days', width: 90 },
  { title: '截至日期', dataIndex: 'work_deadline', width: 100 },
  { title: '状态', key: 'status', width: 80 },
  { title: '打卡状态', key: 'is_status', width: 80 },
  { title: '备注', dataIndex: 'remark', width: 120, ellipsis: true },
  { title: '操作', key: 'action', width: 240, fixed: 'right' as const },
];

// ---------- 加载数据 ----------

async function loadOrders() {
  loading.value = true;
  try {
    const res = await tuzhiOrderListApi({
      page: page.value,
      limit: pageSize.value,
      keyword: keyword.value || undefined,
    });
    orders.value = res.list || [];
    total.value = res.total || 0;
  } catch {
    orders.value = [];
  } finally {
    loading.value = false;
  }
}

async function loadGoods() {
  goodsLoading.value = true;
  try {
    const res = await tuzhiGetGoodsApi();
    goodsList.value = res || [];
  } catch { goodsList.value = []; }
  finally { goodsLoading.value = false; }
}

function onSearch() { page.value = 1; loadOrders(); }

// ---------- 下单 ----------

function openAdd() {
  addForm.value = {
    goods_id: null, username: '', password: '', nickname: '',
    school: '', postname: '', address: '', address_lat: '', address_lng: '',
    work_time: '08:30', off_time: '17:30',
    work_days: ['1','2','3','4','5'], work_deadline: '',
    holiday_status: 0, daily_report: 0, weekly_report: 0, monthly_report: 0,
    weekly_report_time: 1, monthly_report_time: 0, is_off_time: 1,
    xz_push_url: '', images: '', token: '', uuid: '', user_school_id: '', random_phone: '',
  };
  loadGoods();
  addVisible.value = true;
}

async function submitAdd() {
  const f = addForm.value;
  if (!f.goods_id) { message.warning('请选择商品'); return; }
  if (!f.username || !f.password) { message.warning('账号密码不能为空'); return; }
  if (!f.work_deadline) { message.warning('请选择截至日期'); return; }
  addLoading.value = true;
  try {
    const formData = {
      ...f,
      work_days: Array.isArray(f.work_days) ? f.work_days.join(',') : f.work_days,
    };
    const res = await tuzhiAddOrderApi(formData);
    message.success(res?.message || '下单成功');
    addVisible.value = false;
    loadOrders();
  } catch (e: any) {
    message.error(e?.message || '下单失败');
  } finally {
    addLoading.value = false;
  }
}

// ---------- 编辑 ----------

function openEdit(record: any) {
  const workDays = record.work_days
    ? (typeof record.work_days === 'string' ? record.work_days.split(',') : record.work_days)
    : ['1','2','3','4','5'];
  editForm.value = {
    id: record.id,
    goods_id: Number(record.goods_id),
    username: record.username,
    password: record.password,
    nickname: record.nickname || '',
    school: record.school || '',
    postname: record.postname || '',
    address: record.address || '',
    address_lat: record.address_lat || '',
    address_lng: record.address_lng || '',
    work_time: record.work_time || '08:30',
    off_time: record.off_time || '17:30',
    work_days: workDays,
    work_deadline: record.work_deadline || '',
    holiday_status: Number(record.holiday_status) || 0,
    daily_report: Number(record.daily_report) || 0,
    weekly_report: Number(record.weekly_report) || 0,
    monthly_report: Number(record.monthly_report) || 0,
    weekly_report_time: Number(record.weekly_report_time) || 1,
    monthly_report_time: Number(record.monthly_report_time) || 0,
    is_off_time: Number(record.is_off_time) ?? 1,
    xz_push_url: record.xz_push_url || '',
    images: record.images || '',
    token: record.token || '',
    uuid: record.uuid || '',
    user_school_id: record.user_school_id || '',
    random_phone: record.random_phone || '',
  };
  editVisible.value = true;
}

async function submitEdit() {
  const f = editForm.value;
  if (!f.work_deadline) { message.warning('截至日期不能为空'); return; }
  editLoading.value = true;
  try {
    const formData = {
      ...f,
      work_days: Array.isArray(f.work_days) ? f.work_days.join(',') : f.work_days,
    };
    const res = await tuzhiEditOrderApi(formData);
    message.success(res?.message || '修改成功');
    editVisible.value = false;
    loadOrders();
  } catch (e: any) {
    message.error(e?.message || '修改失败');
  } finally {
    editLoading.value = false;
  }
}

// ---------- 删除 ----------

async function handleDelete(id: number) {
  try {
    await tuzhiDeleteOrderApi(id);
    message.success('删除成功');
    loadOrders();
  } catch (e: any) {
    message.error(e?.message || '删除失败');
  }
}

// ---------- 立即打卡 ----------

async function handleCheckIn(id: number) {
  try {
    await tuzhiCheckInWorkApi(id);
    message.success('上班打卡成功');
  } catch (e: any) {
    message.error(e?.message || '操作失败');
  }
}

async function handleCheckOut(id: number) {
  try {
    await tuzhiCheckOutWorkApi(id);
    message.success('下班打卡成功');
  } catch (e: any) {
    message.error(e?.message || '操作失败');
  }
}

// ---------- 同步 ----------

const syncing = ref(false);
async function handleSync() {
  syncing.value = true;
  try {
    const res = await tuzhiSyncOrdersApi();
    message.success((res as any)?.msg || '同步完成');
    loadOrders();
  } catch (e: any) {
    message.error(e?.message || '同步失败');
  } finally {
    syncing.value = false;
  }
}

// ---------- 查看签到信息 ----------

async function handleSignInfo(record: any) {
  signInfoLoading.value = true;
  signInfoVisible.value = true;
  signInfoData.value = null;
  try {
    const data = await tuzhiGetSignInfoApi({
      goods_id: record.goods_id,
      username: record.username,
      password: record.password,
    });
    signInfoData.value = data;
  } catch (e: any) {
    message.error(e?.message || '获取失败');
    signInfoVisible.value = false;
  } finally {
    signInfoLoading.value = false;
  }
}

// ---------- 工具 ----------

function formatTime(ts: any) {
  if (!ts) return '-';
  const n = Number(ts);
  if (n > 1e9) return dayjs.unix(n).format('YYYY-MM-DD HH:mm');
  return String(ts);
}

function remainingDays(deadline: string) {
  if (!deadline) return 0;
  const dl = dayjs(deadline);
  return dl.diff(dayjs(), 'day');
}

function handleDeadlinePick(d: any, ds: string) {
  addForm.value.work_deadline = ds;
}

function handleEditDeadlinePick(d: any, ds: string) {
  editForm.value.work_deadline = ds;
}

onMounted(() => {
  loadOrders();
});
</script>

<template>
  <Page title="凸知打卡" description="凸知实习打卡管理">
    <Card :bordered="false">
      <!-- 工具栏 -->
      <Row :gutter="16" style="margin-bottom: 16px" align="middle">
        <Col :flex="1">
          <Space>
            <Input v-model:value="keyword" placeholder="搜索账号/姓名" style="width:200px" allowClear @pressEnter="onSearch">
              <template #prefix><SearchOutlined /></template>
            </Input>
            <Button type="primary" @click="onSearch">搜索</Button>
          </Space>
        </Col>
        <Col>
          <Space>
            <Button type="primary" @click="openAdd"><PlusOutlined />下单</Button>
            <Button @click="handleSync" :loading="syncing"><SyncOutlined />同步订单</Button>
          </Space>
        </Col>
      </Row>

      <!-- 统计 -->
      <Row :gutter="16" style="margin-bottom: 16px">
        <Col :span="6"><Statistic title="总订单" :value="total" /></Col>
        <Col :span="6"><Statistic title="正常" :value="orders.filter(o => Number(o.is_status) === 1).length" :value-style="{ color: '#52c41a' }" /></Col>
        <Col :span="6"><Statistic title="异常" :value="orders.filter(o => Number(o.is_status) === 0).length" :value-style="{ color: '#f5222d' }" /></Col>
        <Col :span="6"><Statistic title="已完成" :value="orders.filter(o => Number(o.status) === 3).length" :value-style="{ color: '#1890ff' }" /></Col>
      </Row>

      <!-- 表格 -->
      <Table :dataSource="orders" :columns="columns" :loading="loading" :pagination="false"
             rowKey="id" size="small" :scroll="{ x: 1200 }">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'days'">
            {{ record.work_days_ok_num || 0 }}/{{ record.work_days_num || 0 }}
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="statusMap[Number(record.status)]?.color || 'default'">
              {{ statusMap[Number(record.status)]?.text || '未知' }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'is_status'">
            <Tag :color="isStatusMap[Number(record.is_status)]?.color || 'default'">
              {{ isStatusMap[Number(record.is_status)]?.text || '未知' }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <Space size="small" wrap>
              <Button size="small" type="link" @click="openEdit(record)"><EditOutlined />编辑</Button>
              <Button size="small" type="link" @click="handleSignInfo(record)">签到信息</Button>
              <Popconfirm title="确认上班打卡?" @confirm="handleCheckIn(record.id)">
                <Button size="small" type="link" style="color:#52c41a"><PlayCircleOutlined />上班</Button>
              </Popconfirm>
              <Popconfirm title="确认下班打卡?" @confirm="handleCheckOut(record.id)">
                <Button size="small" type="link" style="color:#1890ff"><ClockCircleOutlined />下班</Button>
              </Popconfirm>
              <Popconfirm title="确认删除该订单?" @confirm="handleDelete(record.id)">
                <Button size="small" type="link" danger><DeleteOutlined />删除</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>

      <div style="text-align: right; margin-top: 16px">
        <Pagination v-model:current="page" v-model:pageSize="pageSize" :total="total"
                    showSizeChanger showQuickJumper @change="loadOrders" @showSizeChange="loadOrders" />
      </div>
    </Card>

    <!-- 下单弹窗 -->
    <Modal v-model:open="addVisible" title="凸知打卡下单" width="640px" :confirmLoading="addLoading"
           @ok="submitAdd" okText="提交下单" cancelText="取消">
      <Form layout="vertical" :model="addForm">
        <Row :gutter="16">
          <Col :span="24">
            <FormItem label="选择商品" required>
              <Select v-model:value="addForm.goods_id" placeholder="请选择商品" :loading="goodsLoading" showSearch optionFilterProp="label">
                <SelectOption v-for="g in goodsList" :key="g.id" :value="g.id" :label="g.display_name || g.name">
                  {{ g.display_name || g.name }}
                </SelectOption>
              </Select>
            </FormItem>
          </Col>
        </Row>
        <Row :gutter="16">
          <Col :span="12"><FormItem label="账号" required><Input v-model:value="addForm.username" /></FormItem></Col>
          <Col :span="12"><FormItem label="密码" required><Input v-model:value="addForm.password" /></FormItem></Col>
        </Row>
        <Row :gutter="16">
          <Col :span="12"><FormItem label="姓名"><Input v-model:value="addForm.nickname" /></FormItem></Col>
          <Col :span="12"><FormItem label="学校"><Input v-model:value="addForm.school" /></FormItem></Col>
        </Row>
        <Row :gutter="16">
          <Col :span="12"><FormItem label="岗位名称"><Input v-model:value="addForm.postname" /></FormItem></Col>
          <Col :span="12"><FormItem label="截至日期" required>
            <DatePicker style="width:100%" :value="addForm.work_deadline ? dayjs(addForm.work_deadline) : null"
                        @change="handleDeadlinePick" valueFormat="YYYY-MM-DD" />
          </FormItem></Col>
        </Row>
        <Row :gutter="16">
          <Col :span="24"><FormItem label="地址"><Input v-model:value="addForm.address" /></FormItem></Col>
        </Row>
        <Row :gutter="16">
          <Col :span="12"><FormItem label="纬度"><Input v-model:value="addForm.address_lat" /></FormItem></Col>
          <Col :span="12"><FormItem label="经度"><Input v-model:value="addForm.address_lng" /></FormItem></Col>
        </Row>
        <Row :gutter="16">
          <Col :span="12"><FormItem label="上班打卡时间"><Input v-model:value="addForm.work_time" placeholder="如 08:30" /></FormItem></Col>
          <Col :span="12"><FormItem label="下班打卡时间"><Input v-model:value="addForm.off_time" placeholder="如 17:30" /></FormItem></Col>
        </Row>
        <FormItem label="打卡周期">
          <CheckboxGroup v-model:value="addForm.work_days" :options="weekOptions" />
        </FormItem>
        <Row :gutter="16">
          <Col :span="8"><FormItem label="跳过节假日"><Switch v-model:checked="addForm.holiday_status" :checkedValue="1" :unCheckedValue="0" /></FormItem></Col>
          <Col :span="8"><FormItem label="开启下班打卡"><Switch v-model:checked="addForm.is_off_time" :checkedValue="1" :unCheckedValue="0" /></FormItem></Col>
        </Row>
        <Row :gutter="16">
          <Col :span="8"><FormItem label="日报"><Switch v-model:checked="addForm.daily_report" :checkedValue="1" :unCheckedValue="0" /></FormItem></Col>
          <Col :span="8"><FormItem label="周报"><Switch v-model:checked="addForm.weekly_report" :checkedValue="1" :unCheckedValue="0" /></FormItem></Col>
          <Col :span="8"><FormItem label="月报"><Switch v-model:checked="addForm.monthly_report" :checkedValue="1" :unCheckedValue="0" /></FormItem></Col>
        </Row>
        <FormItem label="息知推送地址"><Input v-model:value="addForm.xz_push_url" placeholder="可选" /></FormItem>
      </Form>
    </Modal>

    <!-- 编辑弹窗 -->
    <Modal v-model:open="editVisible" title="编辑订单" width="640px" :confirmLoading="editLoading"
           @ok="submitEdit" okText="保存" cancelText="取消">
      <Form layout="vertical" :model="editForm">
        <Row :gutter="16">
          <Col :span="12"><FormItem label="账号"><Input v-model:value="editForm.username" /></FormItem></Col>
          <Col :span="12"><FormItem label="密码"><Input v-model:value="editForm.password" /></FormItem></Col>
        </Row>
        <Row :gutter="16">
          <Col :span="12"><FormItem label="姓名"><Input v-model:value="editForm.nickname" /></FormItem></Col>
          <Col :span="12"><FormItem label="学校"><Input v-model:value="editForm.school" /></FormItem></Col>
        </Row>
        <Row :gutter="16">
          <Col :span="12"><FormItem label="岗位名称"><Input v-model:value="editForm.postname" /></FormItem></Col>
          <Col :span="12"><FormItem label="截至日期" required>
            <DatePicker style="width:100%" :value="editForm.work_deadline ? dayjs(editForm.work_deadline) : null"
                        @change="handleEditDeadlinePick" valueFormat="YYYY-MM-DD" />
          </FormItem></Col>
        </Row>
        <Row :gutter="16">
          <Col :span="24"><FormItem label="地址"><Input v-model:value="editForm.address" /></FormItem></Col>
        </Row>
        <Row :gutter="16">
          <Col :span="12"><FormItem label="纬度"><Input v-model:value="editForm.address_lat" /></FormItem></Col>
          <Col :span="12"><FormItem label="经度"><Input v-model:value="editForm.address_lng" /></FormItem></Col>
        </Row>
        <Row :gutter="16">
          <Col :span="12"><FormItem label="上班打卡时间"><Input v-model:value="editForm.work_time" /></FormItem></Col>
          <Col :span="12"><FormItem label="下班打卡时间"><Input v-model:value="editForm.off_time" /></FormItem></Col>
        </Row>
        <FormItem label="打卡周期">
          <CheckboxGroup v-model:value="editForm.work_days" :options="weekOptions" />
        </FormItem>
        <Row :gutter="16">
          <Col :span="8"><FormItem label="跳过节假日"><Switch v-model:checked="editForm.holiday_status" :checkedValue="1" :unCheckedValue="0" /></FormItem></Col>
          <Col :span="8"><FormItem label="开启下班打卡"><Switch v-model:checked="editForm.is_off_time" :checkedValue="1" :unCheckedValue="0" /></FormItem></Col>
        </Row>
        <Row :gutter="16">
          <Col :span="8"><FormItem label="日报"><Switch v-model:checked="editForm.daily_report" :checkedValue="1" :unCheckedValue="0" /></FormItem></Col>
          <Col :span="8"><FormItem label="周报"><Switch v-model:checked="editForm.weekly_report" :checkedValue="1" :unCheckedValue="0" /></FormItem></Col>
          <Col :span="8"><FormItem label="月报"><Switch v-model:checked="editForm.monthly_report" :checkedValue="1" :unCheckedValue="0" /></FormItem></Col>
        </Row>
        <FormItem label="息知推送地址"><Input v-model:value="editForm.xz_push_url" placeholder="可选" /></FormItem>
      </Form>
    </Modal>

    <!-- 签到信息弹窗 -->
    <Modal v-model:open="signInfoVisible" title="签到信息" :footer="null" width="600px">
      <Spin :spinning="signInfoLoading">
        <div v-if="signInfoData" style="max-height:400px;overflow:auto">
          <pre style="white-space:pre-wrap;word-break:break-all">{{ JSON.stringify(signInfoData, null, 2) }}</pre>
        </div>
        <div v-else-if="!signInfoLoading" style="text-align:center;padding:20px;color:#999">暂无数据</div>
      </Spin>
    </Modal>
  </Page>
</template>

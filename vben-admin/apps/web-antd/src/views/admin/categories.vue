<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, Space, Tag, Modal,
  Pagination, message, Select, SelectOption, Switch,
} from 'ant-design-vue';
import {
  PlusOutlined, SearchOutlined, ReloadOutlined, ThunderboltOutlined,
} from '@ant-design/icons-vue';
import {
  getCategoryListPagedApi, saveCategoryApi, deleteCategoryApi, categoryQuickModifyApi,
  getSupplierListApi, type CategoryItem, type SupplierItem,
} from '#/api/admin';

const loading = ref(false);
const list = ref<CategoryItem[]>([]);
const pagination = reactive({ page: 1, limit: 20, total: 0 });
const search = reactive({ keyword: '', status: '' });
const suppliers = ref<SupplierItem[]>([]);

async function loadSuppliers() {
  try {
    const raw = await getSupplierListApi();
    suppliers.value = (raw) || [];
  } catch {}
}

const supplierOptions = computed(() => [
  { value: 0, label: '自动识别（使用订单供应商）' },
  ...suppliers.value.map((s) => ({ value: s.hid, label: `${s.name} (HID:${s.hid} PT:${s.pt})` })),
]);

const statusMap: Record<string, string> = {
  '0': '未启用', '1': '已启用', '2': '启用分类2', '3': '启用分类3', '4': '启用分类4', '5': '启用分类5',
};
const statusColor = (s: string) => {
  if (s === '1') return 'success';
  if (s === '0') return 'error';
  return 'warning';
};

// ===== 数据加载 =====
async function loadData(page = 1) {
  loading.value = true;
  pagination.page = page;
  try {
    const raw = await getCategoryListPagedApi({
      page, limit: pagination.limit,
      keyword: search.keyword || undefined,
      status: search.status || undefined,
    });
    const res = raw;
    list.value = res.list || [];
    pagination.total = res.pagination?.total || 0;
  } catch (e) { console.error(e); }
  finally { loading.value = false; }
}

// ===== 添加/编辑弹窗 =====
const modalVisible = ref(false);
const isEdit = ref(false);
const form = reactive({
  id: 0, sort: '', name: '', status: '1',
  recommend: 0, log: 0, ticket: 0, changepass: 1, allowpause: 0, supplier_report: 0, supplier_report_hid: 0,
});

function openAdd() {
  isEdit.value = false;
  Object.assign(form, { id: 0, sort: '', name: '', status: '1', recommend: 0, log: 0, ticket: 0, changepass: 1, allowpause: 0, supplier_report: 0, supplier_report_hid: 0 });
  modalVisible.value = true;
}

function openEdit(item: CategoryItem) {
  isEdit.value = true;
  Object.assign(form, {
    id: item.id,
    sort: String(item.sort ?? ''),
    name: item.name ?? '',
    status: String(item.status ?? '1'),
    recommend: item.recommend ?? 0,
    log: item.log ?? 0,
    ticket: item.ticket ?? 0,
    changepass: item.changepass ?? 1,
    allowpause: item.allowpause ?? 0,
    supplier_report: item.supplier_report ?? 0,
    supplier_report_hid: item.supplier_report_hid ?? 0,
  });
  modalVisible.value = true;
}

async function submitForm() {
  if (!form.name.trim()) { message.error('请输入分类名称'); return; }
  if (!form.sort.trim()) { message.error('请输入排序'); return; }
  try {
    await saveCategoryApi({
      id: form.id || undefined,
      name: form.name,
      sort: Number(form.sort) || 0,
      status: form.status,
      recommend: form.recommend,
      log: form.log,
      ticket: form.ticket,
      changepass: form.changepass,
      allowpause: form.allowpause,
      supplier_report: form.supplier_report,
      supplier_report_hid: form.supplier_report_hid,
    } as any);
    message.success(isEdit.value ? '修改成功' : '添加成功');
    modalVisible.value = false;
    loadData(isEdit.value ? pagination.page : 1);
  } catch (e: any) { message.error(e?.message || '保存失败'); }
}

// ===== 删除 =====
function confirmDelete(id: number) {
  Modal.confirm({
    title: '提示', content: '确定删除这个分类吗？',
    async onOk() {
      try {
        await deleteCategoryApi(id);
        message.success('删除成功');
        loadData(pagination.page);
      } catch (e: any) { message.error(e?.message || '删除失败'); }
    },
  });
}

// ===== 快速修改弹窗 =====
const quickVisible = ref(false);
const quickForm = reactive({ keyword: '', cid: '' });

function openQuick() {
  Object.assign(quickForm, { keyword: '', cid: '' });
  quickVisible.value = true;
}

async function submitQuick() {
  if (!quickForm.keyword.trim()) { message.error('请填写关键词'); return; }
  if (!quickForm.cid.trim()) { message.error('请填写分类ID'); return; }
  try {
    const raw = await categoryQuickModifyApi(quickForm.keyword, Number(quickForm.cid));
    const res = raw;
    message.success(res.msg || '修改成功');
    quickVisible.value = false;
    loadData(pagination.page);
  } catch (e: any) { message.error(e?.message || '修改失败'); }
}

// ===== 切换状态 =====
async function toggleStatus(item: CategoryItem) {
  const newStatus = item.status === '1' ? '0' : '1';
  Modal.confirm({
    title: '确认操作',
    content: `确定${newStatus === '1' ? '启用' : '禁用'}该分类？`,
    async onOk() {
      try {
        await saveCategoryApi({ id: item.id, name: item.name, sort: item.sort, status: newStatus, recommend: item.recommend, log: item.log, ticket: item.ticket, changepass: item.changepass, allowpause: item.allowpause, supplier_report: item.supplier_report, supplier_report_hid: item.supplier_report_hid } as any);
        message.success('操作成功');
        loadData(pagination.page);
      } catch (e: any) { message.error(e?.message || '操作失败'); }
    },
  });
}

const columns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 70, align: 'center' as const },
  { title: '排序', dataIndex: 'sort', key: 'sort', width: 80, align: 'center' as const },
  { title: '分类名称', dataIndex: 'name', key: 'name', width: 200 },
  { title: '状态', key: 'status', width: 100, align: 'center' as const },
  { title: '开关', key: 'switches', width: 280 },
  { title: '添加时间', dataIndex: 'time', key: 'time', width: 160, align: 'center' as const },
  { title: '操作', key: 'action', width: 140, align: 'center' as const },
];

onMounted(() => { loadData(1); loadSuppliers(); });
</script>

<template>
  <Page title="分类管理" content-class="p-4">

    <!-- 搜索/操作区 -->
    <Card title="分类管理" class="mb-4">
      <template #extra>
        <Space>
          <Button type="primary" size="small" @click="openAdd">
            <template #icon><PlusOutlined /></template>
            添加分类
          </Button>
          <Button size="small" @click="openQuick">
            <template #icon><ThunderboltOutlined /></template>
            快速修改
          </Button>
        </Space>
      </template>

      <div class="flex flex-wrap gap-3 items-center">
        <Input v-model:value="search.keyword" placeholder="输入关键词" class="w-full" style="max-width: 240px; min-width: 120px" @press-enter="loadData(1)" />

        <Select v-model:value="search.status" placeholder="请选择状态" style="max-width: 160px; min-width: 100px" allow-clear>
          <SelectOption value="">全部状态</SelectOption>
          <SelectOption value="1">已启用</SelectOption>
          <SelectOption value="0">未启用</SelectOption>
          <SelectOption value="2">启用分类2</SelectOption>
          <SelectOption value="3">启用分类3</SelectOption>
          <SelectOption value="4">启用分类4</SelectOption>
          <SelectOption value="5">启用分类5</SelectOption>
        </Select>

        <Select v-model:value="pagination.limit" style="max-width: 100px; min-width: 80px" @change="loadData(1)">
          <SelectOption :value="20">20/页</SelectOption>
          <SelectOption :value="50">50/页</SelectOption>
          <SelectOption :value="100">100/页</SelectOption>
          <SelectOption :value="200">200/页</SelectOption>
        </Select>

        <Button type="primary" @click="loadData(1)">
          <template #icon><SearchOutlined /></template>
          搜索
        </Button>
        <Button @click="search.keyword = ''; search.status = ''; loadData(1)">
          <template #icon><ReloadOutlined /></template>
          刷新
        </Button>
      </div>
    </Card>

    <!-- 分类列表 -->
    <Card title="分类列表">
      <Table
        :columns="columns" :data-source="list" :loading="loading"
        :pagination="false" row-key="id" size="small" bordered
        :scroll="{ x: 1050 }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <Tag :color="statusColor(String(record.status))" style="cursor: pointer" @click="toggleStatus(record)">
              {{ statusMap[String(record.status)] || `状态${record.status}` }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'switches'">
            <Space :size="4" wrap>
              <Tag v-if="record.recommend" color="purple">推荐</Tag>
              <Tag v-if="record.log" color="blue">日志</Tag>
              <Tag v-if="record.ticket" color="orange">工单</Tag>
              <Tag v-if="record.changepass" color="cyan">改密</Tag>
              <Tag v-if="record.allowpause" color="green">暂停</Tag>
              <Tag v-if="record.supplier_report" color="red">上游反馈{{ record.supplier_report_hid ? `(HID:${record.supplier_report_hid})` : '' }}</Tag>
              <span v-if="!record.recommend && !record.log && !record.ticket && !record.changepass && !record.allowpause && !record.supplier_report" class="text-gray-400 text-xs">-</span>
            </Space>
          </template>
          <template v-else-if="column.key === 'action'">
            <Space size="small">
              <Button type="primary" size="small" @click="openEdit(record)">编辑</Button>
              <Button danger size="small" @click="confirmDelete(record.id)">删除</Button>
            </Space>
          </template>
        </template>
      </Table>

      <div class="flex justify-center mt-4" v-if="pagination.total > 0">
        <Pagination
          v-model:current="pagination.page"
          :total="pagination.total"
          :page-size="pagination.limit"
          :show-size-changer="false"
          :show-total="(total: number) => `共 ${total} 条`"
          @change="(p: number) => loadData(p)"
        />
      </div>
    </Card>

    <!-- 添加/编辑弹窗 -->
    <Modal v-model:open="modalVisible" :title="isEdit ? '编辑分类' : '添加分类'" :footer="null" :width="600" style="max-width: 95vw">
      <div class="py-2">
        <!-- 基本信息 -->
        <div class="mb-4">
          <div class="text-sm font-semibold text-gray-600 mb-3">基本信息</div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm text-gray-500 mb-1">* 分类名称</label>
              <Input v-model:value="form.name" placeholder="用于识别该分类的名称" />
            </div>
            <div>
              <label class="block text-sm text-gray-500 mb-1">排序</label>
              <Input v-model:value="form.sort" placeholder="数字越小越靠前" />
            </div>
          </div>
        </div>

        <!-- 开关设置 -->
        <div class="mb-4">
          <div class="text-sm font-semibold text-gray-600 mb-3">开关设置</div>
          <div class="grid grid-cols-2 gap-y-4 gap-x-8">
            <div class="flex items-center gap-3">
              <Switch :checked="form.status !== '0'" @change="(v: boolean) => form.status = v ? '1' : '0'" />
              <div>
                <div class="text-sm font-medium">启用状态</div>
                <div class="text-xs text-gray-400">启用后该分类可正常使用</div>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <Switch :checked="form.recommend === 1" @change="(v: boolean) => form.recommend = v ? 1 : 0" />
              <div>
                <div class="text-sm font-medium">推荐分类</div>
                <div class="text-xs text-gray-400">设为推荐分类</div>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <Switch :checked="form.log === 1" @change="(v: boolean) => form.log = v ? 1 : 0" />
              <div>
                <div class="text-sm font-medium">开启日志</div>
                <div class="text-xs text-gray-400">记录操作日志</div>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <Switch :checked="form.ticket === 1" @change="(v: boolean) => form.ticket = v ? 1 : 0" />
              <div>
                <div class="text-sm font-medium">工单开关</div>
                <div class="text-xs text-gray-400">开启工单功能</div>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <Switch :checked="form.changepass === 1" @change="(v: boolean) => form.changepass = v ? 1 : 0" />
              <div>
                <div class="text-sm font-medium">修改密码</div>
                <div class="text-xs text-gray-400">允许修改密码</div>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <Switch :checked="form.allowpause === 1" @change="(v: boolean) => form.allowpause = v ? 1 : 0" />
              <div>
                <div class="text-sm font-medium">允许暂停</div>
                <div class="text-xs text-gray-400">开启订单暂停/开始功能</div>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <Switch :checked="form.supplier_report === 1" @change="(v: boolean) => { form.supplier_report = v ? 1 : 0; if (!v) form.supplier_report_hid = 0; }" />
              <div>
                <div class="text-sm font-medium">上游反馈</div>
                <div class="text-xs text-gray-400">允许向货源方提交反馈</div>
              </div>
            </div>
          </div>
          <!-- 上游反馈供应商选择 -->
          <div v-if="form.supplier_report === 1" class="mt-4">
            <label class="block text-sm text-gray-500 mb-1">反馈供应商（0=自动识别订单供应商）</label>
            <Select v-model:value="form.supplier_report_hid" style="width: 100%" :options="supplierOptions" show-search option-filter-prop="label" />
          </div>
        </div>

        <!-- 底部按钮 -->
        <div class="flex justify-center gap-3 pt-2">
          <Button type="primary" @click="submitForm">保存分类</Button>
          <Button @click="modalVisible = false">取消</Button>
        </div>
      </div>
    </Modal>

    <!-- 快速修改弹窗 -->
    <Modal v-model:open="quickVisible" title="快速修改分类" @ok="submitQuick" ok-text="确定" cancel-text="取消" :width="480" style="max-width: 95vw">
      <div class="space-y-4 py-2">
        <div>
          <label class="block text-sm font-medium mb-1">关键词（示例：强盛）</label>
          <Input v-model:value="quickForm.keyword" placeholder="比如：强盛" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">分类ID（示例：12）</label>
          <Input v-model:value="quickForm.cid" placeholder="比如：12" />
        </div>
        <div class="text-xs text-gray-400">
          提示：将所有商品名称中包含该关键词的商品，批量归入指定分类。
        </div>
      </div>
    </Modal>

  </Page>
</template>

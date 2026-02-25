<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal, Switch,
  Pagination, Popconfirm, message, Select, SelectOption, Tabs, TabPane, Dropdown, Menu, MenuItem,
} from 'ant-design-vue';
import {
  PlusOutlined, SearchOutlined, ReloadOutlined, EditOutlined, DeleteOutlined, DownOutlined,
} from '@ant-design/icons-vue';
import {
  getCategoryListApi, saveCategoryApi, deleteCategoryApi,
  getClassManageListApi, saveClassApi, toggleClassStatusApi, batchDeleteClassApi,
  batchCategoryClassApi, batchPriceClassApi,
  getSupplierListApi,
  type CategoryItem, type ClassItem, type SupplierItem,
} from '#/api/admin';

// ===== 分类 =====
const categories = ref<CategoryItem[]>([]);
const catLoading = ref(false);
const catEditVisible = ref(false);
const catForm = reactive({ id: 0, name: '', status: '1', sort: 0, recommend: 0, log: 0, ticket: 0, changepass: 1, allowpause: 0, supplier_report: 0, supplier_report_hid: 0 });

async function loadCategories() {
  catLoading.value = true;
  try {
    const res = await getCategoryListApi();
    categories.value = res;
    if (!Array.isArray(categories.value)) categories.value = [];
  } catch (e) { console.error(e); }
  finally { catLoading.value = false; }
}

function openCatEdit(cat?: CategoryItem) {
  if (cat) {
    Object.assign(catForm, cat);
  } else {
    Object.assign(catForm, { id: 0, name: '', status: '1', sort: 0, recommend: 0, log: 0, ticket: 0, changepass: 1, allowpause: 0, supplier_report: 0, supplier_report_hid: 0 });
  }
  catEditVisible.value = true;
}

async function handleCatSave() {
  if (!catForm.name.trim()) { message.warning('请填写分类名称'); return; }
  try {
    await saveCategoryApi({ ...catForm, status: String(catForm.status) });
    message.success('保存成功');
    catEditVisible.value = false;
    loadCategories();
  } catch (e: any) { message.error(e?.message || '保存失败'); }
}

async function handleCatDelete(id: number) {
  try {
    await deleteCategoryApi(id);
    message.success('删除成功');
    loadCategories();
  } catch (e: any) { message.error(e?.message || '删除失败'); }
}

// ===== 课程 =====
const classLoading = ref(false);
const classList = ref<ClassItem[]>([]);
const classPagination = reactive({ page: 1, limit: 20, total: 0 });
const classSearch = reactive({ cateId: undefined as number | undefined, keywords: '' });
const classEditVisible = ref(false);
const classForm = reactive({ cid: 0, name: '', price: '0', content: '', cateId: '0', status: 1, hid: '0', sort: 10, noun: '', yunsuan: '*' });

// 货源列表（供选择）
const suppliers = ref<SupplierItem[]>([]);

async function loadClasses(page = 1) {
  classLoading.value = true;
  classPagination.page = page;
  try {
    const raw = await getClassManageListApi({
      page: classPagination.page,
      limit: classPagination.limit,
      cateId: classSearch.cateId,
      keywords: classSearch.keywords,
    });
    const res = raw;
    classList.value = res.list || [];
    classPagination.total = res.pagination?.total || 0;
  } catch (e) { console.error(e); }
  finally { classLoading.value = false; }
}

async function loadSuppliers() {
  try {
    const res = await getSupplierListApi();
    suppliers.value = res;
    if (!Array.isArray(suppliers.value)) suppliers.value = [];
  } catch (e) {}
}

function openClassEdit(cls?: ClassItem) {
  if (cls) {
    Object.assign(classForm, cls);
  } else {
    Object.assign(classForm, { cid: 0, name: '', price: '0', content: '', cateId: '0', status: 1, hid: '0', sort: 10, noun: '', yunsuan: '*' });
  }
  classEditVisible.value = true;
}

async function handleClassSave() {
  if (!classForm.name.trim()) { message.warning('请填写课程名称'); return; }
  try {
    await saveClassApi({ ...classForm, price: String(classForm.price), hid: String(classForm.hid), cateId: String(classForm.cateId) });
    message.success('保存成功');
    classEditVisible.value = false;
    loadClasses(classPagination.page);
  } catch (e: any) { message.error(e?.message || '保存失败'); }
}

async function handleClassToggle(cid: number, status: number) {
  try {
    await toggleClassStatusApi(cid, status ? 0 : 1);
    loadClasses(classPagination.page);
  } catch (e: any) { message.error(e?.message || '操作失败'); }
}

function getCatName(cateId: string | number) {
  return categories.value.find(c => c.id === Number(cateId))?.name || '-';
}

// ===== 全选删除 =====
const selectedClassKeys = ref<number[]>([]);
const classRowSelection = computed(() => ({
  selectedRowKeys: selectedClassKeys.value,
  onChange: (keys: number[]) => { selectedClassKeys.value = keys; },
}));

async function handleBatchDelete() {
  if (selectedClassKeys.value.length === 0) { message.warning('请先勾选要删除的课程'); return; }
  Modal.confirm({
    title: '批量删除',
    content: `确定删除选中的 ${selectedClassKeys.value.length} 个课程吗？此操作不可恢复。`,
    async onOk() {
      try {
        const raw = await batchDeleteClassApi(selectedClassKeys.value);
        const res = raw;
        message.success(res.msg || '删除成功');
        selectedClassKeys.value = [];
        loadClasses(classPagination.page);
      } catch (e: any) { message.error(e?.message || '删除失败'); }
    },
  });
}

// ===== 批量修改分类 =====
const batchCatVisible = ref(false);
const batchCatId = ref<string>('');
async function handleBatchCategory() {
  if (!selectedClassKeys.value.length) { message.warning('请先勾选课程'); return; }
  batchCatId.value = '';
  batchCatVisible.value = true;
}
async function doBatchCategory() {
  if (!batchCatId.value) { message.warning('请选择分类'); return; }
  try {
    const res = await batchCategoryClassApi(selectedClassKeys.value, String(batchCatId.value));
    message.success((res as any).msg || '修改成功');
    batchCatVisible.value = false;
    selectedClassKeys.value = [];
    loadClasses(classPagination.page);
  } catch (e: any) { message.error(e?.message || '修改失败'); }
}

// ===== 批量修改价格 =====
const batchPriceVisible = ref(false);
const batchRate = ref<number>(1);
const batchYunsuan = ref('*');
async function handleBatchPrice() {
  if (!selectedClassKeys.value.length) { message.warning('请先勾选课程'); return; }
  batchRate.value = 1;
  batchYunsuan.value = '*';
  batchPriceVisible.value = true;
}
async function doBatchPrice() {
  if (!batchRate.value && batchRate.value !== 0) { message.warning('请输入倍率/加价'); return; }
  try {
    const res = await batchPriceClassApi(selectedClassKeys.value, batchRate.value, batchYunsuan.value);
    message.success((res as any).msg || '修改成功');
    batchPriceVisible.value = false;
    selectedClassKeys.value = [];
    loadClasses(classPagination.page);
  } catch (e: any) { message.error(e?.message || '修改失败'); }
}

const classColumns = [
  { title: 'CID', dataIndex: 'cid', key: 'cid', width: 70 },
  { title: '课程名称', dataIndex: 'name', key: 'name', ellipsis: true },
  { title: '分类', key: 'cateId', width: 100 },
  { title: '价格', key: 'price', width: 80 },
  { title: '货源', dataIndex: 'hid', key: 'hid', width: 70 },
  { title: '接口编号', dataIndex: 'noun', key: 'noun', width: 100, ellipsis: true },
  { title: '排序', dataIndex: 'sort', key: 'sort', width: 60 },
  { title: '状态', key: 'status', width: 80 },
  { title: '操作', key: 'action', width: 150 },
];

onMounted(() => {
  loadCategories();
  loadClasses(1);
  loadSuppliers();
});
</script>

<template>
  <Page title="课程管理" content-class="p-4">
    <Tabs>
      <TabPane key="class" tab="课程列表">
        <Card>
          <div class="flex flex-wrap justify-between items-center gap-3 mb-4">
            <Space>
              <Select
                v-model:value="classSearch.cateId"
                placeholder="筛选分类"
                allow-clear
                show-search
                option-filter-prop="label"
                style="max-width: 140px; min-width: 100px"
                @change="loadClasses(1)"
              >
                <SelectOption v-for="c in categories" :key="c.id" :value="c.id" :label="c.name">{{ c.name }}</SelectOption>
              </Select>
              <Input v-model:value="classSearch.keywords" placeholder="搜索课程名/CID" allow-clear style="max-width: 180px; min-width: 100px" @pressEnter="loadClasses(1)" />
              <Button type="primary" @click="loadClasses(1)"><template #icon><SearchOutlined /></template></Button>
              <Button @click="loadClasses(classPagination.page)"><template #icon><ReloadOutlined /></template></Button>
            </Space>
            <Space>
              <Dropdown :disabled="selectedClassKeys.length === 0">
                <Button :disabled="selectedClassKeys.length === 0">
                  批量操作<span v-if="selectedClassKeys.length">({{ selectedClassKeys.length }})</span> <DownOutlined />
                </Button>
                <template #overlay>
                  <Menu>
                    <MenuItem key="cat" @click="handleBatchCategory">批量改分类</MenuItem>
                    <MenuItem key="price" @click="handleBatchPrice">批量改价格</MenuItem>
                    <MenuItem key="del" danger @click="handleBatchDelete">批量删除</MenuItem>
                  </Menu>
                </template>
              </Dropdown>
              <Button type="primary" @click="openClassEdit()">
                <template #icon><PlusOutlined /></template>
                添加课程
              </Button>
            </Space>
          </div>

          <Table :columns="classColumns" :data-source="classList" :loading="classLoading" :pagination="false" :row-selection="classRowSelection" row-key="cid" size="small" bordered :scroll="{ x: 1000 }">
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'cateId'">{{ getCatName(record.cateId) }}</template>
              <template v-else-if="column.key === 'price'">¥{{ Number(record.price).toFixed(2) }}</template>
              <template v-else-if="column.key === 'status'">
                <Switch :checked="!!record.status" size="small" @change="handleClassToggle(record.cid, record.status)" />
              </template>
              <template v-else-if="column.key === 'action'">
                <Button type="link" size="small" @click="openClassEdit(record)">
                  <template #icon><EditOutlined /></template>编辑
                </Button>
              </template>
            </template>
          </Table>

          <div class="flex justify-center mt-4">
            <Pagination
              v-model:current="classPagination.page"
              v-model:pageSize="classPagination.limit"
              :total="classPagination.total"
              :page-size-options="['20', '50', '100', '200']"
              show-size-changer
              :show-total="(total: number) => `共 ${total} 条`"
              @change="(p: number) => loadClasses(p)"
              @showSizeChange="(_: number, size: number) => { classPagination.limit = size; loadClasses(1); }"
            />
          </div>
        </Card>
      </TabPane>

      <TabPane key="category" tab="分类管理">
        <Card>
          <div class="flex justify-between items-center mb-4">
            <span class="text-sm text-gray-500">共 {{ categories.length }} 个分类</span>
            <Button type="primary" @click="openCatEdit()">
              <template #icon><PlusOutlined /></template>
              添加分类
            </Button>
          </div>
          <Table :data-source="categories" :loading="catLoading" :pagination="false" row-key="id" size="small" bordered :scroll="{ x: 500 }">
            <Table.Column title="ID" data-index="id" :width="60" />
            <Table.Column title="名称" data-index="name" />
            <Table.Column title="排序" data-index="sort" :width="70" />
            <Table.Column title="状态" :width="80">
              <template #default="{ record }">
                <Tag :color="record.status === 1 ? 'green' : 'default'">{{ record.status === 1 ? '启用' : '禁用' }}</Tag>
              </template>
            </Table.Column>
            <Table.Column title="操作" :width="150">
              <template #default="{ record }">
                <Space size="small">
                  <Button type="link" size="small" @click="openCatEdit(record)">编辑</Button>
                  <Popconfirm title="确定删除此分类？" @confirm="handleCatDelete(record.id)">
                    <Button type="link" size="small" danger>删除</Button>
                  </Popconfirm>
                </Space>
              </template>
            </Table.Column>
          </Table>
        </Card>
      </TabPane>
    </Tabs>

    <!-- 分类编辑弹窗 -->
    <Modal v-model:open="catEditVisible" :title="catForm.id ? '编辑分类' : '添加分类'" @ok="handleCatSave" ok-text="保存">
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium mb-1">名称</label>
          <Input v-model:value="catForm.name" placeholder="分类名称" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">排序</label>
          <InputNumber v-model:value="catForm.sort" :min="0" style="width: 100%" />
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">状态</label>
          <Select v-model:value="catForm.status" style="width: 100%">
            <SelectOption :value="1">启用</SelectOption>
            <SelectOption :value="0">禁用</SelectOption>
          </Select>
        </div>
      </div>
    </Modal>

    <!-- 课程编辑弹窗 -->
    <Modal v-model:open="classEditVisible" :title="classForm.cid ? '编辑课程' : '添加课程'" @ok="handleClassSave" ok-text="保存" width="600px">
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium mb-1">课程名称</label>
          <Input v-model:value="classForm.name" placeholder="课程名称" />
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">价格</label>
            <InputNumber v-model:value="classForm.price" :min="0" :step="0.1" :precision="2" style="width: 100%" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">分类</label>
            <Select
              v-model:value="classForm.cateId"
              placeholder="选择分类"
              show-search
              option-filter-prop="label"
              style="width: 100%"
            >
              <SelectOption v-for="c in categories" :key="c.id" :value="c.id" :label="c.name">{{ c.name }}</SelectOption>
            </Select>
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">货源</label>
            <Select v-model:value="classForm.hid" placeholder="选择货源" style="width: 100%">
              <SelectOption :value="0">无</SelectOption>
              <SelectOption v-for="s in suppliers" :key="s.hid" :value="s.hid">{{ s.name }} (HID:{{ s.hid }})</SelectOption>
            </Select>
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">状态</label>
            <Select v-model:value="classForm.status" style="width: 100%">
              <SelectOption :value="1">上架</SelectOption>
              <SelectOption :value="0">下架</SelectOption>
            </Select>
          </div>
        </div>
        <div class="grid grid-cols-3 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">接口编号 (noun)</label>
            <Input v-model:value="classForm.noun" placeholder="上游课程ID" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">排序</label>
            <InputNumber v-model:value="classForm.sort" :min="0" style="width: 100%" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">加价方式</label>
            <Select v-model:value="classForm.yunsuan" style="width: 100%">
              <SelectOption value="*">乘法 (*)</SelectOption>
              <SelectOption value="+">加法 (+)</SelectOption>
            </Select>
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">描述</label>
          <Input.TextArea v-model:value="classForm.content" :rows="3" placeholder="课程描述" />
        </div>
      </div>
    </Modal>

    <!-- 批量修改分类弹窗 -->
    <Modal v-model:open="batchCatVisible" title="批量修改分类" @ok="doBatchCategory" ok-text="确定">
      <p class="mb-2 text-sm text-gray-500">已选择 {{ selectedClassKeys.length }} 个课程</p>
      <Select v-model:value="batchCatId" placeholder="选择目标分类" show-search option-filter-prop="label" style="width: 100%">
        <SelectOption v-for="c in categories" :key="c.id" :value="c.id" :label="c.name">{{ c.name }}</SelectOption>
      </Select>
    </Modal>

    <!-- 批量修改价格弹窗 -->
    <Modal v-model:open="batchPriceVisible" title="批量修改价格" @ok="doBatchPrice" ok-text="确定">
      <p class="mb-2 text-sm text-gray-500">已选择 {{ selectedClassKeys.length }} 个课程</p>
      <div class="space-y-3">
        <div>
          <label class="block text-sm font-medium mb-1">计算方式</label>
          <Select v-model:value="batchYunsuan" style="width: 100%">
            <SelectOption value="*">乘法（倍率）— 价格 × 倍率</SelectOption>
            <SelectOption value="+">加法（固定）— 价格 + 金额</SelectOption>
          </Select>
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">{{ batchYunsuan === '*' ? '倍率（如 1.5 = 涨价50%）' : '加价金额（元）' }}</label>
          <InputNumber v-model:value="batchRate" :min="0" :step="batchYunsuan === '*' ? 0.1 : 1" :precision="batchYunsuan === '*' ? 4 : 2" style="width: 100%" />
        </div>
      </div>
    </Modal>
  </Page>
</template>

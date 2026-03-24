<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, InputNumber, Space, Tag, Modal, Switch,
  Pagination, Progress, message, Select, SelectOption, Checkbox,
} from 'ant-design-vue';
import {
  SearchOutlined, ReloadOutlined, CloudDownloadOutlined,
  PlusOutlined, WarningOutlined, ThunderboltOutlined, UploadOutlined,
  SwapOutlined, TagOutlined,
} from '@ant-design/icons-vue';
import {
  getSupplierListApi, getSupplierProductsApi, addClassApi,
  importSupplierApi, syncSupplierStatusApi, getCategoryListApi, saveCategoryApi,
  batchReplaceKeywordApi, batchAddPrefixApi,
  type SupplierItem, type SupplierProductItem, type CategoryItem,
} from '#/api/admin';

// 货源列表
const suppliers = ref<SupplierItem[]>([]);
const categories = ref<CategoryItem[]>([]);
const selectedHid = ref<number | undefined>(undefined);
const selectedHyName = computed(() => {
  if (!selectedHid.value) return '';
  return suppliers.value.find(s => s.hid === selectedHid.value)?.name || '';
});

// 商品列表
const loading = ref(false);
const rawList = ref<SupplierProductItem[]>([]);
const search = reactive({ keyword: '', category: '' });
const currentPage = ref(1);
const pageSize = ref(20);
const selectedKeys = ref<string[]>([]);

// 筛选后的列表
const filteredList = computed(() => {
  const kw = (search.keyword || '').trim().toLowerCase();
  const cat = (search.category || '').trim();
  return rawList.value.filter(it => {
    const okKw = kw ? (it.name.toLowerCase().includes(kw) || (it.content || '').toLowerCase().includes(kw)) : true;
    const okCat = cat ? (String(it.fenlei) === cat) : true;
    return okKw && okCat;
  });
});

const totalRows = computed(() => filteredList.value.length);
const displayList = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  return filteredList.value.slice(start, start + pageSize.value);
});

// 上游分类列表（从拉取的数据中提取）
const uniqueCategories = computed(() => {
  const m = new Map<string, string>();
  rawList.value.forEach(it => {
    const key = it.fenlei || '';
    if (!key) return;
    const label = it.category_name || it.fenlei;
    m.set(key, `${label} (${key})`);
  });
  return Array.from(m.entries()).map(([value, label]) => ({ value, label }));
});

// 选择（跨页保留 + 表头全选操作所有筛选结果而非仅当前页）
const rowSelection = computed(() => ({
  selectedRowKeys: selectedKeys.value,
  preserveSelectedRowKeys: true,
  onChange: (keys: string[]) => { selectedKeys.value = keys; },
  onSelectAll: (selected: boolean) => {
    if (selected) {
      // 全选：选中所有筛选结果，而非仅当前页
      selectedKeys.value = filteredList.value.map(it => it.cid);
    } else {
      selectedKeys.value = [];
    }
  },
}));

// 加载货源 + 分类
async function loadBase() {
  try {
    const [sRaw, cRaw] = await Promise.all([getSupplierListApi(), getCategoryListApi()]);
    suppliers.value = sRaw;
    if (!Array.isArray(suppliers.value)) suppliers.value = [];
    categories.value = cRaw;
    if (!Array.isArray(categories.value)) categories.value = [];
  } catch (e) { console.error(e); }
}

// 拉取货源商品
async function fetchProducts() {
  if (!selectedHid.value) { message.warning('请先选择货源'); return; }
  loading.value = true;
  try {
    const pRes = await getSupplierProductsApi(selectedHid.value);
    rawList.value = pRes;
    if (!Array.isArray(rawList.value)) rawList.value = [];
    selectedKeys.value = [];
    currentPage.value = 1;
    message.success(`拉取到 ${rawList.value.length} 个商品`);
  } catch (e: any) { message.error(e?.message || '拉取失败'); rawList.value = []; }
  finally { loading.value = false; }
}

// 检查失效商品
async function checkInvalid() {
  if (!selectedHid.value) { message.warning('请先选择货源'); return; }
  Modal.confirm({
    title: '检查失效商品',
    content: '将从上游拉取最新列表，自动下架本地已删除的课程。',
    async onOk() {
      try {
        const raw = await syncSupplierStatusApi(selectedHid.value!);
        const res = raw;
        message.success(res.msg || '检查完成');
      } catch (e: any) { message.error(e?.message || '检查失败'); }
    },
  });
}

// ===== 单个添加弹窗 =====
const addVisible = ref(false);
const rateHint = ref('1.1');
const addForm = reactive({
  sort: '10', name: '', price: '', getnoun: '', noun: '', content: '',
  queryplat: '0', docking: '0', yunsuan: '*', status: '1', fenlei: '',
});

function openAdd(row: SupplierProductItem) {
  if (!selectedHid.value) { message.warning('请先选择货源'); return; }
  const base = row.price;
  const rate = parseFloat(rateHint.value) || 1;
  const price = isFinite(base) && rate > 0 ? (base * rate).toFixed(2) : '';
  Object.assign(addForm, {
    sort: String(row.sort || 10),
    name: row.name || '',
    price,
    getnoun: row.cid || '',
    noun: row.cid || '',
    content: row.content || '',
    queryplat: String(selectedHid.value),
    docking: String(selectedHid.value),
    yunsuan: '*',
    status: '1',
    fenlei: '',
  });
  addVisible.value = true;
}

async function submitAdd() {
  if (!addForm.name.trim()) { message.warning('平台名字不能为空'); return; }
  if (!addForm.price.trim()) { message.warning('定价不能为空'); return; }
  try {
    await addClassApi({ ...addForm });
    message.success('添加成功');
    addVisible.value = false;
    // 标记 states
    const it = rawList.value.find(x => x.cid === addForm.getnoun);
    if (it) it.states = 1;
  } catch (e: any) { message.error(e?.message || '添加失败'); }
}

// ===== 一键对接弹窗 =====
const yjdjVisible = ref(false);
const yjdjForm = reactive({ category: '999999', name: '', pricee: '1.1', fd: 1 });

function openYjdj() {
  if (!selectedHid.value) { message.warning('请先选择货源'); return; }
  Object.assign(yjdjForm, { category: '999999', name: '', pricee: '1.1', fd: 1 });
  yjdjVisible.value = true;
}

async function submitYjdj() {
  if (!selectedHid.value) return;
  const pricee = parseFloat(yjdjForm.pricee);
  if (!isFinite(pricee) || pricee <= 0) { message.warning('倍率不正确'); return; }
  try {
    const raw = await importSupplierApi({
      hid: selectedHid.value,
      pricee,
      category: yjdjForm.category,
      name: yjdjForm.name,
      fd: yjdjForm.fd,
    });
    const res = raw;
    message.success(res.msg || '对接成功');
    yjdjVisible.value = false;
    // 刷新
    fetchProducts();
  } catch (e: any) { message.error(e?.message || '对接失败'); }
}

// ===== 批量上架弹窗 =====
const batchVisible = ref(false);
const batchRunning = ref(false);
const batchProgress = ref(0);
const batchForm = reactive({
  rate: '1.1', sort: '10', fenlei: '', skipExists: true,
});

const batchCandidates = computed(() => {
  const items = rawList.value.filter(it => selectedKeys.value.includes(it.cid));
  if (batchForm.skipExists) return items.filter(it => it.states !== 1);
  return items;
});

function openBatch() {
  if (!selectedHid.value) { message.warning('请先选择货源'); return; }
  if (selectedKeys.value.length === 0) { message.warning('请先勾选要上架的商品'); return; }
  Object.assign(batchForm, { rate: '1.1', sort: '10', fenlei: '', skipExists: true });
  batchVisible.value = true;
}

async function submitBatch() {
  if (batchRunning.value) return;
  const rate = parseFloat(batchForm.rate);
  if (!isFinite(rate) || rate <= 0) { message.warning('倍率不正确'); return; }
  const items = batchCandidates.value;
  if (!items.length) { message.warning('没有可上架的商品'); return; }

  batchRunning.value = true;
  batchProgress.value = 0;
  let ok = 0, fail = 0;

  for (let i = 0; i < items.length; i++) {
    const row = items[i]!;
    const base = row.price;
    if (!isFinite(base)) { fail++; batchProgress.value = Math.round(((i + 1) / items.length) * 100); continue; }
    const newPrice = (base * rate).toFixed(2);
    try {
      await addClassApi({
        sort: String(row.sort || batchForm.sort || '10'),
        name: row.name,
        price: newPrice,
        getnoun: row.cid,
        noun: row.cid,
        content: row.content || '',
        queryplat: String(selectedHid.value),
        docking: String(selectedHid.value),
        yunsuan: '*',
        status: '1',
        fenlei: batchForm.fenlei,
      });
      ok++;
      row.states = 1;
    } catch { fail++; }
    batchProgress.value = Math.round(((i + 1) / items.length) * 100);
  }

  batchRunning.value = false;
  selectedKeys.value = [];
  message.success(`批量上架完成：成功 ${ok} 个，失败 ${fail} 个`);
  if (fail === 0) batchVisible.value = false;
}

// ===== 新建分类 =====
const newCatName = ref('');
const newCatLoading = ref(false);

async function quickCreateCategory() {
  const name = newCatName.value.trim();
  if (!name) { message.warning('请输入分类名称'); return; }
  newCatLoading.value = true;
  try {
    await saveCategoryApi({ name, status: '1', sort: 0 });
    message.success(`分类「${name}」创建成功`);
    newCatName.value = '';
    // 重新加载分类并自动选中新建的
    const [, cRaw] = await Promise.all([Promise.resolve(), getCategoryListApi()]);
    categories.value = cRaw;
    if (!Array.isArray(categories.value)) categories.value = [];
    const created = categories.value.find(c => c.name === name);
    if (created) {
      batchForm.fenlei = String(created.id);
      addForm.fenlei = String(created.id);
    }
  } catch (e: any) { message.error(e?.message || '创建失败'); }
  finally { newCatLoading.value = false; }
}

// ===== 批量替换关键词 =====
const replaceForm = reactive({ search: '', replace: '', scope: 'all', scopeId: '' });
const replaceLoading = ref(false);

async function submitReplace() {
  if (!replaceForm.search.trim()) { message.warning('请输入要替换的关键词'); return; }
  replaceLoading.value = true;
  try {
    const res = await batchReplaceKeywordApi(replaceForm.search, replaceForm.replace, replaceForm.scope, replaceForm.scopeId);
    message.success((res as any).msg || '替换成功');
  } catch (e: any) { message.error(e?.message || '替换失败'); }
  finally { replaceLoading.value = false; }
}

// ===== 批量添加前缀 =====
const prefixForm = reactive({ prefix: '', scope: 'all', scopeId: '' });
const prefixLoading = ref(false);

async function submitPrefix() {
  if (!prefixForm.prefix.trim()) { message.warning('请输入要添加的前缀'); return; }
  prefixLoading.value = true;
  try {
    const res = await batchAddPrefixApi(prefixForm.prefix, prefixForm.scope, prefixForm.scopeId);
    message.success((res as any).msg || '添加前缀成功');
  } catch (e: any) { message.error(e?.message || '添加前缀失败'); }
  finally { prefixLoading.value = false; }
}

// 全选所有结果
function selectAllFiltered() {
  selectedKeys.value = filteredList.value.map(it => it.cid);
  message.info(`已选中全部 ${selectedKeys.value.length} 个结果`);
}

// 取消所有选择
function clearAllSelected() {
  selectedKeys.value = [];
  message.info('已取消所有选择');
}

// 表格列
const columns = [
  { title: 'YCID', dataIndex: 'cid', key: 'cid', width: 90 },
  { title: '名称', dataIndex: 'name', key: 'name', ellipsis: true },
  { title: '内容', dataIndex: 'content', key: 'content', width: 200, ellipsis: true },
  { title: '分类', key: 'category', width: 120 },
  { title: '价格', key: 'price', width: 90 },
  { title: '排序', dataIndex: 'sort', key: 'sort', width: 70 },
  { title: '状态', key: 'states', width: 90 },
  { title: '操作', key: 'action', width: 80 },
];

onMounted(loadBase);
</script>

<template>
  <Page title="对接插件" content-class="p-4">

    <!-- 操作区 -->
    <Card class="mb-4">
      <div class="flex flex-wrap gap-3 items-center">
        <Select
          v-model:value="selectedHid"
          placeholder="选择货源"
          style="min-width: 220px; max-width: 380px"
          :dropdown-match-select-width="false"
          show-search
          allow-clear
          :filter-option="(input: string, option: any) => String(option.label || '').toLowerCase().includes(input.toLowerCase())"
        >
          <SelectOption v-for="s in suppliers" :key="s.hid" :value="s.hid" :label="`${s.name} (HID:${s.hid})`">
            {{ s.name }} (HID:{{ s.hid }})
          </SelectOption>
        </Select>

        <Button type="primary" :loading="loading" @click="fetchProducts">
          <template #icon><CloudDownloadOutlined /></template>
          拉取商品
        </Button>

        <Button danger @click="checkInvalid">
          <template #icon><WarningOutlined /></template>
          检查失效
        </Button>

        <Button type="primary" class="bg-green-600 border-green-600" @click="openYjdj">
          <template #icon><ThunderboltOutlined /></template>
          一键对接
        </Button>

        <Button
          class="bg-orange-500 border-orange-500 text-white"
          @click="openBatch"
        >
          <template #icon><UploadOutlined /></template>
          批量上架
          <span v-if="selectedKeys.length" class="ml-1">({{ selectedKeys.length }})</span>
        </Button>

        <div class="flex-1" />
        <Tag v-if="selectedHyName" color="blue">当前货源：{{ selectedHyName }}</Tag>
      </div>

      <!-- 筛选 -->
      <div class="flex gap-3 items-center mt-3" v-if="rawList.length > 0">
        <Input v-model:value="search.keyword" placeholder="搜索名称/内容" allow-clear style="width: 200px" @pressEnter="currentPage = 1" />
        <Select
          v-model:value="search.category"
          placeholder="上游分类"
          allow-clear
          show-search
          option-filter-prop="label"
          style="width: 200px"
          @change="currentPage = 1"
        >
          <SelectOption value="" label="全部分类">全部分类</SelectOption>
          <SelectOption v-for="c in uniqueCategories" :key="c.value" :value="c.value" :label="c.label">{{ c.label }}</SelectOption>
        </Select>
        <Button @click="search.keyword = ''; search.category = ''; currentPage = 1">
          <template #icon><ReloadOutlined /></template> 重置
        </Button>

        <Space>
          <Button size="small" @click="selectAllFiltered">全选所有结果</Button>
          <Button size="small" @click="clearAllSelected">清空选择</Button>
        </Space>

        <span class="text-sm text-gray-400">共 {{ totalRows }} 条</span>
      </div>
    </Card>

    <!-- 商品列表 -->
    <Card v-if="rawList.length > 0">
      <Table
        :columns="columns"
        :data-source="displayList"
        :loading="loading"
        :pagination="false"
        :row-selection="rowSelection"
        row-key="cid"
        size="small"
        bordered
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'category'">
            {{ record.category_name || record.fenlei || '-' }}
          </template>
          <template v-else-if="column.key === 'price'">
            <span class="font-semibold">¥{{ Number(record.price).toFixed(2) }}</span>
          </template>
          <template v-else-if="column.key === 'states'">
            <Tag :color="record.states === 1 ? 'green' : 'default'">
              {{ record.states === 1 ? '已添加' : '未添加' }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <Button type="primary" size="small" @click="openAdd(record)">
              <template #icon><PlusOutlined /></template>
            </Button>
          </template>
        </template>
      </Table>

      <div class="flex justify-center mt-4">
        <Pagination
          v-model:current="currentPage"
          :total="totalRows"
          :page-size="pageSize"
          :show-total="(total: number) => `共 ${total} 条`"
        />
      </div>
    </Card>

    <Card v-else-if="!loading">
      <div class="text-center text-gray-400 py-12">
        暂无数据 — 请先选择货源并点击「拉取商品」
      </div>
    </Card>

    <!-- 单个添加弹窗 -->
    <Modal v-model:open="addVisible" title="添加商品" @ok="submitAdd" ok-text="确定" width="680px">
      <div class="space-y-3">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">平台名字</label>
            <Input v-model:value="addForm.name" placeholder="名称" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">定价</label>
            <Input v-model:value="addForm.price" placeholder="例如 10.00" />
          </div>
        </div>
        <div class="grid grid-cols-3 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">排序</label>
            <Input v-model:value="addForm.sort" placeholder="默认10" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">算法</label>
            <Select v-model:value="addForm.yunsuan" style="width: 100%">
              <SelectOption value="*">乘法 (*)</SelectOption>
              <SelectOption value="+">加法 (+)</SelectOption>
            </Select>
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">状态</label>
            <Select v-model:value="addForm.status" style="width: 100%">
              <SelectOption value="1">上架</SelectOption>
              <SelectOption value="0">下架</SelectOption>
            </Select>
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">查询参数 (getnoun)</label>
            <Input v-model:value="addForm.getnoun" placeholder="货源CID" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">对接参数 (noun)</label>
            <Input v-model:value="addForm.noun" placeholder="货源CID" />
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">查询平台</label>
            <Select v-model:value="addForm.queryplat" style="width: 100%">
              <SelectOption value="0">自营</SelectOption>
              <SelectOption v-for="s in suppliers" :key="s.hid" :value="String(s.hid)">{{ s.name }}</SelectOption>
            </Select>
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">对接平台</label>
            <Select v-model:value="addForm.docking" style="width: 100%">
              <SelectOption value="0">自营</SelectOption>
              <SelectOption v-for="s in suppliers" :key="s.hid" :value="String(s.hid)">{{ s.name }}</SelectOption>
            </Select>
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">分类</label>
            <Select
              v-model:value="addForm.fenlei"
              placeholder="选择分类"
              allow-clear
              show-search
              option-filter-prop="label"
              style="width: 100%"
            >
              <SelectOption value="" label="无">无</SelectOption>
              <SelectOption v-for="c in categories" :key="c.id" :value="String(c.id)" :label="c.name">{{ c.name }}</SelectOption>
            </Select>
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">倍率参考</label>
            <Input v-model:value="rateHint" placeholder="例如 1.1" />
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">说明</label>
          <Input.TextArea v-model:value="addForm.content" :rows="2" placeholder="说明/注意事项" />
        </div>
      </div>
    </Modal>

    <!-- 一键对接弹窗 -->
    <Modal v-model:open="yjdjVisible" title="一键对接" @ok="submitYjdj" ok-text="确认对接" width="560px">
      <div class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">上游分类</label>
            <Select v-model:value="yjdjForm.category" style="width: 100%">
              <SelectOption value="999999">全部分类</SelectOption>
              <SelectOption v-for="c in uniqueCategories" :key="c.value" :value="c.value">{{ c.label }}</SelectOption>
            </Select>
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">本地分类名</label>
            <Select
              v-model:value="yjdjForm.name"
              placeholder="选择已有分类或留空"
              allow-clear
              show-search
              option-filter-prop="label"
              style="width: 100%"
            >
              <SelectOption value="" label="不指定">不指定</SelectOption>
              <SelectOption v-for="c in categories" :key="c.id" :value="c.name" :label="c.name">{{ c.name }}</SelectOption>
            </Select>
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">价格倍率</label>
            <Input v-model:value="yjdjForm.pricee" placeholder="例如 1.1" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">模式</label>
            <Select v-model:value="yjdjForm.fd" style="width: 100%">
              <SelectOption :value="1">只更新已对接项目</SelectOption>
              <SelectOption :value="0">更新并上架新项目</SelectOption>
            </Select>
          </div>
        </div>
      </div>
    </Modal>

    <!-- 批量替换关键词 -->
    <Card class="mb-4">
      <template #title>
        <Space><SwapOutlined /> 批量对关键词进行替换</Space>
      </template>
      <div class="flex flex-wrap gap-3 items-end">
        <div>
          <label class="block text-xs text-gray-500 mb-1">要替换的关键词</label>
          <Input v-model:value="replaceForm.search" placeholder="请输入要替换的关键词" style="width: 180px" />
        </div>
        <div>
          <label class="block text-xs text-gray-500 mb-1">替换后的关键词</label>
          <Input v-model:value="replaceForm.replace" placeholder="留空则删除关键词" style="width: 180px" />
        </div>
        <div>
          <label class="block text-xs text-gray-500 mb-1">范围</label>
          <Select v-model:value="replaceForm.scope" style="width: 150px" @change="replaceForm.scopeId = ''">
            <SelectOption value="all">对所有范围执行</SelectOption>
            <SelectOption value="cate">按分类</SelectOption>
            <SelectOption value="docking">按对接平台ID</SelectOption>
          </Select>
        </div>
        <div v-if="replaceForm.scope === 'cate'">
          <label class="block text-xs text-gray-500 mb-1">分类</label>
          <Select v-model:value="replaceForm.scopeId" placeholder="选择分类" allow-clear show-search option-filter-prop="label" style="width: 160px">
            <SelectOption v-for="c in categories" :key="c.id" :value="String(c.id)" :label="c.name">{{ c.name }}</SelectOption>
          </Select>
        </div>
        <div v-if="replaceForm.scope === 'docking'">
          <label class="block text-xs text-gray-500 mb-1">对接平台ID</label>
          <Input v-model:value="replaceForm.scopeId" placeholder="请输入ID" style="width: 150px" />
        </div>
        <Button type="primary" :loading="replaceLoading" @click="submitReplace" style="background: #722ed1; border-color: #722ed1">
          <template #icon><SwapOutlined /></template>
          更新关键词
        </Button>
      </div>
    </Card>

    <!-- 批量添加前缀 -->
    <Card class="mb-4">
      <template #title>
        <Space><TagOutlined /> 批量对商品添加前缀</Space>
      </template>
      <div class="flex flex-wrap gap-3 items-end">
        <div>
          <label class="block text-xs text-gray-500 mb-1">要新增的前缀</label>
          <Input v-model:value="prefixForm.prefix" placeholder="请输入要新增的前缀" style="width: 180px" />
        </div>
        <div>
          <label class="block text-xs text-gray-500 mb-1">范围</label>
          <Select v-model:value="prefixForm.scope" style="width: 150px" @change="prefixForm.scopeId = ''">
            <SelectOption value="all">对所有范围执行</SelectOption>
            <SelectOption value="cate">按分类</SelectOption>
            <SelectOption value="docking">按对接平台ID</SelectOption>
          </Select>
        </div>
        <div v-if="prefixForm.scope === 'cate'">
          <label class="block text-xs text-gray-500 mb-1">分类</label>
          <Select v-model:value="prefixForm.scopeId" placeholder="选择分类" allow-clear show-search option-filter-prop="label" style="width: 160px">
            <SelectOption v-for="c in categories" :key="c.id" :value="String(c.id)" :label="c.name">{{ c.name }}</SelectOption>
          </Select>
        </div>
        <div v-if="prefixForm.scope === 'docking'">
          <label class="block text-xs text-gray-500 mb-1">对接平台ID</label>
          <Input v-model:value="prefixForm.scopeId" placeholder="请输入ID" style="width: 150px" />
        </div>
        <Button type="primary" :loading="prefixLoading" @click="submitPrefix" style="background: #13c2c2; border-color: #13c2c2">
          <template #icon><TagOutlined /></template>
          添加前缀
        </Button>
      </div>
    </Card>

    <!-- 批量上架弹窗 -->
    <Modal v-model:open="batchVisible" title="批量上架" :closable="!batchRunning" :maskClosable="!batchRunning" :footer="null" width="560px">
      <div class="mb-3 text-sm text-gray-500">
        已选 <b class="text-gray-800">{{ selectedKeys.length }}</b> 个，
        本次将上架 <b class="text-gray-800">{{ batchCandidates.length }}</b> 个
        <span v-if="batchForm.skipExists">（已添加的跳过）</span>
      </div>

      <div class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-1">价格倍率</label>
            <Input v-model:value="batchForm.rate" placeholder="例如 1.1" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">默认排序</label>
            <Input v-model:value="batchForm.sort" placeholder="默认 10" />
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div class="col-span-2">
            <label class="block text-sm font-medium mb-1">分类</label>
            <div class="flex gap-2">
              <Select
                v-model:value="batchForm.fenlei"
                placeholder="选择分类"
                allow-clear
                show-search
                option-filter-prop="label"
                style="flex: 1"
              >
                <SelectOption value="" label="无">无</SelectOption>
                <SelectOption v-for="c in categories" :key="c.id" :value="String(c.id)" :label="c.name">{{ c.name }}</SelectOption>
              </Select>
              <Input v-model:value="newCatName" placeholder="新分类名称" style="width: 120px" />
              <Button type="primary" :loading="newCatLoading" @click="quickCreateCategory">
                <template #icon><PlusOutlined /></template>
                新建
              </Button>
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">跳过已添加</label>
            <div class="flex items-center h-8">
              <Switch v-model:checked="batchForm.skipExists" />
              <span class="text-xs text-gray-400 ml-2">开启后已添加的跳过</span>
            </div>
          </div>
        </div>

        <Progress v-if="batchRunning" :percent="batchProgress" :stroke-width="14" />

        <div class="flex justify-end gap-2 mt-2">
          <Button :disabled="batchRunning" @click="batchVisible = false">取消</Button>
          <Button type="primary" class="bg-orange-500 border-orange-500" :loading="batchRunning" @click="submitBatch">
            开始上架
          </Button>
        </div>
      </div>
    </Modal>
  </Page>
</template>

<style>
/* 货源选择器弹出框溢出父容器 overflow:hidden 限制 */
.ant-select-dropdown {
  z-index: 9999 !important;
}
</style>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Card, Table, Button, Input, Space, Tag, Modal, Radio, RadioGroup,
  Pagination, message, Select, SelectOption,
} from 'ant-design-vue';
import {
  PlusOutlined, DeleteOutlined, SearchOutlined,
} from '@ant-design/icons-vue';
import {
  getMiJiaListApi, saveMiJiaApi, deleteMiJiaApi, batchMiJiaApi,
  getCategoryListApi, getClassDropdownApi,
  type MiJiaItem, type CategoryItem, type ClassDropdownItem,
} from '#/api/admin';

const loading = ref(false);
const list = ref<MiJiaItem[]>([]);
const pagination = reactive({ page: 1, limit: 10, total: 0 });
const pagesize = ref(10);
const search = reactive({ selectedUid: undefined as number | undefined, keyword: '' });
const selectedRowKeys = ref<number[]>([]);
const uidOptions = ref<number[]>([]);
const categories = ref<CategoryItem[]>([]);
const classOptions = ref<ClassDropdownItem[]>([]);

// 模式文本映射（按旧系统）
const modeText = (mode: string) => {
  const map: Record<string, string> = {
    '0': '价格的基础上扣除',
    '1': '倍数的基础上扣除',
    '2': '直接定价',
    '3': '按照赠送比例折扣',
    '4': '按倍率定价',
  };
  return map[String(mode)] || '未知类型';
};

// ===== 添加弹窗 =====
const addVisible = ref(false);
const addForm = reactive({
  uid: '',
  setType: 'single' as 'single' | 'batch',
  cid: undefined as number | undefined,
  fenlei: undefined as number | undefined,
  pricingMethod: 'direct' as 'direct' | 'multiplier',
  directPrice: '',
  multiplier: '',
});

const selectedProductPrice = ref(0);
const calculatedPrice = computed(() => {
  const m = Number(addForm.multiplier || 0);
  if (!addForm.cid || !Number.isFinite(m)) return '0.00';
  return (Number(selectedProductPrice.value || 0) * m).toFixed(2);
});

const categoryProducts = computed(() => {
  if (addForm.setType !== 'batch' || !addForm.fenlei) return [];
  return classOptions.value.filter(it => String(it.fenlei) === String(addForm.fenlei));
});

watch(() => addForm.cid, (cid) => {
  if (!cid) { selectedProductPrice.value = 0; return; }
  const p = classOptions.value.find(x => x.cid === cid);
  selectedProductPrice.value = Number(p?.price || 0);
});

// ===== 编辑弹窗 =====
const editVisible = ref(false);
const editForm = reactive({ mid: 0, uid: '', cid: undefined as number | undefined, mode: '2', price: '' });

// ===== 数据加载 =====
async function loadData(page = 1) {
  loading.value = true;
  pagination.page = page;
  try {
    const raw = await getMiJiaListApi({
      page, limit: pagesize.value,
      uid: search.selectedUid,
      keyword: search.keyword || undefined,
    });
    const res = raw;
    list.value = res.list || [];
    pagination.total = res.pagination?.total || 0;
    uidOptions.value = res.uids || [];
    selectedRowKeys.value = [];
  } catch (e) { console.error(e); }
  finally { loading.value = false; }
}

// ===== 添加 =====
function openAdd() {
  Object.assign(addForm, {
    uid: search.selectedUid ? String(search.selectedUid) : '',
    setType: 'single', cid: undefined, fenlei: undefined,
    pricingMethod: 'direct', directPrice: '', multiplier: '',
  });
  selectedProductPrice.value = 0;
  addVisible.value = true;
}

async function submitAdd() {
  if (!addForm.uid || !String(addForm.uid).trim()) { message.error('请输入用户UID'); return; }

  if (addForm.setType === 'single' && !addForm.cid) { message.error('请选择商品'); return; }
  if (addForm.setType === 'batch' && !addForm.fenlei) { message.error('请选择分类'); return; }

  if (addForm.pricingMethod === 'direct') {
    if (!addForm.directPrice.trim()) { message.error('请输入定价金额'); return; }
  } else {
    if (!addForm.multiplier.trim()) { message.error('请输入倍率'); return; }
  }

  try {
    if (addForm.setType === 'single') {
      let finalMode = '2';
      let finalPrice = addForm.directPrice;
      if (addForm.pricingMethod === 'multiplier') {
        finalPrice = calculatedPrice.value;
        finalMode = '2';
      }
      await saveMiJiaApi({ uid: Number(addForm.uid), cid: addForm.cid!, mode: finalMode, price: finalPrice });
      message.success('添加成功');
    } else {
      let batchMode = '2';
      let batchPrice = addForm.directPrice;
      if (addForm.pricingMethod === 'multiplier') {
        batchMode = '4';
        batchPrice = addForm.multiplier;
      }
      const raw = await batchMiJiaApi({ uid: Number(addForm.uid), fenlei: addForm.fenlei!, mode: batchMode, price: batchPrice });
      const res = raw;
      message.success(res.msg || '批量添加成功');
    }
    addVisible.value = false;
    loadData(1);
  } catch (e: any) { message.error(e?.message || '添加失败'); }
}

// ===== 编辑 =====
function openEdit(item: MiJiaItem) {
  Object.assign(editForm, {
    mid: item.mid, uid: String(item.uid), cid: item.cid,
    mode: String(item.mode ?? '2'), price: String(item.price ?? ''),
  });
  editVisible.value = true;
}

async function submitEdit() {
  if (!editForm.uid || !String(editForm.uid).trim()) { message.error('请输入用户UID'); return; }
  if (!editForm.cid) { message.error('请选择商品'); return; }
  if (!editForm.price.trim()) { message.error('请输入金额或倍率'); return; }
  try {
    await saveMiJiaApi({ mid: editForm.mid, uid: Number(editForm.uid), cid: editForm.cid!, mode: editForm.mode, price: editForm.price });
    message.success('更新成功');
    editVisible.value = false;
    loadData(pagination.page);
  } catch (e: any) { message.error(e?.message || '更新失败'); }
}

// ===== 删除 =====
async function deleteOne(mid: number) {
  Modal.confirm({
    title: '提示', content: `确定删除 ID=${mid} 吗？`,
    async onOk() {
      try {
        await deleteMiJiaApi([mid]);
        message.success('删除成功');
        loadData(pagination.page);
      } catch (e: any) { message.error(e?.message || '删除失败'); }
    },
  });
}

async function batchDelete() {
  if (selectedRowKeys.value.length === 0) { message.error('请选择要删除的项'); return; }
  Modal.confirm({
    title: '提示', content: `确定批量删除 ${selectedRowKeys.value.length} 条记录吗？`,
    async onOk() {
      try {
        await deleteMiJiaApi(selectedRowKeys.value);
        message.success('操作成功');
        selectedRowKeys.value = [];
        loadData(1);
      } catch (e: any) { message.error(e?.message || '删除失败'); }
    },
  });
}

const columns = [
  { title: '操作', key: 'action', width: 130, align: 'center' as const },
  { title: 'ID', dataIndex: 'mid', key: 'mid', width: 90, align: 'center' as const },
  { title: 'UID', dataIndex: 'uid', key: 'uid', width: 90, align: 'center' as const },
  { title: '课程', key: 'class', width: 260 },
  { title: '类型', key: 'mode', width: 200 },
  { title: '金额/倍数', dataIndex: 'price', key: 'price', width: 140, align: 'center' as const },
  { title: '添加时间', dataIndex: 'addtime', key: 'addtime', width: 170, align: 'center' as const },
];

const rowSelection = {
  selectedRowKeys: selectedRowKeys,
  onChange: (keys: number[]) => { selectedRowKeys.value = keys; },
};

onMounted(async () => {
  const [, catsRaw, clsRaw] = await Promise.all([
    loadData(1),
    getCategoryListApi().catch(() => []),
    getClassDropdownApi().catch(() => []),
  ]);
  const cats = catsRaw;
  const cls = clsRaw;
  categories.value = Array.isArray(cats) ? cats : [];
  classOptions.value = Array.isArray(cls) ? cls : [];
});
</script>

<template>
  <Page title="密价设置" content-class="p-4">

    <!-- 搜索/操作区 -->
    <Card title="用户密价管理" class="mb-4">
      <template #extra>
        <Button type="primary" size="small" @click="openAdd">
          <template #icon><PlusOutlined /></template>
          添加密价
        </Button>
      </template>

      <div class="flex flex-wrap gap-3 items-center mb-3">
        <Select
          v-model:value="search.selectedUid"
          placeholder="选择用户UID"
          style="width: 180px"
          allow-clear
          show-search
          option-filter-prop="label"
        >
          <SelectOption v-for="u in uidOptions" :key="u" :value="u" :label="String(u)">{{ u }}</SelectOption>
        </Select>

        <Input v-model:value="search.keyword" placeholder="输入关键词（课程名模糊）" style="width: 240px" @press-enter="loadData(1)" />

        <Button type="primary" @click="loadData(1)">
          <template #icon><SearchOutlined /></template>
          搜索
        </Button>

        <Select v-model:value="pagesize" style="width: 100px" @change="loadData(1)">
          <SelectOption :value="10">10/页</SelectOption>
          <SelectOption :value="20">20/页</SelectOption>
          <SelectOption :value="50">50/页</SelectOption>
          <SelectOption :value="100">100/页</SelectOption>
        </Select>
      </div>

      <div class="flex flex-wrap gap-2 items-center">
        <Button danger size="small" @click="batchDelete">
          <template #icon><DeleteOutlined /></template>
          批量删除（选中）
        </Button>
      </div>
    </Card>

    <!-- 密价列表 -->
    <Card title="密价列表">
      <Table
        :columns="columns" :data-source="list" :loading="loading"
        :pagination="false" :row-selection="rowSelection"
        row-key="mid" size="small" bordered
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'action'">
            <Space size="small">
              <Button type="primary" size="small" @click="openEdit(record)">编辑</Button>
              <Button danger size="small" @click="deleteOne(record.mid)">删除</Button>
            </Space>
          </template>
          <template v-else-if="column.key === 'class'">
            {{ record.classname || '-' }}
          </template>
          <template v-else-if="column.key === 'mode'">
            <Tag color="processing">{{ modeText(record.mode) }}</Tag>
          </template>
        </template>
      </Table>

      <div class="flex justify-center mt-4" v-if="pagination.total > 0">
        <Pagination
          v-model:current="pagination.page"
          :total="pagination.total"
          :page-size="pagesize"
          :show-size-changer="false"
          :show-total="(total: number) => `共 ${total} 条`"
          @change="(p: number) => loadData(p)"
        />
      </div>
    </Card>

    <!-- 添加密价弹窗 -->
    <Modal v-model:open="addVisible" title="密价添加" @ok="submitAdd" ok-text="确定" cancel-text="取消" width="640px">
      <div class="space-y-4 py-2">
        <div>
          <label class="block text-sm font-medium mb-1">UID</label>
          <Input v-model:value="addForm.uid" placeholder="请输入用户UID" />
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">设置方式</label>
          <RadioGroup v-model:value="addForm.setType" @change="() => { addForm.cid = undefined; addForm.fenlei = undefined; selectedProductPrice = 0; }">
            <Radio value="single">单个商品设置</Radio>
            <Radio value="batch">按分类批量设置</Radio>
          </RadioGroup>
        </div>

        <!-- 单个商品选择 -->
        <div v-if="addForm.setType === 'single'">
          <label class="block text-sm font-medium mb-1">选择商品</label>
          <Select
            v-model:value="addForm.cid"
            placeholder="请选择商品"
            style="width: 100%"
            show-search
            allow-clear
            option-filter-prop="label"
          >
            <SelectOption v-for="c in classOptions" :key="c.cid" :value="c.cid" :label="`${c.name} (${c.price}币)`">
              {{ c.name }} ({{ c.price }}币)
            </SelectOption>
          </Select>
        </div>

        <!-- 分类批量选择 -->
        <div v-if="addForm.setType === 'batch'">
          <label class="block text-sm font-medium mb-1">选择分类</label>
          <Select
            v-model:value="addForm.fenlei"
            placeholder="请选择分类"
            show-search
            option-filter-prop="label"
            style="width: 100%"
            allow-clear
          >
            <SelectOption v-for="f in categories" :key="f.id" :value="f.id" :label="f.name">{{ f.name }}</SelectOption>
          </Select>

          <div v-if="categoryProducts.length > 0" class="mt-3 p-3 rounded" style="background: #f0f9eb; border-left: 4px solid #67c23a;">
            <div><strong>分类包含商品：</strong>共 {{ categoryProducts.length }} 个商品</div>
            <div class="mt-2" style="max-height: 120px; overflow: auto; font-size: 12px; color: #666;">
              <div v-for="p in categoryProducts" :key="p.cid">{{ p.name }}（原价：{{ p.price }}币）</div>
            </div>
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">定价方式</label>
          <RadioGroup v-model:value="addForm.pricingMethod">
            <Radio value="direct">直接定价</Radio>
            <Radio value="multiplier">按倍率定价</Radio>
          </RadioGroup>
        </div>

        <!-- 直接定价 -->
        <div v-if="addForm.pricingMethod === 'direct'">
          <label class="block text-sm font-medium mb-1">{{ addForm.setType === 'single' ? '定价金额' : '统一定价金额' }}</label>
          <Input v-model:value="addForm.directPrice" placeholder="请输入定价金额" addon-after="币" />
        </div>

        <!-- 倍率定价 -->
        <div v-if="addForm.pricingMethod === 'multiplier'">
          <label class="block text-sm font-medium mb-1">倍率</label>
          <Input v-model:value="addForm.multiplier" placeholder="如 0.8 表示 8 折" addon-after="倍" />
        </div>

        <!-- 单商品倍率预览 -->
        <div v-if="addForm.setType === 'single' && addForm.pricingMethod === 'multiplier' && addForm.cid && addForm.multiplier" class="p-3 rounded" style="background: #f0f9ff; border-left: 4px solid #1890ff;">
          <strong>价格预览：</strong>
          原价 {{ selectedProductPrice }} 币 × {{ addForm.multiplier }} 倍 =
          <span style="color: #f56c6c; font-weight: bold;">{{ calculatedPrice }} 币</span>
        </div>

        <!-- 分类倍率预览 -->
        <div v-if="addForm.setType === 'batch' && addForm.pricingMethod === 'multiplier' && addForm.fenlei && addForm.multiplier && categoryProducts.length > 0" class="p-3 rounded" style="background: #f0f9ff; border-left: 4px solid #1890ff;">
          <strong>批量价格预览（前5个商品）：</strong>
          <div class="mt-2" style="font-size: 12px;">
            <div v-for="p in categoryProducts.slice(0, 5)" :key="p.cid" class="mb-1">
              {{ p.name }}：{{ p.price }}币 × {{ addForm.multiplier }}倍 =
              <span style="color: #f56c6c; font-weight: bold;">{{ (Number(p.price || 0) * Number(addForm.multiplier || 0)).toFixed(2) }}币</span>
            </div>
            <div v-if="categoryProducts.length > 5" style="color: #999;">... 还有 {{ categoryProducts.length - 5 }} 个商品</div>
          </div>
        </div>
      </div>
    </Modal>

    <!-- 编辑密价弹窗 -->
    <Modal v-model:open="editVisible" title="密价修改" @ok="submitEdit" ok-text="确定" cancel-text="取消" width="640px">
      <div class="space-y-4 py-2">
        <div>
          <label class="block text-sm font-medium mb-1">UID</label>
          <Input v-model:value="editForm.uid" placeholder="请输入用户UID" />
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">商品</label>
          <Select
            v-model:value="editForm.cid"
            placeholder="请选择商品"
            style="width: 100%"
            show-search
            allow-clear
            option-filter-prop="label"
          >
            <SelectOption v-for="c in classOptions" :key="c.cid" :value="c.cid" :label="`${c.name} (${c.price}币)`">
              {{ c.name }} ({{ c.price }}币)
            </SelectOption>
          </Select>
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">类型</label>
          <Select v-model:value="editForm.mode" style="width: 100%">
            <SelectOption value="0">价格的基础上扣除</SelectOption>
            <SelectOption value="1">倍数的基础上扣除</SelectOption>
            <SelectOption value="2">直接定价</SelectOption>
            <SelectOption value="4">按倍率定价</SelectOption>
          </Select>
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">金额/倍数</label>
          <Input v-model:value="editForm.price" placeholder="请输入金额或倍率" />
        </div>
      </div>
    </Modal>

  </Page>
</template>

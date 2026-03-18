<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue';

import { Page } from '@vben/common-ui';

import {
  PlusOutlined, DeleteOutlined, SearchOutlined,
} from '@ant-design/icons-vue';
import {
  Card, Table, Button, Input, Space, Tag, Modal, Radio, RadioGroup,
  Pagination, message, Select, SelectOption,
} from 'ant-design-vue';

import {
  getMiJiaListApi, saveMiJiaApi, deleteMiJiaApi, batchMiJiaApi,
  getCategoryListApi, getClassDropdownApi,
  type MiJiaItem, type CategoryItem, type ClassDropdownItem,
} from '#/api/admin';

const loading = ref(false);
const list = ref<MiJiaItem[]>([]);
const pagination = reactive({ page: 1, limit: 10, total: 0 });
const pagesize = ref(10);
const search = reactive({
  selectedUid: undefined as number | undefined,
  keyword: '',
});
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
    '3': '按倍率定价',
    // 兼容未迁移的旧记录：旧版倍率模式曾使用 4。
    '4': '按倍率定价',
  };
  return map[String(mode)] || '未知类型';
};

// ===== 添加弹窗 =====
const addVisible = ref(false);
const addForm = reactive({
  uid: '',
  setType: 'single' as 'batch' | 'single',
  cid: undefined as number | undefined,
  fenlei: undefined as number | undefined,
  mode: '2' as '0' | '1' | '2' | '3',
  priceValue: '',
  multiplier: '',
});

const selectedProductPrice = ref(0);
const isAddMultiplierMode = computed(() => addForm.mode === '3');
const addValueLabel = computed(() => {
  switch (addForm.mode) {
    case '0': {
      return '扣减金额';
    }
    case '1': {
      return '扣减金额';
    }
    case '2': {
      return addForm.setType === 'single' ? '定价金额' : '统一定价金额';
    }
    default: {
      return '金额';
    }
  }
});
const addValuePlaceholder = computed(() => {
  switch (addForm.mode) {
    case '0': {
      return '请输入扣减金额';
    }
    case '1': {
      return '请输入扣减金额';
    }
    case '2': {
      return '请输入定价金额';
    }
    default: {
      return '请输入金额';
    }
  }
});
const calculatedPrice = computed(() => {
  const m = Number(addForm.multiplier || 0);
  if (!addForm.cid || !Number.isFinite(m)) return '0.00';
  return (Number(selectedProductPrice.value || 0) * m).toFixed(2);
});

const categoryProducts = computed(() => {
  if (addForm.setType !== 'batch' || !addForm.fenlei) return [];
  return classOptions.value.filter(
    (it) => String(it.fenlei) === String(addForm.fenlei),
  );
});

watch(
  () => addForm.cid,
  (cid) => {
    if (!cid) {
      selectedProductPrice.value = 0;
      return;
    }
    const p = classOptions.value.find((x) => x.cid === cid);
    selectedProductPrice.value = Number(p?.price || 0);
  },
);

// ===== 编辑弹窗 =====
const editVisible = ref(false);
const editForm = reactive({
  mid: 0,
  uid: '',
  cid: undefined as number | undefined,
  mode: '2',
  price: '',
});

// ===== 数据加载 =====
async function loadData(page = 1) {
  loading.value = true;
  pagination.page = page;
  try {
    const raw = await getMiJiaListApi({
      page,
      limit: pagesize.value,
      uid: search.selectedUid,
      keyword: search.keyword || undefined,
    });
    const res = raw;
    list.value = res.list || [];
    pagination.total = res.pagination?.total || 0;
    uidOptions.value = res.uids || [];
    selectedRowKeys.value = [];
  } catch (error) {
    console.error(error);
  } finally {
    loading.value = false;
  }
}

// ===== 添加 =====
function openAdd() {
  Object.assign(addForm, {
    uid: search.selectedUid ? String(search.selectedUid) : '',
    setType: 'single',
    cid: undefined,
    fenlei: undefined,
    mode: '2',
    priceValue: '',
    multiplier: '',
  });
  selectedProductPrice.value = 0;
  addVisible.value = true;
}

async function submitAdd() {
  if (!addForm.uid || !String(addForm.uid).trim()) {
    message.error('请输入用户UID');
    return;
  }

  if (addForm.setType === 'single' && !addForm.cid) {
    message.error('请选择商品');
    return;
  }
  if (addForm.setType === 'batch' && !addForm.fenlei) {
    message.error('请选择分类');
    return;
  }

  if (isAddMultiplierMode.value) {
    if (!addForm.multiplier.trim()) {
      message.error('请输入倍率');
      return;
    }
  } else {
    if (!addForm.priceValue.trim()) {
      message.error('请输入金额');
      return;
    }
  }

  try {
    if (addForm.setType === 'single') {
      const finalMode = addForm.mode;
      const finalPrice = isAddMultiplierMode.value
        ? addForm.multiplier
        : addForm.priceValue;
      await saveMiJiaApi({
        uid: Number(addForm.uid),
        cid: addForm.cid!,
        mode: finalMode,
        price: finalPrice,
      });
      message.success('添加成功');
    } else {
      const batchMode = addForm.mode;
      const batchPrice = isAddMultiplierMode.value
        ? addForm.multiplier
        : addForm.priceValue;
      const raw = await batchMiJiaApi({
        uid: Number(addForm.uid),
        fenlei: addForm.fenlei!,
        mode: batchMode,
        price: batchPrice,
      });
      const res = raw;
      message.success(res.msg || '批量添加成功');
    }
    addVisible.value = false;
    loadData(1);
  } catch (error: any) { message.error(error?.message || '添加失败'); }
}

// ===== 编辑 =====
function openEdit(item: MiJiaItem) {
  const currentMode = String(item.mode ?? '2');
  Object.assign(editForm, {
    mid: item.mid,
    uid: String(item.uid),
    cid: item.cid,
    // 兼容开发库里尚未迁移的数据，把旧 4 映射成新的 3。
    mode: currentMode === '4' ? '3' : currentMode,
    price: String(item.price ?? ''),
  });
  editVisible.value = true;
}

async function submitEdit() {
  if (!editForm.uid || !String(editForm.uid).trim()) {
    message.error('请输入用户UID');
    return;
  }
  if (!editForm.cid) {
    message.error('请选择商品');
    return;
  }
  if (!editForm.price.trim()) {
    message.error('请输入金额或倍率');
    return;
  }
  try {
    await saveMiJiaApi({
      mid: editForm.mid,
      uid: Number(editForm.uid),
      cid: editForm.cid!,
      mode: editForm.mode,
      price: editForm.price,
    });
    message.success('更新成功');
    editVisible.value = false;
    loadData(pagination.page);
  } catch (error: any) { message.error(error?.message || '更新失败'); }
}

// ===== 删除 =====
async function deleteOne(mid: number) {
  Modal.confirm({
    title: '提示',
    content: `确定删除 ID=${mid} 吗？`,
    async onOk() {
      try {
        await deleteMiJiaApi([mid]);
        message.success('删除成功');
        loadData(pagination.page);
      } catch (error: any) { message.error(error?.message || '删除失败'); }
    },
  });
}

async function batchDelete() {
  if (selectedRowKeys.value.length === 0) {
    message.error('请选择要删除的项');
    return;
  }
  Modal.confirm({
    title: '提示',
    content: `确定批量删除 ${selectedRowKeys.value.length} 条记录吗？`,
    async onOk() {
      try {
        await deleteMiJiaApi(selectedRowKeys.value);
        message.success('操作成功');
        selectedRowKeys.value = [];
        loadData(1);
      } catch (error: any) { message.error(error?.message || '删除失败'); }
    },
  });
}

const columns = [
  { title: '操作', key: 'action', width: 130, align: 'center' as const },
  {
    title: 'ID',
    dataIndex: 'mid',
    key: 'mid',
    width: 90,
    align: 'center' as const,
  },
  {
    title: 'UID',
    dataIndex: 'uid',
    key: 'uid',
    width: 90,
    align: 'center' as const,
  },
  { title: '课程', key: 'class', width: 260 },
  { title: '类型', key: 'mode', width: 200 },
  {
    title: '金额/倍数',
    dataIndex: 'price',
    key: 'price',
    width: 140,
    align: 'center' as const,
  },
  {
    title: '添加时间',
    dataIndex: 'addtime',
    key: 'addtime',
    width: 170,
    align: 'center' as const,
  },
];

const rowSelection = {
  selectedRowKeys,
  onChange: (keys: number[]) => {
    selectedRowKeys.value = keys;
  },
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
  <Page content-class="p-4" title="密价设置">
<!-- 搜索/操作区 -->
    <Card class="mb-4" title="用户密价管理">
      <template #extra>
        <Button size="small" type="primary" @click="openAdd">
          <template #icon><PlusOutlined /></template>
          添加密价
        </Button>
      </template>

      <div class="mb-3 flex flex-wrap items-center gap-3">
        <Select
          v-model:value="search.selectedUid"
          allow-clear
          option-filter-prop="label"
          placeholder="选择用户UID"
          show-search
          style="width: 180px"
        >
          <SelectOption
            v-for="u in uidOptions"
            :key="u"
            :value="u"
            :label="String(u)"
            >{{ u }}</SelectOption
          >
        </Select>

        <Input
          v-model:value="search.keyword"
          placeholder="输入关键词（课程名模糊）"
          style="width: 240px"
          @press-enter="loadData(1)"
        />

        <Button type="primary" @click="loadData(1)">
          <template #icon><SearchOutlined /></template>
          搜索
        </Button>

        <Select
          v-model:value="pagesize"
          style="width: 100px"
          @change="loadData(1)"
        >
          <SelectOption :value="10">10/页</SelectOption>
          <SelectOption :value="20">20/页</SelectOption>
          <SelectOption :value="50">50/页</SelectOption>
          <SelectOption :value="100">100/页</SelectOption>
        </Select>
      </div>

      <div class="flex flex-wrap items-center gap-2">
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
        bordered row-key="mid" size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'action'">
            <Space size="small">
              <Button size="small" type="primary" @click="openEdit(record)">编辑</Button>
              <Button danger size="small" @click="deleteOne(record.mid)"
                >删除</Button
              >
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

      <div v-if="pagination.total > 0" class="flex justify-center mt-4">
        <Pagination
          v-model:current="pagination.page"
          :page-size="pagesize"
          :show-size-changer="false"
          :show-total="(total: number) => `共 ${total} 条`"
          :total="pagination.total"
          @change="(p: number) => loadData(p)"
        />
      </div>
    </Card>

    <!-- 添加密价弹窗 -->
    <Modal
      v-model:open="addVisible"
      title="密价添加"
      @ok="submitAdd"
      ok-text="确定"
      cancel-text="取消"
      width="640px"
    >
      <div class="space-y-4 py-2">
        <div>
          <label class="mb-1 block text-sm font-medium">UID</label>
          <Input v-model:value="addForm.uid" placeholder="请输入用户UID" />
        </div>

        <div>
          <label class="mb-1 block text-sm font-medium">设置方式</label>
          <RadioGroup
            v-model:value="addForm.setType"
            @change="
              () => {
                addForm.cid = undefined;
                addForm.fenlei = undefined;
                selectedProductPrice = 0;
              }
            "
          >
            <Radio value="single">单个商品设置</Radio>
            <Radio value="batch">按分类批量设置</Radio>
          </RadioGroup>
        </div>

        <!-- 单个商品选择 -->
        <div v-if="addForm.setType === 'single'">
          <label class="mb-1 block text-sm font-medium">选择商品</label>
          <Select
            v-model:value="addForm.cid"
            allow-clear
            option-filter-prop="label"
            placeholder="请选择商品"
            show-search
            style="width: 100%"
          >
            <SelectOption
              v-for="c in classOptions"
              :key="c.cid"
              :value="c.cid"
              :label="`${c.name} (${c.price}币)`"
            >
              {{ c.name }} ({{ c.price }}币)
            </SelectOption>
          </Select>
        </div>

        <!-- 分类批量选择 -->
        <div v-if="addForm.setType === 'batch'">
          <label class="mb-1 block text-sm font-medium">选择分类</label>
          <Select
            v-model:value="addForm.fenlei"
            allow-clear
            option-filter-prop="label"
            placeholder="请选择分类"
            show-search
            style="width: 100%"
          >
            <SelectOption
              v-for="f in categories"
              :key="f.id"
              :value="f.id"
              :label="f.name"
              >{{ f.name }}</SelectOption
            >
          </Select>

          <div
            v-if="categoryProducts.length > 0"
            class="mt-3 rounded bg-green-50 p-3 dark:bg-green-900/20"
            style="border-left: 4px solid #67c23a"
          >
            <div>
              <strong>分类包含商品：</strong>共
              {{ categoryProducts.length }} 个商品
            </div>
            <div
              class="mt-2 text-gray-600 dark:text-gray-400"
              style="max-height: 120px; overflow: auto; font-size: 12px"
            >
              <div v-for="p in categoryProducts" :key="p.cid">
                {{ p.name }}（原价：{{ p.price }}币）
              </div>
            </div>
          </div>
        </div>

        <div>
          <label class="mb-1 block text-sm font-medium">类型</label>
          <RadioGroup v-model:value="addForm.mode">
            <Radio value="0">价格的基础上扣除</Radio>
            <Radio value="1">倍数的基础上扣除</Radio>
            <Radio value="2">直接定价</Radio>
            <Radio value="3">按倍率定价</Radio>
          </RadioGroup>
        </div>

        <div v-if="!isAddMultiplierMode">
          <label class="mb-1 block text-sm font-medium">{{
            addValueLabel
          }}</label>
          <Input
            v-model:value="addForm.priceValue"
            :placeholder="addValuePlaceholder"
            addon-after="币"
          />
        </div>

        <div v-else>
          <label class="mb-1 block text-sm font-medium">倍率</label>
          <Input
            v-model:value="addForm.multiplier"
            placeholder="如 0.8 表示 8 折"
            addon-after="倍"
          />
        </div>

        <!-- 单商品倍率预览 -->
        <div
          v-if="
            addForm.setType === 'single' &&
            isAddMultiplierMode &&
            addForm.cid &&
            addForm.multiplier
          "
          class="rounded bg-blue-50 p-3 dark:bg-blue-900/20"
          style="border-left: 4px solid #1890ff"
        >
          <strong>价格预览：</strong>
          原价 {{ selectedProductPrice }} 币 × {{ addForm.multiplier }} 倍 =
          <span style="color: #f56c6c; font-weight: bold"
            >{{ calculatedPrice }} 币</span
          >
        </div>

        <!-- 分类倍率预览 -->
        <div
          v-if="
            addForm.setType === 'batch' &&
            isAddMultiplierMode &&
            addForm.fenlei &&
            addForm.multiplier &&
            categoryProducts.length > 0
          "
          class="rounded bg-blue-50 p-3 dark:bg-blue-900/20"
          style="border-left: 4px solid #1890ff"
        >
          <strong>批量价格预览（前5个商品）：</strong>
          <div class="mt-2" style="font-size: 12px">
            <div
              v-for="p in categoryProducts.slice(0, 5)"
              :key="p.cid"
              class="mb-1"
            >
              {{ p.name }}：{{ p.price }}币 × {{ addForm.multiplier }}倍 =
              <span style="color: #f56c6c; font-weight: bold"
                >{{
                  (
                    Number(p.price || 0) * Number(addForm.multiplier || 0)
                  ).toFixed(2)
                }}币</span
              >
            </div>
            <div
              v-if="categoryProducts.length > 5"
              class="text-gray-400 dark:text-gray-500"
            >
              ... 还有 {{ categoryProducts.length - 5 }} 个商品
            </div>
          </div>
        </div>
      </div>
    </Modal>

    <!-- 编辑密价弹窗 -->
    <Modal
      v-model:open="editVisible"
      title="密价修改"
      @ok="submitEdit"
      ok-text="确定"
      cancel-text="取消"
      width="640px"
    >
      <div class="space-y-4 py-2">
        <div>
          <label class="mb-1 block text-sm font-medium">UID</label>
          <Input v-model:value="editForm.uid" placeholder="请输入用户UID" />
        </div>

        <div>
          <label class="mb-1 block text-sm font-medium">商品</label>
          <Select
            v-model:value="editForm.cid"
            allow-clear
            option-filter-prop="label"
            placeholder="请选择商品"
            show-search
            style="width: 100%"
          >
            <SelectOption
              v-for="c in classOptions"
              :key="c.cid"
              :value="c.cid"
              :label="`${c.name} (${c.price}币)`"
            >
              {{ c.name }} ({{ c.price }}币)
            </SelectOption>
          </Select>
        </div>

        <div>
          <label class="mb-1 block text-sm font-medium">类型</label>
          <Select v-model:value="editForm.mode" style="width: 100%">
            <SelectOption value="0">价格的基础上扣除</SelectOption>
            <SelectOption value="1">倍数的基础上扣除</SelectOption>
            <SelectOption value="2">直接定价</SelectOption>
            <SelectOption value="3">按倍率定价</SelectOption>
          </Select>
        </div>

        <div>
          <label class="mb-1 block text-sm font-medium">金额/倍数</label>
          <Input
            v-model:value="editForm.price"
            placeholder="请输入金额或倍率"
          />
        </div>
      </div>
    </Modal>
  </Page>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { Page } from '@vben/common-ui';
import {
  Button,
  Card,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  message,
} from 'ant-design-vue';
import {
  DeleteOutlined,
  EditOutlined,
  PlusOutlined,
  SearchOutlined,
  UnorderedListOutlined,
} from '@ant-design/icons-vue';
import {
  getClassCategoriesApi,
  getClassListPagedApi,
  type ClassCategory,
  type ClassItem,
} from '#/api/class';
import {
  deleteTenantProductApi,
  getTenantMallCategoriesApi,
  getTenantProductsApi,
  type TenantMallCategory,
  saveTenantProductApi,
  type TenantProduct,
} from '#/api/tenant';

type ProductForm = {
  id: number;
  cid: number;
  retail_price: number;
  sort: number;
  status: number;
  display_name: string;
  cover_url: string;
  description: string;
  category_id?: number;
  category_name: string;
};

const loading = ref(false);
const products = ref<TenantProduct[]>([]);
const categories = ref<ClassCategory[]>([]);
const mallCategories = ref<TenantMallCategory[]>([]);
const selectedClass = ref<ClassItem | null>(null);

const editVisible = ref(false);
const pickerVisible = ref(false);
const pickerLoading = ref(false);
const pickerSubmitting = ref(false);
const pickerSelectingAll = ref(false);
const pickerMode = ref<'single' | 'batch'>('single');

const pickerList = ref<ClassItem[]>([]);
const pickerSelectedRowKeys = ref<number[]>([]);
const pickerSelectedCourses = ref<ClassItem[]>([]);

const pickerPagination = reactive({
  page: 1,
  limit: 10,
  total: 0,
});

const pickerFilters = reactive<{
  fenlei: number | undefined;
  search: string;
}>({
  fenlei: undefined,
  search: '',
});

const pickerBatchPrice = ref<number | undefined>(undefined);
const pickerBatchMarkupRate = ref<number | undefined>(undefined);
const pickerBatchSort = ref(0);
const pickerBatchStatus = ref(1);

const form = reactive<ProductForm>({
  id: 0,
  cid: 0,
  retail_price: 0,
  sort: 0,
  status: 1,
  display_name: '',
  cover_url: '',
  description: '',
  category_id: undefined,
  category_name: '',
});

const categoryOptions = computed(() =>
  categories.value.map((item) => ({
    label: item.name,
    value: item.id,
  })),
);

const mallCategoryOptions = computed(() =>
  mallCategories.value
    .filter((item) => item.status === 1)
    .map((item) => ({
      label: item.name,
      value: item.id,
    })),
);

const pickerRowSelection = computed<any>(() => ({
  type: pickerMode.value === 'single' ? 'radio' : 'checkbox',
  selectedRowKeys: pickerSelectedRowKeys.value,
  onChange: (keys: Array<string | number>, rows: ClassItem[]) => {
    pickerSelectedRowKeys.value = keys.map((item) => Number(item));
    pickerSelectedCourses.value = rows;
  },
}));

async function load() {
  loading.value = true;
  try {
    const res = await getTenantProductsApi();
    products.value = Array.isArray(res) ? res : [];
  } catch (e: any) {
    message.error(e?.message || '加载失败');
  } finally {
    loading.value = false;
  }
}

async function loadCategories() {
  try {
    const res = await getClassCategoriesApi();
    categories.value = Array.isArray(res) ? res : [];
  } catch {
    categories.value = [];
  }
}

async function loadMallCategories() {
  try {
    const res = await getTenantMallCategoriesApi();
    mallCategories.value = Array.isArray(res) ? res : [];
  } catch {
    mallCategories.value = [];
  }
}

function resetForm() {
  Object.assign(form, {
    id: 0,
    cid: 0,
    retail_price: 0,
    sort: 0,
    status: 1,
    display_name: '',
    cover_url: '',
    description: '',
    category_id: undefined,
    category_name: '',
  });
  selectedClass.value = null;
}

function openEdit(p?: TenantProduct) {
  resetForm();
  if (p) {
    Object.assign(form, {
      id: p.id,
      cid: p.cid,
      retail_price: Number(p.retail_price || 0),
      sort: p.sort,
      status: p.status,
      display_name: p.display_name || '',
      cover_url: p.cover_url || '',
      description: p.description || '',
      category_id: p.category_id,
      category_name: p.category_name || '',
    });
    selectedClass.value = {
      cid: p.cid,
      name: p.class_name,
      price: p.supply_price || '0',
      status: p.status,
      fenlei: p.fenlei,
    };
  }
  editVisible.value = true;
}

async function loadPickerCourses(page = 1) {
  pickerLoading.value = true;
  pickerPagination.page = page;
  try {
    const res = await getClassListPagedApi({
      page,
      limit: pickerPagination.limit,
      search: pickerFilters.search.trim() || undefined,
      fenlei: pickerFilters.fenlei,
    });
    pickerList.value = (res?.list ?? []).filter((item) => item.status === 1);
    pickerPagination.total = res?.pagination?.total ?? pickerList.value.length;
  } catch (e: any) {
    message.error(e?.message || '课程库加载失败');
  } finally {
    pickerLoading.value = false;
  }
}

async function selectCurrentPageCourses() {
  if (pickerMode.value !== 'batch') return;
  pickerSelectedRowKeys.value = pickerList.value.map((item) => item.cid);
  pickerSelectedCourses.value = [...pickerList.value];
}

async function selectAllFilteredCourses() {
  if (pickerMode.value !== 'batch') return;
  pickerSelectingAll.value = true;
  try {
    const res = await getClassListPagedApi({
      page: 1,
      limit: Math.max(pickerPagination.total || 0, pickerPagination.limit || 10, 10),
      search: pickerFilters.search.trim() || undefined,
      fenlei: pickerFilters.fenlei,
    });
    const list = (res?.list ?? []).filter((item) => item.status === 1);
    pickerSelectedCourses.value = list;
    pickerSelectedRowKeys.value = list.map((item) => item.cid);
    message.success(`已选中当前筛选结果 ${list.length} 个课程`);
  } catch (e: any) {
    message.error(e?.message || '全选失败');
  } finally {
    pickerSelectingAll.value = false;
  }
}

function clearPickerSelection() {
  pickerSelectedRowKeys.value = [];
  pickerSelectedCourses.value = [];
}

function openPicker(mode: 'single' | 'batch') {
  pickerMode.value = mode;
  pickerVisible.value = true;
  pickerSelectedRowKeys.value = [];
  pickerSelectedCourses.value = [];
  pickerPagination.page = 1;
  if (mode === 'batch') {
    pickerBatchPrice.value = undefined;
    pickerBatchMarkupRate.value = undefined;
    pickerBatchSort.value = 0;
    pickerBatchStatus.value = 1;
  }
  if (mode === 'single' && selectedClass.value) {
    pickerSelectedRowKeys.value = [selectedClass.value.cid];
    pickerSelectedCourses.value = [selectedClass.value];
  }
  void loadPickerCourses(1);
}

function applySelectedCourse(course: ClassItem) {
  selectedClass.value = course;
  form.cid = course.cid;
  if (!form.retail_price || form.retail_price <= 0) {
    form.retail_price = Number(course.price || 0);
  }
  if (!form.display_name) {
    form.display_name = course.name || '';
  }
  if (!form.category_name) {
    form.category_name = course.fenlei || '';
  }
  if (!form.description) {
    form.description = course.content || course.noun || '';
  }
}

function calcBatchRetailPrice(course: ClassItem) {
  if (pickerBatchPrice.value && pickerBatchPrice.value > 0) {
    return pickerBatchPrice.value;
  }

  const supplyPrice = Number(course.price || 0);
  if (!supplyPrice || supplyPrice <= 0) {
    return 0;
  }

  if (
    pickerBatchMarkupRate.value !== undefined &&
    pickerBatchMarkupRate.value !== null &&
    pickerBatchMarkupRate.value >= 0
  ) {
    return Math.round(supplyPrice * (1 + pickerBatchMarkupRate.value / 100) * 100) / 100;
  }

  return supplyPrice;
}

async function handlePickerConfirm() {
  if (!pickerSelectedCourses.value.length) {
    message.warning('请先选择课程');
    return;
  }

  if (pickerMode.value === 'single') {
    applySelectedCourse(pickerSelectedCourses.value[0]!);
    pickerVisible.value = false;
    return;
  }

  pickerSubmitting.value = true;
  try {
    const selected = [...pickerSelectedCourses.value];
    const results = await Promise.allSettled(
      selected.map(async (course) => {
        const retailPrice = calcBatchRetailPrice(course);
        if (!retailPrice || retailPrice <= 0) {
          throw new Error(`课程「${course.name}」供货价无效，请单独设置售价`);
        }
        await saveTenantProductApi({
          cid: course.cid,
          retail_price: retailPrice,
          sort: pickerBatchSort.value,
          status: pickerBatchStatus.value,
        });
      }),
    );

    const failed = results.filter((item) => item.status === 'rejected') as Array<
      PromiseRejectedResult
    >;
    const successCount = results.length - failed.length;

    if (successCount > 0) {
      message.success(`成功保存 ${successCount} 个商品`);
    }
    if (failed.length > 0) {
      message.warning(
        failed
          .slice(0, 3)
          .map((item) => item.reason?.message || '保存失败')
          .join('；'),
      );
    }

    pickerVisible.value = false;
    await load();
  } finally {
    pickerSubmitting.value = false;
  }
}

async function handleSave() {
  if (!form.cid) {
    message.warning('请先从课程库选择商品');
    return;
  }
  if (!form.retail_price || form.retail_price <= 0) {
    message.warning('请填写零售价');
    return;
  }
  try {
    await saveTenantProductApi({ ...form });
    message.success('保存成功');
    editVisible.value = false;
    await loadMallCategories();
    await load();
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  }
}

function handleDelete(cid: number) {
  Modal.confirm({
    title: '确认下架',
    content: '下架后C端将无法购买此商品，确定继续？',
    async onOk() {
      try {
        await deleteTenantProductApi(cid);
        message.success('已下架');
        await load();
      } catch (e: any) {
        message.error(e?.message || '操作失败');
      }
    },
  });
}

onMounted(async () => {
  await Promise.all([load(), loadCategories(), loadMallCategories()]);
});
</script>

<template>
  <Page title="选品管理" content-class="p-4">
    <Card>
      <div class="mb-4 flex items-center justify-between">
        <span class="text-sm text-gray-500">共 {{ products.length }} 个商品</span>
        <Space>
          <Button @click="openPicker('batch')">
            <template #icon><UnorderedListOutlined /></template>
            批量选品
          </Button>
          <Button type="primary" @click="openEdit()">
            <template #icon><PlusOutlined /></template>
            添加商品
          </Button>
        </Space>
      </div>

      <Table
        :data-source="products"
        :loading="loading"
        :pagination="{ pageSize: 20 }"
        row-key="id"
        size="small"
        bordered
      >
        <Table.Column title="课程ID" data-index="cid" :width="90" />
        <Table.Column title="商城展示" :width="240">
          <template #default="{ record }">
            <div class="flex items-center gap-3">
              <img
                v-if="record.cover_url"
                :src="record.cover_url"
                alt="cover"
                class="h-10 w-10 rounded-lg border border-gray-200 object-cover"
              />
              <div
                v-else
                class="flex h-10 w-10 items-center justify-center rounded-lg bg-gray-100 text-xs text-gray-400"
              >
                无图
              </div>
              <div class="min-w-0">
                <div class="truncate font-medium text-gray-800">
                  {{ record.display_name || record.class_name }}
                </div>
                <div class="truncate text-xs text-gray-400">原课程：{{ record.class_name }}</div>
              </div>
            </div>
          </template>
        </Table.Column>
        <Table.Column title="商城分类" :width="120">
          <template #default="{ record }">
            <span>{{ record.category_name || record.fenlei || '-' }}</span>
          </template>
        </Table.Column>
        <Table.Column title="供货价" :width="100">
          <template #default="{ record }">¥{{ record.supply_price }}</template>
        </Table.Column>
        <Table.Column title="零售价" :width="100">
          <template #default="{ record }">
            <span class="font-medium text-red-500">¥{{ record.retail_price }}</span>
          </template>
        </Table.Column>
        <Table.Column title="排序" data-index="sort" :width="70" />
        <Table.Column title="状态" :width="80">
          <template #default="{ record }">
            <Tag :color="record.status === 1 ? 'green' : 'default'">
              {{ record.status === 1 ? '上架' : '下架' }}
            </Tag>
          </template>
        </Table.Column>
        <Table.Column title="操作" :width="140">
          <template #default="{ record }">
            <Space>
              <Button type="link" size="small" @click="openEdit(record)">
                <EditOutlined />
                编辑
              </Button>
              <Button type="link" size="small" danger @click="handleDelete(record.cid)">
                <DeleteOutlined />
                下架
              </Button>
            </Space>
          </template>
        </Table.Column>
      </Table>
    </Card>

    <Modal
      v-model:open="editVisible"
      :title="form.id ? '编辑商品' : '添加商品'"
      @ok="handleSave"
      ok-text="保存"
    >
      <div class="space-y-4">
        <div v-if="!form.id">
          <div class="mb-2 flex items-center justify-between">
            <label class="block text-sm font-medium">课程库选品</label>
            <Button size="small" @click="openPicker('single')">
              <template #icon><SearchOutlined /></template>
              从课程库选择
            </Button>
          </div>
          <div class="text-xs text-gray-400">只能选择存在且已上架的平台课程。</div>
        </div>

        <div v-if="selectedClass" class="rounded-lg border border-gray-200 bg-gray-50 p-3">
          <div class="text-sm font-medium text-gray-800">{{ selectedClass.name }}</div>
          <div class="mt-2 flex flex-wrap gap-2 text-xs text-gray-500">
            <span>CID：{{ selectedClass.cid }}</span>
            <span>供货价：¥{{ selectedClass.price }}</span>
            <span v-if="selectedClass.fenlei">分类：{{ selectedClass.fenlei }}</span>
          </div>
        </div>

        <div>
          <label class="mb-1 block text-sm font-medium">课程ID（CID）</label>
          <InputNumber v-model:value="form.cid" :min="1" style="width: 100%" disabled />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium">商城展示名称</label>
          <Input
            v-model:value="form.display_name"
            placeholder="留空则默认使用原课程名称"
            allow-clear
          />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium">商品封面图</label>
          <Input
            v-model:value="form.cover_url"
            placeholder="支持直链图片地址，如 https://example.com/demo.jpg"
            allow-clear
          />
          <div v-if="form.cover_url" class="mt-2">
            <img
              :src="form.cover_url"
              alt="preview"
              class="h-24 w-24 rounded-lg border border-gray-200 object-cover"
            />
          </div>
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium">商城分类</label>
          <Select
            v-model:value="form.category_id"
            :options="mallCategoryOptions"
            allow-clear
            placeholder="选择商城分类；如无分类请先去「商城分类」创建"
            style="width: 100%"
            @change="
              (value) => {
                const current = mallCategories.find((item) => item.id === value);
                form.category_name = current?.name || '';
              }
            "
          />
          <div class="mt-1 text-xs text-gray-400">
            当前未选时会回退到原课程分类展示。需要新分类请前往“商城分类”页面创建。
          </div>
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium">商品介绍</label>
          <Input.TextArea
            v-model:value="form.description"
            :rows="4"
            placeholder="留空则默认使用原课程介绍"
            show-count
            :maxlength="1000"
          />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium">零售价（元）</label>
          <InputNumber
            v-model:value="form.retail_price"
            :min="0.01"
            :precision="2"
            :step="1"
            style="width: 100%"
          />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium">排序（数字越小越靠前）</label>
          <InputNumber v-model:value="form.sort" :min="0" style="width: 100%" />
        </div>
        <div class="flex items-center gap-3">
          <label class="text-sm font-medium">状态</label>
          <Switch
            v-model:checked="form.status"
            :checked-value="1"
            :un-checked-value="0"
            checked-children="上架"
            un-checked-children="下架"
          />
        </div>
      </div>
    </Modal>

    <Modal
      v-model:open="pickerVisible"
      :title="pickerMode === 'single' ? '从课程库选择商品' : '批量选品'"
      :confirm-loading="pickerSubmitting"
      ok-text="确认"
      width="920px"
      @ok="handlePickerConfirm"
    >
      <div class="space-y-4">
        <div class="flex gap-3">
          <Select
            v-model:value="pickerFilters.fenlei"
            :options="categoryOptions"
            allow-clear
            placeholder="全部分类"
            style="width: 180px"
          />
          <Input
            v-model:value="pickerFilters.search"
            placeholder="输入课程名或 CID 搜索"
            @pressEnter="loadPickerCourses(1)"
          >
            <template #prefix><SearchOutlined /></template>
          </Input>
          <Button type="primary" @click="loadPickerCourses(1)">搜索</Button>
        </div>

        <div
          v-if="pickerMode === 'batch'"
          class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-gray-200 bg-gray-50 px-4 py-3"
        >
          <div class="text-sm text-gray-500">
            当前筛选共 {{ pickerPagination.total }} 个课程，已选 {{ pickerSelectedRowKeys.length }} 个
          </div>
          <Space wrap>
            <Button size="small" @click="selectCurrentPageCourses">全选当前页</Button>
            <Button
              size="small"
              :loading="pickerSelectingAll"
              @click="selectAllFilteredCourses"
            >
              全选当前筛选结果
            </Button>
            <Button size="small" @click="clearPickerSelection">清空选择</Button>
          </Space>
        </div>

        <Table
          :data-source="pickerList"
          :loading="pickerLoading"
          :row-selection="pickerRowSelection"
          :pagination="{
            current: pickerPagination.page,
            pageSize: pickerPagination.limit,
            total: pickerPagination.total,
            onChange: (page: number) => loadPickerCourses(page),
          }"
          row-key="cid"
          size="small"
          bordered
          :scroll="{ x: 760 }"
        >
          <Table.Column title="CID" data-index="cid" :width="90" />
          <Table.Column title="课程名称" data-index="name" ellipsis />
          <Table.Column title="供货价" data-index="price" :width="100">
            <template #default="{ record }">¥{{ record.price }}</template>
          </Table.Column>
          <Table.Column title="分类" data-index="fenlei" :width="120" />
        </Table>

        <div
          v-if="pickerMode === 'batch'"
          class="rounded-lg border border-dashed border-gray-300 bg-gray-50 p-4"
        >
          <div class="mb-3 text-sm font-medium text-gray-700">
            已选 {{ pickerSelectedRowKeys.length }} 个课程，将按以下规则批量上架
          </div>
          <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
            <div>
              <label class="mb-1 block text-sm font-medium">统一零售价（可留空）</label>
              <InputNumber
                v-model:value="pickerBatchPrice"
                :min="0.01"
                :precision="2"
                :step="1"
                style="width: 100%"
                placeholder="留空则使用各自供货价"
              />
            </div>
            <div>
              <label class="mb-1 block text-sm font-medium">加价率 %（可留空）</label>
              <InputNumber
                v-model:value="pickerBatchMarkupRate"
                :min="0"
                :precision="2"
                :step="1"
                style="width: 100%"
                placeholder="例如 20 表示加价 20%"
              />
            </div>
            <div>
              <label class="mb-1 block text-sm font-medium">统一排序</label>
              <InputNumber v-model:value="pickerBatchSort" :min="0" style="width: 100%" />
            </div>
            <div>
              <label class="mb-1 block text-sm font-medium">统一状态</label>
              <Switch
                v-model:checked="pickerBatchStatus"
                :checked-value="1"
                :un-checked-value="0"
                checked-children="上架"
                un-checked-children="下架"
              />
            </div>
          </div>
          <div class="mt-3 text-xs text-gray-500">
            价格优先级：统一零售价 > 加价率 > 原供货价。已存在的商品会覆盖价格、排序和状态；商品不存在或已下架时，后端会拒绝保存。
          </div>
          <div
            v-if="pickerSelectedCourses.length > 0"
            class="mt-2 text-xs text-gray-500"
          >
            示例：{{ pickerSelectedCourses[0]?.name }} 的供货价为 ¥{{ pickerSelectedCourses[0]?.price }}，
            批量上架价将为 ¥{{ calcBatchRetailPrice(pickerSelectedCourses[0]!) }}
          </div>
        </div>
      </div>
    </Modal>
  </Page>
</template>

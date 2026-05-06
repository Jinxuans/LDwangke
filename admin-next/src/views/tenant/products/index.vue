<template>
  <div class="tenant-products-page art-full-height">
    <ElCard class="art-table-card">
      <div class="grid gap-4 xl:grid-cols-[1fr_220px_220px_auto]">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">关键词</label>
          <ElInput v-model="filters.keyword" clearable placeholder="搜索商品名、课程名或 CID" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">商城分类</label>
          <ElSelect v-model="filters.categoryId" class="w-full" clearable placeholder="全部分类">
            <ElOption
              v-for="item in mallCategories"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </ElSelect>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">状态</label>
          <ElSelect v-model="filters.status" class="w-full" clearable placeholder="全部状态">
            <ElOption label="已上架" :value="1" />
            <ElOption label="已下架" :value="0" />
          </ElSelect>
        </div>
        <div class="flex items-end gap-3">
          <ElButton @click="resetFilters">重置</ElButton>
          <ElButton type="primary" @click="openEditDialog()">新增</ElButton>
        </div>
      </div>
    </ElCard>

    <ElCard class="art-table-card mt-4">
      <ArtTableHeader :loading="loading" layout="refresh" @refresh="loadData">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">选品管理</ElTag>
            <ElTag effect="plain">商品 {{ filteredProducts.length }} 个</ElTag>
            <ElTag type="info" effect="plain">分类 {{ mallCategories.length }} 个</ElTag>
            <ElButton plain @click="openPicker('batch')">批量选品</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ElTable v-loading="loading" :data="filteredProducts" row-key="id">
        <ElTableColumn prop="cid" label="CID" width="90" align="center" />
        <ElTableColumn label="商城展示" min-width="260">
          <template #default="{ row }">
            <div class="flex items-center gap-3">
              <img
                v-if="row.cover_url"
                :src="row.cover_url"
                alt="cover"
                class="h-12 w-12 rounded-custom-sm border-full-d object-cover"
              />
              <div
                v-else
                class="flex h-12 w-12 items-center justify-center rounded-custom-sm border-full-d bg-g-100 text-xs text-g-400"
              >
                无图
              </div>
              <div class="min-w-0">
                <p class="line-clamp-1 text-sm font-medium text-g-900">
                  {{ row.display_name || row.class_name || '-' }}
                </p>
                <p class="mt-1 line-clamp-1 text-xs text-g-500">
                  原课程：{{ row.class_name || '-' }}
                </p>
              </div>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn label="分类" min-width="140">
          <template #default="{ row }">
            {{ row.category_name || row.fenlei || '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="供货价" width="120" align="right">
          <template #default="{ row }">¥{{ formatMoney(row.supply_price) }}</template>
        </ElTableColumn>
        <ElTableColumn label="零售价" width="120" align="right">
          <template #default="{ row }">
            <span class="font-semibold text-[var(--el-color-danger)]">¥{{ formatMoney(row.retail_price) }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="sort" label="排序" width="90" align="center" />
        <ElTableColumn label="状态" width="110" align="center">
          <template #default="{ row }">
            <ElTag :type="Number(row.status) === 1 ? 'success' : 'info'" effect="plain">
              {{ Number(row.status) === 1 ? '上架' : '下架' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <div class="flex items-center gap-2">
              <ElButton text type="primary" @click="openEditDialog(row)">编辑</ElButton>
              <ElButton text type="danger" @click="handleDelete(row)">下架</ElButton>
            </div>
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>

    <ElDialog v-model="editVisible" :title="editForm.id ? '编辑商品' : '新增商品'" width="760px" destroy-on-close>
      <div class="grid gap-5 lg:grid-cols-[1.04fr_0.96fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">商品资料</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">这里维护商城展示层字段，不去改动平台课程库原始内容。</p>
          </div>

          <div class="mt-5 space-y-4">
            <div v-if="!editForm.id">
              <div class="mb-2 flex items-center justify-between gap-3">
                <label class="text-sm font-medium text-g-800">课程库选品</label>
                <ElButton size="small" plain @click="openPicker('single')">从课程库选择</ElButton>
              </div>
              <p class="text-xs text-g-500">只能从总课程库里选择有效课程，避免手填 CID 导致无效商品。</p>
            </div>

            <article v-if="selectedClass" class="rounded-custom-sm border-full-d bg-g-100/60 p-4">
              <p class="text-sm font-semibold text-g-900">{{ selectedClass.name || '-' }}</p>
              <div class="mt-3 flex flex-wrap gap-2">
                <ElTag effect="plain">CID {{ selectedClass.cid }}</ElTag>
                <ElTag type="success" effect="plain">供货价 {{ formatMoney(selectedClass.price) }}</ElTag>
                <ElTag v-if="selectedClass.fenlei" type="info" effect="plain">{{ selectedClass.fenlei }}</ElTag>
              </div>
            </article>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">课程 ID</label>
              <ElInputNumber v-model="editForm.cid" class="w-full" :min="1" disabled />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">商城展示名称</label>
              <ElInput v-model="editForm.display_name" placeholder="留空则默认使用原课程名称" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">商品封面图</label>
              <ElInput v-model="editForm.cover_url" placeholder="输入图片直链地址" />
              <div v-if="editForm.cover_url" class="mt-3">
                <img
                  :src="editForm.cover_url"
                  alt="preview"
                  class="h-24 w-24 rounded-custom-sm border-full-d object-cover"
                />
              </div>
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">商品介绍</label>
              <ElInput
                v-model="editForm.description"
                type="textarea"
                :rows="5"
                placeholder="留空则默认使用课程介绍"
              />
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">售卖参数</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">右侧只维护商城侧参数，比如分类、售价、排序和上架状态。</p>
          </div>

          <div class="mt-5 space-y-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">商城分类</label>
              <ElSelect
                v-model="editForm.category_id"
                class="w-full"
                clearable
                placeholder="可选，不选则回退原课程分类"
                @change="handleEditCategoryChange"
              >
                <ElOption
                  v-for="item in enabledMallCategories"
                  :key="item.id"
                  :label="item.name"
                  :value="item.id"
                />
              </ElSelect>
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">零售价</label>
              <ElInputNumber
                v-model="editForm.retail_price"
                class="w-full"
                :min="0.01"
                :precision="2"
                :step="1"
              />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">排序</label>
              <ElInputNumber v-model="editForm.sort" class="w-full" :min="0" :step="1" />
            </div>

            <article class="rounded-custom-sm border-full-d p-4">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-g-900">商品状态</p>
                  <p class="mt-1 text-sm text-g-500">下架后前台不再展示该商品，但数据会保留。</p>
                </div>
                <ElSwitch v-model="editForm.status" :active-value="1" :inactive-value="0" />
              </div>
            </article>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="editVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog
      v-model="pickerVisible"
      :title="pickerMode === 'single' ? '从课程库选择商品' : '批量选品'"
      width="1080px"
      destroy-on-close
    >
      <div class="space-y-4">
        <div class="grid gap-4 xl:grid-cols-[220px_1fr_auto]">
          <ElSelect v-model="pickerFilters.fenlei" class="w-full" clearable placeholder="全部分类">
            <ElOption
              v-for="item in classCategories"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </ElSelect>
          <ElInput
            v-model="pickerFilters.search"
            clearable
            placeholder="输入课程名称或 CID 搜索"
            @keyup.enter="loadPickerCourses(1)"
          />
          <div class="flex gap-3">
            <ElButton type="primary" @click="loadPickerCourses(1)">搜索</ElButton>
            <ElButton @click="resetPickerFilters">重置</ElButton>
          </div>
        </div>

        <div
          v-if="pickerMode === 'batch'"
          class="rounded-custom-sm border-full-d bg-g-100/60 p-4"
        >
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div class="text-sm text-g-600">
              当前筛选共 {{ pickerPagination.total }} 个课程，已选 {{ pickerSelectedIds.length }} 个
            </div>
            <div class="flex flex-wrap gap-2">
              <ElButton size="small" plain @click="selectCurrentPage">全选当前页</ElButton>
              <ElButton size="small" plain :loading="selectingAll" @click="selectAllFiltered">
                全选筛选结果
              </ElButton>
              <ElButton size="small" plain @click="clearPickerSelection">清空</ElButton>
            </div>
          </div>

          <div class="mt-4 grid gap-4 md:grid-cols-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">统一零售价</label>
              <ElInputNumber
                v-model="batchForm.retail_price"
                class="w-full"
                :min="0.01"
                :precision="2"
                :step="1"
                placeholder="可留空"
              />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">加价率 %</label>
              <ElInputNumber
                v-model="batchForm.markup_rate"
                class="w-full"
                :min="0"
                :precision="2"
                :step="1"
                placeholder="可留空"
              />
            </div>
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">统一排序</label>
              <ElInputNumber v-model="batchForm.sort" class="w-full" :min="0" :step="1" />
            </div>
            <div class="rounded-custom-sm border-full-d p-4">
              <div class="flex items-center justify-between gap-3">
                <span class="text-sm font-medium text-g-900">批量状态</span>
                <ElSwitch v-model="batchForm.status" :active-value="1" :inactive-value="0" />
              </div>
            </div>
          </div>
        </div>

        <ElTable
          v-loading="pickerLoading"
          :data="pickerCourses"
          row-key="cid"
          @selection-change="handlePickerSelectionChange"
          @current-change="handlePickerCurrentChange"
        >
          <ElTableColumn v-if="pickerMode === 'batch'" type="selection" width="48" reserve-selection />
          <ElTableColumn prop="cid" label="CID" width="90" align="center" />
          <ElTableColumn prop="name" label="课程名称" min-width="280" />
          <ElTableColumn label="供货价" width="120" align="right">
            <template #default="{ row }">¥{{ formatMoney(row.price) }}</template>
          </ElTableColumn>
          <ElTableColumn prop="fenlei" label="分类" min-width="140" />
          <ElTableColumn v-if="pickerMode === 'single'" label="操作" width="100" fixed="right">
            <template #default="{ row }">
              <ElButton text type="primary" @click="selectSingleCourse(row)">选择</ElButton>
            </template>
          </ElTableColumn>
        </ElTable>

        <div class="flex justify-end">
          <ElPagination
            background
            layout="total, prev, pager, next"
            :current-page="pickerPagination.page"
            :page-size="pickerPagination.limit"
            :total="pickerPagination.total"
            @current-change="loadPickerCourses"
          />
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="pickerVisible = false">取消</ElButton>
          <ElButton
            v-if="pickerMode === 'batch'"
            type="primary"
            :loading="batchSaving"
            @click="handleBatchSave"
          >
            批量上架
          </ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage, ElMessageBox } from 'element-plus'
  import {
    fetchLegacyClassCategories,
    fetchLegacyClassListPaged,
    type LegacyClassCategory,
    type LegacyClassItem
  } from '@/api/legacy/class'
  import {
    deleteTenantProduct,
    fetchTenantMallCategories,
    fetchTenantProducts,
    saveTenantProduct,
    type LegacyTenantMallCategory,
    type LegacyTenantProduct
  } from '@/api/legacy/tenant'

  defineOptions({ name: 'TenantProductsPage' })

  interface ProductEditForm {
    id: number
    cid: number
    retail_price: number
    sort: number
    status: number
    display_name: string
    cover_url: string
    description: string
    category_id?: number
    category_name: string
  }

  const loading = ref(false)
  const saving = ref(false)
  const batchSaving = ref(false)
  const selectingAll = ref(false)
  const editVisible = ref(false)
  const pickerVisible = ref(false)
  const pickerLoading = ref(false)
  const pickerMode = ref<'batch' | 'single'>('single')

  const products = ref<LegacyTenantProduct[]>([])
  const mallCategories = ref<LegacyTenantMallCategory[]>([])
  const classCategories = ref<LegacyClassCategory[]>([])
  const selectedClass = ref<LegacyClassItem | null>(null)
  const pickerCourses = ref<LegacyClassItem[]>([])
  const pickerSelectedIds = ref<number[]>([])
  const pickerSelectedCourses = ref<LegacyClassItem[]>([])

  const filters = reactive({
    categoryId: undefined as number | undefined,
    keyword: '',
    status: undefined as number | undefined
  })

  const pickerFilters = reactive({
    fenlei: undefined as number | undefined,
    search: ''
  })

  const pickerPagination = reactive({
    page: 1,
    limit: 10,
    total: 0
  })

  const batchForm = reactive({
    markup_rate: undefined as number | undefined,
    retail_price: undefined as number | undefined,
    sort: 0,
    status: 1
  })

  const createDefaultEditForm = (): ProductEditForm => ({
    id: 0,
    cid: 0,
    retail_price: 0,
    sort: 0,
    status: 1,
    display_name: '',
    cover_url: '',
    description: '',
    category_id: undefined,
    category_name: ''
  })

  const editForm = reactive<ProductEditForm>(createDefaultEditForm())

  const enabledMallCategories = computed(() =>
    mallCategories.value.filter((item) => Number(item.status) === 1)
  )

  const filteredProducts = computed(() => {
    const keyword = filters.keyword.trim().toLowerCase()
    return products.value.filter((item) => {
      const matchedKeyword =
        !keyword ||
        String(item.display_name || '').toLowerCase().includes(keyword) ||
        String(item.class_name || '').toLowerCase().includes(keyword) ||
        String(item.cid || '').includes(keyword)
      const matchedCategory =
        !filters.categoryId || Number(item.category_id || 0) === Number(filters.categoryId)
      const matchedStatus =
        filters.status === undefined || Number(item.status) === Number(filters.status)
      return matchedKeyword && matchedCategory && matchedStatus
    })
  })

  const formatMoney = (value?: number | string) => Number(value || 0).toFixed(2)

  function resetFilters() {
    filters.categoryId = undefined
    filters.keyword = ''
    filters.status = undefined
  }

  function resetEditForm() {
    Object.assign(editForm, createDefaultEditForm())
    selectedClass.value = null
  }

  function resetPickerFilters() {
    pickerFilters.fenlei = undefined
    pickerFilters.search = ''
    loadPickerCourses(1)
  }

  async function loadData() {
    loading.value = true
    try {
      const [productResult, mallCategoryResult, classCategoryResult] = await Promise.all([
        fetchTenantProducts(),
        fetchTenantMallCategories(),
        fetchLegacyClassCategories()
      ])
      products.value = Array.isArray(productResult) ? productResult : []
      mallCategories.value = Array.isArray(mallCategoryResult) ? mallCategoryResult : []
      classCategories.value = Array.isArray(classCategoryResult) ? classCategoryResult : []
    } finally {
      loading.value = false
    }
  }

  function openEditDialog(product?: LegacyTenantProduct) {
    resetEditForm()
    if (product) {
      Object.assign(editForm, {
        id: product.id,
        cid: product.cid,
        retail_price: Number(product.retail_price || 0),
        sort: Number(product.sort || 0),
        status: Number(product.status || 0),
        display_name: product.display_name || '',
        cover_url: product.cover_url || '',
        description: product.description || '',
        category_id: product.category_id,
        category_name: product.category_name || ''
      })
      selectedClass.value = {
        cid: product.cid,
        name: product.class_name,
        price: String(product.supply_price || '0'),
        status: Number(product.status || 0),
        fenlei: product.fenlei
      }
    }
    editVisible.value = true
  }

  function handleEditCategoryChange(value?: number) {
    const current = mallCategories.value.find((item) => item.id === value)
    editForm.category_name = current?.name || ''
  }

  function openPicker(mode: 'batch' | 'single') {
    pickerMode.value = mode
    pickerVisible.value = true
    pickerSelectedIds.value = []
    pickerSelectedCourses.value = []
    pickerPagination.page = 1
    batchForm.markup_rate = undefined
    batchForm.retail_price = undefined
    batchForm.sort = 0
    batchForm.status = 1
    loadPickerCourses(1)
  }

  async function loadPickerCourses(page = pickerPagination.page) {
    pickerLoading.value = true
    pickerPagination.page = page
    try {
      const result = await fetchLegacyClassListPaged({
        page: pickerPagination.page,
        limit: pickerPagination.limit,
        search: pickerFilters.search.trim() || undefined,
        fenlei: pickerFilters.fenlei
      })
      pickerCourses.value = (result.list || []).filter((item) => Number(item.status) === 1)
      pickerPagination.total = Number(result.pagination?.total || 0)
    } finally {
      pickerLoading.value = false
    }
  }

  function applySingleCourse(course: LegacyClassItem) {
    selectedClass.value = course
    editForm.cid = course.cid
    if (!editForm.retail_price || editForm.retail_price <= 0) {
      editForm.retail_price = Number(course.price || 0)
    }
    if (!editForm.display_name) {
      editForm.display_name = course.name || ''
    }
    if (!editForm.description) {
      editForm.description = course.content || course.noun || ''
    }
    if (!editForm.category_name) {
      editForm.category_name = course.fenlei || ''
    }
    pickerVisible.value = false
  }

  function selectSingleCourse(course: LegacyClassItem) {
    applySingleCourse(course)
  }

  function handlePickerCurrentChange(course?: LegacyClassItem) {
    if (pickerMode.value === 'single' && course) {
      pickerSelectedIds.value = [course.cid]
      pickerSelectedCourses.value = [course]
    }
  }

  function handlePickerSelectionChange(rows: LegacyClassItem[]) {
    if (pickerMode.value !== 'batch') return
    pickerSelectedCourses.value = rows
    pickerSelectedIds.value = rows.map((item) => item.cid)
  }

  function selectCurrentPage() {
    if (pickerMode.value !== 'batch') return
    pickerSelectedCourses.value = [...pickerCourses.value]
    pickerSelectedIds.value = pickerCourses.value.map((item) => item.cid)
  }

  async function selectAllFiltered() {
    if (pickerMode.value !== 'batch') return
    selectingAll.value = true
    try {
      const result = await fetchLegacyClassListPaged({
        page: 1,
        limit: Math.max(Number(pickerPagination.total || 0), pickerPagination.limit, 10),
        search: pickerFilters.search.trim() || undefined,
        fenlei: pickerFilters.fenlei
      })
      const allCourses = (result.list || []).filter((item) => Number(item.status) === 1)
      pickerSelectedCourses.value = allCourses
      pickerSelectedIds.value = allCourses.map((item) => item.cid)
      ElMessage.success(`已选中 ${allCourses.length} 个课程`)
    } finally {
      selectingAll.value = false
    }
  }

  function clearPickerSelection() {
    pickerSelectedCourses.value = []
    pickerSelectedIds.value = []
  }

  function getBatchRetailPrice(course: LegacyClassItem) {
    if (batchForm.retail_price && batchForm.retail_price > 0) {
      return batchForm.retail_price
    }
    const supplyPrice = Number(course.price || 0)
    if (!supplyPrice || supplyPrice <= 0) {
      return 0
    }
    if (batchForm.markup_rate !== undefined && batchForm.markup_rate !== null) {
      return Math.round(supplyPrice * (1 + batchForm.markup_rate / 100) * 100) / 100
    }
    return supplyPrice
  }

  async function handleSave() {
    if (!editForm.cid) {
      ElMessage.warning('请先从课程库选择商品')
      return
    }
    if (!editForm.retail_price || editForm.retail_price <= 0) {
      ElMessage.warning('请先填写零售价')
      return
    }

    saving.value = true
    try {
      await saveTenantProduct({ ...editForm })
      ElMessage.success('商品已保存')
      editVisible.value = false
      await loadData()
    } finally {
      saving.value = false
    }
  }

  async function handleBatchSave() {
    if (!pickerSelectedCourses.value.length) {
      ElMessage.warning('请先选择课程')
      return
    }

    batchSaving.value = true
    try {
      const results = await Promise.allSettled(
        pickerSelectedCourses.value.map((course) => {
          const retailPrice = getBatchRetailPrice(course)
          if (!retailPrice || retailPrice <= 0) {
            throw new Error(`课程「${course.name}」供货价无效`)
          }
          return saveTenantProduct({
            cid: course.cid,
            retail_price: retailPrice,
            sort: batchForm.sort,
            status: batchForm.status
          })
        })
      )

      const failed = results.filter((item) => item.status === 'rejected')
      const successCount = results.length - failed.length
      if (successCount > 0) {
        ElMessage.success(`成功保存 ${successCount} 个商品`)
      }
      if (failed.length > 0) {
        ElMessage.warning(`有 ${failed.length} 个商品保存失败，请检查课程价格或后端限制`)
      }
      pickerVisible.value = false
      await loadData()
    } finally {
      batchSaving.value = false
    }
  }

  async function handleDelete(product: LegacyTenantProduct) {
    try {
      await ElMessageBox.confirm(
        `确定下架商品「${product.display_name || product.class_name || product.cid}」吗？`,
        '下架商品',
        {
          type: 'warning'
        }
      )
    } catch {
      return
    }
    await deleteTenantProduct(product.cid)
    ElMessage.success('商品已下架')
    await loadData()
  }

  onMounted(() => {
    loadData()
  })
</script>

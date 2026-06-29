<template>
  <div class="admin-mijia-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="handleReset"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadData(pagination.current)">
        <template #left>
          <ElSpace wrap>
            <ElButton type="primary" plain @click="openCreateDialog">新增密价</ElButton>
            <ElButton plain :disabled="!selectedIds.length" @click="handleBatchDelete">批量删除</ElButton>
            <ElTag effect="plain">当前页用户 {{ currentUserCount }}</ElTag>
            <ElTag type="warning" effect="plain">已选 {{ selectedIds.length }} 项</ElTag>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        ref="tableRef"
        :loading="loading"
        :data="list"
        :columns="columns"
        :pagination="pagination"
        @pagination:current-change="handleCurrentChange"
        @pagination:size-change="handleSizeChange"
        @selection-change="handleSelectionChange"
      />
    </ElCard>

    <ElDialog v-model="createVisible" title="新增密价" width="920px" destroy-on-close>
      <div class="grid gap-5 lg:grid-cols-[1.05fr_0.95fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">设置参数</h3>
            <p class="mt-1 text-sm text-g-500">单商品或分类规则。</p>
          </div>

          <div class="mt-5 space-y-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">用户 UID</label>
              <ElInput v-model="createForm.uid" placeholder="请输入用户 UID" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">设置方式</label>
              <ElSegmented
                v-model="createForm.setType"
                :options="[
                  { label: '单个商品', value: 'single' },
                  { label: '分类规则', value: 'batch' }
                ]"
                class="w-full"
              />
            </div>

            <div v-if="createForm.setType === 'single'" class="space-y-3">
              <div class="flex flex-wrap items-center justify-between gap-3">
                <label class="text-sm font-medium text-g-800">商品</label>
                <div v-if="showCourseCategoryPanel" class="flex flex-wrap items-center gap-2">
                  <ElTag v-if="shouldShowCourseCategoryPanel && activeCategory === 'collect'" type="warning" effect="plain" size="small">收藏中</ElTag>
                  <ElButton text size="small" @click="showCourseCategoryToggle = !showCourseCategoryToggle">
                    {{ shouldShowCourseCategoryPanel ? '收起分类' : '展开分类' }}
                  </ElButton>
                </div>
              </div>

              <div v-if="shouldShowCourseCategoryPanel" class="space-y-3">
                <div v-if="categoryType === 1" class="max-w-[340px]">
                  <ElSelect
                    :model-value="activeCategory || undefined"
                    clearable
                    placeholder="全部课程"
                    class="w-full"
                    @change="handleCategorySelect"
                  >
                    <ElOption label="收藏课程" value="collect" />
                    <ElOption
                      v-for="item in courseCategories"
                      :key="item.id"
                      :label="item.name"
                      :value="String(item.id)"
                    />
                  </ElSelect>
                </div>
                <div v-else class="space-y-2">
                  <div
                    class="flex flex-wrap gap-1 overflow-hidden transition-all"
                    :class="courseCategoryExpanded ? 'max-h-none' : 'max-h-[7.75rem]'"
                  >
                    <ElButton
                      size="small"
                      class="!ml-0 !h-7 !rounded-[8px] !px-2.5 !text-xs !font-normal !shadow-none"
                      :class="
                        activeCategory === ''
                          ? '!border-[var(--el-color-primary-light-5)] !bg-[var(--el-color-primary-light-9)] !text-[var(--el-color-primary)]'
                          : '!border-[#e5e7eb] !bg-transparent !text-g-600'
                      "
                      @click="changeCategory('')"
                    >
                      全部课程
                    </ElButton>
                    <ElButton
                      size="small"
                      class="!ml-0 !h-7 !rounded-[8px] !px-2.5 !text-xs !font-normal !shadow-none"
                      :class="
                        activeCategory === 'collect'
                          ? '!border-[#ffc9de] !bg-[#fff4f8] !text-[#eb2f96]'
                          : '!border-[#e5e7eb] !bg-transparent !text-g-600'
                      "
                      @click="changeCategory(activeCategory === 'collect' ? '' : 'collect')"
                    >
                      收藏课程
                    </ElButton>
                    <ElButton
                      v-for="item in courseCategories"
                      :key="item.id"
                      size="small"
                      class="!ml-0 !h-7 !rounded-[8px] !px-2.5 !text-xs !font-normal !shadow-none"
                      :class="
                        activeCategory === String(item.id)
                          ? '!border-[var(--el-color-primary-light-5)] !bg-[var(--el-color-primary-light-9)] !text-[var(--el-color-primary)]'
                          : '!border-[#e5e7eb] !bg-transparent !text-g-600'
                      "
                      @click="changeCategory(String(item.id))"
                    >
                      <span>{{ item.name }}</span>
                      <span
                        v-if="item.recommend && activeCategory !== String(item.id)"
                        class="ml-1 rounded-full bg-[#f3e8ff] px-1.5 py-0.5 text-[10px] leading-none text-[#722ed1]"
                      >
                        荐
                      </span>
                    </ElButton>
                  </div>
                  <ElButton v-if="hasOverflowCourseCategories" text size="small" @click="courseCategoryExpanded = !courseCategoryExpanded">
                    {{ courseCategoryExpanded ? '收起分类' : '展开更多分类' }}
                  </ElButton>
                </div>
              </div>

              <ElSelect
                :model-value="createForm.cids"
                class="w-full"
                clearable
                filterable
                multiple
                collapse-tags
                collapse-tags-tooltip
                remote
                reserve-keyword
                placeholder="搜索并选择课程，可多选"
                no-data-text="当前筛选条件下没有课程"
                popper-class="admin-mijia-course-select-dropdown"
                :remote-method="handleCourseKeywordSearch"
                @change="handleCourseSelect($event, 'create')"
                @clear="handleCourseQueryClear('create')"
                @visible-change="handleCourseDropdownVisibleChange"
              >
                <ElOption
                  v-for="course in courseSelectOptions"
                  :key="course.cid"
                  :label="course.name"
                  :value="course.cid"
                >
                  <span class="text-sm text-g-900">{{ course.name }}</span>
                  <span class="ml-1 text-g-500">¥{{ formatLegacyPrice(course.price) }}</span>
                </ElOption>
              </ElSelect>
            </div>

            <div v-else>
              <label class="mb-2 block text-sm font-medium text-g-800">分类</label>
              <ElSelect
                v-model="createForm.fenlei"
                class="w-full"
                clearable
                filterable
                placeholder="请选择分类"
              >
                <ElOption
                  v-for="item in categoryOptions"
                  :key="item.id"
                  :label="`${item.name}（ID ${item.id}）`"
                  :value="item.id"
                />
              </ElSelect>
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">定价模式</label>
              <ElSelect v-model="createForm.mode" class="w-full">
                <ElOption
                  v-for="item in modeOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </ElSelect>
            </div>

            <div v-if="isCreateMultiplierMode">
              <label class="mb-2 block text-sm font-medium text-g-800">倍率</label>
              <ElInput v-model="createForm.multiplier" placeholder="例如 0.8 表示 8 折" />
            </div>

            <div v-else>
              <label class="mb-2 block text-sm font-medium text-g-800">{{ createAmountLabel }}</label>
              <ElInput v-model="createForm.priceValue" :placeholder="createAmountPlaceholder" />
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">规则摘要</h3>
            <p class="mt-1 text-sm text-g-500">确认公式与作用范围。</p>
          </div>

          <div class="mt-5 space-y-4">
            <article class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
              <div class="flex flex-wrap gap-2">
                <ElTag type="primary" effect="plain">{{ createForm.setType === 'single' ? '单个商品' : '分类规则' }}</ElTag>
                <ElTag effect="plain">{{ modeLabel(createForm.mode) }}</ElTag>
              </div>
              <p class="mt-3 text-sm leading-6 text-g-700">{{ createFormulaTip }}</p>
              <p class="mt-1 text-xs leading-5 text-g-500">{{ modeBoundaryTip }}</p>
            </article>

            <article
              v-if="createForm.setType === 'single'"
              class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3"
            >
              <p class="text-sm font-semibold text-g-900">{{ selectedClassLabel }}</p>
              <p v-if="createSelectedClasses.length <= 1" class="mt-2 text-sm leading-6 text-g-500">
                原价 {{ selectedClassPriceLabel }}，最终值 {{ createPreviewText }}。
              </p>
              <div v-else class="mt-3 space-y-2 text-xs text-g-500">
                <p
                  v-for="item in createSelectedClasses.slice(0, 4)"
                  :key="item.cid"
                  class="line-clamp-1 rounded-custom-sm border-full-d bg-[var(--el-bg-color)] px-3 py-2"
                >
                  {{ item.name }}，原价 {{ item.price }}，预估 {{ previewBatchPrice(item.price) }}
                </p>
                <p v-if="createSelectedClasses.length > 4" class="text-g-400">
                  还有 {{ createSelectedClasses.length - 4 }} 个商品未展开
                </p>
              </div>
            </article>

            <article v-else class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
              <p class="text-sm font-semibold text-g-900">{{ selectedCategoryLabel }}</p>
              <p class="mt-2 text-sm leading-6 text-g-500">
                已加载 {{ categoryPreviewProducts.length }} 个示例商品；以后加入此分类的商品会自动套用该规则。
              </p>
              <div v-if="categoryPreviewProducts.length" class="mt-3 space-y-2 text-xs text-g-500">
                <p
                  v-for="item in displayedCategoryPreviewProducts"
                  :key="item.cid"
                  class="line-clamp-1 rounded-custom-sm border-full-d bg-[var(--el-bg-color)] px-3 py-2"
                >
                  {{ item.name }}，原价 {{ item.price }}，预估 {{ previewBatchPrice(item.price) }}
                </p>
                <ElButton v-if="categoryPreviewProducts.length > 4" text size="small" @click="categoryPreviewExpanded = !categoryPreviewExpanded">
                  {{ categoryPreviewExpanded ? '收起示例' : `展开其余 ${categoryPreviewProducts.length - 4} 个示例` }}
                </ElButton>
              </div>
            </article>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="createVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleCreate">保存规则</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="editVisible" title="编辑密价" width="820px" destroy-on-close>
      <div class="grid gap-5 lg:grid-cols-[1.02fr_0.98fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">编辑规则</h3>
            <p class="mt-1 text-sm text-g-500">仅修改当前规则。</p>
          </div>

          <div class="mt-5 space-y-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">用户 UID</label>
              <ElInput v-model="editForm.uid" placeholder="请输入用户 UID" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">规则范围</label>
              <ElSegmented
                v-model="editForm.scopeType"
                :options="[
                  { label: '单个商品', value: 'product' },
                  { label: '分类规则', value: 'category' }
                ]"
                class="w-full"
              />
            </div>

            <div v-if="editForm.scopeType === 'product'" class="space-y-3">
              <div class="flex flex-wrap items-center justify-between gap-3">
                <label class="text-sm font-medium text-g-800">商品</label>
                <div v-if="showCourseCategoryPanel" class="flex flex-wrap items-center gap-2">
                  <ElTag v-if="shouldShowCourseCategoryPanel && activeCategory === 'collect'" type="warning" effect="plain" size="small">收藏中</ElTag>
                  <ElButton text size="small" @click="showCourseCategoryToggle = !showCourseCategoryToggle">
                    {{ shouldShowCourseCategoryPanel ? '收起分类' : '展开分类' }}
                  </ElButton>
                </div>
              </div>

              <div v-if="shouldShowCourseCategoryPanel" class="space-y-3">
                <div v-if="categoryType === 1" class="max-w-[340px]">
                  <ElSelect
                    :model-value="activeCategory || undefined"
                    clearable
                    placeholder="全部课程"
                    class="w-full"
                    @change="handleCategorySelect"
                  >
                    <ElOption label="收藏课程" value="collect" />
                    <ElOption
                      v-for="item in courseCategories"
                      :key="item.id"
                      :label="item.name"
                      :value="String(item.id)"
                    />
                  </ElSelect>
                </div>
                <div v-else class="flex flex-wrap gap-1">
                  <ElButton
                    size="small"
                    class="!ml-0 !h-7 !rounded-[8px] !px-2.5 !text-xs !font-normal !shadow-none"
                    :class="
                      activeCategory === ''
                        ? '!border-[var(--el-color-primary-light-5)] !bg-[var(--el-color-primary-light-9)] !text-[var(--el-color-primary)]'
                        : '!border-[#e5e7eb] !bg-transparent !text-g-600'
                    "
                    @click="changeCategory('')"
                  >
                    全部课程
                  </ElButton>
                  <ElButton
                    size="small"
                    class="!ml-0 !h-7 !rounded-[8px] !px-2.5 !text-xs !font-normal !shadow-none"
                    :class="
                      activeCategory === 'collect'
                        ? '!border-[#ffc9de] !bg-[#fff4f8] !text-[#eb2f96]'
                        : '!border-[#e5e7eb] !bg-transparent !text-g-600'
                    "
                    @click="changeCategory(activeCategory === 'collect' ? '' : 'collect')"
                  >
                    收藏课程
                  </ElButton>
                  <ElButton
                    v-for="item in courseCategories"
                    :key="item.id"
                    size="small"
                    class="!ml-0 !h-7 !rounded-[8px] !px-2.5 !text-xs !font-normal !shadow-none"
                    :class="
                      activeCategory === String(item.id)
                        ? '!border-[var(--el-color-primary-light-5)] !bg-[var(--el-color-primary-light-9)] !text-[var(--el-color-primary)]'
                        : '!border-[#e5e7eb] !bg-transparent !text-g-600'
                    "
                    @click="changeCategory(String(item.id))"
                  >
                    <span>{{ item.name }}</span>
                    <span
                      v-if="item.recommend && activeCategory !== String(item.id)"
                      class="ml-1 rounded-full bg-[#f3e8ff] px-1.5 py-0.5 text-[10px] leading-none text-[#722ed1]"
                    >
                      荐
                    </span>
                  </ElButton>
                </div>
              </div>

              <ElSelect
                :model-value="editForm.cid"
                class="w-full"
                clearable
                filterable
                remote
                reserve-keyword
                placeholder="搜索并选择课程"
                no-data-text="当前筛选条件下没有课程"
                popper-class="admin-mijia-course-select-dropdown"
                :remote-method="handleCourseKeywordSearch"
                @change="handleCourseSelect($event, 'edit')"
                @clear="handleCourseQueryClear('edit')"
                @visible-change="handleCourseDropdownVisibleChange"
              >
                <ElOption
                  v-for="course in courseSelectOptions"
                  :key="course.cid"
                  :label="course.name"
                  :value="course.cid"
                >
                  <span class="text-sm text-g-900">{{ course.name }}</span>
                  <span class="ml-1 text-g-500">¥{{ formatLegacyPrice(course.price) }}</span>
                </ElOption>
              </ElSelect>
            </div>

            <div v-else>
              <label class="mb-2 block text-sm font-medium text-g-800">分类</label>
              <ElSelect
                v-model="editForm.fenlei"
                class="w-full"
                clearable
                filterable
                placeholder="请选择分类"
              >
                <ElOption
                  v-for="item in categoryOptions"
                  :key="item.id"
                  :label="`${item.name}（ID ${item.id}）`"
                  :value="item.id"
                />
              </ElSelect>
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">定价模式</label>
              <ElSelect v-model="editForm.mode" class="w-full">
                <ElOption
                  v-for="item in modeOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </ElSelect>
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">金额 / 倍率</label>
              <ElInput v-model="editForm.price" placeholder="请输入金额或倍率" />
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">规则摘要</h3>
            <p class="mt-1 text-sm text-g-500">确认公式与结果。</p>
          </div>

          <div class="mt-5 space-y-4">
            <article class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
              <div class="flex flex-wrap gap-2">
                <ElTag effect="plain">{{ modeLabel(editForm.mode) }}</ElTag>
                <ElTag type="primary" effect="plain">UID {{ editForm.uid || '-' }}</ElTag>
              </div>
              <p class="mt-3 text-sm leading-6 text-g-700">{{ editFormulaTip }}</p>
              <p class="mt-1 text-xs leading-5 text-g-500">{{ modeBoundaryTip }}</p>
            </article>

            <article class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
              <p class="text-sm font-semibold text-g-900">{{ editScopeLabel }}</p>
              <p class="mt-2 text-sm leading-6 text-g-500">
                {{ editScopeSummary }}
              </p>
            </article>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="editVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleEdit">保存修改</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { nextTick } from 'vue'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import { fetchLegacyAdminCategoryOptions, type LegacyAdminCategory } from '@/api/legacy/admin-categories'
  import {
    deleteLegacyMiJia,
    fetchLegacyMiJiaClassOptions,
    fetchLegacyMiJiaList,
    saveLegacyMiJia,
    type LegacyMiJiaClassOption,
    type LegacyMiJiaItem
  } from '@/api/legacy/admin-mijia'
  import { formatLegacyPrice, useLegacyCourseCatalog } from '@/views/order/shared/useLegacyCourseCatalog'
  import type { LegacyClassItem } from '@/api/legacy/class'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminMiJiaPage' })

  const tableRef = ref()
  const loading = ref(false)
  const saving = ref(false)
  const createVisible = ref(false)
  const editVisible = ref(false)
  const showCourseCategoryToggle = ref(true)
  const courseCategoryExpanded = ref(false)
  const categoryPreviewExpanded = ref(false)

  const list = ref<LegacyMiJiaItem[]>([])
  const uidOptions = ref<number[]>([])
  const fallbackClassOptions = ref<LegacyMiJiaClassOption[]>([])
  const categoryPreviewProducts = ref<LegacyMiJiaClassOption[]>([])
  const categoryOptions = ref<LegacyAdminCategory[]>([])
  const selectedIds = ref<number[]>([])
  const courseDropdownScrollWrap = ref<HTMLElement | null>(null)

  const {
    categories: courseCategories,
    activeCategory,
    keyword,
    courseLoading,
    courses,
    selectedClassId,
    hasMoreCourses,
    showCategoryPanel: showCourseCategoryPanel,
    categoryType,
    loadBaseData: loadCourseBaseData,
    loadCourses,
    loadMoreCourses,
    changeCategory,
    selectCourse
  } = useLegacyCourseCatalog(12)

  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const searchForm = ref({
    uid: '',
    cid: '',
    keyword: ''
  })

  const appliedSearch = reactive({
    uid: '',
    cid: '',
    keyword: ''
  })

  const createForm = reactive({
    uid: '',
    setType: 'single' as 'single' | 'batch',
    cid: undefined as number | undefined,
    cids: [] as number[],
    fenlei: undefined as number | undefined,
    mode: '2',
    priceValue: '',
    multiplier: ''
  })

  const editForm = reactive({
    mid: 0,
    uid: '',
    scopeType: 'product' as 'product' | 'category',
    cid: undefined as number | undefined,
    fenlei: undefined as number | undefined,
    mode: '2',
    price: ''
  })

  const modeOptions = [
    { label: '价格的基础上扣除', value: '0' },
    { label: '倍数的基础上扣除', value: '1' },
    { label: '直接定价', value: '2' },
    { label: '按倍率定价', value: '3' }
  ]

  const formulaTipMap: Record<string, string> = {
    '0': '公式：最终价 = 原本售价 - 扣减金额',
    '1': '公式：最终价 = (商品原价 - 扣减金额) × 用户加价倍数',
    '2': '公式：最终价 = 直接定价金额',
    '3': '公式：最终价 = 商品原价 × 密价倍率',
    '4': '公式：最终价 = 商品原价 × 密价倍率'
  }

  const modeBoundaryTip = '说明：最终价格不会低于 0，也不会高于原本售价。'

  const currentUserCount = computed(() => new Set(list.value.map((item) => item.uid)).size)
  const isCreateMultiplierMode = computed(() => createForm.mode === '3')
  const createFormulaTip = computed(() => formulaTipMap[createForm.mode] || '请按当前模式填写参数')
  const editFormulaTip = computed(() => formulaTipMap[editForm.mode] || '请按当前模式填写参数')
  const createAmountLabel = computed(() => (createForm.mode === '2' ? '定价金额' : '扣减金额'))
  const createAmountPlaceholder = computed(() =>
    createForm.mode === '2' ? '请输入定价金额' : '请输入扣减金额'
  )
  const shouldShowCourseCategoryPanel = computed(
    () => showCourseCategoryPanel.value && showCourseCategoryToggle.value
  )
  const courseSelectOptions = computed(() => courses.value)
  const fallbackClassMap = computed(() => {
    const map = new Map<number, LegacyMiJiaClassOption>()
    fallbackClassOptions.value.forEach((item) => map.set(item.cid, item))
    return map
  })
  const hasOverflowCourseCategories = computed(() => courseCategories.value.length > 14)
  const createSelectedClass = computed(() => findSelectedClass(createForm.cid))
  const createSelectedClasses = computed(() =>
    createForm.cids
      .map((cid) => findSelectedClass(cid))
      .filter((item): item is LegacyClassItem | LegacyMiJiaClassOption => Boolean(item))
  )
  const editSelectedClass = computed(() => findSelectedClass(editForm.cid))
  const editSelectedCategory = computed(() =>
    categoryOptions.value.find((item) => item.id === editForm.fenlei) || null
  )
  const selectedClassLabel = computed(() => {
    if (createForm.cids.length > 1) {
      return `已选择 ${createForm.cids.length} 个商品`
    }
    return createSelectedClass.value?.name || '未选择商品'
  })
  const selectedClassPriceLabel = computed(() => `${createSelectedClass.value?.price || '0'} 币`)
  const editClassLabel = computed(() => editSelectedClass.value?.name || '未选择商品')
  const editClassPriceLabel = computed(() => `${editSelectedClass.value?.price || '0'} 币`)
  const editScopeLabel = computed(() =>
    editForm.scopeType === 'category'
      ? editSelectedCategory.value?.name || '未选择分类'
      : editClassLabel.value
  )
  const editScopeSummary = computed(() => {
    if (editForm.scopeType === 'category') {
      return `已加载 ${categoryPreviewProducts.value.length} 个示例商品；以后加入此分类的商品会自动套用该规则。`
    }
    return `原价 ${editClassPriceLabel.value}，最终值 ${editPreviewText.value}。`
  })
  const selectedCategoryLabel = computed(() => {
    return categoryOptions.value.find((item) => item.id === createForm.fenlei)?.name || '未选择分类'
  })
  const displayedCategoryPreviewProducts = computed(() =>
    categoryPreviewExpanded.value ? categoryPreviewProducts.value : categoryPreviewProducts.value.slice(0, 4)
  )
  const searchItems = computed(() => [
    {
      label: '用户 UID',
      key: 'uid',
      type: 'select',
      props: {
        clearable: true,
        filterable: true,
        placeholder: '全部用户',
        options: uidOptions.value.map((item) => ({
          label: String(item),
          value: String(item)
        }))
      }
    },
    {
      label: '商品 CID',
      key: 'cid',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '输入 CID'
      }
    },
    {
      label: '关键词',
      key: 'keyword',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '搜索课程名称'
      }
    }
  ])

  const modeLabel = (mode: string) =>
    modeOptions.find((item) => item.value === String(mode === '4' ? '3' : mode))?.label || '未知模式'

  const formatPricePreview = (base: string | number, mode: string, input: string) => {
    const baseValue = Number(base || 0)
    const currentValue = Number(input || 0)
    if (!input) {
      return '待填写'
    }

    switch (mode) {
      case '0':
        return `${Math.max(baseValue - currentValue, 0).toFixed(2)} 币`
      case '1':
        return `${Math.max(baseValue - currentValue, 0).toFixed(2)} × 用户倍率`
      case '2':
        return `${currentValue.toFixed(2)} 币`
      case '3':
      case '4':
        return `${(baseValue * currentValue).toFixed(2)} 币`
      default:
        return input
    }
  }

  const createPreviewText = computed(() => {
    const input = isCreateMultiplierMode.value ? createForm.multiplier : createForm.priceValue
    return formatPricePreview(createSelectedClass.value?.price || 0, createForm.mode, input)
  })

  const editPreviewText = computed(() =>
    formatPricePreview(editSelectedClass.value?.price || 0, editForm.mode, editForm.price)
  )

  const previewBatchPrice = (basePrice: string) => {
    const input = isCreateMultiplierMode.value ? createForm.multiplier : createForm.priceValue
    return formatPricePreview(basePrice, createForm.mode, input)
  }

  const validateRuleValue = (mode: string, value: string) => {
    const trimmed = value.trim()
    if (!trimmed) {
      return mode === '3' ? '请输入倍率' : '请输入金额'
    }
    const numeric = Number(trimmed)
    if (!Number.isFinite(numeric)) {
      return '请输入有效数字'
    }
    if (numeric < 0) {
      return '金额或倍率不能小于 0'
    }
    if (mode === '3') {
      if (numeric <= 0) {
        return '倍率必须大于 0'
      }
      if (numeric > 1) {
        return '密价倍率不能大于 1'
      }
    }
    return ''
  }

  const findSelectedClass = (cid?: number): LegacyClassItem | LegacyMiJiaClassOption | null => {
    if (!cid) {
      return null
    }
    return courses.value.find((item) => item.cid === cid) || fallbackClassMap.value.get(cid) || null
  }

  const mergeFallbackClassOptions = (items: LegacyMiJiaClassOption[]) => {
    const map = new Map<number, LegacyMiJiaClassOption>()
    for (const item of fallbackClassOptions.value) {
      map.set(item.cid, item)
    }
    for (const item of items) {
      map.set(item.cid, item)
    }
    fallbackClassOptions.value = Array.from(map.values())
  }

  const ensureClassOption = async (cid?: number) => {
    if (!cid || findSelectedClass(cid)) {
      return
    }
    const classes = await fetchLegacyMiJiaClassOptions({ cid, limit: 1 }).catch(() => [])
    if (Array.isArray(classes)) {
      mergeFallbackClassOptions(classes)
    }
  }

  const handleCategorySelect = async (value?: string) => {
    await changeCategory(value ? String(value) : '')
  }

  const handleCourseKeywordSearch = async (value: string) => {
    keyword.value = value.trim()
    await loadCourses(1)
  }

  const handleCourseQueryClear = async (target: 'create' | 'edit') => {
    if (target === 'create') {
      createForm.cid = undefined
      createForm.cids = []
    } else {
      editForm.cid = undefined
    }
    keyword.value = ''
    await loadCourses(1)
  }

  const handleCourseSelect = (value: number | number[] | undefined, target: 'create' | 'edit') => {
    if (target === 'create') {
      const cids = Array.isArray(value) ? value : value ? [value] : []
      createForm.cids = cids
      createForm.cid = cids[0]
      cids.forEach((cid) => {
        const course = courses.value.find((item) => item.cid === cid)
        if (course) {
          selectCourse(course)
        }
      })
      return
    }

    const cid = Array.isArray(value) ? value[0] : value
    if (!cid) {
      editForm.cid = undefined
      return
    }

    const course = courses.value.find((item) => item.cid === cid)
    if (course) {
      selectCourse(course)
    }
    editForm.cid = cid
  }

  const handleCourseDropdownVisibleChange = async (visible: boolean) => {
    if (visible) {
      await nextTick()
      bindCourseDropdownScroll('.admin-mijia-course-select-dropdown .el-select-dropdown__wrap')
      return
    }

    unbindCourseDropdownScroll()
  }

  const handleCourseDropdownScroll = () => {
    const wrap = courseDropdownScrollWrap.value
    if (!wrap || courseLoading.value || !hasMoreCourses.value) {
      return
    }

    const remaining = wrap.scrollHeight - wrap.scrollTop - wrap.clientHeight
    const preloadDistance = Math.max(160, wrap.clientHeight * 0.6)
    if (remaining <= preloadDistance) {
      void loadMoreCourses()
    }
  }

  const bindCourseDropdownScroll = (selector: string) => {
    unbindCourseDropdownScroll()
    const wrap = document.querySelector(selector) as HTMLElement | null
    if (!wrap) {
      return
    }

    courseDropdownScrollWrap.value = wrap
    wrap.addEventListener('scroll', handleCourseDropdownScroll, { passive: true })
  }

  const unbindCourseDropdownScroll = () => {
    if (!courseDropdownScrollWrap.value) {
      return
    }

    courseDropdownScrollWrap.value.removeEventListener('scroll', handleCourseDropdownScroll)
    courseDropdownScrollWrap.value = null
  }

  const loadCategoryPreviewProducts = async (fenlei?: number) => {
    if (!fenlei) {
      categoryPreviewProducts.value = []
      return
    }
    const classes = await fetchLegacyMiJiaClassOptions({ fenlei, limit: 20 }).catch(() => [])
    categoryPreviewProducts.value = Array.isArray(classes) ? classes : []
  }

  const clearSelection = () => {
    selectedIds.value = []
    tableRef.value?.elTableRef?.clearSelection?.()
  }

  const { columns, columnChecks } = useTableColumns<LegacyMiJiaItem>(() => [
    {
      type: 'selection',
      width: 50,
      fixed: 'left'
    },
    {
      prop: 'mid',
      label: 'MID',
      width: 90,
      align: 'center'
    },
    {
      prop: 'uid',
      label: '用户',
      width: 140,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.username || `UID ${row.uid}`),
          h('p', { class: 'mt-1 text-xs text-g-500' }, `UID ${row.uid}`)
        ])
    },
    {
      prop: 'classname',
      label: '作用范围',
      minWidth: 260,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h(
            'p',
            { class: 'font-semibold text-g-900 line-clamp-1' },
            row.scope_type === 'category' ? row.category_name || '-' : row.classname || '-'
          ),
          h(
            'p',
            { class: 'mt-1 text-xs text-g-500' },
            row.scope_type === 'category' ? `分类 ID ${row.scope_id}` : `CID ${row.cid || row.scope_id}`
          )
        ])
    },
    {
      prop: 'mode',
      label: '模式',
      width: 180,
      formatter: (row) => h(ElTag, { type: 'primary', effect: 'plain' }, () => modeLabel(row.mode))
    },
    {
      prop: 'price',
      label: '金额 / 倍率',
      width: 120,
      formatter: (row) => row.price || '-'
    },
    {
      prop: 'addtime',
      label: '添加时间',
      width: 180
    },
    {
      prop: 'operation',
      label: '操作',
      width: 140,
      fixed: 'right',
      formatter: (row) =>
        h('div', [
          h(ArtButtonTable, {
            type: 'edit',
            onClick: () => openEditDialog(row)
          }),
          h(ArtButtonTable, {
            type: 'delete',
            onClick: () => handleDelete([row.mid])
          })
        ])
    }
  ])

  const loadOptions = async () => {
    const categories = await fetchLegacyAdminCategoryOptions().catch(() => [])
    categoryOptions.value = Array.isArray(categories) ? categories : []
  }

  const loadData = async (page = pagination.current) => {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchLegacyMiJiaList({
        page: pagination.current,
        limit: pagination.size,
        uid: appliedSearch.uid ? Number(appliedSearch.uid) : undefined,
        cid: appliedSearch.cid ? Number(appliedSearch.cid) : undefined,
        keyword: appliedSearch.keyword || undefined
      })

      list.value = result.list || []
      uidOptions.value = result.uids || []
      pagination.total = Number(result.pagination?.total || 0)
      pagination.current = Number(result.pagination?.page || pagination.current)
      pagination.size = Number(result.pagination?.limit || pagination.size)
      clearSelection()
    } finally {
      loading.value = false
    }
  }

  const handleSearch = (params: { uid?: string; cid?: string; keyword?: string }) => {
    appliedSearch.uid = params.uid || ''
    appliedSearch.cid = params.cid || ''
    appliedSearch.keyword = params.keyword?.trim() || ''
    loadData(1)
  }

  const handleReset = () => {
    appliedSearch.uid = ''
    appliedSearch.cid = ''
    appliedSearch.keyword = ''
    loadData(1)
  }

  const handleCurrentChange = (page: number) => loadData(page)

  const handleSizeChange = (size: number) => {
    pagination.size = size
    loadData(1)
  }

  const handleSelectionChange = (rows: LegacyMiJiaItem[]) => {
    selectedIds.value = rows.map((item) => item.mid)
  }

  const resetCreateForm = () => {
    createForm.uid = appliedSearch.uid || ''
    createForm.setType = 'single'
    createForm.cid = undefined
    createForm.cids = []
    createForm.fenlei = undefined
    createForm.mode = '2'
    createForm.priceValue = ''
    createForm.multiplier = ''
    categoryPreviewProducts.value = []
    categoryPreviewExpanded.value = false
  }

  const openCreateDialog = () => {
    resetCreateForm()
    if (courses.value.length === 0) {
      loadCourses(1)
    }
    createVisible.value = true
  }

  const openEditDialog = async (row: LegacyMiJiaItem) => {
    editForm.mid = row.mid
    editForm.uid = String(row.uid)
    editForm.scopeType = row.scope_type === 'category' ? 'category' : 'product'
    editForm.cid = row.scope_type === 'category' ? undefined : row.cid || row.scope_id
    editForm.fenlei = row.scope_type === 'category' ? row.scope_id : undefined
    editForm.mode = row.mode === '4' ? '3' : String(row.mode || '2')
    editForm.price = String(row.price || '')
    if (editForm.scopeType === 'product') {
      keyword.value = ''
      selectedClassId.value = editForm.cid
      if (courses.value.length === 0) {
        await loadCourses(1)
      }
      await ensureClassOption(editForm.cid)
    } else {
      await loadCategoryPreviewProducts(editForm.fenlei)
    }
    editVisible.value = true
  }

  const handleCreate = async () => {
    if (!createForm.uid.trim()) {
      ElMessage.warning('请先填写用户 UID')
      return
    }

    const finalPrice = isCreateMultiplierMode.value ? createForm.multiplier.trim() : createForm.priceValue.trim()
    const valueError = validateRuleValue(createForm.mode, finalPrice)
    if (valueError) {
      ElMessage.warning(valueError)
      return
    }

    saving.value = true
    try {
      if (createForm.setType === 'single') {
        if (!createForm.cids.length) {
          ElMessage.warning('请选择商品')
          return
        }
        for (const cid of createForm.cids) {
          await saveLegacyMiJia({
            uid: Number(createForm.uid),
            cid,
            mode: createForm.mode,
            price: finalPrice
          })
        }
        ElMessage.success(`已创建 ${createForm.cids.length} 条密价规则`)
      } else {
        if (!createForm.fenlei) {
          ElMessage.warning('请选择分类')
          return
        }
        await saveLegacyMiJia({
          uid: Number(createForm.uid),
          scope_type: 'category',
          scope_id: createForm.fenlei,
          fenlei: createForm.fenlei,
          mode: createForm.mode,
          price: finalPrice
        })
        ElMessage.success('分类密价规则已创建')
      }

      createVisible.value = false
      await loadData(1)
    } finally {
      saving.value = false
    }
  }

  const handleEdit = async () => {
    if (!editForm.uid.trim()) {
      ElMessage.warning('请先填写用户 UID')
      return
    }
    if (editForm.scopeType === 'product' && !editForm.cid) {
      ElMessage.warning('请选择商品')
      return
    }
    if (editForm.scopeType === 'category' && !editForm.fenlei) {
      ElMessage.warning('请选择分类')
      return
    }
    const valueError = validateRuleValue(editForm.mode, editForm.price)
    if (valueError) {
      ElMessage.warning(valueError)
      return
    }

    saving.value = true
    try {
      await saveLegacyMiJia({
        mid: editForm.mid,
        uid: Number(editForm.uid),
        cid: editForm.scopeType === 'product' ? editForm.cid : undefined,
        scope_type: editForm.scopeType,
        scope_id: editForm.scopeType === 'category' ? editForm.fenlei : editForm.cid,
        fenlei: editForm.scopeType === 'category' ? editForm.fenlei : undefined,
        mode: editForm.mode,
        price: editForm.price.trim()
      })
      ElMessage.success('密价规则已更新')
      editVisible.value = false
      await loadData(pagination.current)
    } finally {
      saving.value = false
    }
  }

  const handleDelete = async (ids: number[]) => {
    if (!ids.length) {
      return
    }

    const shouldResetPage = ids.length === selectedIds.value.length
    await ElMessageBox.confirm(`确定删除选中的 ${ids.length} 条密价规则吗？`, '删除密价', {
      type: 'warning'
    })
    await deleteLegacyMiJia(ids)
    ElMessage.success('密价规则已删除')
    selectedIds.value = []
    await loadData(shouldResetPage ? 1 : pagination.current)
  }

  const handleBatchDelete = async () => {
    if (!selectedIds.value.length) {
      ElMessage.warning('请先选择要删除的规则')
      return
    }
    await handleDelete([...selectedIds.value])
  }

  onMounted(async () => {
    await Promise.all([loadOptions(), loadCourseBaseData()])
    await Promise.all([loadData(1), loadCourses(1)])
  })

  onUnmounted(() => {
    unbindCourseDropdownScroll()
  })

  watch(
    () => createForm.fenlei,
    (fenlei) => {
      if (createForm.setType === 'batch') {
        loadCategoryPreviewProducts(fenlei)
      }
    }
  )

  watch(
    () => editForm.fenlei,
    (fenlei) => {
      if (editForm.scopeType === 'category') {
        loadCategoryPreviewProducts(fenlei)
      }
    }
  )
</script>

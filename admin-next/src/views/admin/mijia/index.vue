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
            <p class="mt-1 text-sm text-g-500">单商品或按分类批量设置。</p>
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
                  { label: '按分类批量', value: 'batch' }
                ]"
                class="w-full"
              />
            </div>

            <div v-if="createForm.setType === 'single'">
              <label class="mb-2 block text-sm font-medium text-g-800">商品</label>
              <ElSelect
                v-model="createForm.cid"
                class="w-full"
                clearable
                filterable
                placeholder="请选择商品"
              >
                <ElOption
                  v-for="item in classOptions"
                  :key="item.cid"
                  :label="`${item.name}（原价 ${item.price}）`"
                  :value="item.cid"
                />
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
                <ElTag type="primary" effect="plain">{{ createForm.setType === 'single' ? '单个商品' : '按分类批量' }}</ElTag>
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
              <p class="mt-2 text-sm leading-6 text-g-500">
                原价 {{ selectedClassPriceLabel }}，最终值 {{ createPreviewText }}。
              </p>
            </article>

            <article v-else class="rounded-custom-sm border-full-d bg-g-100/40 px-4 py-3">
              <p class="text-sm font-semibold text-g-900">{{ selectedCategoryLabel }}</p>
              <p class="mt-2 text-sm leading-6 text-g-500">
                共命中 {{ categoryProducts.length }} 个商品，保存后会批量覆盖该用户在此分类下的密价。
              </p>
              <div v-if="categoryProducts.length" class="mt-3 space-y-2 text-xs text-g-500">
                <p
                  v-for="item in categoryProducts.slice(0, 6)"
                  :key="item.cid"
                  class="line-clamp-1 rounded-custom-sm border-full-d bg-[var(--el-bg-color)] px-3 py-2"
                >
                  {{ item.name }}，原价 {{ item.price }}，预估 {{ previewBatchPrice(item.price) }}
                </p>
                <p v-if="categoryProducts.length > 6" class="text-g-400">
                  还有 {{ categoryProducts.length - 6 }} 个商品未展开
                </p>
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
              <label class="mb-2 block text-sm font-medium text-g-800">商品</label>
              <ElSelect
                v-model="editForm.cid"
                class="w-full"
                clearable
                filterable
                placeholder="请选择商品"
              >
                <ElOption
                  v-for="item in classOptions"
                  :key="item.cid"
                  :label="`${item.name}（原价 ${item.price}）`"
                  :value="item.cid"
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
              <p class="text-sm font-semibold text-g-900">{{ editClassLabel }}</p>
              <p class="mt-2 text-sm leading-6 text-g-500">
                原价 {{ editClassPriceLabel }}，最终值 {{ editPreviewText }}。
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
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import { fetchLegacyAdminCategoryOptions, type LegacyAdminCategory } from '@/api/legacy/admin-categories'
  import {
    batchSaveLegacyMiJia,
    deleteLegacyMiJia,
    fetchLegacyMiJiaClassOptions,
    fetchLegacyMiJiaList,
    saveLegacyMiJia,
    type LegacyMiJiaClassOption,
    type LegacyMiJiaItem
  } from '@/api/legacy/admin-mijia'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminMiJiaPage' })

  const tableRef = ref()
  const loading = ref(false)
  const saving = ref(false)
  const createVisible = ref(false)
  const editVisible = ref(false)

  const list = ref<LegacyMiJiaItem[]>([])
  const uidOptions = ref<number[]>([])
  const classOptions = ref<LegacyMiJiaClassOption[]>([])
  const categoryOptions = ref<LegacyAdminCategory[]>([])
  const selectedIds = ref<number[]>([])

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
    fenlei: undefined as number | undefined,
    mode: '2',
    priceValue: '',
    multiplier: ''
  })

  const editForm = reactive({
    mid: 0,
    uid: '',
    cid: undefined as number | undefined,
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
  const selectedClass = computed(() =>
    classOptions.value.find((item) => item.cid === createForm.cid) || null
  )
  const editSelectedClass = computed(() =>
    classOptions.value.find((item) => item.cid === editForm.cid) || null
  )
  const selectedClassLabel = computed(() => selectedClass.value?.name || '未选择商品')
  const selectedClassPriceLabel = computed(() => `${selectedClass.value?.price || '0'} 币`)
  const editClassLabel = computed(() => editSelectedClass.value?.name || '未选择商品')
  const editClassPriceLabel = computed(() => `${editSelectedClass.value?.price || '0'} 币`)
  const selectedCategoryLabel = computed(() => {
    return categoryOptions.value.find((item) => item.id === createForm.fenlei)?.name || '未选择分类'
  })
  const categoryProducts = computed(() => {
    if (!createForm.fenlei) {
      return []
    }
    return classOptions.value.filter((item) => String(item.fenlei) === String(createForm.fenlei))
  })

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
      label: '商品',
      key: 'cid',
      type: 'select',
      props: {
        clearable: true,
        filterable: true,
        placeholder: '全部商品',
        options: classOptions.value.map((item) => ({
          label: item.name,
          value: String(item.cid)
        }))
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
    return formatPricePreview(selectedClass.value?.price || 0, createForm.mode, input)
  })

  const editPreviewText = computed(() =>
    formatPricePreview(editSelectedClass.value?.price || 0, editForm.mode, editForm.price)
  )

  const previewBatchPrice = (basePrice: string) => {
    const input = isCreateMultiplierMode.value ? createForm.multiplier : createForm.priceValue
    return formatPricePreview(basePrice, createForm.mode, input)
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
      label: '课程',
      minWidth: 260,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900 line-clamp-1' }, row.classname || '-'),
          h('p', { class: 'mt-1 text-xs text-g-500' }, `CID ${row.cid}`)
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
    const [categories, classes] = await Promise.all([
      fetchLegacyAdminCategoryOptions().catch(() => []),
      fetchLegacyMiJiaClassOptions().catch(() => [])
    ])
    categoryOptions.value = Array.isArray(categories) ? categories : []
    classOptions.value = Array.isArray(classes) ? classes : []
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
    createForm.fenlei = undefined
    createForm.mode = '2'
    createForm.priceValue = ''
    createForm.multiplier = ''
  }

  const openCreateDialog = () => {
    resetCreateForm()
    createVisible.value = true
  }

  const openEditDialog = (row: LegacyMiJiaItem) => {
    editForm.mid = row.mid
    editForm.uid = String(row.uid)
    editForm.cid = row.cid
    editForm.mode = row.mode === '4' ? '3' : String(row.mode || '2')
    editForm.price = String(row.price || '')
    editVisible.value = true
  }

  const handleCreate = async () => {
    if (!createForm.uid.trim()) {
      ElMessage.warning('请先填写用户 UID')
      return
    }

    const finalPrice = isCreateMultiplierMode.value ? createForm.multiplier.trim() : createForm.priceValue.trim()
    if (!finalPrice) {
      ElMessage.warning(isCreateMultiplierMode.value ? '请输入倍率' : '请输入金额')
      return
    }

    saving.value = true
    try {
      if (createForm.setType === 'single') {
        if (!createForm.cid) {
          ElMessage.warning('请选择商品')
          return
        }
        await saveLegacyMiJia({
          uid: Number(createForm.uid),
          cid: createForm.cid,
          mode: createForm.mode,
          price: finalPrice
        })
        ElMessage.success('密价规则已创建')
      } else {
        if (!createForm.fenlei) {
          ElMessage.warning('请选择分类')
          return
        }
        const result = await batchSaveLegacyMiJia({
          uid: Number(createForm.uid),
          fenlei: createForm.fenlei,
          mode: createForm.mode,
          price: finalPrice
        })
        ElMessage.success(result.msg || '批量设置成功')
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
    if (!editForm.cid) {
      ElMessage.warning('请选择商品')
      return
    }
    if (!editForm.price.trim()) {
      ElMessage.warning('请输入金额或倍率')
      return
    }

    saving.value = true
    try {
      await saveLegacyMiJia({
        mid: editForm.mid,
        uid: Number(editForm.uid),
        cid: editForm.cid,
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
    await loadOptions()
    await loadData(1)
  })
</script>

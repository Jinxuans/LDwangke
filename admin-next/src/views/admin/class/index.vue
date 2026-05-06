<template>
  <div class="admin-class-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="handleReset"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">课程 {{ pagination.total }} 条</ElTag>
            <ElTag v-if="selectedIds.length" type="primary" effect="plain">已选 {{ selectedIds.length }} 条</ElTag>
            <ElButton type="primary" plain @click="openCreateDialog">新增课程</ElButton>
            <ElButton plain :disabled="!selectedIds.length" @click="openBatchCategoryDialog">
              批量改分类
            </ElButton>
            <ElButton plain :disabled="!selectedIds.length" @click="handleBatchDelete">批量删除</ElButton>
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

    <ElDialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑课程' : '新增课程'"
      width="960px"
      destroy-on-close
    >
      <div class="grid gap-5 lg:grid-cols-[1.04fr_0.96fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">课程基础信息</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">名称、价格、分类和说明是课程主数据核心字段。</p>
          </div>

          <div class="mt-5 grid gap-4 md:grid-cols-2">
            <div class="md:col-span-2">
              <label class="mb-2 block text-sm font-medium text-g-800">课程名称</label>
              <ElInput v-model="editForm.name" maxlength="120" placeholder="请输入课程名称" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">价格</label>
              <ElInput v-model="editForm.price" placeholder="例如 9.90" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">分类</label>
              <ElSelect v-model="editForm.cateId" class="w-full" filterable placeholder="请选择分类">
                <ElOption
                  v-for="item in categoryOptions"
                  :key="item.id"
                  :label="item.name"
                  :value="String(item.id)"
                />
              </ElSelect>
            </div>

            <div class="md:col-span-2">
              <label class="mb-2 block text-sm font-medium text-g-800">课程描述</label>
              <ElInput
                v-model="editForm.content"
                type="textarea"
                :rows="5"
                resize="none"
                placeholder="可填写课程简介或备注"
              />
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div>
            <p class="text-sm font-semibold text-g-900">对接与状态</p>
            <p class="mt-1 text-sm leading-6 text-g-500">维护对接货源、接口编号、排序和上下架状态。</p>
          </div>

          <div class="mt-5 grid gap-4">
            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">货源</label>
              <ElSelect v-model="editForm.hid" class="w-full" filterable placeholder="可不选择">
                <ElOption label="无对接" value="0" />
                <ElOption
                  v-for="item in supplierOptions"
                  :key="item.hid"
                  :label="`${item.name} (HID:${item.hid})`"
                  :value="String(item.hid)"
                />
              </ElSelect>
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">接口编号</label>
              <ElInput v-model="editForm.noun" placeholder="上游课程编号 / noun" />
            </div>

            <div class="grid gap-4 sm:grid-cols-2">
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">排序</label>
                <ElInputNumber v-model="editForm.sort" class="w-full" :min="0" :max="999999" />
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-g-800">加价方式</label>
                <ElSelect v-model="editForm.yunsuan" class="w-full">
                  <ElOption label="乘法 (*)" value="*" />
                  <ElOption label="加法 (+)" value="+" />
                </ElSelect>
              </div>
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">上架状态</label>
              <ElSegmented
                v-model="editForm.status"
                :options="[
                  { label: '上架', value: 1 },
                  { label: '下架', value: 0 }
                ]"
                class="w-full"
              />
            </div>
          </div>

          <div class="mt-5 space-y-3 rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">课程名称</span>
              <span class="truncate font-medium text-g-900">{{ editForm.name || '未命名课程' }}</span>
            </div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">分类 / 价格</span>
              <span class="font-medium text-g-900">
                {{ getCategoryLabel(editForm.cateId) }} / ¥{{ editForm.price || '0.00' }}
              </span>
            </div>
            <div class="flex flex-wrap gap-2 pt-1">
              <ElTag :type="editForm.status === 1 ? 'success' : 'info'" effect="plain">
                {{ editForm.status === 1 ? '已上架' : '已下架' }}
              </ElTag>
              <ElTag type="primary" effect="plain">{{ getSupplierLabel(editForm.hid) }}</ElTag>
              <ElTag :type="editForm.noun ? 'warning' : 'info'" effect="plain">
                {{ editForm.noun ? `编号 ${editForm.noun}` : '未填编号' }}
              </ElTag>
            </div>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="dialogVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleSave">保存课程</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="batchCategoryVisible" title="批量修改分类" width="520px" destroy-on-close>
      <div class="space-y-4">
        <div class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-4 text-sm text-g-500">
          已选择 {{ selectedIds.length }} 个课程。
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">目标分类</label>
          <ElSelect v-model="batchCategoryId" class="w-full" filterable placeholder="请选择分类">
            <ElOption
              v-for="item in categoryOptions"
              :key="item.id"
              :label="item.name"
              :value="String(item.id)"
            />
          </ElSelect>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="batchCategoryVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="batchSaving" @click="handleBatchCategory">
            保存修改
          </ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import {
    batchChangeLegacyAdminClassCategory,
    batchDeleteLegacyAdminClasses,
    fetchLegacyAdminClasses,
    saveLegacyAdminClass,
    toggleLegacyAdminClassStatus,
    type LegacyAdminClass
  } from '@/api/legacy/admin-classes'
  import { fetchLegacyAdminCategoryOptions, type LegacyAdminCategory } from '@/api/legacy/admin-categories'
  import { fetchLegacyAdminSuppliers, type LegacyAdminSupplier } from '@/api/legacy/admin-suppliers'
  import { ElMessage, ElMessageBox, ElSwitch, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminClassPage' })

  const tableRef = ref()
  const loading = ref(false)
  const saving = ref(false)
  const batchSaving = ref(false)
  const dialogVisible = ref(false)
  const batchCategoryVisible = ref(false)

  const list = ref<LegacyAdminClass[]>([])
  const categoryOptions = ref<LegacyAdminCategory[]>([])
  const supplierOptions = ref<LegacyAdminSupplier[]>([])
  const selectedIds = ref<number[]>([])

  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const searchForm = ref<{
    cateId?: string
    keywords?: string
  }>({
    cateId: undefined,
    keywords: undefined
  })

  const appliedSearch = reactive({
    cateId: undefined as string | undefined,
    keywords: undefined as string | undefined
  })

  const editForm = reactive({
    cid: 0,
    name: '',
    price: '',
    content: '',
    cateId: '',
    status: 1,
    hid: '0',
    sort: 10,
    noun: '',
    yunsuan: '*'
  })

  const batchCategoryId = ref('')

  const isEditing = computed(() => editForm.cid > 0)
  const searchItems = computed(() => [
    {
      label: '分类',
      key: 'cateId',
      type: 'select',
      props: {
        clearable: true,
        filterable: true,
        placeholder: '全部分类',
        options: categoryOptions.value.map((item) => ({
          label: item.name,
          value: String(item.id)
        }))
      }
    },
    {
      label: '关键词',
      key: 'keywords',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '搜索课程名或 CID'
      }
    }
  ])

  const getCategoryLabel = (cateId?: string) =>
    categoryOptions.value.find((item) => String(item.id) === String(cateId || ''))?.name || '未分类'

  const getSupplierLabel = (hid?: string) => {
    if (!hid || hid === '0') return '无对接'
    return supplierOptions.value.find((item) => String(item.hid) === String(hid))?.name || `HID ${hid}`
  }

  const clearSelection = () => {
    selectedIds.value = []
    tableRef.value?.elTableRef?.clearSelection?.()
  }

  const { columns, columnChecks } = useTableColumns<LegacyAdminClass>(() => [
    {
      type: 'selection',
      width: 50,
      fixed: 'left'
    },
    {
      prop: 'cid',
      label: 'CID',
      width: 90
    },
    {
      prop: 'name',
      label: '课程信息',
      minWidth: 280,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900 line-clamp-1' }, row.name || '未命名课程'),
          h('p', { class: 'mt-1 text-xs text-g-500 line-clamp-2' }, row.content || '暂无课程说明')
        ])
    },
    {
      prop: 'cateId',
      label: '分类',
      width: 130,
      formatter: (row) => h(ElTag, { type: 'primary' }, () => getCategoryLabel(row.cateId))
    },
    {
      prop: 'price',
      label: '价格',
      width: 110,
      formatter: (row) =>
        h(
          'span',
          { class: 'font-semibold text-[var(--el-color-success)]' },
          `¥${Number(row.price || 0).toFixed(2)}`
        )
    },
    {
      prop: 'hid',
      label: '货源',
      width: 140,
      formatter: (row) => getSupplierLabel(row.hid)
    },
    {
      prop: 'noun',
      label: '接口编号',
      width: 140
    },
    {
      prop: 'sort',
      label: '排序',
      width: 90
    },
    {
      prop: 'status',
      label: '状态',
      width: 110,
      formatter: (row) =>
        h('div', { class: 'flex items-center gap-2' }, [
          h(ElSwitch, {
            modelValue: Number(row.status) === 1,
            size: 'small',
            onChange: (value: string | number | boolean) => handleToggleStatus(row, Boolean(value))
          }),
          h(ElTag, { type: Number(row.status) === 1 ? 'success' : 'info' }, () =>
            Number(row.status) === 1 ? '上架' : '下架'
          )
        ])
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
          })
        ])
    }
  ])

  const loadOptions = async () => {
    const [categories, suppliers] = await Promise.all([
      fetchLegacyAdminCategoryOptions().catch(() => []),
      fetchLegacyAdminSuppliers().catch(() => [])
    ])
    categoryOptions.value = categories || []
    supplierOptions.value = suppliers || []
  }

  const loadData = async (page = pagination.current) => {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchLegacyAdminClasses({
        page: pagination.current,
        limit: pagination.size,
        cateId: appliedSearch.cateId ? Number(appliedSearch.cateId) : undefined,
        keywords: appliedSearch.keywords
      })
      list.value = result.list || []
      pagination.total = Number(result.pagination?.total || 0)
      pagination.current = Number(result.pagination?.page || pagination.current)
      pagination.size = Number(result.pagination?.limit || pagination.size)
      clearSelection()
    } finally {
      loading.value = false
    }
  }

  const refreshData = async () => {
    await loadData(pagination.current)
  }

  const handleSearch = (params: { cateId?: string; keywords?: string }) => {
    appliedSearch.cateId = params.cateId || undefined
    appliedSearch.keywords = params.keywords?.trim() || undefined
    loadData(1)
  }

  const handleReset = () => {
    appliedSearch.cateId = undefined
    appliedSearch.keywords = undefined
    loadData(1)
  }

  const handleCurrentChange = (page: number) => {
    loadData(page)
  }

  const handleSizeChange = (size: number) => {
    pagination.size = size
    loadData(1)
  }

  const handleSelectionChange = (rows: LegacyAdminClass[]) => {
    selectedIds.value = rows.map((item) => item.cid)
  }

  const resetEditForm = () => {
    editForm.cid = 0
    editForm.name = ''
    editForm.price = ''
    editForm.content = ''
    editForm.cateId = ''
    editForm.status = 1
    editForm.hid = '0'
    editForm.sort = 10
    editForm.noun = ''
    editForm.yunsuan = '*'
  }

  const openCreateDialog = () => {
    resetEditForm()
    dialogVisible.value = true
  }

  const openEditDialog = (record: LegacyAdminClass) => {
    editForm.cid = record.cid
    editForm.name = record.name || ''
    editForm.price = record.price || ''
    editForm.content = record.content || ''
    editForm.cateId = String(record.cateId || '')
    editForm.status = Number(record.status || 0)
    editForm.hid = String(record.hid || '0')
    editForm.sort = Number(record.sort || 0)
    editForm.noun = record.noun || ''
    editForm.yunsuan = record.yunsuan || '*'
    dialogVisible.value = true
  }

  const handleSave = async () => {
    if (!editForm.name.trim()) {
      ElMessage.warning('请先填写课程名称')
      return
    }
    if (!editForm.cateId) {
      ElMessage.warning('请选择分类')
      return
    }
    if (!editForm.price.trim()) {
      ElMessage.warning('请先填写价格')
      return
    }

    saving.value = true
    try {
      await saveLegacyAdminClass({
        cid: editForm.cid || undefined,
        name: editForm.name.trim(),
        price: editForm.price.trim(),
        content: editForm.content.trim(),
        cateId: editForm.cateId,
        status: editForm.status,
        hid: editForm.hid || '0',
        sort: Number(editForm.sort || 0),
        noun: editForm.noun.trim(),
        yunsuan: editForm.yunsuan || '*'
      })
      ElMessage.success(editForm.cid ? '课程已更新' : '课程已创建')
      dialogVisible.value = false
      await loadData(editForm.cid ? pagination.current : 1)
    } finally {
      saving.value = false
    }
  }

  const handleToggleStatus = async (record: LegacyAdminClass, enabled: boolean) => {
    await toggleLegacyAdminClassStatus(record.cid, enabled ? 1 : 0)
    record.status = enabled ? 1 : 0
    ElMessage.success(enabled ? '课程已上架' : '课程已下架')
  }

  const openBatchCategoryDialog = () => {
    if (!selectedIds.value.length) {
      ElMessage.warning('请先选择课程')
      return
    }
    batchCategoryId.value = ''
    batchCategoryVisible.value = true
  }

  const handleBatchCategory = async () => {
    if (!selectedIds.value.length) {
      batchCategoryVisible.value = false
      return
    }
    if (!batchCategoryId.value) {
      ElMessage.warning('请选择目标分类')
      return
    }

    batchSaving.value = true
    try {
      const result = await batchChangeLegacyAdminClassCategory(
        selectedIds.value,
        batchCategoryId.value
      )
      ElMessage.success(result.msg || '分类修改成功')
      batchCategoryVisible.value = false
      await loadData(pagination.current)
    } finally {
      batchSaving.value = false
    }
  }

  const handleBatchDelete = async () => {
    if (!selectedIds.value.length) {
      ElMessage.warning('请先选择课程')
      return
    }

    await ElMessageBox.confirm(
      `确定删除选中的 ${selectedIds.value.length} 个课程吗？该操作不可恢复。`,
      '批量删除课程',
      { type: 'warning' }
    )
    const result = await batchDeleteLegacyAdminClasses(selectedIds.value)
    ElMessage.success(result.msg || '课程已删除')
    await loadData(pagination.current)
  }

  onMounted(async () => {
    await loadOptions()
    await loadData(1)
  })
</script>

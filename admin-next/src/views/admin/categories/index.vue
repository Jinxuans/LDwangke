<template>
  <div class="admin-categories-page art-full-height">
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
            <ElTag effect="plain">分类 {{ pagination.total }} 条</ElTag>
            <ElTag v-if="sortChanged" type="warning" effect="plain">排序待保存</ElTag>
            <ElButton type="primary" plain @click="openCreateDialog">新增分类</ElButton>
            <ElButton plain :disabled="!sortChanged" @click="saveSortChanges">保存当前排序</ElButton>
            <ElButton plain @click="openQuickDialog">快速归类</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="list"
        :columns="columns"
        :pagination="pagination"
        @pagination:current-change="handleCurrentChange"
        @pagination:size-change="handleSizeChange"
      />
    </ElCard>

    <ElDialog
      v-model="dialogVisible"
      :title="isEditing ? '编辑分类' : '新增分类'"
      width="860px"
      destroy-on-close
    >
      <div class="grid gap-5 lg:grid-cols-[1.02fr_0.98fr]">
        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <h3 class="text-lg font-semibold text-g-900">分类基础信息</h3>
            <p class="mt-1.5 text-sm leading-6 text-g-500">定义分类名称、排序和状态，课程与商品会直接使用。</p>
          </div>

          <div class="mt-5 grid gap-4 md:grid-cols-2">
            <div class="md:col-span-2">
              <label class="mb-2 block text-sm font-medium text-g-800">分类名称</label>
              <ElInput v-model="editForm.name" maxlength="40" placeholder="请输入分类名称" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">排序</label>
              <ElInputNumber v-model="editForm.sort" class="w-full" :min="0" :max="9999" />
            </div>

            <div>
              <label class="mb-2 block text-sm font-medium text-g-800">分类状态</label>
              <ElSelect v-model="editForm.status" class="w-full">
                <ElOption
                  v-for="item in statusOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </ElSelect>
            </div>
          </div>
        </section>

        <section class="rounded-custom-sm border-full-d bg-box p-5">
          <div class="border-b-d pb-4">
            <p class="text-lg font-semibold text-g-900">功能开关</p>
            <p class="mt-1 text-sm leading-6 text-g-500">控制日志、工单、改密、暂停和上游反馈等开关。</p>
          </div>

          <div class="mt-5 grid gap-3 sm:grid-cols-2">
            <article
              v-for="item in toggleItems"
              :key="item.key"
              class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-4"
            >
              <div class="flex items-start justify-between gap-3">
                <div>
                  <p class="text-sm font-semibold text-g-900">{{ item.label }}</p>
                  <p class="mt-1 text-sm leading-6 text-g-500">{{ item.desc }}</p>
                </div>
                <ElSwitch
                  :model-value="editForm[item.key] === 1"
                  @change="(value) => updateToggle(item.key, value)"
                />
              </div>
            </article>
          </div>

          <div v-if="editForm.supplier_report === 1" class="mt-5">
            <label class="mb-2 block text-sm font-medium text-g-800">反馈供应商</label>
            <ElSelect
              v-model="editForm.supplier_report_hid"
              class="w-full"
              clearable
              filterable
              placeholder="0 表示自动识别订单供应商"
            >
              <ElOption :value="0" label="自动识别（使用订单供应商）" />
              <ElOption
                v-for="item in supplierOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </ElSelect>
          </div>

          <div class="mt-5 space-y-3 rounded-custom-sm border-full-d bg-g-100/60 p-4">
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">分类名称</span>
              <span class="truncate font-medium text-g-900">{{ editForm.name || '未命名分类' }}</span>
            </div>
            <div class="flex items-center justify-between gap-3 text-sm">
              <span class="text-g-500">排序</span>
              <span class="font-medium text-g-900">{{ editForm.sort }}</span>
            </div>
            <div class="flex flex-wrap gap-2 pt-1">
              <ElTag :type="statusTagType(editForm.status)" effect="plain">{{ statusLabel(editForm.status) }}</ElTag>
              <ElTag v-for="tag in previewTags" :key="tag" type="info" effect="plain">{{ tag }}</ElTag>
              <ElTag v-if="!previewTags.length" type="info" effect="plain">无附加开关</ElTag>
            </div>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="dialogVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleSave">保存分类</ElButton>
        </div>
      </template>
    </ElDialog>

    <ElDialog v-model="quickDialogVisible" title="快速归类" width="560px" destroy-on-close>
      <div class="space-y-4">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">关键词</label>
          <ElInput v-model="quickForm.keyword" placeholder="例如：强盛" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">目标分类 ID</label>
          <ElInputNumber v-model="quickForm.categoryId" class="w-full" :min="1" :max="999999" />
        </div>
        <div class="rounded-custom-sm border-full-d bg-g-100/60 px-4 py-4 text-sm leading-6 text-g-500">
          会把“课程名称包含该关键词”的课程批量归入目标分类，适合老数据快速整理。
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="quickDialogVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="quickSaving" @click="handleQuickModify">开始修改</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import type { LegacyAdminSupplier } from '@/api/legacy/admin-suppliers'
  import { fetchLegacyAdminSuppliers } from '@/api/legacy/admin-suppliers'
  import {
    deleteLegacyAdminCategory,
    fetchLegacyAdminCategories,
    quickModifyLegacyAdminCategory,
    saveLegacyAdminCategory,
    updateLegacyAdminCategorySort,
    type LegacyAdminCategory
  } from '@/api/legacy/admin-categories'
  import { ElMessage, ElMessageBox, ElInputNumber, ElTag } from 'element-plus'

  defineOptions({ name: 'AdminCategoriesPage' })

  const loading = ref(false)
  const saving = ref(false)
  const quickSaving = ref(false)
  const dialogVisible = ref(false)
  const quickDialogVisible = ref(false)

  const list = ref<LegacyAdminCategory[]>([])
  const initialSortMap = ref<Record<number, number>>({})
  const suppliers = ref<LegacyAdminSupplier[]>([])

  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const searchForm = ref<{
    keyword?: string
    status?: string
  }>({
    keyword: undefined,
    status: undefined
  })

  const appliedSearch = reactive({
    keyword: undefined as string | undefined,
    status: undefined as string | undefined
  })

  const editForm = reactive({
    id: 0,
    name: '',
    sort: 10,
    status: '1',
    recommend: 0,
    log: 0,
    ticket: 0,
    changepass: 1,
    allowpause: 0,
    supplier_report: 0,
    supplier_report_hid: 0
  })

  const quickForm = reactive({
    keyword: '',
    categoryId: undefined as number | undefined
  })

  const isEditing = computed(() => editForm.id > 0)
  const sortChanged = computed(() =>
    list.value.some((item) => Number(initialSortMap.value[item.id] ?? item.sort) !== Number(item.sort))
  )

  const statusOptions = [
    { label: '已启用', value: '1' },
    { label: '未启用', value: '0' },
    { label: '启用分类2', value: '2' },
    { label: '启用分类3', value: '3' },
    { label: '启用分类4', value: '4' },
    { label: '启用分类5', value: '5' }
  ]

  type ToggleKey =
    | 'recommend'
    | 'log'
    | 'ticket'
    | 'changepass'
    | 'allowpause'
    | 'supplier_report'

  const toggleItems: Array<{ key: ToggleKey; label: string; desc: string }> = [
    { key: 'recommend', label: '推荐分类', desc: '前台列表可按推荐位优先展示。' },
    { key: 'log', label: '日志开关', desc: '允许课程订单查看日志。' },
    { key: 'ticket', label: '工单开关', desc: '允许针对该分类创建工单。' },
    { key: 'changepass', label: '修改密码', desc: '允许订单执行改密动作。' },
    { key: 'allowpause', label: '允许暂停', desc: '允许订单暂停和恢复。' },
    { key: 'supplier_report', label: '上游反馈', desc: '允许向供应商提交售后反馈。' }
  ]

  const searchItems = computed(() => [
    {
      label: '关键词',
      key: 'keyword',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '搜索分类名称'
      }
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: {
        clearable: true,
        placeholder: '全部状态',
        options: statusOptions
      }
    }
  ])

  const supplierOptions = computed(() =>
    suppliers.value.map((item) => ({
      value: item.hid,
      label: `${item.name} (HID:${item.hid} PT:${item.pt})`
    }))
  )

  const previewTags = computed(() => {
    const tags: string[] = []
    if (editForm.recommend === 1) tags.push('推荐')
    if (editForm.log === 1) tags.push('日志')
    if (editForm.ticket === 1) tags.push('工单')
    if (editForm.changepass === 1) tags.push('改密')
    if (editForm.allowpause === 1) tags.push('暂停')
    if (editForm.supplier_report === 1) tags.push('上游反馈')
    return tags
  })

  const statusLabel = (status: string) =>
    statusOptions.find((item) => item.value === status)?.label || `状态 ${status}`

  const statusTagType = (
    status: string
  ): 'danger' | 'info' | 'primary' | 'success' | 'warning' => {
    if (status === '1') return 'success'
    if (status === '0') return 'info'
    return 'warning'
  }

  const formatSwitchTags = (row: LegacyAdminCategory) => {
    const tags: Array<{ label: string; type: 'danger' | 'info' | 'primary' | 'success' | 'warning' }> = []
    if (row.recommend === 1) tags.push({ label: '推荐', type: 'warning' })
    if (row.log === 1) tags.push({ label: '日志', type: 'primary' })
    if (row.ticket === 1) tags.push({ label: '工单', type: 'warning' })
    if (row.changepass === 1) tags.push({ label: '改密', type: 'success' })
    if (row.allowpause === 1) tags.push({ label: '暂停', type: 'info' })
    if (row.supplier_report === 1) {
      tags.push({
        label: row.supplier_report_hid ? `反馈(HID:${row.supplier_report_hid})` : '反馈(自动)',
        type: 'danger'
      })
    }
    return tags
  }

  const { columns, columnChecks } = useTableColumns<LegacyAdminCategory>(() => [
    {
      prop: 'sort',
      label: '排序',
      width: 110,
      formatter: (row) =>
        h(ElInputNumber, {
          modelValue: row.sort,
          min: 0,
          max: 9999,
          size: 'small',
          controlsPosition: 'right',
          style: { width: '90px' },
          'onUpdate:modelValue': (value: number | undefined) => {
            row.sort = Number(value || 0)
          }
        })
    },
    {
      prop: 'name',
      label: '分类名称',
      minWidth: 220,
      formatter: (row) =>
        h('div', { class: 'leading-6' }, [
          h('p', { class: 'font-semibold text-g-900' }, row.name || '未命名分类'),
          h('p', { class: 'mt-1 text-xs text-g-500' }, `ID ${row.id}`)
        ])
    },
    {
      prop: 'status',
      label: '状态',
      width: 120,
      formatter: (row) => h(ElTag, { type: statusTagType(row.status) }, () => statusLabel(row.status))
    },
    {
      prop: 'switches',
      label: '功能开关',
      minWidth: 280,
      formatter: (row) => {
        const tags = formatSwitchTags(row)
        if (!tags.length) {
          return h('span', { class: 'text-g-400' }, '-')
        }
        return h(
          'div',
          { class: 'flex flex-wrap gap-2' },
          tags.map((tag) => h(ElTag, { type: tag.type }, () => tag.label))
        )
      }
    },
    {
      prop: 'time',
      label: '添加时间',
      width: 180
    },
    {
      prop: 'operation',
      label: '操作',
      width: 170,
      fixed: 'right',
      formatter: (row) =>
        h('div', [
          h(ArtButtonTable, {
            type: 'edit',
            onClick: () => openEditDialog(row)
          }),
          h(ArtButtonTable, {
            type: 'delete',
            onClick: () => handleDelete(row)
          })
        ])
    }
  ])

  const updateInitialSortMap = (items: LegacyAdminCategory[]) => {
    initialSortMap.value = Object.fromEntries(items.map((item) => [item.id, Number(item.sort || 0)]))
  }

  const loadSuppliers = async () => {
    suppliers.value = await fetchLegacyAdminSuppliers().catch(() => [])
  }

  const loadData = async (page = pagination.current) => {
    loading.value = true
    pagination.current = page
    try {
      const result = await fetchLegacyAdminCategories({
        page: pagination.current,
        limit: pagination.size,
        keyword: appliedSearch.keyword,
        status: appliedSearch.status
      })
      list.value = result.list || []
      pagination.total = Number(result.pagination?.total || 0)
      pagination.current = Number(result.pagination?.page || pagination.current)
      pagination.size = Number(result.pagination?.limit || pagination.size)
      updateInitialSortMap(list.value)
    } finally {
      loading.value = false
    }
  }

  const refreshData = async () => {
    await loadData(pagination.current)
  }

  const handleSearch = (params: { keyword?: string; status?: string }) => {
    appliedSearch.keyword = params.keyword?.trim() || undefined
    appliedSearch.status = params.status || undefined
    loadData(1)
  }

  const handleReset = () => {
    appliedSearch.keyword = undefined
    appliedSearch.status = undefined
    loadData(1)
  }

  const handleCurrentChange = (page: number) => {
    loadData(page)
  }

  const handleSizeChange = (size: number) => {
    pagination.size = size
    loadData(1)
  }

  const resetEditForm = () => {
    editForm.id = 0
    editForm.name = ''
    editForm.sort = 10
    editForm.status = '1'
    editForm.recommend = 0
    editForm.log = 0
    editForm.ticket = 0
    editForm.changepass = 1
    editForm.allowpause = 0
    editForm.supplier_report = 0
    editForm.supplier_report_hid = 0
  }

  const openCreateDialog = () => {
    resetEditForm()
    dialogVisible.value = true
  }

  const openEditDialog = (record: LegacyAdminCategory) => {
    editForm.id = record.id
    editForm.name = record.name || ''
    editForm.sort = Number(record.sort || 0)
    editForm.status = record.status || '1'
    editForm.recommend = Number(record.recommend || 0)
    editForm.log = Number(record.log || 0)
    editForm.ticket = Number(record.ticket || 0)
    editForm.changepass = Number(record.changepass ?? 1)
    editForm.allowpause = Number(record.allowpause || 0)
    editForm.supplier_report = Number(record.supplier_report || 0)
    editForm.supplier_report_hid = Number(record.supplier_report_hid || 0)
    dialogVisible.value = true
  }

  const updateToggle = (key: ToggleKey, value: string | number | boolean) => {
    editForm[key] = value ? 1 : 0
    if (key === 'supplier_report' && !value) {
      editForm.supplier_report_hid = 0
    }
  }

  const handleSave = async () => {
    if (!editForm.name.trim()) {
      ElMessage.warning('请先填写分类名称')
      return
    }

    saving.value = true
    try {
      await saveLegacyAdminCategory({
        id: editForm.id || undefined,
        name: editForm.name.trim(),
        sort: Number(editForm.sort || 0),
        status: editForm.status,
        recommend: editForm.recommend,
        log: editForm.log,
        ticket: editForm.ticket,
        changepass: editForm.changepass,
        allowpause: editForm.allowpause,
        supplier_report: editForm.supplier_report,
        supplier_report_hid: editForm.supplier_report_hid
      })
      ElMessage.success(editForm.id ? '分类已更新' : '分类已创建')
      dialogVisible.value = false
      await loadData(editForm.id ? pagination.current : 1)
    } finally {
      saving.value = false
    }
  }

  const handleDelete = async (record: LegacyAdminCategory) => {
    await ElMessageBox.confirm(`确定删除分类「${record.name || record.id}」吗？`, '删除分类', {
      type: 'warning'
    })
    await deleteLegacyAdminCategory(record.id)
    ElMessage.success('分类已删除')
    const nextPage =
      list.value.length === 1 && pagination.current > 1 ? pagination.current - 1 : pagination.current
    await loadData(nextPage)
  }

  const saveSortChanges = async () => {
    if (!sortChanged.value) {
      return
    }
    await updateLegacyAdminCategorySort(
      list.value.map((item) => ({
        id: item.id,
        sort: Number(item.sort || 0)
      }))
    )
    ElMessage.success('分类排序已保存')
    updateInitialSortMap(list.value)
    await loadData(pagination.current)
  }

  const openQuickDialog = () => {
    quickForm.keyword = ''
    quickForm.categoryId = undefined
    quickDialogVisible.value = true
  }

  const handleQuickModify = async () => {
    if (!quickForm.keyword.trim()) {
      ElMessage.warning('请先填写关键词')
      return
    }
    if (!quickForm.categoryId || quickForm.categoryId <= 0) {
      ElMessage.warning('请填写有效的分类 ID')
      return
    }

    quickSaving.value = true
    try {
      const result = await quickModifyLegacyAdminCategory(
        quickForm.keyword.trim(),
        quickForm.categoryId
      )
      ElMessage.success(result.msg || '快速归类完成')
      quickDialogVisible.value = false
      await loadData(pagination.current)
    } finally {
      quickSaving.value = false
    }
  }

  onMounted(() => {
    loadSuppliers()
    loadData(1)
  })
</script>

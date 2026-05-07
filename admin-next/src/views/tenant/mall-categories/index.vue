<template>
  <div class="tenant-mall-categories-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="handleReset"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadData">
        <template #left>
          <ElSpace wrap>
            <ElTag effect="plain">商城分类</ElTag>
            <ElTag effect="plain">分类 {{ filteredList.length }} 个</ElTag>
            <ElTag v-if="sortChangedCount" type="warning" effect="plain">待保存排序 {{ sortChangedCount }} 项</ElTag>
            <ElButton plain :disabled="!sortChangedCount" :loading="sortSaving" @click="handleSaveSort">
              保存排序
            </ElButton>
          </ElSpace>
        </template>
        <template #right>
          <ElButton type="primary" plain @click="openCreateDialog">新增</ElButton>
        </template>
      </ArtTableHeader>

      <ArtTable :loading="loading" :data="filteredList" :columns="columns" :show-table-header="true" row-key="id" />
    </ElCard>

    <ElDialog v-model="dialogVisible" :title="form.id ? '编辑分类' : '新增分类'" width="520px" destroy-on-close>
      <div class="space-y-4">
        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">分类名称</label>
          <ElInput v-model="form.name" maxlength="60" placeholder="请输入分类名称" />
        </div>

        <div>
          <label class="mb-2 block text-sm font-medium text-g-800">排序</label>
          <ElInputNumber v-model="form.sort" class="w-full" :min="0" :step="1" />
        </div>

        <div class="rounded-custom-sm border-full-d p-4">
          <div class="flex items-center justify-between gap-3">
            <div>
              <p class="text-sm font-medium text-g-900">分类状态</p>
              <p class="mt-1 text-sm text-g-500">禁用后前台不再展示该分类，但商品数据仍保留。</p>
            </div>
            <ElSwitch v-model="form.status" :active-value="1" :inactive-value="0" />
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <ElButton @click="dialogVisible = false">取消</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { h } from 'vue'
  import {
    deleteTenantMallCategory,
    fetchTenantMallCategories,
    saveTenantMallCategory,
    updateTenantMallCategorySort,
    type LegacyTenantMallCategory
  } from '@/api/legacy/tenant'
  import { ElButton, ElInputNumber, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import { useTableColumns } from '@/hooks/core/useTableColumns'

  defineOptions({ name: 'TenantMallCategoriesPage' })

  const loading = ref(false)
  const saving = ref(false)
  const sortSaving = ref(false)
  const dialogVisible = ref(false)
  const keyword = ref('')
  const searchForm = ref({
    keyword: ''
  })

  const list = ref<LegacyTenantMallCategory[]>([])
  const originalSortMap = ref<Record<number, number>>({})

  const form = reactive<Partial<LegacyTenantMallCategory>>({
    id: 0,
    name: '',
    sort: 10,
    status: 1
  })

  const filteredList = computed(() => {
    const value = keyword.value.trim().toLowerCase()
    return [...list.value]
      .filter((item) => !value || item.name?.toLowerCase().includes(value))
      .sort((a, b) => Number(a.sort || 0) - Number(b.sort || 0) || Number(a.id || 0) - Number(b.id || 0))
  })

  const sortChangedCount = computed(
    () =>
      list.value.filter((item) => Number(originalSortMap.value[item.id]) !== Number(item.sort || 0)).length
  )

  const searchItems = computed(() => [
    {
      label: '关键词',
      key: 'keyword',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '搜索分类名称'
      }
    }
  ])

  const { columns, columnChecks } = useTableColumns<LegacyTenantMallCategory>(() => [
    { prop: 'id', label: 'ID', width: 90, align: 'center' },
    { prop: 'name', label: '分类名称', minWidth: 220 },
    {
      prop: 'sort',
      label: '排序',
      width: 160,
      align: 'center',
      formatter: (row) =>
        h(ElInputNumber, {
          modelValue: row.sort,
          class: 'w-full',
          min: 0,
          step: 1,
          'onUpdate:modelValue': (value: number) => {
            row.sort = Number(value || 0)
          },
          onChange: () => markSortChanged(row)
        })
    },
    {
      prop: 'status',
      label: '状态',
      width: 120,
      align: 'center',
      formatter: (row) =>
        h(ElTag, { type: Number(row.status) === 1 ? 'success' : 'info', effect: 'plain' }, () =>
          Number(row.status) === 1 ? '启用' : '禁用'
        )
    },
    { prop: 'addtime', label: '添加时间', width: 180 },
    {
      prop: 'operation',
      label: '操作',
      width: 160,
      fixed: 'right',
      formatter: (row) =>
        h('div', { class: 'flex items-center gap-2' }, [
          h(ElButton, { text: true, type: 'primary', onClick: () => openEditDialog(row) }, () => '编辑'),
          h(ElButton, { text: true, type: 'danger', onClick: () => handleDelete(row) }, () => '删除')
        ])
    }
  ])

  function resetForm() {
    Object.assign(form, {
      id: 0,
      name: '',
      sort: 10,
      status: 1
    })
  }

  function handleSearch(params: { keyword?: string }) {
    keyword.value = params.keyword?.trim() || ''
  }

  function handleReset() {
    keyword.value = ''
    searchForm.value.keyword = ''
  }

  function syncOriginalSortMap(items: LegacyTenantMallCategory[]) {
    originalSortMap.value = items.reduce<Record<number, number>>((result, item) => {
      result[item.id] = Number(item.sort || 0)
      return result
    }, {})
  }

  async function loadData() {
    loading.value = true
    try {
      const result = (await fetchTenantMallCategories()) || []
      list.value = Array.isArray(result) ? result : []
      syncOriginalSortMap(list.value)
    } finally {
      loading.value = false
    }
  }

  function openCreateDialog() {
    resetForm()
    dialogVisible.value = true
  }

  function openEditDialog(row: LegacyTenantMallCategory) {
    Object.assign(form, {
      id: row.id,
      name: row.name,
      sort: row.sort,
      status: row.status
    })
    dialogVisible.value = true
  }

  function markSortChanged(row: LegacyTenantMallCategory) {
    row.sort = Number(row.sort || 0)
  }

  async function handleSave() {
    if (!String(form.name || '').trim()) {
      ElMessage.warning('请先填写分类名称')
      return
    }

    saving.value = true
    try {
      await saveTenantMallCategory({
        ...form,
        name: String(form.name || '').trim(),
        sort: Number(form.sort || 0),
        status: Number(form.status || 0)
      })
      ElMessage.success('分类已保存')
      dialogVisible.value = false
      await loadData()
    } finally {
      saving.value = false
    }
  }

  async function handleSaveSort() {
    if (!sortChangedCount.value) return
    sortSaving.value = true
    try {
      await updateTenantMallCategorySort(
        list.value.map((item) => ({
          id: item.id,
          sort: Number(item.sort || 0)
        }))
      )
      ElMessage.success('排序已保存')
      syncOriginalSortMap(list.value)
    } finally {
      sortSaving.value = false
    }
  }

  async function handleDelete(row: LegacyTenantMallCategory) {
    try {
      await ElMessageBox.confirm(`确定删除分类「${row.name}」吗？`, '删除分类', {
        type: 'warning'
      })
    } catch {
      return
    }
    await deleteTenantMallCategory(row.id)
    ElMessage.success('分类已删除')
    await loadData()
  }

  onMounted(() => {
    loadData()
  })
</script>
